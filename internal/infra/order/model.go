package order

import (
	"time"

	domain "trading-stock/internal/domain/order"
)

// OrderModel is the GORM persistence model for orders.
type OrderModel struct {
	ID        string  `gorm:"primaryKey;type:uuid"`
	UserID    string  `gorm:"type:uuid;index;not null"`
	AccountID string  `gorm:"type:uuid;index"`
	Symbol    string  `gorm:"type:varchar(10);index;not null"`
	Price     float64 `gorm:"type:decimal(20,4);not null"`
	Quantity  int     `gorm:"not null"`

	Side      string `gorm:"type:varchar(10);not null"`
	OrderType string `gorm:"column:order_type;type:varchar(20);not null"`
	Status    string `gorm:"type:varchar(20);index;not null"`

	FilledQuantity int     `gorm:"default:0"`
	AvgFillPrice   float64 `gorm:"type:decimal(20,4)"`

	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (OrderModel) TableName() string { return "orders" }

func toOrderModel(o *domain.Order) *OrderModel {
	if o == nil {
		return nil
	}
	return &OrderModel{
		ID:             o.ID,
		UserID:         o.UserID,
		AccountID:      o.AccountID,
		Symbol:         o.Symbol,
		Price:          o.Price,
		Quantity:       o.Quantity,
		Side:           string(o.Side),
		OrderType:      string(o.Type),
		Status:         string(o.Status),
		FilledQuantity: o.FilledQuantity,
		AvgFillPrice:   o.AvgFillPrice,
		CreatedAt:      o.CreatedAt,
		UpdatedAt:      o.UpdatedAt,
	}
}

func (m *OrderModel) toDomain() *domain.Order {
	if m == nil {
		return nil
	}
	return &domain.Order{
		ID:             m.ID,
		UserID:         m.UserID,
		AccountID:      m.AccountID,
		Symbol:         m.Symbol,
		Price:          m.Price,
		Quantity:       m.Quantity,
		Side:           domain.Side(m.Side),
		Type:           domain.OrderType(m.OrderType),
		Status:         domain.Status(m.Status),
		FilledQuantity: m.FilledQuantity,
		AvgFillPrice:   m.AvgFillPrice,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}
