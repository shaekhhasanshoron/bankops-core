package app

import (
	"context"
	"errors"
	"testing"
	"transaction-service/internal/app/saga"
	"transaction-service/internal/domain/entity"
	"transaction-service/internal/ports"
	"transaction-service/internal/ports/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestTransactionSagaOrchestrator_ExecuteTransactionSync_SuccessfulTransfer tests successful transfer execution
func TestTransactionSagaOrchestrator_ExecuteTransactionSync_SuccessfulTransfer(t *testing.T) {
	mockSagaRepo := new(mocks.MockSagaRepo)
	mockAccountClient := new(mocks.MockAccountClient)
	mockTransactionRepo := new(mocks.MockTransactionRepo)
	mockEventRepo := new(mocks.MockEventRepo)

	orchestrator := saga.NewTransactionSagaOrchestrator(
		mockSagaRepo,
		mockAccountClient,
		mockTransactionRepo,
		mockEventRepo,
	)

	ctx := context.Background()
	transaction := &entity.Transaction{
		ID:                   "txn-123",
		SourceAccountID:      "acc-123",
		DestinationAccountID: stringPtr("acc-456"),
		Amount:               100.0,
		Type:                 entity.TransactionTypeTransfer,
		ReferenceID:          "ref-123",
	}
	requester := "user-1"
	requestID := "req-123"

	accountsInfo := []ports.AccountInfo{
		{AccountID: "acc-123", CustomerID: "cust-123", Balance: 500.0, Version: 1},
		{AccountID: "acc-456", CustomerID: "cust-456", Balance: 200.0, Version: 1},
	}

	mockSagaRepo.On("CreateSaga", mock.AnythingOfType("*entity.TransactionSaga")).Return(nil)

	// Validation step
	mockAccountClient.On("ValidateAndGetAccounts",
		ctx,
		[]string{"acc-123", "acc-456"},
		requester,
		requestID,
	).Return(accountsInfo, "", nil)

	mockSagaRepo.On("UpdateSaga", mock.AnythingOfType("*entity.TransactionSaga")).Return(nil)

	// Locking step
	mockAccountClient.On("LockAccounts",
		ctx,
		[]string{"acc-123", "acc-456"},
		"txn-123",
		requester,
		requestID,
	).Return("", nil)

	// Process step
	mockAccountClient.On("UpdateAccountsBalance",
		ctx,
		mock.MatchedBy(func(updates []ports.AccountBalanceUpdate) bool {
			return len(updates) == 2 &&
				updates[0].AccountID == "acc-123" && updates[0].NewBalance == 400.0 &&
				updates[1].AccountID == "acc-456" && updates[1].NewBalance == 300.0
		}),
		requester,
		requestID,
	).Return([]ports.AccountBalanceUpdateResponse{}, "", nil)

	// Complete step
	mockTransactionRepo.On("UpdateTransactionStatus",
		"txn-123",
		entity.TransactionStatusSuccessful,
		"",
	).Return(nil)

	mockAccountClient.On("UnlockAccounts",
		ctx,
		"txn-123",
		requester,
		requestID,
	).Return("", nil)

	mockTransactionRepo.On("UpdateTransactionStatus",
		"txn-123",
		entity.TransactionStatusCompleted,
		"",
	).Return(nil)

	// Execute
	err := orchestrator.ExecuteTransactionSync(
		ctx,
		transaction,
		requester,
		requestID,
	)

	// Assert
	assert.NoError(t, err)
	mockSagaRepo.AssertExpectations(t)
	mockAccountClient.AssertExpectations(t)
	mockTransactionRepo.AssertExpectations(t)
}

// TestInitTransaction_Execute_SuccessfulWithdrawFull tests successful withdraw full transaction
func TestInitTransaction_Execute_SuccessfulWithdrawFull(t *testing.T) {
	mockTransactionRepo := new(mocks.MockTransactionRepo)
	mockAccountClient := new(mocks.MockAccountClient)
	mockSagaRepo := new(mocks.MockSagaRepo)
	mockEventRepo := new(mocks.MockEventRepo)

	initTransaction := NewInitTransaction(
		mockTransactionRepo,
		mockAccountClient,
		mockSagaRepo,
		mockEventRepo,
	)

	ctx := context.Background()
	sourceAccountID := "acc-123"
	var destAccountID *string = nil
	amount := 100.0 // This should be set to 0 for withdraw full
	transactionType := entity.TransactionTypeWithdrawFull
	referenceID := "ref-123"
	requester := "user-1"
	requestID := "req-123"

	accountsInfo := []ports.AccountInfo{
		{AccountID: sourceAccountID, CustomerID: "cust-123", Balance: 500.0},
	}

	expectedTransaction := &entity.Transaction{
		ID:                   "txn-123",
		SourceAccountID:      sourceAccountID,
		DestinationAccountID: nil,
		Amount:               0,
		Type:                 transactionType,
		TransactionStatus:    entity.TransactionStatusCompleted,
	}

	// Mock expectations
	mockAccountClient.On("ValidateAndGetAccounts",
		ctx,
		[]string{sourceAccountID},
		requester,
		requestID,
	).Return(accountsInfo, "", nil)

	mockTransactionRepo.On("CreateTransaction", mock.AnythingOfType("*entity.Transaction")).
		Run(func(args mock.Arguments) {
			tx := args.Get(0).(*entity.Transaction)
			assert.Equal(t, sourceAccountID, tx.SourceAccountID)
			assert.Nil(t, tx.DestinationAccountID)
			assert.Equal(t, 0.0, tx.Amount) // Should be 0 for withdraw full
			assert.Equal(t, transactionType, tx.Type)
		}).
		Return(nil)

	// Mock saga orchestrator execution
	mockSagaRepo.On("CreateSaga", mock.AnythingOfType("*entity.TransactionSaga")).Return(nil)
	mockAccountClient.On("ValidateAndGetAccounts", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(accountsInfo, "", nil)
	mockAccountClient.On("LockAccounts", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return("", nil)
	mockAccountClient.On("UpdateAccountsBalance", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return([]ports.AccountBalanceUpdateResponse{}, "", nil)
	mockAccountClient.On("UnlockAccounts", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return("", nil)
	mockSagaRepo.On("UpdateSaga", mock.AnythingOfType("*entity.TransactionSaga")).Return(nil)
	mockTransactionRepo.On("UpdateTransactionStatus", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	mockTransactionRepo.On("GetTransactionByID", mock.Anything).
		Return(expectedTransaction, nil)

	mockEventRepo.On("CreateEvent", mock.AnythingOfType("*entity.Event")).Return(nil)

	// Execute
	transaction, msg, err := initTransaction.Execute(
		ctx,
		sourceAccountID,
		destAccountID,
		amount,
		transactionType,
		referenceID,
		requester,
		requestID,
	)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, transaction)
	assert.Equal(t, "Transaction completed successfully", msg)
	mockTransactionRepo.AssertExpectations(t)
	mockAccountClient.AssertExpectations(t)
	mockSagaRepo.AssertExpectations(t)
}

// TestInitTransaction_Execute_ValidationFailure_MissingSourceAccount tests validation failure for missing source account
func TestInitTransaction_Execute_ValidationFailure_MissingSourceAccount(t *testing.T) {
	mockTransactionRepo := new(mocks.MockTransactionRepo)
	mockAccountClient := new(mocks.MockAccountClient)
	mockSagaRepo := new(mocks.MockSagaRepo)
	mockEventRepo := new(mocks.MockEventRepo)

	initTransaction := NewInitTransaction(
		mockTransactionRepo,
		mockAccountClient,
		mockSagaRepo,
		mockEventRepo,
	)

	ctx := context.Background()
	sourceAccountID := ""
	destAccountID := "acc-456"
	amount := 100.0
	transactionType := entity.TransactionTypeTransfer
	referenceID := "ref-123"
	requester := "user-1"
	requestID := "req-123"

	// Execute
	transaction, msg, err := initTransaction.Execute(
		ctx,
		sourceAccountID,
		&destAccountID,
		amount,
		transactionType,
		referenceID,
		requester,
		requestID,
	)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, transaction)
	assert.Equal(t, "Source Account ID missing", msg)
}

// TestInitTransaction_Execute_ValidationFailure_MissingReferenceID tests validation failure for missing reference ID
func TestInitTransaction_Execute_ValidationFailure_MissingReferenceID(t *testing.T) {
	mockTransactionRepo := new(mocks.MockTransactionRepo)
	mockAccountClient := new(mocks.MockAccountClient)
	mockSagaRepo := new(mocks.MockSagaRepo)
	mockEventRepo := new(mocks.MockEventRepo)

	initTransaction := NewInitTransaction(
		mockTransactionRepo,
		mockAccountClient,
		mockSagaRepo,
		mockEventRepo,
	)

	ctx := context.Background()
	sourceAccountID := "acc-123"
	destAccountID := "acc-456"
	amount := 100.0
	transactionType := entity.TransactionTypeTransfer
	referenceID := ""
	requester := "user-1"
	requestID := "req-123"

	// Execute
	transaction, msg, err := initTransaction.Execute(
		ctx,
		sourceAccountID,
		&destAccountID,
		amount,
		transactionType,
		referenceID,
		requester,
		requestID,
	)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, transaction)
	assert.Equal(t, "Reference ID missing", msg)
}

// TestInitTransaction_Execute_ValidationFailure_MissingRequester tests validation failure for missing requester
func TestInitTransaction_Execute_ValidationFailure_MissingRequester(t *testing.T) {
	mockTransactionRepo := new(mocks.MockTransactionRepo)
	mockAccountClient := new(mocks.MockAccountClient)
	mockSagaRepo := new(mocks.MockSagaRepo)
	mockEventRepo := new(mocks.MockEventRepo)

	initTransaction := NewInitTransaction(
		mockTransactionRepo,
		mockAccountClient,
		mockSagaRepo,
		mockEventRepo,
	)

	ctx := context.Background()
	sourceAccountID := "acc-123"
	destAccountID := "acc-456"
	amount := 100.0
	transactionType := entity.TransactionTypeTransfer
	referenceID := "ref-123"
	requester := ""
	requestID := "req-123"

	// Execute
	transaction, msg, err := initTransaction.Execute(
		ctx,
		sourceAccountID,
		&destAccountID,
		amount,
		transactionType,
		referenceID,
		requester,
		requestID,
	)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, transaction)
	assert.Equal(t, "Unknown requester", msg)
}

// TestInitTransaction_Execute_ValidationFailure_InvalidTransactionType tests validation failure for invalid transaction type
func TestInitTransaction_Execute_ValidationFailure_InvalidTransactionType(t *testing.T) {
	mockTransactionRepo := new(mocks.MockTransactionRepo)
	mockAccountClient := new(mocks.MockAccountClient)
	mockSagaRepo := new(mocks.MockSagaRepo)
	mockEventRepo := new(mocks.MockEventRepo)

	initTransaction := NewInitTransaction(
		mockTransactionRepo,
		mockAccountClient,
		mockSagaRepo,
		mockEventRepo,
	)

	ctx := context.Background()
	sourceAccountID := "acc-123"
	destAccountID := "acc-456"
	amount := 100.0
	transactionType := "INVALID_TYPE"
	referenceID := "ref-123"
	requester := "user-1"
	requestID := "req-123"

	// Execute
	transaction, msg, err := initTransaction.Execute(
		ctx,
		sourceAccountID,
		&destAccountID,
		amount,
		transactionType,
		referenceID,
		requester,
		requestID,
	)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, transaction)
	assert.Equal(t, "Invalid transaction type", msg)
}

// TestInitTransaction_Execute_ValidationFailure_SameSourceDestination tests validation failure for same source and destination
func TestInitTransaction_Execute_ValidationFailure_SameSourceDestination(t *testing.T) {
	mockTransactionRepo := new(mocks.MockTransactionRepo)
	mockAccountClient := new(mocks.MockAccountClient)
	mockSagaRepo := new(mocks.MockSagaRepo)
	mockEventRepo := new(mocks.MockEventRepo)

	initTransaction := NewInitTransaction(
		mockTransactionRepo,
		mockAccountClient,
		mockSagaRepo,
		mockEventRepo,
	)

	ctx := context.Background()
	sourceAccountID := "acc-123"
	destAccountID := "acc-123" // Same as source
	amount := 100.0
	transactionType := entity.TransactionTypeTransfer
	referenceID := "ref-123"
	requester := "user-1"
	requestID := "req-123"

	// Execute
	transaction, msg, err := initTransaction.Execute(
		ctx,
		sourceAccountID,
		&destAccountID,
		amount,
		transactionType,
		referenceID,
		requester,
		requestID,
	)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, transaction)
	assert.Equal(t, "Cannot transfer amount to the same account", msg)
}

