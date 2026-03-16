package account

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	domain "trading-stock/internal/domain/account"

	"github.com/cockroachdb/apd/v3"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

var projDecCtx = apd.BaseContext.WithPrecision(19)

// Projector is a Kafka consumer that listens to account.events and
// upserts the account_read_models table after each event.
//
// Startup lifecycle:
//
//	wire.go      → NewProjector(...)       (build)
//	lifecycle.go → p.Rebuild(ctx)          (catch-up from EventStore BEFORE Kafka loop)
//	lifecycle.go → go p.Run(ctx)          (then live-stream from Kafka)
//	lifecycle.go → projectorCancel()      (stop on SIGTERM)
type Projector struct {
	reader     *kafka.Reader
	readRepo   domain.ReadModelRepository
	eventStore EventStore // used ONLY during Rebuild(), not in hot Kafka path
	logger     *zap.Logger
}

// NewProjector creates a Kafka reader and wires the projector.
func NewProjector(
	brokers []string,
	readRepo domain.ReadModelRepository,
	eventStore EventStore,
	logger *zap.Logger,
) *Projector {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   KafkaTopicAccountEvents,
		GroupID: "account-projector", // consumer group – guarantees at-least-once

		// Wait for at least 1 byte, receive up to 10 MB per fetch
		MinBytes: 1,
		MaxBytes: 10e6,

		// How long to wait for new messages before returning (avoids tight loop)
		MaxWait: 1 * time.Second,

		// After Rebuild() completes we switch to LastOffset so the Kafka consumer
		// only processes BRAND-NEW events — no double-counting with what Rebuild already applied.
		StartOffset: kafka.LastOffset,

		// Commit offsets explicitly (after p.reader.CommitMessages)
		CommitInterval: 0,
	})

	return &Projector{
		reader:     reader,
		readRepo:   readRepo,
		eventStore: eventStore,
		logger:     logger,
	}
}

// Rebuild performs a full catch-up projection by replaying every event from the
// EventStore (Postgres) and applying them to the read model table.
//
// Call this ONCE synchronously in lifecycle.go BEFORE starting the Kafka consumer
// goroutine. This guarantees the read model is up-to-date with the EventStore
// truth before any live Kafka events arrive.
//
// Strategy:
//   - Load all EventDescriptors ordered by (aggregate_id, version)
//   - Apply each one through the same applyDescriptor() logic used by the Kafka path
//   - Version guard is SKIPPED here because we are replaying in strict order from source-of-truth
func (p *Projector) Rebuild(ctx context.Context) error {
	p.logger.Info("[ Projector ] Starting catch-up rebuild from EventStore...")

	all, err := p.eventStore.LoadAllDescriptors(ctx)
	if err != nil {
		return fmt.Errorf("Rebuild: load all descriptors: %w", err)
	}
	if len(all) == 0 {
		p.logger.Info("[ Projector ] EventStore is empty — nothing to rebuild")
		return nil
	}

	// Process events grouped by aggregate — they arrive pre-sorted by (aggregate_id, version).
	// We keep a running in-memory read model per aggregate to avoid N+1 DB reads.
	readModels := make(map[string]*domain.AccountReadModel) // aggregateID → current rm

	for _, d := range all {
		rm, exists := readModels[d.AggregateID]
		if !exists {
			rm = &domain.AccountReadModel{}
			readModels[d.AggregateID] = rm
		}
		// Apply event payload directly onto the in-memory read model
		if err := applyDescriptor(d, rm); err != nil {
			// Log and skip unknown/corrupted events rather than aborting the full rebuild
			p.logger.Warn("[ Projector ] Skipping unrecognised event during Rebuild",
				zap.String("event_type", string(d.EventType)),
				zap.String("aggregate_id", d.AggregateID),
				zap.Error(err),
			)
		}
		rm.Version = d.Version
		rm.UpdatedAt = d.OccurredAt
	}

	// Persist all rebuilt read models to DB in one pass
	count := 0
	for _, rm := range readModels {
		if err := p.readRepo.Upsert(ctx, rm); err != nil {
			p.logger.Error("[ Projector ] Failed to upsert read model during Rebuild",
				zap.String("aggregate_id", rm.ID),
				zap.Error(err),
			)
			continue
		}
		count++
	}

	p.logger.Info("[ Projector ] Catch-up rebuild complete",
		zap.Int("total_events", len(all)),
		zap.Int("read_models_rebuilt", count),
	)
	return nil
}

