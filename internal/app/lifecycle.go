package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

// Run starts all background workers and the HTTP server, then blocks until
// the process receives an OS signal or the server errors.
func (a *App) Run() error {
	// ── 1. Start background workers ───────────────────────────────────
	// Create ONE cancellable context shared by ALL projectors.
	// Shutdown() calls projectorCancel() to stop them all at once.
	projCtx, cancel := context.WithCancel(context.Background())
	a.projectorCancel = cancel

	// ── Account Projector ─────────────────────────────────────────────
	if a.AccountProjector != nil {
		// Phase 1: Catch-up rebuild (SYNCHRONOUS — guarantees read model is
		// consistent before HTTP traffic starts and before Kafka consumer begins).
		a.Logger.Info("[ Projector ] Account: starting catch-up rebuild...")
		if err := a.AccountProjector.Rebuild(projCtx); err != nil {
			a.Logger.Error("[ Projector ] Account Rebuild failed — read models may be stale",
				zap.Error(err),
			)
		}
		// Phase 2: Live Kafka streaming (async).
		go func() {
			a.Logger.Info("[ Projector ] Account Projector starting live stream...")
			a.AccountProjector.Run(projCtx)
			a.Logger.Info("[ Projector ] Account Projector stopped")
		}()
	}

	// ── Order Projector ───────────────────────────────────────────────
	if a.OrderProjector != nil {
		a.Logger.Info("[ Projector ] Order: starting catch-up rebuild...")
		if err := a.OrderProjector.Rebuild(projCtx); err != nil {
			a.Logger.Error("[ Projector ] Order Rebuild failed — read models may be stale",
				zap.Error(err),
			)
		}
		go func() {
			a.Logger.Info("[ Projector ] Order Projector starting live stream...")
			a.OrderProjector.Run(projCtx)
			a.Logger.Info("[ Projector ] Order Projector stopped")
		}()
	}
	// ── 2. Start HTTP server ───────────────────────────────────────────
	errChan := make(chan error, 1)
	go func() {
		addr := fmt.Sprintf(":%d", a.Config.App.Port)
		a.Logger.Info("Starting HTTP server",
			zap.String("address", addr),
			zap.String("env", a.Config.App.Env),
		)
		if err := a.Echo.Start(addr); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// ── 3. Wait for shutdown signal or server error ────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case <-quit:
		a.Logger.Info("Received shutdown signal")
	case err := <-errChan:
		a.Logger.Error("Server failed to start", zap.Error(err))
		return err
	}

	return a.Shutdown()
}

// Shutdown gracefully stops all resources in reverse startup order.
func (a *App) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	a.Logger.Info("Shutting down gracefully...")

	// ── 1. HTTP server – stop accepting new requests first ────────────
	if err := a.shutdownHTTPServer(ctx); err != nil {
		a.Logger.Error("HTTP server shutdown failed", zap.Error(err))
	}

	// ── 2. Background workers ─────────────────────────────────────────
	// Cancel the projectors' context so their FetchMessage loops exit.
	if a.projectorCancel != nil {
		a.Logger.Info("Stopping Projectors (Account + Order)...")
		a.projectorCancel()
	}

	// ── 3. Infrastructure connections ─────────────────────────────────
	a.closeResources()

	// ── 4. Flush logger ───────────────────────────────────────────────
	_ = a.Logger.Sync() // Ignore sync errors on Windows

	a.Logger.Info("Shutdown completed successfully")
	return nil
}

// shutdownHTTPServer sends a graceful shutdown signal to Echo.
func (a *App) shutdownHTTPServer(ctx context.Context) error {
	if a.Echo == nil {
		return nil
	}
	if err := a.Echo.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown HTTP server: %w", err)
	}
	a.Logger.Info("HTTP server stopped")
	return nil
}

// closeResources closes external infrastructure connections in order.
func (a *App) closeResources() {
	// Database
	if a.DB != nil {
		if sqlDB, err := a.DB.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				a.Logger.Error("Failed to close database", zap.Error(err))
			} else {
				a.Logger.Info("Database connection closed")
			}
		}
	}

	// Redis
	if a.Redis != nil {
		if err := a.Redis.Close(); err != nil {
			a.Logger.Error("Failed to close Redis", zap.Error(err))
		} else {
			a.Logger.Info("Redis connection closed")
		}
	}

	// Kafka (writer flushes pending messages before closing)
	if a.Kafka != nil {
		if err := a.Kafka.Close(); err != nil {
			a.Logger.Error("Failed to close Kafka writer", zap.Error(err))
		} else {
			a.Logger.Info("Kafka writer closed")
		}
	}
}
