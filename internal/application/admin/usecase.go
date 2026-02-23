package admin

import (
	"context"
	"trading-stock/internal/domain/order"
	"trading-stock/internal/domain/user"

	"go.uber.org/zap"
)

// UseCase handles admin business logic
type UseCase interface {
	ListUsers(ctx context.Context) ([]user.User, error)
	GetSystemStats(ctx context.Context) (interface{}, error)
}

type useCase struct {
	userRepo  user.Repository
	orderRepo order.ReadModelRepository // query-side: list/count orders for stats
	logger    *zap.Logger
}

func NewUseCase(userRepo user.Repository, orderRepo order.ReadModelRepository, logger *zap.Logger) UseCase {
	return &useCase{userRepo: userRepo, orderRepo: orderRepo, logger: logger}
}

func (s *useCase) ListUsers(ctx context.Context) ([]user.User, error) {
	users, err := s.userRepo.List(ctx, 20, 0)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (s *useCase) GetSystemStats(ctx context.Context) (interface{}, error) {
	userCount, _ := s.userRepo.Count(ctx)
	// Add more stats if repository methods exist

	return map[string]interface{}{
		"total_users":   userCount,
		"system_status": "operational",
	}, nil
}
