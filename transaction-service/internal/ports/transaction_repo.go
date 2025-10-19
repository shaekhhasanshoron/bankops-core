package ports

import (
	"time"
	"transaction-service/internal/domain/entity"
)

type TransactionRepo interface {
	CreateTransaction(transaction *entity.Transaction) error
	GetTransactionByID(id string) (*entity.Transaction, error)
	GetTransactionByReferenceID(referenceID string) (*entity.Transaction, error)
	UpdateTransactionStatus(id string, transactionStatus string, errorReason string) error
	UpdateTransaction(transaction *entity.Transaction) error
	GetTransactionHistory(accountID string, customerID string, startDate, endDate *time.Time, sortOrder string, page, pageSize int, types []string) ([]*entity.Transaction, int64, error)
	GetPendingTransactions() ([]*entity.Transaction, error)
	GetStuckTransactions() ([]*entity.Transaction, error)
}
