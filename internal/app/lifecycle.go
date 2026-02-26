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
	// ONE shared context cancels ALL background goroutines on shutdown.
	workerCtx, cancel := context.WithCancel(context.Background())
	a.workerCancel = cancel

	// ── Account Projector ─────────────────────────────────────────────
	if a.AccountProjector != nil {
		a.Logger.Info("[ Projector ] Account: starting catch-up rebuild...")
		if err := a.AccountProjector.Rebuild(workerCtx); err != nil {
			a.Logger.Error("[ Projector ] Account Rebuild failed — read models may be stale", zap.Error(err))
		}
		go func() {
			a.Logger.Info("[ Projector ] Account Projector starting live stream...")
			a.AccountProjector.Run(workerCtx)
			a.Logger.Info("[ Projector ] Account Projector stopped")
		}()
	}

	// ── Order Projector ───────────────────────────────────────────────
	if a.OrderProjector != nil {
		a.Logger.Info("[ Projector ] Order: starting catch-up rebuild...")
		if err := a.OrderProjector.Rebuild(workerCtx); err != nil {
			a.Logger.Error("[ Projector ] Order Rebuild failed — read models may be stale", zap.Error(err))
		}
		go func() {
			a.Logger.Info("[ Projector ] Order Projector starting live stream...")
			a.OrderProjector.Run(workerCtx)
			a.Logger.Info("[ Projector ] Order Projector stopped")
		}()
	}

	// ── Outbox Relay ──────────────────────────────────────────────────
	if a.OutboxRelay != nil {
		go func() {
			a.OutboxRelay.Run(workerCtx)
		}()
	}

	// ── Matching Consumer ─────────────────────────────────────────────
	if a.MatchingConsumer != nil {
		go func() {
			a.MatchingConsumer.Run(workerCtx)
		}()
	}

	// ── Order Fill Consumer ───────────────────────────────────────────
	if a.OrderFillConsumer != nil {
		go func() {
			a.OrderFillConsumer.Run(workerCtx)
		}()
	}

	// ── Account Trade Consumer ────────────────────────────────────────
	if a.AccountTradeConsumer != nil {
		go func() {
			a.AccountTradeConsumer.Run(workerCtx)
		}()
	}

	// ── Market Trade Consumer ─────────────────────────────────────────
	if a.MarketTradeConsumer != nil {
		go func() {
			a.MarketTradeConsumer.Run(workerCtx)
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
	if a.workerCancel != nil {
		a.Logger.Info("Stopping all background workers...")
		a.workerCancel()
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
