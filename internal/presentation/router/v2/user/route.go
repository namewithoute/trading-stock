package user

import (
	"trading-stock/internal/presentation/handler/user"

	"github.com/labstack/echo/v4"
)

type UserRouter struct {
	handler *user.UserHandler
}

func NewUserRouter(handler *user.UserHandler) *UserRouter {
	return &UserRouter{handler: handler}
}

// Public routes - không cần auth
func (r *UserRouter) RegisterPublicRoutes(g *echo.Group) {
	users := g.Group("/users")

	// Ví dụ: Xem profile công khai của user khác
	users.GET("/:id/public", r.handler.GetPublicProfile) // GET /api/v1/users/:id/public
}

// Protected routes - cần auth
func (r *UserRouter) RegisterRoutes(g *echo.Group) {
	users := g.Group("/users")

	users.GET("/me", r.handler.GetProfile)                // GET /api/v1/users/me
	users.PUT("/me", r.handler.UpdateProfile)             // PUT /api/v1/users/me
	users.POST("/me/verify-email", r.handler.VerifyEmail) // POST /api/v1/users/me/verify-email
	users.POST("/me/kyc", r.handler.SubmitKYC)            // POST /api/v1/users/me/kyc
}
