package app

import (
	"trading-stock/internal/application"
	"trading-stock/internal/infrastructure"
	infraAccount "trading-stock/internal/infrastructure/account"
	infraMarket "trading-stock/internal/infrastructure/market"
	infraMatching "trading-stock/internal/infrastructure/matching"
	infraOrder "trading-stock/internal/infrastructure/order"
	infraPortfolio "trading-stock/internal/infrastructure/portfolio"
	infraUser "trading-stock/internal/infrastructure/user"
	"trading-stock/internal/presentation/handler"
	"trading-stock/internal/presentation/router"
	"trading-stock/pkg/jwtservice"

	"trading-stock/internal/infrastructure/engine"
)

// wire builds the complete in-process dependency graph.
//
// ── Layer Order (bottom-up, strictly enforced) ───────────────────────────────
//
//	Infrastructure  → builds concrete implementations of domain interfaces
//	      ↓
//	Application     → receives domain interfaces; zero infra imports
//	      ↓
//	Presentation    → receives application use cases; zero domain/infra imports
//
// ── DI Rule ──────────────────────────────────────────────────────────────────
//
//	wire.go is the ONLY file allowed to import infrastructure packages.
//	All other layers communicate exclusively through interfaces.
func (a *App) wire() error {

	// ── 1. Infrastructure Layer ───────────────────────────────────────
	a.Repositories = infrastructure.NewRepositories(a.DB)
	a.Logger.Info("[ Infrastructure ] Repositories initialised")

	// Shared technical utilities (cross-cutting, NOT domain concepts)
	hasher := infraUser.NewBcryptHasher()

	a.JWTService = jwtservice.New(jwtservice.Config{
		AccessSecret:  a.Config.JWT.AccessSecret,
		RefreshSecret: a.Config.JWT.RefreshSecret,
		AccessTTL:     a.Config.JWT.AccessTTL,
		RefreshTTL:    a.Config.JWT.RefreshTTL,
		Issuer:        a.Config.JWT.Issuer,
	})
	a.Logger.Info("[ Infrastructure ] JWT service initialised")

	// ── 1b. Account Event Sourcing infrastructure ─────────────────────
	accountEventStore := infraAccount.NewEventStore(a.DB)
	accountEventSvc := infraAccount.NewEventSourcingService(
		accountEventStore,
		a.Logger,
	)
	a.Logger.Info("[ Infrastructure ] Account EventSourcing service initialised")

	// ── 1c. Account Projector (Kafka consumer → read model) ───────────
	a.AccountProjector = infraAccount.NewProjector(
		a.Config.Kafka.Brokers,
		a.Repositories.AccountReadModelRepo,
		a.Logger,
	)
	a.Logger.Info("[ Infrastructure ] Account Projector initialised")

	// ── 1d. Order Event Sourcing infrastructure ────────────────────────
	orderEventStore := infraOrder.NewEventStore(a.DB)
	orderEventSvc := infraOrder.NewEventSourcingService(
		orderEventStore,
		a.Logger,
	)
	a.Logger.Info("[ Infrastructure ] Order EventSourcing service initialised")

	// ── 1e. Order Projector (Kafka consumer → read model) ─────────────
	a.OrderProjector = infraOrder.NewOrderProjector(
		a.Config.Kafka.Brokers,
		a.Repositories.OrderReadModelRepo,
		a.Logger,
	)
	a.Logger.Info("[ Infrastructure ] Order Projector initialised")

	// ── 1f. Outbox relay ───────────────────────────────────────────────
	// The relay polls outbox_events and pushes pending rows to Kafka so
	// downstream consumers (matching engine, order fill, account, market)
	// receive their messages reliably.
	// a.OutboxRelay = infraOutbox.NewOutboxRelay(a.DB, a.Kafka, a.Logger)
	// a.Logger.Info("[ Infrastructure ] Outbox relay initialised")

	// ── 1g. Matching engine + consumer ────────────────────────────────
	matchingEngine := engine.NewMatchingEngine(engine.MatchingEngineConfig{
		Logger:            a.Logger,
		TradeChannelSize:  2000,
		UpdateChannelSize: 2000,
	})
	a.MatchingEngine = matchingEngine
	if a.Kafka != nil {
		a.EventPublisher = engine.NewEventPublisher(a.Kafka, a.Logger)
		a.Logger.Info("[ Infrastructure ] Matching event publisher initialised")
	} else {
		a.Logger.Warn("[ Infrastructure ] Kafka writer unavailable - matching event publisher disabled")
	}
	matchingSvc := infraMatching.NewMatchingService(matchingEngine, a.DB, a.Logger)
	a.MatchingConsumer = infraMatching.NewMatchingConsumer(a.Config.Kafka.Brokers, matchingSvc, a.Logger)
	a.Logger.Info("[ Infrastructure ] Matching engine + consumer initialised")

	// ── 1h. Order fill consumer ────────────────────────────────────────
	a.OrderFillConsumer = infraOrder.NewOrderFillConsumer(
		a.Config.Kafka.Brokers,
		orderEventSvc,
		a.Logger,
	)
	a.Logger.Info("[ Infrastructure ] Order fill consumer initialised")

	// ── 1h-1. Order update consumer ─────────────────────────────────────
	a.OrderUpdatedConsumer = infraOrder.NewOrderUpdatedConsumer(
		a.Config.Kafka.Brokers,
		orderEventSvc,
		a.Logger,
	)
	a.Logger.Info("[ Infrastructure ] Order updated consumer initialised")

	// ── 1h-2. Market order expire consumer ─────────────────────────────
	a.MarketExpireConsumer = infraOrder.NewMarketExpireConsumer(
		a.Config.Kafka.Brokers,
		orderEventSvc,
		a.Logger,
	)
	a.Logger.Info("[ Infrastructure ] Market expire consumer initialised")

	// ── 1i. Account trade settlement consumer ─────────────────────────
	a.AccountTradeConsumer = infraAccount.NewTradeConsumer(
		a.Config.Kafka.Brokers,
		accountEventSvc,
		a.Repositories.AccountReadModelRepo,
		a.Logger,
	)
	a.Logger.Info("[ Infrastructure ] Account trade consumer initialised")

	// ── 1i-2. Account order update consumer ─────────────────────────────
	a.AccountOrderUpdatedConsumer = infraAccount.NewOrderUpdatedConsumer(
		a.Config.Kafka.Brokers,
		accountEventSvc,
		a.Logger,
	)
	a.Logger.Info("[ Infrastructure ] Account order-updated consumer initialised")

	// ── 1k. Portfolio trade consumer ───────────────────────────────────
	a.PortfolioTradeConsumer = infraPortfolio.NewTradeConsumer(
		a.Config.Kafka.Brokers,
		a.Repositories.Portfolio,
		a.Logger,
	)
	a.Logger.Info("[ Infrastructure ] Portfolio trade consumer initialised")

	// ── 1j. Market data consumer ───────────────────────────────────────
	a.MarketTradeConsumer = infraMarket.NewMarketTradeConsumer(
		a.Config.Kafka.Brokers,
		a.Repositories.Price,
		a.Repositories.Candle,
		a.Logger,
	)
	a.Logger.Info("[ Infrastructure ] Market data consumer initialised")

	// ── 2. Application Layer ───────────────────────────────────────────
	a.Usecases = application.NewUsecases(
		a.Repositories,
		hasher,
		a.JWTService,
		a.Redis,
		accountEventSvc,
		orderEventSvc,
		a.Logger,
	)
	a.Logger.Info("[ Application ] Use cases initialised")

	// ── 3. Presentation Layer ─────────────────────────────────────────
	a.Handlers = handler.NewHandlerGroup(a.Usecases, a.Logger)
	a.Logger.Info("[ Presentation ] Handlers initialised")

	router.RegisterRoutes(a.Echo, a.Handlers, a.JWTService)
	a.Logger.Info("[ Presentation ] Routes registered")

	return nil
}
