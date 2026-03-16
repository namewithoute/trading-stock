package account

import (
	"context"
	"encoding/json"
	"fmt"

	domain "trading-stock/internal/domain/account"
	"trading-stock/internal/infrastructure/outbox"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// KafkaTopicAccountEvents is the Kafka topic for account domain events.
const KafkaTopicAccountEvents = "account.events"

// EventSourcingService implements domain.Repository.
//
// Responsibilities:
//  1. Load       – deserialise + replay all events from EventStore → Aggregate
//  2. Save       – serialise new events → AppendEvents (Postgres) + outbox rows
//
// Wire rule: only wire.go may construct this struct. All other layers receive
// it through the domain.Repository interface.
type EventSourcingService struct {
	eventStore EventStore
	logger     *zap.Logger
}

// NewEventSourcingService constructs the service.
func NewEventSourcingService(
	eventStore EventStore,
	logger *zap.Logger,
) *EventSourcingService {
	return &EventSourcingService{
		eventStore: eventStore,
		logger:     logger,
	}
}

// ─── Load ────────────────────────────────────────────────────────────────────

// Load reconstructs an AccountAggregate from the full event history.
func (s *EventSourcingService) Load(ctx context.Context, aggregateID string) (*domain.AccountAggregate, error) {
	descriptors, err := s.eventStore.LoadEvents(ctx, aggregateID)
	if err != nil {
		return nil, fmt.Errorf("eventStore.LoadEvents(%s): %w", aggregateID, err)
	}
	if len(descriptors) == 0 {
		return nil, domain.ErrAccountNotFound
	}

	events := make([]domain.DomainEvent, 0, len(descriptors))
	for _, d := range descriptors {
		ev, err := DeserialiseEvent(d)
		if err != nil {
			return nil, fmt.Errorf("deserialise event %s v%d: %w", d.EventType, d.Version, err)
		}
		events = append(events, ev)
	}

	return domain.RehydrateAccount(events), nil
}

// ─── Save ────────────────────────────────────────────────────────────────────

// Save persists uncommitted aggregate events and writes Kafka-bound outbox rows.
//
// Guarantees:
//   - EventStore rows and outbox rows commit atomically in the same DB transaction.
//   - Kafka delivery is delegated to the outbox transport (e.g. Debezium).
func (s *EventSourcingService) Save(ctx context.Context, agg *domain.AccountAggregate) error {
	changes := agg.UncommittedEvents()
	if len(changes) == 0 {
		return nil
	}

	// ── Step 1: Serialise → EventDescriptors ──────────────────────────
	// Version before these new events = agg.Version - len(changes)
	baseVersion := agg.Version - len(changes)
	descriptors := make([]EventDescriptor, 0, len(changes))

	for i, ev := range changes {
		payload, err := json.Marshal(ev)
		if err != nil {
			return fmt.Errorf("marshal event %s: %w", ev.GetEventType(), err)
		}
		descriptors = append(descriptors, EventDescriptor{
			ID:          uuid.New().String(),
			AggregateID: agg.ID,
			EventType:   ev.GetEventType(),
			Payload:     payload,
			Version:     baseVersion + i + 1,
			OccurredAt:  ev.GetOccurredAt(),
		})
	}

	// ── Step 2: Persist to EventStore (Postgres) + outbox in same TX ──
	if err := s.eventStore.AppendEventsWithHook(ctx, agg.ID, baseVersion, descriptors,
		func(tx *gorm.DB) error {
			return s.insertOutboxEntries(tx, descriptors)
		},
	); err != nil {
		return fmt.Errorf("eventStore.AppendEventsWithHook: %w", err)
	}

	// Clear the uncommitted buffer only after successful DB commit.
	agg.ClearUncommittedEvents()

	s.logger.Info("Account events persisted to EventStore",
		zap.String("aggregate_id", agg.ID),
		zap.Int("count", len(descriptors)),
		zap.Int("base_version", baseVersion),
	)

	return nil
}

func (s *EventSourcingService) insertOutboxEntries(tx *gorm.DB, descs []EventDescriptor) error {
	for _, d := range descs {
		envelope, err := json.Marshal(d)
		if err != nil {
			return fmt.Errorf("marshal account EventDescriptor %s: %w", d.EventType, err)
		}

		if err := outbox.InsertOutboxEvent(tx, d.ID, KafkaTopicAccountEvents, d.AggregateID, envelope); err != nil {
			return fmt.Errorf("insert outbox account event %s v%d: %w", d.EventType, d.Version, err)
		}
	}

	return nil
}
