package engine

import (
	"container/heap"
	"context"
	"fmt"
	"sync"
	"time"

	"trading-stock/internal/domain/order"

	"github.com/cockroachdb/apd/v3"

	"go.uber.org/zap"
)

// MatchingEngine is the core order matching engine
// It matches buy and sell orders using price-time priority algorithm
type MatchingEngine struct {
	orderBooks map[string]*OrderBook // symbol -> OrderBook
	mu         sync.RWMutex
	logger     *zap.Logger

	// Event channels for publishing trades and order updates
	tradeChannel       chan *Trade
	orderUpdateChannel chan *OrderUpdate
}

// OrderUpdate represents an order status update
type OrderUpdate struct {
	OrderID        string
	Status         order.Status
	FilledQuantity int
	AvgFillPrice   apd.Decimal
	Timestamp      time.Time
}

// MatchingEngineConfig holds configuration for the matching engine
type MatchingEngineConfig struct {
	Logger            *zap.Logger
	TradeChannelSize  int
	UpdateChannelSize int
}

// NewMatchingEngine creates a new matching engine
func NewMatchingEngine(config MatchingEngineConfig) *MatchingEngine {
	if config.TradeChannelSize == 0 {
		config.TradeChannelSize = 1000
	}
	if config.UpdateChannelSize == 0 {
		config.UpdateChannelSize = 1000
	}

	return &MatchingEngine{
		orderBooks:         make(map[string]*OrderBook),
		logger:             config.Logger,
		tradeChannel:       make(chan *Trade, config.TradeChannelSize),
		orderUpdateChannel: make(chan *OrderUpdate, config.UpdateChannelSize),
	}
}

// GetOrCreateOrderBook gets or creates an order book for a symbol
func (me *MatchingEngine) GetOrCreateOrderBook(symbol string) *OrderBook {
	me.mu.Lock()
	defer me.mu.Unlock()

	if ob, exists := me.orderBooks[symbol]; exists {
		return ob
	}

	ob := NewOrderBook(symbol)
	me.orderBooks[symbol] = ob
	me.logger.Info("Created new order book", zap.String("symbol", symbol))
	return ob
}

// GetOrderBook gets an order book for a symbol
func (me *MatchingEngine) GetOrderBook(symbol string) (*OrderBook, error) {
	me.mu.RLock()
	defer me.mu.RUnlock()

	ob, exists := me.orderBooks[symbol]
	if !exists {
		return nil, fmt.Errorf("order book not found for symbol: %s", symbol)
	}
	return ob, nil
}

// SubmitOrder submits an order to the matching engine
// Returns list of trades generated and any error
func (me *MatchingEngine) SubmitOrder(ctx context.Context, o *order.Order) ([]*Trade, error) {
	if o == nil {
		return nil, fmt.Errorf("order cannot be nil")
	}

	// Get or create order book for this symbol
	ob := me.GetOrCreateOrderBook(o.Symbol)

	// Match the order
	trades, err := me.matchOrder(ob, o)
	if err != nil {
		me.logger.Error("Failed to match order",
			zap.String("order_id", o.ID),
			zap.Error(err),
		)
		return nil, err
	}

	// If order is not fully filled, add remaining to order book
	if o.RemainingQuantity() > 0 && o.Type == order.TypeLimit {
		if err := ob.AddOrder(o); err != nil {
			me.logger.Error("Failed to add order to book",
				zap.String("order_id", o.ID),
				zap.Error(err),
			)
			return trades, err
		}

		// Update order status
		if len(trades) > 0 {
			o.Status = order.StatusPartiallyFilled
		} else {
			o.Status = order.StatusPending
		}

		me.publishOrderUpdate(o)
	} else if o.RemainingQuantity() == 0 {
		// Order fully filled
		o.Status = order.StatusFilled
		me.publishOrderUpdate(o)
	} else if o.Type == order.TypeMarket && o.RemainingQuantity() > 0 {
		// Market order not fully filled - reject remaining
		o.Status = order.StatusPartiallyFilled
		me.publishOrderUpdate(o)
	}

	me.logger.Info("Order processed",
		zap.String("order_id", o.ID),
		zap.String("symbol", o.Symbol),
		zap.Int("trades", len(trades)),
		zap.String("status", string(o.Status)),
	)

	return trades, nil
}

