package v2

import (
	"trading-stock/internal/presentation/handler"
	"trading-stock/internal/presentation/middleware"
	"trading-stock/internal/presentation/router/v1/account"
	"trading-stock/internal/presentation/router/v1/admin"
	"trading-stock/internal/presentation/router/v1/auth"
	"trading-stock/internal/presentation/router/v1/market"
	"trading-stock/internal/presentation/router/v1/order"
	"trading-stock/internal/presentation/router/v1/portfolio"
	"trading-stock/internal/presentation/router/v1/risk"
	"trading-stock/internal/presentation/router/v1/trade"
	userRouter "trading-stock/internal/presentation/router/v1/user"
	"trading-stock/pkg/jwtservice"

	"github.com/labstack/echo/v4"
)

// Router handles all /api/v2 routes.
type Router struct {
	echo     *echo.Echo
	handlers *handler.HandlerGroup
	jwtSvc   jwtservice.Service
}

// NewV2Router creates a new v2 Router.
func NewV2Router(e *echo.Echo, handlers *handler.HandlerGroup, jwtSvc jwtservice.Service) *Router {
	return &Router{
		echo:     e,
		handlers: handlers,
		jwtSvc:   jwtSvc,
	}
}

// SubRouter is the standard interface every feature router must satisfy.
type SubRouter interface {
	RegisterPublicRoutes(g *echo.Group)
	RegisterRoutes(g *echo.Group)
}

// Setup registers all public and protected routes under /api/v2.
func (r *Router) Setup() {
	v2 := r.echo.Group("/api/v2")

	public := v2.Group("/public")
	protected := v2.Group("/private", middleware.AuthMiddleware(r.jwtSvc))

	subRouters := []SubRouter{
		auth.NewAuthRouter(r.handlers.AuthHandler),
		userRouter.NewUserRouter(r.handlers.UserHandler),
		account.NewAccountRouter(r.handlers.AccountHandler),
		order.NewOrderRouter(r.handlers.OrderHandler),
		portfolio.NewPortfolioRouter(r.handlers.PortfolioHandler),
		market.NewMarketRouter(r.handlers.MarketHandler),
		trade.NewTradeRouter(r.handlers.TradeHandler),
		admin.NewAdminRouter(r.handlers.AdminHandler),
		risk.NewRiskRouter(r.handlers.RiskHandler),
	}

	for _, sr := range subRouters {
		sr.RegisterPublicRoutes(public)
		sr.RegisterRoutes(protected)
	}
}
