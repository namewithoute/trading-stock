package app

import (
	"trading-stock/internal/application"
	"trading-stock/internal/infrastructure"
	infraAccount "trading-stock/internal/infrastructure/account"
	infraOrder "trading-stock/internal/infrastructure/order"
	infraUser "trading-stock/internal/infrastructure/user"
	"trading-stock/internal/presentation/handler"
	"trading-stock/internal/presentation/router"
	"trading-stock/pkg/jwtservice"
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
	// EventStore is wired here (needs DB) and shared between
	// EventSourcingService (write path) and Projector (rebuild path).
	accountEventStore := infraAccount.NewEventStore(a.DB)
	accountEventSvc := infraAccount.NewEventSourcingService(
		accountEventStore,
		a.Kafka,
		a.Logger,
	)
	a.Logger.Info("[ Infrastructure ] Account EventSourcing service initialised")

	// ── 1c. Account Projector (Kafka consumer → read model) ───────────
	// Build the Projector here so lifecycle.go can call Run(ctx) / stop.
	a.AccountProjector = infraAccount.NewProjector(
		a.Config.Kafka.Brokers,              // Kafka broker addresses from config
		a.Repositories.AccountReadModelRepo, // ReadModelRepository domain interface
		accountEventStore,                   // EventStore — used only during Rebuild()
		a.Logger,
	)
	a.Logger.Info("[ Infrastructure ] Account Projector initialised")

	// ── 1d. Order Event Sourcing infrastructure ─────────────────────────────
	orderEventStore := infraOrder.NewEventStore(a.DB)
	orderEventSvc := infraOrder.NewEventSourcingService(
		orderEventStore,
		a.Kafka,
		a.Logger,
	)
	a.Logger.Info("[ Infrastructure ] Order EventSourcing service initialised")

	// ── 1e. Order Projector (Kafka consumer → read model) ─────────────────
	a.OrderProjector = infraOrder.NewOrderProjector(
		a.Config.Kafka.Brokers,
		a.Repositories.OrderReadModelRepo,
		orderEventStore,
		a.Logger,
	)
	a.Logger.Info("[ Infrastructure ] Order Projector initialised")

	// ── 2. Application Layer ───────────────────────────────────────────
	// All dependencies injected as domain-defined interfaces.
	// NewUsecases has ZERO import from infrastructure packages.
	a.Usecases = application.NewUsecases(
		a.Repositories,
		hasher,
		a.JWTService,
		a.Redis,
		accountEventSvc, // domain.Repository for account (Event Sourcing)
		orderEventSvc,   // domain.Repository for order  (Event Sourcing)
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
