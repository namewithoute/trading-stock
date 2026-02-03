package router

import (
	"trading-stock/internal/handler"
	v1 "trading-stock/internal/router/v1"

	"github.com/labstack/echo/v4"
)

// RouterGroup manages all API versions
type RouterGroup struct {
	echo     *echo.Echo
	handlers *handler.HandlerGroup
}

// NewRouterGroup creates main router with handler group
func NewRouterGroup(e *echo.Echo, handlers *handler.HandlerGroup) *RouterGroup {
	return &RouterGroup{
		echo:     e,
		handlers: handlers,
	}
}

// Setup registers all API version routes
func (m *RouterGroup) Setup() {
	// Setup v1 routes
	v1Router := v1.NewV1Router(m.echo, m.handlers)
	v1Router.Setup()

	// Setup v2 routes (future)
	// v2Router := v2.NewV2Router(m.echo, m.handlers)
	// v2Router.Setup()
}