// TestInitTransaction_Execute_AccountValidationFailure tests account validation failure
func TestInitTransaction_Execute_AccountValidationFailure(t *testing.T) {
	mockTransactionRepo := new(mocks.MockTransactionRepo)
	mockAccountClient := new(mocks.MockAccountClient)
	mockSagaRepo := new(mocks.MockSagaRepo)
	mockEventRepo := new(mocks.MockEventRepo)

	initTransaction := NewInitTransaction(
		mockTransactionRepo,
		mockAccountClient,
		mockSagaRepo,
		mockEventRepo,
	)

	ctx := context.Background()
	sourceAccountID := "acc-123"
	destAccountID := "acc-456"
	amount := 100.0
	transactionType := entity.TransactionTypeTransfer
	referenceID := "ref-123"
	requester := "user-1"
	requestID := "req-123"

	// Mock expectations - account validation fails
	mockAccountClient.On("ValidateAndGetAccounts",
		ctx,
		[]string{sourceAccountID, destAccountID},
		requester,
		requestID,
	).Return(nil, "Account not found", errors.New("validation failed"))

	// Execute
	transaction, msg, err := initTransaction.Execute(
		ctx,
		sourceAccountID,
		&destAccountID,
		amount,
		transactionType,
		referenceID,
		requester,
		requestID,
	)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, transaction)
	assert.Equal(t, "Failed to validate accounts", msg)
	mockAccountClient.AssertExpectations(t)
}

