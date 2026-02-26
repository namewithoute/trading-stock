package market

import "time"

// ─── Shared ───────────────────────────────────────────────────────────────────

// Pagination carries page metadata returned alongside list responses.
type Pagination struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

// ─── Stocks ───────────────────────────────────────────────────────────────────

// ListStocksRequest carries query-string filters for the stock list endpoint.
type ListStocksRequest struct {
	Exchange string `query:"exchange"`
	Sector   string `query:"sector"`
	Search   string `query:"search"`
	Page     int    `query:"page"`
	Limit    int    `query:"limit"`
}

// StockDTO is the compact stock summary returned in list responses.
type StockDTO struct {
	Symbol     string    `json:"symbol"`
	Name       string    `json:"name"`
	Exchange   string    `json:"exchange"`
	Sector     string    `json:"sector"`
	Industry   string    `json:"industry"`
	IsActive   bool      `json:"is_active"`
	IsTradable bool      `json:"is_tradable"`
	Price      float64   `json:"price"`
	Bid        float64   `json:"bid"`
	Ask        float64   `json:"ask"`
	Volume     int64     `json:"volume"`
	PriceAt    time.Time `json:"price_at,omitempty"`
}

// ListStocksResponse wraps a paginated stock list.
type ListStocksResponse struct {
	Stocks     []StockDTO `json:"stocks"`
	Pagination Pagination `json:"pagination"`
}

// StockDetailResponse carries the full per-symbol view.
type StockDetailResponse struct {
	Symbol     string    `json:"symbol"`
	Name       string    `json:"name"`
	Exchange   string    `json:"exchange"`
	Sector     string    `json:"sector"`
	Industry   string    `json:"industry"`
	IsActive   bool      `json:"is_active"`
	IsTradable bool      `json:"is_tradable"`
	CreatedAt  time.Time `json:"created_at"`
	Price      float64   `json:"price"`
	Bid        float64   `json:"bid"`
	Ask        float64   `json:"ask"`
	Spread     float64   `json:"spread"`
	Volume     int64     `json:"volume"`
	PriceAt    time.Time `json:"price_at,omitempty"`
}

// ─── Trending ─────────────────────────────────────────────────────────────────

// TrendingStockDTO is used in the trending list response.
type TrendingStockDTO struct {
	Symbol   string  `json:"symbol"`
	Name     string  `json:"name"`
	Exchange string  `json:"exchange"`
	Price    float64 `json:"price"`
	Bid      float64 `json:"bid"`
	Ask      float64 `json:"ask"`
	Volume   int64   `json:"volume"`
}

// ─── Price ────────────────────────────────────────────────────────────────────

// PriceResponse is the single-symbol current price response.
type PriceResponse struct {
	Symbol    string    `json:"symbol"`
	Price     float64   `json:"price"`
	Bid       float64   `json:"bid"`
	Ask       float64   `json:"ask"`
	Spread    float64   `json:"spread"`
	Volume    int64     `json:"volume"`
	Timestamp time.Time `json:"timestamp"`
}

// PriceHistoryRequest carries query-string filters for price history.
type PriceHistoryRequest struct {
	From string `query:"from"` // RFC3339; defaults to 24 h ago
	To   string `query:"to"`   // RFC3339; defaults to now
}

// ─── Candles ──────────────────────────────────────────────────────────────────

// GetCandlesRequest carries query-string params for the candle endpoint.
type GetCandlesRequest struct {
	Interval string `query:"interval"` // 1m|5m|15m|30m|1h|4h|1d|1w  (default 1d)
	From     string `query:"from"`     // RFC3339; defaults to 30 days ago
	To       string `query:"to"`       // RFC3339; defaults to now
}

// CandleDTO represents a single OHLCV candle.
type CandleDTO struct {
	Timestamp time.Time `json:"timestamp"`
	Open      float64   `json:"open"`
	High      float64   `json:"high"`
	Low       float64   `json:"low"`
	Close     float64   `json:"close"`
	Volume    int64     `json:"volume"`
}

// GetCandlesResponse wraps the candle list.
type GetCandlesResponse struct {
	Symbol   string      `json:"symbol"`
	Interval string      `json:"interval"`
	From     time.Time   `json:"from"`
	To       time.Time   `json:"to"`
	Count    int         `json:"count"`
	Candles  []CandleDTO `json:"candles"`
}

// ─── Watchlist ────────────────────────────────────────────────────────────────

// AddWatchlistRequest is the body for POST /market/watchlist.
type AddWatchlistRequest struct {
	Symbol string `json:"symbol" validate:"required"`
}
