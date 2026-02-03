package bootstrap

import (
	"context"
	"fmt"
	"sync"
	"time"

	"trading-stock/internal/global"
	"trading-stock/internal/initialize"

	"go.uber.org/zap"
)

// ShutdownConfig defines graceful shutdown behavior
type ShutdownConfig struct {
	ServerShutdownTimeout   time.Duration // Timeout for HTTP server shutdown
	ResourceCleanupTimeout  time.Duration // Timeout for closing DB/Redis/Kafka
	GracefulShutdownTimeout time.Duration // Total timeout for entire shutdown process
}

// DefaultShutdownConfig returns production-ready shutdown configuration
func DefaultShutdownConfig() ShutdownConfig {
	return ShutdownConfig{
		ServerShutdownTimeout:   60 * time.Second,  // Wait 60s for in-flight requests
		ResourceCleanupTimeout:  60 * time.Second,  // Wait 60s for resource cleanup
		GracefulShutdownTimeout: 120 * time.Second, // Total 120s max
	}
}

// GracefulShutdown orchestrates clean shutdown of all resources
// Returns error if any cleanup operation fails or times out
func GracefulShutdown(ctx context.Context, cfg ShutdownConfig, shutdownFuncs ...func() error) error {
	global.Logger.Info("Starting graceful shutdown...")

	// Create a context with total shutdown timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, cfg.GracefulShutdownTimeout)
	defer cancel()

	// Track all cleanup operations
	var wg sync.WaitGroup
	errChan := make(chan error, len(shutdownFuncs))

	// Execute all shutdown functions concurrently
	for i, fn := range shutdownFuncs {
		wg.Add(1)
		go func(index int, shutdownFunc func() error) {
			defer wg.Done()

			// Wrap each shutdown function with timeout protection
			done := make(chan error, 1)
			go func() {
				done <- shutdownFunc()
			}()

			select {
			case err := <-done:
				if err != nil {
					global.Logger.Error(
						fmt.Sprintf("Shutdown function %d failed", index),
						zap.Error(err),
					)
					errChan <- err
				}
			case <-shutdownCtx.Done():
				err := fmt.Errorf("shutdown function %d timed out", index)
				global.Logger.Warn(err.Error())
				errChan <- err
			}
		}(i, fn)
	}

	// Wait for all cleanup operations to complete
	doneChan := make(chan struct{})
	go func() {
		wg.Wait()
		close(doneChan)
	}()

	// Wait for completion or timeout
	select {
	case <-doneChan:
		global.Logger.Info("All resources cleaned up successfully")
		close(errChan)

		// Collect all errors
		var errors []error
		for err := range errChan {
			errors = append(errors, err)
		}

		if len(errors) > 0 {
			return fmt.Errorf("shutdown completed with %d errors: %v", len(errors), errors)
		}
		return nil

	case <-shutdownCtx.Done():
		global.Logger.Warn("Graceful shutdown timeout exceeded, forcing exit")
		return fmt.Errorf("shutdown timeout: %w", shutdownCtx.Err())
	}
}

// CloseAllResources closes all external connections with proper error handling
func CloseAllResources() error {
	global.Logger.Info("Cleaning up resources...")

	var errors []error

	// Close Postgres
	if err := closePostgres(); err != nil {
		errors = append(errors, fmt.Errorf("postgres cleanup failed: %w", err))
	}

	// Close Redis
	if err := closeRedis(); err != nil {
		errors = append(errors, fmt.Errorf("redis cleanup failed: %w", err))
	}

	// Close Kafka
	if err := closeKafka(); err != nil {
		errors = append(errors, fmt.Errorf("kafka cleanup failed: %w", err))
	}

	if len(errors) > 0 {
		return fmt.Errorf("resource cleanup errors: %v", errors)
	}

	global.Logger.Info("All resources closed successfully")
	return nil
}

// closePostgres wraps the initialize.ClosePosgresDB with error return
func closePostgres() error {
	if global.DB == nil {
		return nil
	}

	sqlDB, err := global.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get DB instance: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		global.Logger.Error("Failed to close Postgres", zap.Error(err))
		return err
	}

	global.Logger.Info("Postgres connection closed")
	return nil
}

// closeRedis wraps the initialize.CloseRedis with error return
func closeRedis() error {
	if global.Redis == nil {
		return nil
	}

	if err := global.Redis.Close(); err != nil {
		global.Logger.Error("Failed to close Redis", zap.Error(err))
		return err
	}

	global.Logger.Info("Redis connection closed")
	return nil
}

// closeKafka wraps the initialize.CloseKafka with error return
func closeKafka() error {
	if global.Kafka == nil {
		return nil
	}

	// Close() automatically flushes pending messages
	if err := global.Kafka.Close(); err != nil {
		global.Logger.Error("Failed to close Kafka writer", zap.Error(err))
		return err
	}

	global.Logger.Info("Kafka writer closed successfully")
	return nil
}

// Backward compatibility wrappers
func ClosePosgresDB() {
	initialize.ClosePosgresDB()
}

func CloseRedis() {
	initialize.CloseRedis()
}

func CloseKafka() {
	initialize.CloseKafka()
}
