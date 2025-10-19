package app

import (
	"context"
	"fmt"
	"strings"
	"time"
	"transaction-service/internal/app/saga"
	"transaction-service/internal/config"
	"transaction-service/internal/domain/entity"
	custom_err "transaction-service/internal/domain/error"
	"transaction-service/internal/logging"
	"transaction-service/internal/messaging"
	"transaction-service/internal/observability/metrics"
	"transaction-service/internal/ports"
)

type InitTransaction struct {
	transactionRepo ports.TransactionRepo
	accountClient   ports.AccountClient
	sagaRepo        ports.SagaRepo
	eventRepo       ports.EventRepo
}

func NewInitTransaction(
	transactionRepo ports.TransactionRepo,
	accountClient ports.AccountClient,
	sagaRepo ports.SagaRepo,
	eventRepo ports.EventRepo,
) *InitTransaction {

	return &InitTransaction{
		transactionRepo: transactionRepo,
		accountClient:   accountClient,
		sagaRepo:        sagaRepo,
		eventRepo:       eventRepo,
	}
}

func (a *InitTransaction) Execute(
	ctx context.Context,
	sourceAccountID string,
	destinationAccountID *string,
	amount float64,
	transactionType,
	referenceID,
	requester,
	requestId string,
) (*entity.Transaction, string, error) {
	metrics.IncRequestActive()
	defer metrics.DecRequestActive()

	var err error
	defer func() {
		metrics.RecordOperation("init_transaction", err)
	}()

	referenceID = strings.TrimSpace(referenceID)
	transactionType = strings.TrimSpace(transactionType)
	sourceAccountID = strings.TrimSpace(sourceAccountID)
	if destinationAccountID != nil {
		destAccountId := strings.TrimSpace(*destinationAccountID)
		destinationAccountID = &destAccountId
	}

	msg, err := a.validateInput(sourceAccountID, destinationAccountID, transactionType, referenceID, requester)
	if err != nil {
		return nil, msg, err
	}

	if transactionType != entity.TransactionTypeTransfer {
		destinationAccountID = nil
	}
	if transactionType == entity.TransactionTypeWithdrawFull {
		amount = 0
	}

	transaction, _ := entity.NewTransaction(sourceAccountID, destinationAccountID, amount, transactionType, referenceID, requester)
	if config.Current().Recovery.TransactionTimeout > 0 {
		transaction.TimeoutAt = time.Now().Add(config.Current().Recovery.TransactionTimeout)
	}

	msg, err = a.validateTransactionConfig(ctx, transaction, requester, requestId)
	if err != nil {
		return nil, msg, err
	}

	// creating transaction to db
	err = a.transactionRepo.CreateTransaction(transaction)
	if err != nil {
		logging.Logger.Error().
			Err(custom_err.ErrDatabase).
			Str("transaction_id", transaction.ID).
			Str("transaction_type", transaction.Type).
			Msg("Failed to create transaction")
		err = custom_err.ErrDatabase
		return nil, "Failed to create transaction", err
	}
	// Initiating Saga Orchestrator
	err = saga.NewTransactionSagaOrchestrator(
		a.sagaRepo,
		a.accountClient,
		a.transactionRepo,
		a.eventRepo,
	).ExecuteTransactionSync(ctx, transaction, requester, requestId)

	if err != nil {
		_ = a.transactionRepo.UpdateTransactionStatus(transaction.ID, entity.TransactionStatusFailed, err.Error())
		logging.Logger.Error().
			Err(err).
			Str("transaction_id", transaction.ID).
			Str("transaction_type", transaction.Type).
			Msg("Transaction failed")

		eventData := map[string]interface{}{
			"transaction_id": transaction.ID,
			"created_by":     requester,
			"request_id":     requestId,
		}

		event, eventErr := entity.NewEvent(entity.EventTypeTransactionFailed, transaction.ID, entity.EventAggregateTypeTransaction, requester, eventData)
		if eventErr == nil {
			if createErr := a.eventRepo.CreateEvent(event); createErr != nil {
				logging.Logger.Error().Err(createErr).Str("transaction_id", transaction.ID).
					Str("event_type", event.Type).Msg("Failed to create event")
			}
			_ = messaging.GetService().PublishToDefaultTopic(messaging.Message{Content: event.ToString(), Status: true, Type: messaging.MessageTypeTransactionFailed})
		}
		return nil, "Transaction failed: " + err.Error(), err
	}

	updatedTransaction, err := a.transactionRepo.GetTransactionByID(transaction.ID)
	if err != nil {
		err = fmt.Errorf("failed to get transaction status: %w", err)
		return nil, "Failed to get transaction status", err
	}

	eventData := map[string]interface{}{
		"transaction_id": transaction.ID,
		"created_by":     requester,
		"request_id":     requestId,
	}

	event, eventErr := entity.NewEvent(entity.EventTypeTransactionCompleted, transaction.ID, entity.EventAggregateTypeTransaction, requester, eventData)
	if eventErr == nil {
		if createErr := a.eventRepo.CreateEvent(event); createErr != nil {
			logging.Logger.Error().Err(createErr).Str("transaction_id", transaction.ID).
				Str("event_type", event.Type).Msg("Failed to create event")
		}
	}

	_ = messaging.GetService().PublishToDefaultTopic(messaging.Message{Content: updatedTransaction.ToString(), Status: true, Type: messaging.MessageTypeTransactionCompleted})
	return updatedTransaction, "Transaction completed successfully", nil
}

