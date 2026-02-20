package execution

import (
	"context"

	"trading-stock/internal/domain/execution"

	"go.uber.org/zap"
)

// UseCase handles trade history business logic
type UseCase interface {
	ListTrades(ctx context.Context, userID string) (interface{}, error)
}

type useCase struct {
	tradeRepo execution.TradeRepository
	logger    *zap.Logger
}

func NewUseCase(tradeRepo execution.TradeRepository, logger *zap.Logger) UseCase {
	return &useCase{tradeRepo: tradeRepo, logger: logger}
}

func (s *useCase) ListTrades(ctx context.Context, userID string) (interface{}, error) {
	return s.tradeRepo.ListByUser(ctx, userID, 20, 0)
}
