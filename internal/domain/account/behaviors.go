package account

import (
	"github.com/cockroachdb/apd/v3"
	pkgdecimal "trading-stock/pkg/decimal"
)

// ─────────────────────────────────────────────────────────────────────────────
// Behaviors — Domain operations that validate invariants then emit events.
//
// Note on naming: this file is intentionally called "behaviors" (not "commands")
// to distinguish domain command-methods from the CQRS command messages defined
// in the application layer (internal/application/account/command/).
//
// DDD pattern enforced:
//   - Each method checks ALL guard conditions BEFORE calling apply().
//   - If any guard fails, NO event is emitted and the aggregate is unchanged.
//   - Business rules live here — never in handlers or repositories.
// ─────────────────────────────────────────────────────────────────────────────

// ─── Write-side constructor ───────────────────────────────────────────────────

// OpenAccount is the aggregate constructor.
// Validates inputs and emits AccountCreatedEvent to bootstrap a new account.
func OpenAccount(id, userID string, accountType AccountType, currency string) (*AccountAggregate, error) {
	if userID == "" {
		return nil, ErrInvalidUserID
	}
	if !accountType.IsValid() {
		return nil, ErrInvalidAccountType
	}
	if currency == "" {
		return nil, ErrInvalidCurrency
	}

	agg := &AccountAggregate{}
	agg.apply(AccountCreatedEvent{
		AggregateID: id,
		UserID:      userID,
		AccountType: accountType,
		Currency:    currency,
		OccurredAt:  nowUTC(),
	}, true)
	return agg, nil
}

// ─── Money operations ─────────────────────────────────────────────────────────

// Deposit adds funds to the account. Emits MoneyDepositedEvent.
func (a *AccountAggregate) Deposit(amount apd.Decimal) error {
	if a.Status != StatusActive {
		return ErrAccountNotActive
	}
	if amount.Sign() <= 0 {
		return ErrInvalidAmount
	}

	a.apply(MoneyDepositedEvent{
		AggregateID: a.ID,
		Amount:      pkgdecimal.From(amount),
		Currency:    a.Money.Currency,
		OccurredAt:  nowUTC(),
	}, true)
	return nil
}

// Withdraw removes funds from the account. Emits MoneyWithdrawnEvent.
func (a *AccountAggregate) Withdraw(amount apd.Decimal) error {
	if a.Status != StatusActive {
		return ErrAccountNotActive
	}
	if amount.Sign() <= 0 {
		return ErrInvalidAmount
	}
	if a.Money.Balance.Cmp(&amount) < 0 {
		return ErrInsufficientBalance
	}

	a.apply(MoneyWithdrawnEvent{
		AggregateID: a.ID,
		Amount:      pkgdecimal.From(amount),
		Currency:    a.Money.Currency,
		OccurredAt:  nowUTC(),
	}, true)
	return nil
}

// ─── Order-related fund management ───────────────────────────────────────────

// ReserveFunds reduces BuyingPower when a buy order is placed. Emits FundsReservedEvent.
func (a *AccountAggregate) ReserveFunds(amount apd.Decimal) error {
	if a.Status != StatusActive {
		return ErrAccountNotActive
	}
	if amount.Sign() <= 0 {
		return ErrInvalidAmount
	}
	if a.Money.BuyingPower.Cmp(&amount) < 0 {
		return ErrInsufficientBuyingPower
	}

	a.apply(FundsReservedEvent{
		AggregateID: a.ID,
		Amount:      pkgdecimal.From(amount),
		OccurredAt:  nowUTC(),
	}, true)
	return nil
}

// ReleaseFunds returns previously reserved BuyingPower (order cancelled / rejected).
// Emits FundsReleasedEvent.
func (a *AccountAggregate) ReleaseFunds(amount apd.Decimal) error {
	if amount.Sign() <= 0 {
		return ErrInvalidAmount
	}

	a.apply(FundsReleasedEvent{
		AggregateID: a.ID,
		Amount:      pkgdecimal.From(amount),
		OccurredAt:  nowUTC(),
	}, true)
	return nil
}

// SettleTrade finalises cash movement for a filled order execution.
//
// BUY settlement: debits Balance by the settlement amount (payment for securities).
// The BuyingPower was already reduced when funds were reserved at order placement,
// so only the Balance changes here.
//
// SELL settlement: credits both Balance and BuyingPower (cash received from the sale).
//
// Emits TradeSettledEvent.
func (a *AccountAggregate) SettleTrade(tradeID, side string, amount apd.Decimal) error {
	if a.Status != StatusActive {
		return ErrAccountNotActive
	}
	if amount.Sign() <= 0 {
		return ErrInvalidAmount
	}
	if side != "BUY" && side != "SELL" {
		return ErrInvalidAmount // reuse; a dedicated error would be cleaner
	}

	a.apply(TradeSettledEvent{
		AggregateID: a.ID,
		TradeID:     tradeID,
		Side:        side,
		Amount:      pkgdecimal.From(amount),
		OccurredAt:  nowUTC(),
	}, true)
	return nil
}

// ─── Status transitions ───────────────────────────────────────────────────────

// Freeze suspends the account (no trading, read-only). Active → Frozen.
func (a *AccountAggregate) Freeze() error {
	if a.Status != StatusActive {
		return ErrAccountNotActive
	}
	a.apply(StatusChangedEvent{
		AggregateID: a.ID,
		OldStatus:   a.Status,
		NewStatus:   StatusFrozen,
		OccurredAt:  nowUTC(),
	}, true)
	return nil
}

// Unfreeze re-activates a frozen account. Frozen → Active.
func (a *AccountAggregate) Unfreeze() error {
	if a.Status != StatusFrozen {
		return ErrAccountNotFrozen
	}
	a.apply(StatusChangedEvent{
		AggregateID: a.ID,
		OldStatus:   a.Status,
		NewStatus:   StatusActive,
		OccurredAt:  nowUTC(),
	}, true)
	return nil
}

// Close permanently shuts the account. Cannot be re-opened.
func (a *AccountAggregate) Close() error {
	if a.Status == StatusClosed {
		return ErrAccountClosed
	}
	a.apply(StatusChangedEvent{
		AggregateID: a.ID,
		OldStatus:   a.Status,
		NewStatus:   StatusClosed,
		OccurredAt:  nowUTC(),
	}, true)
	return nil
}
