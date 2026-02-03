package risk

import (
	"trading-stock/internal/domain/risk"

	"go.uber.org/zap"
)

// UseCase handles risk management business logic
type UseCase interface {
	// TODO: Add methods
}

type useCase struct {
	riskRepo risk.RiskLimitRepository
	logger   *zap.Logger
}

func NewUseCase(riskRepo risk.RiskLimitRepository, logger *zap.Logger) UseCase {
	return &useCase{riskRepo: riskRepo, logger: logger}
}
