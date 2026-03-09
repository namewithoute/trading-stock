package order

import (
	"time"

	"github.com/cockroachdb/apd/v3"
)

// ─────────────────────────────────────────────────────────────────────────────
// DomainEvent — base interface for all events in the Order bounded context.
// ─────────────────────────────────────────────────────────────────────────────

// EventType categorises every domain event emitted by this aggregate.
type EventType string

const (
	EventOrderPlaced      EventType = "order.placed"
	EventOrderCancelled   EventType = "order.cancelled"
	EventOrderPartialFill EventType = "order.partial_fill"
	EventOrderFilled      EventType = "order.filled"
	EventOrderRejected    EventType = "order.rejected"
	EventOrderExpired     EventType = "order.expired"
)

// DomainEvent is the contract every event in this aggregate must satisfy.
type DomainEvent interface {
	GetEventType() EventType
	GetAggregateID() string
	GetOccurredAt() time.Time
}

// ─────────────────────────────────────────────────────────────────────────────
// Concrete Events
// ─────────────────────────────────────────────────────────────────────────────

// OrderPlacedEvent is emitted when a new order is submitted by a user.
type OrderPlacedEvent struct {
	AggregateID string      `json:"aggregate_id"`
	UserID      string      `json:"user_id"`
	AccountID   string      `json:"account_id"`
	Symbol      string      `json:"symbol"`
	Side        Side        `json:"side"`
	OrderType   OrderType   `json:"order_type"`
	Quantity    int         `json:"quantity"`
	Price       apd.Decimal `json:"price"`
	OccurredAt  time.Time   `json:"occurred_at"`
}

func (e OrderPlacedEvent) GetEventType() EventType  { return EventOrderPlaced }
func (e OrderPlacedEvent) GetAggregateID() string   { return e.AggregateID }
func (e OrderPlacedEvent) GetOccurredAt() time.Time { return e.OccurredAt }

// OrderCancelledEvent is emitted when a user cancels a pending order.
type OrderCancelledEvent struct {
	AggregateID string    `json:"aggregate_id"`
	OccurredAt  time.Time `json:"occurred_at"`
}

func (e OrderCancelledEvent) GetEventType() EventType  { return EventOrderCancelled }
func (e OrderCancelledEvent) GetAggregateID() string   { return e.AggregateID }
func (e OrderCancelledEvent) GetOccurredAt() time.Time { return e.OccurredAt }

// OrderPartialFillEvent is emitted when the matching engine partially executes an order.
type OrderPartialFillEvent struct {
	AggregateID    string      `json:"aggregate_id"`
	FilledQty      int         `json:"filled_qty"`       // quantity filled in THIS trade
	FillPrice      apd.Decimal `json:"fill_price"`       // price of THIS trade
	TotalFilledQty int         `json:"total_filled_qty"` // cumulative filled quantity
	OccurredAt     time.Time   `json:"occurred_at"`
}

func (e OrderPartialFillEvent) GetEventType() EventType  { return EventOrderPartialFill }
func (e OrderPartialFillEvent) GetAggregateID() string   { return e.AggregateID }
func (e OrderPartialFillEvent) GetOccurredAt() time.Time { return e.OccurredAt }

// OrderFilledEvent is emitted when the order is completely executed.
type OrderFilledEvent struct {
	AggregateID    string      `json:"aggregate_id"`
	FilledQty      int         `json:"filled_qty"`
	FillPrice      apd.Decimal `json:"fill_price"`
	TotalFilledQty int         `json:"total_filled_qty"`
	AvgFillPrice   apd.Decimal `json:"avg_fill_price"`
	OccurredAt     time.Time   `json:"occurred_at"`
}

func (e OrderFilledEvent) GetEventType() EventType  { return EventOrderFilled }
func (e OrderFilledEvent) GetAggregateID() string   { return e.AggregateID }
func (e OrderFilledEvent) GetOccurredAt() time.Time { return e.OccurredAt }

// OrderRejectedEvent is emitted when the system rejects an order (e.g. risk check failed).
type OrderRejectedEvent struct {
	AggregateID string    `json:"aggregate_id"`
	Reason      string    `json:"reason"`
	OccurredAt  time.Time `json:"occurred_at"`
}

func (e OrderRejectedEvent) GetEventType() EventType  { return EventOrderRejected }
func (e OrderRejectedEvent) GetAggregateID() string   { return e.AggregateID }
func (e OrderRejectedEvent) GetOccurredAt() time.Time { return e.OccurredAt }

// OrderExpiredEvent is emitted when a time-limited order reaches its expiry.
type OrderExpiredEvent struct {
	AggregateID string    `json:"aggregate_id"`
	OccurredAt  time.Time `json:"occurred_at"`
}

func (e OrderExpiredEvent) GetEventType() EventType  { return EventOrderExpired }
func (e OrderExpiredEvent) GetAggregateID() string   { return e.AggregateID }
func (e OrderExpiredEvent) GetOccurredAt() time.Time { return e.OccurredAt }
