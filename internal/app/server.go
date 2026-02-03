package app

import (
	"time"

	"trading-stock/pkg/logger"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// initHTTPServer initializes the Echo HTTP server with middleware
func (a *App) initHTTPServer() {
	e := echo.New()

	// Hide Echo banner and port message
	e.HideBanner = true
	e.HidePort = true

	// Configure middleware
	e.Use(middleware.RequestID())
	e.Use(logger.ZapLogger(a.Logger)) // Use Zap logger from logger package
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Add timeout middleware
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: 30 * time.Second,
	}))

	a.Echo = e
}
