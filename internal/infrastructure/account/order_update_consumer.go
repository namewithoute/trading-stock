package account

import (
	"context"

	domainAccount "trading-stock/internal/domain/account"
	domainOrder "trading-stock/internal/domain/order"
	infraEvents "trading-stock/internal/infrastructure/events"

	"github.com/cockroachdb/apd/v3"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

var orderUpdateDecCtx = apd.BaseContext.WithPrecision(19)

// OrderUpdatedConsumer consumes trading.orders.updated and releases unused BUY
// reserved funds when an order reaches a final state.
type OrderUpdatedConsumer struct {
	repo   domainAccount.Repository
	reader *kafka.Reader
	logger *zap.Logger
}

// NewOrderUpdatedConsumer constructs the consumer.
func NewOrderUpdatedConsumer(brokers []string, repo domainAccount.Repository, logger *zap.Logger) *OrderUpdatedConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    infraEvents.KafkaTopicOrdersUpdated,
		GroupID:  "account-order-updated-consumer",
		MinBytes: 1,
		MaxBytes: 1 << 20,
	})
	return &OrderUpdatedConsumer{repo: repo, reader: reader, logger: logger}
}

// Run consumes messages until ctx is cancelled.
func (c *OrderUpdatedConsumer) Run(ctx context.Context) {
	c.logger.Info("[ AccountOrderUpdatedConsumer ] started, listening on " + infraEvents.KafkaTopicOrdersUpdated)
	defer func() {
		_ = c.reader.Close()
		c.logger.Info("[ AccountOrderUpdatedConsumer ] stopped")
	}()

	for {
		m, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.logger.Error("[ AccountOrderUpdatedConsumer ] read error", zap.Error(err))
			continue
		}

		var msg infraEvents.OrderUpdatedMessage
		if err := infraEvents.DecodeKafkaPayload(m.Value, &msg); err != nil {
			c.logger.Error("[ AccountOrderUpdatedConsumer ] unmarshal error", zap.Error(err))
			continue
		}

		c.handleFinalBuyOrder(ctx, msg)
	}
}

func (c *OrderUpdatedConsumer) handleFinalBuyOrder(ctx context.Context, msg infraEvents.OrderUpdatedMessage) {
	if msg.Side != domainOrder.SideBuy {
		return
	}
	if msg.AccountID == "" {
		return
	}
	if !msg.Status.IsFinal() {
		return
	}

	agg, err := c.repo.Load(ctx, msg.AccountID)
	if err != nil {
		c.logger.Error("[ AccountOrderUpdatedConsumer ] Load account failed",
			zap.String("account_id", msg.AccountID), zap.Error(err))
		return
	}

	reserved := apd.Decimal{}
	_, _ = orderUpdateDecCtx.Mul(&reserved, &msg.Price.Decimal, apd.New(int64(msg.Quantity), 0))

	executed := apd.Decimal{}
	_, _ = orderUpdateDecCtx.Mul(&executed, &msg.AvgFillPrice.Decimal, apd.New(int64(msg.FilledQuantity), 0))

	release := apd.Decimal{}
	_, _ = orderUpdateDecCtx.Sub(&release, &reserved, &executed)
	if release.Sign() <= 0 {
		return
	}

	if err := agg.ReleaseFunds(release); err != nil {
		c.logger.Warn("[ AccountOrderUpdatedConsumer ] ReleaseFunds skipped",
			zap.String("account_id", msg.AccountID),
			zap.String("order_id", msg.OrderID),
			zap.Error(err))
		return
	}

	if err := c.repo.Save(ctx, agg); err != nil {
		c.logger.Error("[ AccountOrderUpdatedConsumer ] Save account failed",
			zap.String("account_id", msg.AccountID),
			zap.String("order_id", msg.OrderID),
			zap.Error(err))
		return
	}

	c.logger.Info("[ AccountOrderUpdatedConsumer ] released BUY reserve remainder",
		zap.String("account_id", msg.AccountID),
		zap.String("order_id", msg.OrderID),
		zap.String("status", string(msg.Status)),
		zap.String("release_amount", release.String()))
}
