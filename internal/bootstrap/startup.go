package bootstrap

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"trading-stock/internal/config"
	"trading-stock/internal/global"
	"trading-stock/internal/initialize"
	"trading-stock/pkg/logger"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

func Setup() {
	// 1. Load Config (Lấy từ internal/config)
	cfg := config.Load()

	// 2. Init Logger (Truyền config sang pkg/logger)
	var err error
	global.Logger, err = logger.InitLogger(logger.LoggerConfig{
		Level:         cfg.Logger.Level,
		Director:      cfg.Logger.Director,
		ShowLine:      cfg.Logger.ShowLine,
		StacktraceKey: cfg.Logger.StacktraceKey,
		LogInConsole:  cfg.Logger.LogInConsole,
		MaxSize:       cfg.Logger.MaxSize,
		MaxBackups:    cfg.Logger.MaxBackups,
		MaxAge:        cfg.Logger.MaxAge,
	})

	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := initialize.InitPosgresDB(ctx, cfg.Database); err != nil {
		global.Logger.Panic("Failed to initialize postgres", zap.Error(err))
	}

	if err := initialize.InitRedis(ctx, cfg.Redis); err != nil {
		global.Logger.Panic("Failed to initialize redis", zap.Error(err))
	}

	if err := initialize.InitKafka(ctx, cfg.Kafka); err != nil {
		global.Logger.Panic("Failed to initialize kafka", zap.Error(err))
	}

	global.Logger.Info("System Bootstrap: Logger started successfully!")
}

func Run() {
	errChan := make(chan error, 1)

	// 1. Initialize Echo server
	e := echo.New()

	// 2. Configure middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// 3. Define health check route
	e.GET("/ping", func(c echo.Context) error {
		time.Sleep(20 * time.Second)
		return c.JSON(http.StatusOK, map[string]string{
			"message": "pong",
			"status":  "healthy",
		})
	})

	// 4. Start HTTP server in goroutine
	go func() {
		global.Logger.Info("Starting HTTP server on :8080")
		if err := e.Start(":8080"); err != nil && err != http.ErrServerClosed {
			global.Logger.Error("Server startup failed", zap.Error(err))
			errChan <- err
		}
	}()

	// 5. Setup graceful shutdown signal listener
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// 6. Wait for shutdown signal or startup error
	select {
	case <-quit:
		global.Logger.Info("Received shutdown signal (SIGTERM/SIGINT)")
	case err := <-errChan:
		global.Logger.Error("Server failed to start, initiating shutdown", zap.Error(err))
	}

	// 7. Execute graceful shutdown with proper timeout management
	shutdownCfg := DefaultShutdownConfig()

	// Create shutdown context
	ctx, cancel := context.WithTimeout(context.Background(), shutdownCfg.GracefulShutdownTimeout)
	defer cancel()

	// Shutdown HTTP server first (stop accepting new requests)
	global.Logger.Info("Shutting down HTTP server...")
	serverShutdownCtx, serverCancel := context.WithTimeout(ctx, shutdownCfg.ServerShutdownTimeout)
	defer serverCancel()

	if err := e.Shutdown(serverShutdownCtx); err != nil {
		global.Logger.Error("HTTP server forced shutdown", zap.Error(err))
	} else {
		global.Logger.Info("HTTP server stopped gracefully")
	}

	// Close all external resources (DB, Redis, Kafka)
	if err := CloseAllResources(); err != nil {
		global.Logger.Error("Resource cleanup failed", zap.Error(err))
	}

	// Final cleanup
	global.Logger.Info("Server exited properly")
	if err := global.Logger.Sync(); err != nil {
		// Ignore sync errors on Windows (known issue with zap)
		// https://github.com/uber-go/zap/issues/880
	}
}
