package risk

import (
	"context"
	"time"

	domain "trading-stock/internal/domain/risk"

	"gorm.io/gorm"
)

// riskLimitRepository implements domain.RiskLimitRepository
type riskLimitRepository struct {
	db *gorm.DB
}

func NewRiskLimitRepository(db *gorm.DB) domain.RiskLimitRepository {
	return &riskLimitRepository{db: db}
}

func (r *riskLimitRepository) Create(ctx context.Context, rl *domain.RiskLimit) error {
	return r.db.WithContext(ctx).Create(toRiskLimitModel(rl)).Error
}

func (r *riskLimitRepository) GetByID(ctx context.Context, id string) (*domain.RiskLimit, error) {
	var rl RiskLimitModel
	err := r.db.WithContext(ctx).First(&rl, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return rl.toDomain(), nil
}

func (r *riskLimitRepository) GetByAccountID(ctx context.Context, accountID string) (*domain.RiskLimit, error) {
	var rl RiskLimitModel
	err := r.db.WithContext(ctx).Where("account_id = ?", accountID).First(&rl).Error
	if err != nil {
		return nil, err
	}
	return rl.toDomain(), nil
}

func (r *riskLimitRepository) GetByUserID(ctx context.Context, userID string) (*domain.RiskLimit, error) {
	var rl RiskLimitModel
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&rl).Error
	if err != nil {
		return nil, err
	}
	return rl.toDomain(), nil
}

func (r *riskLimitRepository) Update(ctx context.Context, rl *domain.RiskLimit) error {
	return r.db.WithContext(ctx).Save(toRiskLimitModel(rl)).Error
}

func (r *riskLimitRepository) UpdateStatus(ctx context.Context, id string, status domain.LimitStatus) error {
	return r.db.WithContext(ctx).Model(&RiskLimitModel{}).Where("id = ?", id).Update("status", status).Error
}

func (r *riskLimitRepository) ListByStatus(ctx context.Context, status domain.LimitStatus, limit, offset int) ([]*domain.RiskLimit, error) {
	var models []*RiskLimitModel
	err := r.db.WithContext(ctx).Where("status = ?", status).Limit(limit).Offset(offset).Find(&models).Error
	if err != nil {
		return nil, err
	}
	limits := make([]*domain.RiskLimit, 0, len(models))
	for _, m := range models {
		limits = append(limits, m.toDomain())
	}
	return limits, nil
}

func (r *riskLimitRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&RiskLimitModel{}, "id = ?", id).Error
}

// riskMetricsRepository implements domain.RiskMetricsRepository
type riskMetricsRepository struct {
	db *gorm.DB
}

func NewRiskMetricsRepository(db *gorm.DB) domain.RiskMetricsRepository {
	return &riskMetricsRepository{db: db}
}

func (r *riskMetricsRepository) Create(ctx context.Context, rm *domain.RiskMetrics) error {
	return r.db.WithContext(ctx).Create(toRiskMetricsModel(rm)).Error
}

