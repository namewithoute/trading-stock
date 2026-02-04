package account

import (
	"time"

	domain "trading-stock/internal/domain/account"
)

// AccountModel is the GORM persistence model for accounts.
// NOTE: Keep persistence concerns (GORM tags, table name) out of Domain.
type AccountModel struct {
	ID          string `gorm:"primaryKey;type:uuid"`
	UserID      string `gorm:"type:uuid;index;not null"`
	AccountType string `gorm:"type:varchar(20);not null"`

	Balance     float64 `gorm:"type:decimal(20,2);not null;default:0"`
	BuyingPower float64 `gorm:"type:decimal(20,2);not null;default:0"`
	Currency    string  `gorm:"type:varchar(3);not null;default:'USD'"`

	MarginUsed      float64 `gorm:"type:decimal(20,2);default:0"`
	MarginAvailable float64 `gorm:"type:decimal(20,2);default:0"`

	Status string `gorm:"type:varchar(20);not null"`

	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (AccountModel) TableName() string { return "accounts" }

func toAccountModel(a *domain.Account) *AccountModel {
	if a == nil {
		return nil
	}
	return &AccountModel{
		ID:              a.ID,
		UserID:          a.UserID,
		AccountType:     string(a.AccountType),
		Balance:         a.Balance,
		BuyingPower:     a.BuyingPower,
		Currency:        a.Currency,
		MarginUsed:      a.MarginUsed,
		MarginAvailable: a.MarginAvailable,
		Status:          string(a.Status),
		CreatedAt:       a.CreatedAt,
		UpdatedAt:       a.UpdatedAt,
	}
}

func (m *AccountModel) toDomain() *domain.Account {
	if m == nil {
		return nil
	}
	return &domain.Account{
		ID:              m.ID,
		UserID:          m.UserID,
		AccountType:     domain.AccountType(m.AccountType),
		Balance:         m.Balance,
		BuyingPower:     m.BuyingPower,
		Currency:        m.Currency,
		MarginUsed:      m.MarginUsed,
		MarginAvailable: m.MarginAvailable,
		Status:          domain.Status(m.Status),
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}
}
