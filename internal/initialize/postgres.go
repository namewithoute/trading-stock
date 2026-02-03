package initialize

import (
	"context"
	"time"
	"trading-stock/internal/config"
	"trading-stock/internal/domain"
	"trading-stock/internal/global"
	"trading-stock/pkg/utils"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitPosgresDB(ctx context.Context, cfg config.DatabaseConfig) error {
	// Sử dụng Retry Helper
	err := utils.DoWithRetry(ctx, global.Logger, "Postgres", 2*time.Second, func() error {
		// 1. Cố gắng Open connection
		var err error
		var db *gorm.DB

		db, err = gorm.Open(postgres.Open(cfg.Source), &gorm.Config{
			PrepareStmt: true,
		})
		if err != nil {
			return err
		}

		// 2. Open OK -> Check Ping
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}

		if err = sqlDB.PingContext(ctx); err != nil {
			return err
		}

		// 3. OK -> Auto Migrate và Assign Global
		// Lưu ý: AutoMigrate chỉ nên chạy khi kết nối chắc chắn OK.
		// Nhưng nếu thích đơn giản có thể để ở đây hoặc ra ngoài.
		// Để an toàn, ta chỉ connect ở đây. Migrate làm sau.
		global.DB = db
		return nil
	})

	if err != nil {
		return err
	}

	// 4. Cấu hình Connection Pool & Migrate (Chỉ chạy khi đã connect thành công)
	// Lưu ý: global.DB đã được gán bên trong retry function nếu thành công
	sqlDB, _ := global.DB.DB()
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Minute)

	// 5. Auto Migrate
	if err := global.DB.AutoMigrate(
		&domain.Order{},
		&domain.Trade{},
		&domain.Wallet{},
	); err != nil {
		return err
	}

	return nil
}

func ClosePosgresDB() {
	if sqlDB, err := global.DB.DB(); err == nil {
		if err := sqlDB.Close(); err != nil {
			global.Logger.Error("Failed to close Postgres", zap.Error(err))
		} else {
			global.Logger.Info("Postgres connection closed")
		}
	}
}
