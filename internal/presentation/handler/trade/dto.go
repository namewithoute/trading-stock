package trade

import (
	"time"

	"trading-stock/internal/domain/execution"
)

// TradeDTO is the public representation of a trade.
type TradeDTO struct {
	ID          string     `json:"id"`
	BuyOrderID  string     `json:"buy_order_id"`
	SellOrderID string     `json:"sell_order_id"`
	Symbol      string     `json:"symbol"`
	Price       float64    `json:"price"`
	Quantity    int        `json:"quantity"`
	TotalValue  float64    `json:"total_value"`
	BuyerID     string     `json:"buyer_id"`
	SellerID    string     `json:"seller_id"`
	Status      string     `json:"status"`
	SettledAt   *time.Time `json:"settled_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

func toTradeDTO(t *execution.Trade) TradeDTO {
	return TradeDTO{
		ID:          t.ID,
		BuyOrderID:  t.BuyOrderID,
		SellOrderID: t.SellOrderID,
		Symbol:      t.Symbol,
		Price:       t.Price,
		Quantity:    t.Quantity,
		TotalValue:  t.TotalValue(),
		BuyerID:     t.BuyerID,
		SellerID:    t.SellerID,
		Status:      string(t.Status),
		SettledAt:   t.SettledAt,
		CreatedAt:   t.CreatedAt,
	}
}
