package risk

import (
	"context"
	"time"

	"trading-stock/internal/domain/risk"

	"gorm.io/gorm"
)

// riskLimitRepository implements domain.RiskLimitRepository
type riskLimitRepository struct {
	db *gorm.DB
}

func NewRiskLimitRepository(db *gorm.DB) risk.RiskLimitRepository {
	return &riskLimitRepository{db: db}
}

func (r *riskLimitRepository) Create(ctx context.Context, rl *risk.RiskLimit) error {
	return r.db.WithContext(ctx).Create(rl).Error
}

func (r *riskLimitRepository) GetByID(ctx context.Context, id string) (*risk.RiskLimit, error) {
	var rl risk.RiskLimit
	err := r.db.WithContext(ctx).First(&rl, "id = ?", id).Error
	return &rl, err
}

func (r *riskLimitRepository) GetByAccountID(ctx context.Context, accountID string) (*risk.RiskLimit, error) {
	var rl risk.RiskLimit
	err := r.db.WithContext(ctx).Where("account_id = ?", accountID).First(&rl).Error
	return &rl, err
}

func (r *riskLimitRepository) GetByUserID(ctx context.Context, userID string) (*risk.RiskLimit, error) {
	var rl risk.RiskLimit
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&rl).Error
	return &rl, err
}

func (r *riskLimitRepository) Update(ctx context.Context, rl *risk.RiskLimit) error {
	return r.db.WithContext(ctx).Save(rl).Error
}

func (r *riskLimitRepository) UpdateStatus(ctx context.Context, id string, status risk.LimitStatus) error {
	return r.db.WithContext(ctx).Model(&risk.RiskLimit{}).Where("id = ?", id).Update("status", status).Error
}

func (r *riskLimitRepository) ListByStatus(ctx context.Context, status risk.LimitStatus, limit, offset int) ([]*risk.RiskLimit, error) {
	var limits []*risk.RiskLimit
	err := r.db.WithContext(ctx).Where("status = ?", status).Limit(limit).Offset(offset).Find(&limits).Error
	return limits, err
}

func (r *riskLimitRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&risk.RiskLimit{}, "id = ?", id).Error
}

// riskMetricsRepository implements domain.RiskMetricsRepository
type riskMetricsRepository struct {
	db *gorm.DB
}

func NewRiskMetricsRepository(db *gorm.DB) risk.RiskMetricsRepository {
	return &riskMetricsRepository{db: db}
}

func (r *riskMetricsRepository) Create(ctx context.Context, rm *risk.RiskMetrics) error {
	return r.db.WithContext(ctx).Create(rm).Error
}

func (r *riskMetricsRepository) GetByID(ctx context.Context, id string) (*risk.RiskMetrics, error) {
	var rm risk.RiskMetrics
	err := r.db.WithContext(ctx).First(&rm, "id = ?", id).Error
	return &rm, err
}

func (r *riskMetricsRepository) GetByAccountID(ctx context.Context, accountID string) (*risk.RiskMetrics, error) {
	var rm risk.RiskMetrics
	err := r.db.WithContext(ctx).Where("account_id = ?", accountID).First(&rm).Error
	return &rm, err
}

func (r *riskMetricsRepository) Update(ctx context.Context, rm *risk.RiskMetrics) error {
	return r.db.WithContext(ctx).Save(rm).Error
}

func (r *riskMetricsRepository) UpdatePnL(ctx context.Context, accountID string, dailyPnL, weeklyPnL, monthlyPnL float64) error {
	return r.db.WithContext(ctx).Model(&risk.RiskMetrics{}).
		Where("account_id = ?", accountID).
		Updates(map[string]interface{}{
			"daily_pnl":   dailyPnL,
			"weekly_pnl":  weeklyPnL,
			"monthly_pnl": monthlyPnL,
			"updated_at":  time.Now(),
		}).Error
}

func (r *riskMetricsRepository) UpdateExposure(ctx context.Context, accountID string, total, long, short float64) error {
	return r.db.WithContext(ctx).Model(&risk.RiskMetrics{}).
		Where("account_id = ?", accountID).
		Updates(map[string]interface{}{
			"total_exposure": total,
			"long_exposure":  long,
			"short_exposure": short,
			"updated_at":     time.Now(),
		}).Error
}

func (r *riskMetricsRepository) IncrementDailyOrders(ctx context.Context, accountID string) error {
	return r.db.WithContext(ctx).Model(&risk.RiskMetrics{}).
		Where("account_id = ?", accountID).
		UpdateColumn("daily_orders_count", gorm.Expr("daily_orders_count + ?", 1)).Error
}

func (r *riskMetricsRepository) ResetDailyMetrics(ctx context.Context) error {
	return r.db.WithContext(ctx).Model(&risk.RiskMetrics{}).
		Updates(map[string]interface{}{
			"daily_pnl":          0,
			"daily_orders_count": 0,
			"updated_at":         time.Now(),
		}).Error
}

