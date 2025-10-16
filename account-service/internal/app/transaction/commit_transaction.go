package transaction

import (
	"account-service/internal/domain/entity"
	custom_err "account-service/internal/domain/error"
	"account-service/internal/logging"
	"account-service/internal/message_publisher"
	"account-service/internal/observability/metrics"
	"account-service/internal/ports"
	"errors"
	"fmt"
)

// CommitTransaction is a use-case for committing a transaction
type CommitTransaction struct {
	AccountRepo     ports.AccountRepo
	CustomerRepo    ports.CustomerRepo
	EventRepo       ports.EventRepo
	TransactionRepo ports.TransactionRepo
}

// NewCommitTransaction creates a new CommitTransaction use-case
func NewCommitTransaction(transactionRepo ports.TransactionRepo, accountRepo ports.AccountRepo, eventRepo ports.EventRepo) *CommitTransaction {
	return &CommitTransaction{
		AccountRepo:     accountRepo,
		EventRepo:       eventRepo,
		TransactionRepo: transactionRepo,
	}
}

func (t *CommitTransaction) Execute(transactionID, requester, requestId string) (string, error) {
	tx, msg, err := t.validateAndGetTransaction(transactionID)
	if err != nil {
		if errors.Is(err, custom_err.ErrTransactionCompleted) {
			return msg, nil
		}
		return msg, err
	}

	metrics.IncRequestActive()
	defer metrics.DecRequestActive()

	defer func() {
		metrics.RecordOperation("commit_transaction", err)
	}()

	// Verify accounts are still locked for the transaction
	accounts, err := t.AccountRepo.GetAccountsInTransaction(transactionID)
	if err != nil {
		logging.Logger.Error().Err(err).Str("transaction_id", transactionID).Msg("Failed to verify account locks")
		return "Failed to verify account locks", fmt.Errorf("%w: failed to verify account locks", custom_err.ErrDatabase)
	}

	expectedAccountCount := 1
	if tx.RequiresDestinationAccount() {
		expectedAccountCount = 2
	}

	if len(accounts) != expectedAccountCount {
		err = fmt.Errorf("transaction account locks are inconsistent: expected %d, got %d", expectedAccountCount, len(accounts))
		logging.Logger.Error().Err(err).Str("transaction_id", transactionID).Msg("Account lock inconsistency")
		return "Transaction account locks are inconsistent", err
	}

	// Find source account
	var sourceAccount *entity.Account
	for i := range accounts {
		if accounts[i].ID == tx.SourceAccountID {
			sourceAccount = accounts[i]
			break
		}
	}

	if sourceAccount == nil {
		err = fmt.Errorf("could not find source account for transaction: %s", tx.SourceAccountID)
		logging.Logger.Error().Err(err).Str("transaction_id", transactionID).Msg("Missing source account for transaction")
		return "Missing source account for transaction", err
	}

	// Find destination account for transfers
	var destinationAccount *entity.Account
	if tx.RequiresDestinationAccount() {
		for i := range accounts {
			if accounts[i].ID == *tx.DestinationAccountID {
				destinationAccount = accounts[i]
				break
			}
		}

		if destinationAccount == nil {
			err = fmt.Errorf("could not find destination account for transaction: %s", *tx.DestinationAccountID)
			logging.Logger.Error().Err(err).Str("transaction_id", transactionID).Msg("Missing destination account for transaction")
			return "Missing destination account for transaction", err
		}
	}

	// Validate accounts can still transact
	if !sourceAccount.LockedForTx || sourceAccount.ActiveTransactionID == nil || *sourceAccount.ActiveTransactionID != tx.ID {
		err = fmt.Errorf("source account cannot complete transaction: status=%s locked=%t", sourceAccount.Status, sourceAccount.LockedForTx)
		logging.Logger.Error().Err(err).Str("transaction_id", transactionID).Str("source_account_id", sourceAccount.ID).Msg("Source account validation failed")
		return "Source account validation failed", err
	}

	if destinationAccount != nil {
		if !destinationAccount.LockedForTx || destinationAccount.ActiveTransactionID == nil || *destinationAccount.ActiveTransactionID != tx.ID {
			err = fmt.Errorf("destination account cannot complete transaction: status=%s locked=%t", destinationAccount.Status, destinationAccount.LockedForTx)
			logging.Logger.Error().Err(err).Str("transaction_id", transactionID).Str("destination_account_id", destinationAccount.ID).Msg("Destination account validation failed")
			return "Destination account validation failed", err
		}
	}

	// Execute transaction based on type
	switch tx.Type {
	case entity.TransactionTypeTransfer:
		msg, err = t.executeTransfer(tx, sourceAccount, destinationAccount, requester)
	case entity.TransactionTypeWithdrawFull:
		msg, err = t.executeWithdrawFull(tx, sourceAccount, requester)
	case entity.TransactionTypeWithdrawAmount:
		msg, err = t.executeWithdrawAmount(tx, sourceAccount, requester)
	case entity.TransactionTypeAddAmount:
		msg, err = t.executeAddAmount(tx, sourceAccount, requester)
	default:
		msg, err = "Invalid transaction type: "+tx.Type, custom_err.ErrInvalidTransactionType
	}

	if err != nil {
		logging.Logger.Error().Err(err).
			Str("transaction_id", transactionID).
			Str("type", tx.Type).
			Msg("Transaction execution failed")

		if updateErr := t.TransactionRepo.UpdateTransactionStatus(transactionID, entity.TransactionStatusFailed, err.Error()); updateErr != nil {
			logging.Logger.Error().Err(updateErr).Str("transaction_id", transactionID).Msg("Failed to update transaction status after execution failure")
		}
		return "Transaction execution failed", err
	}

	// Update transaction status to completed
	tx.TransactionStatus = entity.TransactionStatusCompleted

	//logging.Logger.Info().
	//	Str("transaction_id", tx.ID).
	//	Str("current_status", tx.TransactionStatus).
	//	Int("current_version", tx.Version).
	//	Msg("Attempting to update transaction to completed")

	if err = t.TransactionRepo.UpdateTransaction(tx); err != nil {
		logging.Logger.Error().Err(err).
			Str("transaction_id", transactionID).
			Str("current_status", tx.TransactionStatus).
			Int("current_version", tx.Version).
			Msg("Failed to update transaction status to completed")
		return "Failed to update transaction status to completed", fmt.Errorf("%w: failed to update transaction status", custom_err.ErrDatabase)
	}

	// After exchanging balance and updating transaction status - complete the lifecycle
	if unlockErr := t.TransactionRepo.CompleteTransactionLifecycle(transactionID); unlockErr != nil {
		logging.Logger.Error().Err(unlockErr).
			Str("transaction_id", transactionID).
			Msg("Failed to complete transaction lifecycle")
	}

	// Record commit event
	eventData := map[string]interface{}{
		"transaction_id": tx.ID,
		"type":           tx.Type,
		"source_balance": sourceAccount.Balance,
	}

	if destinationAccount != nil {
		eventData["destination_balance"] = destinationAccount.Balance
	}

	event, eventErr := entity.NewEvent(entity.EventTypeTransactionCommit, tx.ID, entity.EventAggregateTypeTransaction, tx.CreatedBy, eventData)
	if eventErr == nil {
		if createErr := t.EventRepo.CreateEvent(event); createErr != nil {
			logging.Logger.Error().Err(createErr).Str("transaction_id", tx.ID).Msg("Failed to create transaction commit event")
		}
		_ = message_publisher.Publish(message_publisher.MessagePublishRequest{Message: event.ToString()})
	}

	logging.Logger.Info().
		Str("transaction_id", transactionID).
		Str("type", tx.Type).
		Str("source_account", sourceAccount.ID).
		Float64("amount", tx.Amount).
		Float64("source_new_balance", sourceAccount.Balance).
		Msg("Transaction completed successfully")

	return "", nil
}

