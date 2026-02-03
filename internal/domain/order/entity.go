package order

import "time"

// Order represents a trading order entity
// This is the investor's intention to buy or sell a security
type Order struct {
	ID        string    `json:"id" gorm:"primaryKey;type:uuid"`
	UserID    string    `json:"user_id" gorm:"type:uuid;index;not null"`
	AccountID string    `json:"account_id" gorm:"type:uuid;index"` // Link to trading account
	Symbol    string    `json:"symbol" gorm:"type:varchar(10);index;not null"`
	Price     float64   `json:"price" gorm:"type:decimal(20,4);not null"`
	Quantity  int       `json:"quantity" gorm:"not null"`
	Side      Side      `json:"side" gorm:"type:varchar(4);not null"`
	Type      OrderType `json:"order_type" gorm:"type:varchar(20);not null"`
	Status    Status    `json:"status" gorm:"type:varchar(20);index;not null"`

	// Execution tracking
	FilledQuantity int     `json:"filled_quantity" gorm:"default:0"`
	AvgFillPrice   float64 `json:"avg_fill_price" gorm:"type:decimal(20,4)"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt time.Time `json:"updated_at" gorm:"not null"`
}

// TableName specifies the table name for GORM
func (Order) TableName() string {
	return "orders"
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
