package saga

import (
	"context"
	"fmt"
	"strings"
	"time"
	"transaction-service/internal/domain/entity"
	custom_err "transaction-service/internal/domain/error"
	"transaction-service/internal/logging"
	"transaction-service/internal/ports"
)

type TransactionSagaOrchestrator struct {
	sagaRepo               ports.SagaRepo
	accountClient          ports.AccountClient
	transactionRepo        ports.TransactionRepo
	eventRepo              ports.EventRepo
	sourceAccountInfo      ports.AccountInfo
	destinationAccountInfo *ports.AccountInfo
}

func NewTransactionSagaOrchestrator(
	sagaRepo ports.SagaRepo,
	accountClient ports.AccountClient,
	transactionRepo ports.TransactionRepo,
	eventRepo ports.EventRepo,
) *TransactionSagaOrchestrator {

	return &TransactionSagaOrchestrator{
		sagaRepo:        sagaRepo,
		accountClient:   accountClient,
		transactionRepo: transactionRepo,
		eventRepo:       eventRepo,
	}
}

func (o *TransactionSagaOrchestrator) ExecuteTransactionSync(
	ctx context.Context,
	transaction *entity.Transaction,
	requester, requestId string,
) error {

	saga := entity.NewTransactionSaga(
		transaction.ID,
		transaction.SourceAccountID,
		transaction.DestinationAccountID,
		transaction.Amount,
		transaction.Type,
		transaction.ReferenceID,
	)

	if err := o.sagaRepo.CreateSaga(saga); err != nil {
		logging.Logger.Error().
			Err(err).
			Str("saga_state", entity.TransactionSagaStateInitiated).
			Str("saga_step", entity.TransactionSagaStepInitiate).
			Str("saga_id", saga.ID).
			Str("transaction_id", transaction.ID).
			Str("transaction_type", transaction.Type).
			Msg("Failed to create transaction saga")
		return custom_err.ErrDatabase
	}

	var err error
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		err = o.executeSagaSteps(ctx, saga, requester, requestId)
		if err == nil {
			return nil
		}

		if o.shouldRetry(err) && i < maxRetries-1 {
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}

		return o.compensateSaga(ctx, saga, err, requester, requestId)
	}

	return err
}

func (o *TransactionSagaOrchestrator) executeSagaSteps(
	ctx context.Context,
	saga *entity.TransactionSaga,
	requester, requestId string,
) error {
	accountIDs := saga.GetAccountsToLock()

	// validation step
	err := o.sagaValidationStep(ctx, accountIDs, saga, requester, requestId)
	if err != nil {
		return err
	}

	// locking state
	err = o.sagaLockingStep(ctx, accountIDs, saga, requester, requestId)
	if err != nil {
		return err
	}

	// processing step (transaction will be processed in this step)
	err = o.sagaProcessStep(ctx, saga, requester, requestId)
	if err != nil {
		return err
	}

	// complete step (update transaction and unlock accounts)
	err = o.sagaCompleteStep(ctx, saga, requester, requestId)
	if err != nil {
		return err
	}
	return nil
}

