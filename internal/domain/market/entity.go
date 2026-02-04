package market

import "time"

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
	Price     float64
	Timestamp time.Time

	// Additional price data
	Bid    float64
	Ask    float64
	Volume int64
}

// Candle represents OHLCV (Open, High, Low, Close, Volume) data
type Candle struct {
	ID       string
	Symbol   string
	Interval string // 1m, 5m, 1h, 1d, etc.

	// OHLCV data
	Open   float64
	High   float64
	Low    float64
	Close  float64
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
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
}

// Spread returns the bid-ask spread
func (md *MarketDepth) Spread() float64 {
	if len(md.Bids) == 0 || len(md.Asks) == 0 {
		return 0
	}
	return md.Asks[0].Price - md.Bids[0].Price
}

// MidPrice returns the mid price between best bid and ask
func (md *MarketDepth) MidPrice() float64 {
	if len(md.Bids) == 0 || len(md.Asks) == 0 {
		return 0
	}
	return (md.Bids[0].Price + md.Asks[0].Price) / 2
}
