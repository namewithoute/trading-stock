package risk

import "time"

// RiskMetricsResponse is the HTTP response for risk metrics.
type RiskMetricsResponse struct {
	AccountID          string    `json:"account_id"`
	UserID             string    `json:"user_id"`
	TotalExposure      float64   `json:"total_exposure"`
	LongExposure       float64   `json:"long_exposure"`
	ShortExposure      float64   `json:"short_exposure"`
	PositionsCount     int       `json:"positions_count"`
	LargestPosition    float64   `json:"largest_position"`
	ConcentrationRatio float64   `json:"concentration_ratio"`
	DailyPnL           float64   `json:"daily_pnl"`
	WeeklyPnL          float64   `json:"weekly_pnl"`
	MonthlyPnL         float64   `json:"monthly_pnl"`
	DailyOrdersCount   int       `json:"daily_orders_count"`
	CurrentLeverage    float64   `json:"current_leverage"`
	RiskScore          int       `json:"risk_score"`
	IsHighRisk         bool      `json:"is_high_risk"`
	IsMediumRisk       bool      `json:"is_medium_risk"`
	LastCalculatedAt   time.Time `json:"last_calculated_at"`
}

// RiskAlertResponse is the HTTP response for a single risk alert.
type RiskAlertResponse struct {
	ID         string     `json:"id"`
	AccountID  string     `json:"account_id"`
	UserID     string     `json:"user_id"`
	AlertType  string     `json:"alert_type"`
	Severity   string     `json:"severity"`
	Message    string     `json:"message"`
	Status     string     `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`
}

// CheckOrderRiskRequest is the request body for order risk checks.
type CheckOrderRiskRequest struct {
	AccountID string  `json:"account_id" validate:"required"`
	Symbol    string  `json:"symbol"     validate:"required"`
	Amount    float64 `json:"amount"     validate:"required,gt=0"`
}
