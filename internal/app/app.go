package app

import (
	"context"
	"time"

	"trading-stock/internal/application"
	"trading-stock/internal/config"
	"trading-stock/internal/domain"
	infraAccount "trading-stock/internal/infrastructure/account"
	infraMarket "trading-stock/internal/infrastructure/market"
	infraMatching "trading-stock/internal/infrastructure/matching"
	infraOrder "trading-stock/internal/infrastructure/order"
	infraOutbox "trading-stock/internal/infrastructure/outbox"
	infraPortfolio "trading-stock/internal/infrastructure/portfolio"
	"trading-stock/internal/presentation/handler"
	"trading-stock/pkg/jwtservice"
	"trading-stock/pkg/logger"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// App is the top-level composition root.
// It owns all initialized dependencies and their lifecycle.
//
// Startup sequence:  New() → [config → logger → resources → wire → server]
// Shutdown sequence: Shutdown() → [http → workers → db → redis → kafka → logger.sync]
type App struct {
	Config *config.Config
	Logger *zap.Logger

	// External infrastructure connections
	DB    *gorm.DB
	Redis *redis.Client
	Kafka *kafka.Writer

	// HTTP server
	Echo *echo.Echo

	// DDD layers (wired in wire.go)
	Repositories *domain.Repositories
	Usecases     *application.Usecases
	Handlers     *handler.HandlerGroup
	JWTService   jwtservice.Service

	// ── Background workers ─────────────────────────────────────────────────
	// Projectors: consume domain event topics → maintain read models
	AccountProjector *infraAccount.Projector
	OrderProjector   *infraOrder.Projector

	// Outbox relay: polls outbox_events → publishes to Kafka
	OutboxRelay *infraOutbox.OutboxRelay

	// Matching: consumes orders.accepted → runs engine → writes trades + outbox
	MatchingConsumer *infraMatching.MatchingConsumer

	// Fill consumer: consumes trades.executed → updates order aggregate state
	OrderFillConsumer *infraOrder.OrderFillConsumer

	// Account settlement: consumes trades.executed → settles funds
	AccountTradeConsumer *infraAccount.TradeConsumer

	// Market data: consumes trades.executed → updates price + candle tables
	MarketTradeConsumer *infraMarket.MarketTradeConsumer

	// Portfolio: consumes trades.executed → updates stock positions
	PortfolioTradeConsumer *infraPortfolio.TradeConsumer

	// Single cancel for all background goroutines
	workerCancel context.CancelFunc
}

// New bootstraps the application in a strict, ordered sequence.
// Any step failure returns immediately with a wrapped error.
func New(ctx context.Context) (*App, error) {
	a := &App{}

	// 1. Config — must be first; all other steps read from it
	a.Config = config.Load()

	// 2. Logger — must exist before any error reporting
	if err := a.initLogger(); err != nil {
		return nil, err
	}

	// 3. External connections — enforce a hard startup deadline
	resourceCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := a.initResources(resourceCtx); err != nil {
		return nil, err
	}

	// 4. HTTP server — Echo must exist before routes are registered in wire()
	a.initHTTPServer()

	// 5. Dependency graph — wires repos, use cases, handlers, and routes onto a.Echo
	if err := a.wire(); err != nil {
		return nil, err
	}

	a.Logger.Info("Application ready to serve")
	return a, nil
}

// initLogger builds the structured logger from configuration.
// Uses fmt as fallback since a.Logger is not yet available.
func (a *App) initLogger() error {
	cfg := a.Config.Logger

	log, err := logger.InitLogger(logger.LoggerConfig{
		Level:         cfg.Level,
		Director:      cfg.Director,
		ShowLine:      cfg.ShowLine,
		StacktraceKey: cfg.StacktraceKey,
		LogInConsole:  cfg.LogInConsole,
		MaxSize:       cfg.MaxSize,
		MaxBackups:    cfg.MaxBackups,
		MaxAge:        cfg.MaxAge,
	})
	if err != nil {
		return err
	}

	a.Logger = log
	a.Logger.Info("Logger initialized", zap.String("level", cfg.Level))
	return nil
}
