package jwtservice

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// jwtClaims is the internal struct encoded inside each JWT.
type jwtClaims struct {
	UserID string `json:"uid"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// Config holds the settings required to initialise a JWT service.
type Config struct {
	AccessSecret  string
	RefreshSecret string
	AccessTTL     time.Duration // e.g. 15m
	RefreshTTL    time.Duration // e.g. 168h (7 days)
	Issuer        string
}

// jwtService is the concrete implementation of Service using golang-jwt.
type jwtService struct {
	accessSecret  []byte
	refreshSecret []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
	issuer        string
}

// New creates a new JWT Service from the provided configuration.
func New(cfg Config) Service {
	return &jwtService{
		accessSecret:  []byte(cfg.AccessSecret),
		refreshSecret: []byte(cfg.RefreshSecret),
		accessTTL:     cfg.AccessTTL,
		refreshTTL:    cfg.RefreshTTL,
		issuer:        cfg.Issuer,
	}
}

// GenerateAccessToken creates a short-lived HS256 signed JWT access token.
func (s *jwtService) GenerateAccessToken(_ context.Context, userID, email, role string) (string, error) {
	claims := jwtClaims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.accessTTL)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.accessSecret)
}

// GenerateRefreshToken creates a long-lived HS256 signed JWT refresh token.
// Refresh tokens carry minimal claims (only userID) for security.
func (s *jwtService) GenerateRefreshToken(_ context.Context, userID string) (string, error) {
	claims := jwtClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.refreshTTL)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.refreshSecret)
}

// ValidateToken parses and validates an access token.
// It enforces the signing method, signature, and expiry.
func (s *jwtService) ValidateToken(_ context.Context, tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&jwtClaims{},
		func(t *jwt.Token) (interface{}, error) {
			// Reject tokens signed with an unexpected algorithm
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected token signing method")
			}
			return s.accessSecret, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuedAt(),
	)
	if err != nil {
		return nil, err
	}

	c, ok := token.Claims.(*jwtClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return &Claims{
		UserID: c.UserID,
		Email:  c.Email,
		Role:   c.Role,
	}, nil
}

// ValidateRefreshToken parses and validates a refresh token.
// Uses the refresh secret (separate from the access token secret).
func (s *jwtService) ValidateRefreshToken(_ context.Context, tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&jwtClaims{},
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected token signing method")
			}
			return s.refreshSecret, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil {
		return nil, err
	}

	c, ok := token.Claims.(*jwtClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid refresh token claims")
	}

	return &Claims{UserID: c.UserID}, nil
}
