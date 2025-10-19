package transaction_saga

import (
	"account-service/internal/domain/entity"
	custom_err "account-service/internal/domain/error"
	"account-service/internal/grpc/types"
	mock_repo "account-service/internal/ports/mocks/repo"
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestUpdateAccountBalanceForTransaction_Execute_Success tests success response if all inputs are properly provided
func TestUpdateAccountBalanceForTransaction_Execute_Success(t *testing.T) {
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	updateAccountBalanceForTransaction := NewUpdateAccountBalanceForTransaction(mockAccountRepo)

	accountBalanceUpdates := []types.AccountBalance{
		{AccountID: "acc-1", Balance: 1500.0, Version: 1},
		{AccountID: "acc-2", Balance: 2500.0, Version: 2},
	}
	requester := "user123"

	expectedResponses := []types.AccountBalanceResponse{
		{AccountID: "acc-1", Version: 2},
		{AccountID: "acc-2", Version: 3},
	}

	// Mock account existence checks
	mockAccountRepo.On("GetAccountByID", "acc-1").Return(&entity.Account{ID: "acc-1", Version: 1}, nil)
	mockAccountRepo.On("GetAccountByID", "acc-2").Return(&entity.Account{ID: "acc-2", Version: 2}, nil)

	// Mock the actual balance update
	mockAccountRepo.On("UpdateAccountBalanceLifecycle", accountBalanceUpdates, requester).Return(expectedResponses, nil)

	responses, message, err := updateAccountBalanceForTransaction.Execute(accountBalanceUpdates, requester)

	assert.NoError(t, err)
	assert.Equal(t, "account balances updated successfully", message)
	assert.Equal(t, expectedResponses, responses)

	mockAccountRepo.AssertExpectations(t)
}

// TestUpdateAccountBalanceForTransaction_Execute_EmptyUpdates tests error response when account balance updates are empty
func TestUpdateAccountBalanceForTransaction_Execute_EmptyUpdates(t *testing.T) {
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	updateAccountBalanceForTransaction := NewUpdateAccountBalanceForTransaction(mockAccountRepo)

	var accountBalanceUpdates []types.AccountBalance
	requester := "user123"

	responses, message, err := updateAccountBalanceForTransaction.Execute(accountBalanceUpdates, requester)

	assert.Error(t, err)
	assert.Equal(t, "at least one account balance update is required", message)
	assert.Nil(t, responses)

	mockAccountRepo.AssertNotCalled(t, "GetAccountByID")
	mockAccountRepo.AssertNotCalled(t, "UpdateAccountBalanceLifecycle")
}

// TestUpdateAccountBalanceForTransaction_Execute_NilUpdates tests error response when account balance updates is nil
func TestUpdateAccountBalanceForTransaction_Execute_NilUpdates(t *testing.T) {
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	updateAccountBalanceForTransaction := NewUpdateAccountBalanceForTransaction(mockAccountRepo)

	var accountBalanceUpdates []types.AccountBalance = nil
	requester := "user123"

	responses, message, err := updateAccountBalanceForTransaction.Execute(accountBalanceUpdates, requester)

	assert.Error(t, err)
	assert.Equal(t, "at least one account balance update is required", message)
	assert.Nil(t, responses)

	mockAccountRepo.AssertNotCalled(t, "GetAccountByID")
	mockAccountRepo.AssertNotCalled(t, "UpdateAccountBalanceLifecycle")
}

// TestUpdateAccountBalanceForTransaction_Execute_AccountNotFound tests error response when an account is not found
func TestUpdateAccountBalanceForTransaction_Execute_AccountNotFound(t *testing.T) {
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	updateAccountBalanceForTransaction := NewUpdateAccountBalanceForTransaction(mockAccountRepo)

	accountBalanceUpdates := []types.AccountBalance{
		{AccountID: "acc-1", Balance: 1500.0, Version: 1},
		{AccountID: "acc-nonexistent", Balance: 2500.0, Version: 2},
	}
	requester := "user123"

	// First account exists, second doesn't
	mockAccountRepo.On("GetAccountByID", "acc-1").Return(&entity.Account{ID: "acc-1", Version: 1}, nil)
	mockAccountRepo.On("GetAccountByID", "acc-nonexistent").Return(nil, nil)

	responses, message, err := updateAccountBalanceForTransaction.Execute(accountBalanceUpdates, requester)

	assert.Error(t, err)
	assert.Equal(t, "Account not found", message)
	assert.Equal(t, custom_err.ErrAccountNotFound, err)
	assert.Nil(t, responses)

	mockAccountRepo.AssertExpectations(t)
	mockAccountRepo.AssertNotCalled(t, "UpdateAccountBalanceLifecycle")
}

// TestUpdateAccountBalanceForTransaction_Execute_GetAccountError tests error response when GetAccountByID returns an error
func TestUpdateAccountBalanceForTransaction_Execute_GetAccountError(t *testing.T) {
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	updateAccountBalanceForTransaction := NewUpdateAccountBalanceForTransaction(mockAccountRepo)

	accountBalanceUpdates := []types.AccountBalance{
		{AccountID: "acc-1", Balance: 1500.0, Version: 1},
	}
	requester := "user123"

	mockAccountRepo.On("GetAccountByID", "acc-1").Return(nil, errors.New("database error"))

	responses, message, err := updateAccountBalanceForTransaction.Execute(accountBalanceUpdates, requester)

	assert.Error(t, err)
	assert.Equal(t, "failed to lock accounts for transaction", message)
	assert.Equal(t, custom_err.ErrDatabase, err)
	assert.Nil(t, responses)

	mockAccountRepo.AssertExpectations(t)
	mockAccountRepo.AssertNotCalled(t, "UpdateAccountBalanceLifecycle")
}

// TestUpdateAccountBalanceForTransaction_Execute_UpdateBalanceError tests error response when UpdateAccountBalanceLifecycle returns an error
func TestUpdateAccountBalanceForTransaction_Execute_UpdateBalanceError(t *testing.T) {
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	updateAccountBalanceForTransaction := NewUpdateAccountBalanceForTransaction(mockAccountRepo)

	accountBalanceUpdates := []types.AccountBalance{
		{AccountID: "acc-1", Balance: 1500.0, Version: 1},
		{AccountID: "acc-2", Balance: 2500.0, Version: 2},
	}
	requester := "user123"

	// Mock account existence checks
	mockAccountRepo.On("GetAccountByID", "acc-1").Return(&entity.Account{ID: "acc-1", Version: 1}, nil)
	mockAccountRepo.On("GetAccountByID", "acc-2").Return(&entity.Account{ID: "acc-2", Version: 2}, nil)

	// Mock the actual balance update to return error
	mockAccountRepo.On("UpdateAccountBalanceLifecycle", accountBalanceUpdates, requester).Return(nil, errors.New("update failed"))

	responses, message, err := updateAccountBalanceForTransaction.Execute(accountBalanceUpdates, requester)

	assert.Error(t, err)
	assert.Equal(t, "failed to update account balance", message)
	assert.Equal(t, custom_err.ErrDatabase, err)
	assert.Nil(t, responses)

	mockAccountRepo.AssertExpectations(t)
}

// TestUpdateAccountBalanceForTransaction_Execute_SingleUpdate tests success response with single account balance update
func TestUpdateAccountBalanceForTransaction_Execute_SingleUpdate(t *testing.T) {
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	updateAccountBalanceForTransaction := NewUpdateAccountBalanceForTransaction(mockAccountRepo)

	accountBalanceUpdates := []types.AccountBalance{
		{AccountID: "acc-1", Balance: 1500.0, Version: 1},
	}
	requester := "user123"

	expectedResponses := []types.AccountBalanceResponse{
		{AccountID: "acc-1", Version: 2},
	}

	mockAccountRepo.On("GetAccountByID", "acc-1").Return(&entity.Account{ID: "acc-1", Version: 1}, nil)
	mockAccountRepo.On("UpdateAccountBalanceLifecycle", accountBalanceUpdates, requester).Return(expectedResponses, nil)

	responses, message, err := updateAccountBalanceForTransaction.Execute(accountBalanceUpdates, requester)

	assert.NoError(t, err)
	assert.Equal(t, "account balances updated successfully", message)
	assert.Equal(t, expectedResponses, responses)

	mockAccountRepo.AssertExpectations(t)
}
