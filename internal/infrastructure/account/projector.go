package account

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	domain "trading-stock/internal/domain/account"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// Projector is a Kafka consumer that listens to account.events and
// upserts the account_read_models table after each event.
//
// Lifecycle:
//
//	wire.go  → NewProjector(...)          (build)
//	lifecycle.go → go p.Run(ctx)          (start in goroutine)
//	lifecycle.go → projectorCancel()      (stop on SIGTERM)
type Projector struct {
	reader   *kafka.Reader
	readRepo domain.ReadModelRepository
	eventSvc *EventSourcingService
	logger   *zap.Logger
}

// NewProjector creates a Kafka reader and wires the projector.
func NewProjector(
	brokers []string,
	readRepo domain.ReadModelRepository,
	eventSvc *EventSourcingService,
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

		// Start from the EARLIEST available offset so historical events
		// are replayed on first boot (or after a crash).
		// Once offsets are committed, this group will only receive new events.
		StartOffset: kafka.FirstOffset,

		// Commit offsets explicitly (after p.reader.CommitMessages)
		CommitInterval: 0,
	})

	return &Projector{
		reader:   reader,
		readRepo: readRepo,
		eventSvc: eventSvc,
		logger:   logger,
	}
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
//  2. Replay ALL events for that aggregate from EventStore (full rehydration)
//  3. Upsert the resulting read model into account_read_models
func (p *Projector) project(ctx context.Context, msg kafka.Message) error {
	// Decode outer envelope
	var descriptor domain.EventDescriptor
	if err := json.Unmarshal(msg.Value, &descriptor); err != nil {
		return fmt.Errorf("unmarshal EventDescriptor: %w", err)
	}

	// Replay full aggregate from EventStore
	// NOTE: For accounts with many events, implement snapshots here.
	agg, err := p.eventSvc.Load(ctx, descriptor.AggregateID)
	if err != nil {
		return fmt.Errorf("load aggregate %s: %w", descriptor.AggregateID, err)
	}

	// Persist updated read model
	readModel := agg.ToReadModel()
	if err := p.readRepo.Upsert(ctx, readModel); err != nil {
		return fmt.Errorf("upsert read model for %s: %w", descriptor.AggregateID, err)
	}

	p.logger.Info("[ Projector ] Read model updated",
		zap.String("event_type", string(descriptor.EventType)),
		zap.String("aggregate_id", descriptor.AggregateID),
		zap.Int("version", descriptor.Version),
	)

	return nil
}
