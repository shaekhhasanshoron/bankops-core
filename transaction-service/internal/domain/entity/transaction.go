package entity

import (
	"encoding/json"
	"github.com/google/uuid"
	"time"
)

const (
	TransactionStatusPending    = "pending"
	TransactionStatusSuccessful = "successful" // if amounts are transferred
	TransactionStatusCompleted  = "completed"  // if accounts are unlocked
	TransactionStatusFailed     = "failed"
	TransactionStatusRecovering = "recovering"

	TransactionTypeTransfer       = "transfer"
	TransactionTypeWithdrawFull   = "withdraw_full"
	TransactionTypeWithdrawAmount = "withdraw_amount"
	TransactionTypeAddAmount      = "add_amount"
)

type Transaction struct {
	ID                           string     `gorm:"primaryKey"`
	SourceAccountCustomerID      string     `gorm:"not null;index"`
	SourceAccountID              string     `gorm:"not null;index"`
	DestinationAccountID         *string    `gorm:"index"`
	DestinationAccountCustomerID *string    `gorm:"index"`
	Amount                       float64    `gorm:"not null;check:amount >= 0"`
	Type                         string     `gorm:"not null;index"`
	TransactionStatus            string     `gorm:"not null;index"`
	ReferenceID                  string     `gorm:"index"`
	TimeoutAt                    time.Time  `gorm:"index"`
	Version                      int        `gorm:"default:1"`
	LastRetryAt                  *time.Time `gorm:"null"`
	RetryCount                   int        `gorm:"default:0"`
	ErrorReason                  string     `gorm:"null"`
	CreatedBy                    string     `gorm:"not null"`
	CreatedAt                    time.Time
	UpdatedAt                    time.Time
}

func NewTransaction(sourceAccountID string, destinationAccountID *string, amount float64, transactionType, referenceID, createdBy string) (*Transaction, error) {
	now := time.Now()
	return &Transaction{
		ID:                   uuid.New().String(),
		SourceAccountID:      sourceAccountID,
		DestinationAccountID: destinationAccountID,
		Amount:               amount,
		Type:                 transactionType,
		TransactionStatus:    TransactionStatusPending,
		ReferenceID:          referenceID,
		CreatedAt:            now,
		UpdatedAt:            now,
		CreatedBy:            createdBy,
		TimeoutAt:            now.Add(5 * time.Minute),
		Version:              1,
	}, nil
}

func (t *Transaction) GetAccountsToLock() []string {
	accounts := []string{t.SourceAccountID}
	if t.RequiresDestinationAccount() && t.DestinationAccountID != nil {
		accounts = append(accounts, *t.DestinationAccountID)
	}
	return accounts
}

func (t *Transaction) RequiresDestinationAccount() bool {
	return t.Type == TransactionTypeTransfer
}

func (e *Transaction) ToString() string {
	jsonData, _ := json.Marshal(&e)
	return string(jsonData)
}
