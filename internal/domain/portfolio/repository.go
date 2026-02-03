package portfolio

import "context"

// Repository defines the interface for portfolio data access
type Repository interface {
	// Create creates a new position
	Create(ctx context.Context, position *Position) error

	// GetByID retrieves a position by its ID
	GetByID(ctx context.Context, id string) (*Position, error)

	// GetByAccountAndSymbol retrieves a position by account ID and symbol
	GetByAccountAndSymbol(ctx context.Context, accountID, symbol string) (*Position, error)

	// ListByAccountID retrieves all positions for a specific account
	ListByAccountID(ctx context.Context, accountID string) ([]*Position, error)

	// ListByUserID retrieves all positions for a specific user
	ListByUserID(ctx context.Context, userID string) ([]*Position, error)

	// Update updates an existing position
	Update(ctx context.Context, position *Position) error

	// UpdateCurrentPrice updates the current price and recalculates P&L
	UpdateCurrentPrice(ctx context.Context, id string, price float64) error

	// AddQuantity adds quantity to a position
	AddQuantity(ctx context.Context, id string, quantity int, price float64) error

	// ReduceQuantity reduces quantity from a position
	ReduceQuantity(ctx context.Context, id string, quantity int) error

	// Delete deletes a position (only if quantity is 0)
	Delete(ctx context.Context, id string) error

	// GetTotalValue calculates the total portfolio value for a user
	GetTotalValue(ctx context.Context, userID string) (float64, error)

	// GetTotalUnrealizedPnL calculates the total unrealized P&L for a user
	GetTotalUnrealizedPnL(ctx context.Context, userID string) (float64, error)

	// Count returns the total number of positions
	Count(ctx context.Context) (int64, error)

	// CountByUserID returns the number of positions for a user
	CountByUserID(ctx context.Context, userID string) (int64, error)
}
