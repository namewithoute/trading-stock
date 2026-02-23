package order

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	accountApp "trading-stock/internal/application/account"
	"trading-stock/internal/domain/order"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// UseCase handles order business logic
type UseCase interface {
	CreateOrder(ctx context.Context, userID, accountID, symbol, side, orderType string, price float64, quantity int) (*order.Order, error)
	ListOrders(ctx context.Context, userID, symbol, status string, limit, offset int) ([]*order.Order, error)
	GetOrder(ctx context.Context, id string) (*order.Order, error)
	CancelOrder(ctx context.Context, id string) error
}

type useCase struct {
	orderRepo  order.Repository
	accountSvc accountApp.UseCase // CQRS: replaces the legacy account.Repository
	kafka      *kafka.Writer
	logger     *zap.Logger
}

func NewUseCase(orderRepo order.Repository, accountSvc accountApp.UseCase, kafka *kafka.Writer, logger *zap.Logger) UseCase {
	return &useCase{orderRepo: orderRepo, accountSvc: accountSvc, kafka: kafka, logger: logger}
}

func (s *useCase) CreateOrder(ctx context.Context, userID, accountID, symbol, side, orderType string, price float64, quantity int) (*order.Order, error) {
	o := &order.Order{
		ID:        uuid.New().String(),
		UserID:    userID,
		Symbol:    symbol,
		Price:     price,
		Quantity:  quantity,
		Side:      order.Side(side),
		Type:      order.OrderType(orderType),
		Status:    order.StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// [Business Rule] 1. Get primary account if accountID is not provided
	if accountID == "" {
		acc, err := s.accountSvc.GetPrimaryAccountByUser(ctx, accountApp.GetPrimaryAccountByUserQuery{UserID: userID})
		if err != nil {
			return nil, fmt.Errorf("failed to get user account: %w", err)
		}
		o.AccountID = acc.ID
	} else {
		o.AccountID = accountID
	}

	// [Business Rule] 2. Lock Balance / Reserve Funds for BUY orders
	if o.Side == order.SideBuy {
		totalCost := float64(o.Quantity) * o.Price
		_, err := s.accountSvc.ReserveFunds(ctx, accountApp.ReserveFundsCommand{
			AccountID: o.AccountID,
			Amount:    totalCost,
		})
		if err != nil {
			// e.g. ErrInsufficientBuyingPower
			return nil, fmt.Errorf("failed to reserve funds: %w", err)
		}
		s.logger.Info("Funds reserved for order", zap.String("orderID", o.ID), zap.Float64("amount", totalCost))
	} else if o.Side == order.SideSell {
		// For SELL orders, we need to check if user has enough Portfolio stock!
		// Leaving this open for future implementation (needs portfolioRepo injection)
	}

	if err := s.orderRepo.Create(ctx, o); err != nil {
		s.logger.Error("Failed to create order in repo", zap.Error(err))
		// [Rollback Rule] Must release funds if DB insert fails!
		if o.Side == order.SideBuy {
			_, _ = s.accountSvc.ReleaseFunds(ctx, accountApp.ReleaseFundsCommand{
				AccountID: o.AccountID,
				Amount:    float64(o.Quantity) * o.Price,
			})
		}
		return nil, err
	}

	// Send to Kafka for matching engine
	orderJSON, _ := json.Marshal(o)
	err := s.kafka.WriteMessages(ctx, kafka.Message{
		Topic: "orders.matching.new",
		Key:   []byte(o.Symbol),
		Value: orderJSON,
	})
	if err != nil {
		s.logger.Error("Failed to send order to Kafka", zap.Error(err), zap.String("orderID", o.ID))
		// Note: depending on criticality, we might want to fail the request or retry
	}

	s.logger.Info("Order created successfully", zap.String("orderID", o.ID), zap.String("userID", userID))
	return o, nil
}

func (s *useCase) ListOrders(ctx context.Context, userID, symbol, status string, limit, offset int) ([]*order.Order, error) {
	return s.orderRepo.ListOrdersByUserIDAndSymbolAndStatus(ctx, userID, symbol, status, limit, offset)
}

func (s *useCase) GetOrder(ctx context.Context, id string) (*order.Order, error) {
	return s.orderRepo.GetByID(ctx, id)
}

func (s *useCase) CancelOrder(ctx context.Context, id string) error {
	o, err := s.orderRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if !o.CanBeCancelled() {
		return order.ErrInvalidStatus // Wait, where is this error?
	}

	return s.orderRepo.Cancel(ctx, id)
}
