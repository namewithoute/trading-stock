package account

import (
	"context"
	"fmt"

	domain "trading-stock/internal/domain/account"

	"go.uber.org/zap"
)

// ─────────────────────────────────────────────────────────────────────────────
// QueryHandler — read side of the CQRS split.
//
// All methods SELECT from the account_read_models table only.
// No aggregate loading, no event replay, no writes of any kind.
// ─────────────────────────────────────────────────────────────────────────────

// QueryHandler defines all read operations for the account domain.
type QueryHandler interface {
	GetAccount(ctx context.Context, q GetAccountQuery) (*domain.AccountReadModel, error)
	GetAccountByUser(ctx context.Context, q GetAccountByUserQuery) (*domain.AccountReadModel, error)
	ListAccounts(ctx context.Context, q ListAccountsQuery) ([]*domain.AccountReadModel, error)
	GetPrimaryAccountByUser(ctx context.Context, q GetPrimaryAccountByUserQuery) (*domain.AccountReadModel, error)
}

type queryHandler struct {
	readRepo domain.ReadModelRepository
	logger   *zap.Logger
}

func newQueryHandler(readRepo domain.ReadModelRepository, logger *zap.Logger) QueryHandler {
	return &queryHandler{readRepo: readRepo, logger: logger}
}

func (h *queryHandler) GetAccount(ctx context.Context, q GetAccountQuery) (*domain.AccountReadModel, error) {
	return h.readRepo.GetByID(ctx, q.AccountID)
}

func (h *queryHandler) GetAccountByUser(ctx context.Context, q GetAccountByUserQuery) (*domain.AccountReadModel, error) {
	rm, err := h.readRepo.GetByID(ctx, q.AccountID)
	if err != nil {
		return nil, err
	}
	if rm.UserID != q.UserID {
		return nil, domain.ErrAccountNotFound
	}
	return rm, nil
}

func (h *queryHandler) ListAccounts(ctx context.Context, q ListAccountsQuery) ([]*domain.AccountReadModel, error) {
	return h.readRepo.GetByUserID(ctx, q.UserID)
}

func (h *queryHandler) GetPrimaryAccountByUser(ctx context.Context, q GetPrimaryAccountByUserQuery) (*domain.AccountReadModel, error) {
	accounts, err := h.readRepo.GetByUserID(ctx, q.UserID)
	if err != nil {
		return nil, err
	}
	for _, acc := range accounts {
		if acc.Status == domain.StatusActive {
			return acc, nil
		}
	}
	return nil, fmt.Errorf("no active account found for user %s: %w", q.UserID, domain.ErrAccountNotFound)
}
