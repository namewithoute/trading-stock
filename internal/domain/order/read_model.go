package order

import (
	"time"

	"github.com/cockroachdb/apd/v3"
)

// OrderReadModel is the denormalised, query-optimised view of an order.
type OrderReadModel struct {
	ID             string      `json:"id"`
	UserID         string      `json:"user_id"`
	AccountID      string      `json:"account_id"`
	Symbol         string      `json:"symbol"`
	Side           Side        `json:"side"`
	OrderType      OrderType   `json:"order_type"`
	Quantity       int         `json:"quantity"`
	Price          apd.Decimal `json:"price"`
	FilledQuantity int         `json:"filled_quantity"`
	AvgFillPrice   apd.Decimal `json:"avg_fill_price"`
	Status         Status      `json:"status"`

	// Version matches the latest event version applied to this read model.
	// Used by the idempotent Projector to skip duplicate/out-of-order events.
	Version   int       `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