// TestInitTransaction_Execute_AccountDetailsNotFound tests account details not found scenario
func TestInitTransaction_Execute_AccountDetailsNotFound(t *testing.T) {
	mockTransactionRepo := new(mocks.MockTransactionRepo)
	mockAccountClient := new(mocks.MockAccountClient)
	mockSagaRepo := new(mocks.MockSagaRepo)
	mockEventRepo := new(mocks.MockEventRepo)

	initTransaction := NewInitTransaction(
		mockTransactionRepo,
		mockAccountClient,
		mockSagaRepo,
		mockEventRepo,
	)

	ctx := context.Background()
	sourceAccountID := "acc-123"
	destAccountID := "acc-456"
	amount := 100.0
	transactionType := entity.TransactionTypeTransfer
	referenceID := "ref-123"
	requester := "user-1"
	requestID := "req-123"

	// Mock expectations - return empty accounts info
	mockAccountClient.On("ValidateAndGetAccounts",
		ctx,
		[]string{sourceAccountID, destAccountID},
		requester,
		requestID,
	).Return([]ports.AccountInfo{}, "", nil)

	// Execute
	transaction, msg, err := initTransaction.Execute(
		ctx,
		sourceAccountID,
		&destAccountID,
		amount,
		transactionType,
		referenceID,
		requester,
		requestID,
	)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, transaction)
	assert.Equal(t, "Accounts details not founds", msg)
	mockAccountClient.AssertExpectations(t)
}

