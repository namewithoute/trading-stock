package account

import (
	"trading-stock/internal/domain/account"
	pkgdecimal "trading-stock/pkg/decimal"
)

// AccountListingResponse is the paginated response payload for list endpoints.
type AccountListingResponse struct {
	UserID   string            `json:"user_id"`
	Accounts []AccountResponse `json:"accounts"`
	Total    int               `json:"total"`
}

type CreateAccountRequest struct {
	AccountType string `json:"account_type" validate:"required,oneof=cash margin"`
	Currency    string `json:"currency" validate:"required,iso4217"`
}

// AccountResponse is the serialisable DTO sent to clients.
// It is derived from the AccountReadModel (query side of CQRS).
type AccountResponse struct {
	ID          string              `json:"id"`
	AccountType account.AccountType `json:"account_type"`
	Money       Money               `json:"money"`
	Status      account.Status      `json:"status"`
}

// Money is the nested balance structure in AccountResponse.
type Money struct {
	Balance     pkgdecimal.Decimal `json:"balance"`
	BuyingPower pkgdecimal.Decimal `json:"buying_power"`
	Currency    string             `json:"currency"`
}

// Request DTOs

type DepositRequest struct {
	Amount pkgdecimal.Decimal `json:"amount" validate:"required"`
}

type WithdrawRequest struct {
	Amount pkgdecimal.Decimal `json:"amount" validate:"required"`
}

// ToAccountResponse maps a read model (CQRS query side) to the HTTP response DTO.
func ToAccountResponse(rm *account.AccountReadModel) *AccountResponse {
	return &AccountResponse{
		ID:          rm.ID,
		AccountType: rm.AccountType,
		Money: Money{
			Balance:     pkgdecimal.From(rm.Balance),
			BuyingPower: pkgdecimal.From(rm.BuyingPower),
			Currency:    rm.Currency,
		},
		Status: rm.Status,
	}
}
