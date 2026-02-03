package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// AuthorizationMiddleware checks if user has admin role
func AuthorizationMiddleware(role string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get role from context (set by AuthMiddleware)
			userRole := c.Get("role")
			if userRole == nil {
				return c.JSON(http.StatusForbidden, map[string]string{
					"error": "role not found in context",
				})
			}

			// Check if user has the required role
			if userRole != role {
				return c.JSON(http.StatusForbidden, map[string]string{
					"error": "access required",
				})
			}

			return next(c)
		}
	}
}
