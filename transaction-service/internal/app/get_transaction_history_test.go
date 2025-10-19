package app

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
	"transaction-service/internal/domain/entity"
	mock_repo "transaction-service/internal/ports/mocks"
)

// TestGetTransactionHistory_Execute_SuccessWithAccountID transaction history successful with account id
func TestGetTransactionHistory_Execute_SuccessWithAccountID(t *testing.T) {
	mockTransactionRepo := new(mock_repo.MockTransactionRepo)
	getTransactionHistory := NewGetTransactionHistory(mockTransactionRepo)

	accountID := "acc-123"
	customerID := ""
	types := []string{}
	var startDate, endDate *time.Time = nil, nil
	sortOrder := "desc"
	page := 1
	pageSize := 50
	requester := "user123"
	requestId := "req-456"

	expectedTransactions := []*entity.Transaction{
		{ID: "txn-1", SourceAccountID: accountID, Amount: 100.0},
		{ID: "txn-2", SourceAccountID: accountID, Amount: 200.0},
	}
	expectedTotal := int64(2)

	mockTransactionRepo.On("GetTransactionHistory", accountID, customerID, startDate, endDate, sortOrder, page, pageSize, types).Return(expectedTransactions, expectedTotal, nil)

	transactions, total, err := getTransactionHistory.Execute(accountID, customerID, types, startDate, endDate, sortOrder, page, pageSize, requester, requestId)

	assert.NoError(t, err)
	assert.Equal(t, expectedTransactions, transactions)
	assert.Equal(t, expectedTotal, total)
	mockTransactionRepo.AssertExpectations(t)
}

// TestGetTransactionHistory_Execute_SuccessWithCustomerID tests getting all transaction history by customer id
func TestGetTransactionHistory_Execute_SuccessWithCustomerID(t *testing.T) {
	mockTransactionRepo := new(mock_repo.MockTransactionRepo)
	getTransactionHistory := NewGetTransactionHistory(mockTransactionRepo)

	accountID := ""
	customerID := "comp-123"
	types := []string{"transfer", "withdraw"}
	var startDate, endDate *time.Time = nil, nil
	sortOrder := "desc"
	page := 1
	pageSize := 50
	requester := "user123"
	requestId := "req-456"

	expectedTransactions := []*entity.Transaction{
		{ID: "txn-1", Amount: 100.0},
		{ID: "txn-2", Amount: 200.0},
	}
	expectedTotal := int64(2)

	mockTransactionRepo.On("GetTransactionHistory", accountID, customerID, startDate, endDate, sortOrder, page, pageSize, types).Return(expectedTransactions, expectedTotal, nil)

	transactions, total, err := getTransactionHistory.Execute(accountID, customerID, types, startDate, endDate, sortOrder, page, pageSize, requester, requestId)

	assert.NoError(t, err)
	assert.Equal(t, expectedTransactions, transactions)
	assert.Equal(t, expectedTotal, total)
	mockTransactionRepo.AssertExpectations(t)
}

// TestGetTransactionHistory_Execute_SuccessWithDateRange tests getting all transaction history by date range
func TestGetTransactionHistory_Execute_SuccessWithDateRange(t *testing.T) {
	mockTransactionRepo := new(mock_repo.MockTransactionRepo)
	getTransactionHistory := NewGetTransactionHistory(mockTransactionRepo)

	accountID := "acc-123"
	customerID := ""
	types := []string{}
	startDate := time.Now().AddDate(0, -1, 0) // 1 month ago
	endDate := time.Now()
	sortOrder := "asc"
	page := 1
	pageSize := 50
	requester := "user123"
	requestId := "req-456"

	expectedTransactions := []*entity.Transaction{
		{ID: "txn-1", SourceAccountID: accountID, Amount: 100.0},
	}
	expectedTotal := int64(1)

	mockTransactionRepo.On("GetTransactionHistory", accountID, customerID, &startDate, &endDate, sortOrder, page, pageSize, types).Return(expectedTransactions, expectedTotal, nil)

	transactions, total, err := getTransactionHistory.Execute(accountID, customerID, types, &startDate, &endDate, sortOrder, page, pageSize, requester, requestId)

	assert.NoError(t, err)
	assert.Equal(t, expectedTransactions, transactions)
	assert.Equal(t, expectedTotal, total)
	mockTransactionRepo.AssertExpectations(t)
}

