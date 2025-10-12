package entity

import (
	"time"
)

// Employee entity
type Employee struct {
	ID        uint   `gorm:"primaryKey"`
	Username  string `gorm:"not null"`
	Password  string `gorm:"not null"`
	Role      string `gorm:"not null"`
	Status    string `gorm:"default:valid;not null"` // "valid" or "invalid"
	CreatedAt time.Time
	UpdatedAt time.Time
}
