package domain

type OrderBook struct {
	Symbol string
	Bids   []*Order // BUY
	Asks   []*Order // SELL
}

func (ob *OrderBook) ProcessOrder(order *Order) []Trade {
	// Matching logic goes here
	return []Trade{}
}
