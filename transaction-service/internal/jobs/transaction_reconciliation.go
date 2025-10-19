package jobs

import (
	"context"
	"fmt"
	"time"
	"transaction-service/internal/app/saga"
	"transaction-service/internal/config"
	"transaction-service/internal/domain/entity"
	custom_err "transaction-service/internal/domain/error"
	"transaction-service/internal/logging"
	"transaction-service/internal/ports"
)

type TransactionReconciliationJob struct {
	transactionRepo  ports.TransactionRepo
	accountClient    ports.AccountClient
	sagaOrchestrator *saga.TransactionSagaOrchestrator
}

func NewTransactionReconciliationJob(
	transactionRepo ports.TransactionRepo,
	accountClient ports.AccountClient,
	sagaOrchestrator *saga.TransactionSagaOrchestrator,
) *TransactionReconciliationJob {

	return &TransactionReconciliationJob{
		transactionRepo:  transactionRepo,
		accountClient:    accountClient,
		sagaOrchestrator: sagaOrchestrator,
	}
}

func (j *TransactionReconciliationJob) Start(ctx context.Context) {
	if err := j.RecoverStuckTransactions(ctx); err != nil {
		fmt.Printf("Initial transaction recovery failed: %v\n", err)
	}

	ticker := time.NewTicker(config.Current().Recovery.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := j.RecoverStuckTransactions(ctx); err != nil {
				fmt.Printf("Periodic transaction recovery failed: %v\n", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (j *TransactionReconciliationJob) RecoverStuckTransactions(ctx context.Context) error {
	stuckTransactions, err := j.transactionRepo.GetStuckTransactions()
	if err != nil {
		return fmt.Errorf("failed to get stuck transactions: %w", err)
	}

	for _, transaction := range stuckTransactions {
		if err := j.RecoverSingleTransaction(ctx, transaction); err != nil {
			fmt.Printf("Failed to recover transaction %s: %v\n", transaction.ID, err)
		}
	}
	return nil
}

func (j *TransactionReconciliationJob) RecoverSingleTransaction(ctx context.Context, transaction *entity.Transaction) error {
	message, err := j.accountClient.UnlockAccounts(ctx, transaction.ID, "system", "")
	if err != nil {
		logging.Logger.Warn().
			Err(err).
			Str("transaction_id", transaction.ID).
			Str("transaction_type", transaction.Type).
			Str("message", message).
			Str("job_type", "transaction_recovery").
			Msg("Failed to unlock accounts for transaction")
		return custom_err.ErrAccountUnlockingFailed
	}

	return j.transactionRepo.UpdateTransactionStatus(
		transaction.ID,
		entity.TransactionStatusFailed,
		"recovery: transaction timeout",
	)
}
