package engine

import (
	"container/heap"
	"fmt"
	"sort"
	"sync"

	"trading-stock/internal/domain/order"

	"github.com/cockroachdb/apd/v3"
)

// OrderBook manages buy and sell orders for a specific symbol
// Uses priority queues for efficient order matching
type OrderBook struct {
	Symbol string
	Bids   *BidQueue // Buy orders (max heap - highest price first)
	Asks   *AskQueue // Sell orders (min heap - lowest price first)
	mu     sync.RWMutex
}

// NewOrderBook creates a new order book for a symbol
func NewOrderBook(symbol string) *OrderBook {
	bids := &BidQueue{}
	asks := &AskQueue{}
	heap.Init(bids)
	heap.Init(asks)

	return &OrderBook{
		Symbol: symbol,
		Bids:   bids,
		Asks:   asks,
	}
}

// AddOrder adds an order to the appropriate side of the order book
func (ob *OrderBook) AddOrder(o *order.Order) error {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	if o.Symbol != ob.Symbol {
		return fmt.Errorf("order symbol %s does not match order book symbol %s", o.Symbol, ob.Symbol)
	}

	if o.Side == order.SideBuy {
		heap.Push(ob.Bids, o)
	} else if o.Side == order.SideSell {
		heap.Push(ob.Asks, o)
	} else {
		return fmt.Errorf("invalid order side: %s", o.Side)
	}

	return nil
}

// RemoveOrder removes an order from the order book
func (ob *OrderBook) RemoveOrder(orderID string, side order.Side) error {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	if side == order.SideBuy {
		return ob.removeFromBids(orderID)
	} else if side == order.SideSell {
		return ob.removeFromAsks(orderID)
	}

	return fmt.Errorf("invalid order side: %s", side)
}

// removeFromBids removes an order from the bid queue
func (ob *OrderBook) removeFromBids(orderID string) error {
	for i, o := range *ob.Bids {
		if o.ID == orderID {
			heap.Remove(ob.Bids, i)
			return nil
		}
	}
	return fmt.Errorf("order %s not found in bids", orderID)
}

// removeFromAsks removes an order from the ask queue
func (ob *OrderBook) removeFromAsks(orderID string) error {
	for i, o := range *ob.Asks {
		if o.ID == orderID {
			heap.Remove(ob.Asks, i)
			return nil
		}
	}
	return fmt.Errorf("order %s not found in asks", orderID)
}

// BestBid returns the highest buy order
func (ob *OrderBook) BestBid() *order.Order {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	if ob.Bids.Len() == 0 {
		return nil
	}
	return (*ob.Bids)[0]
}

// BestAsk returns the lowest sell order
func (ob *OrderBook) BestAsk() *order.Order {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	if ob.Asks.Len() == 0 {
		return nil
	}
	return (*ob.Asks)[0]
}

// Spread returns the difference between best ask and best bid
func (ob *OrderBook) Spread() apd.Decimal {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	bestBid := ob.BestBid()
	bestAsk := ob.BestAsk()

	if bestBid == nil || bestAsk == nil {
		return apd.Decimal{}
	}

	var spread apd.Decimal
	_, _ = decCtx.Sub(&spread, &bestAsk.Price, &bestBid.Price)
	return spread
}

// MidPrice returns the mid price between best bid and ask
func (ob *OrderBook) MidPrice() apd.Decimal {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	bestBid := ob.BestBid()
	bestAsk := ob.BestAsk()

	if bestBid == nil || bestAsk == nil {
		return apd.Decimal{}
	}

	var sum, mid apd.Decimal
	_, _ = decCtx.Add(&sum, &bestBid.Price, &bestAsk.Price)
	_, _ = decCtx.Quo(&mid, &sum, apd.New(2, 0))
	return mid
}

// Depth returns the total quantity at each price level
type DepthLevel struct {
	Price    apd.Decimal `json:"price"`
	Quantity int         `json:"orders"`       // Note: json tag preserved
	Orders   int         `json:"orders_count"` // Number of orders at this level
}

// GetBidDepth returns aggregated bid depth
func (ob *OrderBook) GetBidDepth(levels int) []DepthLevel {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	depthMap := make(map[string]*DepthLevel)

	for _, o := range *ob.Bids {
		key := o.Price.String()
		if level, exists := depthMap[key]; exists {
			level.Quantity += o.RemainingQuantity()
			level.Orders++
		} else {
			depthMap[key] = &DepthLevel{
				Price:    o.Price,
				Quantity: o.RemainingQuantity(),
				Orders:   1,
			}
		}
	}

	depths := make([]DepthLevel, 0, len(depthMap))
	for _, level := range depthMap {
		depths = append(depths, *level)
	}

	// Sort by price descending (highest first)
	sort.Slice(depths, func(i, j int) bool {
		return depths[i].Price.Cmp(&depths[j].Price) > 0
	})

	if levels > 0 && levels < len(depths) {
		return depths[:levels]
	}
	return depths
}

