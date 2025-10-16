package db

import (
	"account-service/internal/config"
	"account-service/internal/domain/entity"
	"account-service/internal/logging"
	"errors"
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"time"
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

	if err := prePopulateDefaultCustomer(db); err != nil {
		return nil, fmt.Errorf("failed to prepopulate admin: %w", err)
	}

	logging.Logger.Info().Msg("database initialized successfully")
	return db, nil
}

func runMigrations(db *gorm.DB) error {
	return db.AutoMigrate(
		&entity.Customer{},
		&entity.Account{},
		&entity.Transaction{},
		&entity.Event{},
	)
}

func prePopulateDefaultCustomer(db *gorm.DB) error {
	customerList := []*entity.Customer{}
	newCustomerOne, _ := entity.NewCustomer("Customer One", "system")
	newCustomerTwo, _ := entity.NewCustomer("Customer Two", "system")
	newCustomerThree, _ := entity.NewCustomer("Customer Three", "system")
	customerList = append(customerList, newCustomerOne, newCustomerTwo, newCustomerThree)

	for _, cus := range customerList {
		var customer entity.Customer
		err := db.Where("name = ? AND status = ?", cus.Name, entity.CustomerStatusValid).First(&customer).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if err := db.Create(cus).Error; err != nil {
					logging.Logger.Warn().Err(err).Msg("Unable to populate default customer: " + cus.Name)
				}
			}

		}
	}
	return nil
}
