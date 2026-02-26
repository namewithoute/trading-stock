package market

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"trading-stock/internal/domain/market"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	priceCachePrefix = "market:price:"
	priceCacheTTL    = 5 * time.Second
)

// StockWithPrice bundles a Stock entity with its most recent price tick.
type StockWithPrice struct {
	Stock       *market.Stock
	LatestPrice *market.Price // nil when no price data is available yet
}

// UseCase handles market data business logic.
type UseCase interface {
	// ListStocks returns a paginated, optionally-filtered list of stocks.
	ListStocks(ctx context.Context, exchange, sector, search string, limit, offset int) ([]*StockWithPrice, error)

	// SearchStocks performs a full-text search across symbol and name.
	SearchStocks(ctx context.Context, query string) ([]*StockWithPrice, error)

	// GetStockDetail returns stock metadata + latest price for a single symbol.
	GetStockDetail(ctx context.Context, symbol string) (*StockWithPrice, error)

	// GetLatestPrice returns the most recent price tick; checks Redis first.
	GetLatestPrice(ctx context.Context, symbol string) (*market.Price, error)

	// GetPriceHistory returns all price ticks within [from, to].
	GetPriceHistory(ctx context.Context, symbol string, from, to time.Time) ([]*market.Price, error)

	// GetCandles returns OHLCV candles for the requested interval and time range.
	GetCandles(ctx context.Context, symbol, interval string, from, to time.Time) ([]*market.Candle, error)

	// GetTrendingStocks returns the top N stocks ranked by trading volume.
	GetTrendingStocks(ctx context.Context, limit int) ([]*StockWithPrice, error)
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

// -- ListStocks ---------------------------------------------------------------

func (s *useCase) ListStocks(ctx context.Context, exchange, sector, search string, limit, offset int) ([]*StockWithPrice, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 200 {
		limit = 200
	}

	var stocks []*market.Stock
	var err error

	switch {
	case search != "":
		stocks, err = s.stockRepo.Search(ctx, search)
		if err != nil {
			return nil, fmt.Errorf("search stocks: %w", err)
		}
		stocks = paginate(stocks, offset, limit)

	case exchange != "":
		stocks, err = s.stockRepo.ListByExchange(ctx, exchange)
		if err != nil {
			return nil, fmt.Errorf("list stocks by exchange: %w", err)
		}
		if sector != "" {
			stocks = filterBySector(stocks, sector)
		}
		stocks = paginate(stocks, offset, limit)

	default:
		stocks, err = s.stockRepo.List(ctx, limit, offset)
		if err != nil {
			return nil, fmt.Errorf("list stocks: %w", err)
		}
		if sector != "" {
			stocks = filterBySector(stocks, sector)
		}
	}

	return s.attachPrices(ctx, stocks), nil
}

// -- SearchStocks -------------------------------------------------------------

func (s *useCase) SearchStocks(ctx context.Context, query string) ([]*StockWithPrice, error) {
	stocks, err := s.stockRepo.Search(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("search stocks: %w", err)
	}
	return s.attachPrices(ctx, stocks), nil
}

// -- GetStockDetail -----------------------------------------------------------

func (s *useCase) GetStockDetail(ctx context.Context, symbol string) (*StockWithPrice, error) {
	stock, err := s.stockRepo.GetBySymbol(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("stock not found: %w", err)
	}
	price, _ := s.GetLatestPrice(ctx, symbol)
	return &StockWithPrice{Stock: stock, LatestPrice: price}, nil
}

// -- GetLatestPrice -----------------------------------------------------------

func (s *useCase) GetLatestPrice(ctx context.Context, symbol string) (*market.Price, error) {
	// 1. Redis cache
	if s.redis != nil {
		key := priceCachePrefix + symbol
		val, err := s.redis.Get(ctx, key).Result()
		if err == nil {
			var p market.Price
			if jsonErr := json.Unmarshal([]byte(val), &p); jsonErr == nil {
				return &p, nil
			}
		}
	}

	// 2. DB fallback
	p, err := s.priceRepo.GetLatest(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("price not found for %s: %w", symbol, err)
	}

	// 3. Populate cache
	if s.redis != nil {
		if b, jsonErr := json.Marshal(p); jsonErr == nil {
			_ = s.redis.Set(ctx, priceCachePrefix+symbol, b, priceCacheTTL).Err()
		}
	}
	return p, nil
}

// -- GetPriceHistory ----------------------------------------------------------

func (s *useCase) GetPriceHistory(ctx context.Context, symbol string, from, to time.Time) ([]*market.Price, error) {
	prices, err := s.priceRepo.ListBySymbol(ctx, symbol, from, to)
	if err != nil {
		return nil, fmt.Errorf("price history for %s: %w", symbol, err)
	}
	return prices, nil
}

// -- GetCandles ---------------------------------------------------------------

func (s *useCase) GetCandles(ctx context.Context, symbol, interval string, from, to time.Time) ([]*market.Candle, error) {
	if interval == "" {
		interval = "1d"
	}
	candles, err := s.candleRepo.GetBySymbolAndInterval(ctx, symbol, interval, from, to)
	if err != nil {
		return nil, fmt.Errorf("candles for %s/%s: %w", symbol, interval, err)
	}
	return candles, nil
}

// -- GetTrendingStocks --------------------------------------------------------

func (s *useCase) GetTrendingStocks(ctx context.Context, limit int) ([]*StockWithPrice, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	stocks, err := s.stockRepo.List(ctx, 100, 0)
	if err != nil {
		return nil, fmt.Errorf("list stocks for trending: %w", err)
	}

	withPrices := s.attachPrices(ctx, stocks)

	// Sort descending by volume
	sort.Slice(withPrices, func(i, j int) bool {
		var vi, vj int64
		if withPrices[i].LatestPrice != nil {
			vi = withPrices[i].LatestPrice.Volume
		}
		if withPrices[j].LatestPrice != nil {
			vj = withPrices[j].LatestPrice.Volume
		}
		return vi > vj
	})

	if limit > len(withPrices) {
		limit = len(withPrices)
	}
	return withPrices[:limit], nil
}

// -- helpers ------------------------------------------------------------------

func (s *useCase) attachPrices(ctx context.Context, stocks []*market.Stock) []*StockWithPrice {
	result := make([]*StockWithPrice, 0, len(stocks))
	for _, st := range stocks {
		p, err := s.GetLatestPrice(ctx, st.Symbol)
		if err != nil {
			s.logger.Debug("No price for stock", zap.String("symbol", st.Symbol))
			p = nil
		}
		result = append(result, &StockWithPrice{Stock: st, LatestPrice: p})
	}
	return result
}

func filterBySector(stocks []*market.Stock, sector string) []*market.Stock {
	out := stocks[:0]
	for _, s := range stocks {
		if s.Sector == sector {
			out = append(out, s)
		}
	}
	return out
}

func paginate[T any](slice []T, offset, limit int) []T {
	if offset >= len(slice) {
		return nil
	}
	end := offset + limit
	if end > len(slice) {
		end = len(slice)
	}
	return slice[offset:end]
}
