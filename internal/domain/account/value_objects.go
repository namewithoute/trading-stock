package account

import (
	"errors"
	"fmt"

	"github.com/cockroachdb/apd/v3"
)

// decCtx is the shared arithmetic context for all Money operations.
// Precision 19 covers amounts up to 9_999_999_999_999_999.9999 (19 significant digits).
var decCtx = apd.BaseContext.WithPrecision(19)

// AccountType represents the type of trading account
type AccountType string

const (
	TypeCash   AccountType = "CASH"   // Cash account - no margin trading
	TypeMargin AccountType = "MARGIN" // Margin account - allows borrowing
)

// IsValid checks if the account type is valid
func (t AccountType) IsValid() bool {
	return t == TypeCash || t == TypeMargin
}

// String returns the string representation of AccountType
func (t AccountType) String() string {
	return string(t)
}

// Status represents the account status
type Status string

const (
	StatusActive  Status = "ACTIVE"  // Account is active and can trade
	StatusFrozen  Status = "FROZEN"  // Account is frozen (cannot trade, but can view)
	StatusClosed  Status = "CLOSED"  // Account is permanently closed
	StatusPending Status = "PENDING" // Account is pending approval
)

func (s Status) Compare(status Status) bool {
	return s == status
}

// IsValid checks if the status is valid
func (s Status) IsValid() bool {
	switch s {
	case StatusActive, StatusFrozen, StatusClosed, StatusPending:
		return true
	default:
		return false
	}
}

// String returns the string representation of Status
func (s Status) String() string {
	return string(s)
}

// Domain errors
var (
	ErrInsufficientBalance     = errors.New("insufficient balance")
	ErrInsufficientBuyingPower = errors.New("insufficient buying power")
	ErrAccountFrozen           = errors.New("account is frozen")
	ErrAccountClosed           = errors.New("account is closed")
	ErrInvalidAccountType      = errors.New("invalid account type")
	ErrAccountNotFound         = errors.New("account not found")
	ErrAccountAlreadyExists    = errors.New("account already exists")
	ErrAccountNotActive        = errors.New("account is not active")
	ErrInvalidAmount           = errors.New("invalid amount: must be greater than zero")
	ErrInvalidUserID           = errors.New("invalid user ID")
	ErrInvalidCurrency         = errors.New("invalid currency")
	ErrAccountNotFrozen        = errors.New("account is not frozen")
)

type Money struct {
	Balance     apd.Decimal
	BuyingPower apd.Decimal
	Currency    string
}

func NewMoney(balance, buyingPower apd.Decimal, currency string) Money {
	return Money{
		Balance:     balance,
		BuyingPower: buyingPower,
		Currency:    currency,
	}
}

func (m Money) IsValid() bool {
	return m.Balance.Sign() >= 0 && m.BuyingPower.Sign() >= 0 && m.Currency != ""
}

func (m Money) StringBalance() string {
	return fmt.Sprintf("%s %s", m.Balance.String(), m.Currency)
}

func (m Money) StringBuyingPower() string {
	return fmt.Sprintf("%s %s", m.BuyingPower.String(), m.Currency)
}
