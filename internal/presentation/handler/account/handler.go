package account

import (
	"net/http"
	accountUC "trading-stock/internal/application/account"
	"trading-stock/internal/domain/account"
	"trading-stock/pkg/response"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// AccountHandler handles trading account endpoints
type AccountHandler struct {
	accountUseCase accountUC.UseCase // Uncomment when service is ready
	logger         *zap.Logger
}

// NewAccountHandler creates a new account handler
func NewAccountHandler(accountUseCase accountUC.UseCase, logger *zap.Logger) *AccountHandler {
	return &AccountHandler{
		accountUseCase: accountUseCase,
		logger:         logger,
	}
}

// VerifyAccountExists verifies if account number exists (public)
// GET /api/v1/accounts/verify/:account_number
func (h *AccountHandler) VerifyAccountExists(c echo.Context) error {
	accountNumber := c.Param("account_number")

	if accountNumber == "" {
		return response.Error(c, http.StatusBadRequest, "Account number is required", "account_number_empty")
	}

	acc, err := h.accountUseCase.GetAccount(c.Request().Context(), accountNumber)
	if err != nil {
		h.logger.Error("Failed to get account", zap.Error(err))
		return response.Error(c, http.StatusInternalServerError, "Failed to get account", err.Error())
	}

	if acc == nil {
		h.logger.Error("Account not found", zap.String("account_number", accountNumber))
		return response.Error(c, http.StatusNotFound, "Account not found", account.ErrAccountNotFound.Error())
	}

	if acc.Status != account.StatusActive {
		h.logger.Error("Account is not active", zap.String("account_number", accountNumber))
		return response.Error(c, http.StatusForbidden, "Account is not active", account.ErrAccountNotActive.Error())
	}

	accountResponse := ToAccountResponse(acc)

	return response.Success(c, http.StatusOK, "Account verified successfully", accountResponse)
}

// ListAccounts lists all trading accounts of current user (protected)
// GET /api/v1/accounts
func (h *AccountHandler) ListAccounts(c echo.Context) error {
	userID := c.Get("user_id")

	if userID == nil || userID == "" {
		return response.Error(c, http.StatusUnauthorized, "User ID not found in context", "unauthorized")
	}

	accounts, err := h.accountUseCase.ListAccounts(c.Request().Context(), userID.(string))
	if err != nil {
		h.logger.Error("Failed to list accounts", zap.Error(err))
		return response.Error(c, http.StatusInternalServerError, "Failed to list accounts", err.Error())
	}

	// Prepare data for response
	var accountsResponse []AccountResponse
	for _, acc := range accounts {
		accountsResponse = append(accountsResponse, *ToAccountResponse(acc))
	}

	return response.Success(c, http.StatusOK, "Accounts retrieved successfully", AccountListingResponse{
		UserID:   userID.(string),
		Accounts: accountsResponse,
		Total:    len(accountsResponse),
	})
}

// CreateAccount creates a new trading account (protected)
// POST /api/v1/accounts
func (h *AccountHandler) CreateAccount(c echo.Context) error {
	userID := c.Get("user_id")

	if userID == nil || userID == "" {
		return response.Error(c, http.StatusUnauthorized, "User ID not found in context", "unauthorized")
	}

	acc, err := h.accountUseCase.CreateAccount(c.Request().Context(), userID.(string))
	if err != nil {
		h.logger.Error("Failed to create account", zap.Error(err), zap.String("userID", userID.(string)))
		return response.Error(c, http.StatusInternalServerError, "Failed to create account", err.Error())
	}

	return response.Success(c, http.StatusCreated, "Account created successfully", ToAccountResponse(acc))
}

// GetAccountDetail gets account details (protected)
// GET /api/v1/accounts/:id
func (h *AccountHandler) GetAccountDetail(c echo.Context) error {
	accountID := c.Param("id")
	userID := c.Get("user_id")

	if accountID == "" {
		return response.Error(c, http.StatusBadRequest, "Account ID is required", "account_id_empty")
	}

	if userID == nil || userID == "" {
		return response.Error(c, http.StatusUnauthorized, "User ID not found in context", "unauthorized")
	}

	acc, err := h.accountUseCase.GetAccount(c.Request().Context(), accountID)
	if err != nil {
		h.logger.Error("Failed to get account details", zap.Error(err), zap.String("accountID", accountID))
		return response.Error(c, http.StatusInternalServerError, "Failed to get account details", err.Error())
	}

	if acc == nil {
		return response.Error(c, http.StatusNotFound, "Account not found", account.ErrAccountNotFound.Error())
	}

	// Basic authorization check: verify the account actually belongs to this user
	if acc.UserID != userID.(string) {
		h.logger.Warn("Unauthorized account access attempt", zap.String("userID", userID.(string)), zap.String("accountUserID", acc.UserID))
		return response.Error(c, http.StatusForbidden, "You don't have permission to access this account", "forbidden")
	}

	return response.Success(c, http.StatusOK, "Account details retrieved", ToAccountResponse(acc))
}

// Deposit deposits money to account (protected)
// POST /api/v1/accounts/:id/deposit
func (h *AccountHandler) Deposit(c echo.Context) error {
	accountID := c.Param("id")
	userID := c.Get("user_id")

	if accountID == "" {
		return response.Error(c, http.StatusBadRequest, "Account ID is required", "account_id_empty")
	}

	if userID == nil || userID == "" {
		return response.Error(c, http.StatusUnauthorized, "User ID not found in context", "unauthorized")
	}

	var req DepositRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "Invalid request payload", err.Error())
	}

	if req.Amount <= 0 {
		return response.Error(c, http.StatusBadRequest, "Invalid amount", "Amount must be greater than zero")
	}

	acc, err := h.accountUseCase.Deposit(c.Request().Context(), accountID, userID.(string), req.Amount)
	if err != nil {
		h.logger.Error("Failed to deposit to account", zap.Error(err), zap.String("accountID", accountID))

		// Map domain error to 400 Bad Request
		if err == account.ErrInvalidAmount || err == account.ErrAccountNotFound {
			return response.Error(c, http.StatusBadRequest, "Failed to deposit", err.Error())
		}

		return response.Error(c, http.StatusInternalServerError, "Failed to process deposit", err.Error())
	}

	return response.Success(c, http.StatusOK, "Deposit successful", ToAccountResponse(acc))
}

