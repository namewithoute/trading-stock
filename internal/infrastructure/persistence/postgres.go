package persistence

import (
	"context"
	"fmt"
	"time"

	"trading-stock/internal/config"
	infraAccount "trading-stock/internal/infrastructure/account"
	infraExecution "trading-stock/internal/infrastructure/execution"
	infraMarket "trading-stock/internal/infrastructure/market"
	infraOrder "trading-stock/internal/infrastructure/order"
	infraPortfolio "trading-stock/internal/infrastructure/portfolio"
	infraRisk "trading-stock/internal/infrastructure/risk"
	infraUser "trading-stock/internal/infrastructure/user"
	"trading-stock/pkg/utils"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitPostgresDB initializes PostgreSQL database connection with retry logic
// Returns *gorm.DB instance instead of using global variables
func InitPostgresDB(ctx context.Context, cfg config.DatabaseConfig, log *zap.Logger) (*gorm.DB, error) {
	var db *gorm.DB

	// Configure GORM logger
	gormLogger := logger.Default.LogMode(logger.Silent)
	if cfg.Driver == "debug" {
		gormLogger = logger.Default.LogMode(logger.Info)
	}

	// Use retry with exponential backoff
	retryCfg := utils.DefaultRetryConfig()
	err := utils.DoWithRetry(ctx, log, "PostgreSQL", retryCfg, func() error {
		var err error

		// 1. Open database connection
		db, err = gorm.Open(postgres.Open(cfg.Source), &gorm.Config{
			PrepareStmt: true,
			Logger:      gormLogger,
		})
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}

		// 2. Get underlying *sql.DB
		sqlDB, err := db.DB()
		if err != nil {
			return fmt.Errorf("failed to get database instance: %w", err)
		}

		// 3. Verify connection with ping
		if err = sqlDB.PingContext(ctx); err != nil {
			return fmt.Errorf("failed to ping database: %w", err)
		}

		log.Info("PostgreSQL connection established successfully")
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to initialize PostgreSQL after retries: %w", err)
	}

	// 4. Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Minute)

	log.Info("PostgreSQL connection pool configured",
		zap.Int("max_idle_conns", cfg.MaxIdleConns),
		zap.Int("max_open_conns", cfg.MaxOpenConns),
		zap.Int("conn_max_lifetime_minutes", cfg.ConnMaxLifetime),
	)

	return db, nil
}

// AutoMigrateModels runs database migrations for all domain models
// This should be called separately after successful connection
func AutoMigrateModels(db *gorm.DB, log *zap.Logger) error {
	log.Info("Starting database migrations...")

	// List of all persistence models to migrate
	models := []interface{}{
		// User
		&infraUser.UserModel{},

		// Account – Event Sourcing tables
		&infraAccount.AccountEventModel{},  // append-only write store
		&infraAccount.AccountReadModelDB{}, // denormalised read projection

		// Order
		&infraOrder.OrderModel{},

		// Portfolio
		&infraPortfolio.PositionModel{},

		// Market
		&infraMarket.StockModel{},
		&infraMarket.PriceModel{},
		&infraMarket.CandleModel{},

		// Execution
		&infraExecution.TradeModel{},
		&infraExecution.SettlementModel{},
		&infraExecution.ClearingInstructionModel{},

		// Risk
		&infraRisk.RiskLimitModel{},
		&infraRisk.RiskMetricsModel{},
		&infraRisk.RiskAlertModel{},
	}

	// Run migrations
	if err := db.AutoMigrate(models...); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Info("Database migrations completed successfully",
		zap.Int("models_migrated", len(models)),
	)

	return nil
}

// ClosePostgresDB closes the PostgreSQL database connection
func ClosePostgresDB(db *gorm.DB, log *zap.Logger) error {
	if db == nil {
		return nil
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		log.Error("Failed to close PostgreSQL connection", zap.Error(err))
		return fmt.Errorf("failed to close database: %w", err)
	}

	log.Info("PostgreSQL connection closed successfully")
	return nil
}

// GetDatabaseStats returns database connection pool statistics
func GetDatabaseStats(db *gorm.DB) (map[string]interface{}, error) {
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	stats := sqlDB.Stats()
	return map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration.String(),
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}, nil
}