// matchOrder matches an incoming order against the order book
func (me *MatchingEngine) matchOrder(ob *OrderBook, incomingOrder *order.Order) ([]*Trade, error) {
	trades := make([]*Trade, 0)

	if incomingOrder.Side == order.SideBuy {
		// Match buy order against sell orders (asks)
		trades = me.matchBuyOrder(ob, incomingOrder)
	} else if incomingOrder.Side == order.SideSell {
		// Match sell order against buy orders (bids)
		trades = me.matchSellOrder(ob, incomingOrder)
	} else {
		return nil, fmt.Errorf("invalid order side: %s", incomingOrder.Side)
	}

	// Publish all trades
	for _, trade := range trades {
		me.publishTrade(trade)
	}

	return trades, nil
}

// matchBuyOrder matches a buy order against sell orders
func (me *MatchingEngine) matchBuyOrder(ob *OrderBook, buyOrder *order.Order) []*Trade {
	trades := make([]*Trade, 0)

	for buyOrder.RemainingQuantity() > 0 && ob.Asks.Len() > 0 {
		bestAsk := (*ob.Asks)[0]

		// Check if prices match
		// For limit orders: buy price must be >= sell price
		// For market orders: always match
		if buyOrder.Type == order.TypeLimit && buyOrder.Price.Cmp(&bestAsk.Price) < 0 {
			break // No more matches possible
		}

		// Calculate trade quantity (minimum of remaining quantities)
		tradeQty := min(buyOrder.RemainingQuantity(), bestAsk.RemainingQuantity())
		tradePrice := bestAsk.Price // Price of the resting order (maker)

		// Create trade
		trade := NewTrade(
			buyOrder.ID,
			bestAsk.ID,
			buyOrder.Symbol,
			tradePrice,
			tradeQty,
			buyOrder.UserID,
			bestAsk.UserID,
		)
		trades = append(trades, trade)

		// Update orders
		buyOrder.FilledQuantity += tradeQty
		bestAsk.FilledQuantity += tradeQty

		// Update average fill prices
		me.updateAvgFillPrice(buyOrder, tradePrice, tradeQty)
		me.updateAvgFillPrice(bestAsk, tradePrice, tradeQty)

		// If sell order is fully filled, remove from book
		if bestAsk.RemainingQuantity() == 0 {
			heap.Pop(ob.Asks)
			bestAsk.Status = order.StatusFilled
			me.publishOrderUpdate(bestAsk)
		} else {
			bestAsk.Status = order.StatusPartiallyFilled
			me.publishOrderUpdate(bestAsk)
		}

		me.logger.Debug("Trade executed",
			zap.String("trade_id", trade.ID),
			zap.String("symbol", trade.Symbol),
			zap.String("price", trade.Price.String()),
			zap.Int("quantity", trade.Quantity),
		)
	}

	return trades
}

// matchSellOrder matches a sell order against buy orders
func (me *MatchingEngine) matchSellOrder(ob *OrderBook, sellOrder *order.Order) []*Trade {
	trades := make([]*Trade, 0)

	for sellOrder.RemainingQuantity() > 0 && ob.Bids.Len() > 0 {
		bestBid := (*ob.Bids)[0]

		// Check if prices match
		// For limit orders: sell price must be <= buy price
		// For market orders: always match
		if sellOrder.Type == order.TypeLimit && sellOrder.Price.Cmp(&bestBid.Price) > 0 {
			break // No more matches possible
		}

		// Calculate trade quantity
		tradeQty := min(sellOrder.RemainingQuantity(), bestBid.RemainingQuantity())
		tradePrice := bestBid.Price // Price of the resting order (maker)

		// Create trade
		trade := NewTrade(
			bestBid.ID,
			sellOrder.ID,
			sellOrder.Symbol,
			tradePrice,
			tradeQty,
			bestBid.UserID,
			sellOrder.UserID,
		)
		trades = append(trades, trade)

		// Update orders
		sellOrder.FilledQuantity += tradeQty
		bestBid.FilledQuantity += tradeQty

		// Update average fill prices
		me.updateAvgFillPrice(sellOrder, tradePrice, tradeQty)
		me.updateAvgFillPrice(bestBid, tradePrice, tradeQty)

		// If buy order is fully filled, remove from book
		if bestBid.RemainingQuantity() == 0 {
			heap.Pop(ob.Bids)
			bestBid.Status = order.StatusFilled
			me.publishOrderUpdate(bestBid)
		} else {
			bestBid.Status = order.StatusPartiallyFilled
			me.publishOrderUpdate(bestBid)
		}

		me.logger.Debug("Trade executed",
			zap.String("trade_id", trade.ID),
			zap.String("symbol", trade.Symbol),
			zap.String("price", trade.Price.String()),
			zap.Int("quantity", trade.Quantity),
		)
	}

	return trades
}

