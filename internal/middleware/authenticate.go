package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// AuthMiddleware validates JWT token and sets user info to context
func AuthMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get Authorization header
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "missing authorization header",
				})
			}

			// Extract token (format: "Bearer <token>")
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "invalid authorization format, expected: Bearer <token>",
				})
			}

			token := parts[1]

			// TODO: Validate JWT token
			// 1. Parse and validate token signature
			// 2. Check token expiration
			// 3. Extract claims (user_id, email, role)
			// 4. Check if token is blacklisted (Redis)

			// For now, mock validation
			if token == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "invalid token",
				})
			}

			// Set user info to context (mock data for now)
			// TODO: Replace with actual claims from JWT
			c.Set("user_id", "usr_123")
			c.Set("email", "user@example.com")
			c.Set("role", "user")

			return next(c)
		}
	}
}
