package transaction

import (
	"account-service/internal/config"
	"account-service/internal/domain/entity"
	custom_err "account-service/internal/domain/error"
	"account-service/internal/logging"
	"account-service/internal/observability/metrics"
	"account-service/internal/ports"
	"fmt"
	"time"
)

// IniTransaction is a use-case for initiating a transaction
type IniTransaction struct {
	AccountRepo     ports.AccountRepo
	EventRepo       ports.EventRepo
	TransactionRepo ports.TransactionRepo
}

// NewInitTransaction creates a new IniTransaction use-case
func NewInitTransaction(transactionRepo ports.TransactionRepo, accountRepo ports.AccountRepo, eventRepo ports.EventRepo) *IniTransaction {
	return &IniTransaction{
		AccountRepo:     accountRepo,
		EventRepo:       eventRepo,
		TransactionRepo: transactionRepo,
	}
}

func (t *IniTransaction) Execute(sourceAccountID string, destinationAccountID *string, amount float64, transactionType, referenceID, requester, requestId string) (*entity.Transaction, string, error) {
	metrics.IncRequestActive()
	defer metrics.DecRequestActive()

	var err error
	defer func() {
		metrics.RecordOperation("init_transaction", err)
	}()

	// set initial data
	if transactionType != entity.TransactionTypeTransfer {
		destinationAccountID = nil
	}

	if transactionType == entity.TransactionTypeWithdrawFull {
		amount = 0
	}

	// validation input
	msg, err := t.validateInput(sourceAccountID, referenceID, requester)
	if err != nil {
		return nil, msg, err
	}

	transaction, _ := entity.NewTransaction(sourceAccountID, destinationAccountID, amount, transactionType, referenceID, requester)
	if config.Current().Recovery.TransactionTimeout > 0 {
		transaction.TimeoutAt = time.Now().Add(config.Current().Recovery.TransactionTimeout)
	}

	// validating data by transaction type
	msg, err = t.validateTransactionType(transaction)
	if err != nil {
		return nil, msg, err
	}

	if err = transaction.Validate(); err != nil {
		return nil, "Invalid transaction: " + err.Error(), err
	}

	err = t.TransactionRepo.CreateTransaction(transaction)
	if err != nil {
		logging.Logger.Error().Err(err).
			Str("transaction_id", transaction.ID).
			Str("reference_id", referenceID).
			Str("type", transaction.Type).
			Msg("Failed to create transaction")
		return nil, "Failed to create transaction", fmt.Errorf("%w: failed to create transaction", custom_err.ErrDatabase)
	}

	// Lock accounts for the transaction
	accountsToLock := transaction.GetAccountsToLock()
	if err = t.TransactionRepo.BeginTransactionLifecycle(transaction.ID, accountsToLock); err != nil {
		cleanupErr := t.TransactionRepo.UpdateTransactionStatus(transaction.ID, entity.TransactionStatusFailed, err.Error())
		if cleanupErr != nil {
			logging.Logger.Error().Err(cleanupErr).Str("transaction_id", transaction.ID).Msg("Failed to cleanup transaction after locking failure")
		}

		logging.Logger.Error().Err(err).
			Str("transaction_id", transaction.ID).
			Strs("account_ids", accountsToLock).
			Str("type", transaction.Type).
			Msg("Failed to lock accounts for transaction")
		return nil, "Failed to lock accounts for transaction", fmt.Errorf("%w: failed to lock accounts for transaction", custom_err.ErrDatabase)
	}

	// Record event
	eventData := map[string]interface{}{
		"transaction_id":      transaction.ID,
		"source_account":      sourceAccountID,
		"destination_account": destinationAccountID,
		"amount":              amount,
		"type":                string(transactionType),
		"reference_id":        referenceID,
		"initiated_by":        requester,
		"request_id":          requestId,
	}

	event, eventErr := entity.NewEvent(entity.EventTypeTransactionInit, transaction.ID, entity.EventAggregateTypeTransaction, requester, eventData)
	if eventErr == nil {
		if createErr := t.EventRepo.CreateEvent(event); createErr != nil {
			logging.Logger.Error().Err(createErr).Str("transaction_id", transaction.ID).Msg("Failed to create transaction init event")
		}
	}

	logging.Logger.Info().
		Str("transaction_id", transaction.ID).
		Str("source_account", sourceAccountID).
		Str("destination_account", t.toString(destinationAccountID)).
		Float64("amount", amount).
		Str("type", transactionType).
		Str("reference_id", referenceID).
		Str("requested_by", requester).
		Msg("Transaction initialized successfully")

	return transaction, "Transaction initialized successfully", nil
}

