package order

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	domain "trading-stock/internal/domain/order"
	infraEvents "trading-stock/internal/infrastructure/events"

	"github.com/cockroachdb/apd/v3"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// Projector is a Kafka consumer that listens to order.events and
// upserts the order_read_models table after each event.
//
// Startup lifecycle:
//
//	wire.go      → NewOrderProjector(...)
//	lifecycle.go → p.Rebuild(ctx)    (catch-up from EventStore before Kafka loop)
//	lifecycle.go → go p.Run(ctx)     (live-stream from Kafka)
//	lifecycle.go → cancel()          (stop on SIGTERM)
type Projector struct {
	reader     *kafka.Reader
	readRepo   domain.ReadModelRepository
	eventStore EventStore
	logger     *zap.Logger
}

// NewOrderProjector creates a Kafka reader and wires the projector.
func NewOrderProjector(
	brokers []string,
	readRepo domain.ReadModelRepository,
	eventStore EventStore,
	logger *zap.Logger,
) *Projector {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          KafkaTopicOrderEvents,
		GroupID:        "order-projector",
		MinBytes:       1,
		MaxBytes:       10e6,
		MaxWait:        1 * time.Second,
		StartOffset:    kafka.LastOffset,
		CommitInterval: 0,
	})

	return &Projector{
		reader:     reader,
		readRepo:   readRepo,
		eventStore: eventStore,
		logger:     logger,
	}
}

// ─── Rebuild ──────────────────────────────────────────────────────────────────

// Rebuild performs a full catch-up projection by replaying every event from the
// EventStore (Postgres) before the Kafka consumer loop starts.
func (p *Projector) Rebuild(ctx context.Context) error {
	p.logger.Info("[ OrderProjector ] Starting catch-up rebuild from EventStore...")

	all, err := p.eventStore.LoadAllDescriptors(ctx)
	if err != nil {
		return fmt.Errorf("OrderProjector Rebuild: %w", err)
	}
	if len(all) == 0 {
		p.logger.Info("[ OrderProjector ] EventStore is empty — nothing to rebuild")
		return nil
	}

	// Keep in-memory read models grouped by aggregate to avoid N+1 upserts.
	rms := make(map[string]*domain.OrderReadModel)

	for _, d := range all {
		rm, exists := rms[d.AggregateID]
		if !exists {
			rm = &domain.OrderReadModel{}
			rms[d.AggregateID] = rm
		}
		if err := applyDescriptor(d, rm); err != nil {
			p.logger.Warn("[ OrderProjector ] Skipped unknown event during rebuild",
				zap.String("event_type", string(d.EventType)),
				zap.String("aggregate_id", d.AggregateID),
			)
		}
	}

	// Flush all in-memory read models to DB.
	for _, rm := range rms {
		if err := p.readRepo.Upsert(ctx, rm); err != nil {
			return fmt.Errorf("OrderProjector Rebuild upsert %s: %w", rm.ID, err)
		}
	}

	p.logger.Info("[ OrderProjector ] Rebuild complete",
		zap.Int("aggregates", len(rms)),
		zap.Int("events", len(all)),
	)
	return nil
}

// ─── Run (Kafka live loop) ────────────────────────────────────────────────────

// Run starts the Kafka consumer loop. Blocks until ctx is cancelled.
// Call in a dedicated goroutine after Rebuild() completes.
func (p *Projector) Run(ctx context.Context) {
	p.logger.Info("[ OrderProjector ] Kafka consumer started", zap.String("topic", KafkaTopicOrderEvents))
	defer p.reader.Close()

	for {
		msg, err := p.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				p.logger.Info("[ OrderProjector ] Kafka consumer stopped (context cancelled)")
				return
			}
			p.logger.Error("[ OrderProjector ] FetchMessage error", zap.Error(err))
			continue
		}

		if err := p.handleMessage(ctx, msg); err != nil {
			p.logger.Error("[ OrderProjector ] handleMessage error",
				zap.Error(err),
				zap.String("topic", msg.Topic),
				zap.Int64("offset", msg.Offset),
			)
			// Continue — don't stop the consumer on a single bad message.
		}

		if err := p.reader.CommitMessages(ctx, msg); err != nil {
			p.logger.Error("[ OrderProjector ] CommitMessages error", zap.Error(err))
		}
	}
}

// ─── handleMessage ────────────────────────────────────────────────────────────

func (p *Projector) handleMessage(ctx context.Context, msg kafka.Message) error {
	var d EventDescriptor
	if err := infraEvents.DecodeKafkaPayload(msg.Value, &d); err != nil {
		return fmt.Errorf("unmarshal EventDescriptor: %w", err)
	}

	// Fetch or build the read model for this aggregate.
	rm, err := p.readRepo.GetByID(ctx, d.AggregateID)
	if err != nil {
		// First event for this order — start from blank.
		rm = &domain.OrderReadModel{}
	}

	// Idempotency guard: skip event if already applied.
	if d.Version <= rm.Version {
		return nil
	}

	if err := applyDescriptor(d, rm); err != nil {
		return err
	}

	return p.readRepo.Upsert(ctx, rm)
}

// ─── applyDescriptor ─────────────────────────────────────────────────────────

// applyDescriptor projects a single EventDescriptor onto an OrderReadModel.
// This is the same logic used by both Rebuild and the live Kafka loop.
func applyDescriptor(d EventDescriptor, rm *domain.OrderReadModel) error {
	switch d.EventType {
	case domain.EventOrderPlaced:
		var e domain.OrderPlacedEvent
		if err := json.Unmarshal(d.Payload, &e); err != nil {
			return err
		}
		rm.ID = e.AggregateID
		rm.UserID = e.UserID
		rm.AccountID = e.AccountID
		rm.Symbol = e.Symbol
		rm.Side = e.Side
		rm.OrderType = e.OrderType
		rm.Quantity = e.Quantity
		rm.Price = e.Price.Decimal
		rm.FilledQuantity = 0
		rm.AvgFillPrice = apd.Decimal{}
		rm.Status = domain.StatusPending
		rm.CreatedAt = e.OccurredAt
		rm.UpdatedAt = e.OccurredAt

	case domain.EventOrderCancelled:
		rm.Status = domain.StatusCancelled
		rm.UpdatedAt = d.OccurredAt

	case domain.EventOrderPartialFill:
		var e domain.OrderPartialFillEvent
		if err := json.Unmarshal(d.Payload, &e); err != nil {
			return err
		}
		rm.FilledQuantity = e.TotalFilledQty
		rm.Status = domain.StatusPartiallyFilled
		rm.UpdatedAt = e.OccurredAt

	case domain.EventOrderFilled:
		var e domain.OrderFilledEvent
		if err := json.Unmarshal(d.Payload, &e); err != nil {
			return err
		}
		rm.FilledQuantity = e.TotalFilledQty
		rm.AvgFillPrice = e.AvgFillPrice.Decimal
		rm.Status = domain.StatusFilled
		rm.UpdatedAt = e.OccurredAt

	case domain.EventOrderRejected:
		rm.Status = domain.StatusRejected
		rm.UpdatedAt = d.OccurredAt

	case domain.EventOrderExpired:
		rm.Status = domain.StatusExpired
		rm.UpdatedAt = d.OccurredAt

	default:
		// Unknown events are skipped — forward compatibility.
		return nil
	}

	rm.Version = d.Version
	return nil
}
