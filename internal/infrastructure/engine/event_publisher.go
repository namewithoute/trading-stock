package engine

import (
	"context"
	"encoding/json"
	"fmt"

	infraEvents "trading-stock/internal/infrastructure/events"
	pkgdecimal "trading-stock/pkg/decimal"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// EventPublisher publishes trading events to Kafka
type EventPublisher struct {
	writer *kafka.Writer
	logger *zap.Logger
}

// NewEventPublisher creates a new event publisher
func NewEventPublisher(writer *kafka.Writer, logger *zap.Logger) *EventPublisher {
	return &EventPublisher{
		writer: writer,
		logger: logger,
	}
}

// PublishTrade publishes a trade event to Kafka
func (ep *EventPublisher) PublishTrade(ctx context.Context, trade *Trade) error {
	// Convert trade to JSON
	tradeJSON, err := json.Marshal(trade)
	if err != nil {
		return fmt.Errorf("failed to marshal trade: %w", err)
	}

	// Create Kafka message
	msg := kafka.Message{
		Topic: "trading.trades.executed",
		Key:   []byte(trade.Symbol),
		Value: tradeJSON,
	}

	// Write to Kafka
	if err := ep.writer.WriteMessages(ctx, msg); err != nil {
		ep.logger.Error("Failed to publish trade to Kafka",
			zap.String("trade_id", trade.ID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to publish trade: %w", err)
	}

	ep.logger.Debug("Trade published to Kafka",
		zap.String("trade_id", trade.ID),
		zap.String("symbol", trade.Symbol),
	)

	return nil
}

// PublishOrderUpdate publishes an order update event to Kafka
func (ep *EventPublisher) PublishOrderUpdate(ctx context.Context, update *OrderUpdate) error {
	msgBody := infraEvents.OrderUpdatedMessage{
		EventID:        uuid.New().String(),
		OrderID:        update.OrderID,
		UserID:         update.UserID,
		AccountID:      update.AccountID,
		Symbol:         update.Symbol,
		Side:           update.Side,
		OrderType:      update.OrderType,
		Status:         update.Status,
		Quantity:       update.Quantity,
		FilledQuantity: update.FilledQuantity,
		Price:          pkgdecimal.From(update.Price),
		AvgFillPrice:   pkgdecimal.From(update.AvgFillPrice),
		OccurredAt:     update.Timestamp,
	}

	// Convert update to JSON
	updateJSON, err := json.Marshal(msgBody)
	if err != nil {
		return fmt.Errorf("failed to marshal order update: %w", err)
	}

	// Create Kafka message
	msg := kafka.Message{
		Topic: infraEvents.KafkaTopicOrdersUpdated,
		Key:   []byte(update.OrderID),
		Value: updateJSON,
	}

	// Write to Kafka
	if err := ep.writer.WriteMessages(ctx, msg); err != nil {
		ep.logger.Error("Failed to publish order update to Kafka",
			zap.String("order_id", update.OrderID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to publish order update: %w", err)
	}

	ep.logger.Debug("Order update published to Kafka",
		zap.String("order_id", update.OrderID),
		zap.String("status", string(update.Status)),
	)

	return nil
}

// StartEventConsumer starts consuming events from matching engine and publishes to Kafka
func (ep *EventPublisher) StartEventConsumer(ctx context.Context, engine *MatchingEngine) {
	go ep.consumeTrades(ctx, engine.TradeChannel())
	go ep.consumeOrderUpdates(ctx, engine.OrderUpdateChannel())
	ep.logger.Info("Event consumer started")
}

// consumeTrades consumes trades from the trade channel and publishes to Kafka
func (ep *EventPublisher) consumeTrades(ctx context.Context, tradeChan <-chan *Trade) {
	for {
		select {
		case <-ctx.Done():
			ep.logger.Info("Trade consumer stopped")
			return
		case trade, ok := <-tradeChan:
			if !ok {
				ep.logger.Info("Trade channel closed")
				return
			}
			if err := ep.PublishTrade(ctx, trade); err != nil {
				ep.logger.Error("Failed to publish trade", zap.Error(err))
			}
		}
	}
}

// consumeOrderUpdates consumes order updates and publishes to Kafka
func (ep *EventPublisher) consumeOrderUpdates(ctx context.Context, updateChan <-chan *OrderUpdate) {
	for {
		select {
		case <-ctx.Done():
			ep.logger.Info("Order update consumer stopped")
			return
		case update, ok := <-updateChan:
			if !ok {
				ep.logger.Info("Order update channel closed")
				return
			}
			if err := ep.PublishOrderUpdate(ctx, update); err != nil {
				ep.logger.Error("Failed to publish order update", zap.Error(err))
			}
		}
	}
}
