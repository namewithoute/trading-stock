package utils

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"

	"go.uber.org/zap"
)

// RetryConfig defines retry behavior configuration
type RetryConfig struct {
	MaxAttempts     int           `mapstructure:"max_attempts"`     // Maximum number of retry attempts (0 = infinite until context timeout)
	InitialInterval time.Duration `mapstructure:"initial_interval"` // Initial backoff interval (e.g., 1s)
	MaxInterval     time.Duration `mapstructure:"max_interval"`     // Maximum backoff interval (e.g., 30s)
	Multiplier      float64       `mapstructure:"multiplier"`       // Backoff multiplier (e.g., 2.0 for exponential)
	Jitter          bool          `mapstructure:"jitter"`           // Add randomness to prevent thundering herd
}

// DefaultRetryConfig returns production-ready retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:     10,
		InitialInterval: 1 * time.Second,
		MaxInterval:     30 * time.Second,
		Multiplier:      2.0,
		Jitter:          true,
	}
}

// IsRetryable determines if an error should be retried
// Permanent errors (e.g., auth failures, invalid config) should NOT be retried
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Add your non-retryable error types here
	var permanentErr *PermanentError
	if errors.As(err, &permanentErr) {
		return false
	}

	// By default, retry all errors (network issues, timeouts, etc.)
	return true
}

// PermanentError represents an error that should not be retried
type PermanentError struct {
	Err error
}

func (e *PermanentError) Error() string {
	return fmt.Sprintf("permanent error: %v", e.Err)
}

func (e *PermanentError) Unwrap() error {
	return e.Err
}

// NewPermanentError wraps an error as non-retryable
func NewPermanentError(err error) error {
	return &PermanentError{Err: err}
}

// DoWithRetry executes an operation with exponential backoff retry logic
func DoWithRetry(ctx context.Context, logger *zap.Logger, opName string, cfg RetryConfig, operation func() error) error {
	attempt := 0
	interval := cfg.InitialInterval

	for {
		attempt++

		// 1. Check context cancellation before attempting
		select {
		case <-ctx.Done():
			return fmt.Errorf("%s: context cancelled after %d attempts: %w", opName, attempt-1, ctx.Err())
		default:
		}

		// 2. Execute operation
		err := operation()
		if err == nil {
			if attempt > 1 {
				logger.Info(fmt.Sprintf("%s: succeeded after %d attempts", opName, attempt))
			} else {
				logger.Info(fmt.Sprintf("%s: connected successfully", opName))
			}
			return nil
		}

		// 3. Check if error is retryable
		if !IsRetryable(err) {
			logger.Error(fmt.Sprintf("%s: permanent error, not retrying", opName), zap.Error(err))
			return err
		}

		// 4. Check max attempts
		if cfg.MaxAttempts > 0 && attempt >= cfg.MaxAttempts {
			return fmt.Errorf("%s: max retry attempts (%d) exceeded: %w", opName, cfg.MaxAttempts, err)
		}

		// 5. Log retry attempt
		logger.Warn(
			fmt.Sprintf("%s: attempt %d failed, retrying in %v", opName, attempt, interval),
			zap.Error(err),
			zap.Int("attempt", attempt),
			zap.Duration("next_retry_in", interval),
		)

		// 6. Wait with backoff (or context cancellation)
		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return fmt.Errorf("%s: context cancelled after %d attempts: %w", opName, attempt, ctx.Err())
		case <-timer.C:
			// Continue to next attempt
		}

		// 7. Calculate next backoff interval (exponential with jitter)
		interval = calculateNextInterval(interval, cfg)
	}
}

// calculateNextInterval computes the next backoff duration with exponential growth and optional jitter
func calculateNextInterval(current time.Duration, cfg RetryConfig) time.Duration {
	// Exponential backoff
	next := time.Duration(float64(current) * cfg.Multiplier)

	// Cap at max interval
	if next > cfg.MaxInterval {
		next = cfg.MaxInterval
	}

	// Add jitter (±25% randomness) to prevent thundering herd
	if cfg.Jitter {
		jitterRange := float64(next) * 0.25
		jitterOffset := (rand.Float64() * 2 * jitterRange) - jitterRange
		next = time.Duration(math.Max(float64(cfg.InitialInterval), float64(next)+jitterOffset))
	}

	return next
}
