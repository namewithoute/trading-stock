package account

import (
	"time"

	"github.com/cockroachdb/apd/v3"
)

// ─────────────────────────────────────────────────────────────────────────────
// AccountAggregate — the write-side (Event Sourcing) representation.
//
// DDD rules enforced here:
//  1. State is NEVER set directly — always mutated via Apply().
//  2. Commands validate invariants BEFORE producing events.
//  3. Uncommitted events accumulate until the UseCase calls Save().
// ─────────────────────────────────────────────────────────────────────────────

// AccountAggregate is the Aggregate Root for the Account bounded context.
type AccountAggregate struct {
	// ── Identity ──────────────────────────────────────────────────────────────
	ID     string
	UserID string

	// ── Value Objects ─────────────────────────────────────────────────────────
	AccountType AccountType
	Money       Money
	Status      Status

	// ── Event Sourcing Metadata ───────────────────────────────────────────────
	// Version is the sequence number of the last applied event.
	// Used for optimistic concurrency control in the EventStore.
	Version int

	// uncommittedEvents holds events emitted by the current command,
	// not yet persisted. Drained by the Repository.Save() implementation.
	uncommittedEvents []DomainEvent
}

// ─────────────────────────────────────────────────────────────────────────────
// Rehydration (replay from EventStore)
// ─────────────────────────────────────────────────────────────────────────────

// RehydrateAccount rebuilds an aggregate by replaying its full event history.
// Called by the infrastructure Repository implementation.
func RehydrateAccount(events []DomainEvent) *AccountAggregate {
	agg := &AccountAggregate{}
	for _, e := range events {
		agg.apply(e, false) // false = historical event, do not enqueue
	}
	return agg
}

// ─────────────────────────────────────────────────────────────────────────────
// Apply — state mutation (ONLY entry point that changes aggregate state)
// ─────────────────────────────────────────────────────────────────────────────

// apply mutates aggregate state for a given event.
//
//   - isNew = true  → event was produced by the current command (enqueue for Save)
//   - isNew = false → historical event being replayed (do not enqueue)
func (a *AccountAggregate) apply(event DomainEvent, isNew bool) {
	switch e := event.(type) {
	case AccountCreatedEvent:
		a.ID = e.AggregateID
		a.UserID = e.UserID
		a.AccountType = e.AccountType
		a.Money = Money{Currency: e.Currency} // Balance and BuyingPower zero-valued
		a.Status = StatusActive

	case MoneyDepositedEvent:
		_, _ = decCtx.Add(&a.Money.Balance, &a.Money.Balance, &e.Amount.Decimal)
		_, _ = decCtx.Add(&a.Money.BuyingPower, &a.Money.BuyingPower, &e.Amount.Decimal)

	case MoneyWithdrawnEvent:
		_, _ = decCtx.Sub(&a.Money.Balance, &a.Money.Balance, &e.Amount.Decimal)
		_, _ = decCtx.Sub(&a.Money.BuyingPower, &a.Money.BuyingPower, &e.Amount.Decimal)

	case FundsReservedEvent:
		_, _ = decCtx.Sub(&a.Money.BuyingPower, &a.Money.BuyingPower, &e.Amount.Decimal)

	case FundsReleasedEvent:
		_, _ = decCtx.Add(&a.Money.BuyingPower, &a.Money.BuyingPower, &e.Amount.Decimal)

	case StatusChangedEvent:
		a.Status = e.NewStatus

	case TradeSettledEvent:
		if e.Side == "BUY" {
			// Funds were already reserved (BuyingPower reduced); now debit the actual balance.
			_, _ = decCtx.Sub(&a.Money.Balance, &a.Money.Balance, &e.Amount.Decimal)
		} else {
			// SELL settlement: cash received from selling securities.
			_, _ = decCtx.Add(&a.Money.Balance, &a.Money.Balance, &e.Amount.Decimal)
			_, _ = decCtx.Add(&a.Money.BuyingPower, &a.Money.BuyingPower, &e.Amount.Decimal)
		}
	}

	a.Version++

	if isNew {
		a.uncommittedEvents = append(a.uncommittedEvents, event)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Uncommitted Events — lifecycle managed by UseCase / EventSourcingService
// ─────────────────────────────────────────────────────────────────────────────

// UncommittedEvents returns events not yet persisted to the EventStore.
func (a *AccountAggregate) UncommittedEvents() []DomainEvent {
	return a.uncommittedEvents
}

// ClearUncommittedEvents is called by the Repository after successful persistence.
func (a *AccountAggregate) ClearUncommittedEvents() {
	a.uncommittedEvents = nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Projection helper
// ─────────────────────────────────────────────────────────────────────────────

// ToReadModel converts current aggregate state to the query-optimised read model.
func (a *AccountAggregate) ToReadModel() *AccountReadModel {
	balance := apd.Decimal{}
	_, _ = decCtx.Add(&balance, &a.Money.Balance, &apd.Decimal{})
	buyingPower := apd.Decimal{}
	_, _ = decCtx.Add(&buyingPower, &a.Money.BuyingPower, &apd.Decimal{})
	return &AccountReadModel{
		ID:          a.ID,
		UserID:      a.UserID,
		AccountType: a.AccountType,
		Balance:     balance,
		BuyingPower: buyingPower,
		Currency:    a.Money.Currency,
		Status:      a.Status,
		Version:     a.Version,
		UpdatedAt:   time.Now().UTC(),
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Internal helper
// ─────────────────────────────────────────────────────────────────────────────

// nowUTC is the canonical timestamp factory for all domain events.
func nowUTC() time.Time {
	return time.Now().UTC()
}
