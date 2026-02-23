package account

import (
	"context"

	domain "trading-stock/internal/domain/account"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ─────────────────────────────────────────────────────────────────────────────
// CommandHandler — write side of the CQRS split.
//
// Pattern per command:
//  1. Load aggregate via domain.Repository (implementation-agnostic)
//  2. Run the domain behavior (validates invariants, emits domain events)
//  3. Save aggregate (infrastructure persists uncommitted events)
//  4. Upsert read model synchronously → "read your own writes" guarantee
// ─────────────────────────────────────────────────────────────────────────────

// CommandHandler defines all write operations for the account domain.
type CommandHandler interface {
	CreateAccount(ctx context.Context, cmd CreateAccountCommand) (*domain.AccountReadModel, error)
	Deposit(ctx context.Context, cmd DepositCommand) (*domain.AccountReadModel, error)
	Withdraw(ctx context.Context, cmd WithdrawCommand) (*domain.AccountReadModel, error)
	ReserveFunds(ctx context.Context, cmd ReserveFundsCommand) (*domain.AccountReadModel, error)
	ReleaseFunds(ctx context.Context, cmd ReleaseFundsCommand) (*domain.AccountReadModel, error)
	FreezeAccount(ctx context.Context, cmd FreezeAccountCommand) (*domain.AccountReadModel, error)
	UnfreezeAccount(ctx context.Context, cmd UnfreezeAccountCommand) (*domain.AccountReadModel, error)
	CloseAccount(ctx context.Context, cmd CloseAccountCommand) (*domain.AccountReadModel, error)
}

type commandHandler struct {
	repo     domain.Repository          // load & save aggregates
	readRepo domain.ReadModelRepository // synchronous read-model upsert
	logger   *zap.Logger
}

func newCommandHandler(repo domain.Repository, readRepo domain.ReadModelRepository, logger *zap.Logger) CommandHandler {
	return &commandHandler{repo: repo, readRepo: readRepo, logger: logger}
}

// saveAndProject saves the aggregate then synchronously upserts the read model.
func (h *commandHandler) saveAndProject(ctx context.Context, agg *domain.AccountAggregate) (*domain.AccountReadModel, error) {
	if err := h.repo.Save(ctx, agg); err != nil {
		return nil, err
	}
	rm := agg.ToReadModel()
	if err := h.readRepo.Upsert(ctx, rm); err != nil {
		// Non-fatal: the event store has the truth. Log and continue.
		h.logger.Error("Failed to upsert read model after command",
			zap.String("accountID", agg.ID),
			zap.Error(err),
		)
	}
	return rm, nil
}

// ─── Command implementations ──────────────────────────────────────────────────

func (h *commandHandler) CreateAccount(ctx context.Context, cmd CreateAccountCommand) (*domain.AccountReadModel, error) {
	accountType := domain.AccountType(cmd.AccountType)
	if !accountType.IsValid() {
		accountType = domain.TypeCash
	}
	currency := cmd.Currency
	if currency == "" {
		currency = "USD"
	}
	agg, err := domain.OpenAccount(uuid.New().String(), cmd.UserID, accountType, currency)
	if err != nil {
		return nil, err
	}
	return h.saveAndProject(ctx, agg)
}

func (h *commandHandler) Deposit(ctx context.Context, cmd DepositCommand) (*domain.AccountReadModel, error) {
	agg, err := h.repo.Load(ctx, cmd.AccountID)
	if err != nil {
		return nil, err
	}
	if err := agg.Deposit(cmd.Amount); err != nil {
		return nil, err
	}
	return h.saveAndProject(ctx, agg)
}

func (h *commandHandler) Withdraw(ctx context.Context, cmd WithdrawCommand) (*domain.AccountReadModel, error) {
	agg, err := h.repo.Load(ctx, cmd.AccountID)
	if err != nil {
		return nil, err
	}
	if err := agg.Withdraw(cmd.Amount); err != nil {
		return nil, err
	}
	return h.saveAndProject(ctx, agg)
}

func (h *commandHandler) ReserveFunds(ctx context.Context, cmd ReserveFundsCommand) (*domain.AccountReadModel, error) {
	agg, err := h.repo.Load(ctx, cmd.AccountID)
	if err != nil {
		return nil, err
	}
	if err := agg.ReserveFunds(cmd.Amount); err != nil {
		return nil, err
	}
	return h.saveAndProject(ctx, agg)
}

func (h *commandHandler) ReleaseFunds(ctx context.Context, cmd ReleaseFundsCommand) (*domain.AccountReadModel, error) {
	agg, err := h.repo.Load(ctx, cmd.AccountID)
	if err != nil {
		return nil, err
	}
	if err := agg.ReleaseFunds(cmd.Amount); err != nil {
		return nil, err
	}
	return h.saveAndProject(ctx, agg)
}

func (h *commandHandler) FreezeAccount(ctx context.Context, cmd FreezeAccountCommand) (*domain.AccountReadModel, error) {
	agg, err := h.repo.Load(ctx, cmd.AccountID)
	if err != nil {
		return nil, err
	}
	if err := agg.Freeze(); err != nil {
		return nil, err
	}
	return h.saveAndProject(ctx, agg)
}

func (h *commandHandler) UnfreezeAccount(ctx context.Context, cmd UnfreezeAccountCommand) (*domain.AccountReadModel, error) {
	agg, err := h.repo.Load(ctx, cmd.AccountID)
	if err != nil {
		return nil, err
	}
	if err := agg.Unfreeze(); err != nil {
		return nil, err
	}
	return h.saveAndProject(ctx, agg)
}

func (h *commandHandler) CloseAccount(ctx context.Context, cmd CloseAccountCommand) (*domain.AccountReadModel, error) {
	agg, err := h.repo.Load(ctx, cmd.AccountID)
	if err != nil {
		return nil, err
	}
	if err := agg.Close(); err != nil {
		return nil, err
	}
	return h.saveAndProject(ctx, agg)
}
