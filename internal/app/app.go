package app

import (
	"context"
	"fmt"
	"time"

	"trading-stock/internal/config"
	"trading-stock/internal/domain"
	"trading-stock/internal/handler"
	"trading-stock/internal/infra"
	"trading-stock/internal/initialize"
	"trading-stock/internal/router"
	"trading-stock/internal/usecase"
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

	// Dependency injection
	Repositories *domain.Repositories
	Usecases     *usecase.Usecases
	Handlers     *handler.HandlerGroup
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

	if err := app.initResources(initCtx); err != nil {
		return nil, err
	}

	// Wire dependencies (repositories, services, handlers)
	if err := app.wireDependencies(); err != nil {
		return nil, err
	}

	// Initialize HTTP server (after dependencies are ready)
	app.initHTTPServer()

	router.RegisterRoutes(app.Echo, app.Handlers)

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

// initResources initializes all external dependencies
func (a *App) initResources(ctx context.Context) error {
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

// wireDependencies wires up all dependencies (repositories, services, handlers)
func (a *App) wireDependencies() error {
	// ============================================
	// REPOSITORIES
	// ============================================
	a.Repositories = infra.NewRepositories(a.DB)
	a.Logger.Info("Infrastructure (Repositories) initialized")

	// ============================================
	// USE CASES (SERVICES)
	// ============================================
	a.Usecases = usecase.NewUsecases(
		a.Repositories,
		a.Redis,
		a.Kafka,
		a.Logger,
	)
	a.Logger.Info("Use cases (Services) initialized")

	// ============================================
	// HANDLERS
	// ============================================
	a.Handlers = handler.NewHandlerGroup(a.Usecases)
	a.Logger.Info("Handlers initialized")

	a.Logger.Info("Dependencies wired successfully")
	return nil
}
