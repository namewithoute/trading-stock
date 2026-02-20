package market

import (
	"context"
	"trading-stock/internal/domain/market"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// UseCase handles market data business logic
type UseCase interface {
	ListStocks(ctx context.Context) ([]*market.Stock, error)
	GetStockDetail(ctx context.Context, symbol string) (*market.Stock, error)
	GetLatestPrice(ctx context.Context, symbol string) (*market.Price, error)
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

func (s *useCase) ListStocks(ctx context.Context) ([]*market.Stock, error) {
	return s.stockRepo.List(ctx, 100, 0)
}

func (s *useCase) GetStockDetail(ctx context.Context, symbol string) (*market.Stock, error) {
	return s.stockRepo.GetBySymbol(ctx, symbol)
}

func (s *useCase) GetLatestPrice(ctx context.Context, symbol string) (*market.Price, error) {
	return s.priceRepo.GetLatest(ctx, symbol)
}