// TestGetTransactionHistory_Execute_SuccessWithTransactionTypes tests getting all transaction history by transaction type
func TestGetTransactionHistory_Execute_SuccessWithTransactionTypes(t *testing.T) {
	mockTransactionRepo := new(mock_repo.MockTransactionRepo)
	getTransactionHistory := NewGetTransactionHistory(mockTransactionRepo)

	accountID := "acc-123"
	customerID := ""
	types := []string{"transfer", "withdraw_amount", "add_amount"}
	var startDate, endDate *time.Time = nil, nil
	sortOrder := "desc"
	page := 1
	pageSize := 50
	requester := "user123"
	requestId := "req-456"

	expectedTransactions := []*entity.Transaction{
		{ID: "txn-1", Type: "transfer", Amount: 100.0},
		{ID: "txn-2", Type: "withdraw_amount", Amount: 50.0},
	}
	expectedTotal := int64(2)

	mockTransactionRepo.On("GetTransactionHistory", accountID, customerID, startDate, endDate, sortOrder, page, pageSize, types).Return(expectedTransactions, expectedTotal, nil)

	transactions, total, err := getTransactionHistory.Execute(accountID, customerID, types, startDate, endDate, sortOrder, page, pageSize, requester, requestId)

	assert.NoError(t, err)
	assert.Equal(t, expectedTransactions, transactions)
	assert.Equal(t, expectedTotal, total)
	mockTransactionRepo.AssertExpectations(t)
}

// TestGetTransactionHistory_Execute_SuccessWithPagination tests getting all transaction history by pagination
func TestGetTransactionHistory_Execute_SuccessWithPagination(t *testing.T) {
	mockTransactionRepo := new(mock_repo.MockTransactionRepo)
	getTransactionHistory := NewGetTransactionHistory(mockTransactionRepo)

	accountID := "acc-123"
	customerID := ""
	types := []string{}
	var startDate, endDate *time.Time = nil, nil
	sortOrder := "desc"
	page := 2
	pageSize := 10
	requester := "user123"
	requestId := "req-456"

	expectedTransactions := []*entity.Transaction{
		{ID: "txn-11", Amount: 100.0},
		{ID: "txn-12", Amount: 200.0},
	}
	expectedTotal := int64(15)

	mockTransactionRepo.On("GetTransactionHistory", accountID, customerID, startDate, endDate, sortOrder, page, pageSize, types).Return(expectedTransactions, expectedTotal, nil)

	transactions, total, err := getTransactionHistory.Execute(accountID, customerID, types, startDate, endDate, sortOrder, page, pageSize, requester, requestId)

	assert.NoError(t, err)
	assert.Equal(t, expectedTransactions, transactions)
	assert.Equal(t, expectedTotal, total)
	mockTransactionRepo.AssertExpectations(t)
}

// TestGetTransactionHistory_Execute_SuccessEmptyResult tests for empty transaction history
func TestGetTransactionHistory_Execute_SuccessEmptyResult(t *testing.T) {
	mockTransactionRepo := new(mock_repo.MockTransactionRepo)
	getTransactionHistory := NewGetTransactionHistory(mockTransactionRepo)

	accountID := "acc-123"
	customerID := ""
	types := []string{}
	var startDate, endDate *time.Time = nil, nil
	sortOrder := "desc"
	page := 1
	pageSize := 50
	requester := "user123"
	requestId := "req-456"

	expectedTransactions := []*entity.Transaction{}
	expectedTotal := int64(0)

	mockTransactionRepo.On("GetTransactionHistory", accountID, customerID, startDate, endDate, sortOrder, page, pageSize, types).Return(expectedTransactions, expectedTotal, nil)

	transactions, total, err := getTransactionHistory.Execute(accountID, customerID, types, startDate, endDate, sortOrder, page, pageSize, requester, requestId)

	assert.NoError(t, err)
	assert.Equal(t, expectedTransactions, transactions)
	assert.Equal(t, expectedTotal, total)
	mockTransactionRepo.AssertExpectations(t)
}

