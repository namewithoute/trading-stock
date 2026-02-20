package app

import (
	"errors"
	"net/http"

	"trading-stock/pkg/response"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// GlobalErrorHandler centralizes all error responses.
// It ensures every error – regardless of where it originated –
// is returned to the client in a consistent JSON format via pkg/response.
func (a *App) GlobalErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	message := "Internal Server Error"

	// Type assertion: check if this is an Echo HTTP error (e.g. 404, 401)
	var he *echo.HTTPError
	if errors.As(err, &he) {
		code = he.Code
		if m, ok := he.Message.(string); ok {
			message = m
		}
	}

	// Log the real underlying error for observability (NOT exposed to client)
	a.Logger.Error("API Exception",
		zap.Int("status", code),
		zap.Error(err),
		zap.String("path", c.Request().URL.Path),
		zap.String("method", c.Request().Method),
	)

	// Write JSON response only if not already committed to the client
	if !c.Response().Committed {
		_ = c.JSON(code, response.Response{
			Success: false,
			Code:    code,
			Message: message,
			Error:   err.Error(),
		})
	}
}
