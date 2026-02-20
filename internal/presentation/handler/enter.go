package handler

import (
	"trading-stock/internal/application"
	"trading-stock/internal/presentation/handler/account"
	"trading-stock/internal/presentation/handler/admin"
	"trading-stock/internal/presentation/handler/auth"
	"trading-stock/internal/presentation/handler/market"
	"trading-stock/internal/presentation/handler/order"
	"trading-stock/internal/presentation/handler/portfolio"
	"trading-stock/internal/presentation/handler/trade"
	"trading-stock/internal/presentation/handler/user"
)

// HandlerGroup groups all HTTP handlers for easy dependency injection.
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

// NewHandlerGroup initialises all handlers with their respective use cases.
func NewHandlerGroup(services *application.Usecases) *HandlerGroup {
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
