package auth

import (
	"trading-stock/internal/handler/auth"

	"github.com/labstack/echo/v4"
)

type AuthRouter struct {
	handler *auth.AuthHandler
}

func NewAuthRouter(handler *auth.AuthHandler) *AuthRouter {
	return &AuthRouter{handler: handler}
}

// Public routes - không cần auth
func (r *AuthRouter) RegisterPublicRoutes(g *echo.Group) {
	auth := g.Group("/auth")
	auth.POST("/register", r.handler.Register)
	auth.POST("/login", r.handler.Login)
	auth.POST("/refresh", r.handler.RefreshToken)
}

// Protected routes - cần auth
func (r *AuthRouter) RegisterRoutes(g *echo.Group) {
	auth := g.Group("/auth")
	auth.POST("/logout", r.handler.Logout) // Logout cần auth để invalidate token
}
