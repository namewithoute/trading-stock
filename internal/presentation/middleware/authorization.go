package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// RequireRole is an authorization middleware that allows only users with the
// specified role to proceed. It must be placed AFTER AuthMiddleware.
func RequireRole(role string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userRole, ok := c.Get("role").(string)
			if !ok || userRole == "" {
				return echo.NewHTTPError(http.StatusForbidden, "role not found in context, ensure AuthMiddleware is applied first")
			}

			if userRole != role {
				return echo.NewHTTPError(http.StatusForbidden, "you do not have permission to access this resource")
			}

			return next(c)
		}
	}
}
