package admin

import (
	"trading-stock/internal/presentation/handler/admin"
	"trading-stock/internal/presentation/middleware"

	"github.com/labstack/echo/v4"
)

type AdminRouter struct {
	handler *admin.AdminHandler
}

func NewAdminRouter(handler *admin.AdminHandler) *AdminRouter {
	return &AdminRouter{handler: handler}
}

// Public routes - không có (admin luôn cần auth)
func (r *AdminRouter) RegisterPublicRoutes(g *echo.Group) {
	// Admin không có public routes
}

// Protected routes - cần auth + admin role
func (r *AdminRouter) RegisterRoutes(g *echo.Group) {
	adminGroup := g.Group("/admin", middleware.RequireRole("admin"))

	adminGroup.GET("/users", r.handler.ListUsers)          // GET /api/v1/admin/users
	adminGroup.PUT("/users/:id/kyc", r.handler.ApproveKYC) // PUT /api/v1/admin/users/:id/kyc
	adminGroup.GET("/orders", r.handler.ListAllOrders)     // GET /api/v1/admin/orders
	adminGroup.GET("/stats", r.handler.GetSystemStats)     // GET /api/v1/admin/stats
}