// updateAvgFillPrice updates the average fill price for an order
func (me *MatchingEngine) updateAvgFillPrice(o *order.Order, tradePrice apd.Decimal, tradeQty int) {
	if o.FilledQuantity == tradeQty {
		// First fill
		o.AvgFillPrice = tradePrice
	} else {
		// Calculate weighted average: (prevAvg * prevQty + tradePrice * tradeQty) / totalQty
		prevTotal := new(apd.Decimal)
		_, _ = decCtx.Mul(prevTotal, &o.AvgFillPrice, apd.New(int64(o.FilledQuantity-tradeQty), 0))
		newTotal := new(apd.Decimal)
		_, _ = decCtx.Mul(newTotal, &tradePrice, apd.New(int64(tradeQty), 0))
		sum := new(apd.Decimal)
		_, _ = decCtx.Add(sum, prevTotal, newTotal)
		_, _ = decCtx.Quo(&o.AvgFillPrice, sum, apd.New(int64(o.FilledQuantity), 0))
	}
}

// CancelOrder cancels an order
func (me *MatchingEngine) CancelOrder(ctx context.Context, orderID string, symbol string, side order.Side) error {
	ob, err := me.GetOrderBook(symbol)
	if err != nil {
		return err
	}

	if err := ob.RemoveOrder(orderID, side); err != nil {
		return fmt.Errorf("failed to remove order: %w", err)
	}

	me.logger.Info("Order cancelled",
		zap.String("order_id", orderID),
		zap.String("symbol", symbol),
	)

	return nil
}

// publishTrade publishes a trade to the trade channel
func (me *MatchingEngine) publishTrade(trade *Trade) {
	select {
	case me.tradeChannel <- trade:
		// Trade published successfully
	default:
		me.logger.Warn("Trade channel full, dropping trade",
			zap.String("trade_id", trade.ID),
		)
	}
}

// publishOrderUpdate publishes an order update
func (me *MatchingEngine) publishOrderUpdate(o *order.Order) {
	update := &OrderUpdate{
		OrderID:        o.ID,
		Status:         o.Status,
		FilledQuantity: o.FilledQuantity,
		AvgFillPrice:   o.AvgFillPrice,
		Timestamp:      time.Now(),
	}

	select {
	case me.orderUpdateChannel <- update:
		// Update published successfully
	default:
		me.logger.Warn("Order update channel full, dropping update",
			zap.String("order_id", o.ID),
		)
	}
}

// GetTradeChannel returns the trade channel for consuming trades
func (me *MatchingEngine) GetTradeChannel() <-chan *Trade {
	return me.tradeChannel
}

// GetOrderUpdateChannel returns the order update channel
func (me *MatchingEngine) GetOrderUpdateChannel() <-chan *OrderUpdate {
	return me.orderUpdateChannel
}

// GetAllOrderBooks returns all order books
func (me *MatchingEngine) GetAllOrderBooks() map[string]*OrderBook {
	me.mu.RLock()
	defer me.mu.RUnlock()

	// Return a copy to prevent external modification
	books := make(map[string]*OrderBook, len(me.orderBooks))
	for symbol, ob := range me.orderBooks {
		books[symbol] = ob
	}
	return books
}

// Close closes the matching engine and all channels
func (me *MatchingEngine) Close() {
	close(me.tradeChannel)
	close(me.orderUpdateChannel)
	me.logger.Info("Matching engine closed")
}

// Helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
