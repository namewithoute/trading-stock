package main

import (
	"context"
	"fmt"
	"time"

	"trading-stock/pkg/utils"

	"go.uber.org/zap"
)

// Example 1: Basic retry with default config
func exampleBasicRetry() {
	logger, _ := zap.NewDevelopment()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Use default retry config (10 attempts, exponential backoff)
	cfg := utils.DefaultRetryConfig()

	err := utils.DoWithRetry(ctx, logger, "Database Connection", cfg, func() error {
		// Simulate connection attempt
		fmt.Println("Attempting to connect...")
		// return errors.New("connection refused") // Uncomment to test retry
		return nil // Success!
	})

	if err != nil {
		logger.Error("Failed after all retries", zap.Error(err))
	}
}

// Example 2: Custom retry config for fast-failing operations
func exampleCustomRetry() {
	logger, _ := zap.NewDevelopment()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Custom config: Only 3 attempts, faster backoff
	cfg := utils.RetryConfig{
		MaxAttempts:     3,
		InitialInterval: 500 * time.Millisecond,
		MaxInterval:     2 * time.Second,
		Multiplier:      2.0,
		Jitter:          true,
	}

	err := utils.DoWithRetry(ctx, logger, "API Call", cfg, func() error {
		// Simulate API call
		return fmt.Errorf("rate limit exceeded")
	})

	if err != nil {
		logger.Error("API call failed", zap.Error(err))
	}
}

// Example 3: Permanent error (should not retry)
func examplePermanentError() {
	logger, _ := zap.NewDevelopment()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg := utils.DefaultRetryConfig()

	err := utils.DoWithRetry(ctx, logger, "Authentication", cfg, func() error {
		// Simulate auth failure (permanent error)
		return utils.NewPermanentError(fmt.Errorf("invalid credentials"))
	})

	if err != nil {
		logger.Error("Auth failed (not retried)", zap.Error(err))
	}
}

// Example 4: Context cancellation
func exampleContextCancellation() {
	logger, _ := zap.NewDevelopment()

	// Short timeout to demonstrate cancellation
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cfg := utils.RetryConfig{
		MaxAttempts:     100, // High number, but context will cancel first
		InitialInterval: 1 * time.Second,
		MaxInterval:     5 * time.Second,
		Multiplier:      2.0,
		Jitter:          false,
	}

	err := utils.DoWithRetry(ctx, logger, "Slow Service", cfg, func() error {
		return fmt.Errorf("service unavailable")
	})

	if err != nil {
		logger.Error("Operation cancelled", zap.Error(err))
		// Output: "Slow Service: context cancelled after X attempts: context deadline exceeded"
	}
}

// Example 5: Monitoring retry attempts
func exampleMonitoringRetries() {
	logger, _ := zap.NewDevelopment()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg := utils.DefaultRetryConfig()
	attemptCount := 0

	err := utils.DoWithRetry(ctx, logger, "Flaky Service", cfg, func() error {
		attemptCount++

		// Simulate service that fails first 2 times, then succeeds
		if attemptCount < 3 {
			return fmt.Errorf("temporary failure (attempt %d)", attemptCount)
		}

		fmt.Printf("✅ Success on attempt %d\n", attemptCount)
		return nil
	})

	if err != nil {
		logger.Error("Failed", zap.Error(err))
	} else {
		logger.Info("Succeeded", zap.Int("total_attempts", attemptCount))
	}
}

// Example 6: Exponential backoff visualization
func exampleBackoffVisualization() {
	cfg := utils.RetryConfig{
		MaxAttempts:     5,
		InitialInterval: 1 * time.Second,
		MaxInterval:     30 * time.Second,
		Multiplier:      2.0,
		Jitter:          false, // Disable jitter for predictable output
	}

	fmt.Println("Backoff intervals:")
	interval := cfg.InitialInterval
	for i := 1; i <= cfg.MaxAttempts; i++ {
		fmt.Printf("Attempt %d: wait %v before retry\n", i, interval)

		// Calculate next interval (same logic as in retry.go)
		interval = time.Duration(float64(interval) * cfg.Multiplier)
		if interval > cfg.MaxInterval {
			interval = cfg.MaxInterval
		}
	}

	// Output:
	// Attempt 1: wait 1s before retry
	// Attempt 2: wait 2s before retry
	// Attempt 3: wait 4s before retry
	// Attempt 4: wait 8s before retry
	// Attempt 5: wait 16s before retry
}

func main() {
	fmt.Println("=== Example 1: Basic Retry ===")
	exampleBasicRetry()

	fmt.Println("\n=== Example 2: Custom Retry Config ===")
	exampleCustomRetry()

	fmt.Println("\n=== Example 3: Permanent Error ===")
	examplePermanentError()

	fmt.Println("\n=== Example 4: Context Cancellation ===")
	exampleContextCancellation()

	fmt.Println("\n=== Example 5: Monitoring Retries ===")
	exampleMonitoringRetries()

	fmt.Println("\n=== Example 6: Backoff Visualization ===")
	exampleBackoffVisualization()
}
