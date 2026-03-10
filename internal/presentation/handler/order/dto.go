package order

import (
	"time"

	pkgdecimal "trading-stock/pkg/decimal"
)

// ─── Create ──────────────────────────────────────────────────────────────────

// CreateOrderRequest is the request body for placing a new order.
type CreateOrderRequest struct {
	Symbol    string             `json:"symbol"   validate:"required"`
	Side      string             `json:"side"     validate:"required,oneof=BUY SELL"`
	Type      string             `json:"type"     validate:"required,oneof=MARKET LIMIT STOP_LOSS STOP_LIMIT"`
	Quantity  float64            `json:"quantity" validate:"required,gt=0"`
	Price     pkgdecimal.Decimal `json:"price"`
	AccountID string             `json:"account_id"`
}

// ─── List ─────────────────────────────────────────────────────────────────────

// ListOrdersRequest carries query-string filters for listing orders.
type ListOrdersRequest struct {
	Status string `query:"status"`
	Symbol string `query:"symbol"`
	Page   int    `query:"page"`
	Limit  int    `query:"limit"`
}

// ListOrdersResponse wraps a paginated list of orders.
type ListOrdersResponse struct {
	Orders     []OrderDTO `json:"orders"`
	Pagination Pagination `json:"pagination"`
}

// ─── Detail ───────────────────────────────────────────────────────────────────

// GetOrderDetailResponse carries the full order state returned by GET /orders/:id.
type GetOrderDetailResponse struct {
	OrderID        string             `json:"order_id"`
	AccountID      string             `json:"account_id"`
	Symbol         string             `json:"symbol"`
	Side           string             `json:"side"`
	Type           string             `json:"type"`
	Quantity       int                `json:"quantity"`
	FilledQuantity int                `json:"filled_quantity"`
	Price          pkgdecimal.Decimal `json:"price"`
	AvgFillPrice   pkgdecimal.Decimal `json:"avg_fill_price"`
	Status         string             `json:"status"`
	CreatedAt      time.Time          `json:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at"`
}

// ─── Update ───────────────────────────────────────────────────────────────────

// UpdateOrderRequest carries fields that can be changed on a PENDING order.
// In stock trading this is implemented as cancel + recreate internally.
type UpdateOrderRequest struct {
	Price    pkgdecimal.Decimal `json:"price"`
	Quantity int                `json:"quantity" validate:"required,gt=0"`
}

// ─── Shared sub-types ─────────────────────────────────────────────────────────

// OrderDTO is a compact order summary used inside list responses.
type OrderDTO struct {
	OrderID        string             `json:"order_id"`
	Symbol         string             `json:"symbol"`
	Side           string             `json:"side"`
	Type           string             `json:"type"`
	Quantity       int                `json:"quantity"`
	FilledQuantity int                `json:"filled_quantity"`
	Price          pkgdecimal.Decimal `json:"price"`
	Status         string             `json:"status"`
	CreatedAt      time.Time          `json:"created_at"`
}

// Pagination carries page metadata returned alongside list responses.
type Pagination struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

// Order is kept as a type alias of OrderDTO for backward compatibility with
// existing code that references the old Order struct name.
type Order = OrderDTO
