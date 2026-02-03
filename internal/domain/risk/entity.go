package risk

import (
	"time"
)

// RiskLimit represents risk limits for an account or user
type RiskLimit struct {
	ID        string `json:"id" gorm:"primaryKey;type:uuid"`
	AccountID string `json:"account_id" gorm:"type:uuid;uniqueIndex;not null"`
	UserID    string `json:"user_id" gorm:"type:uuid;index;not null"`

	// Position limits
	MaxPositionSize   int     `json:"max_position_size" gorm:"not null;default:10000"`
	MaxPositionValue  float64 `json:"max_position_value" gorm:"type:decimal(20,2);not null;default:100000"`
	MaxPositionsCount int     `json:"max_positions_count" gorm:"not null;default:50"`

	// Order limits
	MaxOrderSize   int     `json:"max_order_size" gorm:"not null;default:1000"`
	MaxOrderValue  float64 `json:"max_order_value" gorm:"type:decimal(20,2);not null;default:50000"`
	MaxDailyOrders int     `json:"max_daily_orders" gorm:"not null;default:100"`

	// Loss limits
	MaxDailyLoss   float64 `json:"max_daily_loss" gorm:"type:decimal(20,2);not null;default:5000"`
	MaxWeeklyLoss  float64 `json:"max_weekly_loss" gorm:"type:decimal(20,2);not null;default:20000"`
	MaxMonthlyLoss float64 `json:"max_monthly_loss" gorm:"type:decimal(20,2);not null;default:50000"`

	// Leverage limits
	MaxLeverage float64 `json:"max_leverage" gorm:"type:decimal(10,2);not null;default:1.0"`

	// Concentration limits
	MaxConcentration float64 `json:"max_concentration" gorm:"type:decimal(5,2);not null;default:0.25"` // 25% max per position

	Status    LimitStatus `json:"status" gorm:"type:varchar(20);not null;default:'ACTIVE'"`
	CreatedAt time.Time   `json:"created_at" gorm:"not null"`
	UpdatedAt time.Time   `json:"updated_at" gorm:"not null"`
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
	ID        string `json:"id" gorm:"primaryKey;type:uuid"`
	AccountID string `json:"account_id" gorm:"type:uuid;uniqueIndex;not null"`
	UserID    string `json:"user_id" gorm:"type:uuid;index;not null"`

	// Current exposure
	TotalExposure float64 `json:"total_exposure" gorm:"type:decimal(20,2);not null;default:0"`
	LongExposure  float64 `json:"long_exposure" gorm:"type:decimal(20,2);not null;default:0"`
	ShortExposure float64 `json:"short_exposure" gorm:"type:decimal(20,2);not null;default:0"`

	// Position metrics
	PositionsCount     int     `json:"positions_count" gorm:"not null;default:0"`
	LargestPosition    float64 `json:"largest_position" gorm:"type:decimal(20,2);not null;default:0"`
	ConcentrationRatio float64 `json:"concentration_ratio" gorm:"type:decimal(5,4);not null;default:0"`

	// P&L metrics
	DailyPnL   float64 `json:"daily_pnl" gorm:"type:decimal(20,2);not null;default:0"`
	WeeklyPnL  float64 `json:"weekly_pnl" gorm:"type:decimal(20,2);not null;default:0"`
	MonthlyPnL float64 `json:"monthly_pnl" gorm:"type:decimal(20,2);not null;default:0"`

	// Order metrics
	DailyOrdersCount int `json:"daily_orders_count" gorm:"not null;default:0"`

	// Leverage
	CurrentLeverage float64 `json:"current_leverage" gorm:"type:decimal(10,2);not null;default:0"`

	// Risk score (0-100, higher = riskier)
	RiskScore int `json:"risk_score" gorm:"not null;default:0"`

	LastCalculatedAt time.Time `json:"last_calculated_at" gorm:"not null"`
	CreatedAt        time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt        time.Time `json:"updated_at" gorm:"not null"`
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
	ID         string      `json:"id" gorm:"primaryKey;type:uuid"`
	AccountID  string      `json:"account_id" gorm:"type:uuid;index;not null"`
	UserID     string      `json:"user_id" gorm:"type:uuid;index;not null"`
	AlertType  AlertType   `json:"alert_type" gorm:"type:varchar(50);not null"`
	Severity   Severity    `json:"severity" gorm:"type:varchar(20);not null"`
	Message    string      `json:"message" gorm:"type:text;not null"`
	Details    string      `json:"details,omitempty" gorm:"type:jsonb"`
	Status     AlertStatus `json:"status" gorm:"type:varchar(20);not null"`
	ResolvedAt *time.Time  `json:"resolved_at,omitempty"`
	CreatedAt  time.Time   `json:"created_at" gorm:"not null;index"`
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
