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

// NewUsecases wires all use cases with their required dependencies.
//   - hasher     ? implements domain/user.PasswordHasher (lives in infra/user)
//   - jwtService ? pkg/jwtservice.Service (a shared technical utility, not a domain)
func NewUsecases(
	repos *domain.Repositories,
	hasher user.PasswordHasher,
	jwtSvc jwtservice.Service,
	redis *redis.Client,
	kafka *kafka.Writer,
	logger *zap.Logger,
) *Usecases {
	return &Usecases{
		Auth:      auth.NewUseCase(repos.User, hasher, jwtSvc, redis, logger),
		User:      userUC.NewUseCase(repos.User, logger),
		Account:   account.NewUseCase(repos.Account, logger),
		Order:     order.NewUseCase(repos.Order, repos.Account, kafka, logger),
		Portfolio: portfolio.NewUseCase(repos.Portfolio, logger),
		Market:    market.NewUseCase(repos.Stock, repos.Price, repos.Candle, redis, logger),
		Trade:     execution.NewUseCase(repos.Trade, logger),
		Admin:     admin.NewUseCase(repos.User, repos.Order, logger),
		Risk:      risk.NewUseCase(repos.RiskLimit, repos.RiskMetrics, repos.RiskAlert, logger),
	}
}
