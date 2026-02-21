package auth

import (
	"net/http"

	authUC "trading-stock/internal/application/auth"
	"trading-stock/pkg/response"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// AuthHandler handles HTTP requests for all auth-related endpoints.
type AuthHandler struct {
	authUseCase authUC.UseCase
	logger      *zap.Logger
}

// NewAuthHandler creates a new AuthHandler with its UseCase dependency.
func NewAuthHandler(authUseCase authUC.UseCase, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		authUseCase: authUseCase,
		logger:      logger,
	}
}

// Register handles user registration.
// POST /api/v1/auth/register
func (h *AuthHandler) Register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, response.MsgInvalidPayload)
	}

	user, token, err := h.authUseCase.Register(c.Request().Context(), req.Email, req.Password, req.Name)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// Use a dedicated DTO instead of exposing the raw domain.User
	return response.Success(c, http.StatusCreated, response.MsgRegisterSuccess, RegisterResponse{
		UserID: user.ID,
		Email:  user.Email,
		Name:   user.FullName(),
		Token:  token,
	})
}

// Login authenticates a user and returns an access + refresh token pair.
// POST /api/v1/auth/login
func (h *AuthHandler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, response.MsgInvalidPayload)
	}

	user, accessToken, refreshToken, err := h.authUseCase.Login(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		// Return 401 Unauthorized for any auth failure to avoid user enumeration
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	return response.Success(c, http.StatusOK, response.MsgLoginSuccess, LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UserID:       user.ID,
	})
}

// RefreshToken generates a new access token using a valid refresh token.
// POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshToken(c echo.Context) error {
	var req RefreshTokenRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, response.MsgInvalidPayload)
	}

	newAccessToken, err := h.authUseCase.RefreshToken(c.Request().Context(), req.RefreshToken)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, response.MsgInvalidRefresh)
	}

	return response.Success(c, http.StatusOK, response.MsgRefreshSuccess, map[string]string{
		"access_token": newAccessToken,
	})
}

// Logout invalidates the user's refresh token.
// POST /api/v1/auth/logout (protected by AuthMiddleware)
func (h *AuthHandler) Logout(c echo.Context) error {
	var req LogoutRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, response.MsgInvalidPayload)
	}

	// user_id is injected by AuthMiddleware after verifying the JWT
	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, response.MsgUnauthorizedToken)
	}

	if err := h.authUseCase.Logout(c.Request().Context(), userID, req.RefreshToken); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, response.MsgLogoutFailed)
	}

	return response.Success(c, http.StatusOK, response.MsgLogoutSuccess, nil)
}
