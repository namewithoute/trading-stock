package account

import (
	"context"
	"errors"

	domain "trading-stock/internal/domain/account"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// UseCase handles all account business operations using Event Sourcing + CQRS.
//
// ── Write path ────────────────────────────────────────────────────────────────
//
//	Load aggregate → run command → Save events (EventStore + Kafka publish)
//	                             → Upsert read model (synchronous, immediate)
//
// ── Read path ─────────────────────────────────────────────────────────────────
//
//	Query account_read_models directly (no event replay, always fast)
//
// ── Consistency guarantee ──────────────────────────────────────────────────────
//
//	"Read your own writes" is guaranteed because the read model is upserted
//	synchronously within the same HTTP request, before the response is returned.
//	The Kafka/Projector pipeline is a secondary mechanism used for:
//	  - Cross-service projections
//	  - Rebuilding read models after a crash
//	  - Future event-driven integrations
type UseCase interface {
	// Write-side commands
	CreateAccount(ctx context.Context, userID string) (*domain.AccountReadModel, error)
	Deposit(ctx context.Context, id, userID string, amount float64) (*domain.AccountReadModel, error)
	Withdraw(ctx context.Context, id, userID string, amount float64) (*domain.AccountReadModel, error)

	// Read-side queries
	GetAccount(ctx context.Context, id string) (*domain.AccountReadModel, error)
	ListAccounts(ctx context.Context, userID string) ([]*domain.AccountReadModel, error)
}

type useCase struct {
	eventSvc domain.EventSourcingServicePort
	readRepo domain.ReadModelRepository
	logger   *zap.Logger
}

// NewUseCase wires the event sourcing service with the read model repository.
func NewUseCase(
	eventSvc domain.EventSourcingServicePort,
	readRepo domain.ReadModelRepository,
	logger *zap.Logger,
) UseCase {
	return &useCase{
		eventSvc: eventSvc,
		readRepo: readRepo,
		logger:   logger,
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Write-side commands
// ─────────────────────────────────────────────────────────────────────────────

// CreateAccount opens a new trading account for the given user.
func (u *useCase) CreateAccount(ctx context.Context, userID string) (*domain.AccountReadModel, error) {
	id := uuid.New().String()

	// 1. Run command on aggregate root (pure in-memory, no I/O)
	agg := domain.OpenAccount(id, userID, domain.TypeCash, "USD")

	// 2. Persist events (EventStore write + async Kafka publish)
	if err := u.eventSvc.Save(ctx, agg); err != nil {
		u.logger.Error("Failed to save account creation events",
			zap.Error(err), zap.String("userID", userID))
		return nil, err
	}

	// 3. Synchronously upsert the read model so GetAccount works immediately.
	//    ("Read your own writes" consistency guarantee)
	rm := agg.ToReadModel()
	if err := u.readRepo.Upsert(ctx, rm); err != nil {
		// Non-fatal: EventStore already has the events.
		// The Projector will eventually rebuild this from Kafka.
		u.logger.Error("Failed to upsert read model after CreateAccount (will retry via Projector)",
			zap.Error(err), zap.String("accountID", id))
	}

	u.logger.Info("Account created", zap.String("accountID", id), zap.String("userID", userID))
	return rm, nil
}

// Deposit adds funds to an account.
func (u *useCase) Deposit(ctx context.Context, id, userID string, amount float64) (*domain.AccountReadModel, error) {
	// 1. Rehydrate aggregate by replaying all events from EventStore
	agg, err := u.eventSvc.Load(ctx, id)
	if err != nil {
		return nil, err
	}

	// 2. Authorization: account must belong to requesting user
	if agg.UserID != userID {
		return nil, errors.New("unauthorized account access")
	}

	// 3. Domain command – validates business rules, emits MoneyDepositedEvent
	if err := agg.Deposit(amount); err != nil {
		return nil, err
	}

	// 4. Persist new event + async Kafka publish
	if err := u.eventSvc.Save(ctx, agg); err != nil {
		return nil, err
	}

	// 5. Synchronous read model update
	rm := agg.ToReadModel()
	if err := u.readRepo.Upsert(ctx, rm); err != nil {
		u.logger.Error("Failed to upsert read model after Deposit (will retry via Projector)",
			zap.Error(err), zap.String("accountID", id))
	}

	return rm, nil
}

// Withdraw removes funds from an account.
func (u *useCase) Withdraw(ctx context.Context, id, userID string, amount float64) (*domain.AccountReadModel, error) {
	// 1. Rehydrate
	agg, err := u.eventSvc.Load(ctx, id)
	if err != nil {
		return nil, err
	}

	// 2. Authorization
	if agg.UserID != userID {
		return nil, errors.New("unauthorized account access")
	}

	// 3. Domain command – ErrInsufficientBalance enforced inside Aggregate
	if err := agg.Withdraw(amount); err != nil {
		return nil, err
	}

	// 4. Persist
	if err := u.eventSvc.Save(ctx, agg); err != nil {
		return nil, err
	}

	// 5. Synchronous read model update
	rm := agg.ToReadModel()
	if err := u.readRepo.Upsert(ctx, rm); err != nil {
		u.logger.Error("Failed to upsert read model after Withdraw (will retry via Projector)",
			zap.Error(err), zap.String("accountID", id))
	}

	return rm, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Read-side queries – always reads from the pre-built read model table
// ─────────────────────────────────────────────────────────────────────────────

// GetAccount returns a single account read model.
func (u *useCase) GetAccount(ctx context.Context, id string) (*domain.AccountReadModel, error) {
	return u.readRepo.GetByID(ctx, id)
}

// ListAccounts returns all accounts for a user.
func (u *useCase) ListAccounts(ctx context.Context, userID string) ([]*domain.AccountReadModel, error) {
	return u.readRepo.GetByUserID(ctx, userID)
}