// Run is the blocking event loop. Call it in a dedicated goroutine.
// It exits cleanly when ctx is cancelled (e.g. on SIGTERM).
func (p *Projector) Run(ctx context.Context) {
	p.logger.Info("[ Projector ] Account Projector started",
		zap.String("topic", KafkaTopicAccountEvents),
		zap.String("group", "account-projector"),
	)
	defer p.reader.Close()

	for {
		// FetchMessage blocks until a message arrives or MaxWait elapses.
		msg, err := p.reader.FetchMessage(ctx)
		if err != nil {
			// Clean exit when context is cancelled (SIGTERM / shutdown)
			if ctx.Err() != nil {
				p.logger.Info("[ Projector ] Account Projector shutting down gracefully")
				return
			}

			// Transient error (group coordinator not ready, broker restart, etc.)
			// Log as warning and back-off briefly to avoid tight loop spam.
			p.logger.Warn("[ Projector ] FetchMessage error – retrying in 2s",
				zap.Error(err),
			)
			select {
			case <-ctx.Done():
				return
			case <-time.After(2 * time.Second):
				continue
			}
		}

		// Project the event into the read model
		if err := p.project(ctx, msg); err != nil {
			p.logger.Error("[ Projector ] Failed to project event",
				zap.Error(err),
				zap.String("aggregate_id", string(msg.Key)),
			)
			// NOTE: Commit anyway to avoid being stuck on a poison-pill message.
			// In production: route to a dead-letter topic instead of skipping.
		}

		// Commit the Kafka offset after processing (at-least-once semantics)
		if err := p.reader.CommitMessages(ctx, msg); err != nil {
			p.logger.Error("[ Projector ] Failed to commit Kafka offset", zap.Error(err))
		}
	}
}

