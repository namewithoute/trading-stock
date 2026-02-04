package portfolio

import (
	"context"
	"errors"

	domain "trading-stock/internal/domain/portfolio"

	"gorm.io/gorm"
)

// portfolioRepository implements domain.PortfolioRepository interface
type portfolioRepository struct {
	db *gorm.DB
}

// NewPortfolioRepository creates a new portfolio repository
func NewPortfolioRepository(db *gorm.DB) domain.Repository {
	return &portfolioRepository{db: db}
}

// Create creates a new position
func (r *portfolioRepository) Create(ctx context.Context, pos *domain.Position) error {
	return r.db.WithContext(ctx).Create(toPositionModel(pos)).Error
}

// GetByID retrieves a position by its ID
func (r *portfolioRepository) GetByID(ctx context.Context, id string) (*domain.Position, error) {
	var pos PositionModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&pos).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("position not found")
		}
		return nil, err
	}
	return pos.toDomain(), nil
}

// GetByAccountAndSymbol retrieves a position by account ID and symbol
func (r *portfolioRepository) GetByAccountAndSymbol(ctx context.Context, accountID, symbol string) (*domain.Position, error) {
	var pos PositionModel
	err := r.db.WithContext(ctx).Where("account_id = ? AND symbol = ?", accountID, symbol).First(&pos).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return pos.toDomain(), nil
}

// ListByAccountID retrieves all positions for a specific account
func (r *portfolioRepository) ListByAccountID(ctx context.Context, accountID string) ([]*domain.Position, error) {
	var models []*PositionModel
	err := r.db.WithContext(ctx).Where("account_id = ? AND quantity > 0", accountID).Find(&models).Error
	if err != nil {
		return nil, err
	}

	positions := make([]*domain.Position, 0, len(models))
	for _, m := range models {
		positions = append(positions, m.toDomain())
	}
	return positions, nil
}

// ListByUserID retrieves all positions for a specific user
func (r *portfolioRepository) ListByUserID(ctx context.Context, userID string) ([]*domain.Position, error) {
	var models []*PositionModel
	err := r.db.WithContext(ctx).Where("user_id = ? AND quantity > 0", userID).Find(&models).Error
	if err != nil {
		return nil, err
	}

	positions := make([]*domain.Position, 0, len(models))
	for _, m := range models {
		positions = append(positions, m.toDomain())
	}
	return positions, nil
}

// Update updates an existing position
func (r *portfolioRepository) Update(ctx context.Context, pos *domain.Position) error {
	return r.db.WithContext(ctx).Save(toPositionModel(pos)).Error
}

// UpdateCurrentPrice updates the current price and recalculates P&L
func (r *portfolioRepository) UpdateCurrentPrice(ctx context.Context, id string, price float64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var m PositionModel
		if err := tx.Where("id = ?", id).First(&m).Error; err != nil {
			return err
		}

		pos := m.toDomain()
		if pos == nil {
			return errors.New("position not found")
		}

		pos.UpdateCurrentPrice(price)
		return tx.Save(toPositionModel(pos)).Error
	})
}

// AddQuantity adds quantity to a position
func (r *portfolioRepository) AddQuantity(ctx context.Context, id string, quantity int, price float64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var m PositionModel
		if err := tx.Where("id = ?", id).First(&m).Error; err != nil {
			return err
		}

		pos := m.toDomain()
		if pos == nil {
			return errors.New("position not found")
		}

		pos.AddQuantity(quantity, price)
		return tx.Save(toPositionModel(pos)).Error
	})
}

// ReduceQuantity reduces quantity from a position
func (r *portfolioRepository) ReduceQuantity(ctx context.Context, id string, quantity int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var m PositionModel
		if err := tx.Where("id = ?", id).First(&m).Error; err != nil {
			return err
		}

		pos := m.toDomain()
		if pos == nil {
			return errors.New("position not found")
		}

		if err := pos.ReduceQuantity(quantity); err != nil {
			return err
		}

		return tx.Save(toPositionModel(pos)).Error
	})
}

// Delete deletes a position (only if quantity is 0)
func (r *portfolioRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var m PositionModel
		if err := tx.Where("id = ?", id).First(&m).Error; err != nil {
			return err
		}

		pos := m.toDomain()
		if pos == nil {
			return errors.New("position not found")
		}

		if !pos.IsClosed() {
			return errors.New("cannot delete open position")
		}

		return tx.Delete(&PositionModel{}, "id = ?", id).Error
	})
}

// GetTotalValue calculates the total portfolio value for a user
func (r *portfolioRepository) GetTotalValue(ctx context.Context, userID string) (float64, error) {
	var totalValue float64
	err := r.db.WithContext(ctx).
		Model(&PositionModel{}).
		Where("user_id = ?", userID).
		Select("COALESCE(SUM(quantity * current_price), 0)").
		Scan(&totalValue).Error
	return totalValue, err
}

// GetTotalUnrealizedPnL calculates the total unrealized P&L for a user
func (r *portfolioRepository) GetTotalUnrealizedPnL(ctx context.Context, userID string) (float64, error) {
	var totalPnL float64
	err := r.db.WithContext(ctx).
		Model(&PositionModel{}).
		Where("user_id = ?", userID).
		Select("COALESCE(SUM(unrealized_pnl), 0)").
		Scan(&totalPnL).Error
	return totalPnL, err
}

// Count returns the total number of positions
func (r *portfolioRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&PositionModel{}).Count(&count).Error
	return count, err
}

// CountByUserID returns the number of positions for a user
func (r *portfolioRepository) CountByUserID(ctx context.Context, userID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&PositionModel{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return count, err
}
