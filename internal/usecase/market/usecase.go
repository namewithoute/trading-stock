package market

import (
	"context"
	"trading-stock/internal/domain/market"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// UseCase handles market data business logic
type UseCase interface {
	ListStocks(ctx context.Context) (interface{}, error)
	GetStockDetail(ctx context.Context, symbol string) (interface{}, error)
}

type useCase struct {
	stockRepo  market.StockRepository
	priceRepo  market.PriceRepository
	candleRepo market.CandleRepository
	redis      *redis.Client
	logger     *zap.Logger
}

func NewUseCase(
	stockRepo market.StockRepository,
	priceRepo market.PriceRepository,
	candleRepo market.CandleRepository,
	redis *redis.Client,
	logger *zap.Logger,
) UseCase {
	return &useCase{
		stockRepo:  stockRepo,
		priceRepo:  priceRepo,
		candleRepo: candleRepo,
		redis:      redis,
		logger:     logger,
	}
}

func (s *useCase) ListStocks(ctx context.Context) (interface{}, error) {
	return s.stockRepo.List(ctx, 100, 0)
}

func (s *useCase) GetStockDetail(ctx context.Context, symbol string) (interface{}, error) {
	return s.stockRepo.GetBySymbol(ctx, symbol)
}
