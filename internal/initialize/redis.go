package initialize

import (
	"context"
	"time"
	"trading-stock/internal/config"
	"trading-stock/internal/global"
	"trading-stock/pkg/utils"

	"github.com/redis/go-redis/v9"
)

func InitRedis(ctx context.Context, cfg config.RedisConfig) error {
	return utils.DoWithRetry(ctx, global.Logger, "Redis", 2*time.Second, func() error {
		global.Redis = redis.NewClient(&redis.Options{
			Addr:         cfg.Addr,
			Password:     cfg.Password,
			DB:           cfg.DB,
			PoolSize:     cfg.PoolSize,
			MinIdleConns: cfg.MinIdleConns,
		})

		// Test connection
		_, err := global.Redis.Ping(ctx).Result()
		if err != nil {
			return err
		}

		return nil
	})
}

func CloseRedis() {
	if global.Redis != nil {
		global.Redis.Close()
		global.Logger.Info("Redis connection closed")
	}
}
