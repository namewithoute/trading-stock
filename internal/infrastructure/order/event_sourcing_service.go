package order

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	domain "trading-stock/internal/domain/order"
	"trading-stock/internal/infrastructure/outbox"

	"github.com/cockroachdb/apd/v3"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// KafkaTopicOrderEvents is the Kafka topic for order domain events.
const KafkaTopicOrderEvents = "order.events"

// KafkaTopicOrdersAccepted is the outbox topic consumed by the Matching Engine.
const KafkaTopicOrdersAccepted = "orders.accepted"

// OrderAcceptedMessage is the payload written to the outbox for accepted orders.
// The Matching Engine Consumer deserialises this to drive MatchingEngine.SubmitOrder.
type OrderAcceptedMessage struct {
	EventID    string           `json:"event_id"`
	OrderID    string           `json:"order_id"`
	UserID     string           `json:"user_id"`
	AccountID  string           `json:"account_id"`
	Symbol     string           `json:"symbol"`
	Side       domain.Side      `json:"side"`
	OrderType  domain.OrderType `json:"order_type"`
	Price      apd.Decimal      `json:"price"`
	Quantity   int              `json:"quantity"`
	OccurredAt time.Time        `json:"occurred_at"`
}

// EventSourcingService implements domain.Repository.
//
// Responsibilities:
//  1. Load  – deserialise + replay all events from EventStore → Aggregate
//  2. Save  – serialise uncommitted events → AppendEvents (Postgres) → publish (Kafka)
type EventSourcingService struct {
	eventStore EventStore
	publisher  *kafka.Writer
	logger     *zap.Logger
}

