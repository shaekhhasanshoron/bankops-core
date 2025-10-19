package db

import (
	"auth-service/internal/adapters/auth"
	"auth-service/internal/common"
	"auth-service/internal/config"
	"auth-service/internal/domain/entity"
	"auth-service/internal/logging"
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
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	// Run migrations
	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	if err := prePopulateAdmin(db); err != nil {
		return nil, fmt.Errorf("failed to prepopulate admin: %w", err)
	}

	_ = prePopulateDefaultEmployee(db)

	logging.Logger.Info().Msg("database initialized successfully")
	return db, nil
}

func runMigrations(db *gorm.DB) error {
	return db.AutoMigrate(
		&entity.Employee{},
	)
}

func prePopulateAdmin(db *gorm.DB) error {
	var admin entity.Employee

	err := db.Where("username = ? AND status = ?", config.Current().User.AdminUsername, entity.EmployeeStatusValid).First(&admin).Error
	if err == nil {
		return nil
	}

	hashing := auth.NewHashing(config.Current().Auth.HashKey)

	hashedPassword, err := hashing.HashData(config.Current().User.AdminPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	newAdmin, err := entity.NewEmployee(
		config.Current().User.AdminUsername,
		hashedPassword,
		entity.EmployeeRoleAdmin,
		entity.EmployeeAuthMethodPassword,
		common.SystemUserUsername,
	)

	if err != nil {
		return err
	}

	if err := db.Create(newAdmin).Error; err != nil {
		return err
	}

	return nil
}

func prePopulateDefaultEmployee(db *gorm.DB) error {
	employeeList := []*entity.Employee{}
	newEmployeeViewer, _ := entity.NewEmployee(
		"viewer_user",
		"viewer_pass",
		entity.EmployeeRoleViewer,
		entity.EmployeeAuthMethodPassword,
		"system")
	newEmployeeEditor, _ := entity.NewEmployee(
		"editor_user",
		"editor_pass",
		entity.EmployeeRoleEditor,
		entity.EmployeeAuthMethodPassword,
		"system")

	employeeList = append(employeeList, newEmployeeViewer, newEmployeeEditor)

	for _, e := range employeeList {
		var existingEmployee entity.Employee
		err := db.Where("username = ? AND status = ?", e.Username, entity.EmployeeStatusValid).First(&existingEmployee).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				hashing := auth.NewHashing(config.Current().Auth.HashKey)
				hashedPassword, err := hashing.HashData(e.Password)
				if err != nil {
					logging.Logger.Warn().Err(err).Str("employee", e.Username).Msg("unable to hash password")
					continue
				}
				e.Password = hashedPassword
				if err := db.Create(e).Error; err != nil {
					logging.Logger.Warn().Err(err).Str("name", e.Username).Msg("Unable to populate default employee")
				} else {
					logging.Logger.Info().Str("name", e.Username).Msg("Default user created")
				}
			}
		}
	}
	return nil
}
