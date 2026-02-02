package initialize

import (
	"time"
	"trading-stock/internal/config"
	"trading-stock/internal/domain"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var PosgresDB *gorm.DB

func InitPosgresDB(cfg config.DatabaseConfig) (*gorm.DB, error) {
	var err error
	PosgresDB, err = gorm.Open(postgres.Open(cfg.Source), &gorm.Config{
		PrepareStmt: true, // Cache prepared statements for better performance
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := PosgresDB.DB()
	if err != nil {
		return nil, err
	}

	// Cấu hình Connection Pool từ Config
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Minute)

	// Thực hiện AutoMigrate các domain
	err = PosgresDB.AutoMigrate(
		&domain.Order{},
		&domain.Trade{},
		&domain.Wallet{},
	)
	if err != nil {
		return nil, err
	}

	return PosgresDB, nil
}
