package auth

// RegisterRequest is the payload for POST /api/v1/auth/register
type RegisterRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Name     string `json:"name"     validate:"required,min=2"`
}

// LoginRequest is the payload for POST /api/v1/auth/login
type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// RefreshTokenRequest is the payload for POST /api/v1/auth/refresh
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// LogoutRequest is the payload for POST /api/v1/auth/logout
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// LoginResponse is the successful response body for Login
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	UserID       string `json:"user_id"`
}

// RegisterResponse is the successful response body for Register
type RegisterResponse struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	Token  string `json:"token"`
}