func (o *TransactionSagaOrchestrator) sagaValidationStep(
	ctx context.Context,
	accountIDs []string,
	saga *entity.TransactionSaga,
	requester string,
	requestId string) error {

	// If accounts are already locked
	if saga.IsStepsCompletedSuccessfully(
		entity.TransactionSagaStepValidateAccounts,
		entity.TransactionSagaStepLockAccounts) {
		logging.Logger.Debug().
			Str("transaction_id", saga.TransactionID).
			Str("transaction_saga_id", saga.ID).
			Str("saga_step_type", entity.TransactionSagaStepLockAccounts).
			Str("transaction_type", saga.TransactionType).
			Msg("Accounts already locked, moving to next")
		return nil
	}

	var err error
	defer func() {
		if err == nil {
			_ = saga.AppendSuccessfulStep(entity.TransactionSagaStepValidateAccounts)
			saga.CurrentState = entity.TransactionSagaStateValidated
		} else {
			saga.CurrentState = entity.TransactionSagaStateValidationFailed
		}

		_ = o.updateSaga(saga)
	}()

	saga.CurrentStep = entity.TransactionSagaStepValidateAccounts
	saga.CurrentState = entity.TransactionSagaStateValidating
	_ = o.updateSaga(saga)

	accountsInfo, message, err := o.accountClient.ValidateAndGetAccounts(ctx, accountIDs, requester, requestId)
	if err != nil {
		logging.Logger.Error().
			Err(err).
			Str("transaction_id", saga.TransactionID).
			Str("transaction_saga_id", saga.ID).
			Str("transaction_type", saga.TransactionType).
			Str("accountIds", fmt.Sprintf("%v", accountIDs)).
			Str("message", message).
			Msg("Unable to validate accounts saga")
		return custom_err.ErrAccountValidationFailed
	}

	if accountsInfo == nil || len(accountsInfo) == 0 || len(accountsInfo) != len(accountIDs) {
		logging.Logger.Error().Err(custom_err.ErrMissingAccountInfo).
			Str("transaction_id", saga.TransactionID).
			Str("transaction_saga_id", saga.ID).
			Str("transaction_type", saga.TransactionType).
			Str("accountIds", fmt.Sprintf("%v", accountIDs)).
			Msg("Accounts details not found")
		return custom_err.ErrAccountDetailsMissing
	}

	var sourceAccount ports.AccountInfo
	var destinationAccount *ports.AccountInfo
	for _, accountInfo := range accountsInfo {
		if accountInfo.AccountID == saga.SourceAccountID {
			sourceAccount = accountInfo
		}

		if saga.TransactionType == entity.TransactionTypeTransfer &&
			accountInfo.AccountID == *saga.DestinationAccountID {

			destinationAccount = &accountInfo
		}
	}

	o.sourceAccountInfo = sourceAccount
	o.destinationAccountInfo = destinationAccount
	return nil
}

func (o *TransactionSagaOrchestrator) sagaLockingStep(
	ctx context.Context,
	accountIDs []string,
	saga *entity.TransactionSaga,
	requester string,
	requestId string) error {

	// If accounts are already locked
	if saga.IsStepsCompletedSuccessfully(entity.TransactionSagaStepLockAccounts) {
		logging.Logger.Debug().
			Str("transaction_id", saga.TransactionID).
			Str("transaction_saga_id", saga.ID).
			Str("saga_step_type", entity.TransactionSagaStepLockAccounts).
			Str("transaction_type", saga.TransactionType).
			Msg("Accounts already locked, moving to next")
		return nil
	}

	var err error
	defer func() {
		if err == nil {
			_ = saga.AppendSuccessfulStep(entity.TransactionSagaStepLockAccounts)
			saga.CurrentState = entity.TransactionSagaStateLocked
		} else {
			saga.CurrentState = entity.TransactionSagaStateLockFailed
		}

		_ = o.updateSaga(saga)
	}()

	saga.CurrentStep = entity.TransactionSagaStepLockAccounts
	saga.CurrentState = entity.TransactionSagaStateLocking
	_ = o.updateSaga(saga)

	message, err := o.accountClient.LockAccounts(ctx, accountIDs, saga.TransactionID, requester, requestId)
	if err != nil {
		logging.Logger.Error().
			Err(err).
			Str("transaction_id", saga.TransactionID).
			Str("transaction_saga_id", saga.ID).
			Str("transaction_type", saga.TransactionType).
			Str("accountIds", fmt.Sprintf("%v", accountIDs)).
			Str("message", message).
			Msg("Failed to lock accounts")
		return custom_err.ErrAccountLockingFailed
	}
	return nil
}

