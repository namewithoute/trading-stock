package portfolio

import (
	"time"

	"trading-stock/internal/domain/portfolio"
)

// PositionDTO is the HTTP response for a single position.
type PositionDTO struct {
	ID                   string    `json:"id"`
	AccountID            string    `json:"account_id"`
	Symbol               string    `json:"symbol"`
	Quantity             int       `json:"quantity"`
	AvgPrice             float64   `json:"avg_price"`
	CurrentPrice         float64   `json:"current_price"`
	TotalCost            float64   `json:"total_cost"`
	CurrentValue         float64   `json:"current_value"`
	UnrealizedPnL        float64   `json:"unrealized_pnl"`
	UnrealizedPnLPercent float64   `json:"unrealized_pnl_percent"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// PortfolioOverviewResponse is the full portfolio overview response.
type PortfolioOverviewResponse struct {
	UserID        string        `json:"user_id"`
	TotalValue    float64       `json:"total_value"`
	TotalPnL      float64       `json:"total_unrealized_pnl"`
	PositionCount int           `json:"position_count"`
	Positions     []PositionDTO `json:"positions"`
}

// PortfolioPerformanceResponse is the performance summary response.
type PortfolioPerformanceResponse struct {
	UserID             string  `json:"user_id"`
	TotalValue         float64 `json:"total_value"`
	TotalUnrealizedPnL float64 `json:"total_unrealized_pnl"`
	PnLPercent         float64 `json:"pnl_percent"`
}

func toPositionDTO(p *portfolio.Position) PositionDTO {
	return PositionDTO{
		ID:                   p.ID,
		AccountID:            p.AccountID,
		Symbol:               p.Symbol,
		Quantity:             p.Quantity,
		AvgPrice:             p.AvgPrice,
		CurrentPrice:         p.CurrentPrice,
		TotalCost:            p.TotalCost(),
		CurrentValue:         p.CurrentValue(),
		UnrealizedPnL:        p.UnrealizedPnL,
		UnrealizedPnLPercent: p.UnrealizedPnLPercent,
		UpdatedAt:            p.UpdatedAt,
	}
}
