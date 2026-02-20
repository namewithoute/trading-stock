package risk

import (
	"context"
	"trading-stock/internal/domain/risk"

	"go.uber.org/zap"
)

// UseCase handles risk management business logic
type UseCase interface {
	GetRiskMetrics(ctx context.Context, accountID string) (*risk.RiskMetrics, error)
	GetActiveAlerts(ctx context.Context, accountID string) ([]*risk.RiskAlert, error)
	CheckOrderRisk(ctx context.Context, accountID string, symbol string, amount float64) (bool, error)
}

type useCase struct {
	limitRepo   risk.RiskLimitRepository
	metricsRepo risk.RiskMetricsRepository
	alertRepo   risk.RiskAlertRepository
	logger      *zap.Logger
}

func NewUseCase(
	limitRepo risk.RiskLimitRepository,
	metricsRepo risk.RiskMetricsRepository,
	alertRepo risk.RiskAlertRepository,
	logger *zap.Logger,
) UseCase {
	return &useCase{
		limitRepo:   limitRepo,
		metricsRepo: metricsRepo,
		alertRepo:   alertRepo,
		logger:      logger,
	}
}

func (s *useCase) GetRiskMetrics(ctx context.Context, accountID string) (*risk.RiskMetrics, error) {
	return s.metricsRepo.GetByAccountID(ctx, accountID)
}

func (s *useCase) GetActiveAlerts(ctx context.Context, accountID string) ([]*risk.RiskAlert, error) {
	return s.alertRepo.ListByAccountID(ctx, accountID, 10, 0)
}

func (s *useCase) CheckOrderRisk(ctx context.Context, accountID string, symbol string, amount float64) (bool, error) {
	// TODO: Implement risk check logic (e.g. against limits)
	return true, nil
}
