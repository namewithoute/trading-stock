package order

import "time"

type CreateOrderRequest struct {
	Symbol   string  `json:"symbol"    validate:"required"`
	Side     string  `json:"side"      validate:"required"`
	Type     string  `json:"type"      validate:"required"`
	Quantity float64 `json:"quantity"  validate:"required"`
	Price    float64 `json:"price"     validate:"required"`
}

type CreateOrderResponse struct {
	OrderID   string  `json:"order_id"`
	Symbol    string  `json:"symbol"`
	Side      string  `json:"side"`
	Type      string  `json:"type"`
	Quantity  float64 `json:"quantity"`
	Price     float64 `json:"price"`
	Status    string  `json:"status"`
	CreatedAt string  `json:"created_at"`
}

type ListOrdersRequest struct {
	Status string `query:"status"`
	Symbol string `query:"symbol"`
	Page   int    `query:"page"`
	Limit  int    `query:"limit"`
}

type ListOrdersResponse struct {
	Orders     []Order    `json:"orders"`
	Pagination Pagination `json:"pagination"`
}

type Order struct {
	OrderID        string    `json:"order_id"`
	Symbol         string    `json:"symbol"`
	Side           string    `json:"side"`
	Type           string    `json:"type"`
	Quantity       int       `json:"quantity"`
	FilledQuantity int       `json:"filled_quantity"`
	Price          float64   `json:"price"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}

type Pagination struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}
