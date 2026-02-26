package order

import (
	"context"
	"time"

	domain "trading-stock/internal/domain/order"

	"gorm.io/gorm"
)

// ─────────────────────────────────────────────────────────────────────────────
// Infrastructure-layer Event Store types.
//
// These types do NOT belong in the domain because they are persistence
// implementation details (envelope format, serialisation, UUIDs, etc.).
// ─────────────────────────────────────────────────────────────────────────────

// EventDescriptor is the persistence envelope for a single domain event.
// Stored as one row in `order_events` and published as one Kafka message.
type EventDescriptor struct {
	ID          string           `json:"id"`
	AggregateID string           `json:"aggregate_id"`
	EventType   domain.EventType `json:"event_type"`
	Payload     []byte           `json:"payload"` // JSON-serialised concrete event
	Version     int              `json:"version"` // Monotonically increasing per aggregate
	OccurredAt  time.Time        `json:"occurred_at"`
}

// EventStore is the append-only infrastructure interface for order events.
// Implemented by eventStore (Postgres-backed) in event_store.go.
type EventStore interface {
	// AppendEvents persists new events with optimistic concurrency control.
	AppendEvents(ctx context.Context, aggregateID string, expectedVersion int, events []EventDescriptor) error

	// AppendEventsWithHook persists events inside a transaction and then
	// calls afterFn(tx) in the same transaction so callers can insert outbox
	// rows atomically alongside domain events.
	AppendEventsWithHook(ctx context.Context, aggregateID string, expectedVersion int, events []EventDescriptor, afterFn func(tx *gorm.DB) error) error

	// LoadEvents returns all events for an aggregate, ordered by version ASC.
	LoadEvents(ctx context.Context, aggregateID string) ([]EventDescriptor, error)

	// LoadAllDescriptors returns every event ordered by (aggregate_id, version) ASC.
	// Used by the Projector during startup catch-up rebuild.
	LoadAllDescriptors(ctx context.Context) ([]EventDescriptor, error)
}
