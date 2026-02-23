package order

import "time"

// ─────────────────────────────────────────────────────────────────────────────
// OrderAggregate — the write-side (Event Sourcing) representation.
//
// DDD rules enforced here:
//  1. State is NEVER set directly — always mutated via apply().
//  2. Commands (behaviors) validate invariants BEFORE producing events.
//  3. Uncommitted events accumulate until the Repository calls Save().
// ─────────────────────────────────────────────────────────────────────────────

// OrderAggregate is the Aggregate Root for the Order bounded context.
type OrderAggregate struct {
	// ── Identity ──────────────────────────────────────────────────────────────
	ID        string
	UserID    string
	AccountID string

	// ── Value Objects ─────────────────────────────────────────────────────────
	Symbol    string
	Side      Side
	OrderType OrderType
	Quantity  int
	Price     float64

	// ── Fill tracking ─────────────────────────────────────────────────────────
	FilledQuantity int
	AvgFillPrice   float64
	totalFillValue float64 // running sum of (fillQty * fillPrice) for avg calc

	// ── Lifecycle ─────────────────────────────────────────────────────────────
	Status    Status
	CreatedAt time.Time
	UpdatedAt time.Time

	// ── Event Sourcing Metadata ───────────────────────────────────────────────
	Version           int
	uncommittedEvents []DomainEvent
}

// ─────────────────────────────────────────────────────────────────────────────
// Rehydration (replay from EventStore)
// ─────────────────────────────────────────────────────────────────────────────

// RehydrateOrder rebuilds an aggregate by replaying its full event history.
// Called by the infrastructure Repository implementation.
func RehydrateOrder(events []DomainEvent) *OrderAggregate {
	agg := &OrderAggregate{}
	for _, e := range events {
		agg.apply(e, false)
	}
	return agg
}

// ─────────────────────────────────────────────────────────────────────────────
// Apply — state mutation (ONLY entry point that changes aggregate state)
// ─────────────────────────────────────────────────────────────────────────────

