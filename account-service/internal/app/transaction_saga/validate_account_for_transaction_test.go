package transaction_saga

import (
	"account-service/internal/domain/entity"
	custom_err "account-service/internal/domain/error"
	mock_repo "account-service/internal/ports/mocks/repo"
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestValidateAccountForTransaction_Execute_Success tests success response if all inputs are properly provided
func TestValidateAccountForTransaction_Execute_Success(t *testing.T) {
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	validateAccountForTransaction := NewValidateAccountForTransaction(mockAccountRepo)

	transactionID := "tx-123"
	accountIDs := []string{"acc-1", "acc-2", "acc-3"}
	requester := "user123"
	requestID := "req-456"

	// Create accounts that can transact (all CanTransact() conditions must be true)
	validAccounts := []*entity.Account{
		{
			ID:                  "acc-1",
			Status:              entity.AccountStatusValid,
			ActiveStatus:        "active",
			LockedForTx:         false,
			ActiveTransactionID: nil,
			Balance:             1000.0,
		},
		{
			ID:                  "acc-2",
			Status:              entity.AccountStatusValid,
			ActiveStatus:        "active",
			LockedForTx:         false,
			ActiveTransactionID: nil,
			Balance:             2000.0,
		},
		{
			ID:                  "acc-3",
			Status:              entity.AccountStatusValid,
			ActiveStatus:        "active",
			LockedForTx:         false,
			ActiveTransactionID: nil,
			Balance:             3000.0,
		},
	}

	mockAccountRepo.On("GetAccountByID", "acc-1").Return(validAccounts[0], nil)
	mockAccountRepo.On("GetAccountByID", "acc-2").Return(validAccounts[1], nil)
	mockAccountRepo.On("GetAccountByID", "acc-3").Return(validAccounts[2], nil)

	accounts, message, err := validateAccountForTransaction.Execute(transactionID, accountIDs, requester, requestID)

	assert.NoError(t, err)
	assert.Equal(t, "All accounts are valid for transaction", message)
	assert.Equal(t, validAccounts, accounts)

	mockAccountRepo.AssertExpectations(t)
}

// TestValidateAccountForTransaction_Execute_EmptyAccountIDs tests error response when account IDs are empty
func TestValidateAccountForTransaction_Execute_EmptyAccountIDs(t *testing.T) {
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	validateAccountForTransaction := NewValidateAccountForTransaction(mockAccountRepo)

	transactionID := "tx-123"
	accountIDs := []string{}
	requester := "user123"
	requestID := "req-456"

	accounts, message, err := validateAccountForTransaction.Execute(transactionID, accountIDs, requester, requestID)

	assert.Error(t, err)
	assert.Equal(t, "at least one account ID is required", message)
	assert.Equal(t, custom_err.ErrMinimumOneAccountIdRequired, err)
	assert.Nil(t, accounts)

	mockAccountRepo.AssertNotCalled(t, "GetAccountByID")
}

// TestValidateAccountForTransaction_Execute_NilAccountIDs tests error response when account IDs is nil
func TestValidateAccountForTransaction_Execute_NilAccountIDs(t *testing.T) {
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	validateAccountForTransaction := NewValidateAccountForTransaction(mockAccountRepo)

	transactionID := "tx-123"
	var accountIDs []string = nil
	requester := "user123"
	requestID := "req-456"

	accounts, message, err := validateAccountForTransaction.Execute(transactionID, accountIDs, requester, requestID)

	assert.Error(t, err)
	assert.Equal(t, "at least one account ID is required", message)
	assert.Equal(t, custom_err.ErrMinimumOneAccountIdRequired, err)
	assert.Nil(t, accounts)

	mockAccountRepo.AssertNotCalled(t, "GetAccountByID")
}

// TestValidateAccountForTransaction_Execute_AccountNotFound tests error response when an account is not found
func TestValidateAccountForTransaction_Execute_AccountNotFound(t *testing.T) {
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	validateAccountForTransaction := NewValidateAccountForTransaction(mockAccountRepo)

	transactionID := "tx-123"
	accountIDs := []string{"acc-1", "acc-nonexistent"}
	requester := "user123"
	requestID := "req-456"

	validAccount := &entity.Account{
		ID:                  "acc-1",
		Status:              entity.AccountStatusValid,
		ActiveStatus:        "active",
		LockedForTx:         false,
		ActiveTransactionID: nil,
		Balance:             1000.0,
	}

	// First account exists, second doesn't
	mockAccountRepo.On("GetAccountByID", "acc-1").Return(validAccount, nil)
	mockAccountRepo.On("GetAccountByID", "acc-nonexistent").Return(nil, nil)

	accounts, message, err := validateAccountForTransaction.Execute(transactionID, accountIDs, requester, requestID)

	assert.Error(t, err)
	assert.Equal(t, "Account 'acc-nonexistent' not found", message)
	assert.Equal(t, custom_err.ErrAccountNotFound, err)
	assert.Nil(t, accounts)

	mockAccountRepo.AssertExpectations(t)
}

// TestValidateAccountForTransaction_Execute_GetAccountError tests error response when GetAccountByID returns an error
func TestValidateAccountForTransaction_Execute_GetAccountError(t *testing.T) {
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	validateAccountForTransaction := NewValidateAccountForTransaction(mockAccountRepo)

	transactionID := "tx-123"
	accountIDs := []string{"acc-1"}
	requester := "user123"
	requestID := "req-456"

	mockAccountRepo.On("GetAccountByID", "acc-1").Return(nil, errors.New("database error"))

	accounts, message, err := validateAccountForTransaction.Execute(transactionID, accountIDs, requester, requestID)

	assert.Error(t, err)
	assert.Equal(t, "Failed to validate account'acc-1'", message)
	assert.Equal(t, custom_err.ErrDatabase, err)
	assert.Nil(t, accounts)

	mockAccountRepo.AssertExpectations(t)
}

// TestValidateAccountForTransaction_Execute_AccountLockedForTx tests error response when account is locked for transaction
func TestValidateAccountForTransaction_Execute_AccountLockedForTx(t *testing.T) {
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	validateAccountForTransaction := NewValidateAccountForTransaction(mockAccountRepo)

	transactionID := "tx-123"
	accountIDs := []string{"acc-1", "acc-locked"}
	requester := "user123"
	requestID := "req-456"

	validAccount := &entity.Account{
		ID:                  "acc-1",
		Status:              entity.AccountStatusValid,
		ActiveStatus:        "active",
		LockedForTx:         false,
		ActiveTransactionID: nil,
		Balance:             1000.0,
	}

	// Create a locked account that cannot transact
	lockedAccount := &entity.Account{
		ID:                  "acc-locked",
		Status:              entity.AccountStatusValid,
		ActiveStatus:        "active",
		LockedForTx:         true, // This makes CanTransact() return false
		ActiveTransactionID: nil,
		Balance:             2000.0,
	}

	// First account is valid, second is locked
	mockAccountRepo.On("GetAccountByID", "acc-1").Return(validAccount, nil)
	mockAccountRepo.On("GetAccountByID", "acc-locked").Return(lockedAccount, nil)

	accounts, message, err := validateAccountForTransaction.Execute(transactionID, accountIDs, requester, requestID)

	assert.Error(t, err)
	assert.Equal(t, "Account 'acc-locked' cannot transact", message)
	assert.Equal(t, custom_err.ErrAccountLocked, err)
	assert.Nil(t, accounts)

	mockAccountRepo.AssertExpectations(t)
}

// TestValidateAccountForTransaction_Execute_AccountInactiveStatus tests error response when account has inactive status
func TestValidateAccountForTransaction_Execute_AccountInactiveStatus(t *testing.T) {
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	validateAccountForTransaction := NewValidateAccountForTransaction(mockAccountRepo)

	transactionID := "tx-123"
	accountIDs := []string{"acc-inactive"}
	requester := "user123"
	requestID := "req-456"

	inactiveAccount := &entity.Account{
		ID:                  "acc-inactive",
		Status:              entity.AccountStatusValid,
		ActiveStatus:        "inactive", // This makes CanTransact() return false
		LockedForTx:         false,
		ActiveTransactionID: nil,
		Balance:             1000.0,
	}

	mockAccountRepo.On("GetAccountByID", "acc-inactive").Return(inactiveAccount, nil)

	accounts, message, err := validateAccountForTransaction.Execute(transactionID, accountIDs, requester, requestID)

	assert.Error(t, err)
	assert.Equal(t, "Account 'acc-inactive' cannot transact", message)
	assert.Equal(t, custom_err.ErrAccountLocked, err)
	assert.Nil(t, accounts)

	mockAccountRepo.AssertExpectations(t)
}

// TestValidateAccountForTransaction_Execute_AccountWithActiveTransaction tests error response when account has active transaction
func TestValidateAccountForTransaction_Execute_AccountWithActiveTransaction(t *testing.T) {
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	validateAccountForTransaction := NewValidateAccountForTransaction(mockAccountRepo)

	transactionID := "tx-123"
	accountIDs := []string{"acc-active-tx"}
	requester := "user123"
	requestID := "req-456"

	activeTxID := "tx-existing"
	accountWithActiveTx := &entity.Account{
		ID:                  "acc-active-tx",
		Status:              entity.AccountStatusValid,
		ActiveStatus:        "active",
		LockedForTx:         false,
		ActiveTransactionID: &activeTxID, // This makes CanTransact() return false
		Balance:             1000.0,
	}

	mockAccountRepo.On("GetAccountByID", "acc-active-tx").Return(accountWithActiveTx, nil)

	accounts, message, err := validateAccountForTransaction.Execute(transactionID, accountIDs, requester, requestID)

	assert.Error(t, err)
	assert.Equal(t, "Account 'acc-active-tx' cannot transact", message)
	assert.Equal(t, custom_err.ErrAccountLocked, err)
	assert.Nil(t, accounts)

	mockAccountRepo.AssertExpectations(t)
}
