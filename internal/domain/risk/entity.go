package risk

import (
	"time"
)

// RiskLimit represents risk limits for an account or user
type RiskLimit struct {
	ID        string
	AccountID string
	UserID    string

	// Position limits
	MaxPositionSize   int
	MaxPositionValue  float64
	MaxPositionsCount int

	// Order limits
	MaxOrderSize   int
	MaxOrderValue  float64
	MaxDailyOrders int

	// Loss limits
	MaxDailyLoss   float64
	MaxWeeklyLoss  float64
	MaxMonthlyLoss float64

	// Leverage limits
	MaxLeverage float64

	// Concentration limits
	MaxConcentration float64 // 25% max per position

	Status    LimitStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

// IsActive checks if the risk limit is active
func (rl *RiskLimit) IsActive() bool {
	return rl.Status == LimitStatusActive
}

// CheckPositionSize validates if a position size is within limits
func (rl *RiskLimit) CheckPositionSize(size int) bool {
	return size <= rl.MaxPositionSize
}

// CheckPositionValue validates if a position value is within limits
func (rl *RiskLimit) CheckPositionValue(value float64) bool {
	return value <= rl.MaxPositionValue
}

// CheckOrderSize validates if an order size is within limits
func (rl *RiskLimit) CheckOrderSize(size int) bool {
	return size <= rl.MaxOrderSize
}

// CheckOrderValue validates if an order value is within limits
func (rl *RiskLimit) CheckOrderValue(value float64) bool {
	return value <= rl.MaxOrderValue
}

// RiskMetrics represents current risk metrics for an account
type RiskMetrics struct {
	ID        string
	AccountID string
	UserID    string

	// Current exposure
	TotalExposure float64
	LongExposure  float64
	ShortExposure float64

	// Position metrics
	PositionsCount     int
	LargestPosition    float64
	ConcentrationRatio float64

	// P&L metrics
	DailyPnL   float64
	WeeklyPnL  float64
	MonthlyPnL float64

	// Order metrics
	DailyOrdersCount int

	// Leverage
	CurrentLeverage float64

	// Risk score (0-100, higher = riskier)
	RiskScore int

	LastCalculatedAt time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// CalculateRiskScore calculates a risk score based on current metrics
func (rm *RiskMetrics) CalculateRiskScore() int {
	score := 0

	// Leverage risk (0-30 points)
	if rm.CurrentLeverage > 3.0 {
		score += 30
	} else if rm.CurrentLeverage > 2.0 {
		score += 20
	} else if rm.CurrentLeverage > 1.0 {
		score += 10
	}

	// Concentration risk (0-25 points)
	if rm.ConcentrationRatio > 0.5 {
		score += 25
	} else if rm.ConcentrationRatio > 0.3 {
		score += 15
	} else if rm.ConcentrationRatio > 0.2 {
		score += 5
	}

	// Loss risk (0-25 points)
	if rm.DailyPnL < -5000 {
		score += 25
	} else if rm.DailyPnL < -2000 {
		score += 15
	} else if rm.DailyPnL < -1000 {
		score += 5
	}

	// Position count risk (0-20 points)
	if rm.PositionsCount > 50 {
		score += 20
	} else if rm.PositionsCount > 30 {
		score += 10
	}

	rm.RiskScore = score
	return score
}

// IsHighRisk checks if the account is high risk
func (rm *RiskMetrics) IsHighRisk() bool {
	return rm.RiskScore >= 70
}

// IsMediumRisk checks if the account is medium risk
func (rm *RiskMetrics) IsMediumRisk() bool {
	return rm.RiskScore >= 40 && rm.RiskScore < 70
}

// RiskAlert represents a risk alert/violation
type RiskAlert struct {
	ID         string
	AccountID  string
	UserID     string
	AlertType  AlertType
	Severity   Severity
	Message    string
	Details    string
	Status     AlertStatus
	ResolvedAt *time.Time
	CreatedAt  time.Time
}

// IsResolved checks if the alert has been resolved
func (ra *RiskAlert) IsResolved() bool {
	return ra.Status == AlertStatusResolved
}

// Resolve marks the alert as resolved
func (ra *RiskAlert) Resolve() {
	now := time.Now()
	ra.Status = AlertStatusResolved
	ra.ResolvedAt = &now
}

// IsCritical checks if the alert is critical severity
func (ra *RiskAlert) IsCritical() bool {
	return ra.Severity == SeverityCritical
}
