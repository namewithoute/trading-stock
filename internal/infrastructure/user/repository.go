package user

import (
	"context"
	"errors"
	"time"

	domain "trading-stock/internal/domain/user"

	"gorm.io/gorm"
)

// userRepository implements domain.UserRepository interface
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) domain.Repository {
	return &userRepository{db: db}
}

// Create creates a new user
func (r *userRepository) Create(ctx context.Context, u *domain.User) error {
	return r.db.WithContext(ctx).Create(toUserModel(u)).Error
}

// GetByID retrieves a user by their ID
func (r *userRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	var u UserModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return u.toDomain(), nil
}

// GetByEmail retrieves a user by their email address
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var u UserModel
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return u.toDomain(), nil
}

// GetByUsername retrieves a user by their username
func (r *userRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	var u UserModel
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return u.toDomain(), nil
}

// Update updates an existing user
func (r *userRepository) Update(ctx context.Context, u *domain.User) error {
	return r.db.WithContext(ctx).Save(toUserModel(u)).Error
}

// UpdateStatus updates the user's status
func (r *userRepository) UpdateStatus(ctx context.Context, id string, status domain.Status) error {
	return r.db.WithContext(ctx).
		Model(&UserModel{}).
		Where("id = ?", id).
		Update("status", status).
		Error
}

// UpdateKYCStatus updates the user's KYC status
func (r *userRepository) UpdateKYCStatus(ctx context.Context, id string, kycStatus domain.KYCStatus) error {
	return r.db.WithContext(ctx).
		Model(&UserModel{}).
		Where("id = ?", id).
		Update("kyc_status", kycStatus).
		Error
}

// UpdateLastLogin updates the user's last login timestamp
func (r *userRepository) UpdateLastLogin(ctx context.Context, id string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&UserModel{}).
		Where("id = ?", id).
		Update("last_login", &now).
		Error
}

// Delete soft deletes a user (sets status to INACTIVE)
func (r *userRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&UserModel{}).
		Where("id = ?", id).
		Update("status", domain.StatusInactive).
		Error
}

// List retrieves all users with pagination
func (r *userRepository) List(ctx context.Context, limit, offset int) ([]domain.User, error) {
	var models []UserModel
	err := r.db.WithContext(ctx).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	users := make([]domain.User, 0, len(models))
	for i := range models {
		u := models[i].toDomain()
		if u != nil {
			users = append(users, *u)
		}
	}
	return users, nil
}

// Count returns the total number of users
func (r *userRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&UserModel{}).Count(&count).Error
	return count, err
}

// ExistsByEmail checks if a user with the given email exists
func (r *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&UserModel{}).
		Where("email = ?", email).
		Count(&count).Error
	return count > 0, err
}

// ExistsByUsername checks if a user with the given username exists
func (r *userRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&UserModel{}).
		Where("username = ?", username).
		Count(&count).Error
	return count > 0, err
}
