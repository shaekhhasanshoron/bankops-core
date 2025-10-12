package db

import (
	"auth-service/internal/config"
	"auth-service/internal/domain/entity"
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func InitDatabase() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(config.Current().DB.DSN), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate schema: create tables if they don't exist
	if err := db.AutoMigrate(&entity.Employee{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database schema: %v", err)
	}

	return db, nil
}
