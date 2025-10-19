package db

import (
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"time"
	"transaction-service/internal/config"
	"transaction-service/internal/domain/entity"
	"transaction-service/internal/logging"
)

func InitDB() (*gorm.DB, error) {
	gormLogger := logger.New(
		&logging.Logger,
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	db, err := gorm.Open(sqlite.Open(config.Current().DB.DSN), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(25)

	// Run migrations
	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	logging.Logger.Info().Msg("database initialized successfully")
	return db, nil
}

func runMigrations(db *gorm.DB) error {
	return db.AutoMigrate(
		&entity.TransactionSaga{},
		&entity.Transaction{},
		&entity.Event{},
	)
}
