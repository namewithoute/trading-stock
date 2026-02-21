package account

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	domain "trading-stock/internal/domain/account"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// KafkaTopicAccountEvents is the Kafka topic for account domain events.
const KafkaTopicAccountEvents = "account.events"

// EventSourcingService implements domain.EventSourcingServicePort.
//
// Responsibilities:
//  1. Load       – deserialise + replay all events from EventStore → Aggregate
//  2. Save       – serialise new events → AppendEvents (Postgres) → publish (Kafka)
//
// Wire rule: only wire.go may construct this struct. All other layers receive
// it through the EventSourcingServicePort interface.
type EventSourcingService struct {
	eventStore domain.EventStore
	publisher  *kafka.Writer
	logger     *zap.Logger
}

// NewEventSourcingService constructs the service.
// publisher must be a *kafka.Writer already configured with brokers.
// Topic is set per-message so one writer can serve multiple topics.
func NewEventSourcingService(
	eventStore domain.EventStore,
	publisher *kafka.Writer,
	logger *zap.Logger,
) *EventSourcingService {
	return &EventSourcingService{
		eventStore: eventStore,
		publisher:  publisher,
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

// Save persists uncommitted aggregate events then publishes them to Kafka.
//
// Guarantees:
//   - Postgres write happens FIRST (atomically inside AppendEvents).
//   - Kafka publish is attempted synchronously AFTER the DB commit succeeds.
//   - If Kafka publish fails the error is logged but NOT returned, so the
//     HTTP caller still gets a 201/200. The Projector can replay from
//     the EventStore on restart to recover read models.
func (s *EventSourcingService) Save(ctx context.Context, agg *domain.AccountAggregate) error {
	changes := agg.UncommittedEvents()
	if len(changes) == 0 {
		return nil
	}

	// ── Step 1: Serialise → EventDescriptors ──────────────────────────
	// Version before these new events = agg.Version - len(changes)
	baseVersion := agg.Version - len(changes)
	descriptors := make([]domain.EventDescriptor, 0, len(changes))

	for i, ev := range changes {
		payload, err := json.Marshal(ev)
		if err != nil {
			return fmt.Errorf("marshal event %s: %w", ev.GetEventType(), err)
		}
		descriptors = append(descriptors, domain.EventDescriptor{
			ID:          uuid.New().String(),
			AggregateID: agg.ID,
			EventType:   ev.GetEventType(),
			Payload:     payload,
			Version:     baseVersion + i + 1,
			OccurredAt:  ev.GetOccurredAt(),
		})
	}

	// ── Step 2: Persist to EventStore (Postgres) ──────────────────────
	// expectedVersion = state BEFORE this command (optimistic concurrency)
	if err := s.eventStore.AppendEvents(ctx, agg.ID, baseVersion, descriptors); err != nil {
		return fmt.Errorf("eventStore.AppendEvents: %w", err)
	}

	// Clear the uncommitted buffer only after successful DB commit.
	agg.ClearUncommittedEvents()

	s.logger.Info("Account events persisted to EventStore",
		zap.String("aggregate_id", agg.ID),
		zap.Int("count", len(descriptors)),
		zap.Int("base_version", baseVersion),
	)

	// ── Step 3: Publish to Kafka ──────────────────────────────────────
	// Done synchronously so the goroutine pool doesn't grow unbounded.
	// A failure here is logged only – the system is still consistent
	// because the EventStore (source of truth) already has the events.
	if err := s.publishToKafka(ctx, descriptors); err != nil {
		s.logger.Error("Kafka publish failed (events already in EventStore – safe to continue)",
			zap.Error(err),
			zap.String("aggregate_id", agg.ID),
			zap.String("topic", KafkaTopicAccountEvents),
		)
		// Do NOT return err: the DB commit succeeded. Projector can replay.
	}

	return nil
}

// ─── publishToKafka ──────────────────────────────────────────────────────────

// publishToKafka serialises each EventDescriptor as a Kafka message.
// Key = AggregateID ensures all events for the same account land on the same partition.
func (s *EventSourcingService) publishToKafka(ctx context.Context, descs []domain.EventDescriptor) error {
	if s.publisher == nil {
		return fmt.Errorf("kafka writer is nil – skipping publish")
	}

	msgs := make([]kafka.Message, 0, len(descs))
	for _, d := range descs {
		envelope, err := json.Marshal(d)
		if err != nil {
			// Skip this event but keep going with others
			s.logger.Error("Failed to marshal EventDescriptor",
				zap.String("event_type", string(d.EventType)),
				zap.Error(err),
			)
			continue
		}
		msgs = append(msgs, kafka.Message{
			Topic: KafkaTopicAccountEvents, // per-message topic (writer has no default topic)
			Key:   []byte(d.AggregateID),
			Value: envelope,
			Time:  time.Now().UTC(),
		})
	}

	if len(msgs) == 0 {
		return nil
	}

	writeCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := s.publisher.WriteMessages(writeCtx, msgs...); err != nil {
		return fmt.Errorf("kafka.WriteMessages: %w", err)
	}

	s.logger.Info("Account events published to Kafka",
		zap.Int("count", len(msgs)),
		zap.String("topic", KafkaTopicAccountEvents),
	)
	return nil
}
