package account

import (
	"context"
	"time"

	domain "trading-stock/internal/domain/account"
)

// ─────────────────────────────────────────────────────────────────────────────
// Infrastructure-layer Event Store types.
//
// These types do NOT belong in the domain because they are persistence
// implementation details (envelope format, serialisation, UUIDs, etc.).
// The domain only knows about DomainEvent (business facts) and the abstract
// Repository port.
// ─────────────────────────────────────────────────────────────────────────────

// EventDescriptor is the persistence envelope for a single domain event.
// It is stored as one row in `account_events` and published as one Kafka message.
type EventDescriptor struct {
	ID          string           `json:"id"`           // UUID per row
	AggregateID string           `json:"aggregate_id"` // Account ID
	EventType   domain.EventType `json:"event_type"`
	Payload     []byte           `json:"payload"` // JSON-serialised concrete event
	Version     int              `json:"version"` // Monotonically increasing per aggregate
	OccurredAt  time.Time        `json:"occurred_at"`
}

// EventStore is the append-only infrastructure interface for account events.
// It is implemented by eventStore (Postgres-backed) in event_store.go.
// The domain has no knowledge of this interface.
type EventStore interface {
	// AppendEvents persists new events with optimistic concurrency control.
	AppendEvents(ctx context.Context, aggregateID string, expectedVersion int, events []EventDescriptor) error

	// LoadEvents returns all events for an aggregate, ordered by version ASC.
	LoadEvents(ctx context.Context, aggregateID string) ([]EventDescriptor, error)

	// LoadEventsFrom returns events from a given version onwards (post-snapshot use).
	LoadEventsFrom(ctx context.Context, aggregateID string, fromVersion int) ([]EventDescriptor, error)

	// LoadAllDescriptors returns every event ordered by (aggregate_id, version) ASC.
	// Used by the Projector during startup catch-up rebuild.
	LoadAllDescriptors(ctx context.Context) ([]EventDescriptor, error)
}
