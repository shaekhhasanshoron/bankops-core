package account

import (
	"account-service/internal/domain/entity"
	custom_err "account-service/internal/domain/error"
	mock_repo "account-service/internal/ports/mocks/repo"
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestGetAccountBalance_Execute_Success tests success response if all inputs are provided correctly
func TestGetAccountBalance_Execute_Success(t *testing.T) {
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	getAccountBalance := NewGetAccountBalance(mockAccountRepo)

	id := "acc-123"
	requester := "user123"
	requestId := "req-456"
	expectedBalance := 1500.75

	account := &entity.Account{
		ID:      id,
		Balance: expectedBalance,
	}

	mockAccountRepo.On("GetAccountByID", id).Return(account, nil)

	balance, message, err := getAccountBalance.Execute(id, requester, requestId)

	assert.NoError(t, err)
	assert.Equal(t, "Account balance", message)
	assert.Equal(t, expectedBalance, balance)

	mockAccountRepo.AssertExpectations(t)
}

// TestGetAccountBalance_Execute_Success_ZeroBalance tests if account balance is zero
func TestGetAccountBalance_Execute_Success_ZeroBalance(t *testing.T) {
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	getAccountBalance := NewGetAccountBalance(mockAccountRepo)

	id := "acc-123"
	requester := "user123"
	requestId := "req-456"
	expectedBalance := 0.0

	account := &entity.Account{
		ID:      id,
		Balance: expectedBalance,
	}

	mockAccountRepo.On("GetAccountByID", id).Return(account, nil)

	balance, message, err := getAccountBalance.Execute(id, requester, requestId)

	assert.NoError(t, err)
	assert.Equal(t, "Account balance", message)
	assert.Equal(t, expectedBalance, balance)

	mockAccountRepo.AssertExpectations(t)
}

// TestGetAccountBalance_Execute_Success_NegativeBalance tests for negative balance
func TestGetAccountBalance_Execute_Success_NegativeBalance(t *testing.T) {
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	getAccountBalance := NewGetAccountBalance(mockAccountRepo)

	id := "acc-123"
	requester := "user123"
	requestId := "req-456"
	expectedBalance := -500.25

	account := &entity.Account{
		ID:      id,
		Balance: expectedBalance,
	}

	mockAccountRepo.On("GetAccountByID", id).Return(account, nil)

	balance, message, err := getAccountBalance.Execute(id, requester, requestId)

	assert.NoError(t, err)
	assert.Equal(t, "Account balance", message)
	assert.Equal(t, expectedBalance, balance)

	mockAccountRepo.AssertExpectations(t)
}

// TestGetAccountBalance_Execute_EmptyAccountID tests if account id not provided
func TestGetAccountBalance_Execute_EmptyAccountID(t *testing.T) {
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	getAccountBalance := NewGetAccountBalance(mockAccountRepo)

	balance, message, err := getAccountBalance.Execute("", "user123", "req-456")

	assert.Error(t, err)
	assert.ErrorIs(t, err, custom_err.ErrValidationFailed)
	assert.Equal(t, "Invalid request - 'id' account id missing", message)
	assert.Equal(t, 0.0, balance)

	mockAccountRepo.AssertNotCalled(t, "GetAccountByID")
}

// TestGetAccountBalance_Execute_AccountNotFound tests if account not found
func TestGetAccountBalance_Execute_AccountNotFound(t *testing.T) {
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	getAccountBalance := NewGetAccountBalance(mockAccountRepo)

	id := "acc-123"
	requester := "user123"
	requestId := "req-456"

	mockAccountRepo.On("GetAccountByID", id).Return(nil, nil)

	balance, message, err := getAccountBalance.Execute(id, requester, requestId)

	assert.Error(t, err)
	assert.Equal(t, "Account not found", message)
	assert.Equal(t, 0.0, balance)

	mockAccountRepo.AssertExpectations(t)
}

// TestGetAccountBalance_Execute_DatabaseError if database throws an error
func TestGetAccountBalance_Execute_DatabaseError(t *testing.T) {
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	getAccountBalance := NewGetAccountBalance(mockAccountRepo)

	id := "acc-123"
	requester := "user123"
	requestId := "req-456"

	mockAccountRepo.On("GetAccountByID", id).Return(nil, errors.New("database connection failed"))

	balance, message, err := getAccountBalance.Execute(id, requester, requestId)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to verify account")
	assert.Equal(t, "Failed to verify account", message)
	assert.Equal(t, 0.0, balance)

	mockAccountRepo.AssertExpectations(t)
}

// TestGetAccountBalance_Execute_EmptyRequester tests if requester not provided
func TestGetAccountBalance_Execute_EmptyRequester(t *testing.T) {
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	getAccountBalance := NewGetAccountBalance(mockAccountRepo)

	id := "acc-123"
	requester := ""
	requestId := "req-456"
	expectedBalance := 1000.0

	account := &entity.Account{
		ID:      id,
		Balance: expectedBalance,
	}

	mockAccountRepo.On("GetAccountByID", id).Return(account, nil)

	balance, message, err := getAccountBalance.Execute(id, requester, requestId)

	assert.NoError(t, err)
	assert.Equal(t, "Account balance", message)
	assert.Equal(t, expectedBalance, balance)

	mockAccountRepo.AssertExpectations(t)
}

// TestGetAccountBalance_Execute_WhitespaceAccountID tests if white space in input account id
func TestGetAccountBalance_Execute_WhitespaceAccountID(t *testing.T) {
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	getAccountBalance := NewGetAccountBalance(mockAccountRepo)

	balance, message, err := getAccountBalance.Execute("   ", "user123", "req-456")

	assert.Error(t, err)
	assert.ErrorIs(t, err, custom_err.ErrValidationFailed)
	assert.Equal(t, "Invalid request - 'id' account id missing", message)
	assert.Equal(t, 0.0, balance)

	mockAccountRepo.AssertNotCalled(t, "GetAccountByID")
}

// TestGetAccountBalance_Execute_AccountWithFullDetails tests account details
func TestGetAccountBalance_Execute_AccountWithFullDetails(t *testing.T) {
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	getAccountBalance := NewGetAccountBalance(mockAccountRepo)

	id := "acc-123"
	requester := "user123"
	requestId := "req-456"
	expectedBalance := 2500.50

	account := &entity.Account{
		ID:           id,
		CustomerID:   "cust-123",
		AccountType:  entity.AccountTypeSavings,
		Balance:      expectedBalance,
		ActiveStatus: entity.AccountActiveStatusActive,
		Version:      5,
		CreatedBy:    "admin",
	}

	mockAccountRepo.On("GetAccountByID", id).Return(account, nil)

	balance, message, err := getAccountBalance.Execute(id, requester, requestId)

	assert.NoError(t, err)
	assert.Equal(t, "Account balance", message)
	assert.Equal(t, expectedBalance, balance)

	mockAccountRepo.AssertExpectations(t)
}

// TestGetAccountBalance_Execute_ConcurrentAccess tests concurrent access to account balance
func TestGetAccountBalance_Execute_ConcurrentAccess(t *testing.T) {
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	getAccountBalance := NewGetAccountBalance(mockAccountRepo)

	id := "acc-123"
	requester := "user123"
	requestId := "req-456"
	expectedBalance := 5000.0

	account := &entity.Account{
		ID:      id,
		Balance: expectedBalance,
	}

	mockAccountRepo.On("GetAccountByID", id).Return(account, nil).Times(10)

	results := make(chan struct {
		balance float64
		message string
		err     error
	}, 3)

	// Concurrent calls executed
	for i := 0; i < 10; i++ {
		go func() {
			balance, message, err := getAccountBalance.Execute(id, requester, requestId)
			results <- struct {
				balance float64
				message string
				err     error
			}{balance, message, err}
		}()
	}

	for i := 0; i < 10; i++ {
		result := <-results
		assert.NoError(t, result.err)
		assert.Equal(t, "Account balance", result.message)
		assert.Equal(t, expectedBalance, result.balance)
	}

	mockAccountRepo.AssertExpectations(t)
}

// TestGetAccountBalance_Execute_AccountWithDifferentIDs tests using different  mock account ids
func TestGetAccountBalance_Execute_AccountWithDifferentIDs(t *testing.T) {
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	getAccountBalance := NewGetAccountBalance(mockAccountRepo)

	testCases := []struct {
		accountID       string
		expectedBalance float64
	}{
		{"acc-001", 100.0},
		{"acc-002", 200.0},
		{"acc-003", 300.0},
		{"acc-004", 300.0},
		{"acc-005", 300.0},
	}

	for _, tc := range testCases {
		t.Run(tc.accountID, func(t *testing.T) {
			account := &entity.Account{
				ID:      tc.accountID,
				Balance: tc.expectedBalance,
			}

			mockAccountRepo.On("GetAccountByID", tc.accountID).Return(account, nil).Once()

			balance, message, err := getAccountBalance.Execute(tc.accountID, "user123", "req-456")

			assert.NoError(t, err)
			assert.Equal(t, "Account balance", message)
			assert.Equal(t, tc.expectedBalance, balance)
		})
	}

	mockAccountRepo.AssertExpectations(t)
}
