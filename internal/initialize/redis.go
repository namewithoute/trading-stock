package initialize

import (
	"context"
	"trading-stock/internal/config"
	"trading-stock/internal/global"
	"trading-stock/pkg/utils"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func InitRedis(ctx context.Context, cfg config.RedisConfig) error {
	retryCfg := utils.DefaultRetryConfig()
	return utils.DoWithRetry(ctx, global.Logger, "Redis", retryCfg, func() error {
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
		if err := global.Redis.Close(); err != nil {
			global.Logger.Error("Failed to close Redis", zap.Error(err))
		} else {
			global.Logger.Info("Redis connection closed")
		}
	}
}
