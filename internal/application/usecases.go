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
	"trading-stock/internal/domain/user"
	"trading-stock/pkg/jwtservice"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
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
// ── Dependency Inversion Principle ─────────────────────────────────────────
//
//	This function accepts ONLY interfaces defined in the Domain Layer.
//	It must NEVER import or reference Infrastructure packages directly.
//
//	Concrete infra implementations (e.g. EventSourcingService) are built
//	in wire.go and passed in as their Domain interface types.
//
// ── Parameters ──────────────────────────────────────────────────────────────
//
//	repos           – all domain repository interfaces (from infrastructure layer)
//	hasher          – password hashing port (domain/user.PasswordHasher)
//	jwtSvc          – JWT utility (shared pkg; no domain concepts leak)
//	redis           – Redis client (technical utility, not a domain concept)
//	kafkaWriter     – Kafka writer (technical utility)
//	accountEventSvc – Event Sourcing port for Account (domain.EventSourcingServicePort)
//	logger          – structured logger
func NewUsecases(
	repos *domain.Repositories,
	hasher user.PasswordHasher,
	jwtSvc jwtservice.Service,
	redis *redis.Client,
	kafkaWriter *kafka.Writer,
	// Account Event Sourcing port – injected as a domain interface, built in wire.go
	accountRepo domainAccount.Repository,
	logger *zap.Logger,
) *Usecases {
	// Account use case built once and shared with Order (ReserveFunds / ReleaseFunds).
	accountUC := account.NewUseCase(
		accountRepo,                // domain.Repository (infra implements via Event Sourcing)
		repos.AccountReadModelRepo, // ReadModelRepository (domain interface)
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

		// Order: no longer uses the legacy account.Repository — ReserveFunds /
		// ReleaseFunds are issued as CQRS commands through the account UseCase.
		Order: order.NewUseCase(
			repos.Order, accountUC, kafkaWriter, logger,
		),
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
			repos.User, repos.Order, logger,
		),
		Risk: risk.NewUseCase(
			repos.RiskLimit, repos.RiskMetrics, repos.RiskAlert, logger,
		),
	}
}
