package portfolio

import (
	"time"

	domain "trading-stock/internal/domain/portfolio"
)

// PositionModel is the GORM persistence model for portfolio positions.
type PositionModel struct {
	ID        string `gorm:"primaryKey;type:uuid"`
	UserID    string `gorm:"type:uuid;index;not null"`
	AccountID string `gorm:"type:uuid;index;not null"`
	Symbol    string `gorm:"type:varchar(10);index;not null"`

	Quantity     int     `gorm:"not null"`
	AvgPrice     float64 `gorm:"type:decimal(20,4);not null"`
	CurrentPrice float64 `gorm:"type:decimal(20,4)"`

	UnrealizedPnL        float64 `gorm:"type:decimal(20,2)"`
	UnrealizedPnLPercent float64 `gorm:"type:decimal(10,4)"`

	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (PositionModel) TableName() string { return "positions" }

func toPositionModel(p *domain.Position) *PositionModel {
	if p == nil {
		return nil
	}
	return &PositionModel{
		ID:                   p.ID,
		UserID:               p.UserID,
		AccountID:            p.AccountID,
		Symbol:               p.Symbol,
		Quantity:             p.Quantity,
		AvgPrice:             p.AvgPrice,
		CurrentPrice:         p.CurrentPrice,
		UnrealizedPnL:        p.UnrealizedPnL,
		UnrealizedPnLPercent: p.UnrealizedPnLPercent,
		CreatedAt:            p.CreatedAt,
		UpdatedAt:            p.UpdatedAt,
	}
}

func (m *PositionModel) toDomain() *domain.Position {
	if m == nil {
		return nil
	}
	return &domain.Position{
		ID:                   m.ID,
		UserID:               m.UserID,
		AccountID:            m.AccountID,
		Symbol:               m.Symbol,
		Quantity:             m.Quantity,
		AvgPrice:             m.AvgPrice,
		CurrentPrice:         m.CurrentPrice,
		UnrealizedPnL:        m.UnrealizedPnL,
		UnrealizedPnLPercent: m.UnrealizedPnLPercent,
		CreatedAt:            m.CreatedAt,
		UpdatedAt:            m.UpdatedAt,
	}
}
