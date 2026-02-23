package account

// ─────────────────────────────────────────────────────────────────────────────
// Query messages — the application layer's read-side DTOs.
//
// Queries carry the criteria for a single read operation.
// They are immutable value types: no methods, no logic.
// ─────────────────────────────────────────────────────────────────────────────

// GetAccountQuery retrieves a single account by its ID.
type GetAccountQuery struct {
	AccountID string
}

// GetAccountByUserQuery retrieves an account and asserts it belongs to the user.
type GetAccountByUserQuery struct {
	AccountID string
	UserID    string
}

// ListAccountsQuery retrieves all accounts belonging to a user.
type ListAccountsQuery struct {
	UserID string
}

// GetPrimaryAccountByUserQuery retrieves the first active account for a user.
// Used by the Order domain when no explicit AccountID is provided.
type GetPrimaryAccountByUserQuery struct {
	UserID string
}
