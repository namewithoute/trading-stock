package router

import (
	"trading-stock/internal/presentation/handler"
	v1 "trading-stock/internal/presentation/router/v1"
	"trading-stock/pkg/jwtservice"

	"github.com/labstack/echo/v4"
)

// RegisterRoutes sets up all versioned API routes.
func RegisterRoutes(e *echo.Echo, handlers *handler.HandlerGroup, jwtSvc jwtservice.Service) {
	v1Router := v1.NewV1Router(e, handlers, jwtSvc)
	v1Router.Setup()
}