func (a *OrderAggregate) apply(event DomainEvent, isNew bool) {
	switch e := event.(type) {
	case OrderPlacedEvent:
		a.ID = e.AggregateID
		a.UserID = e.UserID
		a.AccountID = e.AccountID
		a.Symbol = e.Symbol
		a.Side = e.Side
		a.OrderType = e.OrderType
		a.Quantity = e.Quantity
		a.Price = e.Price
		a.Status = StatusPending
		a.FilledQuantity = 0
		a.AvgFillPrice = 0
		a.totalFillValue = 0
		a.CreatedAt = e.OccurredAt
		a.UpdatedAt = e.OccurredAt

	case OrderCancelledEvent:
		a.Status = StatusCancelled
		a.UpdatedAt = e.OccurredAt

	case OrderPartialFillEvent:
		a.FilledQuantity = e.TotalFilledQty
		a.totalFillValue += float64(e.FilledQty) * e.FillPrice
		a.AvgFillPrice = a.totalFillValue / float64(a.FilledQuantity)
		a.Status = StatusPartiallyFilled
		a.UpdatedAt = e.OccurredAt

	case OrderFilledEvent:
		a.FilledQuantity = e.TotalFilledQty
		a.AvgFillPrice = e.AvgFillPrice
		a.Status = StatusFilled
		a.UpdatedAt = e.OccurredAt

	case OrderRejectedEvent:
		a.Status = StatusRejected
		a.UpdatedAt = e.OccurredAt

	case OrderExpiredEvent:
		a.Status = StatusExpired
		a.UpdatedAt = e.OccurredAt
	}

	a.Version++
	if isNew {
		a.uncommittedEvents = append(a.uncommittedEvents, event)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Behaviors — domain commands that validate invariants then emit events
// ─────────────────────────────────────────────────────────────────────────────

// PlaceOrder is the aggregate constructor: validates and emits OrderPlacedEvent.
func PlaceOrder(id, userID, accountID, symbol string, side Side, orderType OrderType, quantity int, price float64) (*OrderAggregate, error) {
	if userID == "" {
		return nil, ErrInvalidSide // reuse domain errors; or add ErrInvalidUserID
	}
	if !side.IsValid() {
		return nil, ErrInvalidSide
	}
	if !orderType.IsValid() {
		return nil, ErrInvalidOrderType
	}
	if quantity <= 0 {
		return nil, ErrInvalidOrderType // placeholder; quantity validated here
	}
	if price < 0 {
		return nil, ErrInvalidOrderType
	}

	agg := &OrderAggregate{}
	agg.apply(OrderPlacedEvent{
		AggregateID: id,
		UserID:      userID,
		AccountID:   accountID,
		Symbol:      symbol,
		Side:        side,
		OrderType:   orderType,
		Quantity:    quantity,
		Price:       price,
		OccurredAt:  nowUTC(),
	}, true)
	return agg, nil
}

// Cancel cancels the order if it is in a cancellable state.
func (a *OrderAggregate) Cancel() error {
	if !a.CanBeCancelled() {
		return ErrInvalidStatus
	}
	a.apply(OrderCancelledEvent{
		AggregateID: a.ID,
		OccurredAt:  nowUTC(),
	}, true)
	return nil
}

// RecordFill records a (partial or full) execution from the matching engine.
func (a *OrderAggregate) RecordFill(filledQty int, fillPrice float64) error {
	if a.Status == StatusCancelled || a.Status == StatusFilled {
		return ErrInvalidStatus
	}
	if filledQty <= 0 {
		return ErrInvalidOrderType
	}

	newTotalFilled := a.FilledQuantity + filledQty
	newTotalValue := a.totalFillValue + float64(filledQty)*fillPrice
	newAvg := newTotalValue / float64(newTotalFilled)

	if newTotalFilled >= a.Quantity {
		// Order completely filled
		a.apply(OrderFilledEvent{
			AggregateID:    a.ID,
			FilledQty:      filledQty,
			FillPrice:      fillPrice,
			TotalFilledQty: newTotalFilled,
			AvgFillPrice:   newAvg,
			OccurredAt:     nowUTC(),
		}, true)
	} else {
		// Partial fill
		a.apply(OrderPartialFillEvent{
			AggregateID:    a.ID,
			FilledQty:      filledQty,
			FillPrice:      fillPrice,
			TotalFilledQty: newTotalFilled,
			OccurredAt:     nowUTC(),
		}, true)
	}
	return nil
}

// Reject marks the order as rejected (e.g. risk check failed post-submission).
func (a *OrderAggregate) Reject(reason string) error {
	if a.Status != StatusPending {
		return ErrInvalidStatus
	}
	a.apply(OrderRejectedEvent{
		AggregateID: a.ID,
		Reason:      reason,
		OccurredAt:  nowUTC(),
	}, true)
	return nil
}

// Expire marks the order as expired (called by scheduler for day orders).
func (a *OrderAggregate) Expire() error {
	if !a.CanBeCancelled() {
		return ErrInvalidStatus
	}
	a.apply(OrderExpiredEvent{
		AggregateID: a.ID,
		OccurredAt:  nowUTC(),
	}, true)
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Query helpers (delegated from entity.go logic)
// ─────────────────────────────────────────────────────────────────────────────

func (a *OrderAggregate) CanBeCancelled() bool {
	return a.Status == StatusPending || a.Status == StatusPartiallyFilled
}

func (a *OrderAggregate) CanBeModified() bool {
	return a.Status == StatusPending
}

func (a *OrderAggregate) RemainingQuantity() int {
	return a.Quantity - a.FilledQuantity
}

// ─────────────────────────────────────────────────────────────────────────────
// Uncommitted Events — lifecycle managed by EventSourcingService
// ─────────────────────────────────────────────────────────────────────────────

func (a *OrderAggregate) UncommittedEvents() []DomainEvent {
	return a.uncommittedEvents
}

func (a *OrderAggregate) ClearUncommittedEvents() {
	a.uncommittedEvents = nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Projection helper
// ─────────────────────────────────────────────────────────────────────────────

// ToReadModel converts the aggregate to the query-optimised read model.
func (a *OrderAggregate) ToReadModel() *OrderReadModel {
	return &OrderReadModel{
		ID:             a.ID,
		UserID:         a.UserID,
		AccountID:      a.AccountID,
		Symbol:         a.Symbol,
		Side:           a.Side,
		OrderType:      a.OrderType,
		Quantity:       a.Quantity,
		Price:          a.Price,
		FilledQuantity: a.FilledQuantity,
		AvgFillPrice:   a.AvgFillPrice,
		Status:         a.Status,
		Version:        a.Version,
		CreatedAt:      a.CreatedAt,
		UpdatedAt:      a.UpdatedAt,
	}
}

// nowUTC returns the current UTC time.
func nowUTC() time.Time {
	return time.Now().UTC()
}