// TestGetTransactionHistory_Execute_DefaultPaginationValues tests with default pagination rules
func TestGetTransactionHistory_Execute_DefaultPaginationValues(t *testing.T) {
	mockTransactionRepo := new(mock_repo.MockTransactionRepo)
	getTransactionHistory := NewGetTransactionHistory(mockTransactionRepo)

	accountID := "acc-123"
	customerID := ""
	types := []string{}
	var startDate, endDate *time.Time = nil, nil
	sortOrder := "desc"
	page := 0     // Should default to 1
	pageSize := 0 // Should default to 50
	requester := "user123"
	requestId := "req-456"

	expectedTransactions := []*entity.Transaction{}
	expectedTotal := int64(0)

	// Should be called with default values: page=1, pageSize=50
	mockTransactionRepo.On("GetTransactionHistory", accountID, customerID, startDate, endDate, sortOrder, 1, 50, types).Return(expectedTransactions, expectedTotal, nil)

	transactions, total, err := getTransactionHistory.Execute(accountID, customerID, types, startDate, endDate, sortOrder, page, pageSize, requester, requestId)

	assert.NoError(t, err)
	assert.Equal(t, expectedTransactions, transactions)
	assert.Equal(t, expectedTotal, total)
	mockTransactionRepo.AssertExpectations(t)
}

// TestGetTransactionHistory_Execute_PageSizeBoundaryValues tests with page size boundaries
func TestGetTransactionHistory_Execute_PageSizeBoundaryValues(t *testing.T) {
	mockTransactionRepo := new(mock_repo.MockTransactionRepo)
	getTransactionHistory := NewGetTransactionHistory(mockTransactionRepo)

	accountID := "acc-123"
	customerID := ""
	types := []string{}
	var startDate, endDate *time.Time = nil, nil
	sortOrder := "desc"
	requester := "user123"
	requestId := "req-456"

	testCases := []struct {
		name         string
		pageSize     int
		expectedSize int
	}{
		{"page_size_less_than_1", 0, 50},
		{"page_size_1", 1, 1},
		{"page_size_49", 49, 49},
		{"page_size_50", 50, 50},
		{"page_size_51", 51, 50},
		{"page_size_100", 100, 50},
		{"page_size_101", 101, 50},
		{"page_size_negative", -1, 50},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			expectedTransactions := []*entity.Transaction{}
			expectedTotal := int64(0)

			mockTransactionRepo.On("GetTransactionHistory", accountID, customerID, startDate, endDate, sortOrder, 1, tc.expectedSize, types).Return(expectedTransactions, expectedTotal, nil)

			transactions, total, err := getTransactionHistory.Execute(accountID, customerID, types, startDate, endDate, sortOrder, 1, tc.pageSize, requester, requestId)

			assert.NoError(t, err)
			assert.Equal(t, expectedTransactions, transactions)
			assert.Equal(t, expectedTotal, total)
			mockTransactionRepo.AssertExpectations(t)
		})
	}
}

// TestGetTransactionHistory_Execute_PageBoundaryValues test with page boundaries
func TestGetTransactionHistory_Execute_PageBoundaryValues(t *testing.T) {
	mockTransactionRepo := new(mock_repo.MockTransactionRepo)
	getTransactionHistory := NewGetTransactionHistory(mockTransactionRepo)

	accountID := "acc-123"
	customerID := ""
	types := []string{}
	var startDate, endDate *time.Time = nil, nil
	sortOrder := "desc"
	pageSize := 50
	requester := "user123"
	requestId := "req-456"

	testCases := []struct {
		name         string
		page         int
		expectedPage int
	}{
		{"page_0", 0, 1},
		{"page_1", 1, 1},
		{"page_2", 2, 2},
		{"page_negative", -1, 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			expectedTransactions := []*entity.Transaction{}
			expectedTotal := int64(0)

			mockTransactionRepo.On("GetTransactionHistory", accountID, customerID, startDate, endDate, sortOrder, tc.expectedPage, pageSize, types).Return(expectedTransactions, expectedTotal, nil)

			transactions, total, err := getTransactionHistory.Execute(accountID, customerID, types, startDate, endDate, sortOrder, tc.page, pageSize, requester, requestId)

			assert.NoError(t, err)
			assert.Equal(t, expectedTransactions, transactions)
			assert.Equal(t, expectedTotal, total)
			mockTransactionRepo.AssertExpectations(t)
		})
	}
}

