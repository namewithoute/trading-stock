package engine

import (
	"container/heap"
	"fmt"
	"sync"

	"trading-stock/internal/domain/order"
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
func (ob *OrderBook) Spread() float64 {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	bestBid := ob.BestBid()
	bestAsk := ob.BestAsk()

	if bestBid == nil || bestAsk == nil {
		return 0
	}

	return bestAsk.Price - bestBid.Price
}

// MidPrice returns the mid price between best bid and ask
func (ob *OrderBook) MidPrice() float64 {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	bestBid := ob.BestBid()
	bestAsk := ob.BestAsk()

	if bestBid == nil || bestAsk == nil {
		return 0
	}

	return (bestBid.Price + bestAsk.Price) / 2
}

// Depth returns the total quantity at each price level
type DepthLevel struct {
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
	Orders   int     `json:"orders"` // Number of orders at this level
}

// GetBidDepth returns aggregated bid depth
func (ob *OrderBook) GetBidDepth(levels int) []DepthLevel {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	depthMap := make(map[float64]*DepthLevel)

	for _, o := range *ob.Bids {
		if level, exists := depthMap[o.Price]; exists {
			level.Quantity += o.RemainingQuantity()
			level.Orders++
		} else {
			depthMap[o.Price] = &DepthLevel{
				Price:    o.Price,
				Quantity: o.RemainingQuantity(),
				Orders:   1,
			}
		}
	}

	// Convert map to slice and sort by price DESC
	depths := make([]DepthLevel, 0, len(depthMap))
	for _, level := range depthMap {
		depths = append(depths, *level)
	}

	// Sort by price descending (highest first)
	for i := 0; i < len(depths)-1; i++ {
		for j := i + 1; j < len(depths); j++ {
			if depths[i].Price < depths[j].Price {
				depths[i], depths[j] = depths[j], depths[i]
			}
		}
	}

	// Return only requested number of levels
	if levels > 0 && levels < len(depths) {
		return depths[:levels]
	}
	return depths
}

// GetAskDepth returns aggregated ask depth
func (ob *OrderBook) GetAskDepth(levels int) []DepthLevel {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	depthMap := make(map[float64]*DepthLevel)

	for _, o := range *ob.Asks {
		if level, exists := depthMap[o.Price]; exists {
			level.Quantity += o.RemainingQuantity()
			level.Orders++
		} else {
			depthMap[o.Price] = &DepthLevel{
				Price:    o.Price,
				Quantity: o.RemainingQuantity(),
				Orders:   1,
			}
		}
	}

	// Convert map to slice and sort by price ASC
	depths := make([]DepthLevel, 0, len(depthMap))
	for _, level := range depthMap {
		depths = append(depths, *level)
	}

	// Sort by price ascending (lowest first)
	for i := 0; i < len(depths)-1; i++ {
		for j := i + 1; j < len(depths); j++ {
			if depths[i].Price > depths[j].Price {
				depths[i], depths[j] = depths[j], depths[i]
			}
		}
	}

	// Return only requested number of levels
	if levels > 0 && levels < len(depths) {
		return depths[:levels]
	}
	return depths
}

// Stats returns order book statistics
type OrderBookStats struct {
	Symbol       string  `json:"symbol"`
	BidCount     int     `json:"bid_count"`
	AskCount     int     `json:"ask_count"`
	BestBidPrice float64 `json:"best_bid_price"`
	BestAskPrice float64 `json:"best_ask_price"`
	Spread       float64 `json:"spread"`
	MidPrice     float64 `json:"mid_price"`
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
	if bq[i].Price != bq[j].Price {
		return bq[i].Price > bq[j].Price
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
	if aq[i].Price != aq[j].Price {
		return aq[i].Price < aq[j].Price
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
