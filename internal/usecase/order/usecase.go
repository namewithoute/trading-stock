package order

import (
	"context"
	"encoding/json"
	"time"
	"trading-stock/internal/domain/order"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// UseCase handles order business logic
type UseCase interface {
	CreateOrder(ctx context.Context, userID string, data map[string]interface{}) (*order.Order, error)
	ListOrders(ctx context.Context, userID string) ([]*order.Order, error)
	GetOrder(ctx context.Context, id string) (*order.Order, error)
	CancelOrder(ctx context.Context, id string) error
}

type useCase struct {
	orderRepo order.Repository
	kafka     *kafka.Writer
	logger    *zap.Logger
}

func NewUseCase(orderRepo order.Repository, kafka *kafka.Writer, logger *zap.Logger) UseCase {
	return &useCase{orderRepo: orderRepo, kafka: kafka, logger: logger}
}

func (s *useCase) CreateOrder(ctx context.Context, userID string, data map[string]interface{}) (*order.Order, error) {
	symbol, _ := data["symbol"].(string)
	accountID, _ := data["account_id"].(string)
	side, _ := data["side"].(string)
	orderType, _ := data["order_type"].(string)

	var price float64
	if p, ok := data["price"].(float64); ok {
		price = p
	}

	var quantity int
	if q, ok := data["quantity"].(float64); ok {
		quantity = int(q)
	}

	o := &order.Order{
		ID:        uuid.New().String(),
		UserID:    userID,
		AccountID: accountID,
		Symbol:    symbol,
		Price:     price,
		Quantity:  quantity,
		Side:      order.Side(side),
		Type:      order.OrderType(orderType),
		Status:    order.StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.orderRepo.Create(ctx, o); err != nil {
		s.logger.Error("Failed to create order in repo", zap.Error(err))
		return nil, err
	}

	// Send to Kafka for matching engine
	orderJSON, _ := json.Marshal(o)
	err := s.kafka.WriteMessages(ctx, kafka.Message{
		Key:   []byte(o.ID),
		Value: orderJSON,
	})
	if err != nil {
		s.logger.Error("Failed to send order to Kafka", zap.Error(err), zap.String("orderID", o.ID))
		// Note: depending on criticality, we might want to fail the request or retry
	}

	s.logger.Info("Order created successfully", zap.String("orderID", o.ID), zap.String("userID", userID))
	return o, nil
}

func (s *useCase) ListOrders(ctx context.Context, userID string) ([]*order.Order, error) {
	return s.orderRepo.ListByUserID(ctx, userID, 20, 0)
}

func (s *useCase) GetOrder(ctx context.Context, id string) (*order.Order, error) {
	return s.orderRepo.GetByID(ctx, id)
}

func (s *useCase) CancelOrder(ctx context.Context, id string) error {
	o, err := s.orderRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if !o.CanBeCancelled() {
		return order.ErrInvalidStatus // Wait, where is this error?
	}

	return s.orderRepo.Cancel(ctx, id)
}
