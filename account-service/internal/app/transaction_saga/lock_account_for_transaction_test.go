package transaction_saga

import (
	custom_err "account-service/internal/domain/error"
	mock_repo "account-service/internal/ports/mocks/repo"
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestLockAccountForTransaction_Execute_Success tests success response if all inputs are properly provided
func TestLockAccountForTransaction_Execute_Success(t *testing.T) {
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	lockAccountForTransaction := NewLockAccountForTransaction(mockAccountRepo)

	transactionID := "tx-123"
	accountIDs := []string{"acc-1", "acc-2", "acc-3"}
	requester := "user123"
	requestID := "req-456"

	mockAccountRepo.On("LockAccountsForTransaction", transactionID, accountIDs).Return(nil)

	message, err := lockAccountForTransaction.Execute(transactionID, accountIDs, requester, requestID)

	assert.NoError(t, err)
	assert.Equal(t, "Accounts locked successfully", message)

	mockAccountRepo.AssertExpectations(t)
}

// TestLockAccountForTransaction_Execute_EmptyAccountIDs tests error response when account IDs are empty
func TestLockAccountForTransaction_Execute_EmptyAccountIDs(t *testing.T) {
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	lockAccountForTransaction := NewLockAccountForTransaction(mockAccountRepo)

	transactionID := "tx-123"
	accountIDs := []string{}
	requester := "user123"
	requestID := "req-456"

	message, err := lockAccountForTransaction.Execute(transactionID, accountIDs, requester, requestID)

	assert.Error(t, err)
	assert.Equal(t, "at least one account ID is required", message)
	assert.Equal(t, custom_err.ErrMinimumOneAccountIdRequired, err)

	mockAccountRepo.AssertNotCalled(t, "LockAccountsForTransaction")
}

// TestLockAccountForTransaction_Execute_EmptyTransactionID tests error response when transaction ID is empty
func TestLockAccountForTransaction_Execute_EmptyTransactionID(t *testing.T) {
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	lockAccountForTransaction := NewLockAccountForTransaction(mockAccountRepo)

	transactionID := ""
	accountIDs := []string{"acc-1", "acc-2"}
	requester := "user123"
	requestID := "req-456"

	message, err := lockAccountForTransaction.Execute(transactionID, accountIDs, requester, requestID)

	assert.Error(t, err)
	assert.Equal(t, "transaction id is required", message)
	assert.Equal(t, custom_err.ErrTransactionIdRequired, err)

	mockAccountRepo.AssertNotCalled(t, "LockAccountsForTransaction")
}

// TestLockAccountForTransaction_Execute_WhitespaceTransactionID tests error response when transaction ID is whitespace
func TestLockAccountForTransaction_Execute_WhitespaceTransactionID(t *testing.T) {
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	lockAccountForTransaction := NewLockAccountForTransaction(mockAccountRepo)

	transactionID := "   "
	accountIDs := []string{"acc-1", "acc-2"}
	requester := "user123"
	requestID := "req-456"

	message, err := lockAccountForTransaction.Execute(transactionID, accountIDs, requester, requestID)

	assert.Error(t, err)
	assert.Equal(t, "transaction id is required", message)
	assert.Equal(t, custom_err.ErrTransactionIdRequired, err)

	mockAccountRepo.AssertNotCalled(t, "LockAccountsForTransaction")
}

// TestLockAccountForTransaction_Execute_DatabaseError tests error response when repository returns an error
func TestLockAccountForTransaction_Execute_DatabaseError(t *testing.T) {
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	lockAccountForTransaction := NewLockAccountForTransaction(mockAccountRepo)

	transactionID := "tx-123"
	accountIDs := []string{"acc-1", "acc-2"}
	requester := "user123"
	requestID := "req-456"

	mockAccountRepo.On("LockAccountsForTransaction", transactionID, accountIDs).Return(errors.New("database connection failed"))

	message, err := lockAccountForTransaction.Execute(transactionID, accountIDs, requester, requestID)

	assert.Error(t, err)
	assert.Equal(t, "failed to lock accounts for transaction", message)
	assert.Equal(t, custom_err.ErrDatabase, err)

	mockAccountRepo.AssertExpectations(t)
}

// TestLockAccountForTransaction_Execute_SingleAccountID tests success response with single account ID
func TestLockAccountForTransaction_Execute_SingleAccountID(t *testing.T) {
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	lockAccountForTransaction := NewLockAccountForTransaction(mockAccountRepo)

	transactionID := "tx-123"
	accountIDs := []string{"acc-1"}
	requester := "user123"
	requestID := "req-456"

	mockAccountRepo.On("LockAccountsForTransaction", transactionID, accountIDs).Return(nil)

	message, err := lockAccountForTransaction.Execute(transactionID, accountIDs, requester, requestID)

	assert.NoError(t, err)
	assert.Equal(t, "Accounts locked successfully", message)

	mockAccountRepo.AssertExpectations(t)
}

// TestLockAccountForTransaction_Execute_MultipleAccountIDs tests success response with multiple account IDs
func TestLockAccountForTransaction_Execute_MultipleAccountIDs(t *testing.T) {
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	lockAccountForTransaction := NewLockAccountForTransaction(mockAccountRepo)

	transactionID := "tx-123"
	accountIDs := []string{"acc-1", "acc-2", "acc-3", "acc-4", "acc-5"}
	requester := "user123"
	requestID := "req-456"

	mockAccountRepo.On("LockAccountsForTransaction", transactionID, accountIDs).Return(nil)

	message, err := lockAccountForTransaction.Execute(transactionID, accountIDs, requester, requestID)

	assert.NoError(t, err)
	assert.Equal(t, "Accounts locked successfully", message)

	mockAccountRepo.AssertExpectations(t)
}
