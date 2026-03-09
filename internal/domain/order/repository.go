package order

import "context"

// ─────────────────────────────────────────────────────────────────────────────
// Outbound Ports — interfaces the domain requires from the outside world.
//
// Following Ports & Adapters (Hexagonal Architecture):
//   - The domain defines WHAT it needs.
//   - The infrastructure layer decides HOW to implement it (Event Sourcing).
// ─────────────────────────────────────────────────────────────────────────────

// Repository is the aggregate root repository (write side).
// The infrastructure implements it using Event Sourcing (append-only event log).
type Repository interface {
	// Load rehydrates an OrderAggregate by replaying its full event history.
	Load(ctx context.Context, id string) (*OrderAggregate, error)

	// Save persists all uncommitted events produced by the last command.
	Save(ctx context.Context, agg *OrderAggregate) error
}

// ReadModelRepository is the query-side store for pre-projected order views.
// Written by the Projector; read by the application query handlers.
// Commands never touch it directly.
type ReadModelRepository interface {
	// Upsert creates or replaces the read model for a given order.
	Upsert(ctx context.Context, rm *OrderReadModel) error

	// GetByID returns the read model for a single order.
	GetByID(ctx context.Context, id string) (*OrderReadModel, error)

	// ListByUserID retrieves all orders for a specific user, ordered by created_at DESC.
	ListByUserID(ctx context.Context, userID string, limit, offset int) ([]*OrderReadModel, error)

	// ListByUserIDAndFilter filters by optional symbol and status.
	ListByUserIDAndFilter(ctx context.Context, userID, symbol, status string, limit, offset int) ([]*OrderReadModel, error)

	// ListAll retrieves all orders across all users (admin use).
	ListAll(ctx context.Context, limit, offset int) ([]*OrderReadModel, error)

	// CountAll returns the total number of orders across all users (admin use).
	CountAll(ctx context.Context) (int64, error)
}