// Withdraw withdraws money from account (protected)
// POST /api/v1/accounts/:id/withdraw
func (h *AccountHandler) Withdraw(c echo.Context) error {
	accountID := c.Param("id")
	userID := c.Get("user_id")

	if accountID == "" {
		return response.Error(c, http.StatusBadRequest, "Account ID is required", "account_id_empty")
	}

	if userID == nil || userID == "" {
		return response.Error(c, http.StatusUnauthorized, "User ID not found in context", "unauthorized")
	}

	var req WithdrawRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "Invalid request payload", err.Error())
	}

	if req.Amount <= 0 {
		return response.Error(c, http.StatusBadRequest, "Invalid amount", "Amount must be greater than zero")
	}

	acc, err := h.accountUseCase.Withdraw(c.Request().Context(), accountID, userID.(string), req.Amount)
	if err != nil {
		h.logger.Error("Failed to withdraw from account", zap.Error(err), zap.String("accountID", accountID))

		// Map domain logic errors to 400 Bad Request
		if err == account.ErrInsufficientBalance || err == account.ErrInvalidAmount || err == account.ErrAccountNotFound {
			return response.Error(c, http.StatusBadRequest, "Failed to withdraw", err.Error())
		}

		return response.Error(c, http.StatusInternalServerError, "Failed to process withdrawal", err.Error())
	}

	return response.Success(c, http.StatusOK, "Withdrawal successful", ToAccountResponse(acc))
}