// NewEventSourcingService constructs the service.
func NewEventSourcingService(
	eventStore EventStore,
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

// Load reconstructs an OrderAggregate from the full event history.
func (s *EventSourcingService) Load(ctx context.Context, aggregateID string) (*domain.OrderAggregate, error) {
	descriptors, err := s.eventStore.LoadEvents(ctx, aggregateID)
	if err != nil {
		return nil, fmt.Errorf("order eventStore.LoadEvents(%s): %w", aggregateID, err)
	}
	if len(descriptors) == 0 {
		return nil, domain.ErrOrderNotFound
	}

	events := make([]domain.DomainEvent, 0, len(descriptors))
	for _, d := range descriptors {
		ev, err := DeserialiseEvent(d)
		if err != nil {
			return nil, fmt.Errorf("deserialise order event %s v%d: %w", d.EventType, d.Version, err)
		}
		events = append(events, ev)
	}

	return domain.RehydrateOrder(events), nil
}

// ─── Save ────────────────────────────────────────────────────────────────────

// Save persists uncommitted aggregate events then publishes them to Kafka.
//
// Guarantees:
//   - Postgres write first (inside AppendEvents transaction).
//   - Kafka publish attempted after DB commit; failures are logged only.
func (s *EventSourcingService) Save(ctx context.Context, agg *domain.OrderAggregate) error {
	changes := agg.UncommittedEvents()
	if len(changes) == 0 {
		return nil
	}

	// ── Step 1: Serialise → EventDescriptors ──────────────────────────
	baseVersion := agg.Version - len(changes)
	descriptors := make([]EventDescriptor, 0, len(changes))

	for i, ev := range changes {
		payload, err := json.Marshal(ev)
		if err != nil {
			return fmt.Errorf("marshal order event %s: %w", ev.GetEventType(), err)
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

	// ── Step 2: Persist to EventStore (Postgres) + outbox in same TX ─────
	if err := s.eventStore.AppendEventsWithHook(ctx, agg.ID, baseVersion, descriptors,
		func(tx *gorm.DB) error {
			return s.insertOutboxEntries(tx, changes, descriptors)
		},
	); err != nil {
		return fmt.Errorf("order eventStore.AppendEventsWithHook: %w", err)
	}
	agg.ClearUncommittedEvents()

	s.logger.Info("Order events persisted to EventStore",
		zap.String("aggregate_id", agg.ID),
		zap.Int("count", len(descriptors)),
		zap.Int("base_version", baseVersion),
	)

	// ── Step 3: Publish to Kafka ──────────────────────────────────────
	// Failure is non-fatal: EventStore is source of truth, Projector can replay.
	if err := s.publishToKafka(ctx, descriptors); err != nil {
		s.logger.Error("Kafka publish failed for order events (safe to continue)",
			zap.Error(err),
			zap.String("aggregate_id", agg.ID),
		)
	}

	return nil
}

// ─── publishToKafka ──────────────────────────────────────────────────────────

func (s *EventSourcingService) publishToKafka(ctx context.Context, descs []EventDescriptor) error {
	if s.publisher == nil {
		return fmt.Errorf("kafka writer is nil – skipping publish")
	}

	msgs := make([]kafka.Message, 0, len(descs))
	for _, d := range descs {
		envelope, err := json.Marshal(d)
		if err != nil {
			return fmt.Errorf("marshal event envelope %s: %w", d.EventType, err)
		}
		msgs = append(msgs, kafka.Message{
			Topic: KafkaTopicOrderEvents,
			Key:   []byte(d.AggregateID), // partition by order ID
			Value: envelope,
			Time:  d.OccurredAt,
		})
	}

	writeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return s.publisher.WriteMessages(writeCtx, msgs...)
}

// ─── insertOutboxEntries ──────────────────────────────────────────────────────

// insertOutboxEntries writes outbox rows for events that need downstream fanout.
// Currently only order.placed generates an orders.accepted outbox entry.
// Called inside the DB transaction from AppendEventsWithHook.
func (s *EventSourcingService) insertOutboxEntries(tx *gorm.DB, events []domain.DomainEvent, descriptors []EventDescriptor) error {
	for i, ev := range events {
		if ev.GetEventType() != domain.EventOrderPlaced {
			continue
		}
		placed, ok := ev.(domain.OrderPlacedEvent)
		if !ok {
			continue
		}

		msg := OrderAcceptedMessage{
			EventID:    descriptors[i].ID,
			OrderID:    placed.AggregateID,
			UserID:     placed.UserID,
			AccountID:  placed.AccountID,
			Symbol:     placed.Symbol,
			Side:       placed.Side,
			OrderType:  placed.OrderType,
			Price:      placed.Price.Decimal,
			Quantity:   placed.Quantity,
			OccurredAt: placed.OccurredAt,
		}

		payload, err := json.Marshal(msg)
		if err != nil {
			return fmt.Errorf("marshal OrderAcceptedMessage: %w", err)
		}

		if err := outbox.InsertOutboxEvent(tx, uuid.New().String(), KafkaTopicOrdersAccepted, placed.AggregateID, payload); err != nil {
			return fmt.Errorf("insert outbox order.placed: %w", err)
		}
	}
	return nil
}

// ─── DeserialiseEvent ─────────────────────────────────────────────────────────

// DeserialiseEvent unmarshals an EventDescriptor payload into a typed DomainEvent.
func DeserialiseEvent(d EventDescriptor) (domain.DomainEvent, error) {
	switch d.EventType {
	case domain.EventOrderPlaced:
		var e domain.OrderPlacedEvent
		return e, json.Unmarshal(d.Payload, &e)
	case domain.EventOrderCancelled:
		var e domain.OrderCancelledEvent
		return e, json.Unmarshal(d.Payload, &e)
	case domain.EventOrderPartialFill:
		var e domain.OrderPartialFillEvent
		return e, json.Unmarshal(d.Payload, &e)
	case domain.EventOrderFilled:
		var e domain.OrderFilledEvent
		return e, json.Unmarshal(d.Payload, &e)
	case domain.EventOrderRejected:
		var e domain.OrderRejectedEvent
		return e, json.Unmarshal(d.Payload, &e)
	case domain.EventOrderExpired:
		var e domain.OrderExpiredEvent
		return e, json.Unmarshal(d.Payload, &e)
	default:
		return nil, fmt.Errorf("unknown order event type: %s", d.EventType)
	}
}
