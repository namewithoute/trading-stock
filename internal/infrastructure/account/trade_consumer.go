package account

import (
	"context"
	"encoding/json"
	"fmt"

	domainAccount "trading-stock/internal/domain/account"
	infraEvents "trading-stock/internal/infrastructure/events"

	"github.com/cockroachdb/apd/v3"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

var tradeDecCtx = apd.BaseContext.WithPrecision(19)

// TradeConsumer listens to trades.executed and settles funds in the account
// aggregate for both buyer and seller.
type TradeConsumer struct {
	repo          domainAccount.Repository
	readModelRepo domainAccount.ReadModelRepository
	reader        *kafka.Reader
	logger        *zap.Logger
}

// NewTradeConsumer constructs the consumer.
func NewTradeConsumer(
	brokers []string,
	repo domainAccount.Repository,
	readModelRepo domainAccount.ReadModelRepository,
	logger *zap.Logger,
) *TradeConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    infraEvents.KafkaTopicTradesExecuted,
		GroupID:  "account-trade-consumer",
		MinBytes: 1,
		MaxBytes: 1 << 20,
	})
	return &TradeConsumer{
		repo:          repo,
		readModelRepo: readModelRepo,
		reader:        reader,
		logger:        logger,
	}
}

// Run consumes messages until ctx is cancelled.
func (c *TradeConsumer) Run(ctx context.Context) {
	c.logger.Info("[ AccountTradeConsumer ] started, listening on " + infraEvents.KafkaTopicTradesExecuted)
	defer func() {
		_ = c.reader.Close()
		c.logger.Info("[ AccountTradeConsumer ] stopped")
	}()

	for {
		m, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.logger.Error("[ AccountTradeConsumer ] read error", zap.Error(err))
			continue
		}

		var msg infraEvents.TradeExecutedMessage
		if err := json.Unmarshal(m.Value, &msg); err != nil {
			c.logger.Error("[ AccountTradeConsumer ] unmarshal error", zap.Error(err))
			continue
		}

		amount := apd.Decimal{}
		_, _ = tradeDecCtx.Mul(&amount, &msg.Price, apd.New(int64(msg.Quantity), 0))
		c.settle(ctx, msg.TradeID, msg.BuyerID, "BUY", amount)
		c.settle(ctx, msg.TradeID, msg.SellerID, "SELL", amount)
	}
}

// settle resolves the account ID for a user then calls SettleTrade.
func (c *TradeConsumer) settle(ctx context.Context, tradeID, userID, side string, amount apd.Decimal) {
	accounts, err := c.readModelRepo.GetByUserID(ctx, userID)
	if err != nil || len(accounts) == 0 {
		c.logger.Error("[ AccountTradeConsumer ] account not found for user",
			zap.String("user_id", userID), zap.Error(err))
		return
	}
	accountID := accounts[0].ID
	for _, a := range accounts {
		if string(a.Status) == "active" {
			accountID = a.ID
			break
		}
	}

	agg, err := c.repo.Load(ctx, accountID)
	if err != nil {
		c.logger.Error("[ AccountTradeConsumer ] Load account failed",
			zap.String("account_id", accountID), zap.Error(err))
		return
	}

	if err := agg.SettleTrade(tradeID, side, amount); err != nil {
		c.logger.Error("[ AccountTradeConsumer ] SettleTrade failed",
			zap.String("account_id", accountID), zap.String("side", side), zap.Error(err))
		return
	}

	if err := c.repo.Save(ctx, agg); err != nil {
		c.logger.Error("[ AccountTradeConsumer ] Save account failed",
			zap.String("account_id", accountID), zap.Error(err))
		return
	}

	c.logger.Info(fmt.Sprintf("[ AccountTradeConsumer ] %s settled", side),
		zap.String("trade_id", tradeID),
		zap.String("account_id", accountID),
		zap.String("amount", amount.String()),
	)
}
