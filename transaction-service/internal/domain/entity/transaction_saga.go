package entity

import (
	"github.com/google/uuid"
	"time"
)

const (
	TransactionSagaStepInitiate                = "initiate"
	TransactionSagaStepValidateAccounts        = "validate_accounts"
	TransactionSagaStepLockAccounts            = "lock_accounts"
	TransactionSagaStepProcessTransfer         = "process_transfer"
	TransactionSagaStepComplete                = "complete"
	TransactionSagaStepCompensateFundRollback  = "compensate_fund_rollback"
	TransactionSagaStepCompensateUnlockAccount = "compensate_unlock_accounts"
	TransactionSagaStepCompensate              = "compensate"
	TransactionSagaStepCompensateComplete      = "compensate_complete"
)

const (
	TransactionSagaStateInitiated        = "initiated"
	TransactionSagaStateValidating       = "validating"
	TransactionSagaStateValidated        = "validated"
	TransactionSagaStateValidationFailed = "validation_failed"
	TransactionSagaStateLocking          = "locking"
	TransactionSagaStateLocked           = "locked"
	TransactionSagaStateLockFailed       = "lock_failed"
	TransactionSagaStateProcessing       = "processing"
	TransactionSagaStateCompleted        = "completed"
	TransactionSagaStateFailed           = "failed"
	TransactionSagaStateCompensating     = "compensating"
	TransactionSagaStateCompensateFailed = "compensate_failed"
	TransactionSagaStateCompensated      = "compensated"
)

type TransactionSaga struct {
	ID                   string     `gorm:"primaryKey"`
	TransactionID        string     `gorm:"not null;uniqueIndex"`
	CurrentState         string     `gorm:"not null;index"`
	CurrentStep          string     `gorm:"not null"`
	SourceAccountID      string     `gorm:"not null"`
	DestinationAccountID *string    `gorm:"null"`
	Amount               float64    `gorm:"not null"`
	TransactionType      string     `gorm:"not null"`
	ReferenceID          string     `gorm:"not null"`
	CompensationRequired bool       `gorm:"default:false"`
	CompensationReason   string     `gorm:"null"`
	RetryCount           int        `gorm:"default:0"`
	MaxRetries           int        `gorm:"default:3"`
	LastRetryAt          *time.Time `gorm:"null"`
	NextRetryAt          *time.Time `gorm:"null"`
	TimeoutAt            time.Time  `gorm:"index"`
	CreatedAt            time.Time
	UpdatedAt            time.Time
	Version              int `gorm:"default:1"`
}

func NewTransactionSaga(transactionID, sourceAccountID string, destinationAccountID *string, amount float64, transactionType, referenceID string) *TransactionSaga {
	now := time.Now()
	return &TransactionSaga{
		ID:                   uuid.New().String(),
		TransactionID:        transactionID,
		CurrentState:         TransactionSagaStateInitiated,
		CurrentStep:          TransactionSagaStepValidateAccounts,
		SourceAccountID:      sourceAccountID,
		DestinationAccountID: destinationAccountID,
		Amount:               amount,
		TransactionType:      transactionType,
		ReferenceID:          referenceID,
		MaxRetries:           3,
		TimeoutAt:            now.Add(5 * time.Minute),
		CreatedAt:            now,
		UpdatedAt:            now,
		Version:              1,
	}
}

func (s *TransactionSaga) ShouldTimeout() bool {
	return time.Now().After(s.TimeoutAt)
}

func (s *TransactionSaga) CanRetry() bool {
	return s.RetryCount < s.MaxRetries && !s.ShouldTimeout()
}

func (s *TransactionSaga) MarkForRetry() {
	now := time.Now()
	s.RetryCount++
	s.LastRetryAt = &now
	nextRetry := now.Add(time.Duration(s.RetryCount) * time.Minute)
	s.NextRetryAt = &nextRetry
	s.UpdatedAt = now
}

func (s *TransactionSaga) IncrementVersion() {
	s.Version++
	s.UpdatedAt = time.Now()
}

func (s *TransactionSaga) GetAccountsToLock() []string {
	accounts := []string{s.SourceAccountID}
	if s.DestinationAccountID != nil {
		accounts = append(accounts, *s.DestinationAccountID)
	}
	return accounts
}

func (s *TransactionSaga) RequiresDestinationAccount() bool {
	return s.TransactionType == TransactionTypeTransfer
}
