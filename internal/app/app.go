package app

import (
	"context"
	"fmt"
	"time"

	"trading-stock/internal/config"
	"trading-stock/internal/initialize"
	"trading-stock/pkg/logger"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// App is the main application container with all dependencies
type App struct {
	Config *config.Config
	Logger *zap.Logger
	DB     *gorm.DB
	Redis  *redis.Client
	Kafka  *kafka.Writer
	Echo   *echo.Echo

	// Use cases will be added here
	// UserUseCase     usecase.UserUseCase
	// OrderUseCase    usecase.OrderUseCase
	// PortfolioUseCase usecase.PortfolioUseCase
}

// New creates a new App instance with all dependencies initialized
func New(ctx context.Context) (*App, error) {
	app := &App{}

	// Load configuration
	app.Config = config.Load()

	// Initialize logger
	if err := app.initLogger(); err != nil {
		return nil, err
	}

	// Initialize infrastructure (DB, Redis, Kafka)
	initCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := app.initInfrastructure(initCtx); err != nil {
		return nil, err
	}

	// Wire dependencies (repositories, use cases, handlers)
	if err := app.wireDependencies(); err != nil {
		return nil, err
	}

	// Initialize HTTP server
	app.initHTTPServer()

	app.Logger.Info("Application initialized successfully")
	return app, nil
}

// initLogger initializes the logger
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
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	a.Logger = log
	return nil
}

// initInfrastructure initializes all external dependencies
func (a *App) initInfrastructure(ctx context.Context) error {
	var err error

	// Initialize PostgreSQL
	a.DB, err = initialize.InitPostgresDB(ctx, a.Config.Database, a.Logger)
	if err != nil {
		return fmt.Errorf("postgres initialization failed: %w", err)
	}

	// Run database migrations
	if err := initialize.AutoMigrateModels(a.DB, a.Logger); err != nil {
		return fmt.Errorf("database migration failed: %w", err)
	}

	a.Logger.Info("PostgreSQL connected and migrated successfully")

	// Initialize Redis
	a.Redis, err = initialize.InitRedis(ctx, a.Config.Redis, a.Logger)
	if err != nil {
		return fmt.Errorf("redis initialization failed: %w", err)
	}

	a.Logger.Info("Redis connected successfully")

	// Initialize Kafka
	a.Kafka, err = initialize.InitKafka(ctx, a.Config.Kafka, a.Logger)
	if err != nil {
		return fmt.Errorf("kafka initialization failed: %w", err)
	}

	a.Logger.Info("Kafka connected successfully")

	return nil
}

// wireDependencies wires up all dependencies (repositories, use cases, handlers)
func (a *App) wireDependencies() error {
	// TODO: Initialize repositories
	// userRepo := postgres.NewUserRepository(a.DB)
	// orderRepo := postgres.NewOrderRepository(a.DB)

	// TODO: Initialize use cases
	// a.UserUseCase = usecase.NewUserUseCase(userRepo, a.Logger)
	// a.OrderUseCase = usecase.NewOrderUseCase(orderRepo, a.Kafka, a.Logger)

	// TODO: Initialize handlers and register routes
	// handler.RegisterRoutes(a.Echo, userHandler, orderHandler)

	return nil
}
