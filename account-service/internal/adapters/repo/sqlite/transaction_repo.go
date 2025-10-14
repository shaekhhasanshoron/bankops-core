package sqlite

import (
	"account-service/internal/domain/entity"
	"account-service/internal/ports"
	"errors"
	"gorm.io/gorm"
	"sync"
	"time"
)

var (
	ErrAccountLockedForTransaction = errors.New("account is locked for an ongoing transaction")
)

// TransactionRepo struct to interact with the database.
type TransactionRepo struct {
	DB *gorm.DB
	mu sync.RWMutex
}

// NewTransactionRepo creates a new TransactionRepo instance with an SQLite connection.
func NewTransactionRepo(db *gorm.DB) ports.TransactionRepo {
	return &TransactionRepo{DB: db}
}

func (r *TransactionRepo) CreateTransaction(transaction *entity.Transaction) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.DB.Create(transaction).Error
}

func (r *TransactionRepo) GetTransactionByID(id string) (*entity.Transaction, error) {
	var transaction entity.Transaction
	err := r.DB.Where("id = ?", id).First(&transaction).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &transaction, err
}

func (r *TransactionRepo) GetTransactionByReferenceID(referenceID string) (*entity.Transaction, error) {
	var transaction entity.Transaction
	err := r.DB.Where("reference_id = ?", referenceID).First(&transaction).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &transaction, err
}

func (r *TransactionRepo) BeginTransactionLifecycle(transactionID string, accountIDs []string) error {
	tx := r.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Lock all accounts involved in the transaction
	for _, accountID := range accountIDs {
		var account entity.Account
		if err := tx.Set("gorm:query_option", "FOR UPDATE").
			Where("id = ?", accountID).
			First(&account).Error; err != nil {
			tx.Rollback()
			return err
		}

		if account.LockedForTx {
			tx.Rollback()
			return ErrAccountLockedForTransaction
		}

		// Apply transaction lock
		if err := tx.Model(&entity.Account{}).
			Where("id = ?", accountID).
			Updates(map[string]interface{}{
				"locked_for_tx":         true,
				"active_transaction_id": transactionID,
			}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

func (r *TransactionRepo) CompleteTransactionLifecycle(transactionID string) error {
	tx := r.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Unlock all accounts involved in this transaction
	if err := tx.Model(&entity.Account{}).
		Where("active_transaction_id = ?", transactionID).
		Updates(map[string]interface{}{
			"locked_for_tx":         false,
			"active_transaction_id": nil,
		}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (r *TransactionRepo) UpdateTransactionStatus(id string, transactionStatus string, errorReason string) error {
	updateFields := map[string]interface{}{
		"transaction_status": transactionStatus,
	}

	if errorReason != "" {
		updateFields["error_reason"] = errorReason
	}

	return r.DB.Model(&entity.Transaction{}).
		Where("id = ?", id).
		Updates(updateFields).Error
}

func (r *TransactionRepo) UpdateTransaction(transaction *entity.Transaction) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	currentVersion := transaction.Version
	result := r.DB.Model(transaction).
		Where("id = ? AND version = ?", transaction.ID, currentVersion).
		Updates(map[string]interface{}{
			"transaction_status": transaction.TransactionStatus,
			"error_reason":       transaction.ErrorReason,
			"updated_at":         transaction.UpdatedAt,
			"retry_count":        transaction.RetryCount,
			"last_retry_at":      transaction.LastRetryAt,
			"version":            currentVersion + 1,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("concurrent modification detected")
	}

	transaction.Version = currentVersion + 1
	return nil
}

func (r *TransactionRepo) UpdateTransactionOnRecovery(transaction *entity.Transaction) error {
	return r.DB.Save(transaction).Error
}

func (r *TransactionRepo) GetTransactionHistory(accountID string, customerID string, startDate, endDate *time.Time, sortOrder string, page, pageSize int, types []string) ([]*entity.Transaction, int64, error) {
	var transactions []*entity.Transaction
	var total int64

	offset := (page - 1) * pageSize
	query := r.DB.Model(&entity.Transaction{})

	// Build query - account can be source or destination
	if accountID != "" {
		query = query.Where("source_account_id = ? OR destination_account_id = ?", accountID, accountID)

	} else if customerID != "" {
		var accountIDs []string
		if err := r.DB.Model(&entity.Account{}).
			Where("customer_id = ?", customerID).
			Pluck("id", &accountIDs).Error; err != nil {
			return nil, 0, err
		}

		if len(accountIDs) == 0 {
			return []*entity.Transaction{}, 0, nil
		}

		query = query.Where("source_account_id IN ? OR destination_account_id IN ?",
			accountIDs, entity.TransactionTypeTransfer, accountIDs)
	}

	if startDate != nil {
		query = query.Where("created_at >= ?", startDate)
	}

	if endDate != nil {
		query = query.Where("created_at <= ?", endDate)
	}

	if types != nil && len(types) > 0 {
		query = query.Where("type IN ?", types)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	order := "created_at DESC"
	if sortOrder == "asc" {
		order = "created_at ASC"
	}

	err := query.Offset(offset).Limit(pageSize).Order(order).Find(&transactions).Error
	return transactions, total, err
}

func (r *TransactionRepo) GetPendingTransactions() ([]*entity.Transaction, error) {
	var transactions []*entity.Transaction
	err := r.DB.
		Where("transaction_status IN ?", []string{
			entity.TransactionStatusPending,
			entity.TransactionStatusRecovering,
		}).
		Find(&transactions).Error
	return transactions, err
}

func (r *TransactionRepo) GetLockedButIncompleteTransactions() ([]*entity.Transaction, error) {
	var transactions []*entity.Transaction

	// This query finds transactions that have locked accounts but aren't completed
	err := r.DB.Joins("JOIN accounts ON accounts.active_transaction_id = transactions.id").
		Where("transactions.transaction_status != ?", entity.TransactionStatusCompleted).
		Where("accounts.locked_for_tx = ?", true).
		Group("transactions.id").
		Find(&transactions).Error

	return transactions, err
}

func (r *TransactionRepo) GetStuckTransactions() ([]*entity.Transaction, error) {
	var transactions []*entity.Transaction
	err := r.DB.
		Where("transaction_status IN ? AND timeout_at < ?",
			[]string{
				entity.TransactionStatusPending,
				entity.TransactionStatusRecovering,
			},
			time.Now()).
		Find(&transactions).Error
	return transactions, err
}

func (r *TransactionRepo) ForceUnlockAccounts(transactionID string) error {
	tx := r.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Model(&entity.Account{}).
		Where("active_transaction_id = ? AND status = ?", transactionID, entity.AccountStatusValid).
		Updates(map[string]interface{}{
			"locked_for_tx":         false,
			"active_transaction_id": nil,
		}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (r *TransactionRepo) GetTransactionWithAccounts(transactionID string) (*entity.Transaction, []*entity.Account, error) {
	var transaction entity.Transaction
	var accounts []*entity.Account

	err := r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ?", transactionID).First(&transaction).Error; err != nil {
			return err
		}

		if err := tx.Where("active_transaction_id = ?", transactionID).Find(&accounts).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	return &transaction, accounts, nil
}
