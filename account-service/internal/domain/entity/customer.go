package entity

import (
	custom_err "account-service/internal/domain/error"
	"encoding/json"
	"github.com/google/uuid"
	"strings"
	"time"
)

const (
	CustomerStatusValid   = "valid"
	CustomerStatusInvalid = "invalid"

	CustomerActiveStatusActive      = "active"
	CustomerActiveStatusDeactivated = "deactivated"
)

type Customer struct {
	ID                 string    `gorm:"primaryKey"`
	Name               string    `gorm:"not null"`
	Accounts           []Account `gorm:"foreignKey:CustomerID"`
	ActiveStatus       string    `gorm:"not null;default:active"`
	LockedForOperation bool      `gorm:"default:false;index"`
	Version            int       `gorm:"default:1"`
	Status             string    `gorm:"not null;default:valid"` // soft delete
	CreatedBy          string    `gorm:"null"`
	UpdatedBy          string    `gorm:"null"`
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func NewCustomer(name, requester string) (*Customer, error) {
	if name == "" {
		return nil, custom_err.ErrInvalidCustomer
	}

	if requester == "" {
		requester = "system"
	}

	now := time.Now()
	return &Customer{
		ID:           uuid.New().String(),
		Name:         strings.TrimSpace(name),
		ActiveStatus: CustomerActiveStatusActive,
		Version:      1,
		Status:       CustomerStatusValid,
		CreatedBy:    requester,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

func (c *Customer) CanBeModified() bool {
	return c.Status == CustomerStatusValid &&
		c.ActiveStatus == AccountActiveStatusActive &&
		!c.LockedForOperation &&
		!c.HasAccountsInTransaction()
}

func (c *Customer) HasAccountsInTransaction() bool {
	if c.Accounts == nil || len(c.Accounts) == 0 {
		return false
	}

	for _, account := range c.Accounts {
		if account.HasActiveTransaction() {
			return true
		}
	}
	return false
}

func (c *Customer) LockForOperation() {
	c.LockedForOperation = true
}

func (c *Customer) UnlockFromOperation() {
	c.LockedForOperation = false
}

func (c *Customer) IncrementVersion() {
	c.Version++
	c.UpdatedAt = time.Now()
}

func (e *Customer) ToString() string {
	jsonData, _ := json.Marshal(&e)
	return string(jsonData)
}
