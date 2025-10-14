package ports

import (
	"account-service/internal/domain/entity"
	"time"
)

type TransactionRepo interface {
	CreateTransaction(transaction *entity.Transaction) error
	GetTransactionByID(id string) (*entity.Transaction, error)
	GetTransactionByReferenceID(referenceID string) (*entity.Transaction, error)
	BeginTransactionLifecycle(transactionID string, accountIDs []string) error
	UpdateTransactionStatus(id string, transactionStatus string, errorReason string) error
	CompleteTransactionLifecycle(transactionID string) error
	UpdateTransaction(transaction *entity.Transaction) error
	GetTransactionHistory(accountID string, customerID string, startDate, endDate *time.Time, sortOrder string, page, pageSize int, types []string) ([]*entity.Transaction, int64, error)
	GetPendingTransactions() ([]*entity.Transaction, error)
	GetStuckTransactions() ([]*entity.Transaction, error)
	GetLockedButIncompleteTransactions() ([]*entity.Transaction, error)
	ForceUnlockAccounts(transactionID string) error
	UpdateTransactionOnRecovery(transaction *entity.Transaction) error
	GetTransactionWithAccounts(transactionID string) (*entity.Transaction, []*entity.Account, error)
}
