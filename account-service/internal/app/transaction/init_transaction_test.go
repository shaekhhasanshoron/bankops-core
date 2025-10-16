package transaction

import (
	"account-service/internal/domain/entity"
	mock_repo "account-service/internal/ports/mocks/repo"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

// TestInitTransaction_Execute_SuccessTransfer tests if transfer is initiated
func TestInitTransaction_Execute_SuccessTransfer(t *testing.T) {
	mockTransactionRepo := new(mock_repo.MockTransactionRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	initTransaction := NewInitTransaction(mockTransactionRepo, mockAccountRepo, mockEventRepo)

	sourceAccountID := "acc-123"
	destinationAccountID := "acc-456"
	amount := 100.0
	transactionType := entity.TransactionTypeTransfer
	referenceID := "ref-789"
	requester := "user123"
	requestId := "req-456"

	mockTransactionRepo.On("GetTransactionByReferenceID", referenceID).Return(nil, nil)

	sourceAccount := &entity.Account{ID: sourceAccountID, Balance: 500.0}
	destAccount := &entity.Account{ID: destinationAccountID, Balance: 200.0}

	mockAccountRepo.On("GetAccountByID", sourceAccountID).Return(sourceAccount, nil)
	mockAccountRepo.On("GetAccountByID", destinationAccountID).Return(destAccount, nil)
	mockTransactionRepo.On("CreateTransaction", mock.AnythingOfType("*entity.Transaction")).Return(nil)
	mockTransactionRepo.On("BeginTransactionLifecycle", mock.Anything, mock.Anything).Return(nil)
	mockEventRepo.On("CreateEvent", mock.AnythingOfType("*entity.Event")).Return(nil)

	transaction, message, err := initTransaction.Execute(
		sourceAccountID, &destinationAccountID, amount, transactionType, referenceID, requester, requestId,
	)

	assert.NoError(t, err)
	assert.Equal(t, "Transaction initialized successfully", message)
	assert.NotNil(t, transaction)
	assert.Equal(t, transactionType, transaction.Type)
	assert.Equal(t, amount, transaction.Amount)

	mockTransactionRepo.AssertExpectations(t)
	mockAccountRepo.AssertExpectations(t)
	mockEventRepo.AssertExpectations(t)
}

// TestInitTransaction_Execute_SuccessWithdrawAmount test for withdraw amount
func TestInitTransaction_Execute_SuccessWithdrawAmount(t *testing.T) {
	mockTransactionRepo := new(mock_repo.MockTransactionRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	initTransaction := NewInitTransaction(mockTransactionRepo, mockAccountRepo, mockEventRepo)

	sourceAccountID := "acc-123"
	var destinationAccountID *string = nil
	amount := 50.0
	transactionType := entity.TransactionTypeWithdrawAmount
	referenceID := "ref-789"
	requester := "user123"
	requestId := "req-456"

	mockTransactionRepo.On("GetTransactionByReferenceID", referenceID).Return(nil, nil)
	sourceAccount := &entity.Account{ID: sourceAccountID, Balance: 500.0}
	mockAccountRepo.On("GetAccountByID", sourceAccountID).Return(sourceAccount, nil)
	mockTransactionRepo.On("CreateTransaction", mock.AnythingOfType("*entity.Transaction")).Return(nil)
	mockTransactionRepo.On("BeginTransactionLifecycle", mock.Anything, mock.Anything).Return(nil)
	mockEventRepo.On("CreateEvent", mock.AnythingOfType("*entity.Event")).Return(nil)

	transaction, message, err := initTransaction.Execute(
		sourceAccountID, destinationAccountID, amount, transactionType, referenceID, requester, requestId,
	)

	assert.NoError(t, err)
	assert.Equal(t, "Transaction initialized successfully", message)
	assert.NotNil(t, transaction)
	assert.Equal(t, transactionType, transaction.Type)
	assert.Equal(t, amount, transaction.Amount)

	mockTransactionRepo.AssertExpectations(t)
}

// TestInitTransaction_Execute_SuccessWithdrawFull tests success for withdraw full amount
func TestInitTransaction_Execute_SuccessWithdrawFull(t *testing.T) {
	mockTransactionRepo := new(mock_repo.MockTransactionRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	initTransaction := NewInitTransaction(mockTransactionRepo, mockAccountRepo, mockEventRepo)

	sourceAccountID := "acc-123"
	var destinationAccountID *string = nil
	amount := 100.0
	transactionType := entity.TransactionTypeWithdrawFull
	referenceID := "ref-789"
	requester := "user123"
	requestId := "req-456"

	mockTransactionRepo.On("GetTransactionByReferenceID", referenceID).Return(nil, nil)
	sourceAccount := &entity.Account{ID: sourceAccountID, Balance: 500.0}
	mockAccountRepo.On("GetAccountByID", sourceAccountID).Return(sourceAccount, nil)
	mockTransactionRepo.On("CreateTransaction", mock.AnythingOfType("*entity.Transaction")).Return(nil)
	mockTransactionRepo.On("BeginTransactionLifecycle", mock.Anything, mock.Anything).Return(nil)
	mockEventRepo.On("CreateEvent", mock.AnythingOfType("*entity.Event")).Return(nil)

	transaction, message, err := initTransaction.Execute(
		sourceAccountID, destinationAccountID, amount, transactionType, referenceID, requester, requestId,
	)

	assert.NoError(t, err)
	assert.Equal(t, "Transaction initialized successfully", message)
	assert.NotNil(t, transaction)
	assert.Equal(t, transactionType, transaction.Type)
	assert.Equal(t, 0.0, transaction.Amount)

	mockTransactionRepo.AssertExpectations(t)
}

// TestInitTransaction_Execute_SuccessAddAmount tests success response for adding amount
func TestInitTransaction_Execute_SuccessAddAmount(t *testing.T) {
	mockTransactionRepo := new(mock_repo.MockTransactionRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	initTransaction := NewInitTransaction(mockTransactionRepo, mockAccountRepo, mockEventRepo)

	sourceAccountID := "acc-123"
	var destinationAccountID *string = nil
	amount := 200.0
	transactionType := entity.TransactionTypeAddAmount
	referenceID := "ref-789"
	requester := "user123"
	requestId := "req-456"

	mockTransactionRepo.On("GetTransactionByReferenceID", referenceID).Return(nil, nil)
	sourceAccount := &entity.Account{ID: sourceAccountID, Balance: 500.0}
	mockAccountRepo.On("GetAccountByID", sourceAccountID).Return(sourceAccount, nil)
	mockTransactionRepo.On("CreateTransaction", mock.AnythingOfType("*entity.Transaction")).Return(nil)
	mockTransactionRepo.On("BeginTransactionLifecycle", mock.Anything, mock.Anything).Return(nil)
	mockEventRepo.On("CreateEvent", mock.AnythingOfType("*entity.Event")).Return(nil)

	transaction, message, err := initTransaction.Execute(
		sourceAccountID, destinationAccountID, amount, transactionType, referenceID, requester, requestId,
	)

	assert.NoError(t, err)
	assert.Equal(t, "Transaction initialized successfully", message)
	assert.NotNil(t, transaction)
	assert.Equal(t, transactionType, transaction.Type)
	assert.Equal(t, amount, transaction.Amount)

	mockTransactionRepo.AssertExpectations(t)
}

// TestInitTransaction_Execute_ErrorTransferMultipleValidations tests error if input are not provided correctly
func TestInitTransaction_Execute_ErrorTransferMultipleValidations(t *testing.T) {
	mockTransactionRepo := new(mock_repo.MockTransactionRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	initTransaction := NewInitTransaction(mockTransactionRepo, mockAccountRepo, mockEventRepo)

	sourceAccountID := "acc-123"
	var destinationAccountID *string = nil
	amount := 0.0
	transactionType := entity.TransactionTypeTransfer
	referenceID := "ref-789"
	requester := "user123"
	requestId := "req-456"

	mockTransactionRepo.On("GetTransactionByReferenceID", referenceID).Return(nil, nil)

	transaction, message, err := initTransaction.Execute(
		sourceAccountID, destinationAccountID, amount, transactionType, referenceID, requester, requestId,
	)

	assert.Error(t, err)
	assert.Contains(t, message, "Destination account required")
	assert.NotContains(t, message, "Invalid amount")
	assert.Nil(t, transaction)

	mockTransactionRepo.AssertExpectations(t)
	mockAccountRepo.AssertNotCalled(t, "GetAccountByID")
}

// TestInitTransaction_Execute_ErrorEmptySourceAccountID tests error if source account id not provided
func TestInitTransaction_Execute_ErrorEmptySourceAccountID(t *testing.T) {
	mockTransactionRepo := new(mock_repo.MockTransactionRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	initTransaction := NewInitTransaction(mockTransactionRepo, mockAccountRepo, mockEventRepo)

	sourceAccountID := ""
	destinationAccountID := "acc-456"
	amount := 100.0
	transactionType := entity.TransactionTypeTransfer
	referenceID := "ref-789"
	requester := "user123"
	requestId := "req-456"

	transaction, message, err := initTransaction.Execute(
		sourceAccountID, &destinationAccountID, amount, transactionType, referenceID, requester, requestId,
	)

	assert.Error(t, err)
	assert.Contains(t, message, "Source Account ID missing")
	assert.Nil(t, transaction)

	mockTransactionRepo.AssertNotCalled(t, "GetTransactionByReferenceID")
	mockTransactionRepo.AssertNotCalled(t, "CreateTransaction")
}

// TestInitTransaction_Execute_ErrorEmptyReferenceID if reference id not provided
func TestInitTransaction_Execute_ErrorEmptyReferenceID(t *testing.T) {
	mockTransactionRepo := new(mock_repo.MockTransactionRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	initTransaction := NewInitTransaction(mockTransactionRepo, mockAccountRepo, mockEventRepo)

	sourceAccountID := "acc-123"
	destinationAccountID := "acc-456"
	amount := 100.0
	transactionType := entity.TransactionTypeTransfer
	referenceID := ""
	requester := "user123"
	requestId := "req-456"

	transaction, message, err := initTransaction.Execute(
		sourceAccountID, &destinationAccountID, amount, transactionType, referenceID, requester, requestId,
	)

	assert.Error(t, err)
	assert.Contains(t, message, "Reference ID missing")
	assert.Nil(t, transaction)

	mockTransactionRepo.AssertNotCalled(t, "GetTransactionByReferenceID")
	mockTransactionRepo.AssertNotCalled(t, "CreateTransaction")
}

// TestInitTransaction_Execute_ErrorEmptyRequester tests if empty provider is provided
func TestInitTransaction_Execute_ErrorEmptyRequester(t *testing.T) {
	mockTransactionRepo := new(mock_repo.MockTransactionRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	initTransaction := NewInitTransaction(mockTransactionRepo, mockAccountRepo, mockEventRepo)

	sourceAccountID := "acc-123"
	destinationAccountID := "acc-456"
	amount := 100.0
	transactionType := entity.TransactionTypeTransfer
	referenceID := "ref-789"
	requester := ""
	requestId := "req-456"

	transaction, message, err := initTransaction.Execute(
		sourceAccountID, &destinationAccountID, amount, transactionType, referenceID, requester, requestId,
	)

	assert.Error(t, err)
	assert.Contains(t, message, "Unknown requester")
	assert.Nil(t, transaction)

	mockTransactionRepo.AssertNotCalled(t, "GetTransactionByReferenceID")
	mockTransactionRepo.AssertNotCalled(t, "CreateTransaction")
}

// TestInitTransaction_Execute_ErrorDuplicateReferenceID tests if duplicate reference id already exists
func TestInitTransaction_Execute_ErrorDuplicateReferenceID(t *testing.T) {
	mockTransactionRepo := new(mock_repo.MockTransactionRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	initTransaction := NewInitTransaction(mockTransactionRepo, mockAccountRepo, mockEventRepo)

	sourceAccountID := "acc-123"
	destinationAccountID := "acc-456"
	amount := 100.0
	transactionType := entity.TransactionTypeTransfer
	referenceID := "ref-789"
	requester := "user123"
	requestId := "req-456"

	existingTransaction := &entity.Transaction{ID: "txn-123", ReferenceID: referenceID}
	mockTransactionRepo.On("GetTransactionByReferenceID", referenceID).Return(existingTransaction, nil)

	transaction, message, err := initTransaction.Execute(
		sourceAccountID, &destinationAccountID, amount, transactionType, referenceID, requester, requestId,
	)

	assert.Error(t, err)
	assert.Contains(t, message, "Duplicate transaction reference")
	assert.Nil(t, transaction)

	mockTransactionRepo.AssertExpectations(t)
	mockAccountRepo.AssertNotCalled(t, "GetAccountByID")
	mockTransactionRepo.AssertNotCalled(t, "CreateTransaction")
}

// TestInitTransaction_Execute_ErrorDatabaseCheckReference tests if database throws error while checking reference
func TestInitTransaction_Execute_ErrorDatabaseCheckReference(t *testing.T) {
	mockTransactionRepo := new(mock_repo.MockTransactionRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	initTransaction := NewInitTransaction(mockTransactionRepo, mockAccountRepo, mockEventRepo)

	sourceAccountID := "acc-123"
	destinationAccountID := "acc-456"
	amount := 100.0
	transactionType := entity.TransactionTypeTransfer
	referenceID := "ref-789"
	requester := "user123"
	requestId := "req-456"

	mockTransactionRepo.On("GetTransactionByReferenceID", referenceID).Return(nil, fmt.Errorf("database error"))

	transaction, message, err := initTransaction.Execute(
		sourceAccountID, &destinationAccountID, amount, transactionType, referenceID, requester, requestId,
	)

	assert.Error(t, err)
	assert.Contains(t, message, "Failed to check duplicate reference")
	assert.Nil(t, transaction)

	mockTransactionRepo.AssertExpectations(t)
	mockAccountRepo.AssertNotCalled(t, "GetAccountByID")
	mockTransactionRepo.AssertNotCalled(t, "CreateTransaction")
}

// TestInitTransaction_Execute_ErrorTransferMissingDestination tests if destination account id missing when transfer_type=transfer
func TestInitTransaction_Execute_ErrorTransferMissingDestination(t *testing.T) {
	mockTransactionRepo := new(mock_repo.MockTransactionRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	initTransaction := NewInitTransaction(mockTransactionRepo, mockAccountRepo, mockEventRepo)

	sourceAccountID := "acc-123"
	var destinationAccountID *string = nil
	amount := 100.0
	transactionType := entity.TransactionTypeTransfer
	referenceID := "ref-789"
	requester := "user123"
	requestId := "req-456"

	mockTransactionRepo.On("GetTransactionByReferenceID", referenceID).Return(nil, nil)

	transaction, message, err := initTransaction.Execute(
		sourceAccountID, destinationAccountID, amount, transactionType, referenceID, requester, requestId,
	)

	assert.Error(t, err)
	assert.Contains(t, message, "Destination account required")
	assert.Nil(t, transaction)

	mockTransactionRepo.AssertExpectations(t)
	mockAccountRepo.AssertExpectations(t)
	mockTransactionRepo.AssertNotCalled(t, "CreateTransaction")
}

// TestInitTransaction_Execute_ErrorTransferDestinationNotFound tests error if destination account id missing while transfer_type=transfer
func TestInitTransaction_Execute_ErrorTransferDestinationNotFound(t *testing.T) {
	mockTransactionRepo := new(mock_repo.MockTransactionRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	initTransaction := NewInitTransaction(mockTransactionRepo, mockAccountRepo, mockEventRepo)

	sourceAccountID := "acc-123"
	destinationAccountID := "acc-456"
	amount := 100.0
	transactionType := entity.TransactionTypeTransfer
	referenceID := "ref-789"
	requester := "user123"
	requestId := "req-456"

	mockTransactionRepo.On("GetTransactionByReferenceID", referenceID).Return(nil, nil)
	mockAccountRepo.On("GetAccountByID", destinationAccountID).Return(nil, nil) // Destination not found

	transaction, message, err := initTransaction.Execute(
		sourceAccountID, &destinationAccountID, amount, transactionType, referenceID, requester, requestId,
	)

	assert.Error(t, err)
	assert.Contains(t, message, "Destination account not found")
	assert.Nil(t, transaction)

	mockTransactionRepo.AssertExpectations(t)
	mockAccountRepo.AssertExpectations(t)
	mockTransactionRepo.AssertNotCalled(t, "CreateTransaction")
}

// TestInitTransaction_Execute_ErrorTransferInvalidAmount tests if invalid amount is provided while making transfer_type=transfer
func TestInitTransaction_Execute_ErrorTransferInvalidAmount(t *testing.T) {
	mockTransactionRepo := new(mock_repo.MockTransactionRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	initTransaction := NewInitTransaction(mockTransactionRepo, mockAccountRepo, mockEventRepo)

	sourceAccountID := "acc-123"
	destinationAccountID := "acc-456"
	amount := 0.0
	transactionType := entity.TransactionTypeTransfer
	referenceID := "ref-789"
	requester := "user123"
	requestId := "req-456"

	mockTransactionRepo.On("GetTransactionByReferenceID", referenceID).Return(nil, nil)
	destAccount := &entity.Account{ID: destinationAccountID, Balance: 200.0}
	mockAccountRepo.On("GetAccountByID", destinationAccountID).Return(destAccount, nil)

	transaction, message, err := initTransaction.Execute(
		sourceAccountID, &destinationAccountID, amount, transactionType, referenceID, requester, requestId,
	)

	assert.Error(t, err)
	assert.Contains(t, message, "Invalid amount for transaction")
	assert.Nil(t, transaction)

	mockTransactionRepo.AssertExpectations(t)
	mockAccountRepo.AssertExpectations(t)
	mockTransactionRepo.AssertNotCalled(t, "CreateTransaction")
}

// TestInitTransaction_Execute_ErrorWithdrawAmountInvalidAmount tests if invalid amount is provided when transfer_type=withdraw_amount
func TestInitTransaction_Execute_ErrorWithdrawAmountInvalidAmount(t *testing.T) {
	mockTransactionRepo := new(mock_repo.MockTransactionRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	initTransaction := NewInitTransaction(mockTransactionRepo, mockAccountRepo, mockEventRepo)

	sourceAccountID := "acc-123"
	var destinationAccountID *string = nil
	amount := 0.0
	transactionType := entity.TransactionTypeWithdrawAmount
	referenceID := "ref-789"
	requester := "user123"
	requestId := "req-456"

	mockTransactionRepo.On("GetTransactionByReferenceID", referenceID).Return(nil, nil)

	transaction, message, err := initTransaction.Execute(
		sourceAccountID, destinationAccountID, amount, transactionType, referenceID, requester, requestId,
	)

	assert.Error(t, err)
	assert.Contains(t, message, "Invalid amount for transaction")
	assert.Nil(t, transaction)

	mockTransactionRepo.AssertExpectations(t)
	mockAccountRepo.AssertExpectations(t)
	mockTransactionRepo.AssertNotCalled(t, "CreateTransaction")
}

// TestInitTransaction_Execute_ErrorAddAmountInvalidAmount tests error for invalid amount
func TestInitTransaction_Execute_ErrorAddAmountInvalidAmount(t *testing.T) {
	mockTransactionRepo := new(mock_repo.MockTransactionRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	initTransaction := NewInitTransaction(mockTransactionRepo, mockAccountRepo, mockEventRepo)

	sourceAccountID := "acc-123"
	var destinationAccountID *string = nil
	amount := 0.0 // Invalid amount for add
	transactionType := entity.TransactionTypeAddAmount
	referenceID := "ref-789"
	requester := "user123"
	requestId := "req-456"

	mockTransactionRepo.On("GetTransactionByReferenceID", referenceID).Return(nil, nil)

	transaction, message, err := initTransaction.Execute(
		sourceAccountID, destinationAccountID, amount, transactionType, referenceID, requester, requestId,
	)

	assert.Error(t, err)
	assert.Contains(t, message, "Invalid amount for transaction")
	assert.Nil(t, transaction)

	mockTransactionRepo.AssertExpectations(t)
	mockAccountRepo.AssertExpectations(t)
	mockTransactionRepo.AssertNotCalled(t, "CreateTransaction")
}

// TestInitTransaction_Execute_ErrorSourceAccountNotFound tests if source account not found
func TestInitTransaction_Execute_ErrorSourceAccountNotFound(t *testing.T) {
	mockTransactionRepo := new(mock_repo.MockTransactionRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	initTransaction := NewInitTransaction(mockTransactionRepo, mockAccountRepo, mockEventRepo)

	sourceAccountID := "acc-123"
	var destinationAccountID *string = nil
	amount := 100.0
	transactionType := entity.TransactionTypeAddAmount
	referenceID := "ref-789"
	requester := "user123"
	requestId := "req-456"

	mockTransactionRepo.On("GetTransactionByReferenceID", referenceID).Return(nil, nil)
	mockAccountRepo.On("GetAccountByID", sourceAccountID).Return(nil, nil) // Source account not found

	transaction, message, err := initTransaction.Execute(
		sourceAccountID, destinationAccountID, amount, transactionType, referenceID, requester, requestId,
	)

	assert.Error(t, err)
	assert.Contains(t, message, "Source account not found")
	assert.Nil(t, transaction)

	mockTransactionRepo.AssertExpectations(t)
	mockAccountRepo.AssertExpectations(t)
	mockTransactionRepo.AssertNotCalled(t, "CreateTransaction")
}

// TestInitTransaction_Execute_ErrorWithdrawFullEmptyAccount tests error while withdraw full amount when account already empty
func TestInitTransaction_Execute_ErrorWithdrawFullEmptyAccount(t *testing.T) {
	mockTransactionRepo := new(mock_repo.MockTransactionRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	initTransaction := NewInitTransaction(mockTransactionRepo, mockAccountRepo, mockEventRepo)

	sourceAccountID := "acc-123"
	var destinationAccountID *string = nil
	amount := 100.0
	transactionType := entity.TransactionTypeWithdrawFull
	referenceID := "ref-789"
	requester := "user123"
	requestId := "req-456"

	mockTransactionRepo.On("GetTransactionByReferenceID", referenceID).Return(nil, nil)
	sourceAccount := &entity.Account{ID: sourceAccountID, Balance: 0.0}
	mockAccountRepo.On("GetAccountByID", sourceAccountID).Return(sourceAccount, nil)

	transaction, message, err := initTransaction.Execute(
		sourceAccountID, destinationAccountID, amount, transactionType, referenceID, requester, requestId,
	)

	assert.Error(t, err)
	assert.Contains(t, message, "Account already empty")
	assert.Nil(t, transaction)

	mockTransactionRepo.AssertExpectations(t)
	mockAccountRepo.AssertExpectations(t)
	mockTransactionRepo.AssertNotCalled(t, "CreateTransaction")
}

// TestInitTransaction_Execute_ErrorCreateTransaction tests database throws an error while creating a new transaction
func TestInitTransaction_Execute_ErrorCreateTransaction(t *testing.T) {
	mockTransactionRepo := new(mock_repo.MockTransactionRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	initTransaction := NewInitTransaction(mockTransactionRepo, mockAccountRepo, mockEventRepo)

	sourceAccountID := "acc-123"
	var destinationAccountID *string = nil
	amount := 100.0
	transactionType := entity.TransactionTypeAddAmount
	referenceID := "ref-789"
	requester := "user123"
	requestId := "req-456"

	mockTransactionRepo.On("GetTransactionByReferenceID", referenceID).Return(nil, nil)
	sourceAccount := &entity.Account{ID: sourceAccountID, Balance: 500.0}
	mockAccountRepo.On("GetAccountByID", sourceAccountID).Return(sourceAccount, nil)
	mockTransactionRepo.On("CreateTransaction", mock.AnythingOfType("*entity.Transaction")).Return(fmt.Errorf("database error"))

	transaction, message, err := initTransaction.Execute(
		sourceAccountID, destinationAccountID, amount, transactionType, referenceID, requester, requestId,
	)

	assert.Error(t, err)
	assert.Contains(t, message, "Failed to create transaction")
	assert.Nil(t, transaction)

	mockTransactionRepo.AssertExpectations(t)
	mockAccountRepo.AssertExpectations(t)
	mockTransactionRepo.AssertNotCalled(t, "BeginTransactionLifecycle")
}

// TestInitTransaction_Execute_ErrorBeginTransactionLifecycle tests error while transaction lifecycle starts
func TestInitTransaction_Execute_ErrorBeginTransactionLifecycle(t *testing.T) {
	mockTransactionRepo := new(mock_repo.MockTransactionRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	initTransaction := NewInitTransaction(mockTransactionRepo, mockAccountRepo, mockEventRepo)

	sourceAccountID := "acc-123"
	var destinationAccountID *string = nil
	amount := 100.0
	transactionType := entity.TransactionTypeAddAmount
	referenceID := "ref-789"
	requester := "user123"
	requestId := "req-456"

	mockTransactionRepo.On("GetTransactionByReferenceID", referenceID).Return(nil, nil)
	sourceAccount := &entity.Account{ID: sourceAccountID, Balance: 500.0}
	mockAccountRepo.On("GetAccountByID", sourceAccountID).Return(sourceAccount, nil)
	mockTransactionRepo.On("CreateTransaction", mock.AnythingOfType("*entity.Transaction")).Return(nil)
	mockTransactionRepo.On("BeginTransactionLifecycle", mock.Anything, mock.Anything).Return(fmt.Errorf("locking error"))
	mockTransactionRepo.On("UpdateTransactionStatus", mock.Anything, entity.TransactionStatusFailed, mock.Anything).Return(nil)

	transaction, message, err := initTransaction.Execute(
		sourceAccountID, destinationAccountID, amount, transactionType, referenceID, requester, requestId,
	)

	assert.Error(t, err)
	assert.Contains(t, message, "Failed to lock accounts for transaction")
	assert.Nil(t, transaction)

	mockTransactionRepo.AssertExpectations(t)
	mockAccountRepo.AssertExpectations(t)
}

// TestInitTransaction_Execute_ErrorEventCreationDoesNotFailTransaction tests transaction stay succeeded even when event didn't published error
func TestInitTransaction_Execute_ErrorEventCreationDoesNotFailTransaction(t *testing.T) {
	mockTransactionRepo := new(mock_repo.MockTransactionRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	initTransaction := NewInitTransaction(mockTransactionRepo, mockAccountRepo, mockEventRepo)

	sourceAccountID := "acc-123"
	var destinationAccountID *string = nil
	amount := 100.0
	transactionType := entity.TransactionTypeAddAmount
	referenceID := "ref-789"
	requester := "user123"
	requestId := "req-456"

	mockTransactionRepo.On("GetTransactionByReferenceID", referenceID).Return(nil, nil)
	sourceAccount := &entity.Account{ID: sourceAccountID, Balance: 500.0}
	mockAccountRepo.On("GetAccountByID", sourceAccountID).Return(sourceAccount, nil)
	mockTransactionRepo.On("CreateTransaction", mock.AnythingOfType("*entity.Transaction")).Return(nil)
	mockTransactionRepo.On("BeginTransactionLifecycle", mock.Anything, mock.Anything).Return(nil)
	mockEventRepo.On("CreateEvent", mock.AnythingOfType("*entity.Event")).Return(fmt.Errorf("event error"))

	transaction, message, err := initTransaction.Execute(
		sourceAccountID, destinationAccountID, amount, transactionType, referenceID, requester, requestId,
	)

	assert.NoError(t, err)
	assert.Equal(t, "Transaction initialized successfully", message)
	assert.NotNil(t, transaction)

	mockTransactionRepo.AssertExpectations(t)
	mockAccountRepo.AssertExpectations(t)
	mockEventRepo.AssertExpectations(t)
}