func (t *CommitTransaction) validateAndGetTransaction(transactionID string) (*entity.Transaction, string, error) {
	tx, err := t.TransactionRepo.GetTransactionByID(transactionID)
	if err != nil {
		logging.Logger.Error().Err(err).Str("transaction_id", transactionID).Msg("Failed to get transaction")
		return nil, "Failed to get transaction", custom_err.ErrDatabase
	}
	if tx == nil {
		err := custom_err.ErrTransactionNotFound
		logging.Logger.Error().Err(err).Str("transaction_id", transactionID).Msg("Transaction not found")
		return nil, "Transaction not found", custom_err.ErrTransactionNotFound
	}

	if tx.TransactionStatus == entity.TransactionStatusCompleted {
		logging.Logger.Info().Err(err).Str("transaction_id", transactionID).Msg("Transaction already completed")
		return nil, "Transaction already completed", custom_err.ErrTransactionCompleted
	}

	if tx.TransactionStatus == entity.TransactionStatusFailed || tx.TransactionStatus == entity.TransactionStatusCancelled {
		err := fmt.Errorf("%w", custom_err.ErrTransactionFailed)
		logging.Logger.Error().Err(err).Str("transaction_id", transactionID).Msg("Transaction failed")
		return nil, "Transaction failed", err
	}
	return tx, "", nil
}

