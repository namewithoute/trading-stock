package order

import "time"

// OrderReadModel is the denormalised, query-optimised view of an order.
// It lives in the `order_read_models` table and is safely reconstructed
// from the EventStore by the Projector.
//
// In CQRS architecture, this model is used STRICTLY for Read operations.
// Writes operate on the OrderAggregate via Event Sourcing.
type OrderReadModel struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	AccountID      string    `json:"account_id"`
	Symbol         string    `json:"symbol"`
	Side           Side      `json:"side"`
	OrderType      OrderType `json:"order_type"`
	Quantity       int       `json:"quantity"`
	Price          float64   `json:"price"`
	FilledQuantity int       `json:"filled_quantity"`
	AvgFillPrice   float64   `json:"avg_fill_price"`
	Status         Status    `json:"status"`

	// Version matches the latest event version applied to this read model.
	// Used by the idempotent Projector to skip duplicate/out-of-order events.
	Version   int       `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
