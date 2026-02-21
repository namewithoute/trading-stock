package account

import (
	"context"
	"time"
)

// AccountReadModel is the denormalised, query-optimised view of an account.
// It lives in the `account_read_models` table and is rebuilt by the Projector.
type AccountReadModel struct {
	ID          string      `json:"id"`
	UserID      string      `json:"user_id"`
	AccountType AccountType `json:"account_type"`
	Balance     float64     `json:"balance"`
	BuyingPower float64     `json:"buying_power"`
	Currency    string      `json:"currency"`
	Status      Status      `json:"status"`
	Version     int         `json:"version"` // Matches latest event version
	UpdatedAt   time.Time   `json:"updated_at"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Event Store Repository – write side
// ─────────────────────────────────────────────────────────────────────────────

// EventStore is the append-only event log for AccountAggregate.
type EventStore interface {
	// AppendEvents saves new events for an aggregate (within one DB transaction).
	// expectedVersion is used for optimistic concurrency: if the last stored version
	// doesn't match, an error is returned to prevent lost-update bugs.
	AppendEvents(ctx context.Context, aggregateID string, expectedVersion int, events []EventDescriptor) error

	// LoadEvents fetches all events for an aggregate ordered by version ASC.
	LoadEvents(ctx context.Context, aggregateID string) ([]EventDescriptor, error)

	// LoadEventsFrom fetches events starting from a specific version (used after Snapshot).
	LoadEventsFrom(ctx context.Context, aggregateID string, fromVersion int) ([]EventDescriptor, error)
}

// ─────────────────────────────────────────────────────────────────────────────
// Read Model Repository – query side
// ─────────────────────────────────────────────────────────────────────────────

// ReadModelRepository is the query interface for account read models.
type ReadModelRepository interface {
	// Upsert creates or replaces the read model for a given account.
	Upsert(ctx context.Context, rm *AccountReadModel) error

	// GetByID returns the read model for a single account.
	GetByID(ctx context.Context, id string) (*AccountReadModel, error)

	// GetByUserID returns all read models for a user.
	GetByUserID(ctx context.Context, userID string) ([]*AccountReadModel, error)
}
