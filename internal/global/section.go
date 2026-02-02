package global

import (
	"trading-stock/internal/config"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	Logger *zap.Logger
	Config *config.Config
	DB     *gorm.DB
	Redis  *redis.Client
	Kafka  *kafka.Writer
)
