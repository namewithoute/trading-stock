package user

import "context"

// Repository defines the interface for user data access
type Repository interface {
	// Create creates a new user in the database
	Create(ctx context.Context, user *User) error

	// GetByID retrieves a user by their ID
	GetByID(ctx context.Context, id string) (*User, error)

	// GetByEmail retrieves a user by their email address
	GetByEmail(ctx context.Context, email string) (*User, error)

	// GetByUsername retrieves a user by their username
	GetByUsername(ctx context.Context, username string) (*User, error)

	// Update updates an existing user
	Update(ctx context.Context, user *User) error

	// UpdateStatus updates the user's status
	UpdateStatus(ctx context.Context, id string, status Status) error

	// UpdateKYCStatus updates the user's KYC status
	UpdateKYCStatus(ctx context.Context, id string, kycStatus KYCStatus) error

	// UpdateLastLogin updates the user's last login timestamp
	UpdateLastLogin(ctx context.Context, id string) error

	// Delete soft deletes a user (sets status to INACTIVE)
	Delete(ctx context.Context, id string) error

	// List retrieves all users with pagination
	List(ctx context.Context, limit, offset int) ([]User, error)

	// Count returns the total number of users
	Count(ctx context.Context) (int64, error)

	// ExistsByEmail checks if a user with the given email exists
	ExistsByEmail(ctx context.Context, email string) (bool, error)

	// ExistsByUsername checks if a user with the given username exists
	ExistsByUsername(ctx context.Context, username string) (bool, error)
}
