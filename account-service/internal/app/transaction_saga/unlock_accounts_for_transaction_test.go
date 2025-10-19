package transaction_saga

import (
	custom_err "account-service/internal/domain/error"
	mock_repo "account-service/internal/ports/mocks/repo"
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestUnlockAccountsForTransaction_Execute_Success tests success response if all inputs are properly provided
func TestUnlockAccountsForTransaction_Execute_Success(t *testing.T) {
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	unlockAccountsForTransaction := NewUnlockAccountsForTransaction(mockAccountRepo)

	transactionID := "tx-123"
	requester := "user123"
	requestID := "req-456"

	mockAccountRepo.On("UnlockAccountsForTransaction", transactionID).Return(nil)

	message, err := unlockAccountsForTransaction.Execute(transactionID, requester, requestID)

	assert.NoError(t, err)
	assert.Equal(t, "Accounts unlocked successfully", message)

	mockAccountRepo.AssertExpectations(t)
}

// TestUnlockAccountsForTransaction_Execute_EmptyTransactionID tests error response when transaction ID is empty
func TestUnlockAccountsForTransaction_Execute_EmptyTransactionID(t *testing.T) {
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	unlockAccountsForTransaction := NewUnlockAccountsForTransaction(mockAccountRepo)

	transactionID := ""
	requester := "user123"
	requestID := "req-456"

	message, err := unlockAccountsForTransaction.Execute(transactionID, requester, requestID)

	assert.Error(t, err)
	assert.Equal(t, "transaction id is required", message)
	assert.Equal(t, custom_err.ErrTransactionIdRequired, err)

	mockAccountRepo.AssertNotCalled(t, "UnlockAccountsForTransaction")
}

// TestUnlockAccountsForTransaction_Execute_WhitespaceTransactionID tests error response when transaction ID is whitespace
func TestUnlockAccountsForTransaction_Execute_WhitespaceTransactionID(t *testing.T) {
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	unlockAccountsForTransaction := NewUnlockAccountsForTransaction(mockAccountRepo)

	transactionID := "   "
	requester := "user123"
	requestID := "req-456"

	message, err := unlockAccountsForTransaction.Execute(transactionID, requester, requestID)

	assert.Error(t, err)
	assert.Equal(t, "transaction id is required", message)
	assert.Equal(t, custom_err.ErrTransactionIdRequired, err)

	mockAccountRepo.AssertNotCalled(t, "UnlockAccountsForTransaction")
}

// TestUnlockAccountsForTransaction_Execute_DatabaseError tests error response when repository returns an error
func TestUnlockAccountsForTransaction_Execute_DatabaseError(t *testing.T) {
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	unlockAccountsForTransaction := NewUnlockAccountsForTransaction(mockAccountRepo)

	transactionID := "tx-123"
	requester := "user123"
	requestID := "req-456"

	mockAccountRepo.On("UnlockAccountsForTransaction", transactionID).Return(errors.New("database connection failed"))

	message, err := unlockAccountsForTransaction.Execute(transactionID, requester, requestID)

	assert.Error(t, err)
	assert.Equal(t, "failed to unlock accounts for transaction", message)
	assert.Equal(t, custom_err.ErrDatabase, err)

	mockAccountRepo.AssertExpectations(t)
}

// TestUnlockAccountsForTransaction_Execute_LongTransactionID tests success response with long transaction ID
func TestUnlockAccountsForTransaction_Execute_LongTransactionID(t *testing.T) {
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	unlockAccountsForTransaction := NewUnlockAccountsForTransaction(mockAccountRepo)

	transactionID := "tx-very-long-transaction-id-1234567890-abcdefghijklmnopqrstuvwxyz"
	requester := "user123"
	requestID := "req-456"

	mockAccountRepo.On("UnlockAccountsForTransaction", transactionID).Return(nil)

	message, err := unlockAccountsForTransaction.Execute(transactionID, requester, requestID)

	assert.NoError(t, err)
	assert.Equal(t, "Accounts unlocked successfully", message)

	mockAccountRepo.AssertExpectations(t)
}

// TestUnlockAccountsForTransaction_Execute_SpecialCharacters tests success response with special characters in transaction ID
func TestUnlockAccountsForTransaction_Execute_SpecialCharacters(t *testing.T) {
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	unlockAccountsForTransaction := NewUnlockAccountsForTransaction(mockAccountRepo)

	transactionID := "tx-123-特殊字符-🚀"
	requester := "user123"
	requestID := "req-456"

	mockAccountRepo.On("UnlockAccountsForTransaction", transactionID).Return(nil)

	message, err := unlockAccountsForTransaction.Execute(transactionID, requester, requestID)

	assert.NoError(t, err)
	assert.Equal(t, "Accounts unlocked successfully", message)

	mockAccountRepo.AssertExpectations(t)
}

// TestUnlockAccountsForTransaction_Execute_EmptyRequester tests success response when requester is empty (if allowed by business logic)
func TestUnlockAccountsForTransaction_Execute_EmptyRequester(t *testing.T) {
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	unlockAccountsForTransaction := NewUnlockAccountsForTransaction(mockAccountRepo)

	transactionID := "tx-123"
	requester := ""
	requestID := "req-456"

	mockAccountRepo.On("UnlockAccountsForTransaction", transactionID).Return(nil)

	message, err := unlockAccountsForTransaction.Execute(transactionID, requester, requestID)

	assert.NoError(t, err)
	assert.Equal(t, "Accounts unlocked successfully", message)

	mockAccountRepo.AssertExpectations(t)
}