func (r *riskMetricsRepository) ListHighRisk(ctx context.Context, minScore int, limit int) ([]*risk.RiskMetrics, error) {
	var metrics []*risk.RiskMetrics
	err := r.db.WithContext(ctx).Where("risk_score >= ?", minScore).Limit(limit).Find(&metrics).Error
	return metrics, err
}

func (r *riskMetricsRepository) GetAggregateMetrics(ctx context.Context) (map[string]interface{}, error) {
	var stats struct {
		TotalExposure  float64 `gorm:"column:total"`
		TotalPnL       float64 `gorm:"column:pnl"`
		ActiveAccounts int64   `gorm:"column:count"`
	}
	err := r.db.WithContext(ctx).Model(&risk.RiskMetrics{}).
		Select("SUM(total_exposure) as total, SUM(daily_pnl) as pnl, COUNT(*) as count").
		Scan(&stats).Error

	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"total_exposure":  stats.TotalExposure,
		"total_daily_pnl": stats.TotalPnL,
		"active_accounts": stats.ActiveAccounts,
	}
	return result, nil
}

// riskAlertRepository implements domain.RiskAlertRepository
type riskAlertRepository struct {
	db *gorm.DB
}

func NewRiskAlertRepository(db *gorm.DB) risk.RiskAlertRepository {
	return &riskAlertRepository{db: db}
}

func (r *riskAlertRepository) Create(ctx context.Context, ra *risk.RiskAlert) error {
	return r.db.WithContext(ctx).Create(ra).Error
}

func (r *riskAlertRepository) GetByID(ctx context.Context, id string) (*risk.RiskAlert, error) {
	var ra risk.RiskAlert
	err := r.db.WithContext(ctx).First(&ra, "id = ?", id).Error
	return &ra, err
}

func (r *riskAlertRepository) ListByAccountID(ctx context.Context, accountID string, limit, offset int) ([]*risk.RiskAlert, error) {
	var alerts []*risk.RiskAlert
	err := r.db.WithContext(ctx).Where("account_id = ?", accountID).Limit(limit).Offset(offset).Order("created_at DESC").Find(&alerts).Error
	return alerts, err
}

func (r *riskAlertRepository) ListByUserID(ctx context.Context, userID string, limit, offset int) ([]*risk.RiskAlert, error) {
	var alerts []*risk.RiskAlert
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Limit(limit).Offset(offset).Order("created_at DESC").Find(&alerts).Error
	return alerts, err
}

func (r *riskAlertRepository) ListByType(ctx context.Context, alertType risk.AlertType, limit, offset int) ([]*risk.RiskAlert, error) {
	var alerts []*risk.RiskAlert
	err := r.db.WithContext(ctx).Where("alert_type = ?", alertType).Limit(limit).Offset(offset).Order("created_at DESC").Find(&alerts).Error
	return alerts, err
}

func (r *riskAlertRepository) ListBySeverity(ctx context.Context, severity risk.Severity, limit, offset int) ([]*risk.RiskAlert, error) {
	var alerts []*risk.RiskAlert
	err := r.db.WithContext(ctx).Where("severity = ?", severity).Limit(limit).Offset(offset).Order("created_at DESC").Find(&alerts).Error
	return alerts, err
}

func (r *riskAlertRepository) ListActive(ctx context.Context, limit, offset int) ([]*risk.RiskAlert, error) {
	var alerts []*risk.RiskAlert
	err := r.db.WithContext(ctx).Where("status = ?", risk.AlertStatusActive).Limit(limit).Offset(offset).Order("created_at DESC").Find(&alerts).Error
	return alerts, err
}

func (r *riskAlertRepository) ListCritical(ctx context.Context, limit int) ([]*risk.RiskAlert, error) {
	var alerts []*risk.RiskAlert
	err := r.db.WithContext(ctx).Where("severity = ?", risk.SeverityCritical).Limit(limit).Order("created_at DESC").Find(&alerts).Error
	return alerts, err
}

func (r *riskAlertRepository) Update(ctx context.Context, ra *risk.RiskAlert) error {
	return r.db.WithContext(ctx).Save(ra).Error
}

func (r *riskAlertRepository) Resolve(ctx context.Context, id string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&risk.RiskAlert{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":      risk.AlertStatusResolved,
			"resolved_at": &now,
		}).Error
}

func (r *riskAlertRepository) CountByStatus(ctx context.Context, status risk.AlertStatus) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&risk.RiskAlert{}).Where("status = ?", status).Count(&count).Error
	return count, err
}

func (r *riskAlertRepository) CountBySeverity(ctx context.Context, severity risk.Severity) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&risk.RiskAlert{}).Where("severity = ?", severity).Count(&count).Error
	return count, err
}

func (r *riskAlertRepository) DeleteOldResolved(ctx context.Context, olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)
	return r.db.WithContext(ctx).Where("status = ? AND resolved_at < ?", risk.AlertStatusResolved, cutoff).Delete(&risk.RiskAlert{}).Error
}
