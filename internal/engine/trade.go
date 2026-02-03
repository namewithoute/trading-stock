package engine

import (
	"time"

	"github.com/google/uuid"
)

// Trade represents a matched trade between a buy and sell order
type Trade struct {
	ID          string    `json:"id"`
	BuyOrderID  string    `json:"buy_order_id"`
	SellOrderID string    `json:"sell_order_id"`
	Symbol      string    `json:"symbol"`
	Price       float64   `json:"price"`
	Quantity    int       `json:"quantity"`
	BuyerID     string    `json:"buyer_id"`
	SellerID    string    `json:"seller_id"`
	Timestamp   time.Time `json:"timestamp"`
}

// NewTrade creates a new trade
func NewTrade(buyOrderID, sellOrderID, symbol string, price float64, quantity int, buyerID, sellerID string) *Trade {
	return &Trade{
		ID:          uuid.New().String(),
		BuyOrderID:  buyOrderID,
		SellOrderID: sellOrderID,
		Symbol:      symbol,
		Price:       price,
		Quantity:    quantity,
		BuyerID:     buyerID,
		SellerID:    sellerID,
		Timestamp:   time.Now(),
	}
}

// TotalValue returns the total value of the trade
func (t *Trade) TotalValue() float64 {
	return t.Price * float64(t.Quantity)
}
