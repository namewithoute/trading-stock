package user

import (
	"net/http"
	userUC "trading-stock/internal/application/user"
	"trading-stock/pkg/response"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// UserHandler handles user management endpoints
type UserHandler struct {
	userUseCase userUC.UseCase
	logger      *zap.Logger
}

// NewUserHandler creates a new user handler
func NewUserHandler(userUseCase userUC.UseCase, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		userUseCase: userUseCase,
		logger:      logger,
	}
}

// GetPublicProfile gets public profile of a user (public endpoint)
// GET /api/v1/users/:id/public
func (h *UserHandler) GetPublicProfile(c echo.Context) error {
	userID := c.Param("id")
	if userID == "" {
		return response.Error(c, http.StatusBadRequest, "User ID is required", "user_id_empty")
	}

	u, err := h.userUseCase.GetProfile(c.Request().Context(), userID)
	if err != nil {
		h.logger.Error("GetPublicProfile failed", zap.Error(err), zap.String("user_id", userID))
		return response.Error(c, http.StatusNotFound, "User not found", err.Error())
	}

	return response.Success(c, http.StatusOK, "Public profile retrieved", PublicProfileResponse{
		UserID:    u.ID,
		Username:  u.Username,
		FullName:  u.FullName(),
		CreatedAt: u.CreatedAt,
	})
}

// GetProfile gets current user's profile (protected)
// GET /api/v1/users/me
func (h *UserHandler) GetProfile(c echo.Context) error {
	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		return response.Error(c, http.StatusUnauthorized, "User not authenticated", "unauthorized")
	}

	u, err := h.userUseCase.GetProfile(c.Request().Context(), userID)
	if err != nil {
		h.logger.Error("GetProfile failed", zap.Error(err), zap.String("user_id", userID))
		return response.Error(c, http.StatusInternalServerError, "Failed to get profile", err.Error())
	}

	return response.Success(c, http.StatusOK, "Profile retrieved", UserProfileResponse{
		UserID:        u.ID,
		Email:         u.Email,
		Username:      u.Username,
		FirstName:     u.FirstName,
		LastName:      u.LastName,
		Phone:         u.Phone,
		EmailVerified: u.EmailVerified,
		KYCStatus:     string(u.KYCStatus),
		Status:        string(u.Status),
		Role:          string(u.Role),
		CreatedAt:     u.CreatedAt,
		LastLogin:     u.LastLogin,
	})
}

// UpdateProfile updates current user's profile (protected)
// PUT /api/v1/users/me
func (h *UserHandler) UpdateProfile(c echo.Context) error {
	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		return response.Error(c, http.StatusUnauthorized, "User not authenticated", "unauthorized")
	}

	var req UpdateProfileRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "Invalid request payload", err.Error())
	}

	data := map[string]interface{}{
		"first_name": req.FirstName,
		"last_name":  req.LastName,
		"phone":      req.Phone,
	}

	if err := h.userUseCase.UpdateProfile(c.Request().Context(), userID, data); err != nil {
		h.logger.Error("UpdateProfile failed", zap.Error(err), zap.String("user_id", userID))
		return response.Error(c, http.StatusInternalServerError, "Failed to update profile", err.Error())
	}

	u, _ := h.userUseCase.GetProfile(c.Request().Context(), userID)
	if u == nil {
		return response.Success(c, http.StatusOK, "Profile updated successfully", nil)
	}

	return response.Success(c, http.StatusOK, "Profile updated successfully", UserProfileResponse{
		UserID:        u.ID,
		Email:         u.Email,
		Username:      u.Username,
		FirstName:     u.FirstName,
		LastName:      u.LastName,
		Phone:         u.Phone,
		EmailVerified: u.EmailVerified,
		KYCStatus:     string(u.KYCStatus),
		Status:        string(u.Status),
		Role:          string(u.Role),
		CreatedAt:     u.CreatedAt,
		LastLogin:     u.LastLogin,
	})
}

// VerifyEmail verifies user's email (protected)
// POST /api/v1/users/me/verify-email
func (h *UserHandler) VerifyEmail(c echo.Context) error {
	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		return response.Error(c, http.StatusUnauthorized, "User not authenticated", "unauthorized")
	}

	if err := h.userUseCase.VerifyEmail(c.Request().Context(), userID); err != nil {
		h.logger.Error("VerifyEmail failed", zap.Error(err), zap.String("user_id", userID))
		return response.Error(c, http.StatusInternalServerError, "Failed to verify email", err.Error())
	}

	return response.Success(c, http.StatusOK, "Email verified successfully", map[string]string{
		"user_id": userID,
		"status":  "verified",
	})
}

// SubmitKYC submits KYC documents (protected)
// POST /api/v1/users/me/kyc
func (h *UserHandler) SubmitKYC(c echo.Context) error {
	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		return response.Error(c, http.StatusUnauthorized, "User not authenticated", "unauthorized")
	}

	if err := h.userUseCase.SubmitKYC(c.Request().Context(), userID); err != nil {
		h.logger.Error("SubmitKYC failed", zap.Error(err), zap.String("user_id", userID))
		return response.Error(c, http.StatusInternalServerError, "Failed to submit KYC", err.Error())
	}

	return response.Success(c, http.StatusOK, "KYC submitted successfully", map[string]string{
		"user_id": userID,
		"status":  "pending",
	})
}
