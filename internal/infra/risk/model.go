package risk

import (
	"time"

	domain "trading-stock/internal/domain/risk"
)

// RiskLimitModel is the GORM persistence model for risk limits.
type RiskLimitModel struct {
	ID        string `gorm:"primaryKey;type:uuid"`
	AccountID string `gorm:"type:uuid;uniqueIndex;not null"`
	UserID    string `gorm:"type:uuid;index;not null"`

	MaxPositionSize   int     `gorm:"not null;default:10000"`
	MaxPositionValue  float64 `gorm:"type:decimal(20,2);not null;default:100000"`
	MaxPositionsCount int     `gorm:"not null;default:50"`

	MaxOrderSize   int     `gorm:"not null;default:1000"`
	MaxOrderValue  float64 `gorm:"type:decimal(20,2);not null;default:50000"`
	MaxDailyOrders int     `gorm:"not null;default:100"`

	MaxDailyLoss   float64 `gorm:"type:decimal(20,2);not null;default:5000"`
	MaxWeeklyLoss  float64 `gorm:"type:decimal(20,2);not null;default:20000"`
	MaxMonthlyLoss float64 `gorm:"type:decimal(20,2);not null;default:50000"`

	MaxLeverage      float64   `gorm:"type:decimal(10,2);not null;default:1.0"`
	MaxConcentration float64   `gorm:"type:decimal(5,2);not null;default:0.25"`
	Status           string    `gorm:"type:varchar(20);not null;default:'ACTIVE'"`
	CreatedAt        time.Time `gorm:"not null"`
	UpdatedAt        time.Time `gorm:"not null"`
}

func (RiskLimitModel) TableName() string { return "risk_limits" }

func toRiskLimitModel(rl *domain.RiskLimit) *RiskLimitModel {
	if rl == nil {
		return nil
	}
	return &RiskLimitModel{
		ID:                rl.ID,
		AccountID:         rl.AccountID,
		UserID:            rl.UserID,
		MaxPositionSize:   rl.MaxPositionSize,
		MaxPositionValue:  rl.MaxPositionValue,
		MaxPositionsCount: rl.MaxPositionsCount,
		MaxOrderSize:      rl.MaxOrderSize,
		MaxOrderValue:     rl.MaxOrderValue,
		MaxDailyOrders:    rl.MaxDailyOrders,
		MaxDailyLoss:      rl.MaxDailyLoss,
		MaxWeeklyLoss:     rl.MaxWeeklyLoss,
		MaxMonthlyLoss:    rl.MaxMonthlyLoss,
		MaxLeverage:       rl.MaxLeverage,
		MaxConcentration:  rl.MaxConcentration,
		Status:            string(rl.Status),
		CreatedAt:         rl.CreatedAt,
		UpdatedAt:         rl.UpdatedAt,
	}
}

func (m *RiskLimitModel) toDomain() *domain.RiskLimit {
	if m == nil {
		return nil
	}
	return &domain.RiskLimit{
		ID:                m.ID,
		AccountID:         m.AccountID,
		UserID:            m.UserID,
		MaxPositionSize:   m.MaxPositionSize,
		MaxPositionValue:  m.MaxPositionValue,
		MaxPositionsCount: m.MaxPositionsCount,
		MaxOrderSize:      m.MaxOrderSize,
		MaxOrderValue:     m.MaxOrderValue,
		MaxDailyOrders:    m.MaxDailyOrders,
		MaxDailyLoss:      m.MaxDailyLoss,
		MaxWeeklyLoss:     m.MaxWeeklyLoss,
		MaxMonthlyLoss:    m.MaxMonthlyLoss,
		MaxLeverage:       m.MaxLeverage,
		MaxConcentration:  m.MaxConcentration,
		Status:            domain.LimitStatus(m.Status),
		CreatedAt:         m.CreatedAt,
		UpdatedAt:         m.UpdatedAt,
	}
}

// RiskMetricsModel is the GORM persistence model for risk metrics.
type RiskMetricsModel struct {
	ID        string `gorm:"primaryKey;type:uuid"`
	AccountID string `gorm:"type:uuid;uniqueIndex;not null"`
	UserID    string `gorm:"type:uuid;index;not null"`

	TotalExposure float64 `gorm:"type:decimal(20,2);not null;default:0"`
	LongExposure  float64 `gorm:"type:decimal(20,2);not null;default:0"`
	ShortExposure float64 `gorm:"type:decimal(20,2);not null;default:0"`

	PositionsCount     int     `gorm:"not null;default:0"`
	LargestPosition    float64 `gorm:"type:decimal(20,2);not null;default:0"`
	ConcentrationRatio float64 `gorm:"type:decimal(5,4);not null;default:0"`

	DailyPnL   float64 `gorm:"type:decimal(20,2);not null;default:0"`
	WeeklyPnL  float64 `gorm:"type:decimal(20,2);not null;default:0"`
	MonthlyPnL float64 `gorm:"type:decimal(20,2);not null;default:0"`

	DailyOrdersCount int `gorm:"not null;default:0"`

	CurrentLeverage float64 `gorm:"type:decimal(10,2);not null;default:0"`
	RiskScore       int     `gorm:"not null;default:0"`

	LastCalculatedAt time.Time `gorm:"not null"`
	CreatedAt        time.Time `gorm:"not null"`
	UpdatedAt        time.Time `gorm:"not null"`
}