func (t *IniTransaction) validateInput(sourceAccountID string, referenceID, requester string) (string, error) {
	if sourceAccountID == "" {
		err := fmt.Errorf("%w: source account ID is required", custom_err.ErrValidationFailed)
		logging.Logger.Warn().Err(err).Msg("Source Account ID missing")
		return "Source Account ID missing", err
	}

	if referenceID == "" {
		err := fmt.Errorf("%w", custom_err.ErrMissingReferenceID)
		logging.Logger.Warn().Err(err).Msg("Reference ID missing")
		return "Reference ID missing", err
	}

	if requester == "" {
		err := fmt.Errorf("%w: requester is required", custom_err.ErrValidationFailed)
		logging.Logger.Error().Err(err).Msg("Unknown requester")
		return "Unknown requester", err
	}

	existingTx, err := t.TransactionRepo.GetTransactionByReferenceID(referenceID)
	if err != nil {
		logging.Logger.Error().Err(err).Str("reference_id", referenceID).Msg("Failed to check duplicate reference")
		return "Failed to check duplicate reference", fmt.Errorf("%w: failed to check duplicate reference", custom_err.ErrDatabase)
	}

	if existingTx != nil {
		err := fmt.Errorf("%w", custom_err.ErrDuplicateReference)
		logging.Logger.Error().Err(err).Str("reference_id", referenceID).Msg("Duplicate transaction reference")
		return "Duplicate transaction reference", err
	}
	return "", nil
}

func (t *IniTransaction) validateTransactionType(transaction *entity.Transaction) (string, error) {
	switch transaction.Type {
	case entity.TransactionTypeTransfer:
		if transaction.DestinationAccountID == nil {
			err := fmt.Errorf("%w", custom_err.ErrMissingDestinationAccount)
			logging.Logger.Error().Err(err).Msg("Destination account required")
			return "Destination account required", err
		}

		destAccount, err := t.AccountRepo.GetAccountByID(*transaction.DestinationAccountID)
		if err != nil {
			logging.Logger.Error().Err(err).Msg("Failed to verify destination account")
			return "Destination account required", fmt.Errorf("%w: failed to verify destination account", custom_err.ErrDatabase)
		}

		if destAccount == nil {
			err := fmt.Errorf("%w", custom_err.ErrAccountNotFound)
			logging.Logger.Error().Err(err).Msg("Destination account not found")
			return "Destination account not found", err
		}

		if transaction.Amount <= 0 {
			err := fmt.Errorf("%w", custom_err.ErrInvalidAmount)
			logging.Logger.Error().Err(err).Msg("Invalid amount for transaction")
			return "Invalid amount for transaction", err
		}

	case entity.TransactionTypeWithdrawAmount:
		if transaction.Amount <= 0 {
			err := fmt.Errorf("%w", custom_err.ErrInvalidWithdrawAmount)
			logging.Logger.Error().Err(err).Msg("Invalid amount for transaction")
			return "Invalid amount for transaction", err
		}

	case entity.TransactionTypeAddAmount:
		if transaction.Amount <= 0 {
			err := fmt.Errorf("%w", custom_err.ErrInvalidAddAmount)
			logging.Logger.Error().Err(err).Msg("Invalid amount for transaction")
			return "Invalid amount for transaction", err
		}
	}

	sourceAccount, err := t.AccountRepo.GetAccountByID(transaction.SourceAccountID)
	if err != nil {
		err := fmt.Errorf("%w", custom_err.ErrInvalidAddAmount)
		logging.Logger.Error().Err(err).Msg("Failed to verify source account")
		return "Failed to verify source account", fmt.Errorf("%w: failed to verify source account", custom_err.ErrDatabase)
	}

	if sourceAccount == nil {
		err := fmt.Errorf("%w", custom_err.ErrAccountNotFound)
		logging.Logger.Error().Err(err).Msg("Source account not found")
		return "Source account not found", err
	}

	if transaction.Type == entity.TransactionTypeWithdrawFull && sourceAccount.Balance <= 0 {
		err := fmt.Errorf("%w", custom_err.ErrAccountEmpty)
		logging.Logger.Error().Err(err).Msg("Account already empty")
		return "Account already empty", err
	}

	return "", nil
}

func (t *IniTransaction) toString(str *string) string {
	if str == nil {
		return ""
	}
	return *str
}
