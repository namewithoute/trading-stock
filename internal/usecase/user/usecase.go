package user

import (
	"context"
	"trading-stock/internal/domain/user"

	"go.uber.org/zap"
)

// UseCase handles user business logic
type UseCase interface {
	GetProfile(ctx context.Context, userID string) (*user.User, error)
	UpdateProfile(ctx context.Context, userID string, data map[string]interface{}) error
}

type useCase struct {
	userRepo user.Repository
	logger   *zap.Logger
}

func NewUseCase(userRepo user.Repository, logger *zap.Logger) UseCase {
	return &useCase{userRepo: userRepo, logger: logger}
}

func (s *useCase) GetProfile(ctx context.Context, userID string) (*user.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

func (s *useCase) UpdateProfile(ctx context.Context, userID string, data map[string]interface{}) error {
	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if firstName, ok := data["first_name"].(string); ok {
		u.FirstName = firstName
	}
	if lastName, ok := data["last_name"].(string); ok {
		u.LastName = lastName
	}

	return s.userRepo.Update(ctx, u)
}
