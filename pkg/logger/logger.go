package logger

import (
	"os"
	"path/filepath"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// LoggerConfig defines logger configuration
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

// InitLogger initializes Zap logger with configuration
func InitLogger(cfg LoggerConfig) (*zap.Logger, error) {
	// Create log directory
	if err := os.MkdirAll(cfg.Director, 0755); err != nil {
		return nil, err
	}

	// Set log level
	level := getLogLevel(cfg.Level)

	// Encoder configuration
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

	// Console Core (with colors)
	if cfg.LogInConsole {
		cores = append(cores, zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			zapcore.AddSync(os.Stdout),
			level,
		))
	}

	// File Core (JSON format)
	fileEncoderConfig := encoderConfig
	fileEncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder // No colors in file

	fileEncoder := zapcore.NewJSONEncoder(fileEncoderConfig)
	writer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   filepath.Join(cfg.Director, "all.log"),
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
		Compress:   true,
	})
	cores = append(cores, zapcore.NewCore(fileEncoder, writer, level))

	// Create logger instance
	core := zapcore.NewTee(cores...)
	logger := zap.New(core)

	// Add caller info if enabled
	if cfg.ShowLine {
		logger = logger.WithOptions(zap.AddCaller())
	}

	return logger, nil
}

// getLogLevel converts string level to zapcore.Level
func getLogLevel(l string) zapcore.Level {
	switch l {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

// customTimeEncoder formats time in custom format
func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

// ZapLogger returns Echo middleware that logs HTTP requests using Zap
func ZapLogger(logger *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			// Process request
			err := next(c)
			if err != nil {
				c.Error(err)
			}

			// Get request info
			req := c.Request()
			res := c.Response()

			// Get request ID from context (set by RequestID middleware)
			id := req.Header.Get(echo.HeaderXRequestID)
			if id == "" {
				id = res.Header().Get(echo.HeaderXRequestID)
			}

			// Build log fields
			fields := []zap.Field{
				zap.String("request_id", id),
				zap.String("method", req.Method),
				zap.String("uri", req.RequestURI),
				zap.String("remote_ip", c.RealIP()),
				zap.Int("status", res.Status),
				zap.Int64("bytes_in", req.ContentLength),
				zap.Int64("bytes_out", res.Size),
				zap.Duration("latency", time.Since(start)),
				zap.String("user_agent", req.UserAgent()),
			}

			// Add error if present
			if err != nil {
				fields = append(fields, zap.Error(err))
			}

			// Log based on status code
			switch {
			case res.Status >= 500:
				logger.Error("HTTP request", fields...)
			case res.Status >= 400:
				logger.Warn("HTTP request", fields...)
			case res.Status >= 300:
				logger.Info("HTTP request", fields...)
			default:
				logger.Debug("HTTP request", fields...)
			}

			return nil
		}
	}
}
