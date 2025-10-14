package sqlite

import (
	"account-service/internal/domain/entity"
	"account-service/internal/domain/value"
	"account-service/internal/ports"
	"errors"
	"gorm.io/gorm"
	"sync"
	"time"
)

// AccountRepo struct to interact with the database.
type AccountRepo struct {
	DB *gorm.DB
	mu sync.RWMutex
}

// NewAccountRepo creates a new AccountRepo instance with an SQLite connection.
func NewAccountRepo(db *gorm.DB) ports.AccountRepo {
	return &AccountRepo{DB: db}
}

// CreateAccount creates an account
func (r *AccountRepo) CreateAccount(account *entity.Account) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.DB.Create(account).Error
}

// GetAccountByID gets account by account ID
func (r *AccountRepo) GetAccountByID(id string) (*entity.Account, error) {
	var account entity.Account
	err := r.DB.Where("id = ? AND status = ?", id, entity.AccountStatusValid).First(&account).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &account, err
}

// GetAccountByIDForUpdate gets account for update by account ID
func (r *AccountRepo) GetAccountByIDForUpdate(id string) (*entity.Account, error) {
	var account entity.Account

	tx := r.DB.Begin()
	defer tx.Rollback()

	err := tx.Set("gorm:query_option", "FOR UPDATE").
		Where("id = ? AND status = ?", id, entity.AccountStatusValid).
		First(&account).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &account, nil
}

// GetAccountByCustomerID gets account list by customer ID
func (r *AccountRepo) GetAccountByCustomerID(customerID string) ([]*entity.Account, error) {
	var accounts []*entity.Account
	err := r.DB.Where("customer_id = ? AND status = ?", customerID, entity.AccountStatusValid).Find(&accounts).Error
	return accounts, err
}

// UpdateAccount update account
func (r *AccountRepo) UpdateAccount(account *entity.Account) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	result := r.DB.Model(account).
		Where("id = ? AND version = ? AND status = ?", account.ID, account.Version, entity.AccountStatusValid).
		Updates(map[string]interface{}{
			"balance":               account.Balance,
			"active_status":         account.ActiveStatus,
			"updated_at":            account.UpdatedAt,
			"updated_by":            account.UpdatedBy,
			"locked_for_tx":         account.LockedForTx,
			"active_transaction_id": account.ActiveTransactionID,
			"version":               account.Version + 1,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return value.ErrConcurrentModification
	}

	return nil
}

// DeleteAccount deletes account by account ID
func (r *AccountRepo) DeleteAccount(id, requester string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	//return r.db.WithContext(ctx).Where("id = ?", id).Delete(&domain.Account{}).Error

	var account entity.Account
	if err := r.DB.Where("id = ? AND status = ?", id, entity.AccountStatusValid).First(&account).Error; err != nil {
		return err
	}

	account.Status = entity.AccountStatusInvalid
	account.UpdatedBy = requester
	if err := r.DB.Save(&account).Error; err != nil {
		return err
	}
	return nil
}

// DeleteAllAccountsByCustomerID deletes all accounts by customer id
func (r *AccountRepo) DeleteAllAccountsByCustomerID(customerID, requester string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.DB.Model(&entity.Account{}).
		Where("customer_id = ? AND status = ?", customerID, entity.AccountStatusValid).
		Updates(map[string]interface{}{
			"updated_by": requester,
			"status":     entity.AccountStatusInvalid,
		}).Error
}

func (r *AccountRepo) GetAccountsByFiltersWithPagination(filters map[string]interface{}, page, pageSize int) ([]*entity.Account, int64, error) {
	var accounts []*entity.Account
	var totalCount int64

	query := r.DB.Model(&entity.Account{})

	if status, ok := filters["status"]; ok {
		query = query.Where("status = ?", status)
	}

	if customerID, ok := filters["customer_id"]; ok {
		query = query.Where("customer_id = ?", customerID)
	}

	if minBalance, ok := filters["min_balance"]; ok {
		query = query.Where("balance >= ?", minBalance)
	}

	if lockedForTx, ok := filters["locked_for_tx"]; ok {
		query = query.Where("locked_for_tx = ?", lockedForTx)
	}

	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize

	err := query.
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&accounts).Error

	if err != nil {
		return nil, 0, err
	}

	return accounts, totalCount, nil
}

