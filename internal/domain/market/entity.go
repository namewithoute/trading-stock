package market

import (
	"time"

	"github.com/cockroachdb/apd/v3"
)

var decCtx = apd.BaseContext.WithPrecision(19)

// Stock represents a tradable security/symbol
type Stock struct {
	ID       string
	Symbol   string
	Name     string
	Exchange string

	// Stock information
	Sector   string
	Industry string

	// Trading status
	IsActive   bool
	IsTradable bool

	// Timestamps
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Price represents real-time or historical price data
type Price struct {
	ID        string
	Symbol    string
	Price     apd.Decimal
	Timestamp time.Time

	// Additional price data
	Bid    apd.Decimal
	Ask    apd.Decimal
	Volume int64
}

// Candle represents OHLCV (Open, High, Low, Close, Volume) data
type Candle struct {
	ID       string
	Symbol   string
	Interval string // 1m, 5m, 1h, 1d, etc.

	// OHLCV data
	Open   apd.Decimal
	High   apd.Decimal
	Low    apd.Decimal
	Close  apd.Decimal
	Volume int64

	// Timestamp
	Timestamp time.Time
}

// MarketDepth represents the order book depth (bid/ask levels)
type MarketDepth struct {
	Symbol    string       `json:"symbol"`
	Bids      []PriceLevel `json:"bids"` // Buy orders
	Asks      []PriceLevel `json:"asks"` // Sell orders
	Timestamp time.Time    `json:"timestamp"`
}

// PriceLevel represents a single price level in the order book
type PriceLevel struct {
	Price    apd.Decimal `json:"price"`
	Quantity int         `json:"quantity"`
}

// Spread returns the bid-ask spread
func (md *MarketDepth) Spread() apd.Decimal {
	if len(md.Bids) == 0 || len(md.Asks) == 0 {
		return apd.Decimal{}
	}
	var spread apd.Decimal
	_, _ = decCtx.Sub(&spread, &md.Asks[0].Price, &md.Bids[0].Price)
	return spread
}

// MidPrice returns the mid price between best bid and ask
func (md *MarketDepth) MidPrice() apd.Decimal {
	if len(md.Bids) == 0 || len(md.Asks) == 0 {
		return apd.Decimal{}
	}
	var sum, mid apd.Decimal
	_, _ = decCtx.Add(&sum, &md.Bids[0].Price, &md.Asks[0].Price)
	_, _ = decCtx.Quo(&mid, &sum, apd.New(2, 0))
	return mid
}
