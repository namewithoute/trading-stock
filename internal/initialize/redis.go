package initialize

import (
	"context"
	"fmt"

	"trading-stock/internal/config"
	"trading-stock/pkg/utils"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// InitRedis initializes Redis connection with retry logic
// Returns *redis.Client instance instead of using global variables
func InitRedis(ctx context.Context, cfg config.RedisConfig, log *zap.Logger) (*redis.Client, error) {
	var client *redis.Client

	// Use retry with exponential backoff
	retryCfg := utils.DefaultRetryConfig()
	err := utils.DoWithRetry(ctx, log, "Redis", retryCfg, func() error {
		// Create Redis client
		client = redis.NewClient(&redis.Options{
			Addr:         cfg.Addr,
			Password:     cfg.Password,
			DB:           cfg.DB,
			PoolSize:     cfg.PoolSize,
			MinIdleConns: cfg.MinIdleConns,
		})

		// Test connection with ping
		if err := client.Ping(ctx).Err(); err != nil {
			return fmt.Errorf("failed to ping Redis: %w", err)
		}

		log.Info("Redis connection established successfully")
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to initialize Redis after retries: %w", err)
	}

	log.Info("Redis connection pool configured",
		zap.String("addr", cfg.Addr),
		zap.Int("db", cfg.DB),
		zap.Int("pool_size", cfg.PoolSize),
		zap.Int("min_idle_conns", cfg.MinIdleConns),
	)

	return client, nil
}

// CloseRedis closes the Redis connection
func CloseRedis(client *redis.Client, log *zap.Logger) error {
	if client == nil {
		return nil
	}

	if err := client.Close(); err != nil {
		log.Error("Failed to close Redis connection", zap.Error(err))
		return fmt.Errorf("failed to close Redis: %w", err)
	}

	log.Info("Redis connection closed successfully")
	return nil
}

// GetRedisStats returns Redis connection pool statistics
func GetRedisStats(ctx context.Context, client *redis.Client) (map[string]interface{}, error) {
	if client == nil {
		return nil, fmt.Errorf("redis client is nil")
	}

	stats := client.PoolStats()

	// Get server info
	info, err := client.Info(ctx, "server").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get Redis info: %w", err)
	}

	return map[string]interface{}{
		"pool_hits":     stats.Hits,
		"pool_misses":   stats.Misses,
		"pool_timeouts": stats.Timeouts,
		"total_conns":   stats.TotalConns,
		"idle_conns":    stats.IdleConns,
		"stale_conns":   stats.StaleConns,
		"server_info":   info,
	}, nil
}
