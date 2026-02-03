package usecase

import (
	"trading-stock/internal/infra"
	"trading-stock/internal/usecase/account"
	"trading-stock/internal/usecase/admin"
	"trading-stock/internal/usecase/auth"
	"trading-stock/internal/usecase/execution"
	"trading-stock/internal/usecase/market"
	"trading-stock/internal/usecase/order"
	"trading-stock/internal/usecase/portfolio"
	"trading-stock/internal/usecase/user"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// Services groups all use cases together
type Usecases struct {
	Auth      auth.UseCase
	User      user.UseCase
	Account   account.UseCase
	Order     order.UseCase
	Portfolio portfolio.UseCase
	Market    market.UseCase
	Trade     execution.UseCase
	Admin     admin.UseCase
}

// NewServices creates all use cases with dependencies
func NewUsecases(
	repos *infra.Repositories,
	redis *redis.Client,
	kafka *kafka.Writer,
	logger *zap.Logger,
) *Usecases {
	return &Usecases{
		Auth:      auth.NewUseCase(repos.User, redis, logger),
		User:      user.NewUseCase(repos.User, logger),
		Account:   account.NewUseCase(repos.Account, logger),
		Order:     order.NewUseCase(repos.Order, kafka, logger),
		Portfolio: portfolio.NewUseCase(repos.Portfolio, logger),
		Market:    market.NewUseCase(repos.Stock, repos.Price, repos.Candle, redis, logger),
		Trade:     execution.NewUseCase(repos.Trade, logger),
		Admin:     admin.NewUseCase(repos.User, repos.Order, logger),
	}
}
