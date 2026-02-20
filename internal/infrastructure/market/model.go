package market

import (
	"time"

	domain "trading-stock/internal/domain/market"
)

// StockModel is the GORM persistence model for stocks.
type StockModel struct {
	ID       string `gorm:"primaryKey;type:uuid"`
	Symbol   string `gorm:"uniqueIndex;type:varchar(10);not null"`
	Name     string `gorm:"type:varchar(255);not null"`
	Exchange string `gorm:"type:varchar(50)"`

	Sector   string `gorm:"type:varchar(100)"`
	Industry string `gorm:"type:varchar(100)"`

	IsActive   bool `gorm:"default:true"`
	IsTradable bool `gorm:"default:true"`

	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (StockModel) TableName() string { return "stocks" }

func toStockModel(s *domain.Stock) *StockModel {
	if s == nil {
		return nil
	}
	return &StockModel{
		ID:         s.ID,
		Symbol:     s.Symbol,
		Name:       s.Name,
		Exchange:   s.Exchange,
		Sector:     s.Sector,
		Industry:   s.Industry,
		IsActive:   s.IsActive,
		IsTradable: s.IsTradable,
		CreatedAt:  s.CreatedAt,
		UpdatedAt:  s.UpdatedAt,
	}
}

func (m *StockModel) toDomain() *domain.Stock {
	if m == nil {
		return nil
	}
	return &domain.Stock{
		ID:         m.ID,
		Symbol:     m.Symbol,
		Name:       m.Name,
		Exchange:   m.Exchange,
		Sector:     m.Sector,
		Industry:   m.Industry,
		IsActive:   m.IsActive,
		IsTradable: m.IsTradable,
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
	}
}

// PriceModel is the GORM persistence model for prices.
type PriceModel struct {
	ID        string    `gorm:"primaryKey;type:uuid"`
	Symbol    string    `gorm:"index;type:varchar(10);not null"`
	Price     float64   `gorm:"type:decimal(20,4);not null"`
	Timestamp time.Time `gorm:"index;not null"`

	Bid    float64 `gorm:"type:decimal(20,4)"`
	Ask    float64 `gorm:"type:decimal(20,4)"`
	Volume int64
}

func (PriceModel) TableName() string { return "prices" }

func toPriceModel(p *domain.Price) *PriceModel {
	if p == nil {
		return nil
	}
	return &PriceModel{
		ID:        p.ID,
		Symbol:    p.Symbol,
		Price:     p.Price,
		Timestamp: p.Timestamp,
		Bid:       p.Bid,
		Ask:       p.Ask,
		Volume:    p.Volume,
	}
}

func (m *PriceModel) toDomain() *domain.Price {
	if m == nil {
		return nil
	}
	return &domain.Price{
		ID:        m.ID,
		Symbol:    m.Symbol,
		Price:     m.Price,
		Timestamp: m.Timestamp,
		Bid:       m.Bid,
		Ask:       m.Ask,
		Volume:    m.Volume,
	}
}

// CandleModel is the GORM persistence model for candles.
type CandleModel struct {
	ID       string `gorm:"primaryKey;type:uuid"`
	Symbol   string `gorm:"index;type:varchar(10);not null"`
	Interval string `gorm:"type:varchar(10);not null"`

	Open   float64 `gorm:"type:decimal(20,4);not null"`
	High   float64 `gorm:"type:decimal(20,4);not null"`
	Low    float64 `gorm:"type:decimal(20,4);not null"`
	Close  float64 `gorm:"type:decimal(20,4);not null"`
	Volume int64   `gorm:"not null"`

	Timestamp time.Time `gorm:"index;not null"`
}

func (CandleModel) TableName() string { return "candles" }

func toCandleModel(c *domain.Candle) *CandleModel {
	if c == nil {
		return nil
	}
	return &CandleModel{
		ID:        c.ID,
		Symbol:    c.Symbol,
		Interval:  c.Interval,
		Open:      c.Open,
		High:      c.High,
		Low:       c.Low,
		Close:     c.Close,
		Volume:    c.Volume,
		Timestamp: c.Timestamp,
	}
}

func (m *CandleModel) toDomain() *domain.Candle {
	if m == nil {
		return nil
	}
	return &domain.Candle{
		ID:        m.ID,
		Symbol:    m.Symbol,
		Interval:  m.Interval,
		Open:      m.Open,
		High:      m.High,
		Low:       m.Low,
		Close:     m.Close,
		Volume:    m.Volume,
		Timestamp: m.Timestamp,
	}
}
