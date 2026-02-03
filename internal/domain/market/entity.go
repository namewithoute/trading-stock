package market

import "time"

// Stock represents a tradable security/symbol
type Stock struct {
	ID       string `json:"id" gorm:"primaryKey;type:uuid"`
	Symbol   string `json:"symbol" gorm:"uniqueIndex;type:varchar(10);not null"`
	Name     string `json:"name" gorm:"type:varchar(255);not null"`
	Exchange string `json:"exchange" gorm:"type:varchar(50)"`

	// Stock information
	Sector   string `json:"sector" gorm:"type:varchar(100)"`
	Industry string `json:"industry" gorm:"type:varchar(100)"`

	// Trading status
	IsActive   bool `json:"is_active" gorm:"default:true"`
	IsTradable bool `json:"is_tradable" gorm:"default:true"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt time.Time `json:"updated_at" gorm:"not null"`
}

// TableName specifies the table name for GORM
func (Stock) TableName() string {
	return "stocks"
}

// Price represents real-time or historical price data
type Price struct {
	ID        string    `json:"id" gorm:"primaryKey;type:uuid"`
	Symbol    string    `json:"symbol" gorm:"index;type:varchar(10);not null"`
	Price     float64   `json:"price" gorm:"type:decimal(20,4);not null"`
	Timestamp time.Time `json:"timestamp" gorm:"index;not null"`

	// Additional price data
	Bid    float64 `json:"bid,omitempty" gorm:"type:decimal(20,4)"`
	Ask    float64 `json:"ask,omitempty" gorm:"type:decimal(20,4)"`
	Volume int64   `json:"volume,omitempty"`
}

// TableName specifies the table name for GORM
func (Price) TableName() string {
	return "prices"
}

// Candle represents OHLCV (Open, High, Low, Close, Volume) data
type Candle struct {
	ID       string `json:"id" gorm:"primaryKey;type:uuid"`
	Symbol   string `json:"symbol" gorm:"index;type:varchar(10);not null"`
	Interval string `json:"interval" gorm:"type:varchar(10);not null"` // 1m, 5m, 1h, 1d, etc.

	// OHLCV data
	Open   float64 `json:"open" gorm:"type:decimal(20,4);not null"`
	High   float64 `json:"high" gorm:"type:decimal(20,4);not null"`
	Low    float64 `json:"low" gorm:"type:decimal(20,4);not null"`
	Close  float64 `json:"close" gorm:"type:decimal(20,4);not null"`
	Volume int64   `json:"volume" gorm:"not null"`

	// Timestamp
	Timestamp time.Time `json:"timestamp" gorm:"index;not null"`
}

// TableName specifies the table name for GORM
func (Candle) TableName() string {
	return "candles"
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