func (t *CommitTransaction) executeTransfer(tx *entity.Transaction, sourceAccount, destinationAccount *entity.Account, requester string) (string, error) {
	if sourceAccount.Balance < tx.Amount {
		return "Insufficient balance", custom_err.ErrInsufficientBalance
	}

	// Update balances
	newSourceBalance := sourceAccount.Balance - tx.Amount
	newDestBalance := destinationAccount.Balance + tx.Amount

	// Update source account balance
	if err := t.AccountRepo.UpdateAccountBalance(sourceAccount.ID, newSourceBalance, sourceAccount.Version, requester); err != nil {
		logging.Logger.Err(err).
			Str("account_id", sourceAccount.ID).
			Float64("old_balance", sourceAccount.Balance).
			Float64("new_balance", newSourceBalance).
			Msg("Failed to update source account balance")
		return "Failed to update source account balance", fmt.Errorf("%w: failed to update source account balance", custom_err.ErrDatabase)
	}

	// Update destination account balance
	if err := t.AccountRepo.UpdateAccountBalance(destinationAccount.ID, newDestBalance, destinationAccount.Version, requester); err != nil {
		// Rollback source account balance
		rollbackErr := t.AccountRepo.UpdateAccountBalance(sourceAccount.ID, sourceAccount.Balance, sourceAccount.Version+1, requester)
		if rollbackErr != nil {
			logging.Logger.Error().Err(rollbackErr).
				Str("account_id", sourceAccount.ID).
				Msg("CRITICAL: Failed to rollback source account balance after transfer failure")
			// Manual intervention will be required in such scenarios
		}

		logging.Logger.Error().Err(err).
			Str("account_id", destinationAccount.ID).
			Float64("old_balance", destinationAccount.Balance).
			Float64("new_balance", newDestBalance).
			Msg("Failed to update destination account balance, rolled back source account")
		return "Failed to update destination account balance, rolled back source account", fmt.Errorf("%w: failed to update destination account balance", custom_err.ErrDatabase)
	}

	return "", nil
}

func (t *CommitTransaction) executeWithdrawFull(tx *entity.Transaction, sourceAccount *entity.Account, requester string) (string, error) {
	if sourceAccount.Balance <= 0 {
		return "Account already empty", custom_err.ErrAccountEmpty
	}

	// Set amount to full balance and withdraw all
	newBalance := 0.0

	if err := t.AccountRepo.UpdateAccountBalance(sourceAccount.ID, newBalance, sourceAccount.Version, requester); err != nil {
		logging.Logger.Error().Err(err).
			Str("account_id", sourceAccount.ID).
			Float64("old_balance", sourceAccount.Balance).
			Float64("new_balance", newBalance).
			Msg("Failed to update account balance for full withdrawal")
		return "Failed to update account balance for full withdrawal", fmt.Errorf("%w: failed to update account balance", custom_err.ErrDatabase)
	}
	return "", nil
}

func (t *CommitTransaction) executeWithdrawAmount(tx *entity.Transaction, sourceAccount *entity.Account, requester string) (string, error) {
	if sourceAccount.Balance < tx.Amount {
		return "Insufficient balance", custom_err.ErrInsufficientBalance
	}

	newBalance := sourceAccount.Balance - tx.Amount

	if err := t.AccountRepo.UpdateAccountBalance(sourceAccount.ID, newBalance, sourceAccount.Version, requester); err != nil {
		logging.Logger.Error().Err(err).
			Str("account_id", sourceAccount.ID).
			Float64("old_balance", sourceAccount.Balance).
			Float64("new_balance", newBalance).
			Msg("Failed to update account balance for amount withdrawal")
		return "Failed to update account balance", fmt.Errorf("%w: failed to update account balance", custom_err.ErrDatabase)
	}

	return "", nil
}

func (t *CommitTransaction) executeAddAmount(tx *entity.Transaction, sourceAccount *entity.Account, requester string) (string, error) {
	newBalance := sourceAccount.Balance + tx.Amount

	if err := t.AccountRepo.UpdateAccountBalance(sourceAccount.ID, newBalance, sourceAccount.Version, requester); err != nil {
		logging.Logger.Error().Err(err).
			Str("account_id", sourceAccount.ID).
			Float64("old_balance", sourceAccount.Balance).
			Float64("new_balance", newBalance).
			Msg("Failed to update account balance for amount addition")
		return "Failed to update account balance", fmt.Errorf("%w: failed to update account balance", custom_err.ErrDatabase)
	}

	return "", nil
}
