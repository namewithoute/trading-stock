package app

import (
	"context"
	"fmt"

	"trading-stock/internal/infrastructure/persistence"

	"go.uber.org/zap"
)

// initResources connects to all external infrastructure services.
//
// Connections are established in dependency order:
//  1. PostgreSQL — primary data store (migrations run here)
//  2. Redis      — cache and session store
//  3. Kafka      — event streaming
//
// Called with a deadline context to prevent indefinite blocking on startup.
func (a *App) initResources(ctx context.Context) error {
	// ── PostgreSQL ────────────────────────────────────────────────────
	db, err := persistence.InitPostgresDB(ctx, a.Config.Database, a.Logger)
	if err != nil {
		return fmt.Errorf("postgres: %w", err)
	}
	a.DB = db

	if err := persistence.AutoMigrateModels(a.DB, a.Logger); err != nil {
		return fmt.Errorf("postgres migration: %w", err)
	}
	a.Logger.Info("[ Infrastructure ] PostgreSQL connected and migrated")

	// ── Redis ─────────────────────────────────────────────────────────
	rdb, err := persistence.InitRedis(ctx, a.Config.Redis, a.Logger)
	if err != nil {
		return fmt.Errorf("redis: %w", err)
	}
	a.Redis = rdb
	a.Logger.Info("[ Infrastructure ] Redis connected")

	// ── Kafka ─────────────────────────────────────────────────────────
	kw, err := persistence.InitKafka(ctx, a.Config.Kafka, a.Logger)
	if err != nil {
		// NOTE: Kafka failure is non-fatal in some environments.
		// Change to return err if Kafka is strictly required.
		a.Logger.Warn("[ Infrastructure ] Kafka unavailable – continuing without it",
			zap.Error(err),
		)
	} else {
		a.Kafka = kw
		a.Logger.Info("[ Infrastructure ] Kafka connected")
	}

	return nil
}
