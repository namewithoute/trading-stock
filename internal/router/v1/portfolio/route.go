package portfolio

import (
	"trading-stock/internal/handler/portfolio"

	"github.com/labstack/echo/v4"
)

type PortfolioRouter struct {
	handler *portfolio.PortfolioHandler
}

func NewPortfolioRouter(handler *portfolio.PortfolioHandler) *PortfolioRouter {
	return &PortfolioRouter{handler: handler}
}

// Protected routes - cần auth
func (r *PortfolioRouter) RegisterRoutes(g *echo.Group) {
	portfolio := g.Group("/portfolio")

	portfolio.GET("", r.handler.GetOverview)                   // GET /api/v1/portfolio
	portfolio.GET("/positions", r.handler.ListPositions)       // GET /api/v1/portfolio/positions
	portfolio.GET("/positions/:symbol", r.handler.GetPosition) // GET /api/v1/portfolio/positions/:symbol
	portfolio.GET("/performance", r.handler.GetPerformance)    // GET /api/v1/portfolio/performance
}

func (r *PortfolioRouter) RegisterPublicRoutes(g *echo.Group) {

}
