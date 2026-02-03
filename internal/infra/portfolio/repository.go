package portfolio

import (
	"context"
	"errors"

	"trading-stock/internal/domain/portfolio"

	"gorm.io/gorm"
)

// portfolioRepository implements domain.PortfolioRepository interface
type portfolioRepository struct {
	db *gorm.DB
}

// NewPortfolioRepository creates a new portfolio repository
func NewPortfolioRepository(db *gorm.DB) portfolio.Repository {
	return &portfolioRepository{db: db}
}

// Create creates a new position
func (r *portfolioRepository) Create(ctx context.Context, pos *portfolio.Position) error {
	return r.db.WithContext(ctx).Create(pos).Error
}

// GetByID retrieves a position by its ID
func (r *portfolioRepository) GetByID(ctx context.Context, id string) (*portfolio.Position, error) {
	var pos portfolio.Position
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&pos).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("position not found")
		}
		return nil, err
	}
	return &pos, nil
}

// GetByAccountAndSymbol retrieves a position by account ID and symbol
func (r *portfolioRepository) GetByAccountAndSymbol(ctx context.Context, accountID, symbol string) (*portfolio.Position, error) {
	var pos portfolio.Position
	err := r.db.WithContext(ctx).Where("account_id = ? AND symbol = ?", accountID, symbol).First(&pos).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &pos, nil
}

// ListByAccountID retrieves all positions for a specific account
func (r *portfolioRepository) ListByAccountID(ctx context.Context, accountID string) ([]*portfolio.Position, error) {
	var positions []*portfolio.Position
	err := r.db.WithContext(ctx).Where("account_id = ? AND quantity > 0", accountID).Find(&positions).Error
	return positions, err
}

// ListByUserID retrieves all positions for a specific user
func (r *portfolioRepository) ListByUserID(ctx context.Context, userID string) ([]*portfolio.Position, error) {
	var positions []*portfolio.Position
	err := r.db.WithContext(ctx).Where("user_id = ? AND quantity > 0", userID).Find(&positions).Error
	return positions, err
}

// Update updates an existing position
func (r *portfolioRepository) Update(ctx context.Context, pos *portfolio.Position) error {
	return r.db.WithContext(ctx).Save(pos).Error
}

// UpdateCurrentPrice updates the current price and recalculates P&L
func (r *portfolioRepository) UpdateCurrentPrice(ctx context.Context, id string, price float64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var pos portfolio.Position
		if err := tx.Where("id = ?", id).First(&pos).Error; err != nil {
			return err
		}

		pos.UpdateCurrentPrice(price)
		return tx.Save(&pos).Error
	})
}

// AddQuantity adds quantity to a position
func (r *portfolioRepository) AddQuantity(ctx context.Context, id string, quantity int, price float64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var pos portfolio.Position
		if err := tx.Where("id = ?", id).First(&pos).Error; err != nil {
			return err
		}

		pos.AddQuantity(quantity, price)
		return tx.Save(&pos).Error
	})
}

// ReduceQuantity reduces quantity from a position
func (r *portfolioRepository) ReduceQuantity(ctx context.Context, id string, quantity int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var pos portfolio.Position
		if err := tx.Where("id = ?", id).First(&pos).Error; err != nil {
			return err
		}

		if err := pos.ReduceQuantity(quantity); err != nil {
			return err
		}

		return tx.Save(&pos).Error
	})
}

// Delete deletes a position (only if quantity is 0)
func (r *portfolioRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var pos portfolio.Position
		if err := tx.Where("id = ?", id).First(&pos).Error; err != nil {
			return err
		}

		if !pos.IsClosed() {
			return errors.New("cannot delete open position")
		}

		return tx.Delete(&pos).Error
	})
}

// GetTotalValue calculates the total portfolio value for a user
func (r *portfolioRepository) GetTotalValue(ctx context.Context, userID string) (float64, error) {
	var totalValue float64
	err := r.db.WithContext(ctx).
		Model(&portfolio.Position{}).
		Where("user_id = ?", userID).
		Select("COALESCE(SUM(quantity * current_price), 0)").
		Scan(&totalValue).Error
	return totalValue, err
}

// GetTotalUnrealizedPnL calculates the total unrealized P&L for a user
func (r *portfolioRepository) GetTotalUnrealizedPnL(ctx context.Context, userID string) (float64, error) {
	var totalPnL float64
	err := r.db.WithContext(ctx).
		Model(&portfolio.Position{}).
		Where("user_id = ?", userID).
		Select("COALESCE(SUM(unrealized_pnl), 0)").
		Scan(&totalPnL).Error
	return totalPnL, err
}

// Count returns the total number of positions
func (r *portfolioRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&portfolio.Position{}).Count(&count).Error
	return count, err
}

// CountByUserID returns the number of positions for a user
func (r *portfolioRepository) CountByUserID(ctx context.Context, userID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&portfolio.Position{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return count, err
}
