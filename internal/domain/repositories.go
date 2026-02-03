package domain

import (
	"trading-stock/internal/domain/account"
	"trading-stock/internal/domain/execution"
	"trading-stock/internal/domain/market"
	"trading-stock/internal/domain/order"
	"trading-stock/internal/domain/portfolio"
	"trading-stock/internal/domain/risk"
	"trading-stock/internal/domain/user"
)

// Repositories groups all repository interfaces for dependency injection
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
