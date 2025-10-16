package entity

import (
	custom_err "account-service/internal/domain/error"
	"github.com/google/uuid"
	"time"
)

const (
	AccountTypeSavings = "savings"
	AccountTypeCurrent = "current"

	AccountStatusValid   = "valid"
	AccountStatusInvalid = "invalid"

	AccountActiveStatusActive   = "active"
	AccountActiveStatusInactive = "inactive"
	AccountActiveStatusLocked   = "locked"
)

type Account struct {
	ID                  string  `gorm:"primaryKey"`
	CustomerID          string  `gorm:"not null;index"`
	Balance             float64 `gorm:"not null;default:0;check:balance >= 0"`
	AccountType         string  `gorm:"not null;default:'savings'"`
	ActiveStatus        string  `gorm:"not null;default:'active'"`
	LockedForTx         bool    `gorm:"default:false;index"` // locked for Transaction
	ActiveTransactionID *string `gorm:"index"`
	Version             int     `gorm:"default:1"` // Optimistic locking
	Status              string  `gorm:"not null;default:valid"`
	CreatedBy           string  `gorm:"null"`
	UpdatedBy           string  `gorm:"null"`
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

func NewAccount(customerID string, accountType string, initialDeposit float64, requester string) (*Account, error) {
	if customerID == "" {
		return nil, custom_err.ErrInvalidCustomer
	}

	if requester == "" {
		requester = "system"
	}

	now := time.Now()
	return &Account{
		ID:           uuid.New().String(),
		CustomerID:   customerID,
		Balance:      initialDeposit,
		AccountType:  accountType,
		ActiveStatus: AccountActiveStatusActive,
		Version:      1,
		Status:       AccountStatusValid,
		CreatedBy:    requester,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

func (a *Account) CanTransact() bool {
	return a.Status == AccountStatusValid &&
		a.ActiveStatus == AccountActiveStatusActive &&
		!a.LockedForTx &&
		a.ActiveTransactionID == nil
}

func (a *Account) LockForTransaction(transactionID string) {
	a.LockedForTx = true
	a.ActiveTransactionID = &transactionID
}

func (a *Account) UnlockFromTransaction() {
	a.LockedForTx = false
	a.ActiveTransactionID = nil
}

func (a *Account) HasActiveTransaction() bool {
	return a.ActiveTransactionID != nil
}

func (a *Account) IncrementVersion() {
	a.Version++
	a.UpdatedAt = time.Now()
}

func (a *Account) Validate() error {
	if a.Balance < 0 {
		return custom_err.ErrNegativeBalance
	}
	if a.CustomerID == "" {
		return custom_err.ErrMissingCustomerID
	}
	return nil
}
