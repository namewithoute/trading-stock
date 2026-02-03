package account

import "context"

// Repository defines the interface for account data access
type Repository interface {
	// Create creates a new account
	Create(ctx context.Context, account *Account) error

	// GetByID retrieves an account by its ID
	GetByID(ctx context.Context, id string) (*Account, error)

	// GetByUserID retrieves all accounts for a specific user
	GetByUserID(ctx context.Context, userID string) ([]*Account, error)

	// GetPrimaryAccount retrieves the primary (first) account for a user
	GetPrimaryAccount(ctx context.Context, userID string) (*Account, error)

	// Update updates an existing account
	Update(ctx context.Context, account *Account) error

	// UpdateBalance updates the account balance and buying power
	UpdateBalance(ctx context.Context, id string, balance, buyingPower float64) error

	// UpdateStatus updates the account status
	UpdateStatus(ctx context.Context, id string, status Status) error

	// Deposit adds funds to an account
	Deposit(ctx context.Context, id string, amount float64) error

	// Withdraw removes funds from an account
	Withdraw(ctx context.Context, id string, amount float64) error

	// ReserveFunds reserves funds for an order
	ReserveFunds(ctx context.Context, id string, amount float64) error

	// ReleaseFunds releases reserved funds
	ReleaseFunds(ctx context.Context, id string, amount float64) error

	// Delete soft deletes an account (sets status to CLOSED)
	Delete(ctx context.Context, id string) error

	// List retrieves all accounts with pagination
	List(ctx context.Context, limit, offset int) ([]*Account, error)

	// Count returns the total number of accounts
	Count(ctx context.Context) (int64, error)

	// CountByUserID returns the number of accounts for a user
	CountByUserID(ctx context.Context, userID string) (int64, error)
}
