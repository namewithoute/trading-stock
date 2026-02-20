package jwtservice

import "context"

// Claims holds the verified user identity extracted from a JWT token.
type Claims struct {
	UserID string
	Email  string
	Role   string
}

// Service is the interface for issuing and validating JWT tokens.
// It is used as a shared utility across middleware and use cases.
type Service interface {
	// GenerateAccessToken issues a short-lived signed JWT access token.
	GenerateAccessToken(ctx context.Context, userID, email, role string) (string, error)

	// GenerateRefreshToken issues a long-lived signed JWT refresh token.
	GenerateRefreshToken(ctx context.Context, userID string) (string, error)

	// ValidateToken parses and verifies an access token, returning its Claims.
	ValidateToken(ctx context.Context, tokenString string) (*Claims, error)
}
