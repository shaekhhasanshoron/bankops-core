package bootstrap

import (
	"auth-service/internal/auth"
	"auth-service/internal/common"
	"auth-service/internal/config"
	"auth-service/internal/domain/entity"
	"gorm.io/gorm"
)

// CreateAdmin checks if the admin exists, and creates it if not
func CreateAdmin(db *gorm.DB) error {
	var admin entity.Employee

	err := db.Where("username = ? AND status = ?", "admin", common.EmployeeStatusValid).First(&admin).Error
	if err == nil {
		return nil
	}

	admin = entity.Employee{
		Username: config.Current().User.AdminUsername,
		Role:     "admin",
		Status:   common.EmployeeStatusValid,
	}

	hashedPassword, err := auth.HashData(config.Current().User.AdminPassword)
	if err != nil {
		return err
	}
	admin.Password = hashedPassword

	if err := db.Create(&admin).Error; err != nil {
		return err
	}

	return nil
}
