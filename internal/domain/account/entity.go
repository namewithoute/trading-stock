package account

import "time"

// Account represents a trading account entity
// A user can have multiple accounts (e.g., cash account, margin account)
type Account struct {
	ID          string
	UserID      string
	AccountType AccountType

	// Balance information
	// Balance     float64
	// BuyingPower float64
	// Currency    string

	Money Money

	// Margin account specific (only for margin accounts)
	MarginUsed      float64
	MarginAvailable float64

	// Status
	Status Status

	// Timestamps
	CreatedAt time.Time
	UpdatedAt time.Time
}

// IsActive checks if the account is active
func (a *Account) IsActive() bool {
	return a.Status == StatusActive
}

// CanTrade checks if the account can perform trading operations
func (a *Account) CanTrade() bool {
	return a.IsActive() && a.Money.BuyingPower > 0
}

// HasSufficientBalance checks if the account has sufficient balance for a purchase
func (a *Account) HasSufficientBalance(amount float64) bool {
	return a.Money.BuyingPower >= amount
}

// Deposit adds funds to the account
func (a *Account) Deposit(amount float64) {
	a.Money.Balance += amount
	a.Money.BuyingPower += amount
	a.UpdatedAt = time.Now()
}

// Withdraw removes funds from the account
func (a *Account) Withdraw(amount float64) error {
	if a.Money.Balance < amount {
		return ErrInsufficientBalance
	}
	a.Money.Balance -= amount
	a.Money.BuyingPower -= amount
	a.UpdatedAt = time.Now()
	return nil
}

// ReserveFunds reserves funds for an order (reduces buying power)
func (a *Account) ReserveFunds(amount float64) error {
	if a.Money.BuyingPower < amount {
		return ErrInsufficientBuyingPower
	}
	a.Money.BuyingPower -= amount
	a.UpdatedAt = time.Now()
	return nil
}

// ReleaseFunds releases reserved funds (increases buying power)
func (a *Account) ReleaseFunds(amount float64) {
	a.Money.BuyingPower += amount
	a.UpdatedAt = time.Now()
}
