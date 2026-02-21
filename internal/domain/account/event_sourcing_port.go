package account

import "context"

// EventSourcingServicePort is the domain-level port that the Application layer
// uses to load and save aggregates. It hides the Kafka/Postgres details from
// the business logic (Dependency Inversion Principle).
type EventSourcingServicePort interface {
	// Load rehydrates an AccountAggregate from its event history.
	Load(ctx context.Context, aggregateID string) (*AccountAggregate, error)

	// Save persists uncommitted events and publishes them to Kafka.
	Save(ctx context.Context, agg *AccountAggregate) error
}