// UpdateAccountBalance update account balance by account ID
func (r *AccountRepo) UpdateAccountBalance(id string, newBalance float64, currentVersion int, requester string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	result := r.DB.Model(&entity.Account{}).
		Where("id = ? AND version = ? AND status = ?", id, currentVersion, entity.AccountStatusValid).
		Updates(map[string]interface{}{
			"balance":    newBalance,
			"updated_at": time.Now(),
			"updated_by": requester,
			"version":    currentVersion + 1,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return value.ErrConcurrentModification
	}

	return nil
}

// LockAccountForTransaction locks account while transacting balance by account ID and transaction id
func (r *AccountRepo) LockAccountForTransaction(id string, transactionID string) error {
	r.mu.Lock()
	tx := r.DB.Begin()

	defer func() {
		tx.Rollback()
		r.mu.Unlock()
	}()

	// Lock the account for update
	var account entity.Account
	err := tx.Set("gorm:query_option", "FOR UPDATE").
		Where("id = ? AND status = ?", id, entity.AccountStatusValid).
		First(&account).Error

	if err != nil {
		return err
	}

	// Check if already locked
	if account.LockedForTx {
		return value.ErrAccountLocked
	}

	// Apply lock
	result := tx.Model(&entity.Account{}).
		Where("id = ? AND version = ? AND status = ?", id, account.Version, entity.AccountStatusValid).
		Updates(map[string]interface{}{
			"locked_for_tx":         true,
			"active_transaction_id": transactionID,
			"version":               account.Version + 1,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return value.ErrConcurrentModification
	}

	return tx.Commit().Error
}

// UnlockAccountFromTransaction unlocks an account while transacting balance by account ID
func (r *AccountRepo) UnlockAccountFromTransaction(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	result := r.DB.Model(&entity.Account{}).
		Where("id = ? AND status = ?", id, entity.AccountStatusValid).
		Updates(map[string]interface{}{
			"locked_for_tx":         false,
			"active_transaction_id": nil,
		})

	return result.Error
}

// CheckTransactionLock checks if account in transaction or not by account ID
func (r *AccountRepo) CheckTransactionLock(id string) error {
	account, err := r.GetAccountByID(id)
	if err != nil {
		return err
	}
	if account == nil {
		return value.ErrAccountNotFound
	}

	if account.LockedForTx || account.ActiveTransactionID != nil {
		return value.ErrAccountLocked
	}

	return nil
}

// GetAccountsInTransaction get all the accounts by transaction ID
func (r *AccountRepo) GetAccountsInTransaction(transactionID string) ([]*entity.Account, error) {
	var accounts []*entity.Account
	err := r.DB.
		Where("active_transaction_id = ? AND status = ?", transactionID, entity.AccountStatusValid).
		Find(&accounts).Error
	return accounts, err
}

// GetCustomerAccountsInTransaction get all the accounts that are in transaction by customer id
func (r *AccountRepo) GetCustomerAccountsInTransaction(customerID string) ([]*entity.Account, error) {
	var accounts []*entity.Account
	err := r.DB.
		Where("customer_id = ? AND locked_for_tx = ? AND status = ?", customerID, true, entity.AccountStatusValid).
		Find(&accounts).Error
	return accounts, err
}

// GetCustomerAccountsInTransactionOrHasBalance get all the customer accounts either in transaction or has balance
func (r *AccountRepo) GetCustomerAccountsInTransactionOrHasBalance(customerID string) ([]*entity.Account, error) {
	var accounts []*entity.Account
	err := r.DB.
		Where("customer_id = ? AND status = ?", customerID, entity.AccountStatusValid).
		Where("locked_for_tx = ? OR balance > ?", true, 0).
		Find(&accounts).Error
	return accounts, err
}

// ForceUnlockAllAccounts unlock all the accounts by transaction ID
func (r *AccountRepo) ForceUnlockAllAccounts(transactionID string) error {
	return r.DB.Model(&entity.Account{}).
		Where("active_transaction_id = ? AND status = ?", transactionID, entity.AccountStatusValid).
		Updates(map[string]interface{}{
			"locked_for_tx":         false,
			"active_transaction_id": nil,
		}).Error
}

// IncrementVersion increment version by account id
func (r *AccountRepo) IncrementVersion(id string, currentVersion int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	result := r.DB.Model(&entity.Account{}).
		Where("id = ? AND version = ? AND status = ?", id, currentVersion, entity.AccountStatusValid).
		Update("version", currentVersion+1)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return value.ErrConcurrentModification
	}

	return nil
}
