package application

import (
	"trading-stock/internal/application/account"
	"trading-stock/internal/application/admin"
	"trading-stock/internal/application/auth"
	"trading-stock/internal/application/execution"
	"trading-stock/internal/application/market"
	"trading-stock/internal/application/order"
	"trading-stock/internal/application/portfolio"
	"trading-stock/internal/application/risk"
	userUC "trading-stock/internal/application/user"
	"trading-stock/internal/domain"
	domainAccount "trading-stock/internal/domain/account"
	domainOrder "trading-stock/internal/domain/order"
	"trading-stock/internal/domain/user"
	"trading-stock/pkg/jwtservice"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Usecases aggregates all application use cases (Application Layer).
type Usecases struct {
	Auth      auth.UseCase
	User      userUC.UseCase
	Account   account.UseCase
	Order     order.UseCase
	Portfolio portfolio.UseCase
	Market    market.UseCase
	Trade     execution.UseCase
	Admin     admin.UseCase
	Risk      risk.UseCase
}

// NewUsecases constructs all use cases.
//
// ── Dependency Inversion Principle ─────────────────────────────────────
//
//	This function accepts ONLY interfaces defined in the Domain Layer.
//	It must NEVER import or reference Infrastructure packages directly.
//
// ── Parameters ──────────────────────────────────────────────
//
//	repos        – all domain repository interfaces (from infrastructure layer)
//	hasher       – password hashing port (domain/user.PasswordHasher)
//	jwtSvc       – JWT utility (shared pkg; no domain concepts leak)
//	redis        – Redis client (technical utility, not a domain concept)
//	accountRepo  – Event Sourcing port for Account (domain.Repository)
//	orderRepo    – Event Sourcing port for Order (domain.Repository)
//	logger       – structured logger
func NewUsecases(
	repos *domain.Repositories,
	hasher user.PasswordHasher,
	jwtSvc jwtservice.Service,
	redis *redis.Client,
	// Account Event Sourcing port – injected as domain interface, built in wire.go
	accountRepo domainAccount.Repository,
	// Order Event Sourcing port – injected as domain interface, built in wire.go
	orderRepo domainOrder.Repository,
	logger *zap.Logger,
) *Usecases {
	// Account use case built once and shared with Order (ReserveFunds / ReleaseFunds).
	accountUC := account.NewUseCase(
		accountRepo,                // domain.Repository (infra implements via Event Sourcing)
		repos.AccountReadModelRepo, // ReadModelRepository (domain interface)
		logger,
	)

	// Order use case: write side via Event Sourcing, read side via ReadModelRepository.
	orderUC := order.NewUseCase(
		orderRepo,                // domain.Repository (ES write side)
		repos.OrderReadModelRepo, // ReadModelRepository (query side)
		accountUC,                // CQRS: ReserveFunds / ReleaseFunds
		logger,
	)

	return &Usecases{
		Auth: auth.NewUseCase(
			repos.User, hasher, jwtSvc, redis, logger,
		),
		User: userUC.NewUseCase(
			repos.User, logger,
		),

		// Account: write commands go through Event Sourcing port;
		//          reads go through the Read Model repository.
		Account: accountUC,

		// Order: write commands go through Event Sourcing port;
		//        reads go through the Read Model repository.
		Order: orderUC,

		Portfolio: portfolio.NewUseCase(
			repos.Portfolio, logger,
		),
		Market: market.NewUseCase(
			repos.Stock, repos.Price, repos.Candle, redis, logger,
		),
		Trade: execution.NewUseCase(
			repos.Trade, logger,
		),
		Admin: admin.NewUseCase(
			repos.User, repos.OrderReadModelRepo, logger,
		),
		Risk: risk.NewUseCase(
			repos.RiskLimit, repos.RiskMetrics, repos.RiskAlert, logger,
		),
	}
}
