package account

import (
	"account-service/internal/domain/entity"
	"account-service/internal/domain/value"
	mock_repo "account-service/internal/ports/mocks/repo"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

// TestDeleteAccount_Execute_Success_SingleAccount test success on single account delete if scope is single and an account is deleted using accountID
func TestDeleteAccount_Execute_Success_SingleAccount(t *testing.T) {
	mockCustomerRepo := new(mock_repo.MockCustomerRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	deleteAccount := NewDeleteAccount(mockAccountRepo, mockCustomerRepo, mockEventRepo)

	scope := "single"
	id := "acc-123"
	requester := "user123"
	requestId := "req-456"

	account := &entity.Account{
		ID:         "acc-123",
		CustomerID: "cust-123",
		Balance:    0.0,
	}

	mockAccountRepo.On("GetAccountByID", id).Return(account, nil)
	mockAccountRepo.On("CheckTransactionLock", id).Return(nil)
	mockAccountRepo.On("DeleteAccount", id, requester).Return(nil)
	mockEventRepo.On("CreateEvent", mock.AnythingOfType("*entity.Event")).Return(nil)

	message, err := deleteAccount.Execute(scope, id, requester, requestId)

	assert.NoError(t, err)
	assert.Equal(t, "Accounts deleted successfully", message)

	mockAccountRepo.AssertExpectations(t)
	mockEventRepo.AssertExpectations(t)
}

// TestDeleteAccount_Execute_Success_AllAccounts test success on all accounts delete if scope is 'all' and all accounts are deleted using companyID
func TestDeleteAccount_Execute_Success_AllAccounts(t *testing.T) {
	mockCustomerRepo := new(mock_repo.MockCustomerRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	deleteAccount := NewDeleteAccount(mockAccountRepo, mockCustomerRepo, mockEventRepo)

	scope := "all"
	id := "cust-123"
	requester := "user123"
	requestId := "req-456"

	mockCustomerRepo.On("Exists", id).Return(true, nil)
	mockAccountRepo.On("GetCustomerAccountsInTransactionOrHasBalance", id).Return([]*entity.Account{}, nil)
	mockAccountRepo.On("DeleteAllAccountsByCustomerID", id, requester).Return(nil)
	mockEventRepo.On("CreateEvent", mock.AnythingOfType("*entity.Event")).Return(nil)

	message, err := deleteAccount.Execute(scope, id, requester, requestId)

	assert.NoError(t, err)
	assert.Equal(t, "Accounts deleted successfully", message)

	mockCustomerRepo.AssertExpectations(t)
	mockAccountRepo.AssertExpectations(t)
	mockEventRepo.AssertExpectations(t)
}

// TestDeleteAccount_Execute_InvalidScope test if scope is not all/single
func TestDeleteAccount_Execute_InvalidScope(t *testing.T) {
	mockCustomerRepo := new(mock_repo.MockCustomerRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	deleteAccount := NewDeleteAccount(mockAccountRepo, mockCustomerRepo, mockEventRepo)

	message, err := deleteAccount.Execute("invalid", "acc-123", "user123", "req-456")

	assert.Error(t, err)
	assert.ErrorIs(t, err, value.ErrValidationFailed)
	assert.Equal(t, "Invalid request - 'scope' missing or invalid", message)

	mockCustomerRepo.AssertNotCalled(t, "Exists")
	mockAccountRepo.AssertNotCalled(t, "GetAccountByID")
}

// TestDeleteAccount_Execute_EmptyScope if scope is empty
func TestDeleteAccount_Execute_EmptyScope(t *testing.T) {
	mockCustomerRepo := new(mock_repo.MockCustomerRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	deleteAccount := NewDeleteAccount(mockAccountRepo, mockCustomerRepo, mockEventRepo)

	message, err := deleteAccount.Execute("", "acc-123", "user123", "req-456")

	assert.Error(t, err)
	assert.ErrorIs(t, err, value.ErrValidationFailed)
	assert.Equal(t, "Invalid request - 'scope' missing or invalid", message)

	mockCustomerRepo.AssertNotCalled(t, "Exists")
	mockAccountRepo.AssertNotCalled(t, "GetAccountByID")
}

// TestDeleteAccount_Execute_EmptyID_SingleScope tests if account id is empty when scope = single
func TestDeleteAccount_Execute_EmptyID_SingleScope(t *testing.T) {
	mockCustomerRepo := new(mock_repo.MockCustomerRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	deleteAccount := NewDeleteAccount(mockAccountRepo, mockCustomerRepo, mockEventRepo)

	message, err := deleteAccount.Execute("single", "", "user123", "req-456")

	assert.Error(t, err)
	assert.ErrorIs(t, err, value.ErrValidationFailed)
	assert.Equal(t, "Invalid request - 'id' account id missing", message)

	mockCustomerRepo.AssertNotCalled(t, "Exists")
	mockAccountRepo.AssertNotCalled(t, "GetAccountByID")
}

// TestDeleteAccount_Execute_EmptyRequester tests if requester not provided
func TestDeleteAccount_Execute_EmptyRequester(t *testing.T) {
	mockCustomerRepo := new(mock_repo.MockCustomerRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	deleteAccount := NewDeleteAccount(mockAccountRepo, mockCustomerRepo, mockEventRepo)

	message, err := deleteAccount.Execute("single", "acc-123", "", "req-456")

	assert.Error(t, err)
	assert.ErrorIs(t, err, value.ErrValidationFailed)
	assert.Equal(t, "Unknown requester", message)

	mockCustomerRepo.AssertNotCalled(t, "Exists")
	mockAccountRepo.AssertNotCalled(t, "GetAccountByID")
}

// TestDeleteAccount_Execute_CustomerNotFound_AllScope tests if customer exists with provided id when scope=all
func TestDeleteAccount_Execute_CustomerNotFound_AllScope(t *testing.T) {
	mockCustomerRepo := new(mock_repo.MockCustomerRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	deleteAccount := NewDeleteAccount(mockAccountRepo, mockCustomerRepo, mockEventRepo)

	scope := "all"
	id := "cust-123"
	requester := "user123"
	requestId := "req-456"

	mockCustomerRepo.On("Exists", id).Return(false, nil)

	message, err := deleteAccount.Execute(scope, id, requester, requestId)

	assert.Error(t, err)
	assert.ErrorIs(t, err, value.ErrCustomerNotFound)
	assert.Equal(t, "Customer not found", message)

	mockCustomerRepo.AssertExpectations(t)
	mockAccountRepo.AssertNotCalled(t, "GetCustomerAccountsInTransactionOrHasBalance")
}

// TestDeleteAccount_Execute_DatabaseError_CheckCustomer_AllScope tests if database throws error when scope=all and id is provided
func TestDeleteAccount_Execute_DatabaseError_CheckCustomer_AllScope(t *testing.T) {
	mockCustomerRepo := new(mock_repo.MockCustomerRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	deleteAccount := NewDeleteAccount(mockAccountRepo, mockCustomerRepo, mockEventRepo)

	scope := "all"
	id := "cust-123"
	requester := "user123"
	requestId := "req-456"

	mockCustomerRepo.On("Exists", id).Return(false, errors.New("database error"))

	message, err := deleteAccount.Execute(scope, id, requester, requestId)

	assert.Error(t, err)
	assert.ErrorIs(t, err, value.ErrDatabase)
	assert.Equal(t, "Failed to verify customer", message)

	mockCustomerRepo.AssertExpectations(t)
	mockAccountRepo.AssertNotCalled(t, "GetCustomerAccountsInTransactionOrHasBalance")
}

// TestDeleteAccount_Execute_AccountsLockedOrHasBalance_AllScope tests if any account under company in transaction
func TestDeleteAccount_Execute_AccountsLockedOrHasBalance_AllScope(t *testing.T) {
	mockCustomerRepo := new(mock_repo.MockCustomerRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	deleteAccount := NewDeleteAccount(mockAccountRepo, mockCustomerRepo, mockEventRepo)

	scope := "all"
	id := "cust-123"
	requester := "user123"
	requestId := "req-456"

	accounts := []*entity.Account{
		{ID: "acc-1", CustomerID: id, Balance: 100.0},
		{ID: "acc-2", CustomerID: id, Balance: 0.0},
	}

	mockCustomerRepo.On("Exists", id).Return(true, nil)
	mockAccountRepo.On("GetCustomerAccountsInTransactionOrHasBalance", id).Return(accounts, nil)

	message, err := deleteAccount.Execute(scope, id, requester, requestId)

	assert.Error(t, err)
	assert.ErrorIs(t, err, value.ErrAccountLocked)
	assert.Equal(t, "Account deletion blocked. Some accounts are in transaction or has balance", message)

	mockCustomerRepo.AssertExpectations(t)
	mockAccountRepo.AssertExpectations(t)
	mockAccountRepo.AssertNotCalled(t, "DeleteAllAccountsByCustomerID")
}

// TestDeleteAccount_Execute_DatabaseError_CheckAccounts_AllScope tests if no account found under company
func TestDeleteAccount_Execute_DatabaseError_CheckAccounts_AllScope(t *testing.T) {
	mockCustomerRepo := new(mock_repo.MockCustomerRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	deleteAccount := NewDeleteAccount(mockAccountRepo, mockCustomerRepo, mockEventRepo)

	scope := "all"
	id := "cust-123"
	requester := "user123"
	requestId := "req-456"

	mockCustomerRepo.On("Exists", id).Return(true, nil)
	mockAccountRepo.On("GetCustomerAccountsInTransactionOrHasBalance", id).Return(nil, errors.New("database error"))

	message, err := deleteAccount.Execute(scope, id, requester, requestId)

	assert.Error(t, err)
	assert.ErrorIs(t, err, value.ErrDatabase)
	assert.Equal(t, "Failed to verify accounts", message)

	mockCustomerRepo.AssertExpectations(t)
	mockAccountRepo.AssertExpectations(t)
	mockAccountRepo.AssertNotCalled(t, "DeleteAllAccountsByCustomerID")
}

// TestDeleteAccount_Execute_DatabaseError_DeleteAllAccounts tests if database throws error while deleting all accounts
func TestDeleteAccount_Execute_DatabaseError_DeleteAllAccounts(t *testing.T) {
	mockCustomerRepo := new(mock_repo.MockCustomerRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	deleteAccount := NewDeleteAccount(mockAccountRepo, mockCustomerRepo, mockEventRepo)

	scope := "all"
	id := "cust-123"
	requester := "user123"
	requestId := "req-456"

	mockCustomerRepo.On("Exists", id).Return(true, nil)
	mockAccountRepo.On("GetCustomerAccountsInTransactionOrHasBalance", id).Return([]*entity.Account{}, nil)
	mockAccountRepo.On("DeleteAllAccountsByCustomerID", id, requester).Return(errors.New("database error"))

	message, err := deleteAccount.Execute(scope, id, requester, requestId)

	assert.Error(t, err)
	assert.ErrorIs(t, err, value.ErrDatabase)
	assert.Equal(t, "Failed to delete accounts", message)

	mockCustomerRepo.AssertExpectations(t)
	mockAccountRepo.AssertExpectations(t)
	mockEventRepo.AssertNotCalled(t, "CreateEvent")
}

// TestDeleteAccount_Execute_AccountNotFound_SingleScope tests if scope=single and account not found with account id
func TestDeleteAccount_Execute_AccountNotFound_SingleScope(t *testing.T) {
	mockCustomerRepo := new(mock_repo.MockCustomerRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	deleteAccount := NewDeleteAccount(mockAccountRepo, mockCustomerRepo, mockEventRepo)

	scope := "single"
	id := "acc-123"
	requester := "user123"
	requestId := "req-456"

	mockAccountRepo.On("GetAccountByID", id).Return(nil, nil)

	message, err := deleteAccount.Execute(scope, id, requester, requestId)

	assert.Error(t, err)
	assert.Equal(t, "Account not found", message)

	mockAccountRepo.AssertExpectations(t)
	mockAccountRepo.AssertNotCalled(t, "CheckTransactionLock")
}

// TestDeleteAccount_Execute_DatabaseError_GetAccount_SingleScope if database throws error while getting account for scope=single
func TestDeleteAccount_Execute_DatabaseError_GetAccount_SingleScope(t *testing.T) {
	mockCustomerRepo := new(mock_repo.MockCustomerRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	deleteAccount := NewDeleteAccount(mockAccountRepo, mockCustomerRepo, mockEventRepo)

	scope := "single"
	id := "acc-123"
	requester := "user123"
	requestId := "req-456"

	mockAccountRepo.On("GetAccountByID", id).Return(nil, errors.New("database error"))

	message, err := deleteAccount.Execute(scope, id, requester, requestId)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to verify account")
	assert.Equal(t, "Failed to verify account", message)

	mockAccountRepo.AssertExpectations(t)
	mockAccountRepo.AssertNotCalled(t, "CheckTransactionLock")
}

// TestDeleteAccount_Execute_AccountHasBalance_SingleScope tests if account has balance in scope=single
func TestDeleteAccount_Execute_AccountHasBalance_SingleScope(t *testing.T) {
	mockCustomerRepo := new(mock_repo.MockCustomerRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	deleteAccount := NewDeleteAccount(mockAccountRepo, mockCustomerRepo, mockEventRepo)

	scope := "single"
	id := "acc-123"
	requester := "user123"
	requestId := "req-456"

	account := &entity.Account{
		ID:      "acc-123",
		Balance: 100.50,
	}

	mockAccountRepo.On("GetAccountByID", id).Return(account, nil)

	message, err := deleteAccount.Execute(scope, id, requester, requestId)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot delete account with positive balance")
	assert.Equal(t, "Account deletion blocked. Account has balance", message)

	mockAccountRepo.AssertExpectations(t)
	mockAccountRepo.AssertNotCalled(t, "CheckTransactionLock")
	mockAccountRepo.AssertNotCalled(t, "DeleteAccount")
}

// TestDeleteAccount_Execute_AccountInTransaction_SingleScope tests if account is in transaction while delete
func TestDeleteAccount_Execute_AccountInTransaction_SingleScope(t *testing.T) {
	mockCustomerRepo := new(mock_repo.MockCustomerRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	deleteAccount := NewDeleteAccount(mockAccountRepo, mockCustomerRepo, mockEventRepo)

	scope := "single"
	id := "acc-123"
	requester := "user123"
	requestId := "req-456"

	account := &entity.Account{
		ID:      "acc-123",
		Balance: 0.0,
	}

	mockAccountRepo.On("GetAccountByID", id).Return(account, nil)
	mockAccountRepo.On("CheckTransactionLock", id).Return(errors.New("account locked in transaction"))

	message, err := deleteAccount.Execute(scope, id, requester, requestId)

	assert.Error(t, err)
	assert.Equal(t, "Failed to verify accounts", message)

	mockAccountRepo.AssertExpectations(t)
	mockAccountRepo.AssertNotCalled(t, "DeleteAccount")
}

// TestDeleteAccount_Execute_DatabaseError_DeleteAccount_SingleScope tests if database throws error
func TestDeleteAccount_Execute_DatabaseError_DeleteAccount_SingleScope(t *testing.T) {
	mockCustomerRepo := new(mock_repo.MockCustomerRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	deleteAccount := NewDeleteAccount(mockAccountRepo, mockCustomerRepo, mockEventRepo)

	scope := "single"
	id := "acc-123"
	requester := "user123"
	requestId := "req-456"

	account := &entity.Account{
		ID:         "acc-123",
		CustomerID: "cust-123",
		Balance:    0.0,
	}

	mockAccountRepo.On("GetAccountByID", id).Return(account, nil)
	mockAccountRepo.On("CheckTransactionLock", id).Return(nil)
	mockAccountRepo.On("DeleteAccount", id, requester).Return(errors.New("database error"))

	message, err := deleteAccount.Execute(scope, id, requester, requestId)

	assert.Error(t, err)
	assert.ErrorIs(t, err, value.ErrDatabase)
	assert.Equal(t, "Failed to delete account", message)

	mockAccountRepo.AssertExpectations(t)
	mockEventRepo.AssertNotCalled(t, "CreateEvent")
}
