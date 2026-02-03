package account

import "time"

// Account represents a trading account entity
// A user can have multiple accounts (e.g., cash account, margin account)
type Account struct {
	ID          string      `json:"id" gorm:"primaryKey;type:uuid"`
	UserID      string      `json:"user_id" gorm:"type:uuid;index;not null"`
	AccountType AccountType `json:"account_type" gorm:"type:varchar(20);not null"`

	// Balance information
	Balance     float64 `json:"balance" gorm:"type:decimal(20,2);not null;default:0"`
	BuyingPower float64 `json:"buying_power" gorm:"type:decimal(20,2);not null;default:0"`
	Currency    string  `json:"currency" gorm:"type:varchar(3);not null;default:'USD'"`

	// Margin account specific (only for margin accounts)
	MarginUsed      float64 `json:"margin_used,omitempty" gorm:"type:decimal(20,2);default:0"`
	MarginAvailable float64 `json:"margin_available,omitempty" gorm:"type:decimal(20,2);default:0"`

	// Status
	Status Status `json:"status" gorm:"type:varchar(20);not null"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt time.Time `json:"updated_at" gorm:"not null"`
}

// TableName specifies the table name for GORM
func (Account) TableName() string {
	return "accounts"
}

// IsActive checks if the account is active
func (a *Account) IsActive() bool {
	return a.Status == StatusActive
}

// CanTrade checks if the account can perform trading operations
func (a *Account) CanTrade() bool {
	return a.IsActive() && a.BuyingPower > 0
}

// HasSufficientBalance checks if the account has sufficient balance for a purchase
func (a *Account) HasSufficientBalance(amount float64) bool {
	return a.BuyingPower >= amount
}

// Deposit adds funds to the account
func (a *Account) Deposit(amount float64) {
	a.Balance += amount
	a.BuyingPower += amount
	a.UpdatedAt = time.Now()
}

// Withdraw removes funds from the account
func (a *Account) Withdraw(amount float64) error {
	if a.Balance < amount {
		return ErrInsufficientBalance
	}
	a.Balance -= amount
	a.BuyingPower -= amount
	a.UpdatedAt = time.Now()
	return nil
}

// ReserveFunds reserves funds for an order (reduces buying power)
func (a *Account) ReserveFunds(amount float64) error {
	if a.BuyingPower < amount {
		return ErrInsufficientBuyingPower
	}
	a.BuyingPower -= amount
	a.UpdatedAt = time.Now()
	return nil
}

// ReleaseFunds releases reserved funds (increases buying power)
func (a *Account) ReleaseFunds(amount float64) {
	a.BuyingPower += amount
	a.UpdatedAt = time.Now()
}
