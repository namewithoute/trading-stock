package engine

import (
	"time"

	"github.com/cockroachdb/apd/v3"
	"github.com/google/uuid"
)

var decCtx = apd.BaseContext.WithPrecision(19)

// Trade represents a matched trade between a buy and sell order
type Trade struct {
	ID          string      `json:"id"`
	BuyOrderID  string      `json:"buy_order_id"`
	SellOrderID string      `json:"sell_order_id"`
	Symbol      string      `json:"symbol"`
	Price       apd.Decimal `json:"price"`
	Quantity    int         `json:"quantity"`
	BuyerID     string      `json:"buyer_id"`
	SellerID    string      `json:"seller_id"`
	Timestamp   time.Time   `json:"timestamp"`
}

// NewTrade creates a new trade
func NewTrade(buyOrderID, sellOrderID, symbol string, price apd.Decimal, quantity int, buyerID, sellerID string) *Trade {
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
func (t *Trade) TotalValue() apd.Decimal {
	var result apd.Decimal
	_, _ = decCtx.Mul(&result, &t.Price, apd.New(int64(t.Quantity), 0))
	return result
}
