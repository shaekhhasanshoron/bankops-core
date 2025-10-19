package customer

import (
	"account-service/internal/domain/entity"
	custom_err "account-service/internal/domain/error"
	mock_repo "account-service/internal/ports/mocks/repo"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

// TestDeleteCustomer_Execute_Success tests successful delete of customer
func TestDeleteCustomer_Execute_Success(t *testing.T) {
	mockCustomerRepo := new(mock_repo.MockCustomerRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	deleteCustomer := NewDeleteCustomer(mockCustomerRepo, mockEventRepo)

	id := "customer123"
	requester := "requester123"
	requestId := "req-456"

	customer := &entity.Customer{
		ID:   id,
		Name: "Test Customer",
		Accounts: []entity.Account{
			{
				ID:           "acc-1",
				ActiveStatus: entity.AccountActiveStatusActive,
				Balance:      0,
			},
		},
	}

	mockCustomerRepo.On("GetCustomerByID", id).Return(customer, nil)
	mockCustomerRepo.On("CheckModificationAllowed", id).Return(nil)
	mockCustomerRepo.On("DeleteCustomerByID", id, requester).Return(nil)
	mockEventRepo.On("CreateEvent", mock.AnythingOfType("*entity.Event")).Return(nil)

	message, err := deleteCustomer.Execute(id, requester, requestId)

	assert.NoError(t, err)
	assert.Equal(t, "Customer deleted successfully", message)

	mockCustomerRepo.AssertExpectations(t)
	mockEventRepo.AssertExpectations(t)
}

// TestDeleteCustomer_Execute_Success_NoAccounts tests customer with no accounts
func TestDeleteCustomer_Execute_Success_NoAccounts(t *testing.T) {
	mockCustomerRepo := new(mock_repo.MockCustomerRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	deleteCustomer := NewDeleteCustomer(mockCustomerRepo, mockEventRepo)

	id := "cust-123"
	requester := "user123"
	requestId := "req-456"

	customer := &entity.Customer{
		ID:       id,
		Name:     "Test Customer",
		Accounts: []entity.Account{},
	}

	mockCustomerRepo.On("GetCustomerByID", id).Return(customer, nil)
	mockCustomerRepo.On("CheckModificationAllowed", id).Return(nil)
	mockCustomerRepo.On("DeleteCustomerByID", id, requester).Return(nil)
	mockEventRepo.On("CreateEvent", mock.AnythingOfType("*entity.Event")).Return(nil)

	message, err := deleteCustomer.Execute(id, requester, requestId)

	assert.NoError(t, err)
	assert.Equal(t, "Customer deleted successfully", message)

	mockCustomerRepo.AssertExpectations(t)
	mockEventRepo.AssertExpectations(t)
}

// TestDeleteCustomer_Execute_EmptyCustomerID tests if customer id not provided
func TestDeleteCustomer_Execute_EmptyCustomerID(t *testing.T) {
	mockCustomerRepo := new(mock_repo.MockCustomerRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	deleteCustomer := NewDeleteCustomer(mockCustomerRepo, mockEventRepo)

	message, err := deleteCustomer.Execute("", "user123", "req-456")

	assert.Error(t, err)
	assert.ErrorIs(t, err, custom_err.ErrValidationFailed)
	assert.Equal(t, "Missing required data", message)

	mockCustomerRepo.AssertNotCalled(t, "GetCustomerByID")
	mockCustomerRepo.AssertNotCalled(t, "CheckModificationAllowed")
	mockCustomerRepo.AssertNotCalled(t, "DeleteCustomerByID")
}

// TestDeleteCustomer_Execute_EmptyRequester tests if requester is missing
func TestDeleteCustomer_Execute_EmptyRequester(t *testing.T) {
	mockCustomerRepo := new(mock_repo.MockCustomerRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	deleteCustomer := NewDeleteCustomer(mockCustomerRepo, mockEventRepo)

	message, err := deleteCustomer.Execute("cust-123", "", "req-456")

	assert.Error(t, err)
	assert.ErrorIs(t, err, custom_err.ErrValidationFailed)
	assert.Equal(t, "Unknown requester", message)

	mockCustomerRepo.AssertNotCalled(t, "GetCustomerByID")
	mockCustomerRepo.AssertNotCalled(t, "CheckModificationAllowed")
	mockCustomerRepo.AssertNotCalled(t, "DeleteCustomerByID")
}

// TestDeleteCustomer_Execute_EmptyRequester tests if requester id is missing
func TestDeleteCustomer_Execute_EmptyRequestId(t *testing.T) {
	mockCustomerRepo := new(mock_repo.MockCustomerRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	deleteCustomer := NewDeleteCustomer(mockCustomerRepo, mockEventRepo)

	id := "cust-123"
	requester := "user123"
	requestId := ""

	customer := &entity.Customer{
		ID:       id,
		Name:     "Test Customer",
		Accounts: []entity.Account{},
	}
	mockCustomerRepo.On("GetCustomerByID", id).Return(customer, nil)
	mockCustomerRepo.On("CheckModificationAllowed", id).Return(nil)
	mockCustomerRepo.On("DeleteCustomerByID", id, requester).Return(nil)
	mockEventRepo.On("CreateEvent", mock.AnythingOfType("*entity.Event")).Return(nil)

	message, err := deleteCustomer.Execute(id, requester, requestId)

	assert.NoError(t, err)
	assert.Equal(t, "Customer deleted successfully", message)
}

// TestDeleteCustomer_Execute_CustomerNotFound_NilCustomer tests if customer not found
func TestDeleteCustomer_Execute_CustomerNotFound_NilCustomer(t *testing.T) {
	mockCustomerRepo := new(mock_repo.MockCustomerRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	deleteCustomer := NewDeleteCustomer(mockCustomerRepo, mockEventRepo)

	id := "cust-123"
	requester := "user123"
	requestId := "req-456"

	mockCustomerRepo.On("GetCustomerByID", id).Return(nil, nil)

	message, err := deleteCustomer.Execute(id, requester, requestId)

	assert.Error(t, err)
	assert.ErrorIs(t, err, custom_err.ErrCustomerNotFound)
	assert.Equal(t, "Customer not found", message)

	mockCustomerRepo.AssertNotCalled(t, "CheckModificationAllowed")
	mockCustomerRepo.AssertNotCalled(t, "DeleteCustomerByID")
	mockEventRepo.AssertNotCalled(t, "CreateEvent")
}

// TestDeleteCustomer_Execute_ModificationNotAllowed tests if modification not allowed
func TestDeleteCustomer_Execute_ModificationNotAllowed(t *testing.T) {
	mockCustomerRepo := new(mock_repo.MockCustomerRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	deleteCustomer := NewDeleteCustomer(mockCustomerRepo, mockEventRepo)

	id := "cust-123"
	requester := "user123"
	requestId := "req-456"

	customer := &entity.Customer{
		ID:   id,
		Name: "Test Customer",
		Accounts: []entity.Account{
			{
				ID:           "acc-1",
				ActiveStatus: entity.AccountActiveStatusInactive,
				Balance:      0,
			},
		},
	}

	// Mock expectations
	mockCustomerRepo.On("GetCustomerByID", id).Return(customer, nil)
	mockCustomerRepo.On("CheckModificationAllowed", id).Return(errors.New("modification blocked"))

	// Execute
	message, err := deleteCustomer.Execute(id, requester, requestId)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "Customer deletion not allowed right now", message)

	mockCustomerRepo.AssertNotCalled(t, "DeleteCustomerByID")
	mockEventRepo.AssertNotCalled(t, "CreateEvent")
}

// TestDeleteCustomer_Execute_ModificationNotAllowed tests if modification not allowed
func TestDeleteCustomer_Execute_MultipleAccounts_ActiveWithBalance(t *testing.T) {
	mockCustomerRepo := new(mock_repo.MockCustomerRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	deleteCustomer := NewDeleteCustomer(mockCustomerRepo, mockEventRepo)

	id := "cust-123"
	requester := "user123"
	requestId := "req-456"

	customer := &entity.Customer{
		ID:   id,
		Name: "Test Customer",
		Accounts: []entity.Account{
			{
				ID:           "acc-1",
				ActiveStatus: entity.AccountActiveStatusActive,
				Balance:      0,
			},
			{
				ID:           "acc-2",
				ActiveStatus: entity.AccountActiveStatusActive,
				Balance:      50.25,
			},
			{
				ID:           "acc-3",
				ActiveStatus: entity.AccountActiveStatusActive,
				Balance:      0,
			},
		},
	}

	mockCustomerRepo.On("GetCustomerByID", id).Return(customer, nil)
	mockCustomerRepo.On("CheckModificationAllowed", id).Return(nil)

	_, err := deleteCustomer.Execute(id, requester, requestId)

	assert.Error(t, err)

	mockCustomerRepo.AssertNotCalled(t, "DeleteCustomerByID")
	mockEventRepo.AssertNotCalled(t, "CreateEvent")
}

// TestDeleteCustomer_Execute_DatabaseError_DeleteCustomer tests if database error occurs while deleting customer
func TestDeleteCustomer_Execute_DatabaseError_DeleteCustomer(t *testing.T) {
	mockCustomerRepo := new(mock_repo.MockCustomerRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	deleteCustomer := NewDeleteCustomer(mockCustomerRepo, mockEventRepo)

	id := "cust-123"
	requester := "user123"
	requestId := "req-456"

	customer := &entity.Customer{
		ID:   id,
		Name: "Test Customer",
		Accounts: []entity.Account{
			{
				ID:           "acc-1",
				ActiveStatus: entity.CustomerActiveStatusActive,
				Balance:      0,
			},
		},
	}

	mockCustomerRepo.On("GetCustomerByID", id).Return(customer, nil)
	mockCustomerRepo.On("CheckModificationAllowed", id).Return(nil)
	mockCustomerRepo.On("DeleteCustomerByID", id, requester).Return(errors.New("database error"))

	message, err := deleteCustomer.Execute(id, requester, requestId)

	assert.Error(t, err)
	assert.ErrorIs(t, err, custom_err.ErrDatabase)
	assert.Equal(t, "Customer deletion failed", message)

	mockCustomerRepo.AssertExpectations(t)
	mockEventRepo.AssertNotCalled(t, "CreateEvent")
}

func TestDeleteCustomer_Execute_EventCreationFails(t *testing.T) {
	// Setup
	mockCustomerRepo := new(mock_repo.MockCustomerRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	deleteCustomer := NewDeleteCustomer(mockCustomerRepo, mockEventRepo)

	id := "cust-123"
	requester := "user123"
	requestId := "req-456"

	customer := &entity.Customer{
		ID:   id,
		Name: "Test Customer",
		Accounts: []entity.Account{
			{
				ID:           "acc-1",
				ActiveStatus: entity.AccountActiveStatusActive,
				Balance:      0,
			},
		},
	}

	mockCustomerRepo.On("GetCustomerByID", id).Return(customer, nil)
	mockCustomerRepo.On("CheckModificationAllowed", id).Return(nil)
	mockCustomerRepo.On("DeleteCustomerByID", id, requester).Return(nil)
	mockEventRepo.On("CreateEvent", mock.AnythingOfType("*entity.Event")).Return(errors.New("event storage failed"))

	message, err := deleteCustomer.Execute(id, requester, requestId)

	assert.NoError(t, err)
	assert.Equal(t, "Customer deleted successfully", message)

	mockCustomerRepo.AssertExpectations(t)
	mockEventRepo.AssertExpectations(t)
}