func (r *riskMetricsRepository) GetByID(ctx context.Context, id string) (*domain.RiskMetrics, error) {
	var rm RiskMetricsModel
	err := r.db.WithContext(ctx).First(&rm, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return rm.toDomain(), nil
}

func (r *riskMetricsRepository) GetByAccountID(ctx context.Context, accountID string) (*domain.RiskMetrics, error) {
	var rm RiskMetricsModel
	err := r.db.WithContext(ctx).Where("account_id = ?", accountID).First(&rm).Error
	if err != nil {
		return nil, err
	}
	return rm.toDomain(), nil
}

func (r *riskMetricsRepository) Update(ctx context.Context, rm *domain.RiskMetrics) error {
	return r.db.WithContext(ctx).Save(toRiskMetricsModel(rm)).Error
}

func (r *riskMetricsRepository) UpdatePnL(ctx context.Context, accountID string, dailyPnL, weeklyPnL, monthlyPnL float64) error {
	return r.db.WithContext(ctx).Model(&RiskMetricsModel{}).
		Where("account_id = ?", accountID).
		Updates(map[string]interface{}{
			"daily_pnl":   dailyPnL,
			"weekly_pnl":  weeklyPnL,
			"monthly_pnl": monthlyPnL,
			"updated_at":  time.Now(),
		}).Error
}

func (r *riskMetricsRepository) UpdateExposure(ctx context.Context, accountID string, total, long, short float64) error {
	return r.db.WithContext(ctx).Model(&RiskMetricsModel{}).
		Where("account_id = ?", accountID).
		Updates(map[string]interface{}{
			"total_exposure": total,
			"long_exposure":  long,
			"short_exposure": short,
			"updated_at":     time.Now(),
		}).Error
}

func (r *riskMetricsRepository) IncrementDailyOrders(ctx context.Context, accountID string) error {
	return r.db.WithContext(ctx).Model(&RiskMetricsModel{}).
		Where("account_id = ?", accountID).
		UpdateColumn("daily_orders_count", gorm.Expr("daily_orders_count + ?", 1)).Error
}

func (r *riskMetricsRepository) ResetDailyMetrics(ctx context.Context) error {
	return r.db.WithContext(ctx).Model(&RiskMetricsModel{}).
		Updates(map[string]interface{}{
			"daily_pnl":          0,
			"daily_orders_count": 0,
			"updated_at":         time.Now(),
		}).Error
}

func (r *riskMetricsRepository) ListHighRisk(ctx context.Context, minScore int, limit int) ([]*domain.RiskMetrics, error) {
	var models []*RiskMetricsModel
	err := r.db.WithContext(ctx).Where("risk_score >= ?", minScore).Limit(limit).Find(&models).Error
	if err != nil {
		return nil, err
	}
	metrics := make([]*domain.RiskMetrics, 0, len(models))
	for _, m := range models {
		metrics = append(metrics, m.toDomain())
	}
	return metrics, nil
}

func (r *riskMetricsRepository) GetAggregateMetrics(ctx context.Context) (map[string]interface{}, error) {
	var stats struct {
		TotalExposure  float64 `gorm:"column:total"`
		TotalPnL       float64 `gorm:"column:pnl"`
		ActiveAccounts int64   `gorm:"column:count"`
	}
	err := r.db.WithContext(ctx).Model(&RiskMetricsModel{}).
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

func NewRiskAlertRepository(db *gorm.DB) domain.RiskAlertRepository {
	return &riskAlertRepository{db: db}
}

func (r *riskAlertRepository) Create(ctx context.Context, ra *domain.RiskAlert) error {
	return r.db.WithContext(ctx).Create(toRiskAlertModel(ra)).Error
}

func (r *riskAlertRepository) GetByID(ctx context.Context, id string) (*domain.RiskAlert, error) {
	var ra RiskAlertModel
	err := r.db.WithContext(ctx).First(&ra, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return ra.toDomain(), nil
}

func (r *riskAlertRepository) ListByAccountID(ctx context.Context, accountID string, limit, offset int) ([]*domain.RiskAlert, error) {
	var models []*RiskAlertModel
	err := r.db.WithContext(ctx).Where("account_id = ?", accountID).Limit(limit).Offset(offset).Order("created_at DESC").Find(&models).Error
	if err != nil {
		return nil, err
	}
	alerts := make([]*domain.RiskAlert, 0, len(models))
	for _, m := range models {
		alerts = append(alerts, m.toDomain())
	}
	return alerts, nil
}

func (r *riskAlertRepository) ListByUserID(ctx context.Context, userID string, limit, offset int) ([]*domain.RiskAlert, error) {
	var models []*RiskAlertModel
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Limit(limit).Offset(offset).Order("created_at DESC").Find(&models).Error
	if err != nil {
		return nil, err
	}
	alerts := make([]*domain.RiskAlert, 0, len(models))
	for _, m := range models {
		alerts = append(alerts, m.toDomain())
	}
	return alerts, nil
}

func (r *riskAlertRepository) ListByType(ctx context.Context, alertType domain.AlertType, limit, offset int) ([]*domain.RiskAlert, error) {
	var models []*RiskAlertModel
	err := r.db.WithContext(ctx).Where("alert_type = ?", alertType).Limit(limit).Offset(offset).Order("created_at DESC").Find(&models).Error
	if err != nil {
		return nil, err
	}
	alerts := make([]*domain.RiskAlert, 0, len(models))
	for _, m := range models {
		alerts = append(alerts, m.toDomain())
	}
	return alerts, nil
}

func (r *riskAlertRepository) ListBySeverity(ctx context.Context, severity domain.Severity, limit, offset int) ([]*domain.RiskAlert, error) {
	var models []*RiskAlertModel
	err := r.db.WithContext(ctx).Where("severity = ?", severity).Limit(limit).Offset(offset).Order("created_at DESC").Find(&models).Error
	if err != nil {
		return nil, err
	}
	alerts := make([]*domain.RiskAlert, 0, len(models))
	for _, m := range models {
		alerts = append(alerts, m.toDomain())
	}
	return alerts, nil
}

func (r *riskAlertRepository) ListActive(ctx context.Context, limit, offset int) ([]*domain.RiskAlert, error) {
	var models []*RiskAlertModel
	err := r.db.WithContext(ctx).Where("status = ?", domain.AlertStatusActive).Limit(limit).Offset(offset).Order("created_at DESC").Find(&models).Error
	if err != nil {
		return nil, err
	}
	alerts := make([]*domain.RiskAlert, 0, len(models))
	for _, m := range models {
		alerts = append(alerts, m.toDomain())
	}
	return alerts, nil
}

func (r *riskAlertRepository) ListCritical(ctx context.Context, limit int) ([]*domain.RiskAlert, error) {
	var models []*RiskAlertModel
	err := r.db.WithContext(ctx).Where("severity = ?", domain.SeverityCritical).Limit(limit).Order("created_at DESC").Find(&models).Error
	if err != nil {
		return nil, err
	}
	alerts := make([]*domain.RiskAlert, 0, len(models))
	for _, m := range models {
		alerts = append(alerts, m.toDomain())
	}
	return alerts, nil
}

func (r *riskAlertRepository) Update(ctx context.Context, ra *domain.RiskAlert) error {
	return r.db.WithContext(ctx).Save(toRiskAlertModel(ra)).Error
}

func (r *riskAlertRepository) Resolve(ctx context.Context, id string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&RiskAlertModel{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":      domain.AlertStatusResolved,
			"resolved_at": &now,
		}).Error
}

func (r *riskAlertRepository) CountByStatus(ctx context.Context, status domain.AlertStatus) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&RiskAlertModel{}).Where("status = ?", status).Count(&count).Error
	return count, err
}

func (r *riskAlertRepository) CountBySeverity(ctx context.Context, severity domain.Severity) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&RiskAlertModel{}).Where("severity = ?", severity).Count(&count).Error
	return count, err
}

func (r *riskAlertRepository) DeleteOldResolved(ctx context.Context, olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)
	return r.db.WithContext(ctx).Where("status = ? AND resolved_at < ?", domain.AlertStatusResolved, cutoff).Delete(&RiskAlertModel{}).Error
}
