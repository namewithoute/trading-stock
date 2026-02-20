package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Kafka    KafkaConfig    `mapstructure:"kafka"`
	Logger   LoggerConfig   `mapstructure:"logger"`
	Init     InitConfig     `mapstructure:"init"`
	JWT      JWTConfig      `mapstructure:"jwt"`
}

type InitConfig struct {
	StartupTimeout time.Duration `mapstructure:"startup_timeout"`
	RetryConfig    RetryConfig   `mapstructure:"retry_config"`
}

type RetryConfig struct {
	MaxAttempts     int           `mapstructure:"max_attempts"`
	InitialInterval time.Duration `mapstructure:"initial_interval"`
	MaxInterval     time.Duration `mapstructure:"max_interval"`
	Multiplier      float64       `mapstructure:"multiplier"`
	Jitter          bool          `mapstructure:"jitter"`
}

type AppConfig struct {
	Port int    `mapstructure:"port"`
	Env  string `mapstructure:"env"`
}

type DatabaseConfig struct {
	Driver          string `mapstructure:"driver"`
	Source          string `mapstructure:"source"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
}

type RedisConfig struct {
	Addr         string `mapstructure:"addr"`
	Password     string `mapstructure:"password"`
	DB           int    `mapstructure:"db"`
	PoolSize     int    `mapstructure:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns"`
}

type KafkaConfig struct {
	Brokers      []string `mapstructure:"brokers"`
	BatchSize    int      `mapstructure:"batch_size"`
	BatchTimeout int      `mapstructure:"batch_timeout"`
}

type LoggerConfig struct {
	Level         string `mapstructure:"level"`
	Director      string `mapstructure:"director"`
	ShowLine      bool   `mapstructure:"show_line"`
	StacktraceKey string `mapstructure:"stacktrace_key"`
	LogInConsole  bool   `mapstructure:"log_in_console"`
	MaxSize       int    `mapstructure:"max_size"`
	MaxBackups    int    `mapstructure:"max_backups"`
	MaxAge        int    `mapstructure:"max_age"`
}

// JWTConfig holds secrets and TTL settings for JWT token generation.
// Access tokens are short-lived; refresh tokens are long-lived.
// NEVER commit real secrets to source control – use environment variables in production.
type JWTConfig struct {
	AccessSecret  string        `mapstructure:"access_secret"`
	RefreshSecret string        `mapstructure:"refresh_secret"`
	AccessTTL     time.Duration `mapstructure:"access_ttl"`  // e.g. "15m"
	RefreshTTL    time.Duration `mapstructure:"refresh_ttl"` // e.g. "168h" (7 days)
	Issuer        string        `mapstructure:"issuer"`
}

// Load reads configuration from a YAML file or environment variables.
func Load() *Config {
	v := viper.New()
	v.SetConfigName("dev")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./internal/configs")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		panic(fmt.Errorf("unable to decode into struct: %v", err))
	}

	return &cfg
}
