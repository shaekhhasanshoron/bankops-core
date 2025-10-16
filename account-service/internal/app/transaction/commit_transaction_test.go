package transaction

import (
	"account-service/internal/domain/entity"
	"account-service/internal/domain/value"
	mock_repo "account-service/internal/ports/mocks/repo"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

// TestCommitTransaction_Execute_SuccessTransfer tests successful transaction for valid inputs from transfer_type=transfer
func TestCommitTransaction_Execute_SuccessTransfer(t *testing.T) {
	mockTransactionRepo := new(mock_repo.MockTransactionRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	commitTransaction := NewCommitTransaction(mockTransactionRepo, mockAccountRepo, mockEventRepo)

	transactionID := "txn-123"
	requester := "user123"
	requestId := "req-456"

	destAccountID := "acc-456"
	tx := &entity.Transaction{
		ID:                   transactionID,
		Type:                 entity.TransactionTypeTransfer,
		SourceAccountID:      "acc-123",
		DestinationAccountID: &destAccountID,
		Amount:               100.0,
		TransactionStatus:    entity.TransactionStatusPending,
		CreatedBy:            "user123",
	}
	mockTransactionRepo.On("GetTransactionByID", transactionID).Return(tx, nil)

	sourceAccount := &entity.Account{
		ID:                  "acc-123",
		Balance:             500.0,
		LockedForTx:         true,
		ActiveTransactionID: &transactionID,
		Version:             1,
	}
	destAccount := &entity.Account{
		ID:                  "acc-456",
		Balance:             200.0,
		LockedForTx:         true,
		ActiveTransactionID: &transactionID,
		Version:             1,
	}
	accounts := []*entity.Account{sourceAccount, destAccount}
	mockAccountRepo.On("GetAccountsInTransaction", transactionID).Return(accounts, nil)
	mockAccountRepo.On("UpdateAccountBalance", "acc-123", 400.0, 1, requester).Return(nil)
	mockAccountRepo.On("UpdateAccountBalance", "acc-456", 300.0, 1, requester).Return(nil)
	mockTransactionRepo.On("UpdateTransaction", mock.AnythingOfType("*entity.Transaction")).Return(nil)
	mockTransactionRepo.On("CompleteTransactionLifecycle", transactionID).Return(nil)
	mockEventRepo.On("CreateEvent", mock.AnythingOfType("*entity.Event")).Return(nil)

	message, err := commitTransaction.Execute(transactionID, requester, requestId)

	assert.NoError(t, err)
	assert.Equal(t, "", message)
	mockTransactionRepo.AssertExpectations(t)
	mockAccountRepo.AssertExpectations(t)
	mockEventRepo.AssertExpectations(t)
}

// TestCommitTransaction_Execute_SuccessWithdrawAmount tests successful transaction for valid inputs from transfer_type=withdraw_amount
func TestCommitTransaction_Execute_SuccessWithdrawAmount(t *testing.T) {
	mockTransactionRepo := new(mock_repo.MockTransactionRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	commitTransaction := NewCommitTransaction(mockTransactionRepo, mockAccountRepo, mockEventRepo)

	transactionID := "txn-123"
	requester := "user123"
	requestId := "req-456"

	tx := &entity.Transaction{
		ID:                transactionID,
		Type:              entity.TransactionTypeWithdrawAmount,
		SourceAccountID:   "acc-123",
		Amount:            100.0,
		TransactionStatus: entity.TransactionStatusPending,
		CreatedBy:         "user123",
	}
	mockTransactionRepo.On("GetTransactionByID", transactionID).Return(tx, nil)

	sourceAccount := &entity.Account{
		ID:                  "acc-123",
		Balance:             500.0,
		LockedForTx:         true,
		ActiveTransactionID: &transactionID,
		Version:             1,
	}
	accounts := []*entity.Account{sourceAccount}
	mockAccountRepo.On("GetAccountsInTransaction", transactionID).Return(accounts, nil)

	mockAccountRepo.On("UpdateAccountBalance", "acc-123", 400.0, 1, requester).Return(nil)
	mockTransactionRepo.On("UpdateTransaction", mock.AnythingOfType("*entity.Transaction")).Return(nil)
	mockTransactionRepo.On("CompleteTransactionLifecycle", transactionID).Return(nil)
	mockEventRepo.On("CreateEvent", mock.AnythingOfType("*entity.Event")).Return(nil)

	message, err := commitTransaction.Execute(transactionID, requester, requestId)

	assert.NoError(t, err)
	assert.Equal(t, "", message)
	mockTransactionRepo.AssertExpectations(t)
}

// TestCommitTransaction_Execute_SuccessWithdrawFull tests successful transaction for valid inputs from transfer_type=withdraw_full
func TestCommitTransaction_Execute_SuccessWithdrawFull(t *testing.T) {
	mockTransactionRepo := new(mock_repo.MockTransactionRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	commitTransaction := NewCommitTransaction(mockTransactionRepo, mockAccountRepo, mockEventRepo)

	transactionID := "txn-123"
	requester := "user123"
	requestId := "req-456"

	tx := &entity.Transaction{
		ID:                transactionID,
		Type:              entity.TransactionTypeWithdrawFull,
		SourceAccountID:   "acc-123",
		Amount:            0.0,
		TransactionStatus: entity.TransactionStatusPending,
		CreatedBy:         "user123",
	}
	mockTransactionRepo.On("GetTransactionByID", transactionID).Return(tx, nil)

	sourceAccount := &entity.Account{
		ID:                  "acc-123",
		Balance:             500.0,
		LockedForTx:         true,
		ActiveTransactionID: &transactionID,
		Version:             1,
	}
	accounts := []*entity.Account{sourceAccount}
	mockAccountRepo.On("GetAccountsInTransaction", transactionID).Return(accounts, nil)

	mockAccountRepo.On("UpdateAccountBalance", "acc-123", 0.0, 1, requester).Return(nil)
	mockTransactionRepo.On("UpdateTransaction", mock.AnythingOfType("*entity.Transaction")).Return(nil)
	mockTransactionRepo.On("CompleteTransactionLifecycle", transactionID).Return(nil)
	mockEventRepo.On("CreateEvent", mock.AnythingOfType("*entity.Event")).Return(nil)

	message, err := commitTransaction.Execute(transactionID, requester, requestId)

	assert.NoError(t, err)
	assert.Equal(t, "", message)
	mockTransactionRepo.AssertExpectations(t)
}

// TestCommitTransaction_Execute_SuccessAddAmount tests successful transaction for valid inputs from transfer_type=add_amount
func TestCommitTransaction_Execute_SuccessAddAmount(t *testing.T) {
	mockTransactionRepo := new(mock_repo.MockTransactionRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	commitTransaction := NewCommitTransaction(mockTransactionRepo, mockAccountRepo, mockEventRepo)

	transactionID := "txn-123"
	requester := "user123"
	requestId := "req-456"

	tx := &entity.Transaction{
		ID:                transactionID,
		Type:              entity.TransactionTypeAddAmount,
		SourceAccountID:   "acc-123",
		Amount:            100.0,
		TransactionStatus: entity.TransactionStatusPending,
		CreatedBy:         "user123",
	}
	mockTransactionRepo.On("GetTransactionByID", transactionID).Return(tx, nil)

	sourceAccount := &entity.Account{
		ID:                  "acc-123",
		Balance:             500.0,
		LockedForTx:         true,
		ActiveTransactionID: &transactionID,
		Version:             1,
	}
	accounts := []*entity.Account{sourceAccount}
	mockAccountRepo.On("GetAccountsInTransaction", transactionID).Return(accounts, nil)

	mockAccountRepo.On("UpdateAccountBalance", "acc-123", 600.0, 1, requester).Return(nil)
	mockTransactionRepo.On("UpdateTransaction", mock.AnythingOfType("*entity.Transaction")).Return(nil)
	mockTransactionRepo.On("CompleteTransactionLifecycle", transactionID).Return(nil)
	mockEventRepo.On("CreateEvent", mock.AnythingOfType("*entity.Event")).Return(nil)

	message, err := commitTransaction.Execute(transactionID, requester, requestId)

	assert.NoError(t, err)
	assert.Equal(t, "", message)
	mockTransactionRepo.AssertExpectations(t)
}

// TestCommitTransaction_Execute_ErrorTransactionNotFound tests if transaction not found
func TestCommitTransaction_Execute_ErrorTransactionNotFound(t *testing.T) {
	mockTransactionRepo := new(mock_repo.MockTransactionRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	commitTransaction := NewCommitTransaction(mockTransactionRepo, mockAccountRepo, mockEventRepo)

	transactionID := "txn-not-found"
	requester := "user123"
	requestId := "req-456"

	mockTransactionRepo.On("GetTransactionByID", transactionID).Return(nil, nil)

	message, err := commitTransaction.Execute(transactionID, requester, requestId)

	assert.Error(t, err)
	assert.ErrorIs(t, err, value.ErrTransactionNotFound)
	assert.Contains(t, message, "Transaction not found")
	mockTransactionRepo.AssertExpectations(t)
	mockAccountRepo.AssertNotCalled(t, "GetAccountsInTransaction")
}

// TestCommitTransaction_Execute_ErrorGetTransactionDatabaseError tests if database throws any error while getting transaction
func TestCommitTransaction_Execute_ErrorGetTransactionDatabaseError(t *testing.T) {
	mockTransactionRepo := new(mock_repo.MockTransactionRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	commitTransaction := NewCommitTransaction(mockTransactionRepo, mockAccountRepo, mockEventRepo)

	transactionID := "txn-123"
	requester := "user123"
	requestId := "req-456"

	mockTransactionRepo.On("GetTransactionByID", transactionID).Return(nil, fmt.Errorf("database error"))

	message, err := commitTransaction.Execute(transactionID, requester, requestId)

	assert.Error(t, err)
	//assert.ErrorIs(t, err, value.ErrDatabase)
	assert.Contains(t, message, "Failed to get transaction")
	mockTransactionRepo.AssertExpectations(t)
	mockAccountRepo.AssertNotCalled(t, "GetAccountsInTransaction")
}

// TestCommitTransaction_Execute_ErrorTransactionAlreadyCompleted tests if transaction already completed
func TestCommitTransaction_Execute_ErrorTransactionAlreadyCompleted(t *testing.T) {
	mockTransactionRepo := new(mock_repo.MockTransactionRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	commitTransaction := NewCommitTransaction(mockTransactionRepo, mockAccountRepo, mockEventRepo)

	transactionID := "txn-123"
	requester := "user123"
	requestId := "req-456"

	tx := &entity.Transaction{
		ID:                transactionID,
		TransactionStatus: entity.TransactionStatusCompleted,
	}
	mockTransactionRepo.On("GetTransactionByID", transactionID).Return(tx, nil)

	message, err := commitTransaction.Execute(transactionID, requester, requestId)

	assert.NoError(t, err)
	assert.Equal(t, "Transaction already completed", message)
	mockTransactionRepo.AssertExpectations(t)
	mockAccountRepo.AssertNotCalled(t, "GetAccountsInTransaction")
}

// TestCommitTransaction_Execute_ErrorTransactionFailed test if transaction already failed
func TestCommitTransaction_Execute_ErrorTransactionFailed(t *testing.T) {
	mockTransactionRepo := new(mock_repo.MockTransactionRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	commitTransaction := NewCommitTransaction(mockTransactionRepo, mockAccountRepo, mockEventRepo)

	transactionID := "txn-123"
	requester := "user123"
	requestId := "req-456"

	tx := &entity.Transaction{
		ID:                transactionID,
		TransactionStatus: entity.TransactionStatusFailed,
	}
	mockTransactionRepo.On("GetTransactionByID", transactionID).Return(tx, nil)

	message, err := commitTransaction.Execute(transactionID, requester, requestId)

	assert.Error(t, err)
	assert.Contains(t, message, "Transaction failed")
	mockTransactionRepo.AssertExpectations(t)
	mockAccountRepo.AssertNotCalled(t, "GetAccountsInTransaction")
}

// TestCommitTransaction_Execute_ErrorGetAccountsInTransaction tests if database throws error while getting accounts in transaction
func TestCommitTransaction_Execute_ErrorGetAccountsInTransaction(t *testing.T) {
	mockTransactionRepo := new(mock_repo.MockTransactionRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	commitTransaction := NewCommitTransaction(mockTransactionRepo, mockAccountRepo, mockEventRepo)

	transactionID := "txn-123"
	requester := "user123"
	requestId := "req-456"

	tx := &entity.Transaction{
		ID:                   transactionID,
		Type:                 entity.TransactionTypeTransfer,
		SourceAccountID:      "acc-123",
		DestinationAccountID: toStringPtr("acc-456"),
		TransactionStatus:    entity.TransactionStatusPending,
	}
	mockTransactionRepo.On("GetTransactionByID", transactionID).Return(tx, nil)
	mockAccountRepo.On("GetAccountsInTransaction", transactionID).Return(nil, fmt.Errorf("database error"))

	message, err := commitTransaction.Execute(transactionID, requester, requestId)

	assert.Error(t, err)
	assert.Contains(t, message, "Failed to verify account locks")
	mockTransactionRepo.AssertExpectations(t)
	mockAccountRepo.AssertExpectations(t)
}

// TestCommitTransaction_Execute_ErrorAccountLockInconsistencyTransfer tests transaction error if accounts are locked
func TestCommitTransaction_Execute_ErrorAccountLockInconsistencyTransfer(t *testing.T) {
	mockTransactionRepo := new(mock_repo.MockTransactionRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	commitTransaction := NewCommitTransaction(mockTransactionRepo, mockAccountRepo, mockEventRepo)

	transactionID := "txn-123"
	requester := "user123"
	requestId := "req-456"

	tx := &entity.Transaction{
		ID:                   transactionID,
		Type:                 entity.TransactionTypeTransfer,
		SourceAccountID:      "acc-123",
		DestinationAccountID: toStringPtr("acc-456"),
		TransactionStatus:    entity.TransactionStatusPending,
	}
	mockTransactionRepo.On("GetTransactionByID", transactionID).Return(tx, nil)

	// Only one account returned, but transfer requires two
	accounts := []*entity.Account{
		{ID: "acc-123", LockedForTx: true, ActiveTransactionID: &transactionID},
	}
	mockAccountRepo.On("GetAccountsInTransaction", transactionID).Return(accounts, nil)

	message, err := commitTransaction.Execute(transactionID, requester, requestId)

	assert.Error(t, err)
	assert.Contains(t, message, "Transaction account locks are inconsistent")
	mockTransactionRepo.AssertExpectations(t)
	mockAccountRepo.AssertExpectations(t)
}

// TestCommitTransaction_Execute_ErrorAccountLockInconsistencySingle test account lock error
func TestCommitTransaction_Execute_ErrorAccountLockInconsistencySingle(t *testing.T) {
	mockTransactionRepo := new(mock_repo.MockTransactionRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	commitTransaction := NewCommitTransaction(mockTransactionRepo, mockAccountRepo, mockEventRepo)

	transactionID := "txn-123"
	requester := "user123"
	requestId := "req-456"

	tx := &entity.Transaction{
		ID:                transactionID,
		Type:              entity.TransactionTypeWithdrawAmount,
		SourceAccountID:   "acc-123",
		TransactionStatus: entity.TransactionStatusPending,
	}
	mockTransactionRepo.On("GetTransactionByID", transactionID).Return(tx, nil)

	accounts := []*entity.Account{}
	mockAccountRepo.On("GetAccountsInTransaction", transactionID).Return(accounts, nil)

	message, err := commitTransaction.Execute(transactionID, requester, requestId)

	assert.Error(t, err)
	assert.Contains(t, message, "Transaction account locks are inconsistent")
	mockTransactionRepo.AssertExpectations(t)
	mockAccountRepo.AssertExpectations(t)
}

// TestCommitTransaction_Execute_ErrorSourceAccountNotFound tests source account not found before transaction
func TestCommitTransaction_Execute_ErrorSourceAccountNotFound(t *testing.T) {
	mockTransactionRepo := new(mock_repo.MockTransactionRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	commitTransaction := NewCommitTransaction(mockTransactionRepo, mockAccountRepo, mockEventRepo)

	transactionID := "txn-123"
	requester := "user123"
	requestId := "req-456"

	tx := &entity.Transaction{
		ID:                transactionID,
		Type:              entity.TransactionTypeWithdrawAmount,
		SourceAccountID:   "acc-123",
		TransactionStatus: entity.TransactionStatusPending,
	}
	mockTransactionRepo.On("GetTransactionByID", transactionID).Return(tx, nil)

	accounts := []*entity.Account{
		{ID: "acc-456", LockedForTx: true, ActiveTransactionID: &transactionID}, // Wrong account
	}
	mockAccountRepo.On("GetAccountsInTransaction", transactionID).Return(accounts, nil)

	message, err := commitTransaction.Execute(transactionID, requester, requestId)

	assert.Error(t, err)
	assert.Contains(t, message, "Missing source account for transaction")
	mockTransactionRepo.AssertExpectations(t)
	mockAccountRepo.AssertExpectations(t)
}

// TestCommitTransaction_Execute_ErrorDestinationAccountNotFound if destination account not matches with the current transaction
func TestCommitTransaction_Execute_ErrorDestinationAccountNotFound(t *testing.T) {
	mockTransactionRepo := new(mock_repo.MockTransactionRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	commitTransaction := NewCommitTransaction(mockTransactionRepo, mockAccountRepo, mockEventRepo)

	transactionID := "txn-123"
	requester := "user123"
	requestId := "req-456"

	destID := "acc-456"
	tx := &entity.Transaction{
		ID:                   transactionID,
		Type:                 entity.TransactionTypeTransfer,
		SourceAccountID:      "acc-123",
		DestinationAccountID: &destID,
		TransactionStatus:    entity.TransactionStatusPending,
	}
	mockTransactionRepo.On("GetTransactionByID", transactionID).Return(tx, nil)

	accounts := []*entity.Account{
		{ID: "acc-123", LockedForTx: true, ActiveTransactionID: &transactionID},
		{ID: "dest-acc-123", LockedForTx: true, ActiveTransactionID: &transactionID},
	}
	mockAccountRepo.On("GetAccountsInTransaction", transactionID).Return(accounts, nil)

	message, err := commitTransaction.Execute(transactionID, requester, requestId)

	assert.Error(t, err)
	assert.Contains(t, message, "Missing destination account for transaction")
	mockTransactionRepo.AssertExpectations(t)
	mockAccountRepo.AssertExpectations(t)
}

// TestCommitTransaction_Execute_ErrorSourceAccountNotLocked tests amount locked
func TestCommitTransaction_Execute_ErrorSourceAccountNotLocked(t *testing.T) {
	mockTransactionRepo := new(mock_repo.MockTransactionRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	commitTransaction := NewCommitTransaction(mockTransactionRepo, mockAccountRepo, mockEventRepo)

	transactionID := "txn-123"
	requester := "user123"
	requestId := "req-456"

	tx := &entity.Transaction{
		ID:                transactionID,
		Type:              entity.TransactionTypeWithdrawAmount,
		SourceAccountID:   "acc-123",
		TransactionStatus: entity.TransactionStatusPending,
	}
	mockTransactionRepo.On("GetTransactionByID", transactionID).Return(tx, nil)

	sourceAccount := &entity.Account{
		ID:                  "acc-123",
		LockedForTx:         false,
		ActiveTransactionID: &transactionID,
	}
	accounts := []*entity.Account{sourceAccount}
	mockAccountRepo.On("GetAccountsInTransaction", transactionID).Return(accounts, nil)

	message, err := commitTransaction.Execute(transactionID, requester, requestId)

	assert.Error(t, err)
	assert.Contains(t, message, "Source account validation failed")
	mockTransactionRepo.AssertExpectations(t)
	mockAccountRepo.AssertExpectations(t)
}

// TestCommitTransaction_Execute_ErrorSourceAccountWrongTransaction tests for wrong transaction
func TestCommitTransaction_Execute_ErrorSourceAccountWrongTransaction(t *testing.T) {
	mockTransactionRepo := new(mock_repo.MockTransactionRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	commitTransaction := NewCommitTransaction(mockTransactionRepo, mockAccountRepo, mockEventRepo)

	transactionID := "txn-123"
	requester := "user123"
	requestId := "req-456"

	tx := &entity.Transaction{
		ID:                transactionID,
		Type:              entity.TransactionTypeWithdrawAmount,
		SourceAccountID:   "acc-123",
		TransactionStatus: entity.TransactionStatusPending,
	}
	mockTransactionRepo.On("GetTransactionByID", transactionID).Return(tx, nil)

	wrongTxID := "txn-other"
	sourceAccount := &entity.Account{
		ID:                  "acc-123",
		LockedForTx:         true,
		ActiveTransactionID: &wrongTxID,
	}
	accounts := []*entity.Account{sourceAccount}
	mockAccountRepo.On("GetAccountsInTransaction", transactionID).Return(accounts, nil)

	message, err := commitTransaction.Execute(transactionID, requester, requestId)

	assert.Error(t, err)
	assert.Contains(t, message, "Source account validation failed")
	mockTransactionRepo.AssertExpectations(t)
	mockAccountRepo.AssertExpectations(t)
}

// TestCommitTransaction_Execute_ErrorDestinationAccountNotLocked tests destination account failed
func TestCommitTransaction_Execute_ErrorDestinationAccountNotLocked(t *testing.T) {
	mockTransactionRepo := new(mock_repo.MockTransactionRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	commitTransaction := NewCommitTransaction(mockTransactionRepo, mockAccountRepo, mockEventRepo)

	transactionID := "txn-123"
	requester := "user123"
	requestId := "req-456"

	destID := "acc-456"
	tx := &entity.Transaction{
		ID:                   transactionID,
		Type:                 entity.TransactionTypeTransfer,
		SourceAccountID:      "acc-123",
		DestinationAccountID: &destID,
		TransactionStatus:    entity.TransactionStatusPending,
	}
	mockTransactionRepo.On("GetTransactionByID", transactionID).Return(tx, nil)

	sourceAccount := &entity.Account{
		ID:                  "acc-123",
		LockedForTx:         true,
		ActiveTransactionID: &transactionID,
	}
	destAccount := &entity.Account{
		ID:                  "acc-456",
		LockedForTx:         false,
		ActiveTransactionID: &transactionID,
	}
	accounts := []*entity.Account{sourceAccount, destAccount}
	mockAccountRepo.On("GetAccountsInTransaction", transactionID).Return(accounts, nil)

	message, err := commitTransaction.Execute(transactionID, requester, requestId)

	assert.Error(t, err)
	assert.Contains(t, message, "Destination account validation failed")
	mockTransactionRepo.AssertExpectations(t)
	mockAccountRepo.AssertExpectations(t)
}

// TestCommitTransaction_Execute_ErrorTransferInsufficientBalance tests for insufficient balance
func TestCommitTransaction_Execute_ErrorTransferInsufficientBalance(t *testing.T) {
	mockTransactionRepo := new(mock_repo.MockTransactionRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	commitTransaction := NewCommitTransaction(mockTransactionRepo, mockAccountRepo, mockEventRepo)

	transactionID := "txn-123"
	requester := "user123"
	requestId := "req-456"

	destID := "acc-456"
	tx := &entity.Transaction{
		ID:                   transactionID,
		Type:                 entity.TransactionTypeTransfer,
		SourceAccountID:      "acc-123",
		DestinationAccountID: &destID,
		Amount:               600.0, // More than balance
		TransactionStatus:    entity.TransactionStatusPending,
	}
	mockTransactionRepo.On("GetTransactionByID", transactionID).Return(tx, nil)

	sourceAccount := &entity.Account{
		ID:                  "acc-123",
		Balance:             500.0, // Insufficient
		LockedForTx:         true,
		ActiveTransactionID: &transactionID,
		Version:             1,
	}
	destAccount := &entity.Account{
		ID:                  "acc-456",
		Balance:             200.0,
		LockedForTx:         true,
		ActiveTransactionID: &transactionID,
		Version:             1,
	}
	accounts := []*entity.Account{sourceAccount, destAccount}
	mockAccountRepo.On("GetAccountsInTransaction", transactionID).Return(accounts, nil)
	mockTransactionRepo.On("UpdateTransactionStatus", transactionID, entity.TransactionStatusFailed, mock.Anything).Return(nil)

	message, err := commitTransaction.Execute(transactionID, requester, requestId)

	assert.Error(t, err)
	assert.Equal(t, value.ErrInsufficientBalance, err)
	assert.Contains(t, message, "Transaction execution failed")
	mockTransactionRepo.AssertExpectations(t)
	mockAccountRepo.AssertExpectations(t)
}

// TestCommitTransaction_Execute_ErrorWithdrawAmountInsufficientBalance tests transaction balance is larger than available account balance  for transfer_type=withdraw_amount
func TestCommitTransaction_Execute_ErrorWithdrawAmountInsufficientBalance(t *testing.T) {
	mockTransactionRepo := new(mock_repo.MockTransactionRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	commitTransaction := NewCommitTransaction(mockTransactionRepo, mockAccountRepo, mockEventRepo)

	transactionID := "txn-123"
	requester := "user123"
	requestId := "req-456"

	tx := &entity.Transaction{
		ID:                transactionID,
		Type:              entity.TransactionTypeWithdrawAmount,
		SourceAccountID:   "acc-123",
		Amount:            600.0,
		TransactionStatus: entity.TransactionStatusPending,
	}
	mockTransactionRepo.On("GetTransactionByID", transactionID).Return(tx, nil)

	sourceAccount := &entity.Account{
		ID:                  "acc-123",
		Balance:             500.0,
		LockedForTx:         true,
		ActiveTransactionID: &transactionID,
		Version:             1,
	}
	accounts := []*entity.Account{sourceAccount}
	mockAccountRepo.On("GetAccountsInTransaction", transactionID).Return(accounts, nil)
	mockTransactionRepo.On("UpdateTransactionStatus", transactionID, entity.TransactionStatusFailed, mock.Anything).Return(nil)

	message, err := commitTransaction.Execute(transactionID, requester, requestId)

	assert.Error(t, err)
	assert.Equal(t, value.ErrInsufficientBalance, err)
	assert.Contains(t, message, "Transaction execution failed")
	mockTransactionRepo.AssertExpectations(t)
	mockAccountRepo.AssertExpectations(t)
}

// TestCommitTransaction_Execute_ErrorWithdrawFullEmptyAccount test withdraw full amount but account already empty
func TestCommitTransaction_Execute_ErrorWithdrawFullEmptyAccount(t *testing.T) {
	mockTransactionRepo := new(mock_repo.MockTransactionRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	commitTransaction := NewCommitTransaction(mockTransactionRepo, mockAccountRepo, mockEventRepo)

	transactionID := "txn-123"
	requester := "user123"
	requestId := "req-456"

	tx := &entity.Transaction{
		ID:                transactionID,
		Type:              entity.TransactionTypeWithdrawFull,
		SourceAccountID:   "acc-123",
		TransactionStatus: entity.TransactionStatusPending,
	}
	mockTransactionRepo.On("GetTransactionByID", transactionID).Return(tx, nil)

	sourceAccount := &entity.Account{
		ID:                  "acc-123",
		Balance:             0.0,
		LockedForTx:         true,
		ActiveTransactionID: &transactionID,
		Version:             1,
	}
	accounts := []*entity.Account{sourceAccount}
	mockAccountRepo.On("GetAccountsInTransaction", transactionID).Return(accounts, nil)
	mockTransactionRepo.On("UpdateTransactionStatus", transactionID, entity.TransactionStatusFailed, mock.Anything).Return(nil)

	message, err := commitTransaction.Execute(transactionID, requester, requestId)

	assert.Error(t, err)
	assert.Equal(t, value.ErrAccountEmpty, err)
	assert.Contains(t, message, "Transaction execution failed")
	mockTransactionRepo.AssertExpectations(t)
	mockAccountRepo.AssertExpectations(t)
}

// TestCommitTransaction_Execute_ErrorTransferSourceBalanceUpdateFails test source balance update fails
func TestCommitTransaction_Execute_ErrorTransferSourceBalanceUpdateFails(t *testing.T) {
	mockTransactionRepo := new(mock_repo.MockTransactionRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	commitTransaction := NewCommitTransaction(mockTransactionRepo, mockAccountRepo, mockEventRepo)

	transactionID := "txn-123"
	requester := "user123"
	requestId := "req-456"

	destID := "acc-456"
	tx := &entity.Transaction{
		ID:                   transactionID,
		Type:                 entity.TransactionTypeTransfer,
		SourceAccountID:      "acc-123",
		DestinationAccountID: &destID,
		Amount:               100.0,
		TransactionStatus:    entity.TransactionStatusPending,
	}
	mockTransactionRepo.On("GetTransactionByID", transactionID).Return(tx, nil)

	sourceAccount := &entity.Account{
		ID:                  "acc-123",
		Balance:             500.0,
		LockedForTx:         true,
		ActiveTransactionID: &transactionID,
		Version:             1,
	}
	destAccount := &entity.Account{
		ID:                  "acc-456",
		Balance:             200.0,
		LockedForTx:         true,
		ActiveTransactionID: &transactionID,
		Version:             1,
	}
	accounts := []*entity.Account{sourceAccount, destAccount}
	mockAccountRepo.On("GetAccountsInTransaction", transactionID).Return(accounts, nil)
	mockTransactionRepo.On("UpdateTransactionStatus", transactionID, entity.TransactionStatusFailed, mock.Anything).Return(nil)

	mockAccountRepo.On("UpdateAccountBalance", "acc-123", 400.0, 1, requester).Return(fmt.Errorf("database error"))

	message, err := commitTransaction.Execute(transactionID, requester, requestId)

	assert.Error(t, err)
	assert.Contains(t, message, "Transaction execution failed")
	mockTransactionRepo.AssertExpectations(t)
	mockAccountRepo.AssertExpectations(t)
	mockTransactionRepo.AssertNotCalled(t, "UpdateTransaction")
}

// TestCommitTransaction_Execute_ErrorTransferDestinationBalanceUpdateFails if destination balance fails
func TestCommitTransaction_Execute_ErrorTransferDestinationBalanceUpdateFails(t *testing.T) {
	mockTransactionRepo := new(mock_repo.MockTransactionRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	commitTransaction := NewCommitTransaction(mockTransactionRepo, mockAccountRepo, mockEventRepo)

	transactionID := "txn-123"
	requester := "user123"
	requestId := "req-456"

	destID := "acc-456"
	tx := &entity.Transaction{
		ID:                   transactionID,
		Type:                 entity.TransactionTypeTransfer,
		SourceAccountID:      "acc-123",
		DestinationAccountID: &destID,
		Amount:               100.0,
		TransactionStatus:    entity.TransactionStatusPending,
	}
	mockTransactionRepo.On("GetTransactionByID", transactionID).Return(tx, nil)
	mockTransactionRepo.On("UpdateTransactionStatus", transactionID, entity.TransactionStatusFailed, mock.Anything).Return(nil)

	sourceAccount := &entity.Account{
		ID:                  "acc-123",
		Balance:             500.0,
		LockedForTx:         true,
		ActiveTransactionID: &transactionID,
		Version:             1,
	}
	destAccount := &entity.Account{
		ID:                  "acc-456",
		Balance:             200.0,
		LockedForTx:         true,
		ActiveTransactionID: &transactionID,
		Version:             1,
	}
	accounts := []*entity.Account{sourceAccount, destAccount}
	mockAccountRepo.On("GetAccountsInTransaction", transactionID).Return(accounts, nil)

	mockAccountRepo.On("UpdateAccountBalance", "acc-123", 400.0, 1, requester).Return(nil)
	mockAccountRepo.On("UpdateAccountBalance", "acc-456", 300.0, 1, requester).Return(fmt.Errorf("database error"))
	mockAccountRepo.On("UpdateAccountBalance", "acc-123", 500.0, 2, requester).Return(nil)

	message, err := commitTransaction.Execute(transactionID, requester, requestId)

	assert.Error(t, err)
	assert.Contains(t, message, "Transaction execution failed")
	mockTransactionRepo.AssertExpectations(t)
	mockAccountRepo.AssertExpectations(t)
	mockTransactionRepo.AssertNotCalled(t, "UpdateTransaction")
}

// TestCommitTransaction_Execute_ErrorUpdateTransactionFails tests error when transaction fails
func TestCommitTransaction_Execute_ErrorUpdateTransactionFails(t *testing.T) {
	mockTransactionRepo := new(mock_repo.MockTransactionRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	commitTransaction := NewCommitTransaction(mockTransactionRepo, mockAccountRepo, mockEventRepo)

	transactionID := "txn-123"
	requester := "user123"
	requestId := "req-456"

	tx := &entity.Transaction{
		ID:                transactionID,
		Type:              entity.TransactionTypeAddAmount,
		SourceAccountID:   "acc-123",
		Amount:            100.0,
		TransactionStatus: entity.TransactionStatusPending,
	}
	mockTransactionRepo.On("GetTransactionByID", transactionID).Return(tx, nil)

	sourceAccount := &entity.Account{
		ID:                  "acc-123",
		Balance:             500.0,
		LockedForTx:         true,
		ActiveTransactionID: &transactionID,
		Version:             1,
	}
	accounts := []*entity.Account{sourceAccount}
	mockAccountRepo.On("GetAccountsInTransaction", transactionID).Return(accounts, nil)

	mockAccountRepo.On("UpdateAccountBalance", "acc-123", 600.0, 1, requester).Return(nil)
	mockTransactionRepo.On("UpdateTransaction", mock.AnythingOfType("*entity.Transaction")).Return(fmt.Errorf("database error"))

	message, err := commitTransaction.Execute(transactionID, requester, requestId)

	assert.Error(t, err)
	assert.Contains(t, message, "Failed to update transaction status to completed")
	mockTransactionRepo.AssertExpectations(t)
	mockAccountRepo.AssertExpectations(t)
}

// TestCommitTransaction_Execute_ErrorCompleteLifecycleFails tests error when transaction lifecycle update failed
func TestCommitTransaction_Execute_ErrorCompleteLifecycleFails(t *testing.T) {
	mockTransactionRepo := new(mock_repo.MockTransactionRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	commitTransaction := NewCommitTransaction(mockTransactionRepo, mockAccountRepo, mockEventRepo)

	transactionID := "txn-123"
	requester := "user123"
	requestId := "req-456"

	tx := &entity.Transaction{
		ID:                transactionID,
		Type:              entity.TransactionTypeAddAmount,
		SourceAccountID:   "acc-123",
		Amount:            100.0,
		TransactionStatus: entity.TransactionStatusPending,
	}
	mockTransactionRepo.On("GetTransactionByID", transactionID).Return(tx, nil)

	sourceAccount := &entity.Account{
		ID:                  "acc-123",
		Balance:             500.0,
		LockedForTx:         true,
		ActiveTransactionID: &transactionID,
		Version:             1,
	}
	accounts := []*entity.Account{sourceAccount}
	mockAccountRepo.On("GetAccountsInTransaction", transactionID).Return(accounts, nil)

	mockAccountRepo.On("UpdateAccountBalance", "acc-123", 600.0, 1, requester).Return(nil)
	mockTransactionRepo.On("UpdateTransaction", mock.AnythingOfType("*entity.Transaction")).Return(nil)
	mockTransactionRepo.On("CompleteTransactionLifecycle", transactionID).Return(fmt.Errorf("unlock error"))
	mockEventRepo.On("CreateEvent", mock.AnythingOfType("*entity.Event")).Return(fmt.Errorf("event error"))

	message, err := commitTransaction.Execute(transactionID, requester, requestId)

	assert.NoError(t, err)
	assert.Equal(t, "", message)
	mockTransactionRepo.AssertExpectations(t)
	mockAccountRepo.AssertExpectations(t)
}

// TestCommitTransaction_Execute_ErrorInvalidTransactionType tests error on invalid transaction type
func TestCommitTransaction_Execute_ErrorInvalidTransactionType(t *testing.T) {
	mockTransactionRepo := new(mock_repo.MockTransactionRepo)
	mockAccountRepo := new(mock_repo.MockAccountRepo)
	mockEventRepo := new(mock_repo.MockEventRepo)

	commitTransaction := NewCommitTransaction(mockTransactionRepo, mockAccountRepo, mockEventRepo)

	transactionID := "txn-123"
	requester := "user123"
	requestId := "req-456"

	tx := &entity.Transaction{
		ID:                transactionID,
		Type:              "invalid-type",
		SourceAccountID:   "acc-123",
		TransactionStatus: entity.TransactionStatusPending,
	}
	mockTransactionRepo.On("GetTransactionByID", transactionID).Return(tx, nil)

	sourceAccount := &entity.Account{
		ID:                  "acc-123",
		Balance:             500.0,
		LockedForTx:         true,
		ActiveTransactionID: &transactionID,
		Version:             1,
	}
	accounts := []*entity.Account{sourceAccount}
	mockAccountRepo.On("GetAccountsInTransaction", transactionID).Return(accounts, nil)
	mockTransactionRepo.On("UpdateTransactionStatus", transactionID, entity.TransactionStatusFailed, mock.Anything).Return(nil)

	message, err := commitTransaction.Execute(transactionID, requester, requestId)

	assert.Error(t, err)
	assert.Contains(t, message, "Transaction execution failed")
	mockTransactionRepo.AssertExpectations(t)
	mockAccountRepo.AssertExpectations(t)
}

func toStringPtr(s string) *string {
	return &s
}
