package user

import "time"

// User represents a user entity in the trading system
type User struct {
	ID       string
	Email    string
	Username string
	Password string // Never expose in JSON via DTOs

	// Profile information
	FirstName string
	LastName  string
	Phone     string

	// Status and verification
	Status        Status
	EmailVerified bool
	KYCStatus     KYCStatus

	// Timestamps
	CreatedAt time.Time
	UpdatedAt time.Time
	LastLogin *time.Time
}

// IsActive checks if the user account is active
func (u *User) IsActive() bool {
	return u.Status == StatusActive
}

// CanTrade checks if the user can perform trading operations
func (u *User) CanTrade() bool {
	return u.IsActive() && u.EmailVerified && u.KYCStatus == KYCApproved
}

// FullName returns the user's full name
func (u *User) FullName() string {
	if u.FirstName == "" && u.LastName == "" {
		return u.Username
	}
	return u.FirstName + " " + u.LastName
}
