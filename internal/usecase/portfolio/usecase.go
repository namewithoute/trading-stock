package portfolio

import (
	"context"
	"trading-stock/internal/domain/portfolio"

	"go.uber.org/zap"
)

// UseCase handles portfolio business logic
type UseCase interface {
	GetOverview(ctx context.Context, userID string) (interface{}, error)
}

type useCase struct {
	portfolioRepo portfolio.Repository
	logger        *zap.Logger
}

func NewUseCase(portfolioRepo portfolio.Repository, logger *zap.Logger) UseCase {
	return &useCase{portfolioRepo: portfolioRepo, logger: logger}
}

func (s *useCase) GetOverview(ctx context.Context, userID string) (interface{}, error) {
	return s.portfolioRepo.ListByUserID(ctx, userID)
}
