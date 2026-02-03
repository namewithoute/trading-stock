package admin

import (
	"context"
	"trading-stock/internal/domain/order"
	"trading-stock/internal/domain/user"

	"go.uber.org/zap"
)

// UseCase handles admin business logic
type UseCase interface {
	ListUsers(ctx context.Context) (interface{}, error)
	GetSystemStats(ctx context.Context) (interface{}, error)
}

type useCase struct {
	userRepo  user.Repository
	orderRepo order.Repository
	logger    *zap.Logger
}

func NewUseCase(userRepo user.Repository, orderRepo order.Repository, logger *zap.Logger) UseCase {
	return &useCase{userRepo: userRepo, orderRepo: orderRepo, logger: logger}
}

func (s *useCase) ListUsers(ctx context.Context) (interface{}, error) {
	return s.userRepo.List(ctx, 20, 0)
}

func (s *useCase) GetSystemStats(ctx context.Context) (interface{}, error) {
	// TODO: Implement system stats aggregation
	return nil, nil
}
