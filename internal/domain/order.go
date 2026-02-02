package domain

import (
	"time"
)

// Order: Ý định của nhà đầu tư
type Order struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"` // Cần biết lệnh này của ai
	Symbol    string    `json:"symbol"`
	Price     float64   `json:"price"`
	Quantity  int       `json:"quantity"`
	Side      Side      `json:"side"`       // Thêm vào để phân biệt Mua/Bán
	OrderType string    `json:"order_type"` // LIMIT, MARKET...
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"` // Dùng time.Time thay vì string
	UpdatedAt time.Time `json:"updated_at"`
}

type Side string

const (
	Buy  Side = "BUY"
	Sell Side = "SELL"
)
