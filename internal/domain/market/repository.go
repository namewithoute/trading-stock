package market

import (
	"context"
	"time"
)

// StockRepository defines the interface for stock data access
type StockRepository interface {
	// Create creates a new stock
	Create(ctx context.Context, stock *Stock) error

	// GetByID retrieves a stock by its ID
	GetByID(ctx context.Context, id string) (*Stock, error)

	// GetBySymbol retrieves a stock by its symbol
	GetBySymbol(ctx context.Context, symbol string) (*Stock, error)

	// Update updates an existing stock
	Update(ctx context.Context, stock *Stock) error

	// List retrieves all stocks with pagination
	List(ctx context.Context, limit, offset int) ([]*Stock, error)

	// ListByExchange retrieves stocks by exchange
	ListByExchange(ctx context.Context, exchange string) ([]*Stock, error)

	// Search searches for stocks by name or symbol
	Search(ctx context.Context, query string) ([]*Stock, error)

	// Delete deletes a stock
	Delete(ctx context.Context, id string) error
}

// PriceRepository defines the interface for price data access
type PriceRepository interface {
	// Create creates a new price record
	Create(ctx context.Context, price *Price) error

	// GetLatest retrieves the latest price for a symbol
	GetLatest(ctx context.Context, symbol string) (*Price, error)

	// GetBySymbolAndTime retrieves price at a specific time
	GetBySymbolAndTime(ctx context.Context, symbol string, timestamp time.Time) (*Price, error)

	// ListBySymbol retrieves price history for a symbol
	ListBySymbol(ctx context.Context, symbol string, from, to time.Time) ([]*Price, error)

	// BatchCreate creates multiple price records
	BatchCreate(ctx context.Context, prices []*Price) error
}

// CandleRepository defines the interface for candle data access
type CandleRepository interface {
	// Create creates a new candle
	Create(ctx context.Context, candle *Candle) error

	// Update updates an existing candle (used to update OHLCV fields).
	Update(ctx context.Context, candle *Candle) error

	// GetBySymbolAndInterval retrieves candles for a symbol and interval
	GetBySymbolAndInterval(ctx context.Context, symbol, interval string, from, to time.Time) ([]*Candle, error)

	// GetLatest retrieves the latest candle for a symbol and interval
	GetLatest(ctx context.Context, symbol, interval string) (*Candle, error)

	// BatchCreate creates multiple candles
	BatchCreate(ctx context.Context, candles []*Candle) error

	// Delete deletes old candles (for cleanup)
	DeleteOlderThan(ctx context.Context, timestamp time.Time) error
}