func (RiskMetricsModel) TableName() string { return "risk_metrics" }

func toRiskMetricsModel(rm *domain.RiskMetrics) *RiskMetricsModel {
	if rm == nil {
		return nil
	}
	return &RiskMetricsModel{
		ID:                 rm.ID,
		AccountID:          rm.AccountID,
		UserID:             rm.UserID,
		TotalExposure:      rm.TotalExposure,
		LongExposure:       rm.LongExposure,
		ShortExposure:      rm.ShortExposure,
		PositionsCount:     rm.PositionsCount,
		LargestPosition:    rm.LargestPosition,
		ConcentrationRatio: rm.ConcentrationRatio,
		DailyPnL:           rm.DailyPnL,
		WeeklyPnL:          rm.WeeklyPnL,
		MonthlyPnL:         rm.MonthlyPnL,
		DailyOrdersCount:   rm.DailyOrdersCount,
		CurrentLeverage:    rm.CurrentLeverage,
		RiskScore:          rm.RiskScore,
		LastCalculatedAt:   rm.LastCalculatedAt,
		CreatedAt:          rm.CreatedAt,
		UpdatedAt:          rm.UpdatedAt,
	}
}

func (m *RiskMetricsModel) toDomain() *domain.RiskMetrics {
	if m == nil {
		return nil
	}
	return &domain.RiskMetrics{
		ID:                 m.ID,
		AccountID:          m.AccountID,
		UserID:             m.UserID,
		TotalExposure:      m.TotalExposure,
		LongExposure:       m.LongExposure,
		ShortExposure:      m.ShortExposure,
		PositionsCount:     m.PositionsCount,
		LargestPosition:    m.LargestPosition,
		ConcentrationRatio: m.ConcentrationRatio,
		DailyPnL:           m.DailyPnL,
		WeeklyPnL:          m.WeeklyPnL,
		MonthlyPnL:         m.MonthlyPnL,
		DailyOrdersCount:   m.DailyOrdersCount,
		CurrentLeverage:    m.CurrentLeverage,
		RiskScore:          m.RiskScore,
		LastCalculatedAt:   m.LastCalculatedAt,
		CreatedAt:          m.CreatedAt,
		UpdatedAt:          m.UpdatedAt,
	}
}

// RiskAlertModel is the GORM persistence model for risk alerts.
type RiskAlertModel struct {
	ID        string `gorm:"primaryKey;type:uuid"`
	AccountID string `gorm:"type:uuid;index;not null"`
	UserID    string `gorm:"type:uuid;index;not null"`

	AlertType string `gorm:"type:varchar(50);not null"`
	Severity  string `gorm:"type:varchar(20);not null"`
	Message   string `gorm:"type:text;not null"`
	Details   string `gorm:"type:jsonb"`
	Status    string `gorm:"type:varchar(20);not null"`

	ResolvedAt *time.Time
	CreatedAt  time.Time `gorm:"not null;index"`
}

func (RiskAlertModel) TableName() string { return "risk_alerts" }

func toRiskAlertModel(ra *domain.RiskAlert) *RiskAlertModel {
	if ra == nil {
		return nil
	}
	return &RiskAlertModel{
		ID:         ra.ID,
		AccountID:  ra.AccountID,
		UserID:     ra.UserID,
		AlertType:  string(ra.AlertType),
		Severity:   string(ra.Severity),
		Message:    ra.Message,
		Details:    ra.Details,
		Status:     string(ra.Status),
		ResolvedAt: ra.ResolvedAt,
		CreatedAt:  ra.CreatedAt,
	}
}

func (m *RiskAlertModel) toDomain() *domain.RiskAlert {
	if m == nil {
		return nil
	}
	return &domain.RiskAlert{
		ID:         m.ID,
		AccountID:  m.AccountID,
		UserID:     m.UserID,
		AlertType:  domain.AlertType(m.AlertType),
		Severity:   domain.Severity(m.Severity),
		Message:    m.Message,
		Details:    m.Details,
		Status:     domain.AlertStatus(m.Status),
		ResolvedAt: m.ResolvedAt,
		CreatedAt:  m.CreatedAt,
	}
}
