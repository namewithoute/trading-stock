package account

import (
	domain "trading-stock/internal/domain/account"

	"go.uber.org/zap"
)

// ─────────────────────────────────────────────────────────────────────────────
// UseCase — CQRS Facade for the Account domain.
//
// UseCase composes CommandHandler (write side) and QueryHandler (read side)
// into a single interface so the Presentation layer has one dependency.
//
// ── Write path ────────────────────────────────────────────────────────────────
//
//	CommandHandler: Load aggregate → run domain command → Save events
//	                              → Upsert read model (synchronous, immediate)
//
// ── Read path ─────────────────────────────────────────────────────────────────
//
//	QueryHandler: SELECT directly from account_read_models — no event replay.
//
// ── Consistency guarantee ──────────────────────────────────────────────────────
//
//	"Read your own writes" is guaranteed because the CommandHandler upserts
//	the read model synchronously before returning.  The Kafka/Projector
//	pipeline serves as a secondary recovery mechanism.
// ─────────────────────────────────────────────────────────────────────────────

// UseCase is the combined CQRS entry point consumed by the Presentation layer.
// It embeds command.CommandHandler (writes) and query.QueryHandler (reads)
// without merging their implementations — each side can evolve independently.
type UseCase interface {
	CommandHandler
	QueryHandler
}

// useCase is the thin facade that delegates to the two specialised handlers.
type useCase struct {
	CommandHandler // write-side implementation
	QueryHandler   // read-side implementation
}

// NewUseCase wires the CQRS handlers together.
func NewUseCase(
	repo domain.Repository,
	readRepo domain.ReadModelRepository,
	logger *zap.Logger,
) UseCase {
	return &useCase{
		CommandHandler: newCommandHandler(repo, readRepo, logger),
		QueryHandler:   newQueryHandler(readRepo, logger),
	}
}
