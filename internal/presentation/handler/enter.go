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

	"go.uber.org/zap"
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
func NewHandlerGroup(services *application.Usecases, logger *zap.Logger) *HandlerGroup {
	return &HandlerGroup{
		AuthHandler:      auth.NewAuthHandler(services.Auth, logger),
		UserHandler:      user.NewUserHandler(services.User, logger),
		AccountHandler:   account.NewAccountHandler(services.Account, logger),
		OrderHandler:     order.NewOrderHandler(services.Order, logger),
		PortfolioHandler: portfolio.NewPortfolioHandler(services.Portfolio, logger),
		MarketHandler:    market.NewMarketHandler(services.Market, logger),
		TradeHandler:     trade.NewTradeHandler(services.Trade, logger),
		AdminHandler:     admin.NewAdminHandler(services.Admin, logger),
	}
}
