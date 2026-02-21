package account

import "time"

// ─────────────────────────────────────────────────────────────────────────────
// Event Types
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

// DomainEvent is the base interface every event in this aggregate must satisfy.
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

// ─────────────────────────────────────────────────────────────────────────────
// EventDescriptor – the envelope persisted in the event store table.
// ─────────────────────────────────────────────────────────────────────────────

// EventDescriptor is stored in `account_events` with the JSON payload.
type EventDescriptor struct {
	ID          string    `json:"id"`           // UUID per event row
	AggregateID string    `json:"aggregate_id"` // Account ID
	EventType   EventType `json:"event_type"`
	Payload     []byte    `json:"payload"` // JSON-encoded concrete event
	Version     int       `json:"version"` // Monotonically increasing per aggregate
	OccurredAt  time.Time `json:"occurred_at"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Aggregate Root – Event Sourcing version of Account
// ─────────────────────────────────────────────────────────────────────────────

// AccountAggregate is the write-side representation of an account.
// Its state is rebuilt exclusively by replaying DomainEvents.
type AccountAggregate struct {
	// Derived read state (rebuilt via Apply)
	ID          string
	UserID      string
	AccountType AccountType
	Money       Money
	Status      Status
	Version     int

	// Uncommitted changes produced by Commands – stored then cleared by UseCase.
	uncommittedEvents []DomainEvent
}

// RehydrateAccount rebuilds an aggregate from its full event history (replays).
func RehydrateAccount(events []DomainEvent) *AccountAggregate {
	agg := &AccountAggregate{}
	for _, e := range events {
		agg.apply(e, false)
	}
	return agg
}

// apply mutates internal state for a given event.
// isNew=true means the event was produced by the current command (enqueue for saving).
func (a *AccountAggregate) apply(event DomainEvent, isNew bool) {
	switch e := event.(type) {
	case AccountCreatedEvent:
		a.ID = e.AggregateID
		a.UserID = e.UserID
		a.AccountType = e.AccountType
		a.Money = Money{Balance: 0, BuyingPower: 0, Currency: e.Currency}
		a.Status = StatusActive

	case MoneyDepositedEvent:
		a.Money.Balance += e.Amount
		a.Money.BuyingPower += e.Amount

	case MoneyWithdrawnEvent:
		a.Money.Balance -= e.Amount
		a.Money.BuyingPower -= e.Amount

	case FundsReservedEvent:
		a.Money.BuyingPower -= e.Amount

	case FundsReleasedEvent:
		a.Money.BuyingPower += e.Amount

	case StatusChangedEvent:
		a.Status = e.NewStatus
	}

	a.Version++

	if isNew {
		a.uncommittedEvents = append(a.uncommittedEvents, event)
	}
}

// ─── Commands ─────────────────────────────────────────────────────────────────

// OpenAccount command – produces AccountCreatedEvent.
func OpenAccount(id, userID string, accountType AccountType, currency string) *AccountAggregate {
	agg := &AccountAggregate{}
	agg.apply(AccountCreatedEvent{
		AggregateID: id,
		UserID:      userID,
		AccountType: accountType,
		Currency:    currency,
		OccurredAt:  nowUTC(),
	}, true)
	return agg
}

// Deposit command – produces MoneyDepositedEvent.
func (a *AccountAggregate) Deposit(amount float64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	a.apply(MoneyDepositedEvent{
		AggregateID: a.ID,
		Amount:      amount,
		Currency:    a.Money.Currency,
		OccurredAt:  nowUTC(),
	}, true)
	return nil
}

// Withdraw command – produces MoneyWithdrawnEvent.
func (a *AccountAggregate) Withdraw(amount float64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	if a.Money.Balance < amount {
		return ErrInsufficientBalance
	}
	a.apply(MoneyWithdrawnEvent{
		AggregateID: a.ID,
		Amount:      amount,
		Currency:    a.Money.Currency,
		OccurredAt:  nowUTC(),
	}, true)
	return nil
}

// ReserveFunds command – reduces BuyingPower.
func (a *AccountAggregate) ReserveFunds(amount float64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	if a.Money.BuyingPower < amount {
		return ErrInsufficientBuyingPower
	}
	a.apply(FundsReservedEvent{
		AggregateID: a.ID,
		Amount:      amount,
		OccurredAt:  nowUTC(),
	}, true)
	return nil
}

// ReleaseFunds command – returns reserved BuyingPower.
func (a *AccountAggregate) ReleaseFunds(amount float64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	a.apply(FundsReleasedEvent{
		AggregateID: a.ID,
		Amount:      amount,
		OccurredAt:  nowUTC(),
	}, true)
	return nil
}

// ChangeStatus command – transitions account status.
func (a *AccountAggregate) ChangeStatus(newStatus Status) {
	a.apply(StatusChangedEvent{
		AggregateID: a.ID,
		OldStatus:   a.Status,
		NewStatus:   newStatus,
		OccurredAt:  nowUTC(),
	}, true)
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

// UncommittedEvents returns events that have not yet been persisted.
func (a *AccountAggregate) UncommittedEvents() []DomainEvent {
	return a.uncommittedEvents
}

// ClearUncommittedEvents should be called by the repository after successful save.
func (a *AccountAggregate) ClearUncommittedEvents() {
	a.uncommittedEvents = nil
}

// ToReadModel converts aggregate state into the read model struct.
func (a *AccountAggregate) ToReadModel() *AccountReadModel {
	return &AccountReadModel{
		ID:          a.ID,
		UserID:      a.UserID,
		AccountType: a.AccountType,
		Balance:     a.Money.Balance,
		BuyingPower: a.Money.BuyingPower,
		Currency:    a.Money.Currency,
		Status:      a.Status,
		Version:     a.Version,
	}
}

// nowUTC is a small helper to keep event timestamps consistent.
func nowUTC() time.Time {
	return time.Now().UTC()
}
