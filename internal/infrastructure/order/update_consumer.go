package order

import (
	"context"

	domain "trading-stock/internal/domain/order"
	infraEvents "trading-stock/internal/infrastructure/events"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// OrderUpdatedConsumer consumes trading.orders.updated and applies status/fill
// progress to the order aggregate.
type OrderUpdatedConsumer struct {
	repo   domain.Repository
	reader *kafka.Reader
	logger *zap.Logger
}

// NewOrderUpdatedConsumer constructs the consumer.
func NewOrderUpdatedConsumer(brokers []string, repo domain.Repository, logger *zap.Logger) *OrderUpdatedConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    infraEvents.KafkaTopicOrdersUpdated,
		GroupID:  "order-updated-consumer",
		MinBytes: 1,
		MaxBytes: 1 << 20,
	})
	return &OrderUpdatedConsumer{repo: repo, reader: reader, logger: logger}
}

// Run consumes messages until ctx is cancelled.
func (c *OrderUpdatedConsumer) Run(ctx context.Context) {
	c.logger.Info("[ OrderUpdatedConsumer ] started, listening on " + infraEvents.KafkaTopicOrdersUpdated)
	defer func() {
		_ = c.reader.Close()
		c.logger.Info("[ OrderUpdatedConsumer ] stopped")
	}()

	for {
		m, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.logger.Error("[ OrderUpdatedConsumer ] read error", zap.Error(err))
			continue
		}

		var msg infraEvents.OrderUpdatedMessage
		if err := infraEvents.DecodeKafkaPayload(m.Value, &msg); err != nil {
			c.logger.Error("[ OrderUpdatedConsumer ] unmarshal error", zap.Error(err))
			continue
		}

		c.applyUpdate(ctx, msg)
	}
}

func (c *OrderUpdatedConsumer) applyUpdate(ctx context.Context, msg infraEvents.OrderUpdatedMessage) {
	agg, err := c.repo.Load(ctx, msg.OrderID)
	if err != nil {
		c.logger.Error("[ OrderUpdatedConsumer ] Load order failed",
			zap.String("order_id", msg.OrderID), zap.Error(err))
		return
	}

	changed := false

	if msg.FilledQuantity > agg.FilledQuantity {
		delta := msg.FilledQuantity - agg.FilledQuantity
		if err := agg.RecordFill(delta, msg.AvgFillPrice.Decimal); err != nil {
			c.logger.Warn("[ OrderUpdatedConsumer ] RecordFill skipped",
				zap.String("order_id", msg.OrderID),
				zap.Int("delta", delta),
				zap.Error(err))
		} else {
			changed = true
		}
	}

	if msg.Status == domain.StatusExpired && !agg.Status.IsFinal() {
		if err := agg.Expire(); err != nil {
			c.logger.Warn("[ OrderUpdatedConsumer ] Expire skipped",
				zap.String("order_id", msg.OrderID), zap.Error(err))
		} else {
			changed = true
		}
	}

	if !changed {
		return
	}

	if err := c.repo.Save(ctx, agg); err != nil {
		c.logger.Error("[ OrderUpdatedConsumer ] Save order failed",
			zap.String("order_id", msg.OrderID), zap.Error(err))
	}
}
