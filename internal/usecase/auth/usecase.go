package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"trading-stock/internal/domain/user"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// UseCase handles authentication business logic
type UseCase interface {
	Register(ctx context.Context, email, password, name string) (*user.User, string, error)
	Login(ctx context.Context, email, password string) (*user.User, string, string, error)
	RefreshToken(ctx context.Context, refreshToken string) (string, error)
	Logout(ctx context.Context, userID, token string) error
	ValidateToken(ctx context.Context, token string) (*user.User, error)
}

type useCase struct {
	userRepo user.Repository
	redis    *redis.Client
	logger   *zap.Logger
}

// NewUseCase creates a new auth use case
func NewUseCase(
	userRepo user.Repository,
	redis *redis.Client,
	logger *zap.Logger,
) UseCase {
	return &useCase{
		userRepo: userRepo,
		redis:    redis,
		logger:   logger,
	}
}

// Register registers a new user
func (s *useCase) Register(ctx context.Context, email, password, name string) (*user.User, string, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, email)
	if err == nil && existingUser != nil {
		return nil, "", errors.New("email already registered")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Failed to hash password", zap.Error(err))
		return nil, "", errors.New("failed to process password")
	}

	// Create user
	newUser := &user.User{
		Email:     email,
		Password:  string(hashedPassword),
		FirstName: name,
		Status:    user.StatusActive,
	}

	if err := s.userRepo.Create(ctx, newUser); err != nil {
		s.logger.Error("Failed to create user", zap.Error(err))
		return nil, "", errors.New("failed to create user")
	}

	// Generate JWT token (TODO: implement JWT package)
	token := "mock_jwt_token_" + newUser.ID

	s.logger.Info("User registered successfully", zap.String("user_id", newUser.ID))
	return newUser, token, nil
}

// Login authenticates user and returns tokens
func (s *useCase) Login(ctx context.Context, email, password string) (*user.User, string, string, error) {
	// Find user by email
	u, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, "", "", errors.New("invalid email or password")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		return nil, "", "", errors.New("invalid email or password")
	}

	// Check if user is active
	if !u.IsActive() {
		return nil, "", "", errors.New("account is not active")
	}

	// Generate tokens (TODO: implement JWT package)
	accessToken := "mock_access_token_" + u.ID
	refreshToken := "mock_refresh_token_" + u.ID

	// Store refresh token in Redis
	refreshKey := fmt.Sprintf("refresh_token:%s", u.ID)
	if err := s.redis.Set(ctx, refreshKey, refreshToken, 7*24*time.Hour).Err(); err != nil {
		s.logger.Error("Failed to store refresh token", zap.Error(err))
	}

	s.logger.Info("User logged in successfully", zap.String("user_id", u.ID))
	return u, accessToken, refreshToken, nil
}

// RefreshToken refreshes access token
func (s *useCase) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	// TODO: Validate refresh token and generate new access token
	// For now, return mock token
	newAccessToken := "mock_new_access_token"
	return newAccessToken, nil
}

// Logout logs out user by invalidating tokens
func (s *useCase) Logout(ctx context.Context, userID, token string) error {
	// Add token to blacklist in Redis
	blacklistKey := fmt.Sprintf("blacklist:%s", token)
	if err := s.redis.Set(ctx, blacklistKey, "1", 24*time.Hour).Err(); err != nil {
		s.logger.Error("Failed to blacklist token", zap.Error(err))
		return errors.New("failed to logout")
	}

	// Remove refresh token
	refreshKey := fmt.Sprintf("refresh_token:%s", userID)
	s.redis.Del(ctx, refreshKey)

	s.logger.Info("User logged out successfully", zap.String("user_id", userID))
	return nil
}

// ValidateToken validates JWT token
func (s *useCase) ValidateToken(ctx context.Context, token string) (*user.User, error) {
	// Check if token is blacklisted
	blacklistKey := fmt.Sprintf("blacklist:%s", token)
	exists, err := s.redis.Exists(ctx, blacklistKey).Result()
	if err == nil && exists > 0 {
		return nil, errors.New("token has been revoked")
	}

	// TODO: Validate JWT and extract user ID
	// For now, return mock user
	return nil, errors.New("not implemented")
}
