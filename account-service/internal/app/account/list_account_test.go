package account

import (
	"account-service/internal/domain/entity"
	custom_err "account-service/internal/domain/error"
	mock_repo "account-service/internal/ports/mocks/repo"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestListAccount_Execute_SuccessWithDefaultPagination test success if input are feault
func TestListAccount_Execute_SuccessWithDefaultPagination(t *testing.T) {
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	listAccount := NewListAccount(mockAccountRepo)

	customerID := "cust-123"
	minBalance := ""
	inTransaction := ""
	page := 0
	pageSize := 0
	requester := "user123"
	requestId := "req-456"
	setOrder := "desc"

	expectedAccounts := []*entity.Account{
		{ID: "acc-1", CustomerID: customerID, Balance: 1000.0},
		{ID: "acc-2", CustomerID: customerID, Balance: 2000.0},
	}
	expectedTotalCount := int64(2)

	mockAccountRepo.On("GetAccountsByFiltersWithPagination",
		map[string]interface{}{
			"status":      "valid",
			"customer_id": customerID,
		},
		1, 100,
	).Return(expectedAccounts, expectedTotalCount, nil)

	accounts, totalCount, totalPages, message, err := listAccount.Execute(
		customerID, minBalance, inTransaction, page, pageSize, setOrder, requester, requestId,
	)

	assert.NoError(t, err)
	assert.Equal(t, expectedAccounts, accounts)
	assert.Equal(t, expectedTotalCount, totalCount)
	assert.Equal(t, int64(1), totalPages)
	assert.Equal(t, "Account List", message)
	mockAccountRepo.AssertExpectations(t)
}

// TestListAccount_Execute_SuccessWithCustomPagination tests with custom pagination
func TestListAccount_Execute_SuccessWithCustomPagination(t *testing.T) {
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	listAccount := NewListAccount(mockAccountRepo)

	customerID := "cust-123"
	minBalance := ""
	inTransaction := ""
	page := 2
	pageSize := 10
	requester := "user123"
	requestId := "req-456"
	setOrder := "desc"

	expectedAccounts := []*entity.Account{
		{ID: "acc-1", CustomerID: customerID, Balance: 1000.0},
	}
	expectedTotalCount := int64(15)

	mockAccountRepo.On("GetAccountsByFiltersWithPagination",
		map[string]interface{}{
			"status":      "valid",
			"customer_id": customerID,
		},
		page, pageSize,
	).Return(expectedAccounts, expectedTotalCount, nil)

	accounts, totalCount, totalPages, message, err := listAccount.Execute(
		customerID, minBalance, inTransaction, page, pageSize, setOrder, requester, requestId,
	)

	assert.NoError(t, err)
	assert.Equal(t, expectedAccounts, accounts)
	assert.Equal(t, expectedTotalCount, totalCount)
	assert.Equal(t, int64(2), totalPages)
	assert.Equal(t, "Account List", message)
	mockAccountRepo.AssertExpectations(t)
}

// TestListAccount_Execute_SuccessWithMinBalanceFilter tests with minimum filter balance
func TestListAccount_Execute_SuccessWithMinBalanceFilter(t *testing.T) {
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	listAccount := NewListAccount(mockAccountRepo)

	customerID := "cust-123"
	minBalance := "500.50"
	inTransaction := ""
	page := 1
	pageSize := 20
	requester := "user123"
	requestId := "req-456"
	setOrder := "desc"

	expectedAccounts := []*entity.Account{
		{ID: "acc-1", CustomerID: customerID, Balance: 1000.0},
	}
	expectedTotalCount := int64(1)

	mockAccountRepo.On("GetAccountsByFiltersWithPagination",
		map[string]interface{}{
			"status":      "valid",
			"customer_id": customerID,
			"min_balance": 500.5,
		},
		page, pageSize,
	).Return(expectedAccounts, expectedTotalCount, nil)

	accounts, totalCount, totalPages, message, err := listAccount.Execute(
		customerID, minBalance, inTransaction, page, pageSize, setOrder, requester, requestId,
	)

	assert.NoError(t, err)
	assert.Equal(t, expectedAccounts, accounts)
	assert.Equal(t, expectedTotalCount, totalCount)
	assert.Equal(t, expectedTotalCount, totalCount)
	assert.Equal(t, "Account List", message)
	assert.Equal(t, int64(1), totalPages)
	mockAccountRepo.AssertExpectations(t)
}

// TestListAccount_Execute_SuccessWithInTransactionTrue tests when in_transaction=true is provided in the query
func TestListAccount_Execute_SuccessWithInTransactionTrue(t *testing.T) {
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	listAccount := NewListAccount(mockAccountRepo)

	customerID := "cust-123"
	minBalance := ""
	inTransaction := "true"
	page := 1
	pageSize := 100
	requester := "user123"
	requestId := "req-456"
	setOrder := "desc"

	expectedAccounts := []*entity.Account{
		{ID: "acc-1", CustomerID: customerID, Balance: 1000.0, LockedForTx: true},
	}
	expectedTotalCount := int64(1)

	mockAccountRepo.On("GetAccountsByFiltersWithPagination",
		map[string]interface{}{
			"status":        "valid",
			"customer_id":   customerID,
			"locked_for_tx": true,
		},
		page, pageSize,
	).Return(expectedAccounts, expectedTotalCount, nil)

	accounts, totalCount, totalPages, message, err := listAccount.Execute(
		customerID, minBalance, inTransaction, page, pageSize, setOrder, requester, requestId,
	)

	assert.NoError(t, err)
	assert.Equal(t, expectedAccounts, accounts)
	assert.Equal(t, expectedTotalCount, totalCount)
	assert.Equal(t, int64(1), totalPages)
	assert.Equal(t, "Account List", message)
	mockAccountRepo.AssertExpectations(t)
}

// TestListAccount_Execute_SuccessWithInTransactionFalse tests when in_transaction=false is provided in the query
func TestListAccount_Execute_SuccessWithInTransactionFalse(t *testing.T) {
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	listAccount := NewListAccount(mockAccountRepo)

	customerID := "cust-123"
	minBalance := ""
	inTransaction := "false"
	page := 1
	pageSize := 100
	requester := "user123"
	requestId := "req-456"
	setOrder := "desc"

	expectedAccounts := []*entity.Account{
		{ID: "acc-1", CustomerID: customerID, Balance: 1000.0, LockedForTx: false},
	}
	expectedTotalCount := int64(1)

	mockAccountRepo.On("GetAccountsByFiltersWithPagination",
		map[string]interface{}{
			"status":        "valid",
			"customer_id":   customerID,
			"locked_for_tx": false,
		},
		page, pageSize,
	).Return(expectedAccounts, expectedTotalCount, nil)

	accounts, totalCount, totalPages, message, err := listAccount.Execute(
		customerID, minBalance, inTransaction, page, pageSize, setOrder, requester, requestId,
	)

	assert.NoError(t, err)
	assert.Equal(t, expectedAccounts, accounts)
	assert.Equal(t, expectedTotalCount, totalCount)
	assert.Equal(t, "Account List", message)
	assert.Equal(t, int64(1), totalPages)
	mockAccountRepo.AssertExpectations(t)
}

// TestListAccount_Execute_SuccessWithAllFilters test if all filters are provided
func TestListAccount_Execute_SuccessWithAllFilters(t *testing.T) {
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	listAccount := NewListAccount(mockAccountRepo)

	customerID := "cust-123"
	minBalance := "1000.0"
	inTransaction := "true"
	page := 1
	pageSize := 50
	requester := "user123"
	requestId := "req-456"
	setOrder := "desc"

	expectedAccounts := []*entity.Account{
		{ID: "acc-1", CustomerID: customerID, Balance: 1500.0, LockedForTx: true},
	}
	expectedTotalCount := int64(1)

	mockAccountRepo.On("GetAccountsByFiltersWithPagination",
		map[string]interface{}{
			"status":        "valid",
			"customer_id":   customerID,
			"min_balance":   1000.0,
			"locked_for_tx": true,
		},
		page, pageSize,
	).Return(expectedAccounts, expectedTotalCount, nil)

	accounts, totalCount, totalPages, message, err := listAccount.Execute(
		customerID, minBalance, inTransaction, page, pageSize, setOrder, requester, requestId,
	)

	assert.NoError(t, err)
	assert.Equal(t, expectedAccounts, accounts)
	assert.Equal(t, expectedTotalCount, totalCount)
	assert.Equal(t, int64(1), totalPages)
	assert.Equal(t, "Account List", message)
	mockAccountRepo.AssertExpectations(t)
}

// TestListAccount_Execute_SuccessWithDefaultFilter test no filter provided
func TestListAccount_Execute_SuccessWithDefaultFilter(t *testing.T) {
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	listAccount := NewListAccount(mockAccountRepo)

	customerID := ""
	minBalance := ""
	inTransaction := ""
	page := 1
	pageSize := 100
	requester := "user123"
	requestId := "req-456"
	setOrder := "desc"

	expectedAccounts := []*entity.Account{
		{ID: "acc-1", CustomerID: "cust-1", Balance: 1000.0},
		{ID: "acc-2", CustomerID: "cust-2", Balance: 2000.0},
	}
	expectedTotalCount := int64(2)

	mockAccountRepo.On("GetAccountsByFiltersWithPagination",
		map[string]interface{}{
			"status": "valid",
		},
		page, pageSize,
	).Return(expectedAccounts, expectedTotalCount, nil)

	accounts, totalCount, totalPages, message, err := listAccount.Execute(
		customerID, minBalance, inTransaction, page, pageSize, setOrder, requester, requestId,
	)

	assert.NoError(t, err)
	assert.Equal(t, expectedAccounts, accounts)
	assert.Equal(t, expectedTotalCount, totalCount)
	assert.Equal(t, int64(1), totalPages)
	assert.Equal(t, "Account List", message)
	mockAccountRepo.AssertExpectations(t)
}

// TestListAccount_Execute_ErrorInvalidMinBalanceFormat tests for invalid balance
func TestListAccount_Execute_ErrorInvalidMinBalanceFormat(t *testing.T) {
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	listAccount := NewListAccount(mockAccountRepo)

	customerID := "cust-123"
	minBalance := "invalid-balance"
	inTransaction := ""
	page := 1
	pageSize := 100
	requester := "user123"
	requestId := "req-456"
	setOrder := "desc"

	accounts, totalCount, totalPages, message, err := listAccount.Execute(
		customerID, minBalance, inTransaction, page, pageSize, setOrder, requester, requestId,
	)

	assert.ErrorIs(t, err, custom_err.ErrValidationFailed)
	assert.Contains(t, message, "Invalid request - customer id missing")
	assert.Nil(t, accounts)
	assert.Equal(t, int64(0), totalCount)
	assert.Equal(t, int64(0), totalPages)
	mockAccountRepo.AssertNotCalled(t, "GetAccountsByFiltersWithPagination")
}

// TestListAccount_Execute_ErrorNegativeMinBalance tests for negative balance
func TestListAccount_Execute_ErrorNegativeMinBalance(t *testing.T) {
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	listAccount := NewListAccount(mockAccountRepo)

	customerID := "cust-123"
	minBalance := "-100.0"
	inTransaction := ""
	page := 1
	pageSize := 100
	requester := "user123"
	requestId := "req-456"
	setOrder := "desc"

	accounts, totalCount, totalPages, message, err := listAccount.Execute(
		customerID, minBalance, inTransaction, page, pageSize, setOrder, requester, requestId,
	)

	assert.Error(t, err)
	assert.Contains(t, message, "Invalid request - invalid balance")
	assert.Nil(t, accounts)
	assert.Equal(t, int64(0), totalCount)
	assert.Equal(t, int64(0), totalPages)
	mockAccountRepo.AssertNotCalled(t, "GetAccountsByFiltersWithPagination")
}
