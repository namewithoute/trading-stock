package order

import (
	"context"
	"fmt"

	accountApp "trading-stock/internal/application/account"
	"trading-stock/internal/domain/order"

	"github.com/cockroachdb/apd/v3"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

var decCtx = apd.BaseContext.WithPrecision(19)

// UseCase handles order business logic
type UseCase interface {
	CreateOrder(ctx context.Context, userID, accountID, symbol, side, orderType string, price apd.Decimal, quantity int) (*order.OrderReadModel, error)
	ListOrders(ctx context.Context, userID, symbol, status string, limit, offset int) ([]*order.OrderReadModel, error)
	GetOrder(ctx context.Context, id string) (*order.OrderReadModel, error)
	CancelOrder(ctx context.Context, id string) error
	// UpdateOrder cancels the existing PENDING order then re-creates it with updated
	// price / quantity. Caller must verify ownership before invoking.
	UpdateOrder(ctx context.Context, userID, orderID string, newPrice apd.Decimal, newQuantity int) (*order.OrderReadModel, error)
}

type useCase struct {
	orderRepo  order.Repository          // ES write side (Load + Save)
	readRepo   order.ReadModelRepository // query side
	accountSvc accountApp.UseCase        // CQRS: ReserveFunds / ReleaseFunds
	logger     *zap.Logger
}

func NewUseCase(
	orderRepo order.Repository,
	readRepo order.ReadModelRepository,
	accountSvc accountApp.UseCase,
	logger *zap.Logger,
) UseCase {
	return &useCase{
		orderRepo:  orderRepo,
		readRepo:   readRepo,
		accountSvc: accountSvc,
		logger:     logger,
	}
}

// ─── CreateOrder ──────────────────────────────────────────────────────────────

func (s *useCase) CreateOrder(ctx context.Context, userID, accountID, symbol, side, orderType string, price apd.Decimal, quantity int) (*order.OrderReadModel, error) {
	// [Business Rule] 1. Resolve primary account if not provided
	if accountID == "" {
		acc, err := s.accountSvc.GetPrimaryAccountByUser(ctx, accountApp.GetPrimaryAccountByUserQuery{UserID: userID})
		if err != nil {
			return nil, fmt.Errorf("failed to get user account: %w", err)
		}
		accountID = acc.ID
	}

	// [Business Rule] 2. Reserve funds for BUY orders before placing
	if order.Side(side) == order.SideBuy {
		var totalCost apd.Decimal
		_, _ = decCtx.Mul(&totalCost, &price, apd.New(int64(quantity), 0))
		_, err := s.accountSvc.ReserveFunds(ctx, accountApp.ReserveFundsCommand{
			AccountID: accountID,
			Amount:    totalCost,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to reserve funds: %w", err)
		}
		s.logger.Info("Funds reserved for BUY order",
			zap.String("userID", userID),
			zap.String("amount", totalCost.String()),
		)
	}

	// 3. Create the aggregate (validates domain invariants, emits OrderPlacedEvent)
	agg, err := order.PlaceOrder(
		uuid.New().String(),
		userID,
		accountID,
		symbol,
		order.Side(side),
		order.OrderType(orderType),
		quantity,
		price,
	)

	if err != nil {
		// Rollback reserved funds if aggregate creation fails
		if order.Side(side) == order.SideBuy {
			var rollbackAmt apd.Decimal
			_, _ = decCtx.Mul(&rollbackAmt, &price, apd.New(int64(quantity), 0))
			_, _ = s.accountSvc.ReleaseFunds(ctx, accountApp.ReleaseFundsCommand{
				AccountID: accountID,
				Amount:    rollbackAmt,
			})
		}
		return nil, err
	}

	// 4. Persist to EventStore (+ Kafka publish via EventSourcingService)
	if err := s.orderRepo.Save(ctx, agg); err != nil {
		s.logger.Error("Failed to save order events", zap.Error(err))
		// Rollback reserved funds on persistence failure
		if order.Side(side) == order.SideBuy {
			var rollbackAmt apd.Decimal
			_, _ = decCtx.Mul(&rollbackAmt, &price, apd.New(int64(quantity), 0))
			_, _ = s.accountSvc.ReleaseFunds(ctx, accountApp.ReleaseFundsCommand{
				AccountID: accountID,
				Amount:    rollbackAmt,
			})
		}
		return nil, err
	}

	// 5. Return read model built from aggregate state.
	// The persistent read model will be updated asynchronously by the Kafka projector.
	rm := agg.ToReadModel()

	s.logger.Info("Order created",
		zap.String("orderID", agg.ID),
		zap.String("userID", userID),
		zap.String("side", side),
		zap.String("symbol", symbol),
	)
	return rm, nil
}

// ─── ListOrders ───────────────────────────────────────────────────────────────

func (s *useCase) ListOrders(ctx context.Context, userID, symbol, status string, limit, offset int) ([]*order.OrderReadModel, error) {
	return s.readRepo.ListByUserIDAndFilter(ctx, userID, symbol, status, limit, offset)
}

// ─── GetOrder ─────────────────────────────────────────────────────────────────

func (s *useCase) GetOrder(ctx context.Context, id string) (*order.OrderReadModel, error) {
	return s.readRepo.GetByID(ctx, id)
}

// ─── CancelOrder ──────────────────────────────────────────────────────────────

func (s *useCase) CancelOrder(ctx context.Context, id string) error {
	// 1. Load aggregate from EventStore (replay all events)
	agg, err := s.orderRepo.Load(ctx, id)
	if err != nil {
		return fmt.Errorf("order not found: %w", err)
	}

	// 2. Invoke domain behavior (validates CanBeCancelled → emits OrderCancelledEvent)
	if err := agg.Cancel(); err != nil {
		return fmt.Errorf("cannot cancel order: %w", err)
	}

	// 3. Persist new event
	if err := s.orderRepo.Save(ctx, agg); err != nil {
		return fmt.Errorf("failed to persist cancel event: %w", err)
	}

	// 4. Release reserved funds for BUY orders (unfilled quantity only)
	// Read model will be updated asynchronously by the Kafka projector.
	if agg.Side == order.SideBuy {
		remaining := agg.RemainingQuantity()
		if remaining > 0 {
			var releaseAmount apd.Decimal
			_, _ = decCtx.Mul(&releaseAmount, &agg.Price, apd.New(int64(remaining), 0))
			_, releaseErr := s.accountSvc.ReleaseFunds(ctx, accountApp.ReleaseFundsCommand{
				AccountID: agg.AccountID,
				Amount:    releaseAmount,
			})
			if releaseErr != nil {
				// Non-fatal: order is already cancelled in EventStore
				s.logger.Error("Failed to release funds after cancel",
					zap.Error(releaseErr),
					zap.String("orderID", id),
					zap.String("releaseAmount", releaseAmount.String()),
				)
			} else {
				s.logger.Info("Funds released after cancel",
					zap.String("orderID", id),
					zap.String("releaseAmount", releaseAmount.String()),
				)
			}
		}
	}

	return nil
}

// ─── UpdateOrder ──────────────────────────────────────────────────────────────

func (s *useCase) UpdateOrder(ctx context.Context, userID, orderID string, newPrice apd.Decimal, newQuantity int) (*order.OrderReadModel, error) {
	// 1. Load aggregate to verify ownership + cancellability
	agg, err := s.orderRepo.Load(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("order not found: %w", err)
	}

	if agg.UserID != userID {
		return nil, fmt.Errorf("access denied: order does not belong to user")
	}

	if !agg.CanBeModified() {
		return nil, fmt.Errorf("order cannot be modified with status %s: %w", agg.Status, order.ErrInvalidStatus)
	}

	// 2. Cancel old order (releases reserved funds for BUY)
	if err := s.CancelOrder(ctx, orderID); err != nil {
		return nil, fmt.Errorf("failed to cancel old order before update: %w", err)
	}

	// 3. Place new order with updated parameters
	updated, err := s.CreateOrder(ctx, agg.UserID, agg.AccountID, agg.Symbol, string(agg.Side), string(agg.OrderType), newPrice, newQuantity)
	if err != nil {
		return nil, fmt.Errorf("failed to create replacement order: %w", err)
	}

	s.logger.Info("Order updated (cancel+recreate)",
		zap.String("oldOrderID", orderID),
		zap.String("newOrderID", updated.ID),
		zap.String("newPrice", newPrice.String()),
		zap.Int("newQuantity", newQuantity),
	)
	return updated, nil
}