func (o *TransactionSagaOrchestrator) sagaProcessStep(
	ctx context.Context,
	saga *entity.TransactionSaga,
	requester string,
	requestId string) error {

	// If accounts are already locked
	if saga.IsStepsCompletedSuccessfully(entity.TransactionSagaStepProcessTransfer) {
		logging.Logger.Debug().
			Str("transaction_id", saga.TransactionID).
			Str("transaction_saga_id", saga.ID).
			Str("saga_step_type", entity.TransactionSagaStepProcessTransfer).
			Str("transaction_type", saga.TransactionType).
			Msg("Accounts amount has been transferred; moving to following step")
		return nil
	}

	var err error
	defer func() {
		if err == nil {
			_ = saga.AppendSuccessfulStep(entity.TransactionSagaStepProcessTransfer)
			saga.CurrentState = entity.TransactionSagaStateCompleted
		} else {
			saga.CurrentState = entity.TransactionSagaStateFailed
		}

		_ = o.updateSaga(saga)
	}()

	saga.CurrentStep = entity.TransactionSagaStepProcessTransfer
	saga.CurrentState = entity.TransactionSagaStateProcessing
	_ = o.updateSaga(saga)

	updates, err := o.calculateAndValidateBalanceUpdates(saga)
	if err != nil {
		logging.Logger.Warn().
			Err(err).
			Str("transaction_id", saga.TransactionID).
			Str("transaction_saga_id", saga.ID).
			Str("transaction_type", saga.TransactionType).
			Msg("failed to calculate and validate amounts")
		return custom_err.ErrAccountEmpty
	}

	_, message, err := o.accountClient.UpdateAccountsBalance(ctx, updates, requester, requestId)
	if err != nil {
		logging.Logger.Warn().
			Err(err).
			Str("transaction_id", saga.TransactionID).
			Str("transaction_saga_id", saga.ID).
			Str("transaction_type", saga.TransactionType).
			Str("message", message).
			Msg("failed to update account balance; job will unlock the accounts")
		return custom_err.ErrTransactionFailed
	}

	return nil
}

func (o *TransactionSagaOrchestrator) sagaCompleteStep(
	ctx context.Context,
	saga *entity.TransactionSaga,
	requester string,
	requestId string) error {

	var err error
	defer func() {
		if err == nil {
			_ = saga.AppendSuccessfulStep(entity.TransactionSagaStepComplete)
			saga.CurrentState = entity.TransactionSagaStateCompleted
		} else {
			saga.CurrentState = entity.TransactionSagaStateFailed
		}

		_ = o.updateSaga(saga)
	}()

	saga.CurrentStep = entity.TransactionSagaStepComplete
	saga.CurrentState = entity.TransactionSagaStateProcessing
	_ = o.updateSaga(saga)

	err = o.transactionRepo.UpdateTransactionStatus(saga.TransactionID, entity.TransactionStatusSuccessful, "")
	if err != nil {
		logging.Logger.Error().
			Err(err).
			Str("transaction_id", saga.TransactionID).
			Str("transaction_saga_id", saga.ID).
			Str("transaction_type", saga.TransactionType).
			Str("transaction_status", entity.TransactionStatusSuccessful).
			Msg("Failed to update transaction status")
		return custom_err.ErrTransactionStatusUpdateFailed
	}

	message, err := o.accountClient.UnlockAccounts(ctx, saga.TransactionID, requester, requestId)
	if err != nil {
		logging.Logger.Warn().
			Err(err).
			Str("transaction_id", saga.TransactionID).
			Str("transaction_saga_id", saga.ID).
			Str("transaction_type", saga.TransactionType).
			Str("message", message).
			Msg("failed to unlock accounts; Retry for a while otherwise, it will recover by job")
		return custom_err.ErrAccountUnlockingFailed
	}

	// This ensures that accounts are unlocked successfully
	err = o.transactionRepo.UpdateTransactionStatus(saga.TransactionID, entity.TransactionStatusCompleted, "")
	if err != nil {
		logging.Logger.Error().
			Err(err).
			Str("transaction_id", saga.TransactionID).
			Str("transaction_saga_id", saga.ID).
			Str("transaction_type", saga.TransactionType).
			Str("transaction_status", entity.TransactionStatusCompleted).
			Msg("Failed to update transaction status")
		return custom_err.ErrTransactionStatusUpdateFailed
	}
	return nil
}

