package initialize

import (
	"context"
	"time"
	"trading-stock/internal/config"

	"github.com/segmentio/kafka-go"
)

func InitKafka(cfg config.KafkaConfig) (*kafka.Writer, error) {
	// In Kafka, we usually don't have a single "connection" object like DB.
	// Instead, we verify if brokers are reachable.
	for _, broker := range cfg.Brokers {
		conn, err := kafka.DialContext(context.Background(), "tcp", broker)
		if err != nil {
			return nil, err
		}
		conn.Close()
	}
	return GetKafkaWriter(cfg, "orders"), nil
}

// GetKafkaWriter returns a new Kafka writer for a specific topic with optimization
func GetKafkaWriter(cfg config.KafkaConfig, topic string) *kafka.Writer {
	return &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    cfg.BatchSize,
		BatchTimeout: time.Duration(cfg.BatchTimeout) * time.Millisecond,
		Async:        true, // Send messages asynchronously for better performance
	}
}

// GetKafkaReader returns a new Kafka reader for a specific topic and group
func GetKafkaReader(cfg config.KafkaConfig, topic, groupID string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:  cfg.Brokers,
		GroupID:  groupID,
		Topic:    topic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})
}
