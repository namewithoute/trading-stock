package engine

import (
	"context"
	"testing"
	"time"

	"trading-stock/internal/domain/order"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// TestOrderBookBasicOperations tests basic order book operations
func TestOrderBookBasicOperations(t *testing.T) {
	ob := NewOrderBook("AAPL")

	// Create test orders
	buyOrder := &order.Order{
		ID:        uuid.New().String(),
		UserID:    "user1",
		Symbol:    "AAPL",
		Price:     150.00,
		Quantity:  10,
		Side:      order.SideBuy,
		Type:      order.TypeLimit,
		Status:    order.StatusPending,
		CreatedAt: time.Now(),
	}

	sellOrder := &order.Order{
		ID:        uuid.New().String(),
		UserID:    "user2",
		Symbol:    "AAPL",
		Price:     151.00,
		Quantity:  5,
		Side:      order.SideSell,
		Type:      order.TypeLimit,
		Status:    order.StatusPending,
		CreatedAt: time.Now(),
	}

	// Add orders
	if err := ob.AddOrder(buyOrder); err != nil {
		t.Fatalf("Failed to add buy order: %v", err)
	}

	if err := ob.AddOrder(sellOrder); err != nil {
		t.Fatalf("Failed to add sell order: %v", err)
	}

	// Check best bid and ask
	bestBid := ob.BestBid()
	if bestBid == nil || bestBid.Price != 150.00 {
		t.Errorf("Expected best bid price 150.00, got %v", bestBid)
	}

	bestAsk := ob.BestAsk()
	if bestAsk == nil || bestAsk.Price != 151.00 {
		t.Errorf("Expected best ask price 151.00, got %v", bestAsk)
	}

	// Check spread
	spread := ob.Spread()
	if spread != 1.00 {
		t.Errorf("Expected spread 1.00, got %f", spread)
	}
}

// TestMatchingEngineSimpleMatch tests a simple order match
func TestMatchingEngineSimpleMatch(t *testing.T) {
	logger := zap.NewNop()
	engine := NewMatchingEngine(MatchingEngineConfig{
		Logger: logger,
	})

	// Create a sell order first (resting order)
	sellOrder := &order.Order{
		ID:        uuid.New().String(),
		UserID:    "seller1",
		Symbol:    "AAPL",
		Price:     150.00,
		Quantity:  10,
		Side:      order.SideSell,
		Type:      order.TypeLimit,
		Status:    order.StatusPending,
		CreatedAt: time.Now(),
	}

	// Submit sell order (should go to order book)
	trades, err := engine.SubmitOrder(context.Background(), sellOrder)
	if err != nil {
		t.Fatalf("Failed to submit sell order: %v", err)
	}
	if len(trades) != 0 {
		t.Errorf("Expected 0 trades for resting order, got %d", len(trades))
	}

	// Create a matching buy order
	buyOrder := &order.Order{
		ID:        uuid.New().String(),
		UserID:    "buyer1",
		Symbol:    "AAPL",
		Price:     150.00,
		Quantity:  10,
		Side:      order.SideBuy,
		Type:      order.TypeLimit,
		Status:    order.StatusPending,
		CreatedAt: time.Now(),
	}

	// Submit buy order (should match with sell order)
	trades, err = engine.SubmitOrder(context.Background(), buyOrder)
	if err != nil {
		t.Fatalf("Failed to submit buy order: %v", err)
	}

	// Should generate 1 trade
	if len(trades) != 1 {
		t.Fatalf("Expected 1 trade, got %d", len(trades))
	}

	trade := trades[0]
	if trade.Price != 150.00 {
		t.Errorf("Expected trade price 150.00, got %f", trade.Price)
	}
	if trade.Quantity != 10 {
		t.Errorf("Expected trade quantity 10, got %d", trade.Quantity)
	}

	// Both orders should be fully filled
	if buyOrder.Status != order.StatusFilled {
		t.Errorf("Expected buy order to be filled, got status %s", buyOrder.Status)
	}
	if sellOrder.Status != order.StatusFilled {
		t.Errorf("Expected sell order to be filled, got status %s", sellOrder.Status)
	}
}

// TestMatchingEnginePartialFill tests partial order filling
func TestMatchingEnginePartialFill(t *testing.T) {
	logger := zap.NewNop()
	engine := NewMatchingEngine(MatchingEngineConfig{
		Logger: logger,
	})

	// Create a sell order (resting)
	sellOrder := &order.Order{
		ID:        uuid.New().String(),
		UserID:    "seller1",
		Symbol:    "AAPL",
		Price:     150.00,
		Quantity:  10,
		Side:      order.SideSell,
		Type:      order.TypeLimit,
		Status:    order.StatusPending,
		CreatedAt: time.Now(),
	}

	engine.SubmitOrder(context.Background(), sellOrder)

	// Create a smaller buy order
	buyOrder := &order.Order{
		ID:        uuid.New().String(),
		UserID:    "buyer1",
		Symbol:    "AAPL",
		Price:     150.00,
		Quantity:  5, // Only 5 shares
		Side:      order.SideBuy,
		Type:      order.TypeLimit,
		Status:    order.StatusPending,
		CreatedAt: time.Now(),
	}

	trades, err := engine.SubmitOrder(context.Background(), buyOrder)
	if err != nil {
		t.Fatalf("Failed to submit buy order: %v", err)
	}

	// Should generate 1 trade for 5 shares
	if len(trades) != 1 {
		t.Fatalf("Expected 1 trade, got %d", len(trades))
	}

	if trades[0].Quantity != 5 {
		t.Errorf("Expected trade quantity 5, got %d", trades[0].Quantity)
	}

	// Buy order should be fully filled
	if buyOrder.Status != order.StatusFilled {
		t.Errorf("Expected buy order to be filled, got status %s", buyOrder.Status)
	}

	// Sell order should be partially filled
	if sellOrder.Status != order.StatusPartiallyFilled {
		t.Errorf("Expected sell order to be partially filled, got status %s", sellOrder.Status)
	}

	if sellOrder.RemainingQuantity() != 5 {
		t.Errorf("Expected 5 remaining quantity, got %d", sellOrder.RemainingQuantity())
	}
}

// TestPriceTimePriority tests price-time priority matching
func TestPriceTimePriority(t *testing.T) {
	logger := zap.NewNop()
	engine := NewMatchingEngine(MatchingEngineConfig{
		Logger: logger,
	})

	// Add multiple sell orders at different prices
	sellOrder1 := &order.Order{
		ID:        uuid.New().String(),
		UserID:    "seller1",
		Symbol:    "AAPL",
		Price:     151.00,
		Quantity:  10,
		Side:      order.SideSell,
		Type:      order.TypeLimit,
		Status:    order.StatusPending,
		CreatedAt: time.Now(),
	}

	sellOrder2 := &order.Order{
		ID:        uuid.New().String(),
		UserID:    "seller2",
		Symbol:    "AAPL",
		Price:     150.00, // Better price
		Quantity:  10,
		Side:      order.SideSell,
		Type:      order.TypeLimit,
		Status:    order.StatusPending,
		CreatedAt: time.Now().Add(1 * time.Second),
	}

	engine.SubmitOrder(context.Background(), sellOrder1)
	engine.SubmitOrder(context.Background(), sellOrder2)

	// Create buy order that should match with better price first
	buyOrder := &order.Order{
		ID:        uuid.New().String(),
		UserID:    "buyer1",
		Symbol:    "AAPL",
		Price:     151.00,
		Quantity:  5,
		Side:      order.SideBuy,
		Type:      order.TypeLimit,
		Status:    order.StatusPending,
		CreatedAt: time.Now(),
	}

	trades, err := engine.SubmitOrder(context.Background(), buyOrder)
	if err != nil {
		t.Fatalf("Failed to submit buy order: %v", err)
	}

	// Should match with sellOrder2 (better price)
	if len(trades) != 1 {
		t.Fatalf("Expected 1 trade, got %d", len(trades))
	}

	if trades[0].Price != 150.00 {
		t.Errorf("Expected trade at better price 150.00, got %f", trades[0].Price)
	}

	if trades[0].SellOrderID != sellOrder2.ID {
		t.Errorf("Expected to match with sellOrder2")
	}
}

// TestMarketOrder tests market order execution
func TestMarketOrder(t *testing.T) {
	logger := zap.NewNop()
	engine := NewMatchingEngine(MatchingEngineConfig{
		Logger: logger,
	})

	// Add sell order
	sellOrder := &order.Order{
		ID:        uuid.New().String(),
		UserID:    "seller1",
		Symbol:    "AAPL",
		Price:     150.00,
		Quantity:  10,
		Side:      order.SideSell,
		Type:      order.TypeLimit,
		Status:    order.StatusPending,
		CreatedAt: time.Now(),
	}

	engine.SubmitOrder(context.Background(), sellOrder)

	// Create market buy order (should match at any price)
	marketBuyOrder := &order.Order{
		ID:        uuid.New().String(),
		UserID:    "buyer1",
		Symbol:    "AAPL",
		Price:     0, // Market order
		Quantity:  10,
		Side:      order.SideBuy,
		Type:      order.TypeMarket,
		Status:    order.StatusPending,
		CreatedAt: time.Now(),
	}

	trades, err := engine.SubmitOrder(context.Background(), marketBuyOrder)
	if err != nil {
		t.Fatalf("Failed to submit market order: %v", err)
	}

	// Should match immediately
	if len(trades) != 1 {
		t.Fatalf("Expected 1 trade, got %d", len(trades))
	}

	// Should execute at the sell order price
	if trades[0].Price != 150.00 {
		t.Errorf("Expected trade price 150.00, got %f", trades[0].Price)
	}
}
