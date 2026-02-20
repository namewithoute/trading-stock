package order

import "context"

// Repository defines the interface for order data access
// This is implemented by the infrastructure layer (e.g., PostgreSQL)
type Repository interface {
	// Create creates a new order in the database
	Create(ctx context.Context, order *Order) error

	// GetByID retrieves an order by its ID
	GetByID(ctx context.Context, id string) (*Order, error)

	// Update updates an existing order
	Update(ctx context.Context, order *Order) error

	// Cancel cancels an order by setting its status to CANCELLED
	Cancel(ctx context.Context, id string) error

	// ListByUserID retrieves all orders for a specific user
	// Results are ordered by created_at DESC
	ListByUserID(ctx context.Context, userID string, limit, offset int) ([]*Order, error)

	ListOrdersByUserIDAndSymbolAndStatus(ctx context.Context, userID, symbol string, status string, limit, offset int) ([]*Order, error)

	// ListByStatus retrieves all orders with a specific status
	ListByStatus(ctx context.Context, status Status, limit, offset int) ([]*Order, error)

	// ListBySymbol retrieves all orders for a specific symbol
	ListBySymbol(ctx context.Context, symbol string, limit, offset int) ([]*Order, error)

	// ListPendingOrdersByUserAndSymbol retrieves pending orders for a user and symbol
	// This is useful for checking existing orders before creating new ones
	ListPendingOrdersByUserAndSymbol(ctx context.Context, userID, symbol string) ([]*Order, error)

	// CountByUserID returns the total number of orders for a user
	CountByUserID(ctx context.Context, userID string) (int64, error)

	// Delete permanently deletes an order (use with caution)
	Delete(ctx context.Context, id string) error
}
