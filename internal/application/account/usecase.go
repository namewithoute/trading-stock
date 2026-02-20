package account

import (
	"context"
	"time"
	"trading-stock/internal/domain/account"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// UseCase handles account business logic
type UseCase interface {
	ListAccounts(ctx context.Context, userID string) ([]*account.Account, error)
	CreateAccount(ctx context.Context, userID string) (*account.Account, error)
	GetAccount(ctx context.Context, id string) (*account.Account, error)
}

type useCase struct {
	accountRepo account.Repository
	logger      *zap.Logger
}

func NewUseCase(accountRepo account.Repository, logger *zap.Logger) UseCase {
	return &useCase{accountRepo: accountRepo, logger: logger}
}

func (s *useCase) ListAccounts(ctx context.Context, userID string) ([]*account.Account, error) {
	return s.accountRepo.GetByUserID(ctx, userID)
}

func (s *useCase) CreateAccount(ctx context.Context, userID string) (*account.Account, error) {
	acc := &account.Account{
		ID:          uuid.New().String(),
		UserID:      userID,
		AccountType: account.TypeCash,
		Balance:     0,
		BuyingPower: 0,
		Currency:    "USD",
		Status:      account.StatusActive,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.accountRepo.Create(ctx, acc); err != nil {
		s.logger.Error("Failed to create account", zap.Error(err), zap.String("userID", userID))
		return nil, err
	}

	s.logger.Info("Account created successfully", zap.String("accountID", acc.ID), zap.String("userID", userID))
	return acc, nil
}

func (s *useCase) GetAccount(ctx context.Context, id string) (*account.Account, error) {
	return s.accountRepo.GetByID(ctx, id)
}
