package router

import (
	"trading-stock/internal/handler"
	v1 "trading-stock/internal/router/v1"

	"github.com/labstack/echo/v4"
)

func RegisterRoutes(e *echo.Echo, handlers *handler.HandlerGroup) {
	v1Router := v1.NewV1Router(e, handlers)
	v1Router.Setup()
}
