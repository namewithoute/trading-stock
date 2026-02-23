package account

// ─────────────────────────────────────────────────────────────────────────────
// Command messages — the application layer's write-side DTOs.
//
// Commands carry the intent and required data for a single state-changing
// operation. They are immutable value types: no methods, no logic.
// ─────────────────────────────────────────────────────────────────────────────

// CreateAccountCommand opens a new trading account for a user.
type CreateAccountCommand struct {
	UserID      string // required
	AccountType string // "CASH" or "MARGIN"; defaults to "CASH" if empty
	Currency    string // ISO-4217 code; defaults to "USD" if empty
}

// DepositCommand adds funds to an existing account.
type DepositCommand struct {
	AccountID string
	Amount    float64
}

// WithdrawCommand removes funds from an account.
type WithdrawCommand struct {
	AccountID string
	Amount    float64
}

// ReserveFundsCommand reduces BuyingPower when a buy order is placed.
type ReserveFundsCommand struct {
	AccountID string
	Amount    float64
}

// ReleaseFundsCommand restores BuyingPower when a buy order is cancelled.
type ReleaseFundsCommand struct {
	AccountID string
	Amount    float64
}

// FreezeAccountCommand suspends an account (no trading allowed).
type FreezeAccountCommand struct {
	AccountID string
}

// UnfreezeAccountCommand re-activates a frozen account.
type UnfreezeAccountCommand struct {
	AccountID string
}

// CloseAccountCommand permanently closes an account.
type CloseAccountCommand struct {
	AccountID string
}
