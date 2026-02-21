package account

import (
	"trading-stock/internal/domain/account"
)

// AccountListingResponse is the paginated response payload for list endpoints.
type AccountListingResponse struct {
	UserID   string            `json:"user_id"`
	Accounts []AccountResponse `json:"accounts"`
	Total    int               `json:"total"`
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
	Balance     float64 `json:"balance"`
	BuyingPower float64 `json:"buying_power"`
	Currency    string  `json:"currency"`
}

// Request DTOs

type DepositRequest struct {
	Amount float64 `json:"amount" validate:"required,gt=0"`
}

type WithdrawRequest struct {
	Amount float64 `json:"amount" validate:"required,gt=0"`
}

// ToAccountResponse maps a read model (CQRS query side) to the HTTP response DTO.
func ToAccountResponse(rm *account.AccountReadModel) *AccountResponse {
	return &AccountResponse{
		ID:          rm.ID,
		AccountType: rm.AccountType,
		Money: Money{
			Balance:     rm.Balance,
			BuyingPower: rm.BuyingPower,
			Currency:    rm.Currency,
		},
		Status: rm.Status,
	}
}
