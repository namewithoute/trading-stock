package execution

import (
	"context"

	"trading-stock/internal/domain/execution"

	"go.uber.org/zap"
)

// UseCase handles trade history business logic
type UseCase interface {
	ListTrades(ctx context.Context, userID string) ([]*execution.Trade, error)
	GetTradeDetail(ctx context.Context, tradeID string) (*execution.Trade, error)
	GetMarketTrades(ctx context.Context, symbol string, limit int) ([]*execution.Trade, error)
}

type useCase struct {
	tradeRepo execution.TradeRepository
	logger    *zap.Logger
}

func NewUseCase(tradeRepo execution.TradeRepository, logger *zap.Logger) UseCase {
	return &useCase{tradeRepo: tradeRepo, logger: logger}
}

func (s *useCase) ListTrades(ctx context.Context, userID string) ([]*execution.Trade, error) {
	return s.tradeRepo.ListByUser(ctx, userID, 20, 0)
}

func (s *useCase) GetTradeDetail(ctx context.Context, tradeID string) (*execution.Trade, error) {
	return s.tradeRepo.GetByID(ctx, tradeID)
}

func (s *useCase) GetMarketTrades(ctx context.Context, symbol string, limit int) ([]*execution.Trade, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.tradeRepo.ListBySymbol(ctx, symbol, limit, 0)
}