// TestGetTransactionHistory_Execute_SortOrderValidation test with sort validation
func TestGetTransactionHistory_Execute_SortOrderValidation(t *testing.T) {
	mockTransactionRepo := new(mock_repo.MockTransactionRepo)
	getTransactionHistory := NewGetTransactionHistory(mockTransactionRepo)

	accountID := "acc-123"
	customerID := ""
	types := []string{}
	var startDate, endDate *time.Time = nil, nil
	page := 1
	pageSize := 50
	requester := "user123"
	requestId := "req-456"

	testCases := []struct {
		name          string
		sortOrder     string
		expectedOrder string
	}{
		{"sort_order_asc", "asc", "asc"},
		{"sort_order_desc", "desc", "desc"},
		{"sort_order_empty", "", "desc"},
		{"sort_order_invalid", "invalid", "desc"},
		{"sort_order_ASC", "ASC", "desc"},
		{"sort_order_DESC", "DESC", "desc"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			expectedTransactions := []*entity.Transaction{}
			expectedTotal := int64(0)

			mockTransactionRepo.On("GetTransactionHistory", accountID, customerID, startDate, endDate, tc.expectedOrder, page, pageSize, types).Return(expectedTransactions, expectedTotal, nil)

			transactions, total, err := getTransactionHistory.Execute(accountID, customerID, types, startDate, endDate, tc.sortOrder, page, pageSize, requester, requestId)

			assert.NoError(t, err)
			assert.Equal(t, expectedTransactions, transactions)
			assert.Equal(t, expectedTotal, total)
			mockTransactionRepo.AssertExpectations(t)
		})
	}
}

// TestGetTransactionHistory_Execute_ErrorDateRangeValidation test with date range
func TestGetTransactionHistory_Execute_ErrorDateRangeValidation(t *testing.T) {
	mockTransactionRepo := new(mock_repo.MockTransactionRepo)
	getTransactionHistory := NewGetTransactionHistory(mockTransactionRepo)

	accountID := "acc-123"
	customerID := ""
	types := []string{}
	startDate := time.Now().AddDate(0, 0, 1)
	endDate := time.Now().AddDate(0, 0, -1)
	sortOrder := "desc"
	page := 1
	pageSize := 50
	requester := "user123"
	requestId := "req-456"

	transactions, total, err := getTransactionHistory.Execute(accountID, customerID, types, &startDate, &endDate, sortOrder, page, pageSize, requester, requestId)

	assert.Error(t, err)
	assert.Nil(t, transactions)
	assert.Equal(t, int64(0), total)
	mockTransactionRepo.AssertNotCalled(t, "GetTransactionHistory")
}

// TestGetTransactionHistory_Execute_ErrorDatabaseFailure tests if database throws error
func TestGetTransactionHistory_Execute_ErrorDatabaseFailure(t *testing.T) {
	mockTransactionRepo := new(mock_repo.MockTransactionRepo)
	getTransactionHistory := NewGetTransactionHistory(mockTransactionRepo)

	accountID := "acc-123"
	customerID := ""
	types := []string{}
	var startDate, endDate *time.Time = nil, nil
	sortOrder := "desc"
	page := 1
	pageSize := 50
	requester := "user123"
	requestId := "req-456"

	mockTransactionRepo.On("GetTransactionHistory", accountID, customerID, startDate, endDate, sortOrder, page, pageSize, types).Return([]*entity.Transaction{}, int64(0), fmt.Errorf("database error"))

	transactions, total, err := getTransactionHistory.Execute(accountID, customerID, types, startDate, endDate, sortOrder, page, pageSize, requester, requestId)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get transaction history")
	assert.Empty(t, transactions)
	assert.Equal(t, int64(0), total)
	mockTransactionRepo.AssertExpectations(t)
}

// TestGetTransactionHistory_Execute_SuccessWithAllFilters test for all filter
func TestGetTransactionHistory_Execute_SuccessWithAllFilters(t *testing.T) {
	mockTransactionRepo := new(mock_repo.MockTransactionRepo)
	getTransactionHistory := NewGetTransactionHistory(mockTransactionRepo)

	accountID := "acc-123"
	customerID := "comp-456"
	types := []string{"transfer", "withdraw_amount"}
	startDate := time.Now().AddDate(0, -1, 0)
	endDate := time.Now()
	sortOrder := "asc"
	page := 2
	pageSize := 25
	requester := "user123"
	requestId := "req-456"

	expectedTransactions := []*entity.Transaction{
		{ID: "txn-1", Type: "transfer", Amount: 100.0},
		{ID: "txn-2", Type: "withdraw_amount", Amount: 50.0},
	}
	expectedTotal := int64(2)

	mockTransactionRepo.On("GetTransactionHistory", accountID, customerID, &startDate, &endDate, sortOrder, page, pageSize, types).Return(expectedTransactions, expectedTotal, nil)

	transactions, total, err := getTransactionHistory.Execute(accountID, customerID, types, &startDate, &endDate, sortOrder, page, pageSize, requester, requestId)

	assert.NoError(t, err)
	assert.Equal(t, expectedTransactions, transactions)
	assert.Equal(t, expectedTotal, total)
	mockTransactionRepo.AssertExpectations(t)
}
