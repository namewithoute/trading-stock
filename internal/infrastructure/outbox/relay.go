package outbox

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	pollInterval = 200 * time.Millisecond
	batchSize    = 50
)

// OutboxRelay polls the outbox_events table and forwards unpublished rows to
// Kafka, then marks them as processed. It runs as a long-lived background
// goroutine and stops cleanly when ctx is cancelled.
type OutboxRelay struct {
	db        *gorm.DB
	publisher *kafka.Writer
	logger    *zap.Logger
}

// NewOutboxRelay constructs a relay. publisher must be a *kafka.Writer that
// already has brokers configured; the topic is taken from each row's Topic field.
func NewOutboxRelay(db *gorm.DB, publisher *kafka.Writer, logger *zap.Logger) *OutboxRelay {
	return &OutboxRelay{db: db, publisher: publisher, logger: logger}
}

// Run blocks until ctx is cancelled, polling and flushing the outbox.
func (r *OutboxRelay) Run(ctx context.Context) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	r.logger.Info("[ OutboxRelay ] started")
	for {
		select {
		case <-ctx.Done():
			r.logger.Info("[ OutboxRelay ] stopped")
			return
		case <-ticker.C:
			if err := r.flush(ctx); err != nil {
				r.logger.Error("[ OutboxRelay ] flush error", zap.Error(err))
			}
		}
	}
}

// flush picks up to batchSize unprocessed rows, publishes them, and marks them done.
func (r *OutboxRelay) flush(ctx context.Context) error {
	var rows []OutboxEventModel
	if err := r.db.WithContext(ctx).
		Where("processed_at IS NULL").
		Order("created_at ASC").
		Limit(batchSize).
		Find(&rows).Error; err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}

	msgs := make([]kafka.Message, 0, len(rows))
	for _, row := range rows {
		msgs = append(msgs, kafka.Message{
			Topic: row.Topic,
			Key:   []byte(row.MessageKey),
			Value: row.Payload,
		})
	}

	if err := r.publisher.WriteMessages(ctx, msgs...); err != nil {
		return err
	}

	// Mark processed in bulk
	ids := make([]string, len(rows))
	for i, row := range rows {
		ids[i] = row.ID
	}
	now := time.Now().UTC()
	return r.db.WithContext(ctx).
		Model(&OutboxEventModel{}).
		Where("id IN ?", ids).
		Update("processed_at", now).Error
}

// InsertOutboxEvent inserts a single outbox row inside the provided DB transaction tx.
// Call this within the same tx.Transaction() block as your domain event append.
func InsertOutboxEvent(tx *gorm.DB, id, topic, key string, payload []byte) error {
	row := OutboxEventModel{
		ID:         id,
		Topic:      topic,
		MessageKey: key,
		Payload:    payload,
		CreatedAt:  time.Now().UTC(),
	}
	return tx.Create(&row).Error
}
