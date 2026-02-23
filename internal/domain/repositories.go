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

// Repositories groups all domain repository interfaces.
//
// IMPORTANT: This struct contains ONLY repository interfaces
// (data access contracts defined in the Domain layer).
//
// Infrastructure implementations are injected by wire.go.
// Cross-cutting services (EventSourcingServicePort, KafkaPublisher, etc.)
// do NOT belong here – they are injected directly into the Application layer
// via dedicated dependency structs.
type Repositories struct {
	// ── User Context ──────────────────────────────────────────────────
	User user.Repository

	// ── Account Context ────────────────────────────────────────────────
	// Write side: account.Repository is injected directly in wire.go
	// (it depends on Kafka, so it cannot be wired in NewRepositories).
	// Query (read) side served by ReadModelRepository.
	AccountReadModelRepo account.ReadModelRepository

	// ── Order Context ─────────────────────────────────────────────────
	// Write side: order.Repository (EventSourcingService) is injected directly
	// in wire.go (depends on Kafka so cannot be wired in NewRepositories).
	// Query (read) side served by ReadModelRepository.
	OrderReadModelRepo order.ReadModelRepository

	// ── Portfolio Context ─────────────────────────────────────────────
	Portfolio portfolio.Repository

	// ── Market Context ────────────────────────────────────────────────
	Stock  market.StockRepository
	Price  market.PriceRepository
	Candle market.CandleRepository

	// ── Execution Context ─────────────────────────────────────────────
	Trade      execution.TradeRepository
	Settlement execution.SettlementRepository
	Clearing   execution.ClearingRepository

	// ── Risk Context ──────────────────────────────────────────────────
	RiskLimit   risk.RiskLimitRepository
	RiskMetrics risk.RiskMetricsRepository
	RiskAlert   risk.RiskAlertRepository
}
