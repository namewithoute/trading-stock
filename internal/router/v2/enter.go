package v1

import (
	"trading-stock/internal/handler"
	"trading-stock/internal/middleware"
	"trading-stock/internal/router/v1/account"
	"trading-stock/internal/router/v1/admin"
	"trading-stock/internal/router/v1/auth"
	"trading-stock/internal/router/v1/market"
	"trading-stock/internal/router/v1/order"
	"trading-stock/internal/router/v1/portfolio"
	"trading-stock/internal/router/v1/trade"
	"trading-stock/internal/router/v1/user"

	"github.com/labstack/echo/v4"
)

// Router handles v1 API routes
type Router struct {
	echo     *echo.Echo
	handlers *handler.HandlerGroup
}

// NewV2Router creates a new v2 router with handler group
func NewV2Router(e *echo.Echo, handlers *handler.HandlerGroup) *Router {
	return &Router{
		echo:     e,
		handlers: handlers,
	}
}

type SubRouter interface {
	RegisterPublicRoutes(g *echo.Group)
	RegisterRoutes(g *echo.Group)
}

func (r *Router) Setup() {
	v1 := r.echo.Group("/api/v2")

	public := v1.Group("/public")
	protected := v1.Group("/private", middleware.AuthMiddleware())

	subRouters := []SubRouter{
		auth.NewAuthRouter(r.handlers.AuthHandler),
		user.NewUserRouter(r.handlers.UserHandler),
		account.NewAccountRouter(r.handlers.AccountHandler),
		order.NewOrderRouter(r.handlers.OrderHandler),
		portfolio.NewPortfolioRouter(r.handlers.PortfolioHandler),
		market.NewMarketRouter(r.handlers.MarketHandler),
		trade.NewTradeRouter(r.handlers.TradeHandler),
		admin.NewAdminRouter(r.handlers.AdminHandler),
	}

	for _, sr := range subRouters {
		sr.RegisterPublicRoutes(public)
		sr.RegisterRoutes(protected)
	}
}
