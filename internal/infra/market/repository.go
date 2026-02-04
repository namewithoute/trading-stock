package market

import (
	"context"
	"time"

	domain "trading-stock/internal/domain/market"

	"gorm.io/gorm"
)

// stockRepository implements domain.StockRepository
type stockRepository struct {
	db *gorm.DB
}

func NewStockRepository(db *gorm.DB) domain.StockRepository {
	return &stockRepository{db: db}
}

func (r *stockRepository) Create(ctx context.Context, s *domain.Stock) error {
	return r.db.WithContext(ctx).Create(toStockModel(s)).Error
}

func (r *stockRepository) GetByID(ctx context.Context, id string) (*domain.Stock, error) {
	var s StockModel
	err := r.db.WithContext(ctx).First(&s, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return s.toDomain(), nil
}

func (r *stockRepository) GetBySymbol(ctx context.Context, symbol string) (*domain.Stock, error) {
	var s StockModel
	err := r.db.WithContext(ctx).First(&s, "symbol = ?", symbol).Error
	if err != nil {
		return nil, err
	}
	return s.toDomain(), nil
}

func (r *stockRepository) Update(ctx context.Context, s *domain.Stock) error {
	return r.db.WithContext(ctx).Save(toStockModel(s)).Error
}

func (r *stockRepository) List(ctx context.Context, limit, offset int) ([]*domain.Stock, error) {
	var models []*StockModel
	err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&models).Error
	if err != nil {
		return nil, err
	}
	stocks := make([]*domain.Stock, 0, len(models))
	for _, m := range models {
		stocks = append(stocks, m.toDomain())
	}
	return stocks, nil
}

func (r *stockRepository) ListByExchange(ctx context.Context, exchange string) ([]*domain.Stock, error) {
	var models []*StockModel
	err := r.db.WithContext(ctx).Where("exchange = ?", exchange).Find(&models).Error
	if err != nil {
		return nil, err
	}
	stocks := make([]*domain.Stock, 0, len(models))
	for _, m := range models {
		stocks = append(stocks, m.toDomain())
	}
	return stocks, nil
}

func (r *stockRepository) Search(ctx context.Context, query string) ([]*domain.Stock, error) {
	var models []*StockModel
	q := "%" + query + "%"
	err := r.db.WithContext(ctx).Where("symbol LIKE ? OR name LIKE ?", q, q).Find(&models).Error
	if err != nil {
		return nil, err
	}
	stocks := make([]*domain.Stock, 0, len(models))
	for _, m := range models {
		stocks = append(stocks, m.toDomain())
	}
	return stocks, nil
}

func (r *stockRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&StockModel{}, "id = ?", id).Error
}

// priceRepository implements domain.PriceRepository
type priceRepository struct {
	db *gorm.DB
}

func NewPriceRepository(db *gorm.DB) domain.PriceRepository {
	return &priceRepository{db: db}
}

func (r *priceRepository) Create(ctx context.Context, p *domain.Price) error {
	return r.db.WithContext(ctx).Create(toPriceModel(p)).Error
}

func (r *priceRepository) GetLatest(ctx context.Context, symbol string) (*domain.Price, error) {
	var p PriceModel
	err := r.db.WithContext(ctx).Where("symbol = ?", symbol).Order("timestamp DESC").First(&p).Error
	if err != nil {
		return nil, err
	}
	return p.toDomain(), nil
}

func (r *priceRepository) GetBySymbolAndTime(ctx context.Context, symbol string, timestamp time.Time) (*domain.Price, error) {
	var p PriceModel
	err := r.db.WithContext(ctx).Where("symbol = ? AND timestamp <= ?", symbol, timestamp).Order("timestamp DESC").First(&p).Error
	if err != nil {
		return nil, err
	}
	return p.toDomain(), nil
}

func (r *priceRepository) ListBySymbol(ctx context.Context, symbol string, from, to time.Time) ([]*domain.Price, error) {
	var models []*PriceModel
	err := r.db.WithContext(ctx).Where("symbol = ? AND timestamp BETWEEN ? AND ?", symbol, from, to).Order("timestamp ASC").Find(&models).Error
	if err != nil {
		return nil, err
	}
	prices := make([]*domain.Price, 0, len(models))
	for _, m := range models {
		prices = append(prices, m.toDomain())
	}
	return prices, nil
}

func (r *priceRepository) BatchCreate(ctx context.Context, prices []*domain.Price) error {
	models := make([]*PriceModel, 0, len(prices))
	for _, p := range prices {
		models = append(models, toPriceModel(p))
	}
	return r.db.WithContext(ctx).Create(&models).Error
}

// candleRepository implements domain.CandleRepository
type candleRepository struct {
	db *gorm.DB
}

func NewCandleRepository(db *gorm.DB) domain.CandleRepository {
	return &candleRepository{db: db}
}

func (r *candleRepository) Create(ctx context.Context, c *domain.Candle) error {
	return r.db.WithContext(ctx).Create(toCandleModel(c)).Error
}

func (r *candleRepository) GetBySymbolAndInterval(ctx context.Context, symbol, interval string, from, to time.Time) ([]*domain.Candle, error) {
	var models []*CandleModel
	err := r.db.WithContext(ctx).Where("symbol = ? AND interval = ? AND timestamp BETWEEN ? AND ?", symbol, interval, from, to).Order("timestamp ASC").Find(&models).Error
	if err != nil {
		return nil, err
	}
	candles := make([]*domain.Candle, 0, len(models))
	for _, m := range models {
		candles = append(candles, m.toDomain())
	}
	return candles, nil
}

func (r *candleRepository) GetLatest(ctx context.Context, symbol, interval string) (*domain.Candle, error) {
	var c CandleModel
	err := r.db.WithContext(ctx).Where("symbol = ? AND interval = ?", symbol, interval).Order("timestamp DESC").First(&c).Error
	if err != nil {
		return nil, err
	}
	return c.toDomain(), nil
}

func (r *candleRepository) BatchCreate(ctx context.Context, candles []*domain.Candle) error {
	models := make([]*CandleModel, 0, len(candles))
	for _, c := range candles {
		models = append(models, toCandleModel(c))
	}
	return r.db.WithContext(ctx).Create(&models).Error
}

func (r *candleRepository) DeleteOlderThan(ctx context.Context, timestamp time.Time) error {
	return r.db.WithContext(ctx).Where("timestamp < ?", timestamp).Delete(&CandleModel{}).Error
}
