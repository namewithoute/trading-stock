package market

import (
	"context"
	"time"

	"trading-stock/internal/domain/market"

	"gorm.io/gorm"
)

// stockRepository implements market.StockRepository
type stockRepository struct {
	db *gorm.DB
}

func NewStockRepository(db *gorm.DB) market.StockRepository {
	return &stockRepository{db: db}
}

func (r *stockRepository) Create(ctx context.Context, s *market.Stock) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *stockRepository) GetByID(ctx context.Context, id string) (*market.Stock, error) {
	var s market.Stock
	err := r.db.WithContext(ctx).First(&s, "id = ?", id).Error
	return &s, err
}

func (r *stockRepository) GetBySymbol(ctx context.Context, symbol string) (*market.Stock, error) {
	var s market.Stock
	err := r.db.WithContext(ctx).First(&s, "symbol = ?", symbol).Error
	return &s, err
}

func (r *stockRepository) Update(ctx context.Context, s *market.Stock) error {
	return r.db.WithContext(ctx).Save(s).Error
}

func (r *stockRepository) List(ctx context.Context, limit, offset int) ([]*market.Stock, error) {
	var stocks []*market.Stock
	err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&stocks).Error
	return stocks, err
}

func (r *stockRepository) ListByExchange(ctx context.Context, exchange string) ([]*market.Stock, error) {
	var stocks []*market.Stock
	err := r.db.WithContext(ctx).Where("exchange = ?", exchange).Find(&stocks).Error
	return stocks, err
}

func (r *stockRepository) Search(ctx context.Context, query string) ([]*market.Stock, error) {
	var stocks []*market.Stock
	q := "%" + query + "%"
	err := r.db.WithContext(ctx).Where("symbol LIKE ? OR name LIKE ?", q, q).Find(&stocks).Error
	return stocks, err
}

func (r *stockRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&market.Stock{}, "id = ?", id).Error
}

// priceRepository implements market.PriceRepository
type priceRepository struct {
	db *gorm.DB
}

func NewPriceRepository(db *gorm.DB) market.PriceRepository {
	return &priceRepository{db: db}
}

func (r *priceRepository) Create(ctx context.Context, p *market.Price) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *priceRepository) GetLatest(ctx context.Context, symbol string) (*market.Price, error) {
	var p market.Price
	err := r.db.WithContext(ctx).Where("symbol = ?", symbol).Order("timestamp DESC").First(&p).Error
	return &p, err
}

func (r *priceRepository) GetBySymbolAndTime(ctx context.Context, symbol string, timestamp time.Time) (*market.Price, error) {
	var p market.Price
	err := r.db.WithContext(ctx).Where("symbol = ? AND timestamp <= ?", symbol, timestamp).Order("timestamp DESC").First(&p).Error
	return &p, err
}

func (r *priceRepository) ListBySymbol(ctx context.Context, symbol string, from, to time.Time) ([]*market.Price, error) {
	var prices []*market.Price
	err := r.db.WithContext(ctx).Where("symbol = ? AND timestamp BETWEEN ? AND ?", symbol, from, to).Order("timestamp ASC").Find(&prices).Error
	return prices, err
}

func (r *priceRepository) BatchCreate(ctx context.Context, prices []*market.Price) error {
	return r.db.WithContext(ctx).Create(&prices).Error
}

// candleRepository implements market.CandleRepository
type candleRepository struct {
	db *gorm.DB
}

func NewCandleRepository(db *gorm.DB) market.CandleRepository {
	return &candleRepository{db: db}
}

func (r *candleRepository) Create(ctx context.Context, c *market.Candle) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *candleRepository) GetBySymbolAndInterval(ctx context.Context, symbol, interval string, from, to time.Time) ([]*market.Candle, error) {
	var candles []*market.Candle
	err := r.db.WithContext(ctx).Where("symbol = ? AND interval = ? AND timestamp BETWEEN ? AND ?", symbol, interval, from, to).Order("timestamp ASC").Find(&candles).Error
	return candles, err
}

func (r *candleRepository) GetLatest(ctx context.Context, symbol, interval string) (*market.Candle, error) {
	var c market.Candle
	err := r.db.WithContext(ctx).Where("symbol = ? AND interval = ?", symbol, interval).Order("timestamp DESC").First(&c).Error
	return &c, err
}

func (r *candleRepository) BatchCreate(ctx context.Context, candles []*market.Candle) error {
	return r.db.WithContext(ctx).Create(&candles).Error
}

func (r *candleRepository) DeleteOlderThan(ctx context.Context, timestamp time.Time) error {
	return r.db.WithContext(ctx).Where("timestamp < ?", timestamp).Delete(&market.Candle{}).Error
}