// TestInitTransaction_Execute_InsufficientBalance tests insufficient balance scenario
func TestInitTransaction_Execute_InsufficientBalance(t *testing.T) {
	mockTransactionRepo := new(mocks.MockTransactionRepo)
	mockAccountClient := new(mocks.MockAccountClient)
	mockSagaRepo := new(mocks.MockSagaRepo)
	mockEventRepo := new(mocks.MockEventRepo)

	initTransaction := NewInitTransaction(
		mockTransactionRepo,
		mockAccountClient,
		mockSagaRepo,
		mockEventRepo,
	)

	ctx := context.Background()
	sourceAccountID := "acc-123"
	destAccountID := "acc-456"
	amount := 600.0 // More than available balance
	transactionType := entity.TransactionTypeTransfer
	referenceID := "ref-123"
	requester := "user-1"
	requestID := "req-123"

	accountsInfo := []ports.AccountInfo{
		{AccountID: sourceAccountID, CustomerID: "cust-123", Balance: 500.0}, // Only 500 available
		{AccountID: destAccountID, CustomerID: "cust-456", Balance: 200.0},
	}

	// Mock expectations
	mockAccountClient.On("ValidateAndGetAccounts",
		ctx,
		[]string{sourceAccountID, destAccountID},
		requester,
		requestID,
	).Return(accountsInfo, "", nil)

	// Execute
	transaction, msg, err := initTransaction.Execute(
		ctx,
		sourceAccountID,
		&destAccountID,
		amount,
		transactionType,
		referenceID,
		requester,
		requestID,
	)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, transaction)
	assert.Equal(t, "Source account has insufficient balance", msg)
	mockAccountClient.AssertExpectations(t)
}

