package customer

import (
	"account-service/internal/domain/entity"
	custom_err "account-service/internal/domain/error"
	mock_repo "account-service/internal/ports/mocks/repo"
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestListCustomer_Execute_Success check for success response
func TestListCustomer_Execute_Success(t *testing.T) {
	mockCustomerRepo := new(mock_repo.MockCustomerRepo)
	listCustomer := NewListCustomer(mockCustomerRepo)

	page := 1
	pageSize := 10
	requestId := "req-123"
	sortOrder := "desc"

	customers := []*entity.Customer{
		{ID: "cus-1", Name: "Customer One"},
		{ID: "cus-2", Name: "Customer Two"},
	}
	totalCount := int64(50)

	mockCustomerRepo.On("ListCustomer", page, pageSize).Return(customers, totalCount, nil)

	result, resultTotalCount, totalPages, message, err := listCustomer.Execute(page, pageSize, requestId, sortOrder)

	assert.NoError(t, err)
	assert.Equal(t, "Customer List", message)
	assert.Equal(t, customers, result)
	assert.Equal(t, totalCount, resultTotalCount)
	assert.Equal(t, int64(5), totalPages)

	mockCustomerRepo.AssertExpectations(t)
}

// TestListCustomer_Execute_Success_EmptyResults checks if empty results can be handled
func TestListCustomer_Execute_Success_EmptyResults(t *testing.T) {
	mockCustomerRepo := new(mock_repo.MockCustomerRepo)
	listCustomer := NewListCustomer(mockCustomerRepo)

	page := 1
	pageSize := 10
	requestId := "req-123"
	sortOrder := "desc"

	var customers []*entity.Customer
	totalCount := int64(0)

	mockCustomerRepo.On("ListCustomer", page, pageSize).Return(customers, totalCount, nil)

	result, resultTotalCount, totalPages, message, err := listCustomer.Execute(page, pageSize, requestId, sortOrder)

	assert.NoError(t, err)
	assert.Equal(t, "Customer List", message)
	assert.Equal(t, customers, result)
	assert.Equal(t, totalCount, resultTotalCount)
	assert.Equal(t, int64(0), totalPages)

	mockCustomerRepo.AssertExpectations(t)
}

// TestListCustomer_Execute_PageBoundaryAdjustment checks if page counts are not boundary it will be adjusted
func TestListCustomer_Execute_PageBoundaryAdjustment(t *testing.T) {
	mockCustomerRepo := new(mock_repo.MockCustomerRepo)
	listCustomer := NewListCustomer(mockCustomerRepo)

	page := 0
	pageSize := 150
	requestId := "req-123"
	sortOrder := "desc"

	customers := []*entity.Customer{
		{ID: "cust-1", Name: "Customer One"},
	}
	totalCount := int64(1)

	mockCustomerRepo.On("ListCustomer", 1, 100).Return(customers, totalCount, nil)

	result, resultTotalCount, totalPages, message, err := listCustomer.Execute(page, pageSize, requestId, sortOrder)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "Customer List", message)
	assert.Equal(t, customers, result)
	assert.Equal(t, totalCount, resultTotalCount)
	assert.Equal(t, int64(1), totalPages)

	mockCustomerRepo.AssertExpectations(t)
}

// TestListCustomer_Execute_DatabaseError checks if database throws an error
func TestListCustomer_Execute_DatabaseError(t *testing.T) {
	mockCustomerRepo := new(mock_repo.MockCustomerRepo)
	listCustomer := NewListCustomer(mockCustomerRepo)

	page := 1
	pageSize := 10
	requestId := "req-123"
	sortOrder := "desc"

	var nilCustomers []*entity.Customer
	mockCustomerRepo.On("ListCustomer", page, pageSize).Return(nilCustomers, int64(0), errors.New("database error"))

	result, resultTotalCount, totalPages, message, err := listCustomer.Execute(page, pageSize, requestId, sortOrder)

	assert.Error(t, err)
	assert.ErrorIs(t, err, custom_err.ErrDatabase)
	assert.Equal(t, "Failed to list customer", message)
	assert.Nil(t, result)
	assert.Equal(t, int64(0), resultTotalCount)
	assert.Equal(t, int64(0), totalPages)

	mockCustomerRepo.AssertExpectations(t)
}

// TestListCustomer_Execute_InvalidPageSizeTooSmall checks if page size is invalid or not
func TestListCustomer_Execute_InvalidPageSizeTooSmall(t *testing.T) {
	mockCustomerRepo := new(mock_repo.MockCustomerRepo)
	listCustomer := NewListCustomer(mockCustomerRepo)

	page := 1
	pageSize := 0
	requestId := "req-123"
	sortOrder := "desc"

	customers := []*entity.Customer{
		{ID: "cust-1", Name: "Customer One"},
	}
	totalCount := int64(1)

	mockCustomerRepo.On("ListCustomer", page, 100).Return(customers, totalCount, nil)

	result, resultTotalCount, totalPages, message, err := listCustomer.Execute(page, pageSize, requestId, sortOrder)

	assert.NoError(t, err)
	assert.Equal(t, "Customer List", message)
	assert.Equal(t, customers, result)
	assert.Equal(t, totalCount, resultTotalCount)
	assert.Equal(t, totalPages, int64(1))

	mockCustomerRepo.AssertExpectations(t)
}

// TestNewListCustomer checks customer list
func TestNewListCustomer(t *testing.T) {
	mockCustomerRepo := new(mock_repo.MockCustomerRepo)

	useCase := NewListCustomer(mockCustomerRepo)

	assert.NotNil(t, useCase)
	assert.Equal(t, mockCustomerRepo, useCase.CustomerRepo)
}