// project handles a single Kafka message:
//  1. Decode EventDescriptor envelope from the message payload
//  2. Version guard — skip duplicates / out-of-order events
//  3. Apply event onto the current read model via applyDescriptor()
//  4. Upsert the updated read model into account_read_models
func (p *Projector) project(ctx context.Context, msg kafka.Message) error {
	// 1. Decode outer envelope
	var descriptor EventDescriptor
	if err := json.Unmarshal(msg.Value, &descriptor); err != nil {
		return fmt.Errorf("unmarshal EventDescriptor: %w", err)
	}

	// 2. Fetch the CURRENT read model (not EventStore)
	rm, err := p.readRepo.GetByID(ctx, descriptor.AggregateID)
	if err != nil && descriptor.EventType != domain.EventAccountCreated {
		return fmt.Errorf("get read model: %w", err)
	}

	// 3. Version guard — idempotent & ordering protection
	if rm != nil {
		if descriptor.Version <= rm.Version {
			// Duplicate or already-applied event (safe to discard)
			p.logger.Info("[ Projector ] Skipping duplicate/out-of-order event",
				zap.Int("event_version", descriptor.Version),
				zap.Int("rm_version", rm.Version),
				zap.String("aggregate_id", descriptor.AggregateID),
			)
			return nil // Idempotent!
		}
		if descriptor.Version > rm.Version+1 {
			// Gap detected — event(s) lost in transit. In production: send to DLQ.
			return fmt.Errorf("missing events: expected v%d, got v%d (aggregate=%s)",
				rm.Version+1, descriptor.Version, descriptor.AggregateID)
		}
	} else {
		rm = &domain.AccountReadModel{}
	}

	// 4. Apply event payload directly onto the read model (shared with Rebuild path)
	if err := applyDescriptor(descriptor, rm); err != nil {
		return err
	}

	// 5. Stamp version + timestamp
	rm.Version = descriptor.Version
	rm.UpdatedAt = time.Now()

	// 6. Persist to Read DB
	if err := p.readRepo.Upsert(ctx, rm); err != nil {
		return fmt.Errorf("upsert read model: %w", err)
	}

	p.logger.Info("[ Projector ] Read model updated from Kafka event",
		zap.String("event_type", string(descriptor.EventType)),
		zap.String("aggregate_id", descriptor.AggregateID),
		zap.Int("version", descriptor.Version),
	)

	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// applyDescriptor — shared projection logic
// ─────────────────────────────────────────────────────────────────────────────

// applyDescriptor applies a single EventDescriptor payload onto an in-memory
// AccountReadModel. It is a pure function (no I/O) reused by both:
//   - Rebuild()  — replaying events from EventStore on startup
//   - project()  — applying live events arriving from Kafka
func applyDescriptor(d EventDescriptor, rm *domain.AccountReadModel) error {
	switch d.EventType {
	case domain.EventAccountCreated:
		var evt domain.AccountCreatedEvent
		if err := json.Unmarshal(d.Payload, &evt); err != nil {
			return fmt.Errorf("unmarshal AccountCreatedEvent: %w", err)
		}
		rm.ID = evt.AggregateID
		rm.UserID = evt.UserID
		rm.AccountType = evt.AccountType
		rm.Currency = evt.Currency
		rm.Status = domain.StatusActive

	case domain.EventMoneyDeposited:
		var evt domain.MoneyDepositedEvent
		if err := json.Unmarshal(d.Payload, &evt); err != nil {
			return fmt.Errorf("unmarshal MoneyDepositedEvent: %w", err)
		}
		_, _ = projDecCtx.Add(&rm.Balance, &rm.Balance, &evt.Amount.Decimal)
		_, _ = projDecCtx.Add(&rm.BuyingPower, &rm.BuyingPower, &evt.Amount.Decimal)

	case domain.EventMoneyWithdrawn:
		var evt domain.MoneyWithdrawnEvent
		if err := json.Unmarshal(d.Payload, &evt); err != nil {
			return fmt.Errorf("unmarshal MoneyWithdrawnEvent: %w", err)
		}
		_, _ = projDecCtx.Sub(&rm.Balance, &rm.Balance, &evt.Amount.Decimal)
		_, _ = projDecCtx.Sub(&rm.BuyingPower, &rm.BuyingPower, &evt.Amount.Decimal)

	case domain.EventFundsReserved:
		var evt domain.FundsReservedEvent
		if err := json.Unmarshal(d.Payload, &evt); err != nil {
			return fmt.Errorf("unmarshal FundsReservedEvent: %w", err)
		}
		_, _ = projDecCtx.Sub(&rm.BuyingPower, &rm.BuyingPower, &evt.Amount.Decimal)

	case domain.EventFundsReleased:
		var evt domain.FundsReleasedEvent
		if err := json.Unmarshal(d.Payload, &evt); err != nil {
			return fmt.Errorf("unmarshal FundsReleasedEvent: %w", err)
		}
		_, _ = projDecCtx.Add(&rm.BuyingPower, &rm.BuyingPower, &evt.Amount.Decimal)

	case domain.EventStatusChanged:
		var evt domain.StatusChangedEvent
		if err := json.Unmarshal(d.Payload, &evt); err != nil {
			return fmt.Errorf("unmarshal StatusChangedEvent: %w", err)
		}
		rm.Status = evt.NewStatus
	case domain.EventTradeSettled:
		var evt domain.TradeSettledEvent
		if err := json.Unmarshal(d.Payload, &evt); err != nil {
			return fmt.Errorf("unmarshal TradeSettledEvent: %w", err)
		}
		if evt.Side == "BUY" {
			_, _ = projDecCtx.Sub(&rm.Balance, &rm.Balance, &evt.Amount.Decimal)
		} else {
			_, _ = projDecCtx.Add(&rm.Balance, &rm.Balance, &evt.Amount.Decimal)
			_, _ = projDecCtx.Add(&rm.BuyingPower, &rm.BuyingPower, &evt.Amount.Decimal)
		}
	default:
		return fmt.Errorf("unknown event type: %s", d.EventType)
	}
	return nil
}
