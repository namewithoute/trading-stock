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

// Run starts the HTTP server and handles graceful shutdown
func (a *App) Run() error {
	// Start server in goroutine
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

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case <-quit:
		a.Logger.Info("Received shutdown signal")
	case err := <-errChan:
		a.Logger.Error("Server failed to start", zap.Error(err))
		return err
	}

	// Graceful shutdown
	return a.Shutdown()
}

// Shutdown gracefully shuts down all resources
func (a *App) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	a.Logger.Info("Shutting down gracefully...")

	// Shutdown HTTP server
	if err := a.shutdownHTTPServer(ctx); err != nil {
		a.Logger.Error("HTTP server shutdown failed", zap.Error(err))
	}

	// Close infrastructure connections
	a.closeResources()

	// Sync logger
	_ = a.Logger.Sync() // Ignore sync errors on Windows

	a.Logger.Info("Shutdown completed successfully")
	return nil
}

// shutdownHTTPServer gracefully shuts down the HTTP server
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

// closeInfrastructure closes all infrastructure connections
func (a *App) closeResources() {
	// Close database
	if a.DB != nil {
		if sqlDB, err := a.DB.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				a.Logger.Error("Failed to close database", zap.Error(err))
			} else {
				a.Logger.Info("Database connection closed")
			}
		}
	}

	// Close Redis
	if a.Redis != nil {
		if err := a.Redis.Close(); err != nil {
			a.Logger.Error("Failed to close Redis", zap.Error(err))
		} else {
			a.Logger.Info("Redis connection closed")
		}
	}

	// Close Kafka
	if a.Kafka != nil {
		if err := a.Kafka.Close(); err != nil {
			a.Logger.Error("Failed to close Kafka", zap.Error(err))
		} else {
			a.Logger.Info("Kafka writer closed")
		}
	}
}
