package infrastructure

import (
	"trading-stock/internal/domain"

	implAccount "trading-stock/internal/infrastructure/account"
	implExecution "trading-stock/internal/infrastructure/execution"
	implMarket "trading-stock/internal/infrastructure/market"
	implOrder "trading-stock/internal/infrastructure/order"
	implPortfolio "trading-stock/internal/infrastructure/portfolio"
	implRisk "trading-stock/internal/infrastructure/risk"
	implUser "trading-stock/internal/infrastructure/user"

	"gorm.io/gorm"
)

// NewRepositories initializes all repository implementations
func NewRepositories(db *gorm.DB) *domain.Repositories {
	return &domain.Repositories{
		User:                 implUser.NewUserRepository(db),
		AccountReadModelRepo: implAccount.NewReadModelRepository(db),
		Order:                implOrder.NewOrderRepository(db),
		Portfolio:            implPortfolio.NewPortfolioRepository(db),
		Stock:                implMarket.NewStockRepository(db),
		Price:                implMarket.NewPriceRepository(db),
		Candle:               implMarket.NewCandleRepository(db),
		Trade:                implExecution.NewTradeRepository(db),
		Settlement:           implExecution.NewSettlementRepository(db),
		Clearing:             implExecution.NewClearingRepository(db),
		RiskLimit:            implRisk.NewRiskLimitRepository(db),
		RiskMetrics:          implRisk.NewRiskMetricsRepository(db),
		RiskAlert:            implRisk.NewRiskAlertRepository(db),
	}
}
