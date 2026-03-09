package order

import (
	"time"

	"github.com/cockroachdb/apd/v3"
)

// Order represents a trading order entity
// This is the investor's intention to buy or sell a security
type Order struct {
	ID        string
	UserID    string
	AccountID string
	Symbol    string
	Price     apd.Decimal
	Quantity  int
	Side      Side
	Type      OrderType
	Status    Status

	FilledQuantity int
	AvgFillPrice   apd.Decimal

	CreatedAt time.Time
	UpdatedAt time.Time
}

// IsFullyFilled checks if the order is completely filled
func (o *Order) IsFullyFilled() bool {
	return o.FilledQuantity >= o.Quantity
}

// IsPartiallyFilled checks if the order is partially filled
func (o *Order) IsPartiallyFilled() bool {
	return o.FilledQuantity > 0 && o.FilledQuantity < o.Quantity
}

// RemainingQuantity returns the unfilled quantity
func (o *Order) RemainingQuantity() int {
	return o.Quantity - o.FilledQuantity
}

// CanBeCancelled checks if the order can be cancelled
func (o *Order) CanBeCancelled() bool {
	return o.Status == StatusPending || o.Status == StatusPartiallyFilled
}

// CanBeModified checks if the order can be modified
func (o *Order) CanBeModified() bool {
	return o.Status == StatusPending
}
