package user

import "time"

// User represents a user entity in the trading system
type User struct {
	ID       string `json:"id" gorm:"primaryKey;type:uuid"`
	Email    string `json:"email" gorm:"uniqueIndex;type:varchar(255);not null"`
	Username string `json:"username" gorm:"uniqueIndex;type:varchar(100);not null"`
	Password string `json:"-" gorm:"type:varchar(255);not null"` // Never expose in JSON

	// Profile information
	FirstName string `json:"first_name" gorm:"type:varchar(100)"`
	LastName  string `json:"last_name" gorm:"type:varchar(100)"`
	Phone     string `json:"phone" gorm:"type:varchar(20)"`

	// Status and verification
	Status        Status    `json:"status" gorm:"type:varchar(20);not null"`
	EmailVerified bool      `json:"email_verified" gorm:"default:false"`
	KYCStatus     KYCStatus `json:"kyc_status" gorm:"type:varchar(20);default:'PENDING'"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at" gorm:"not null"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"not null"`
	LastLogin *time.Time `json:"last_login,omitempty"`
}

// TableName specifies the table name for GORM
func (User) TableName() string {
	return "users"
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
