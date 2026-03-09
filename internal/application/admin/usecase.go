package admin

import (
	"context"
	"trading-stock/internal/domain/order"
	"trading-stock/internal/domain/user"

	"go.uber.org/zap"
)

// UseCase handles admin business logic
type UseCase interface {
	ListUsers(ctx context.Context, limit, offset int) ([]user.User, error)
	GetSystemStats(ctx context.Context) (interface{}, error)
	ApproveKYC(ctx context.Context, userID string, status user.KYCStatus) error
	ListAllOrders(ctx context.Context, limit, offset int) ([]*order.OrderReadModel, int64, error)
}

type useCase struct {
	userRepo  user.Repository
	orderRepo order.ReadModelRepository
	logger    *zap.Logger
}

func NewUseCase(userRepo user.Repository, orderRepo order.ReadModelRepository, logger *zap.Logger) UseCase {
	return &useCase{userRepo: userRepo, orderRepo: orderRepo, logger: logger}
}

func (s *useCase) ListUsers(ctx context.Context, limit, offset int) ([]user.User, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.userRepo.List(ctx, limit, offset)
}

func (s *useCase) GetSystemStats(ctx context.Context) (interface{}, error) {
	userCount, _ := s.userRepo.Count(ctx)
	orderCount, _ := s.orderRepo.CountAll(ctx)

	return map[string]interface{}{
		"total_users":   userCount,
		"total_orders":  orderCount,
		"system_status": "operational",
	}, nil
}

func (s *useCase) ApproveKYC(ctx context.Context, userID string, status user.KYCStatus) error {
	return s.userRepo.UpdateKYCStatus(ctx, userID, status)
}

func (s *useCase) ListAllOrders(ctx context.Context, limit, offset int) ([]*order.OrderReadModel, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	orders, err := s.orderRepo.ListAll(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.orderRepo.CountAll(ctx)
	if err != nil {
		return nil, 0, err
	}
	return orders, total, nil
}
