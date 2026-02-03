package usecase

import (
	// Import các interface từ Domain layer

	// Import Use Case implementations
	"trading-stock/internal/domain"
	"trading-stock/internal/usecase/account"
	"trading-stock/internal/usecase/admin"
	"trading-stock/internal/usecase/auth"
	"trading-stock/internal/usecase/execution"
	"trading-stock/internal/usecase/market"
	"trading-stock/internal/usecase/order"
	"trading-stock/internal/usecase/portfolio"
	"trading-stock/internal/usecase/risk"
	"trading-stock/internal/usecase/user"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// Repositories struct này bây giờ chỉ chứa các Interface từ Domain.
// Nó giúp gom nhóm các phụ thuộc mà không cần biết triển khai cụ thể là gì.

type Usecases struct {
	Auth      auth.UseCase
	User      user.UseCase
	Account   account.UseCase
	Order     order.UseCase
	Portfolio portfolio.UseCase
	Market    market.UseCase
	Trade     execution.UseCase
	Admin     admin.UseCase
	Risk      risk.UseCase
}

// NewUsecases bây giờ nhận vào struct Repositories (local) chứa các Interface.
// Chú ý: Đã loại bỏ hoàn toàn sự phụ thuộc vào package 'infra'.
func NewUsecases(
	repos *domain.Repositories,
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
		Risk:      risk.NewUseCase(repos.RiskLimit, repos.RiskMetrics, repos.RiskAlert, logger),
	}
}
