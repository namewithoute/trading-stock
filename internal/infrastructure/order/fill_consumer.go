package order

import (
	"context"
	"encoding/json"

	domain "trading-stock/internal/domain/order"
	infraEvents "trading-stock/internal/infrastructure/events"

	"github.com/cockroachdb/apd/v3"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// OrderFillConsumer consumes trades.executed messages and updates the order
// aggregate via RecordFill, keeping the order read model up to date.
type OrderFillConsumer struct {
	repo   domain.Repository
	reader *kafka.Reader
	logger *zap.Logger
}

// NewOrderFillConsumer constructs the consumer.
func NewOrderFillConsumer(brokers []string, repo domain.Repository, logger *zap.Logger) *OrderFillConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    infraEvents.KafkaTopicTradesExecuted,
		GroupID:  "order-fill-consumer",
		MinBytes: 1,
		MaxBytes: 1 << 20,
	})
	return &OrderFillConsumer{repo: repo, reader: reader, logger: logger}
}

// Run consumes messages until ctx is cancelled.
func (c *OrderFillConsumer) Run(ctx context.Context) {
	c.logger.Info("[ OrderFillConsumer ] started, listening on " + infraEvents.KafkaTopicTradesExecuted)
	defer func() {
		_ = c.reader.Close()
		c.logger.Info("[ OrderFillConsumer ] stopped")
	}()

	for {
		m, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.logger.Error("[ OrderFillConsumer ] read error", zap.Error(err))
			continue
		}

		var msg infraEvents.TradeExecutedMessage
		if err := json.Unmarshal(m.Value, &msg); err != nil {
			c.logger.Error("[ OrderFillConsumer ] unmarshal error", zap.Error(err))
			continue
		}

		// Update both sides of the trade
		c.recordFill(ctx, msg.BuyOrderID, msg.Quantity, msg.Price.Decimal)
		c.recordFill(ctx, msg.SellOrderID, msg.Quantity, msg.Price.Decimal)
	}
}

// recordFill loads the aggregate, calls RecordFill, then saves.
func (c *OrderFillConsumer) recordFill(ctx context.Context, orderID string, qty int, price apd.Decimal) {
	agg, err := c.repo.Load(ctx, orderID)
	if err != nil {
		c.logger.Error("[ OrderFillConsumer ] Load order failed",
			zap.String("order_id", orderID), zap.Error(err))
		return
	}

	if err := agg.RecordFill(qty, price); err != nil {
		c.logger.Warn("[ OrderFillConsumer ] RecordFill skipped",
			zap.String("order_id", orderID),
			zap.String("status", string(agg.Status)),
			zap.Error(err))
		return
	}

	if err := c.repo.Save(ctx, agg); err != nil {
		c.logger.Error("[ OrderFillConsumer ] Save order failed",
			zap.String("order_id", orderID), zap.Error(err))
	}
}