// TestInitTransaction_Execute_GetTransactionStatusFailure tests failure to get transaction status after saga success
func TestInitTransaction_Execute_GetTransactionStatusFailure(t *testing.T) {
	mockTransactionRepo := new(mocks.MockTransactionRepo)
	mockAccountClient := new(mocks.MockAccountClient)
	mockSagaRepo := new(mocks.MockSagaRepo)
	mockEventRepo := new(mocks.MockEventRepo)

	initTransaction := NewInitTransaction(
		mockTransactionRepo,
		mockAccountClient,
		mockSagaRepo,
		mockEventRepo,
	)

	ctx := context.Background()
	sourceAccountID := "acc-123"
	destAccountID := "acc-456"
	amount := 100.0
	transactionType := entity.TransactionTypeTransfer
	referenceID := "ref-123"
	requester := "user-1"
	requestID := "req-123"

	accountsInfo := []ports.AccountInfo{
		{AccountID: sourceAccountID, CustomerID: "cust-123", Balance: 500.0},
		{AccountID: destAccountID, CustomerID: "cust-456", Balance: 200.0},
	}

	// Mock expectations
	mockAccountClient.On("ValidateAndGetAccounts",
		ctx,
		[]string{sourceAccountID, destAccountID},
		requester,
		requestID,
	).Return(accountsInfo, "", nil)

	mockTransactionRepo.On("CreateTransaction", mock.AnythingOfType("*entity.Transaction")).
		Return(nil)

	// Mock successful saga execution
	mockSagaRepo.On("CreateSaga", mock.AnythingOfType("*entity.TransactionSaga")).Return(nil)
	mockAccountClient.On("ValidateAndGetAccounts", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(accountsInfo, "", nil)
	mockAccountClient.On("LockAccounts", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return("", nil)
	mockAccountClient.On("UpdateAccountsBalance", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return([]ports.AccountBalanceUpdateResponse{}, "", nil)
	mockAccountClient.On("UnlockAccounts", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return("", nil)
	mockSagaRepo.On("UpdateSaga", mock.AnythingOfType("*entity.TransactionSaga")).Return(nil)
	mockTransactionRepo.On("UpdateTransactionStatus", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// Transaction status retrieval fails
	mockTransactionRepo.On("GetTransactionByID", mock.Anything).
		Return(nil, errors.New("database error"))

	// Execute
	transaction, msg, err := initTransaction.Execute(
		ctx,
		sourceAccountID,
		&destAccountID,
		amount,
		transactionType,
		referenceID,
		requester,
		requestID,
	)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, transaction)
	assert.Equal(t, "Failed to get transaction status", msg)
	mockTransactionRepo.AssertExpectations(t)
	mockAccountClient.AssertExpectations(t)
	mockSagaRepo.AssertExpectations(t)
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
