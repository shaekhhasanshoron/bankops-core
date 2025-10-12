package entity

import (
	"fmt"
	"github.com/google/uuid"
	"time"
)

const (
	EmployeeStatusValid   = "valid"
	EmployeeStatusInvalid = "invalid"

	EmployeeActiveStatusActive      = "active"
	EmployeeActiveStatusDeactivated = "deactivated"

	EmployeeRoleAdmin  = "admin"
	EmployeeRoleViewer = "viewer"
	EmployeeRoleEditor = "editor"

	EmployeeAuthMethodPassword = "password"
	EmployeeAuthMethodSSO      = "sso"
	EmployeeAuthMethodPasskey  = "passkey"
)

// Employee entity
type Employee struct {
	ID           string `gorm:"primaryKey"`
	Username     string `gorm:"not null"`
	AuthMethod   string `gorm:"not null"`
	Password     string `gorm:"null"`
	Role         string `gorm:"not null"`
	ActiveStatus string `gorm:"not null"`
	Status       string `gorm:"default:valid;not null"`
	CreatedBy    string `gorm:"null"` // username
	UpdatedBy    string `gorm:"null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewEmployee(id, username, password, role, authMethod, requester string) (*Employee, error) {
	if id == "" {
		id = uuid.New().String()
	}

	if requester == "" {
		requester = "system"
	}

	if authMethod == "" {
		authMethod = EmployeeAuthMethodPassword
	}

	if authMethod == EmployeeAuthMethodPassword {
		if password == "" {
			return nil, fmt.Errorf("password is required for password auth method")
		}
	} else {
		password = ""
	}

	return &Employee{
		ID:           id,
		Username:     username,
		AuthMethod:   authMethod,
		Password:     password,
		Role:         role,
		ActiveStatus: EmployeeActiveStatusActive,
		Status:       EmployeeStatusValid,
		CreatedBy:    requester,
		UpdatedBy:    "",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}, nil
}
