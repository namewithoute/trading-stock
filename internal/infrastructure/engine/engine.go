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

// SubmitRequest is the message sent to a per-symbol goroutine.
type SubmitRequest struct {
	Order  *order.Order
	Result chan<- SubmitResult
}

// CancelRequest asks a symbol goroutine to remove an order from its book.
type CancelRequest struct {
	OrderID string
	Side    order.Side
	Result  chan<- error
}

// symbolMessage is the union type routed to each symbol goroutine.
type symbolMessage struct {
	submit *SubmitRequest
	cancel *CancelRequest
}

// SubmitResult carries the outcome of a SubmitRequest back to the caller.
type SubmitResult struct {
	Trades []*Trade
	Err    error
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
	SymbolChannelSize int // per-symbol channel buffer (default 500)
}

// MatchingEngine runs one goroutine per symbol.
// Orders are routed to the correct goroutine via per-symbol channels,
// so each OrderBook is accessed by exactly one goroutine (no locks needed).
type MatchingEngine struct {
	channels map[string]chan symbolMessage // symbol → channel
	mu       sync.RWMutex                  // protects channels map only
	logger   *zap.Logger

	symbolChanSize int
	cancel         context.CancelFunc
	wg             sync.WaitGroup

	// Event channels for publishing trades and order updates
	tradeChannel       chan *Trade
	orderUpdateChannel chan *OrderUpdate
}

// NewMatchingEngine creates a new matching engine.
// Call RegisterSymbols() to start per-symbol goroutines, then use Submit/Cancel.
func NewMatchingEngine(config MatchingEngineConfig) *MatchingEngine {
	if config.TradeChannelSize == 0 {
		config.TradeChannelSize = 1000
	}
	if config.UpdateChannelSize == 0 {
		config.UpdateChannelSize = 1000
	}
	if config.SymbolChannelSize == 0 {
		config.SymbolChannelSize = 500
	}

	return &MatchingEngine{
		channels:           make(map[string]chan symbolMessage),
		logger:             config.Logger,
		symbolChanSize:     config.SymbolChannelSize,
		tradeChannel:       make(chan *Trade, config.TradeChannelSize),
		orderUpdateChannel: make(chan *OrderUpdate, config.UpdateChannelSize),
	}
}

// RegisterSymbols starts one goroutine per symbol. Must be called once at startup.
func (me *MatchingEngine) RegisterSymbols(ctx context.Context, symbols []string) {
	ctx, me.cancel = context.WithCancel(ctx)

	me.mu.Lock()
	defer me.mu.Unlock()

	for _, sym := range symbols {
		if _, exists := me.channels[sym]; exists {
			continue
		}
		ch := make(chan symbolMessage, me.symbolChanSize)
		me.channels[sym] = ch

		ob := NewOrderBook(sym)
		me.wg.Add(1)
		go me.symbolLoop(ctx, sym, ob, ch)
	}
	me.logger.Info("[ MatchingEngine ] registered symbol goroutines",
		zap.Int("count", len(symbols)),
	)
}

// EnsureSymbol dynamically registers a symbol if not yet known (hot path guard).
func (me *MatchingEngine) EnsureSymbol(ctx context.Context, symbol string) {
	me.mu.RLock()
	_, exists := me.channels[symbol]
	me.mu.RUnlock()
	if exists {
		return
	}

	me.mu.Lock()
	defer me.mu.Unlock()
	// double-check
	if _, exists := me.channels[symbol]; exists {
		return
	}
	ch := make(chan symbolMessage, me.symbolChanSize)
	me.channels[symbol] = ch
	ob := NewOrderBook(symbol)
	me.wg.Add(1)
	go me.symbolLoop(ctx, symbol, ob, ch)
	me.logger.Info("[ MatchingEngine ] dynamically registered symbol", zap.String("symbol", symbol))
}

// Stop cancels all symbol goroutines and waits for them to drain.
func (me *MatchingEngine) Stop() {
	if me.cancel != nil {
		me.cancel()
	}
	me.wg.Wait()
}

