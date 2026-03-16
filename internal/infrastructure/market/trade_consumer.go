package market

import (
	"context"
	"time"

	domainMarket "trading-stock/internal/domain/market"
	infraEvents "trading-stock/internal/infrastructure/events"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// MarketTradeConsumer listens to trades.executed and updates per-symbol price and
// OHLCV candle data in the market data store.
type MarketTradeConsumer struct {
	priceRepo  domainMarket.PriceRepository
	candleRepo domainMarket.CandleRepository
	reader     *kafka.Reader
	logger     *zap.Logger
}

// NewMarketTradeConsumer constructs the market data trade consumer.
func NewMarketTradeConsumer(
	brokers []string,
	priceRepo domainMarket.PriceRepository,
	candleRepo domainMarket.CandleRepository,
	logger *zap.Logger,
) *MarketTradeConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    infraEvents.KafkaTopicTradesExecuted,
		GroupID:  "market-data-consumer",
		MinBytes: 1,
		MaxBytes: 1 << 20,
	})
	return &MarketTradeConsumer{
		priceRepo:  priceRepo,
		candleRepo: candleRepo,
		reader:     reader,
		logger:     logger,
	}
}

// Run consumes messages until ctx is cancelled.
func (c *MarketTradeConsumer) Run(ctx context.Context) {
	c.logger.Info("[ MarketTradeConsumer ] started, listening on " + infraEvents.KafkaTopicTradesExecuted)
	defer func() {
		_ = c.reader.Close()
		c.logger.Info("[ MarketTradeConsumer ] stopped")
	}()

	for {
		m, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.logger.Error("[ MarketTradeConsumer ] read error", zap.Error(err))
			continue
		}

		var msg infraEvents.TradeExecutedMessage
		if err := infraEvents.DecodeKafkaPayload(m.Value, &msg); err != nil {
			c.logger.Error("[ MarketTradeConsumer ] unmarshal error", zap.Error(err))
			continue
		}

		if err := c.handleTrade(ctx, msg); err != nil {
			c.logger.Error("[ MarketTradeConsumer ] handleTrade error",
				zap.String("trade_id", msg.TradeID), zap.Error(err))
		}
	}
}

// handleTrade upserts a price record and updates the 1-minute candle.
func (c *MarketTradeConsumer) handleTrade(ctx context.Context, msg infraEvents.TradeExecutedMessage) error {
	now := msg.OccurredAt
	if now.IsZero() {
		now = time.Now().UTC()
	}

	// ── 1. Append price tick ────────────────────────────────────────────────
	price := &domainMarket.Price{
		ID:        uuid.New().String(),
		Symbol:    msg.Symbol,
		Price:     msg.Price.Decimal,
		Volume:    int64(msg.Quantity),
		Timestamp: now,
	}
	if err := c.priceRepo.Create(ctx, price); err != nil {
		return err
	}

	// ── 2. Upsert 1-minute OHLCV candle ────────────────────────────────────
	interval := "1m"
	// Truncate to minute boundary
	minuteStart := now.Truncate(time.Minute)

	from := minuteStart
	to := minuteStart.Add(time.Minute)
	candles, err := c.candleRepo.GetBySymbolAndInterval(ctx, msg.Symbol, interval, from, to)
	if err != nil {
		return err
	}

	if len(candles) == 0 {
		// First trade in this minute — open a new candle
		candle := &domainMarket.Candle{
			ID:        uuid.New().String(),
			Symbol:    msg.Symbol,
			Interval:  interval,
			Open:      msg.Price.Decimal,
			High:      msg.Price.Decimal,
			Low:       msg.Price.Decimal,
			Close:     msg.Price.Decimal,
			Volume:    int64(msg.Quantity),
			Timestamp: minuteStart,
		}
		if err := c.candleRepo.Create(ctx, candle); err != nil {
			return err
		}
	} else {
		// Update existing candle
		existing := candles[len(candles)-1]
		if msg.Price.Decimal.Cmp(&existing.High) > 0 {
			existing.High = msg.Price.Decimal
		}
		if msg.Price.Decimal.Cmp(&existing.Low) < 0 {
			existing.Low = msg.Price.Decimal
		}
		existing.Close = msg.Price.Decimal
		existing.Volume += int64(msg.Quantity)

		// Repository doesn't expose Update; re-create via BatchCreate with the updated record
		// We delete-and-reinsert approach is avoided since the candle table may use
		// a unique constraint on (symbol, interval, timestamp).
		// Instead, do a direct GORM save through the concrete priceRepo's db.
		// We expose a helper in model.go to allow updates.
		if err := c.candleRepo.Update(ctx, existing); err != nil {
			return err
		}
	}

	c.logger.Debug("[ MarketTradeConsumer ] price + candle updated",
		zap.String("symbol", msg.Symbol),
		zap.String("price", msg.Price.Decimal.String()),
		zap.Int("qty", msg.Quantity),
	)
	return nil
}
