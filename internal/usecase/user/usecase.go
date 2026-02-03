package user

import (
	"context"
	"trading-stock/internal/domain/user"

	"go.uber.org/zap"
)

// UseCase handles user business logic
type UseCase interface {
	GetProfile(ctx context.Context, userID string) (interface{}, error)
	UpdateProfile(ctx context.Context, userID string, data map[string]interface{}) error
}

type useCase struct {
	userRepo user.Repository
	logger   *zap.Logger
}

func NewUseCase(userRepo user.Repository, logger *zap.Logger) UseCase {
	return &useCase{userRepo: userRepo, logger: logger}
}

func (s *useCase) GetProfile(ctx context.Context, userID string) (interface{}, error) {
	return s.userRepo.GetByID(ctx, userID)
}

func (s *useCase) UpdateProfile(ctx context.Context, userID string, data map[string]interface{}) error {
	// TODO: Implement update logic
	return nil
}
