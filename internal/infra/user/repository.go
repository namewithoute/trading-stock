package user

import (
	"context"
	"errors"
	"time"

	"trading-stock/internal/domain/user"

	"gorm.io/gorm"
)

// userRepository implements user.Repository interface
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) user.Repository {
	return &userRepository{db: db}
}

// Create creates a new user
func (r *userRepository) Create(ctx context.Context, u *user.User) error {
	return r.db.WithContext(ctx).Create(u).Error
}

// GetByID retrieves a user by their ID
func (r *userRepository) GetByID(ctx context.Context, id string) (*user.User, error) {
	var u user.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &u, nil
}

// GetByEmail retrieves a user by their email address
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	var u user.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &u, nil
}

// GetByUsername retrieves a user by their username
func (r *userRepository) GetByUsername(ctx context.Context, username string) (*user.User, error) {
	var u user.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &u, nil
}

// Update updates an existing user
func (r *userRepository) Update(ctx context.Context, u *user.User) error {
	return r.db.WithContext(ctx).Save(u).Error
}

// UpdateStatus updates the user's status
func (r *userRepository) UpdateStatus(ctx context.Context, id string, status user.Status) error {
	return r.db.WithContext(ctx).
		Model(&user.User{}).
		Where("id = ?", id).
		Update("status", status).
		Error
}

// UpdateKYCStatus updates the user's KYC status
func (r *userRepository) UpdateKYCStatus(ctx context.Context, id string, kycStatus user.KYCStatus) error {
	return r.db.WithContext(ctx).
		Model(&user.User{}).
		Where("id = ?", id).
		Update("kyc_status", kycStatus).
		Error
}

// UpdateLastLogin updates the user's last login timestamp
func (r *userRepository) UpdateLastLogin(ctx context.Context, id string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&user.User{}).
		Where("id = ?", id).
		Update("last_login", &now).
		Error
}

// Delete soft deletes a user (sets status to INACTIVE)
func (r *userRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&user.User{}).
		Where("id = ?", id).
		Update("status", user.StatusInactive).
		Error
}

// List retrieves all users with pagination
func (r *userRepository) List(ctx context.Context, limit, offset int) ([]*user.User, error) {
	var users []*user.User
	err := r.db.WithContext(ctx).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&users).Error
	return users, err
}

// Count returns the total number of users
func (r *userRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&user.User{}).Count(&count).Error
	return count, err
}

// ExistsByEmail checks if a user with the given email exists
func (r *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&user.User{}).
		Where("email = ?", email).
		Count(&count).Error
	return count > 0, err
}

// ExistsByUsername checks if a user with the given username exists
func (r *userRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&user.User{}).
		Where("username = ?", username).
		Count(&count).Error
	return count > 0, err
}
