package entity

import (
	"account-service/internal/domain/value"
	"github.com/google/uuid"
	"time"
)

const (
	TransactionPending    = "pending"
	TransactionCompleted  = "completed"
	TransactionFailed     = "failed"
	TransactionCancelled  = "cancelled"
	TransactionRecovering = "recovering"
)

type Transaction struct {
	ID                string     `gorm:"primaryKey"`
	FromAccountID     string     `gorm:"not null;index"`
	ToAccountID       string     `gorm:"not null;index"`
	Amount            float64    `gorm:"not null;check:amount > 0"`
	TransactionStatus string     `gorm:"not null;index"`
	ReferenceID       string     `gorm:"uniqueIndex"`
	ErrorReason       string     `json:"error_reason,omitempty"`
	RetryCount        int        `gorm:"default:0"`
	LastRetryAt       *time.Time `json:"last_retry_at,omitempty"`
	TimeoutAt         time.Time  `gorm:"index"`
	Version           int        `gorm:"default:1"` // Optimistic locking
	Status            string     `gorm:"not null;default:valid"`
	CreatedBy         string     `gorm:"null"`
	UpdatedBy         string     `gorm:"null"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func NewTransaction(fromAccountID, toAccountID string, amount float64, referenceID, createdBy string) (*Transaction, error) {
	now := time.Now()
	return &Transaction{
		ID:            uuid.New().String(),
		FromAccountID: fromAccountID,
		ToAccountID:   toAccountID,
		Amount:        amount,
		Status:        TransactionPending,
		ReferenceID:   referenceID,
		CreatedAt:     now,
		UpdatedAt:     now,
		CreatedBy:     createdBy,
		TimeoutAt:     now.Add(5 * time.Minute), // Default timeout
		Version:       1,
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
	t.Status = TransactionRecovering
	t.UpdatedAt = now
}

func (t *Transaction) IncrementVersion() {
	t.Version++
	t.UpdatedAt = time.Now()
}

func (t *Transaction) Validate() error {
	if t.Amount <= 0 {
		return value.ErrInvalidAmount
	}
	if t.FromAccountID == t.ToAccountID {
		return value.ErrSameAccountTransfer
	}
	if t.ReferenceID == "" {
		return value.ErrMissingReferenceID
	}
	return nil
}
