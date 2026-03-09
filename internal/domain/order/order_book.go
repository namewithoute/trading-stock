package order

import (
	"sort"

	"github.com/cockroachdb/apd/v3"
)

var obDecCtx = apd.BaseContext.WithPrecision(19)

// OrderBook represents a collection of buy and sell orders for a specific symbol
// This is used for order matching in the execution engine
type OrderBook struct {
	Symbol string   `json:"symbol"`
	Bids   []*Order `json:"bids"` // Buy orders (sorted by price DESC)
	Asks   []*Order `json:"asks"` // Sell orders (sorted by price ASC)
}

// NewOrderBook creates a new order book for a symbol
func NewOrderBook(symbol string) *OrderBook {
	return &OrderBook{
		Symbol: symbol,
		Bids:   make([]*Order, 0),
		Asks:   make([]*Order, 0),
	}
}

// AddOrder adds an order to the appropriate side of the order book
func (ob *OrderBook) AddOrder(order *Order) {
	if order.Side == SideBuy {
		ob.Bids = append(ob.Bids, order)
		// TODO: Sort bids by price DESC
	} else {
		ob.Asks = append(ob.Asks, order)
		// TODO: Sort asks by price ASC
	}
}

// RemoveOrder removes an order from the order book
func (ob *OrderBook) RemoveOrder(orderID string) {
	// Remove from bids
	for i, order := range ob.Bids {
		if order.ID == orderID {
			ob.Bids = append(ob.Bids[:i], ob.Bids[i+1:]...)
			return
		}
	}

	// Remove from asks
	for i, order := range ob.Asks {
		if order.ID == orderID {
			ob.Asks = append(ob.Asks[:i], ob.Asks[i+1:]...)
			return
		}
	}
}

// BestBid returns the highest buy order price
func (ob *OrderBook) BestBid() *Order {
	if len(ob.Bids) == 0 {
		return nil
	}
	// Assuming bids are sorted DESC
	return ob.Bids[0]
}

// BestAsk returns the lowest sell order price
func (ob *OrderBook) BestAsk() *Order {
	if len(ob.Asks) == 0 {
		return nil
	}
	// Assuming asks are sorted ASC
	return ob.Asks[0]
}

// Spread returns the difference between best ask and best bid
func (ob *OrderBook) Spread() apd.Decimal {
	bestBid := ob.BestBid()
	bestAsk := ob.BestAsk()

	if bestBid == nil || bestAsk == nil {
		return apd.Decimal{}
	}

	var spread apd.Decimal
	_, _ = obDecCtx.Sub(&spread, &bestAsk.Price, &bestBid.Price)
	return spread
}

// Depth returns the total quantity at each price level
type Depth struct {
	Price    apd.Decimal `json:"price"`
	Quantity int         `json:"quantity"`
}

// GetBidDepth returns the bid depth (aggregated by price)
func (ob *OrderBook) GetBidDepth() []Depth {
	depthMap := make(map[string]*Depth)
	for _, order := range ob.Bids {
		key := order.Price.String()
		if d, ok := depthMap[key]; ok {
			d.Quantity += order.RemainingQuantity()
		} else {
			depthMap[key] = &Depth{Price: order.Price, Quantity: order.RemainingQuantity()}
		}
	}

	depths := make([]Depth, 0, len(depthMap))
	for _, d := range depthMap {
		depths = append(depths, *d)
	}
	sort.Slice(depths, func(i, j int) bool { return depths[i].Price.Cmp(&depths[j].Price) > 0 })
	return depths
}

// GetAskDepth returns the ask depth (aggregated by price)
func (ob *OrderBook) GetAskDepth() []Depth {
	depthMap := make(map[string]*Depth)
	for _, order := range ob.Asks {
		key := order.Price.String()
		if d, ok := depthMap[key]; ok {
			d.Quantity += order.RemainingQuantity()
		} else {
			depthMap[key] = &Depth{Price: order.Price, Quantity: order.RemainingQuantity()}
		}
	}

	depths := make([]Depth, 0, len(depthMap))
	for _, d := range depthMap {
		depths = append(depths, *d)
	}
	sort.Slice(depths, func(i, j int) bool { return depths[i].Price.Cmp(&depths[j].Price) < 0 })
	return depths
}
