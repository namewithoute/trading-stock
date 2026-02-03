package infra

import (
	"trading-stock/internal/domain"

	implAccount "trading-stock/internal/infra/account"
	implExecution "trading-stock/internal/infra/execution"
	implMarket "trading-stock/internal/infra/market"
	implOrder "trading-stock/internal/infra/order"
	implPortfolio "trading-stock/internal/infra/portfolio"
	implRisk "trading-stock/internal/infra/risk"
	implUser "trading-stock/internal/infra/user"

	"gorm.io/gorm"
)

// NewRepositories initializes all repository implementations
func NewRepositories(db *gorm.DB) *domain.Repositories {
	return &domain.Repositories{
		User:        implUser.NewUserRepository(db),
		Account:     implAccount.NewAccountRepository(db),
		Order:       implOrder.NewOrderRepository(db),
		Portfolio:   implPortfolio.NewPortfolioRepository(db),
		Stock:       implMarket.NewStockRepository(db),
		Price:       implMarket.NewPriceRepository(db),
		Candle:      implMarket.NewCandleRepository(db),
		Trade:       implExecution.NewTradeRepository(db),
		Settlement:  implExecution.NewSettlementRepository(db),
		Clearing:    implExecution.NewClearingRepository(db),
		RiskLimit:   implRisk.NewRiskLimitRepository(db),
		RiskMetrics: implRisk.NewRiskMetricsRepository(db),
		RiskAlert:   implRisk.NewRiskAlertRepository(db),
	}
}
