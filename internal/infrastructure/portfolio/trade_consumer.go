package portfolio

import (
	"context"
	"encoding/json"

	domain "trading-stock/internal/domain/portfolio"
	infraEvents "trading-stock/internal/infrastructure/events"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// TradeConsumer listens to trades.executed and updates portfolio positions
// for both buyer (add shares) and seller (reduce shares).
type TradeConsumer struct {
	portfolioRepo domain.Repository
	reader        *kafka.Reader
	logger        *zap.Logger
}

// NewTradeConsumer constructs the consumer.
func NewTradeConsumer(
	brokers []string,
	portfolioRepo domain.Repository,
	logger *zap.Logger,
) *TradeConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    infraEvents.KafkaTopicTradesExecuted,
		GroupID:  "portfolio-trade-consumer",
		MinBytes: 1,
		MaxBytes: 1 << 20,
	})
	return &TradeConsumer{
		portfolioRepo: portfolioRepo,
		reader:        reader,
		logger:        logger,
	}
}

// Run consumes messages until ctx is cancelled.
func (c *TradeConsumer) Run(ctx context.Context) {
	c.logger.Info("[ PortfolioTradeConsumer ] started, listening on " + infraEvents.KafkaTopicTradesExecuted)
	defer func() {
		_ = c.reader.Close()
		c.logger.Info("[ PortfolioTradeConsumer ] stopped")
	}()

	for {
		m, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.logger.Error("[ PortfolioTradeConsumer ] read error", zap.Error(err))
			continue
		}

		var msg infraEvents.TradeExecutedMessage
		if err := json.Unmarshal(m.Value, &msg); err != nil {
			c.logger.Error("[ PortfolioTradeConsumer ] unmarshal error", zap.Error(err))
			continue
		}

		priceFloat, _ := msg.Price.Decimal.Float64()

		// Buyer gains shares
		c.updatePosition(ctx, msg.BuyerID, msg.Symbol, msg.Quantity, priceFloat, "BUY")

		// Seller loses shares
		c.updatePosition(ctx, msg.SellerID, msg.Symbol, msg.Quantity, priceFloat, "SELL")
	}
}

// updatePosition creates or updates a portfolio position for a given user+symbol.
func (c *TradeConsumer) updatePosition(ctx context.Context, userID, symbol string, quantity int, price float64, side string) {
	// Find existing position by user's account — we look up by userID across all accounts.
	positions, err := c.portfolioRepo.ListByUserID(ctx, userID)
	if err != nil {
		c.logger.Error("[ PortfolioTradeConsumer ] ListByUserID failed",
			zap.String("user_id", userID), zap.Error(err))
		return
	}

	// Find an existing position for this symbol.
	var existing *domain.Position
	for _, p := range positions {
		if p.Symbol == symbol {
			existing = p
			break
		}
	}

	if side == "BUY" {
		if existing != nil {
			existing.AddQuantity(quantity, price)
			if err := c.portfolioRepo.Update(ctx, existing); err != nil {
				c.logger.Error("[ PortfolioTradeConsumer ] Update position failed",
					zap.String("position_id", existing.ID), zap.Error(err))
				return
			}
		} else {
			pos := &domain.Position{
				ID:           uuid.New().String(),
				UserID:       userID,
				Symbol:       symbol,
				Quantity:     quantity,
				AvgPrice:     price,
				CurrentPrice: price,
			}
			pos.CalculateUnrealizedPnL()
			if err := c.portfolioRepo.Create(ctx, pos); err != nil {
				c.logger.Error("[ PortfolioTradeConsumer ] Create position failed",
					zap.String("user_id", userID), zap.String("symbol", symbol), zap.Error(err))
				return
			}
		}
		c.logger.Info("[ PortfolioTradeConsumer ] BUY position updated",
			zap.String("user_id", userID), zap.String("symbol", symbol), zap.Int("qty", quantity))
	} else {
		// SELL
		if existing == nil {
			c.logger.Warn("[ PortfolioTradeConsumer ] SELL but no existing position",
				zap.String("user_id", userID), zap.String("symbol", symbol))
			return
		}
		if err := existing.ReduceQuantity(quantity); err != nil {
			c.logger.Error("[ PortfolioTradeConsumer ] ReduceQuantity failed",
				zap.String("position_id", existing.ID), zap.Error(err))
			return
		}
		if existing.IsClosed() {
			if err := c.portfolioRepo.Delete(ctx, existing.ID); err != nil {
				c.logger.Error("[ PortfolioTradeConsumer ] Delete closed position failed",
					zap.String("position_id", existing.ID), zap.Error(err))
			}
		} else {
			if err := c.portfolioRepo.Update(ctx, existing); err != nil {
				c.logger.Error("[ PortfolioTradeConsumer ] Update position failed",
					zap.String("position_id", existing.ID), zap.Error(err))
			}
		}
		c.logger.Info("[ PortfolioTradeConsumer ] SELL position updated",
			zap.String("user_id", userID), zap.String("symbol", symbol), zap.Int("qty", quantity))
	}
}
