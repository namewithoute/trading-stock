package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Kafka    KafkaConfig    `mapstructure:"kafka"`
	Logger   LoggerConfig   `mapstructure:"logger"`
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

// Load reads metadata from a file or environment variables
func Load() *Config {
	v := viper.New()
	v.SetConfigName("dev")  // Tên file config (không cần đuôi)
	v.SetConfigType("yaml") // Hoặc json, env
	v.AddConfigPath(".")    // Tìm ở thư mục root
	v.AddConfigPath("./internal/configs")
	// Đọc lệnh từ biến môi trường (Ví dụ: DB_SOURCE)
	v.AutomaticEnv()
	if err := v.ReadInConfig(); err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		panic(fmt.Errorf("unable to decode into struct, %v", err))
	}
	return &cfg
}
