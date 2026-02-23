package account

import "context"

// ─────────────────────────────────────────────────────────────────────────────
// Outbound Ports — interfaces the domain requires from the outside world.
//
// Following Ports & Adapters (Hexagonal Architecture):
//   - The domain defines WHAT it needs.
//   - The infrastructure layer decides HOW to implement it.
//
// The domain has no knowledge of Event Sourcing, Postgres, Kafka, or any
// persistence mechanism. That is intentional.
// ─────────────────────────────────────────────────────────────────────────────

// Repository is the aggregate root repository.
// The infrastructure layer is free to implement it with Event Sourcing,
// a relational table, a document store, or any other mechanism.
type Repository interface {
	// Load rehydrates an AccountAggregate from its full history.
	Load(ctx context.Context, id string) (*AccountAggregate, error)

	// Save persists all uncommitted changes produced by the last command.
	Save(ctx context.Context, agg *AccountAggregate) error
}

// ReadModelRepository is the query-side store for pre-projected account views.
// Written by the infrastructure projector; read by the application query handlers.
// Commands never touch it directly.
type ReadModelRepository interface {
	// Upsert creates or replaces the read model for a given account.
	Upsert(ctx context.Context, rm *AccountReadModel) error

	// GetByID returns the read model for a single account.
	GetByID(ctx context.Context, id string) (*AccountReadModel, error)

	// GetByUserID returns all read models belonging to a user.
	GetByUserID(ctx context.Context, userID string) ([]*AccountReadModel, error)
}
