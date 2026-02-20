package account

import (
	"context"
	"errors"

	domain "trading-stock/internal/domain/account"

	"gorm.io/gorm"
)

// accountRepository implements domain.AccountRepository interface
type accountRepository struct {
	db *gorm.DB
}

// NewAccountRepository creates a new account repository
func NewAccountRepository(db *gorm.DB) domain.Repository {
	return &accountRepository{db: db}
}

// Create creates a new account
func (r *accountRepository) Create(ctx context.Context, acc *domain.Account) error {
	return r.db.WithContext(ctx).Create(toAccountModel(acc)).Error
}

// GetByID retrieves an account by its ID
func (r *accountRepository) GetByID(ctx context.Context, id string) (*domain.Account, error) {
	var acc AccountModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&acc).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("account not found")
		}
		return nil, err
	}
	return acc.toDomain(), nil
}

// GetByUserID retrieves all accounts for a specific user
func (r *accountRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.Account, error) {
	var models []*AccountModel
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	accounts := make([]*domain.Account, 0, len(models))
	for _, m := range models {
		accounts = append(accounts, m.toDomain())
	}
	return accounts, nil
}

// GetPrimaryAccount retrieves the primary (first) account for a user
func (r *accountRepository) GetPrimaryAccount(ctx context.Context, userID string) (*domain.Account, error) {
	var acc AccountModel
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at ASC").
		First(&acc).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("no account found for user")
		}
		return nil, err
	}
	return acc.toDomain(), nil
}

// Update updates an existing account
func (r *accountRepository) Update(ctx context.Context, acc *domain.Account) error {
	return r.db.WithContext(ctx).Save(toAccountModel(acc)).Error
}

// UpdateBalance updates the account balance and buying power
func (r *accountRepository) UpdateBalance(ctx context.Context, id string, balance, buyingPower float64) error {
	return r.db.WithContext(ctx).
		Model(&AccountModel{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"balance":      balance,
			"buying_power": buyingPower,
		}).Error
}

// UpdateStatus updates the account status
func (r *accountRepository) UpdateStatus(ctx context.Context, id string, status domain.Status) error {
	return r.db.WithContext(ctx).
		Model(&AccountModel{}).
		Where("id = ?", id).
		Update("status", status).
		Error
}

// Deposit adds funds to an account (atomic operation)
func (r *accountRepository) Deposit(ctx context.Context, id string, amount float64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.Model(&AccountModel{}).
			Where("id = ?", id).
			Updates(map[string]interface{}{
				"balance":      gorm.Expr("balance + ?", amount),
				"buying_power": gorm.Expr("buying_power + ?", amount),
			}).Error
	})
}

// Withdraw removes funds from an account (atomic operation)
func (r *accountRepository) Withdraw(ctx context.Context, id string, amount float64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Check if sufficient balance
		var acc AccountModel
		if err := tx.Where("id = ?", id).First(&acc).Error; err != nil {
			return err
		}

		if acc.Balance < amount {
			return errors.New("insufficient balance")
		}

		return tx.Model(&AccountModel{}).
			Where("id = ?", id).
			Updates(map[string]interface{}{
				"balance":      gorm.Expr("balance - ?", amount),
				"buying_power": gorm.Expr("buying_power - ?", amount),
			}).Error
	})
}

// ReserveFunds reserves funds for an order (reduces buying power)
func (r *accountRepository) ReserveFunds(ctx context.Context, id string, amount float64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Check if sufficient buying power
		var acc AccountModel
		if err := tx.Where("id = ?", id).First(&acc).Error; err != nil {
			return err
		}

		if acc.BuyingPower < amount {
			return errors.New("insufficient buying power")
		}

		return tx.Model(&AccountModel{}).
			Where("id = ?", id).
			Update("buying_power", gorm.Expr("buying_power - ?", amount)).
			Error
	})
}

// ReleaseFunds releases reserved funds (increases buying power)
func (r *accountRepository) ReleaseFunds(ctx context.Context, id string, amount float64) error {
	return r.db.WithContext(ctx).
		Model(&AccountModel{}).
		Where("id = ?", id).
		Update("buying_power", gorm.Expr("buying_power + ?", amount)).
		Error
}

// Delete soft deletes an account (sets status to CLOSED)
func (r *accountRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&AccountModel{}).
		Where("id = ?", id).
		Update("status", domain.StatusClosed).
		Error
}

// List retrieves all accounts with pagination
func (r *accountRepository) List(ctx context.Context, limit, offset int) ([]*domain.Account, error) {
	var models []*AccountModel
	err := r.db.WithContext(ctx).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	accounts := make([]*domain.Account, 0, len(models))
	for _, m := range models {
		accounts = append(accounts, m.toDomain())
	}
	return accounts, nil
}

// Count returns the total number of accounts
func (r *accountRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&AccountModel{}).Count(&count).Error
	return count, err
}

// CountByUserID returns the number of accounts for a user
func (r *accountRepository) CountByUserID(ctx context.Context, userID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&AccountModel{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return count, err
}
