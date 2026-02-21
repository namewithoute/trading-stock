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

	// ── Account Context – Data Repositories ──────────────────────────
	// NOTE: Account command/query is implemented via Event Sourcing.
	// The legacy CRUD repository is kept for backward compatibility
	// with Order UseCase that does a balance pre-check.
	Account              account.Repository
	AccountEventStore    account.EventStore
	AccountReadModelRepo account.ReadModelRepository

	// ── Order Context ─────────────────────────────────────────────────
	Order order.Repository

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
