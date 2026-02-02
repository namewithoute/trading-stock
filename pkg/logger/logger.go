package logger

import (
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// LoggerConfig định nghĩa những gì Logger cần, không phụ thuộc vào internal/config
type LoggerConfig struct {
	Level         string
	Director      string
	ShowLine      bool
	StacktraceKey string
	LogInConsole  bool
	MaxSize       int
	MaxBackups    int
	MaxAge        int
}

func InitLogger(cfg LoggerConfig) (*zap.Logger, error) {
	var logger *zap.Logger

	// 1. Tạo thư mục log
	if err := os.MkdirAll(cfg.Director, 0755); err != nil {
		return nil, err
	}

	// 2. Thiết lập level
	level := getLogLevel(cfg.Level)

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  cfg.StacktraceKey,
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     customTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var cores []zapcore.Core

	// Console Core
	if cfg.LogInConsole {
		cores = append(cores, zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			zapcore.AddSync(os.Stdout),
			level,
		))
	}

	// File Core (JSON)
	fileEncoder := zapcore.NewJSONEncoder(encoderConfig)
	writer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   filepath.Join(cfg.Director, "all.log"),
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
		Compress:   true,
	})
	cores = append(cores, zapcore.NewCore(fileEncoder, writer, level))

	// Tạo Logger instance
	core := zapcore.NewTee(cores...)
	logger = zap.New(core)

	if cfg.ShowLine {
		logger = logger.WithOptions(zap.AddCaller())
	}

	return logger, nil
}

func getLogLevel(l string) zapcore.Level {
	switch l {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}
