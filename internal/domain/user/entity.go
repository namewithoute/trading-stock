package user

import (
	"errors"
	"time"
)

var (
	ErrInvalidPassword = errors.New("invalid email or password")
	ErrUserInactive    = errors.New("user account is inactive")
)

// User represents a user entity in the trading system
type User struct {
	ID       string
	Email    string
	Password string // This is the Hashed Password
	Username string

	// Profile information
	FirstName string
	LastName  string
	Phone     string

	// Status and verification
	Status        Status
	EmailVerified bool
	KYCStatus     KYCStatus
	Role          Role

	// Timestamps
	CreatedAt time.Time
	UpdatedAt time.Time
	LastLogin *time.Time
}

// Authenticate checks if the password is correct and user is active
func (u *User) Authenticate(plainPassword string, hasher PasswordHasher) error {
	if !hasher.Compare(u.Password, plainPassword) {
		return ErrInvalidPassword
	}
	if !u.IsActive() {
		return ErrUserInactive
	}
	return nil
}

// UpdatePassword hashes and updates the user's password
func (u *User) UpdatePassword(newPassword string, hasher PasswordHasher) error {
	hash, err := hasher.Hash(newPassword)
	if err != nil {
		return err
	}
	u.Password = hash
	return nil
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
