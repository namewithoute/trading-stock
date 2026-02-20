package app

import (
	"trading-stock/internal/application"
	"trading-stock/internal/infrastructure"
	infraUser "trading-stock/internal/infrastructure/user"
	"trading-stock/internal/presentation/handler"
	"trading-stock/internal/presentation/router"
	"trading-stock/pkg/jwtservice"
)

// wire builds the entire in-process dependency graph.
//
// Strict bottom-up DDD order:
//
//	Infrastructure Layer (Repositories)
//	↓
//	Application Layer   (Use Cases / Services)
//	↓
//	Presentation Layer  (Handlers + Routes)
func (a *App) wire() error {
	// ── Infrastructure Layer ──────────────────────────────────────────
	a.Repositories = infrastructure.NewRepositories(a.DB)
	a.Logger.Info("[ Infrastructure ] Repositories initialized")

	// ── Shared Technical Services ─────────────────────────────────────
	// These are cross-cutting utilities, not domain concepts.
	hasher := infraUser.NewBcryptHasher()

	a.JWTService = jwtservice.New(jwtservice.Config{
		AccessSecret:  a.Config.JWT.AccessSecret,
		RefreshSecret: a.Config.JWT.RefreshSecret,
		AccessTTL:     a.Config.JWT.AccessTTL,
		RefreshTTL:    a.Config.JWT.RefreshTTL,
		Issuer:        a.Config.JWT.Issuer,
	})
	a.Logger.Info("[ Infrastructure ] JWT service initialized")

	// ── Application Layer ─────────────────────────────────────────────
	a.Usecases = application.NewUsecases(
		a.Repositories,
		hasher,
		a.JWTService,
		a.Redis,
		a.Kafka,
		a.Logger,
	)
	a.Logger.Info("[ Application ] Use cases initialized")

	// ── Presentation Layer ────────────────────────────────────────────
	a.Handlers = handler.NewHandlerGroup(a.Usecases)
	a.Logger.Info("[ Presentation ] Handlers initialized")

	router.RegisterRoutes(a.Echo, a.Handlers, a.JWTService)
	a.Logger.Info("[ Presentation ] Routes registered")

	return nil
}
