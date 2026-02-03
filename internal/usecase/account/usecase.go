package account

import (
	"context"
	"trading-stock/internal/domain/account"

	"go.uber.org/zap"
)

// UseCase handles account business logic
type UseCase interface {
	ListAccounts(ctx context.Context, userID string) (interface{}, error)
	CreateAccount(ctx context.Context, userID string) (interface{}, error)
}

type useCase struct {
	accountRepo account.Repository
	logger      *zap.Logger
}

func NewUseCase(accountRepo account.Repository, logger *zap.Logger) UseCase {
	return &useCase{accountRepo: accountRepo, logger: logger}
}

func (s *useCase) ListAccounts(ctx context.Context, userID string) (interface{}, error) {
	return s.accountRepo.GetByUserID(ctx, userID)
}

func (s *useCase) CreateAccount(ctx context.Context, userID string) (interface{}, error) {
	// TODO: Implement create account logic
	return nil, nil
}
