package order

import (
	"context"
	"trading-stock/internal/domain/order"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// UseCase handles order business logic
type UseCase interface {
	CreateOrder(ctx context.Context, userID string, data map[string]interface{}) (interface{}, error)
	ListOrders(ctx context.Context, userID string) (interface{}, error)
}

type useCase struct {
	orderRepo order.Repository
	kafka     *kafka.Writer
	logger    *zap.Logger
}

func NewUseCase(orderRepo order.Repository, kafka *kafka.Writer, logger *zap.Logger) UseCase {
	return &useCase{orderRepo: orderRepo, kafka: kafka, logger: logger}
}

func (s *useCase) CreateOrder(ctx context.Context, userID string, data map[string]interface{}) (interface{}, error) {
	// TODO: Implement create order + send to Kafka
	return nil, nil
}

func (s *useCase) ListOrders(ctx context.Context, userID string) (interface{}, error) {
	return s.orderRepo.ListByUserID(ctx, userID, 20, 0)
}
