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

// TestCreateAccount_Execute_Success tests success response if all inputs are properly provided
func TestCreateAccount_Execute_Success(t *testing.T) {
	mockCustomerRepo := new(mock_repo.MockCustomerRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	createAccount := NewCreateAccount(mockAccountRepo, mockCustomerRepo, mockEventRepo)

	customerID := "cust-123"
	initialDeposit := 100.0
	requester := "user123"
	requestId := "req-456"

	mockCustomerRepo.On("Exists", customerID).Return(true, nil)
	mockAccountRepo.On("CreateAccount", mock.AnythingOfType("*entity.Account")).Return(nil)
	mockEventRepo.On("CreateEvent", mock.AnythingOfType("*entity.Event")).Return(nil)

	account, message, err := createAccount.Execute(customerID, initialDeposit, requester, requestId)

	assert.NoError(t, err)
	assert.Equal(t, "Account successfully created", message)
	assert.NotNil(t, account)
	assert.Equal(t, customerID, account.CustomerID)
	assert.Equal(t, initialDeposit, account.Balance)
	assert.Equal(t, requester, account.CreatedBy)
	assert.Equal(t, entity.AccountTypeSavings, account.AccountType)

	mockCustomerRepo.AssertExpectations(t)
	mockAccountRepo.AssertExpectations(t)
	mockEventRepo.AssertExpectations(t)
}

// TestCreateAccount_Execute_EmptyCustomerID tests if customer id not provided in input
func TestCreateAccount_Execute_EmptyCustomerID(t *testing.T) {
	mockCustomerRepo := new(mock_repo.MockCustomerRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	createAccount := NewCreateAccount(mockAccountRepo, mockCustomerRepo, mockEventRepo)

	account, message, err := createAccount.Execute("", 100.0, "user123", "req-456")

	assert.Error(t, err)
	assert.ErrorIs(t, err, value.ErrValidationFailed)
	assert.Equal(t, "Required missing fields", message)
	assert.Nil(t, account)

	mockCustomerRepo.AssertNotCalled(t, "Exists")
	mockAccountRepo.AssertNotCalled(t, "CreateAccount")
}

// TestCreateAccount_Execute_NegativeInitialDeposit tests if initial deposit amount is invalid
func TestCreateAccount_Execute_NegativeInitialDeposit(t *testing.T) {
	mockCustomerRepo := new(mock_repo.MockCustomerRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	createAccount := NewCreateAccount(mockAccountRepo, mockCustomerRepo, mockEventRepo)

	account, message, err := createAccount.Execute("cust-123", -50.0, "user123", "req-456")

	assert.Error(t, err)
	assert.ErrorIs(t, err, value.ErrInvalidAmount)
	assert.Equal(t, "Invalid request", message)
	assert.Nil(t, account)

	mockCustomerRepo.AssertNotCalled(t, "Exists")
	mockAccountRepo.AssertNotCalled(t, "CreateAccount")
}

// TestCreateAccount_Execute_EmptyRequester check if request is empty
func TestCreateAccount_Execute_EmptyRequester(t *testing.T) {
	mockCustomerRepo := new(mock_repo.MockCustomerRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	createAccount := NewCreateAccount(mockAccountRepo, mockCustomerRepo, mockEventRepo)

	account, message, err := createAccount.Execute("cust-123", 100.0, "", "req-456")

	assert.Error(t, err)
	assert.ErrorIs(t, err, value.ErrValidationFailed)
	assert.Equal(t, "Unknown requester", message)
	assert.Nil(t, account)

	mockCustomerRepo.AssertNotCalled(t, "Exists")
	mockAccountRepo.AssertNotCalled(t, "CreateAccount")
}

// TestCreateAccount_Execute_DatabaseError_CheckCustomer tests if database trows error while checking customer
func TestCreateAccount_Execute_DatabaseError_CheckCustomer(t *testing.T) {
	mockCustomerRepo := new(mock_repo.MockCustomerRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	createAccount := NewCreateAccount(mockAccountRepo, mockCustomerRepo, mockEventRepo)

	customerID := "cust-123"
	initialDeposit := 100.0
	requester := "user123"
	requestId := "req-456"

	mockCustomerRepo.On("Exists", customerID).Return(false, errors.New("database error"))

	account, message, err := createAccount.Execute(customerID, initialDeposit, requester, requestId)

	assert.Error(t, err)
	assert.ErrorIs(t, err, value.ErrDatabase)
	assert.Equal(t, "Failed to verify customer", message)
	assert.Nil(t, account)

	mockCustomerRepo.AssertExpectations(t)
	mockAccountRepo.AssertNotCalled(t, "CreateAccount")
}

// TestCreateAccount_Execute_CustomerNotFound tests if customer not found
func TestCreateAccount_Execute_CustomerNotFound(t *testing.T) {
	mockCustomerRepo := new(mock_repo.MockCustomerRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	createAccount := NewCreateAccount(mockAccountRepo, mockCustomerRepo, mockEventRepo)

	customerID := "cust-123"
	initialDeposit := 100.0
	requester := "user123"
	requestId := "req-456"

	mockCustomerRepo.On("Exists", customerID).Return(false, nil)

	account, message, err := createAccount.Execute(customerID, initialDeposit, requester, requestId)

	assert.Error(t, err)
	assert.ErrorIs(t, err, value.ErrCustomerNotFound)
	assert.Equal(t, "Customer not found", message)
	assert.Nil(t, account)

	mockCustomerRepo.AssertExpectations(t)
	mockAccountRepo.AssertNotCalled(t, "CreateAccount")
}

// TestCreateAccount_Execute_DatabaseError_CreateAccount tests if database throws error while creating account
func TestCreateAccount_Execute_DatabaseError_CreateAccount(t *testing.T) {
	mockCustomerRepo := new(mock_repo.MockCustomerRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	createAccount := NewCreateAccount(mockAccountRepo, mockCustomerRepo, mockEventRepo)

	customerID := "cust-123"
	initialDeposit := 100.0
	requester := "user123"
	requestId := "req-456"

	// Mock expectations
	mockCustomerRepo.On("Exists", customerID).Return(true, nil)
	mockAccountRepo.On("CreateAccount", mock.AnythingOfType("*entity.Account")).Return(errors.New("database error"))

	// Execute
	account, message, err := createAccount.Execute(customerID, initialDeposit, requester, requestId)

	// Assert
	assert.Error(t, err)
	assert.ErrorIs(t, err, value.ErrDatabase)
	assert.Equal(t, "Failed to create account", message)
	assert.Nil(t, account)

	mockCustomerRepo.AssertExpectations(t)
	mockAccountRepo.AssertExpectations(t)
	mockEventRepo.AssertNotCalled(t, "CreateEvent")
}
