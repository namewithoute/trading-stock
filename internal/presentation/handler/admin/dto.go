package admin

import (
	"time"

	"trading-stock/internal/domain/order"
	"trading-stock/internal/domain/user"
)

// ApproveKYCRequest is the request body for approving/rejecting KYC.
type ApproveKYCRequest struct {
	Status string `json:"status" validate:"required,oneof=approved rejected"`
	Reason string `json:"reason"`
}

// UserAdminDTO is the admin view of a user.
type UserAdminDTO struct {
	UserID        string     `json:"user_id"`
	Email         string     `json:"email"`
	Username      string     `json:"username"`
	FirstName     string     `json:"first_name"`
	LastName      string     `json:"last_name"`
	Phone         string     `json:"phone"`
	Status        string     `json:"status"`
	KYCStatus     string     `json:"kyc_status"`
	EmailVerified bool       `json:"email_verified"`
	Role          string     `json:"role"`
	CreatedAt     time.Time  `json:"created_at"`
	LastLogin     *time.Time `json:"last_login,omitempty"`
}

func toUserAdminDTO(u user.User) UserAdminDTO {
	return UserAdminDTO{
		UserID:        u.ID,
		Email:         u.Email,
		Username:      u.Username,
		FirstName:     u.FirstName,
		LastName:      u.LastName,
		Phone:         u.Phone,
		Status:        string(u.Status),
		KYCStatus:     string(u.KYCStatus),
		EmailVerified: u.EmailVerified,
		Role:          string(u.Role),
		CreatedAt:     u.CreatedAt,
		LastLogin:     u.LastLogin,
	}
}

// OrderAdminDTO is the admin view of an order.
type OrderAdminDTO struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	AccountID      string    `json:"account_id"`
	Symbol         string    `json:"symbol"`
	Side           string    `json:"side"`
	OrderType      string    `json:"order_type"`
	Quantity       int       `json:"quantity"`
	Price          float64   `json:"price"`
	FilledQuantity int       `json:"filled_quantity"`
	AvgFillPrice   float64   `json:"avg_fill_price"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func toOrderAdminDTO(o *order.OrderReadModel) OrderAdminDTO {
	return OrderAdminDTO{
		ID:             o.ID,
		UserID:         o.UserID,
		AccountID:      o.AccountID,
		Symbol:         o.Symbol,
		Side:           string(o.Side),
		OrderType:      string(o.OrderType),
		Quantity:       o.Quantity,
		Price:          o.Price,
		FilledQuantity: o.FilledQuantity,
		AvgFillPrice:   o.AvgFillPrice,
		Status:         string(o.Status),
		CreatedAt:      o.CreatedAt,
		UpdatedAt:      o.UpdatedAt,
	}
}
