package entity

import (
	"errors"
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

var (
	ErrPasswordRequired = errors.New("password required")
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
	CreatedBy    string `gorm:"null"`
	UpdatedBy    string `gorm:"null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewEmployee(username, password, role, authMethod, requester string) (*Employee, error) {
	if requester == "" {
		requester = "system"
	}

	if authMethod == "" {
		authMethod = EmployeeAuthMethodPassword
	}

	if authMethod == EmployeeAuthMethodPassword {
		if password == "" {
			return nil, ErrPasswordRequired
		}
	} else {
		password = ""
	}

	return &Employee{
		ID:           uuid.New().String(),
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

func (e *Employee) IsValid() bool {
	return e.ID != "" && e.Username != ""
}
