package portfolio

import (
	"context"
	"trading-stock/internal/domain/portfolio"

	"go.uber.org/zap"
)

// UseCase handles portfolio business logic
type UseCase interface {
	GetOverview(ctx context.Context, userID string) ([]*portfolio.Position, error)
	GetTotalValue(ctx context.Context, userID string) (float64, error)
}

type useCase struct {
	portfolioRepo portfolio.Repository
	logger        *zap.Logger
}

func NewUseCase(portfolioRepo portfolio.Repository, logger *zap.Logger) UseCase {
	return &useCase{portfolioRepo: portfolioRepo, logger: logger}
}

func (s *useCase) GetOverview(ctx context.Context, userID string) ([]*portfolio.Position, error) {
	return s.portfolioRepo.ListByUserID(ctx, userID)
}

func (s *useCase) GetTotalValue(ctx context.Context, userID string) (float64, error) {
	return s.portfolioRepo.GetTotalValue(ctx, userID)
}
