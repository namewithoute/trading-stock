package risk

import (
	"trading-stock/internal/presentation/handler/risk"

	"github.com/labstack/echo/v4"
)

// RiskRouter registers risk-management endpoints.
type RiskRouter struct {
	handler *risk.RiskHandler
}

// NewRiskRouter creates a new RiskRouter.
func NewRiskRouter(handler *risk.RiskHandler) *RiskRouter {
	return &RiskRouter{handler: handler}
}

// RegisterPublicRoutes registers routes that do not require authentication.
func (r *RiskRouter) RegisterPublicRoutes(_ *echo.Group) {}

// RegisterRoutes registers protected risk endpoints.
func (r *RiskRouter) RegisterRoutes(g *echo.Group) {
	riskGroup := g.Group("/risk")

	riskGroup.GET("/metrics/:account_id", r.handler.GetRiskMetrics) // GET /api/v1/risk/metrics/:account_id
	riskGroup.GET("/alerts/:account_id", r.handler.GetActiveAlerts) // GET /api/v1/risk/alerts/:account_id
	riskGroup.POST("/check", r.handler.CheckOrderRisk)              // POST /api/v1/risk/check
}
