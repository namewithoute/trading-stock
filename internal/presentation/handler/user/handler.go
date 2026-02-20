package user

import (
	"net/http"
	userUC "trading-stock/internal/application/user"

	"github.com/labstack/echo/v4"
)

// UserHandler handles user management endpoints
type UserHandler struct {
	UserUseCase userUC.UseCase // Uncomment when service is ready
}

// NewUserHandler creates a new user handler
func NewUserHandler(UserUseCase userUC.UseCase) *UserHandler {
	return &UserHandler{
		UserUseCase: UserUseCase,
	}
}

// GetPublicProfile gets public profile of a user (public endpoint)
// GET /api/v1/users/:id/public
func (h *UserHandler) GetPublicProfile(c echo.Context) error {
	userID := c.Param("id")

	// TODO: Implement get public profile logic
	// 1. Get user ID from URL param
	// 2. Fetch public user info from database (exclude sensitive data)
	// 3. Return public profile

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Get public profile - TODO: implement",
		"data": map[string]interface{}{
			"user_id":     userID,
			"username":    "john_doe",
			"avatar":      "https://example.com/avatar.jpg",
			"joined_date": "2024-01-01",
		},
	})
}

// GetProfile gets current user's profile (protected)
// GET /api/v1/users/me
func (h *UserHandler) GetProfile(c echo.Context) error {
	userID := c.Get("user_id")

	// TODO: Implement get profile logic
	// 1. Get user ID from context (set by auth middleware)
	// 2. Fetch user info from database
	// 3. Return full profile

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Get profile - TODO: implement",
		"data": map[string]interface{}{
			"user_id":    userID,
			"email":      "user@example.com",
			"name":       "John Doe",
			"phone":      "+84123456789",
			"kyc_status": "pending",
		},
	})
}

// UpdateProfile updates current user's profile (protected)
// PUT /api/v1/users/me
func (h *UserHandler) UpdateProfile(c echo.Context) error {
	userID := c.Get("user_id")

	// TODO: Implement update profile logic
	// 1. Get user ID from context
	// 2. Parse request body (name, phone, avatar, etc.)
	// 3. Validate input
	// 4. Update user in database
	// 5. Return updated profile

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Update profile - TODO: implement",
		"user_id": userID,
	})
}

// VerifyEmail verifies user's email (protected)
// POST /api/v1/users/me/verify-email
func (h *UserHandler) VerifyEmail(c echo.Context) error {
	userID := c.Get("user_id")

	// TODO: Implement email verification logic
	// 1. Get user ID from context
	// 2. Parse verification code from request
	// 3. Validate code
	// 4. Update email_verified status
	// 5. Return success

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Email verified successfully",
		"user_id": userID,
	})
}

// SubmitKYC submits KYC documents (protected)
// POST /api/v1/users/me/kyc
func (h *UserHandler) SubmitKYC(c echo.Context) error {
	userID := c.Get("user_id")

	// TODO: Implement KYC submission logic
	// 1. Get user ID from context
	// 2. Parse KYC documents (ID card, selfie, address proof)
	// 3. Upload files to storage (S3/MinIO)
	// 4. Create KYC record in database
	// 5. Set status to "pending"
	// 6. Return success

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "KYC submitted successfully",
		"user_id": userID,
		"status":  "pending",
	})
}
