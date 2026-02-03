package app

import (
	"net/http"
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

	// Health check endpoint
	e.GET("/health", a.healthCheckHandler)

	// API v1 group
	v1 := e.Group("/api/v1")
	_ = v1 // Will be used for registering routes

	a.Echo = e
}

// healthCheckHandler returns the health status of the application
func (a *App) healthCheckHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"service":   "trading-stock-api",
	})
}
