package admin

import (
	"trading-stock/internal/handler/admin"
	"trading-stock/internal/middleware"

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
	// g đã là /api/v1/admin rồi
	g.Group("/admin", middleware.AuthorizationMiddleware("admin"))
	g.GET("/users", r.handler.ListUsers)          // GET /api/v1/admin/users
	g.PUT("/users/:id/kyc", r.handler.ApproveKYC) // PUT /api/v1/admin/users/:id/kyc
	g.GET("/orders", r.handler.ListAllOrders)     // GET /api/v1/admin/orders
	g.GET("/stats", r.handler.GetSystemStats)     // GET /api/v1/admin/stats
}
