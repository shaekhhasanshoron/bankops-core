package customer

import (
	"account-service/internal/domain/entity"
	custom_err "account-service/internal/domain/error"
	mock_repo "account-service/internal/ports/mocks/repo"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestCreateCustomer_Execute_Success_NewCustomer tests a new customer will be created if valid inputs are provided
func TestCreateCustomer_Execute_Success_NewCustomer(t *testing.T) {
	mockCustomerRepo := new(mock_repo.MockCustomerRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	createCustomer := NewCreateCustomer(mockCustomerRepo, mockEventRepo)

	name := "Test Customer"
	requester := "user123"
	requestId := "req-123"

	mockCustomerRepo.On("GetCustomerByName", name).Return(nil, errors.New("not found"))

	expectedCustomer := &entity.Customer{
		Name:      name,
		CreatedBy: requester,
	}

	mockCustomerRepo.On("CreateCustomer", mock.AnythingOfType("*entity.Customer")).
		Return(expectedCustomer, nil)

	mockEventRepo.On("CreateEvent", mock.AnythingOfType("*entity.Event")).Return(nil)

	customer, message, err := createCustomer.Execute(name, requester, requestId)

	assert.NoError(t, err)
	assert.Equal(t, "Customer created successfully", message)
	assert.NotNil(t, customer)

	assert.Equal(t, name, customer.Name)
	assert.Equal(t, requester, customer.CreatedBy)
	assert.NotEmpty(t, customer.ID)
	assert.Equal(t, "active", customer.ActiveStatus)
	assert.Equal(t, "valid", customer.Status)
	assert.Equal(t, 1, customer.Version)

	mockCustomerRepo.AssertExpectations(t)
	mockEventRepo.AssertExpectations(t)
}

// TestCreateCustomer_Execute_CustomerAlreadyExists tests if a customer already exists
func TestCreateCustomer_Execute_CustomerAlreadyExists(t *testing.T) {
	mockCustomerRepo := new(mock_repo.MockCustomerRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	createCustomer := NewCreateCustomer(mockCustomerRepo, mockEventRepo)

	name := "Existing Customer"
	requester := "user123"
	requestId := "req-123"

	existingCustomer := &entity.Customer{
		ID:   "existing-cust-123",
		Name: name,
	}

	mockCustomerRepo.On("GetCustomerByName", name).Return(existingCustomer, nil)

	customer, message, err := createCustomer.Execute(name, requester, requestId)

	assert.Error(t, err)
	assert.ErrorIs(t, err, custom_err.ErrCustomerExists)
	assert.Equal(t, "Customer already exists", message)
	assert.Nil(t, customer)

	mockCustomerRepo.AssertNotCalled(t, "CreateCustomer")
	mockEventRepo.AssertNotCalled(t, "CreateEvent")
}

// TestCreateCustomer_Execute_Success_WithSpecialCharacters tests name that contains special character
func TestCreateCustomer_Execute_Success_WithSpecialCharacters(t *testing.T) {
	mockCustomerRepo := new(mock_repo.MockCustomerRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	createCustomer := NewCreateCustomer(mockCustomerRepo, mockEventRepo)

	name := "Test Customer & Co. (Neura)"
	requester := "user123"
	requestId := "req-123"

	mockCustomerRepo.On("GetCustomerByName", name).Return(nil, errors.New("not found"))
	mockCustomerRepo.On("CreateCustomer", mock.AnythingOfType("*entity.Customer")).Return(nil, nil)
	mockEventRepo.On("CreateEvent", mock.AnythingOfType("*entity.Event")).Return(nil)

	customer, message, err := createCustomer.Execute(name, requester, requestId)

	assert.NoError(t, err)
	assert.Equal(t, "Customer created successfully", message)
	assert.NotNil(t, customer)
	assert.Equal(t, name, customer.Name)
	assert.Equal(t, requester, customer.CreatedBy)

	mockCustomerRepo.AssertExpectations(t)
	mockEventRepo.AssertExpectations(t)
}

// TestCreateCustomer_Execute_EmptyName tests empty name input
func TestCreateCustomer_Execute_EmptyName(t *testing.T) {
	mockCustomerRepo := new(mock_repo.MockCustomerRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	createCustomer := NewCreateCustomer(mockCustomerRepo, mockEventRepo)

	customer, message, err := createCustomer.Execute("", "user123", "req-123")

	assert.Error(t, err)
	assert.ErrorIs(t, err, custom_err.ErrValidationFailed)
	assert.Equal(t, "Required missing fields", message)
	assert.Nil(t, customer)

	mockCustomerRepo.AssertNotCalled(t, "GetCustomerByName")
	mockCustomerRepo.AssertNotCalled(t, "CreateCustomer")
}

// TestCreateCustomer_Execute_DatabaseError_CreateCustomer tests if database throws error while creating customer
func TestCreateCustomer_Execute_DatabaseError_CreateCustomer(t *testing.T) {
	mockCustomerRepo := new(mock_repo.MockCustomerRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	createCustomer := NewCreateCustomer(mockCustomerRepo, mockEventRepo)

	name := "Test Customer"
	requester := "user123"
	requestId := "req-123"

	mockCustomerRepo.On("GetCustomerByName", name).Return(nil, errors.New("not found"))
	mockCustomerRepo.On("CreateCustomer", mock.AnythingOfType("*entity.Customer")).Return(nil, errors.New("database error"))

	customer, message, err := createCustomer.Execute(name, requester, requestId)

	assert.Error(t, err)
	assert.ErrorIs(t, err, custom_err.ErrDatabase)
	assert.Equal(t, "Failed to create customer", message)
	assert.Nil(t, customer)

	mockCustomerRepo.AssertExpectations(t)
	mockEventRepo.AssertNotCalled(t, "CreateEvent")
}

// TestCreateCustomer_Execute_DatabaseError_CheckExisting tests database error
func TestCreateCustomer_Execute_DatabaseError_CheckExisting(t *testing.T) {
	mockCustomerRepo := new(mock_repo.MockCustomerRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	createCustomer := NewCreateCustomer(mockCustomerRepo, mockEventRepo)

	name := "Test Customer"
	requester := "user123"
	requestId := "req-123"

	mockCustomerRepo.On("GetCustomerByName", name).Return(nil, errors.New("database connection failed"))
	mockCustomerRepo.On("CreateCustomer", mock.AnythingOfType("*entity.Customer")).Return(nil, errors.New("database connection error"))

	customer, message, err := createCustomer.Execute(name, requester, requestId)

	assert.Error(t, err)
	assert.Contains(t, message, "Failed to create customer")
	assert.Nil(t, customer)

	mockCustomerRepo.AssertNotCalled(t, "CreateCustomer")
}
