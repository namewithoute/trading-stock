package order

import (
	"context"
	"time"

	domain "trading-stock/internal/domain/order"
	infraEvents "trading-stock/internal/infrastructure/events"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// MarketExpireConsumer consumes orders.market_expired messages and expires the
// order aggregate for market orders whose unfilled remainder must be closed.
//
// Race-condition handling: the outbox event is written in the same transaction
// as the trade events, but this consumer may run before all fills have been
// applied by the OrderFillConsumer. A bounded retry loop waits until the
// aggregate's FilledQuantity matches the expected value from the engine.
type MarketExpireConsumer struct {
	repo   domain.Repository
	reader *kafka.Reader
	logger *zap.Logger
}

// NewMarketExpireConsumer constructs the consumer.
func NewMarketExpireConsumer(brokers []string, repo domain.Repository, logger *zap.Logger) *MarketExpireConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    infraEvents.KafkaTopicOrdersMarketExpired,
		GroupID:  "market-expire-consumer",
		MinBytes: 1,
		MaxBytes: 1 << 20,
	})
	return &MarketExpireConsumer{repo: repo, reader: reader, logger: logger}
}

// Run consumes messages until ctx is cancelled.
func (c *MarketExpireConsumer) Run(ctx context.Context) {
	c.logger.Info("[ MarketExpireConsumer ] started, listening on " + infraEvents.KafkaTopicOrdersMarketExpired)
	defer func() {
		_ = c.reader.Close()
		c.logger.Info("[ MarketExpireConsumer ] stopped")
	}()

	for {
		m, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.logger.Error("[ MarketExpireConsumer ] read error", zap.Error(err))
			continue
		}

		var msg infraEvents.MarketOrderExpiredMessage
		if err := infraEvents.DecodeKafkaPayload(m.Value, &msg); err != nil {
			c.logger.Error("[ MarketExpireConsumer ] unmarshal error", zap.Error(err))
			continue
		}

		c.expireMarketOrder(ctx, msg)
	}
}

// expireMarketOrder loads the aggregate, waits for all fills to arrive, then expires.
func (c *MarketExpireConsumer) expireMarketOrder(ctx context.Context, msg infraEvents.MarketOrderExpiredMessage) {
	const maxRetries = 30
	const retryDelay = 100 * time.Millisecond

	for attempt := 0; attempt < maxRetries; attempt++ {
		agg, err := c.repo.Load(ctx, msg.OrderID)
		if err != nil {
			c.logger.Error("[ MarketExpireConsumer ] Load order failed",
				zap.String("order_id", msg.OrderID), zap.Error(err))
			return
		}

		// Already in a terminal state (filled, cancelled, expired) — nothing to do.
		if agg.Status.IsFinal() {
			c.logger.Debug("[ MarketExpireConsumer ] order already in terminal state",
				zap.String("order_id", msg.OrderID),
				zap.String("status", string(agg.Status)))
			return
		}

		// Wait until all fills from the engine have been applied by OrderFillConsumer.
		if agg.FilledQuantity < msg.FilledQuantity {
			if attempt < maxRetries-1 {
				select {
				case <-time.After(retryDelay):
					continue
				case <-ctx.Done():
					return
				}
			}
			c.logger.Warn("[ MarketExpireConsumer ] fills not yet applied after retries",
				zap.String("order_id", msg.OrderID),
				zap.Int("expected", msg.FilledQuantity),
				zap.Int("actual", agg.FilledQuantity))
			return
		}

		// All fills applied — expire the remainder.
		if err := agg.Expire(); err != nil {
			c.logger.Warn("[ MarketExpireConsumer ] Expire skipped",
				zap.String("order_id", msg.OrderID),
				zap.String("status", string(agg.Status)),
				zap.Error(err))
			return
		}

		if err := c.repo.Save(ctx, agg); err != nil {
			c.logger.Error("[ MarketExpireConsumer ] Save order failed",
				zap.String("order_id", msg.OrderID), zap.Error(err))
			return
		}

		c.logger.Info("[ MarketExpireConsumer ] market order remainder expired",
			zap.String("order_id", msg.OrderID),
			zap.Int("filled_quantity", agg.FilledQuantity),
			zap.Int("total_quantity", agg.Quantity))
		return
	}
}
