package account

import "errors"

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
)
