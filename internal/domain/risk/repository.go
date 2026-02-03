package risk

import (
	"context"
	"time"
)

// RiskLimitRepository defines the interface for risk limit data access
type RiskLimitRepository interface {
	// Create creates a new risk limit
	Create(ctx context.Context, limit *RiskLimit) error

	// GetByID retrieves a risk limit by ID
	GetByID(ctx context.Context, id string) (*RiskLimit, error)

	// GetByAccountID retrieves risk limit by account ID
	GetByAccountID(ctx context.Context, accountID string) (*RiskLimit, error)

	// GetByUserID retrieves risk limit by user ID
	GetByUserID(ctx context.Context, userID string) (*RiskLimit, error)

	// Update updates a risk limit
	Update(ctx context.Context, limit *RiskLimit) error

	// UpdateStatus updates the status of a risk limit
	UpdateStatus(ctx context.Context, id string, status LimitStatus) error

	// ListByStatus retrieves risk limits by status
	ListByStatus(ctx context.Context, status LimitStatus, limit, offset int) ([]*RiskLimit, error)

	// Delete deletes a risk limit
	Delete(ctx context.Context, id string) error
}

// RiskMetricsRepository defines the interface for risk metrics data access
type RiskMetricsRepository interface {
	// Create creates new risk metrics
	Create(ctx context.Context, metrics *RiskMetrics) error

	// GetByID retrieves risk metrics by ID
	GetByID(ctx context.Context, id string) (*RiskMetrics, error)

	// GetByAccountID retrieves risk metrics by account ID
	GetByAccountID(ctx context.Context, accountID string) (*RiskMetrics, error)

	// Update updates risk metrics
	Update(ctx context.Context, metrics *RiskMetrics) error

	// UpdatePnL updates P&L metrics
	UpdatePnL(ctx context.Context, accountID string, dailyPnL, weeklyPnL, monthlyPnL float64) error

	// UpdateExposure updates exposure metrics
	UpdateExposure(ctx context.Context, accountID string, total, long, short float64) error

	// IncrementDailyOrders increments daily orders count
	IncrementDailyOrders(ctx context.Context, accountID string) error

	// ResetDailyMetrics resets daily metrics (called at day end)
	ResetDailyMetrics(ctx context.Context) error

	// ListHighRisk retrieves accounts with high risk scores
	ListHighRisk(ctx context.Context, minScore int, limit int) ([]*RiskMetrics, error)

	// GetAggregateMetrics gets aggregate risk metrics across all accounts
	GetAggregateMetrics(ctx context.Context) (map[string]interface{}, error)
}

// RiskAlertRepository defines the interface for risk alert data access
type RiskAlertRepository interface {
	// Create creates a new risk alert
	Create(ctx context.Context, alert *RiskAlert) error

	// GetByID retrieves a risk alert by ID
	GetByID(ctx context.Context, id string) (*RiskAlert, error)

	// ListByAccountID retrieves alerts by account ID
	ListByAccountID(ctx context.Context, accountID string, limit, offset int) ([]*RiskAlert, error)

	// ListByUserID retrieves alerts by user ID
	ListByUserID(ctx context.Context, userID string, limit, offset int) ([]*RiskAlert, error)

	// ListByType retrieves alerts by type
	ListByType(ctx context.Context, alertType AlertType, limit, offset int) ([]*RiskAlert, error)

	// ListBySeverity retrieves alerts by severity
	ListBySeverity(ctx context.Context, severity Severity, limit, offset int) ([]*RiskAlert, error)

	// ListActive retrieves all active alerts
	ListActive(ctx context.Context, limit, offset int) ([]*RiskAlert, error)

	// ListCritical retrieves all critical alerts
	ListCritical(ctx context.Context, limit int) ([]*RiskAlert, error)

	// Update updates a risk alert
	Update(ctx context.Context, alert *RiskAlert) error

	// Resolve marks an alert as resolved
	Resolve(ctx context.Context, id string) error

	// CountByStatus counts alerts by status
	CountByStatus(ctx context.Context, status AlertStatus) (int64, error)

	// CountBySeverity counts alerts by severity
	CountBySeverity(ctx context.Context, severity Severity) (int64, error)

	// DeleteOldResolved deletes resolved alerts older than specified duration
	DeleteOldResolved(ctx context.Context, olderThan time.Duration) error
}