func (a *InitTransaction) validateInput(sourceAccountID string, destinationAccountID *string, transactionType, referenceID, requester string) (string, error) {
	sourceAccountID = strings.TrimSpace(sourceAccountID)
	if sourceAccountID == "" {
		logging.Logger.Error().Err(custom_err.ErrMissingSourceAccountId).Msg("Source Account ID required")
		return "Source Account ID missing", custom_err.ErrMissingSourceAccountId
	}
	referenceID = strings.TrimSpace(referenceID)
	if referenceID == "" {
		logging.Logger.Error().Err(custom_err.ErrMissingReferenceID).Msg("Reference ID missing")
		return "Reference ID missing", custom_err.ErrMissingReferenceID
	}

	requester = strings.TrimSpace(requester)
	if requester == "" {
		logging.Logger.Error().Err(custom_err.ErrMissingRequester).Msg("Unknown requester")
		return "Unknown requester", custom_err.ErrMissingRequester
	}

	validTypes := map[string]bool{
		entity.TransactionTypeTransfer:       true,
		entity.TransactionTypeWithdrawFull:   true,
		entity.TransactionTypeWithdrawAmount: true,
		entity.TransactionTypeAddAmount:      true,
	}

	if !validTypes[transactionType] {
		logging.Logger.Error().Err(custom_err.ErrInvalidTransactionType).Str("transaction_type", transactionType).Msg("Invalid transaction type")
		return "Invalid transaction type", custom_err.ErrInvalidTransactionType
	}

	if transactionType == entity.TransactionTypeTransfer {
		if destinationAccountID == nil || strings.TrimSpace(*destinationAccountID) == "" {
			logging.Logger.Error().Err(custom_err.ErrInvalidRequest).Msg("Destination Account ID required")
			return "Destination Account ID required", custom_err.ErrInvalidRequest
		}

		if sourceAccountID == strings.TrimSpace(*destinationAccountID) {
			logging.Logger.Error().Err(custom_err.ErrInvalidRequest).Msg("Cannot transfer amount to the same account")
			return "Cannot transfer amount to the same account", custom_err.ErrInvalidRequest
		}
	}
	return "", nil
}

func (a *InitTransaction) validateTransactionConfig(ctx context.Context, transaction *entity.Transaction, requester, requestId string) (string, error) {
	if transaction.Type != entity.TransactionTypeWithdrawFull && transaction.Amount <= 0 {
		logging.Logger.Error().Err(custom_err.ErrInvalidAmount).
			Str("transaction_type", transaction.Type).
			Str("transaction_amount", fmt.Sprintf("%v", transaction.Amount)).
			Msg("Invalid transaction amount")

		return "Invalid amount for transaction", custom_err.ErrInvalidAmount
	}

	var destAccountId string
	accountIds := []string{transaction.SourceAccountID}
	if transaction.Type == entity.TransactionTypeTransfer {
		destAccountId = *transaction.DestinationAccountID
		accountIds = append(accountIds, *transaction.DestinationAccountID)
	}

	accountsInfo, message, err := a.accountClient.ValidateAndGetAccounts(ctx, accountIds, requester, requestId)
	if err != nil {
		logging.Logger.Error().
			Err(err).
			Str("message", message).
			Str("transaction_type", transaction.Type).
			Str("source_account_id", transaction.SourceAccountID).
			Str("destination_account_id", destAccountId).
			Msg("Failed to validate accounts")
		return "Failed to validate accounts", custom_err.ErrValidationFailed
	}

	if accountsInfo == nil || len(accountsInfo) == 0 || len(accountsInfo) != len(accountIds) {
		logging.Logger.Error().Err(custom_err.ErrMissingAccountInfo).
			Str("transaction_type", transaction.Type).
			Str("source_account_id", transaction.SourceAccountID).
			Str("destination_account_id", destAccountId).
			Msg("Accounts details not found")
		return "Accounts details not founds", custom_err.ErrValidationFailed
	}

	var sourceAccount ports.AccountInfo
	for _, accountInfo := range accountsInfo {
		if accountInfo.AccountID == transaction.SourceAccountID {
			sourceAccount = accountInfo
			transaction.SourceAccountCustomerID = accountInfo.CustomerID
		}
		if transaction.Type == entity.TransactionTypeTransfer && accountInfo.AccountID == *transaction.DestinationAccountID {
			transaction.DestinationAccountCustomerID = &accountInfo.CustomerID
		}
	}

	if transaction.Type != entity.TransactionTypeWithdrawFull &&
		transaction.Type != entity.TransactionTypeAddAmount &&
		sourceAccount.Balance < transaction.Amount {

		logging.Logger.Error().Err(custom_err.ErrInsufficientBalance).
			Str("transaction_type", transaction.Type).
			Str("source_account_id", transaction.SourceAccountID).
			Msg("Accounts details not found")
		return "Source account has insufficient balance", custom_err.ErrInsufficientBalance
	}
	return "", nil
}
