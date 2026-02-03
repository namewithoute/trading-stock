package handler

import (
	"trading-stock/internal/handler/account"
	"trading-stock/internal/handler/admin"
	"trading-stock/internal/handler/auth"
	"trading-stock/internal/handler/market"
	"trading-stock/internal/handler/order"
	"trading-stock/internal/handler/portfolio"
	"trading-stock/internal/handler/trade"
	"trading-stock/internal/handler/user"
	"trading-stock/internal/usecase"
)

// HandlerGroup groups all handlers together for easy dependency injection
type HandlerGroup struct {
	AuthHandler      *auth.AuthHandler
	UserHandler      *user.UserHandler
	AccountHandler   *account.AccountHandler
	OrderHandler     *order.OrderHandler
	PortfolioHandler *portfolio.PortfolioHandler
	MarketHandler    *market.MarketHandler
	TradeHandler     *trade.TradeHandler
	AdminHandler     *admin.AdminHandler
}

// NewHandlerGroup creates a new handler group with all handlers initialized
// Pass use cases to inject dependencies
func NewHandlerGroup(services *usecase.Usecases) *HandlerGroup {

	return &HandlerGroup{
		AuthHandler:      auth.NewAuthHandler(services.Auth),
		UserHandler:      user.NewUserHandler(services.User),
		AccountHandler:   account.NewAccountHandler(services.Account),
		OrderHandler:     order.NewOrderHandler(services.Order),
		PortfolioHandler: portfolio.NewPortfolioHandler(services.Portfolio),
		MarketHandler:    market.NewMarketHandler(services.Market),
		TradeHandler:     trade.NewTradeHandler(services.Trade),
		AdminHandler:     admin.NewAdminHandler(services.Admin),
	}
}
