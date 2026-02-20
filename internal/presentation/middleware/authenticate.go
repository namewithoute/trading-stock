package middleware

import (
	"net/http"
	"strings"

	"trading-stock/pkg/jwtservice"

	"github.com/labstack/echo/v4"
)

// AuthMiddleware validates the JWT Bearer token from the Authorization header.
// On success, it injects user_id, email, and role into the Echo context.
// Protected routes MUST be wrapped with this middleware.
func AuthMiddleware(jwtSvc jwtservice.Service) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header")
			}

			// Expect format: "Bearer <token>"
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" || parts[1] == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid authorization format, expected: Bearer <token>")
			}

			// Validate JWT via the shared jwtservice utility
			claims, err := jwtSvc.ValidateToken(c.Request().Context(), parts[1])
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired token")
			}

			// Inject verified claims into context for downstream handlers
			c.Set("user_id", claims.UserID)
			c.Set("email", claims.Email)
			c.Set("role", claims.Role)

			return next(c)
		}
	}
}
