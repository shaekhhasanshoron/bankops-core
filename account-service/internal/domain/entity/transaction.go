package entity

import (
	"account-service/internal/domain/value"
	"github.com/google/uuid"
	"time"
)

const (
	TransactionStatusPending    = "pending"
	TransactionStatusCompleted  = "completed"
	TransactionStatusFailed     = "failed"
	TransactionStatusCancelled  = "cancelled"
	TransactionStatusRecovering = "recovering"

	TransactionTypeWithdrawFull   = "withdraw_full"
	TransactionTypeWithdrawAmount = "withdraw_amount"
	TransactionTypeAddAmount      = "add_amount"
	TransactionTypeTransfer       = "transfer"
)

type Transaction struct {
	ID                   string     `gorm:"primaryKey"`
	SourceAccountID      string     `gorm:"not null;index"`
	DestinationAccountID *string    `gorm:"index"`
	Amount               float64    `gorm:"not null;check:amount >= 0" json:"amount"`
	Type                 string     `gorm:"not null;index" json:"type"`
	TransactionStatus    string     `gorm:"not null;index"`
	ReferenceID          string     `gorm:"uniqueIndex"`
	TimeoutAt            time.Time  `gorm:"index"`
	Version              int        `gorm:"default:1"`
	LastRetryAt          *time.Time `gorm:"null"`
	RetryCount           int        `gorm:"default:0"`
	ErrorReason          string     `gorm:"null"`
	CreatedBy            string     `gorm:"not null"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
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

func (t *Transaction) ShouldTimeout() bool {
	return time.Now().After(t.TimeoutAt)
}

func (t *Transaction) CanRetry() bool {
	return t.RetryCount < 3 && !t.ShouldTimeout()
}

func (t *Transaction) MarkForRetry() {
	now := time.Now()
	t.RetryCount++
	t.LastRetryAt = &now
	t.TransactionStatus = TransactionStatusRecovering
	t.UpdatedAt = now
}

func (t *Transaction) IncrementVersion() {
	t.Version++
	t.UpdatedAt = time.Now()
}

func (t *Transaction) Validate() error {
	if t.Amount < 0 {
		return value.ErrInvalidAmount
	}

	if t.ReferenceID == "" {
		return value.ErrMissingReferenceID
	}

	switch t.Type {
	case TransactionTypeTransfer:
		if t.DestinationAccountID == nil {
			return value.ErrMissingDestinationAccount
		}
		if t.SourceAccountID == *t.DestinationAccountID {
			return value.ErrSameAccountTransfer
		}
		if t.Amount <= 0 {
			return value.ErrInvalidAmount
		}

	case TransactionTypeWithdrawFull:
		t.DestinationAccountID = nil
	case TransactionTypeWithdrawAmount:
		if t.Amount <= 0 {
			return value.ErrInvalidAmount
		}
		t.DestinationAccountID = nil
	case TransactionTypeAddAmount:
		if t.Amount <= 0 {
			return value.ErrInvalidAmount
		}
		t.DestinationAccountID = nil
	default:
		return value.ErrInvalidTransactionType
	}

	return nil
}

func (t *Transaction) RequiresDestinationAccount() bool {
	return t.Type == TransactionTypeTransfer
}

func (t *Transaction) GetAccountsToLock() []string {
	accounts := []string{t.SourceAccountID}
	if t.RequiresDestinationAccount() && t.DestinationAccountID != nil {
		accounts = append(accounts, *t.DestinationAccountID)
	}
	return accounts
}
