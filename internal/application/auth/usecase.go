package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"trading-stock/internal/domain/user"
	"trading-stock/pkg/jwtservice"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// UseCase defines the application-level authentication operations.
type UseCase interface {
	Register(ctx context.Context, email, password, name string) (*user.User, string, error)
	Login(ctx context.Context, email, password string) (*user.User, string, string, error)
	RefreshToken(ctx context.Context, refreshToken string) (string, error)
	Logout(ctx context.Context, userID, refreshToken string) error
	ValidateToken(ctx context.Context, tokenString string) (*jwtservice.Claims, error)
}

type useCase struct {
	userRepo   user.Repository
	hasher     user.PasswordHasher
	jwtService jwtservice.Service
	redis      *redis.Client
	logger     *zap.Logger
}

// NewUseCase creates a new auth UseCase with all required dependencies.
func NewUseCase(
	userRepo user.Repository,
	hasher user.PasswordHasher,
	jwtSvc jwtservice.Service,
	redis *redis.Client,
	logger *zap.Logger,
) UseCase {
	return &useCase{
		userRepo:   userRepo,
		hasher:     hasher,
		jwtService: jwtSvc,
		redis:      redis,
		logger:     logger,
	}
}

// Register creates a new user account and issues a first-time access token.
func (s *useCase) Register(ctx context.Context, email, password, name string) (*user.User, string, error) {
	existingUser, err := s.userRepo.GetByEmail(ctx, email)
	if err == nil && existingUser != nil {
		return nil, "", errors.New("email already registered")
	}

	newUser := &user.User{
		ID:        uuid.New().String(),
		Email:     email,
		Username:  email,
		FirstName: name,
		Status:    user.StatusActive,
	}

	if err := newUser.UpdatePassword(password, s.hasher); err != nil {
		s.logger.Error("Failed to hash password", zap.Error(err))
		return nil, "", errors.New("failed to process password")
	}

	if err := s.userRepo.Create(ctx, newUser); err != nil {
		s.logger.Error("Failed to create user", zap.Error(err))
		return nil, "", errors.New("failed to create user")
	}

	token, err := s.jwtService.GenerateAccessToken(ctx, newUser.ID, newUser.Email, string(user.RoleUser))
	if err != nil {
		s.logger.Error("Failed to generate access token", zap.Error(err))
		return nil, "", errors.New("user created but failed to generate token")
	}

	s.logger.Info("User registered successfully", zap.String("user_id", newUser.ID))
	return newUser, token, nil
}

// Login verifies credentials and returns a JWT access + refresh token pair.
func (s *useCase) Login(ctx context.Context, email, password string) (*user.User, string, string, error) {
	u, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, "", "", user.ErrInvalidPassword
	}

	if err := u.Authenticate(password, s.hasher); err != nil {
		return nil, "", "", err
	}

	accessToken, err := s.jwtService.GenerateAccessToken(ctx, u.ID, u.Email, string(user.RoleUser))
	if err != nil {
		s.logger.Error("Failed to generate access token", zap.Error(err))
		return nil, "", "", errors.New("failed to generate access token")
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(ctx, u.ID)
	if err != nil {
		s.logger.Error("Failed to generate refresh token", zap.Error(err))
		return nil, "", "", errors.New("failed to generate refresh token")
	}

	refreshKey := fmt.Sprintf("refresh_token:%s", u.ID)
	if err := s.redis.Set(ctx, refreshKey, refreshToken, 7*24*time.Hour).Err(); err != nil {
		s.logger.Error("Failed to store refresh token in Redis", zap.Error(err))
	}

	s.logger.Info("User logged in successfully", zap.String("user_id", u.ID))
	return u, accessToken, refreshToken, nil
}

// RefreshToken validates a refresh token and issues a new access token.
func (s *useCase) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	claims, err := s.jwtService.ValidateToken(ctx, refreshToken)
	if err != nil {
		return "", errors.New("invalid or expired refresh token")
	}

	refreshKey := fmt.Sprintf("refresh_token:%s", claims.UserID)
	storedToken, err := s.redis.Get(ctx, refreshKey).Result()
	if err != nil || storedToken != refreshToken {
		return "", errors.New("refresh token has been revoked")
	}

	newAccessToken, err := s.jwtService.GenerateAccessToken(ctx, claims.UserID, claims.Email, claims.Role)
	if err != nil {
		return "", errors.New("failed to generate new access token")
	}

	return newAccessToken, nil
}

// Logout invalidates the user session by deleting the refresh token from Redis.
func (s *useCase) Logout(ctx context.Context, userID, refreshToken string) error {
	blacklistKey := fmt.Sprintf("blacklist:%s", refreshToken)
	if err := s.redis.Set(ctx, blacklistKey, "1", 7*24*time.Hour).Err(); err != nil {
		s.logger.Error("Failed to blacklist refresh token", zap.Error(err))
		return errors.New("failed to logout")
	}

	refreshKey := fmt.Sprintf("refresh_token:%s", userID)
	s.redis.Del(ctx, refreshKey)

	s.logger.Info("User logged out successfully", zap.String("user_id", userID))
	return nil
}

// ValidateToken validates an access token and returns its Claims.
// Used by AuthMiddleware on every protected request.
func (s *useCase) ValidateToken(ctx context.Context, tokenString string) (*jwtservice.Claims, error) {
	blacklistKey := fmt.Sprintf("blacklist:%s", tokenString)
	exists, err := s.redis.Exists(ctx, blacklistKey).Result()
	if err == nil && exists > 0 {
		return nil, errors.New("token has been revoked")
	}

	return s.jwtService.ValidateToken(ctx, tokenString)
}
