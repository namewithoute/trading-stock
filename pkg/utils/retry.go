package utils

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// DoWithRetry executes an operation with smart retry logic based on context timeout
func DoWithRetry(ctx context.Context, logger *zap.Logger, opName string, retryInterval time.Duration, operation func() error) error {
	for {
		// 1. Check timeout immediately
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for %s: %w", opName, ctx.Err())
		default:
		}

		// 2. Execute operation
		if err := operation(); err == nil {
			logger.Info(fmt.Sprintf("Successfully connected to %s", opName))
			return nil // Success!
		} else {
			logger.Warn(fmt.Sprintf("Failed to connect to %s, retrying...", opName), zap.Error(err))
		}

		// 3. Wait (Backoff) or Cancel
		timer := time.NewTimer(retryInterval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return fmt.Errorf("timeout waiting for %s: %w", opName, ctx.Err())
		case <-timer.C:
			// Continue loop
		}
	}
}
