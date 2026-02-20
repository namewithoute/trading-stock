package user

import (
	"time"

	domain "trading-stock/internal/domain/user"
)

// UserModel is the GORM persistence model for users.
type UserModel struct {
	ID       string `gorm:"primaryKey;type:uuid"`
	Email    string `gorm:"uniqueIndex;type:varchar(255);not null"`
	Username string `gorm:"uniqueIndex;type:varchar(100);not null"`
	Password string `gorm:"type:varchar(255);not null"`

	FirstName string `gorm:"type:varchar(100)"`
	LastName  string `gorm:"type:varchar(100)"`
	Phone     string `gorm:"type:varchar(20)"`

	Status        string `gorm:"type:varchar(20);not null"`
	EmailVerified bool   `gorm:"default:false"`
	KYCStatus     string `gorm:"type:varchar(20);default:'PENDING'"`

	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
	LastLogin *time.Time
}

func (UserModel) TableName() string { return "users" }

func toUserModel(u *domain.User) *UserModel {
	if u == nil {
		return nil
	}
	return &UserModel{
		ID:            u.ID,
		Email:         u.Email,
		Username:      u.Username,
		Password:      u.Password,
		FirstName:     u.FirstName,
		LastName:      u.LastName,
		Phone:         u.Phone,
		Status:        string(u.Status),
		EmailVerified: u.EmailVerified,
		KYCStatus:     string(u.KYCStatus),
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
		LastLogin:     u.LastLogin,
	}
}

func (m *UserModel) toDomain() *domain.User {
	if m == nil {
		return nil
	}
	return &domain.User{
		ID:            m.ID,
		Email:         m.Email,
		Username:      m.Username,
		Password:      m.Password,
		FirstName:     m.FirstName,
		LastName:      m.LastName,
		Phone:         m.Phone,
		Status:        domain.Status(m.Status),
		EmailVerified: m.EmailVerified,
		KYCStatus:     domain.KYCStatus(m.KYCStatus),
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
		LastLogin:     m.LastLogin,
	}
}
