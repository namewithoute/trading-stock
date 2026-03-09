package user

import "time"

// PublicProfileResponse is the reduced profile returned on public endpoints.
type PublicProfileResponse struct {
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	FullName  string    `json:"full_name"`
	CreatedAt time.Time `json:"created_at"`
}

// UserProfileResponse is the full profile returned on protected endpoints.
type UserProfileResponse struct {
	UserID        string     `json:"user_id"`
	Email         string     `json:"email"`
	Username      string     `json:"username"`
	FirstName     string     `json:"first_name"`
	LastName      string     `json:"last_name"`
	Phone         string     `json:"phone"`
	EmailVerified bool       `json:"email_verified"`
	KYCStatus     string     `json:"kyc_status"`
	Status        string     `json:"status"`
	Role          string     `json:"role"`
	CreatedAt     time.Time  `json:"created_at"`
	LastLogin     *time.Time `json:"last_login,omitempty"`
}

// UpdateProfileRequest is the request body for updating a user's profile.
type UpdateProfileRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
}
