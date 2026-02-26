package matching

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	domainExecution "trading-stock/internal/domain/execution"
	domainOrder "trading-stock/internal/domain/order"
	"trading-stock/internal/infrastructure/engine"
	infraEvents "trading-stock/internal/infrastructure/events"
	infraExecution "trading-stock/internal/infrastructure/execution"
	infraOutbox "trading-stock/internal/infrastructure/outbox"
	infraOrder "trading-stock/internal/infrastructure/order"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// KafkaTopicTradesExecuted is the Kafka topic for executed trades.
const KafkaTopicTradesExecuted = "trades.executed"

// MatchingService bridges Kafka consumer → MatchingEngine → persistence + outbox.
type MatchingService struct {
	engine    *engine.MatchingEngine
	db        *gorm.DB
	logger    *zap.Logger
}

// NewMatchingService constructs the service.
func NewMatchingService(
	eng *engine.MatchingEngine,
	db *gorm.DB,
	logger *zap.Logger,
) *MatchingService {
	return &MatchingService{
		engine: eng,
		db:     db,
		logger: logger,
	}
}

// HandleOrderAccepted is called for each orders.accepted message.
// It submits the order to the matching engine, persists each generated trade,
// and writes outbox rows for trades.executed — all in one DB transaction.
func (s *MatchingService) HandleOrderAccepted(ctx context.Context, msg infraOrder.OrderAcceptedMessage) error {
	// ── Adapt → legacy *order.Order for the engine ────────────────────────
	o := &domainOrder.Order{
		ID:        msg.OrderID,
		UserID:    msg.UserID,
		AccountID: msg.AccountID,
		Symbol:    msg.Symbol,
		Side:      msg.Side,
		Type:      msg.OrderType,
		Price:     msg.Price,
		Quantity:  msg.Quantity,
		Status:    domainOrder.StatusPending,
		CreatedAt: msg.OccurredAt,
		UpdatedAt: msg.OccurredAt,
	}

	// ── Submit to matching engine ──────────────────────────────────────────
	trades, err := s.engine.SubmitOrder(ctx, o)
	if err != nil {
		return fmt.Errorf("MatchingEngine.SubmitOrder(%s): %w", msg.OrderID, err)
	}

	if len(trades) == 0 {
		s.logger.Info("[ Matching ] order queued, no immediate fill",
			zap.String("order_id", msg.OrderID),
			zap.String("symbol", msg.Symbol),
		)
		return nil
	}

	// ── Persist trades + outbox in one transaction ─────────────────────────
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, t := range trades {
			tradeID := t.ID
			if tradeID == "" {
				tradeID = uuid.New().String()
			}
			now := time.Now().UTC()

			// 1. Persist trade record
			tr := &domainExecution.Trade{
				ID:          tradeID,
				BuyOrderID:  t.BuyOrderID,
				SellOrderID: t.SellOrderID,
				Symbol:      t.Symbol,
				Price:       t.Price,
				Quantity:    t.Quantity,
				BuyerID:     t.BuyerID,
				SellerID:    t.SellerID,
				Status:      domainExecution.TradeStatusPending,
				CreatedAt:   now,
			}

			if err := infraExecution.SaveTradeWithTx(tx, tr); err != nil {
				return fmt.Errorf("persist trade %s: %w", tradeID, err)
			}

			// 2. Write outbox entry for trades.executed
			execMsg := infraEvents.TradeExecutedMessage{
				EventID:     uuid.New().String(),
				TradeID:     tradeID,
				Symbol:      t.Symbol,
				Price:       t.Price,
				Quantity:    t.Quantity,
				BuyOrderID:  t.BuyOrderID,
				SellOrderID: t.SellOrderID,
				BuyerID:     t.BuyerID,
				SellerID:    t.SellerID,
				OccurredAt:  now,
			}
			payload, err := json.Marshal(execMsg)
			if err != nil {
				return fmt.Errorf("marshal TradeExecutedMessage: %w", err)
			}

			if err := infraOutbox.InsertOutboxEvent(tx, execMsg.EventID, KafkaTopicTradesExecuted, tradeID, payload); err != nil {
				return fmt.Errorf("insert outbox trades.executed: %w", err)
			}

			s.logger.Info("[ Matching ] trade executed and persisted",
				zap.String("trade_id", tradeID),
				zap.String("symbol", t.Symbol),
				zap.Float64("price", t.Price),
				zap.Int("quantity", t.Quantity),
			)
		}
		return nil
	})
}

// ─── MatchingConsumer ─────────────────────────────────────────────────────────

// MatchingConsumer is a long-running Kafka consumer that reads orders.accepted
// messages and forwards them to MatchingService.
type MatchingConsumer struct {
	service *MatchingService
	reader  *kafka.Reader
	logger  *zap.Logger
}

// NewMatchingConsumer constructs the consumer.
func NewMatchingConsumer(brokers []string, service *MatchingService, logger *zap.Logger) *MatchingConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    infraOrder.KafkaTopicOrdersAccepted,
		GroupID:  "matching-engine-consumer",
		MinBytes: 1,
		MaxBytes: 1 << 20, // 1 MB
	})
	return &MatchingConsumer{service: service, reader: reader, logger: logger}
}

// Run consumes messages until ctx is cancelled.
func (c *MatchingConsumer) Run(ctx context.Context) {
	c.logger.Info("[ MatchingConsumer ] started, listening on " + infraOrder.KafkaTopicOrdersAccepted)
	defer func() {
		_ = c.reader.Close()
		c.logger.Info("[ MatchingConsumer ] stopped")
	}()

	for {
		m, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.logger.Error("[ MatchingConsumer ] read error", zap.Error(err))
			continue
		}

		var msg infraOrder.OrderAcceptedMessage
		if err := json.Unmarshal(m.Value, &msg); err != nil {
			c.logger.Error("[ MatchingConsumer ] unmarshal error",
				zap.Error(err),
				zap.ByteString("value", m.Value),
			)
			continue
		}

		if err := c.service.HandleOrderAccepted(ctx, msg); err != nil {
			c.logger.Error("[ MatchingConsumer ] HandleOrderAccepted error",
				zap.Error(err),
				zap.String("order_id", msg.OrderID),
			)
		}
	}
}
