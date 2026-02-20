package persistence

import (
	"context"
	"fmt"
	"time"

	"trading-stock/internal/config"
	"trading-stock/pkg/utils"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// InitKafka initializes Kafka writer with retry logic
// Returns *kafka.Writer instance instead of using global variables
func InitKafka(ctx context.Context, cfg config.KafkaConfig, log *zap.Logger) (*kafka.Writer, error) {
	// 1. Test broker connectivity with retry
	retryCfg := utils.DefaultRetryConfig()
	err := utils.DoWithRetry(ctx, log, "Kafka Connection", retryCfg, func() error {
		for _, broker := range cfg.Brokers {
			conn, err := kafka.Dial("tcp", broker)
			if err == nil {
				conn.Close()
				log.Info("Successfully connected to Kafka broker", zap.String("broker", broker))
				return nil
			}
			log.Warn("Failed to connect to broker",
				zap.String("broker", broker),
				zap.Error(err),
			)
		}
		return fmt.Errorf("failed to connect to any Kafka broker")
	})

	if err != nil {
		return nil, fmt.Errorf("failed to initialize Kafka after retries: %w", err)
	}

	// 2. Create Kafka writer with optimized configuration
	writer := &kafka.Writer{
		Addr:     kafka.TCP(cfg.Brokers...),
		Balancer: &kafka.LeastBytes{}, // Smart load balancing based on message size

		// Performance tuning
		// Batch up to 100 messages or wait 10ms before sending
		BatchSize:    cfg.BatchSize,
		BatchTimeout: time.Duration(cfg.BatchTimeout) * time.Millisecond,

		// Compression for reduced network usage
		Compression: kafka.Snappy,

		// Timeout configuration
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,

		// Retry policy: 3 attempts with backoff
		MaxAttempts: 3,

		// Async writes for better performance
		Async: false, // Set to true for fire-and-forget (loses delivery guarantee)
	}

	log.Info("Kafka writer initialized successfully",
		zap.Strings("brokers", cfg.Brokers),
		zap.Int("batch_size", cfg.BatchSize),
		zap.Int("batch_timeout_ms", cfg.BatchTimeout),
	)

	return writer, nil
}

// CloseKafka closes the Kafka writer
// This will flush any pending messages before closing
func CloseKafka(writer *kafka.Writer, log *zap.Logger) error {
	if writer == nil {
		return nil
	}

	// Close() automatically flushes pending messages in the buffer
	if err := writer.Close(); err != nil {
		log.Error("Failed to close Kafka writer", zap.Error(err))
		return fmt.Errorf("failed to close Kafka: %w", err)
	}

	log.Info("Kafka writer closed successfully")
	return nil
}

// GetKafkaStats returns Kafka writer statistics
func GetKafkaStats(writer *kafka.Writer) map[string]interface{} {
	if writer == nil {
		return nil
	}

	stats := writer.Stats()
	return map[string]interface{}{
		"writes":         stats.Writes,
		"messages":       stats.Messages,
		"bytes":          stats.Bytes,
		"errors":         stats.Errors,
		"batch_time_avg": stats.BatchTime.Avg,
		"batch_time_max": stats.BatchTime.Max,
		"batch_size_avg": stats.BatchSize.Avg,
		"batch_size_max": stats.BatchSize.Max,
		"write_time_avg": stats.WriteTime.Avg,
		"write_time_max": stats.WriteTime.Max,
		"wait_time_avg":  stats.WaitTime.Avg,
		"wait_time_max":  stats.WaitTime.Max,
		"retries":        stats.Retries,
		"batch_timeout":  stats.BatchTimeout,
	}
}

// PublishMessage publishes a single message to Kafka
// This is a helper function for testing or simple use cases
func PublishMessage(ctx context.Context, writer *kafka.Writer, topic string, key, value []byte) error {
	if writer == nil {
		return fmt.Errorf("kafka writer is nil")
	}

	msg := kafka.Message{
		Topic: topic,
		Key:   key,
		Value: value,
		Time:  time.Now(),
	}

	if err := writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

// PublishMessages publishes multiple messages to Kafka in a batch
func PublishMessages(ctx context.Context, writer *kafka.Writer, messages []kafka.Message) error {
	if writer == nil {
		return fmt.Errorf("kafka writer is nil")
	}

	if len(messages) == 0 {
		return nil
	}

	if err := writer.WriteMessages(ctx, messages...); err != nil {
		return fmt.Errorf("failed to publish messages: %w", err)
	}

	return nil
}
