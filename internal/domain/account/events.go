package account

import "time"

// ─────────────────────────────────────────────────────────────────────────────
// DomainEvent — base interface for all events in the Account bounded context.
// ─────────────────────────────────────────────────────────────────────────────

// EventType categorises every domain event emitted by this aggregate.
type EventType string

const (
	EventAccountCreated EventType = "account.created"
	EventMoneyDeposited EventType = "account.money_deposited"
	EventMoneyWithdrawn EventType = "account.money_withdrawn"
	EventFundsReserved  EventType = "account.funds_reserved"
	EventFundsReleased  EventType = "account.funds_released"
	EventStatusChanged  EventType = "account.status_changed"
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

// AccountCreatedEvent is emitted when a new trading account is opened.
type AccountCreatedEvent struct {
	AggregateID string      `json:"aggregate_id"`
	UserID      string      `json:"user_id"`
	AccountType AccountType `json:"account_type"`
	Currency    string      `json:"currency"`
	OccurredAt  time.Time   `json:"occurred_at"`
}

func (e AccountCreatedEvent) GetEventType() EventType  { return EventAccountCreated }
func (e AccountCreatedEvent) GetAggregateID() string   { return e.AggregateID }
func (e AccountCreatedEvent) GetOccurredAt() time.Time { return e.OccurredAt }

// MoneyDepositedEvent is emitted when funds are added to an account.
type MoneyDepositedEvent struct {
	AggregateID string    `json:"aggregate_id"`
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	OccurredAt  time.Time `json:"occurred_at"`
}

func (e MoneyDepositedEvent) GetEventType() EventType  { return EventMoneyDeposited }
func (e MoneyDepositedEvent) GetAggregateID() string   { return e.AggregateID }
func (e MoneyDepositedEvent) GetOccurredAt() time.Time { return e.OccurredAt }

// MoneyWithdrawnEvent is emitted when funds are removed from an account.
type MoneyWithdrawnEvent struct {
	AggregateID string    `json:"aggregate_id"`
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	OccurredAt  time.Time `json:"occurred_at"`
}

func (e MoneyWithdrawnEvent) GetEventType() EventType  { return EventMoneyWithdrawn }
func (e MoneyWithdrawnEvent) GetAggregateID() string   { return e.AggregateID }
func (e MoneyWithdrawnEvent) GetOccurredAt() time.Time { return e.OccurredAt }

// FundsReservedEvent is emitted when buying power is reduced for a pending order.
type FundsReservedEvent struct {
	AggregateID string    `json:"aggregate_id"`
	Amount      float64   `json:"amount"`
	OccurredAt  time.Time `json:"occurred_at"`
}

func (e FundsReservedEvent) GetEventType() EventType  { return EventFundsReserved }
func (e FundsReservedEvent) GetAggregateID() string   { return e.AggregateID }
func (e FundsReservedEvent) GetOccurredAt() time.Time { return e.OccurredAt }

// FundsReleasedEvent is emitted when reserved buying power is returned (order cancelled).
type FundsReleasedEvent struct {
	AggregateID string    `json:"aggregate_id"`
	Amount      float64   `json:"amount"`
	OccurredAt  time.Time `json:"occurred_at"`
}

func (e FundsReleasedEvent) GetEventType() EventType  { return EventFundsReleased }
func (e FundsReleasedEvent) GetAggregateID() string   { return e.AggregateID }
func (e FundsReleasedEvent) GetOccurredAt() time.Time { return e.OccurredAt }

// StatusChangedEvent is emitted when account status transitions (e.g. ACTIVE → FROZEN).
type StatusChangedEvent struct {
	AggregateID string    `json:"aggregate_id"`
	OldStatus   Status    `json:"old_status"`
	NewStatus   Status    `json:"new_status"`
	OccurredAt  time.Time `json:"occurred_at"`
}

func (e StatusChangedEvent) GetEventType() EventType  { return EventStatusChanged }
func (e StatusChangedEvent) GetAggregateID() string   { return e.AggregateID }
func (e StatusChangedEvent) GetOccurredAt() time.Time { return e.OccurredAt }