// GetAskDepth returns aggregated ask depth
func (ob *OrderBook) GetAskDepth(levels int) []DepthLevel {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	depthMap := make(map[string]*DepthLevel)

	for _, o := range *ob.Asks {
		key := o.Price.String()
		if level, exists := depthMap[key]; exists {
			level.Quantity += o.RemainingQuantity()
			level.Orders++
		} else {
			depthMap[key] = &DepthLevel{
				Price:    o.Price,
				Quantity: o.RemainingQuantity(),
				Orders:   1,
			}
		}
	}

	depths := make([]DepthLevel, 0, len(depthMap))
	for _, level := range depthMap {
		depths = append(depths, *level)
	}

	// Sort by price ascending (lowest first)
	sort.Slice(depths, func(i, j int) bool {
		return depths[i].Price.Cmp(&depths[j].Price) < 0
	})

	if levels > 0 && levels < len(depths) {
		return depths[:levels]
	}
	return depths
}

// Stats returns order book statistics
type OrderBookStats struct {
	Symbol       string      `json:"symbol"`
	BidCount     int         `json:"bid_count"`
	AskCount     int         `json:"ask_count"`
	BestBidPrice apd.Decimal `json:"best_bid_price"`
	BestAskPrice apd.Decimal `json:"best_ask_price"`
	Spread       apd.Decimal `json:"spread"`
	MidPrice     apd.Decimal `json:"mid_price"`
}

// GetStats returns order book statistics
func (ob *OrderBook) GetStats() OrderBookStats {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	stats := OrderBookStats{
		Symbol:   ob.Symbol,
		BidCount: ob.Bids.Len(),
		AskCount: ob.Asks.Len(),
	}

	bestBid := ob.BestBid()
	bestAsk := ob.BestAsk()

	if bestBid != nil {
		stats.BestBidPrice = bestBid.Price
	}
	if bestAsk != nil {
		stats.BestAskPrice = bestAsk.Price
	}

	stats.Spread = ob.Spread()
	stats.MidPrice = ob.MidPrice()

	return stats
}

// Clear removes all orders from the order book
func (ob *OrderBook) Clear() {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	ob.Bids = &BidQueue{}
	ob.Asks = &AskQueue{}
	heap.Init(ob.Bids)
	heap.Init(ob.Asks)
}

// ===== Priority Queue Implementations =====

// BidQueue is a max heap for buy orders (highest price first, then earliest time)
type BidQueue []*order.Order

func (bq BidQueue) Len() int { return len(bq) }

func (bq BidQueue) Less(i, j int) bool {
	// Max heap: higher price has priority
	cmp := bq[i].Price.Cmp(&bq[j].Price)
	if cmp != 0 {
		return cmp > 0
	}
	// If same price, earlier time has priority (FIFO)
	return bq[i].CreatedAt.Before(bq[j].CreatedAt)
}

func (bq BidQueue) Swap(i, j int) {
	bq[i], bq[j] = bq[j], bq[i]
}

func (bq *BidQueue) Push(x interface{}) {
	*bq = append(*bq, x.(*order.Order))
}

func (bq *BidQueue) Pop() interface{} {
	old := *bq
	n := len(old)
	item := old[n-1]
	*bq = old[0 : n-1]
	return item
}

// AskQueue is a min heap for sell orders (lowest price first, then earliest time)
type AskQueue []*order.Order

func (aq AskQueue) Len() int { return len(aq) }

func (aq AskQueue) Less(i, j int) bool {
	// Min heap: lower price has priority
	cmp := aq[i].Price.Cmp(&aq[j].Price)
	if cmp != 0 {
		return cmp < 0
	}
	// If same price, earlier time has priority (FIFO)
	return aq[i].CreatedAt.Before(aq[j].CreatedAt)
}

func (aq AskQueue) Swap(i, j int) {
	aq[i], aq[j] = aq[j], aq[i]
}

func (aq *AskQueue) Push(x interface{}) {
	*aq = append(*aq, x.(*order.Order))
}

func (aq *AskQueue) Pop() interface{} {
	old := *aq
	n := len(old)
	item := old[n-1]
	*aq = old[0 : n-1]
	return item
}
