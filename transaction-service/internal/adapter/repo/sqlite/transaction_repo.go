package sqlite

import (
	"errors"
	"gorm.io/gorm"
	"time"
	"transaction-service/internal/domain/entity"
	"transaction-service/internal/ports"
)

type TransactionRepo struct {
	DB *gorm.DB
}

func NewTransactionRepo(db *gorm.DB) ports.TransactionRepo {
	return &TransactionRepo{DB: db}
}

func (r *TransactionRepo) CreateTransaction(transaction *entity.Transaction) error {
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

func (r *TransactionRepo) UpdateTransactionStatus(id string, status string, errorReason string) error {
	updateFields := map[string]interface{}{
		"transaction_status": status,
		"updated_at":         time.Now(),
	}

	if errorReason != "" {
		updateFields["error_reason"] = errorReason
	}

	return r.DB.Model(&entity.Transaction{}).
		Where("id = ?", id).
		Updates(updateFields).Error
}

func (r *TransactionRepo) UpdateTransaction(transaction *entity.Transaction) error {
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

func (r *TransactionRepo) GetTransactionHistory(accountID string, customerID string, startDate, endDate *time.Time, sortOrder string, page, pageSize int, types []string) ([]*entity.Transaction, int64, error) {
	var transactions []*entity.Transaction
	var total int64

	offset := (page - 1) * pageSize
	query := r.DB.Model(&entity.Transaction{})

	// Build query - account can be source or destination
	if accountID != "" {
		query = query.Where("source_account_id = ? OR destination_account_id = ?", accountID, accountID)

	} else if customerID != "" {
		query = query.Where("source_account_customer_id = ? OR destination_account_customer_id = ?", customerID, customerID)
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

// GetStuckTransactions returns transactions that are not completed/failed
func (r *TransactionRepo) GetStuckTransactions() ([]*entity.Transaction, error) {
	var transactions []*entity.Transaction
	err := r.DB.
		Where("transaction_status IN ? AND timeout_at < ?",
			[]string{
				entity.TransactionStatusPending,
				entity.TransactionStatusRecovering,
				entity.TransactionStatusSuccessful,
			},
			time.Now()).
		Find(&transactions).Error
	return transactions, err
}