// SubmitOrder sends an order to the correct symbol goroutine and waits for the result.
func (me *MatchingEngine) SubmitOrder(ctx context.Context, o *order.Order) ([]*Trade, error) {
	if o == nil {
		return nil, fmt.Errorf("order cannot be nil")
	}

	ch, err := me.channelFor(o.Symbol)
	if err != nil {
		return nil, err
	}

	resCh := make(chan SubmitResult, 1)
	select {
	case ch <- symbolMessage{submit: &SubmitRequest{Order: o, Result: resCh}}:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	select {
	case res := <-resCh:
		return res.Trades, res.Err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// CancelOrder sends a cancel request to the correct symbol goroutine.
func (me *MatchingEngine) CancelOrder(ctx context.Context, orderID string, symbol string, side order.Side) error {
	ch, err := me.channelFor(symbol)
	if err != nil {
		return err
	}

	resCh := make(chan error, 1)
	select {
	case ch <- symbolMessage{cancel: &CancelRequest{OrderID: orderID, Side: side, Result: resCh}}:
	case <-ctx.Done():
		return ctx.Err()
	}

	select {
	case err := <-resCh:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// channelFor returns the channel for a given symbol.
func (me *MatchingEngine) channelFor(symbol string) (chan symbolMessage, error) {
	me.mu.RLock()
	ch, ok := me.channels[symbol]
	me.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("no order book registered for symbol: %s", symbol)
	}
	return ch, nil
}

// ── per-symbol goroutine ──────────────────────────────────────────────────────

func (me *MatchingEngine) symbolLoop(ctx context.Context, symbol string, ob *OrderBook, ch <-chan symbolMessage) {
	defer me.wg.Done()
	me.logger.Info("[ MatchingEngine ] symbol goroutine started", zap.String("symbol", symbol))

	for {
		select {
		case <-ctx.Done():
			me.logger.Info("[ MatchingEngine ] symbol goroutine stopped", zap.String("symbol", symbol))
			return
		case msg := <-ch:
			if msg.submit != nil {
				trades, err := me.processOrder(ob, msg.submit.Order)
				msg.submit.Result <- SubmitResult{Trades: trades, Err: err}
			}
			if msg.cancel != nil {
				err := ob.RemoveOrder(msg.cancel.OrderID, msg.cancel.Side)
				msg.cancel.Result <- err
			}
		}
	}
}

// processOrder matches an order, updates statuses, publishes events — all single-threaded.
func (me *MatchingEngine) processOrder(ob *OrderBook, o *order.Order) ([]*Trade, error) {
	trades, err := me.matchOrder(ob, o)
	if err != nil {
		me.logger.Error("Failed to match order",
			zap.String("order_id", o.ID),
			zap.Error(err),
		)
		return nil, err
	}

	// Publish all trades
	for _, trade := range trades {
		me.publishTrade(trade)
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

		if len(trades) > 0 {
			o.Status = order.StatusPartiallyFilled
		} else {
			o.Status = order.StatusPending
		}
		me.publishOrderUpdate(o)
	} else if o.RemainingQuantity() == 0 {
		o.Status = order.StatusFilled
		me.publishOrderUpdate(o)
	} else if o.Type == order.TypeMarket && o.RemainingQuantity() > 0 {
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
		trades = me.matchBuyOrder(ob, incomingOrder)
	} else if incomingOrder.Side == order.SideSell {
		trades = me.matchSellOrder(ob, incomingOrder)
	} else {
		return nil, fmt.Errorf("invalid order side: %s", incomingOrder.Side)
	}

	return trades, nil
}

// matchBuyOrder matches a buy order against sell orders
func (me *MatchingEngine) matchBuyOrder(ob *OrderBook, buyOrder *order.Order) []*Trade {
	trades := make([]*Trade, 0)

	for buyOrder.RemainingQuantity() > 0 && ob.Asks.Len() > 0 {
		bestAsk := (*ob.Asks)[0]

		if buyOrder.Type == order.TypeLimit && buyOrder.Price.Cmp(&bestAsk.Price) < 0 {
			break
		}

		tradeQty := min(buyOrder.RemainingQuantity(), bestAsk.RemainingQuantity())
		tradePrice := bestAsk.Price

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

		buyOrder.FilledQuantity += tradeQty
		bestAsk.FilledQuantity += tradeQty

		me.updateAvgFillPrice(buyOrder, tradePrice, tradeQty)
		me.updateAvgFillPrice(bestAsk, tradePrice, tradeQty)

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

		if sellOrder.Type == order.TypeLimit && sellOrder.Price.Cmp(&bestBid.Price) > 0 {
			break
		}

		tradeQty := min(sellOrder.RemainingQuantity(), bestBid.RemainingQuantity())
		tradePrice := bestBid.Price

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

		sellOrder.FilledQuantity += tradeQty
		bestBid.FilledQuantity += tradeQty

		me.updateAvgFillPrice(sellOrder, tradePrice, tradeQty)
		me.updateAvgFillPrice(bestBid, tradePrice, tradeQty)

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
		o.AvgFillPrice = tradePrice
	} else {
		prevTotal := new(apd.Decimal)
		_, _ = decCtx.Mul(prevTotal, &o.AvgFillPrice, apd.New(int64(o.FilledQuantity-tradeQty), 0))
		newTotal := new(apd.Decimal)
		_, _ = decCtx.Mul(newTotal, &tradePrice, apd.New(int64(tradeQty), 0))
		sum := new(apd.Decimal)
		_, _ = decCtx.Add(sum, prevTotal, newTotal)
		_, _ = decCtx.Quo(&o.AvgFillPrice, sum, apd.New(int64(o.FilledQuantity), 0))
	}
}

// publishTrade publishes a trade to the trade channel
func (me *MatchingEngine) publishTrade(trade *Trade) {
	select {
	case me.tradeChannel <- trade:
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
	default:
		me.logger.Warn("Order update channel full, dropping update",
			zap.String("order_id", o.ID),
		)
	}
}

// TradeChannel returns the read-only trade event channel.
func (me *MatchingEngine) TradeChannel() <-chan *Trade {
	return me.tradeChannel
}

// OrderUpdateChannel returns the read-only order update event channel.
func (me *MatchingEngine) OrderUpdateChannel() <-chan *OrderUpdate {
	return me.orderUpdateChannel
}

// GetOrCreateOrderBook is kept for backward compatibility (tests).
// In production the symbol goroutine owns the book; this creates a detached copy.
func (me *MatchingEngine) GetOrCreateOrderBook(symbol string) *OrderBook {
	return NewOrderBook(symbol)
}
