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
	// 1. Khởi tạo Echo
	e := echo.New()
	// 2. Cấu hình Middleware cơ bản
	e.Use(middleware.Logger())  // Log các request HTTP
	e.Use(middleware.Recover()) // Tránh crash server khi có panic
	// 3. Định nghĩa Route đơn giản để test
	e.GET("/ping", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"message": "pong",
		})
	})
	// 4. Khởi chạy Server trong một Goroutine để không làm block chương trình
	go func() {
		if err := e.Start(":8080"); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// 5. Thiết lập Graceful Shutdown (Lắng nghe tín hiệu từ OS)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	// Chờ đợi 1 trong 2 sự kiện
	select {
	case <-quit:
		global.Logger.Info("Shutting down server (Signal received)...")
	case err := <-errChan:
		global.Logger.Error("Shutting down server (Startup failed)", zap.Error(err))
	}
	global.Logger.Info("Shutting down server...")
	// Timeout 10 giây để server đóng các kết nối đang dang dở
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		global.Logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	global.Logger.Info("Cleaning up resources...")

	done := make(chan struct{})
	go func() {
		defer close(done)
		initialize.ClosePosgresDB()
		initialize.CloseRedis()
		initialize.CloseKafka()
	}()

	select {
	case <-done:
		global.Logger.Info("Resources closed successfully")
	case <-ctx.Done():
		global.Logger.Warn("Timeout during resource cleanup, forcing exit")
	}
	global.Logger.Info("Server exited properly")
	global.Logger.Sync()
}
