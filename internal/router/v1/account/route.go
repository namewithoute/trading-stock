package account

import (
	"trading-stock/internal/handler/account"

	"github.com/labstack/echo/v4"
)

type AccountRouter struct {
	handler *account.AccountHandler
}

func NewAccountRouter(handler *account.AccountHandler) *AccountRouter {
	return &AccountRouter{handler: handler}
}

// Public routes - không cần auth
func (r *AccountRouter) RegisterPublicRoutes(g *echo.Group) {
	// Thường account không có public routes
	// Nhưng có thể có endpoint để check account number có tồn tại không (cho transfer)
	accounts := g.Group("/accounts")
	accounts.GET("/verify/:account_number", r.handler.VerifyAccountExists)
}

// Protected routes - cần auth
func (r *AccountRouter) RegisterRoutes(g *echo.Group) {
	accounts := g.Group("/accounts")

	accounts.GET("", r.handler.ListAccounts)           // GET /api/v1/accounts
	accounts.POST("", r.handler.CreateAccount)         // POST /api/v1/accounts
	accounts.GET("/:id", r.handler.GetAccountDetail)   // GET /api/v1/accounts/:id
	accounts.POST("/:id/deposit", r.handler.Deposit)   // POST /api/v1/accounts/:id/deposit
	accounts.POST("/:id/withdraw", r.handler.Withdraw) // POST /api/v1/accounts/:id/withdraw
}
