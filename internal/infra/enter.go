package infra

import (
	"trading-stock/internal/domain/account"
	"trading-stock/internal/domain/execution"
	"trading-stock/internal/domain/market"
	"trading-stock/internal/domain/order"
	"trading-stock/internal/domain/portfolio"
	"trading-stock/internal/domain/risk"
	"trading-stock/internal/domain/user"

	implAccount "trading-stock/internal/infra/account"
	implExecution "trading-stock/internal/infra/execution"
	implMarket "trading-stock/internal/infra/market"
	implOrder "trading-stock/internal/infra/order"
	implPortfolio "trading-stock/internal/infra/portfolio"
	implRisk "trading-stock/internal/infra/risk"
	implUser "trading-stock/internal/infra/user"

	"gorm.io/gorm"
)

// Repositories groups all repositories together for dependency injection
type Repositories struct {
	User        user.Repository
	Account     account.Repository
	Order       order.Repository
	Portfolio   portfolio.Repository
	Stock       market.StockRepository
	Price       market.PriceRepository
	Candle      market.CandleRepository
	Trade       execution.TradeRepository
	Settlement  execution.SettlementRepository
	Clearing    execution.ClearingRepository
	RiskLimit   risk.RiskLimitRepository
	RiskMetrics risk.RiskMetricsRepository
	RiskAlert   risk.RiskAlertRepository
}

// NewRepositories initializes all repository implementations
func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
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