func (o *TransactionSagaOrchestrator) calculateAndValidateBalanceUpdates(
	saga *entity.TransactionSaga,
) ([]ports.AccountBalanceUpdate, error) {

	var updates []ports.AccountBalanceUpdate

	switch saga.TransactionType {
	case entity.TransactionTypeTransfer:
		if saga.DestinationAccountID == nil {
			return nil, fmt.Errorf("destination account required for transfer")
		}
		if saga.Amount <= 0 {
			return nil, fmt.Errorf("invalid amount for transfer")
		}
		if o.sourceAccountInfo.Balance < saga.Amount {
			return nil, fmt.Errorf("insufficient balance in source account")
		}

		updates = append(updates, ports.AccountBalanceUpdate{
			AccountID:  saga.SourceAccountID,
			NewBalance: o.sourceAccountInfo.Balance - saga.Amount,
			Version:    o.sourceAccountInfo.Version,
		})

		updates = append(updates, ports.AccountBalanceUpdate{
			AccountID:  *saga.DestinationAccountID,
			NewBalance: o.destinationAccountInfo.Balance + saga.Amount,
			Version:    o.destinationAccountInfo.Version,
		})

	case entity.TransactionTypeWithdrawFull:
		if o.sourceAccountInfo.Balance <= 0 {
			return nil, fmt.Errorf("account already empty")
		}
		updates = append(updates, ports.AccountBalanceUpdate{
			AccountID:  saga.SourceAccountID,
			NewBalance: 0,
			Version:    o.sourceAccountInfo.Version,
		})

	case entity.TransactionTypeWithdrawAmount:
		if saga.Amount <= 0 {
			return nil, fmt.Errorf("invalid amount for withdrawal")
		}
		if o.sourceAccountInfo.Balance < saga.Amount {
			return nil, fmt.Errorf("insufficient balance")
		}
		updates = append(updates, ports.AccountBalanceUpdate{
			AccountID:  saga.SourceAccountID,
			NewBalance: o.sourceAccountInfo.Balance - saga.Amount,
			Version:    o.sourceAccountInfo.Version,
		})

	case entity.TransactionTypeAddAmount:
		if saga.Amount <= 0 {
			return nil, fmt.Errorf("invalid amount for addition")
		}
		updates = append(updates, ports.AccountBalanceUpdate{
			AccountID:  saga.SourceAccountID,
			NewBalance: o.sourceAccountInfo.Balance + saga.Amount,
			Version:    o.sourceAccountInfo.Version,
		})

	default:
		return nil, fmt.Errorf("invalid transaction type: %s", saga.TransactionType)
	}

	return updates, nil
}

func (o *TransactionSagaOrchestrator) compensateSaga(
	ctx context.Context,
	saga *entity.TransactionSaga,
	originalErr error,
	requester, requestId string,
) error {

	var err error
	defer func() {
		if err == nil {
			saga.CurrentState = entity.TransactionSagaStateCompensated
		} else {
			saga.CurrentState = entity.TransactionSagaStateFailed
		}

		_ = o.updateSaga(saga)
	}()

	saga.CurrentStep = entity.TransactionSagaStepCompensate
	saga.CurrentState = entity.TransactionSagaStateCompensating
	_ = o.updateSaga(saga)

	message, err := o.accountClient.UnlockAccounts(ctx, saga.TransactionID, requester, requestId)
	if err != nil {
		logging.Logger.Warn().
			Err(err).
			Str("transaction_id", saga.TransactionID).
			Str("transaction_saga_id", saga.ID).
			Str("transaction_type", saga.TransactionType).
			Str("message", message).
			Msg("Failed to unlock accounts during compensation; recovery job will handle it")
		return custom_err.ErrAccountUnlockingFailed
	}

	if err := o.transactionRepo.UpdateTransactionStatus(
		saga.TransactionID,
		entity.TransactionStatusFailed,
		fmt.Sprintf("compensated: %v", originalErr),
	); err != nil {
		logging.Logger.Error().
			Err(err).
			Str("transaction_id", saga.TransactionID).
			Str("transaction_saga_id", saga.ID).
			Str("transaction_type", saga.TransactionType).
			Str("message", message).
			Msg("Failed to mark transaction as failed after compensation")
		return custom_err.ErrFailedToMarkTransactionAsFailed
	}

	return originalErr
}

func (o *TransactionSagaOrchestrator) shouldRetry(err error) bool {
	errorStr := err.Error()
	transientPatterns := []string{
		"timeout",
		"unavailable",
		"network",
		"connection refused",
		"connection failed",
		"deadline exceeded",
		"temporary",
	}

	for _, pattern := range transientPatterns {
		if strings.Contains(strings.ToLower(errorStr), pattern) {
			return true
		}
	}
	return false
}

func (o *TransactionSagaOrchestrator) updateSaga(saga *entity.TransactionSaga) error {
	if err := o.sagaRepo.UpdateSaga(saga); err != nil {
		logging.Logger.Warn().
			Err(err).
			Str("saga_state", saga.CurrentState).
			Str("saga_step", saga.CurrentStep).
			Str("saga_id", saga.ID).
			Str("transaction_id", saga.TransactionID).
			Str("transaction_type", saga.TransactionType).
			Msg("Unable to update transaction saga")
	}
	return nil
}
