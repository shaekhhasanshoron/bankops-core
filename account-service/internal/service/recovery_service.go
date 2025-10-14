package service

import (
	apptx "account-service/internal/app/transaction"
	"account-service/internal/config"
	"account-service/internal/domain/entity"
	"account-service/internal/logging"
	"account-service/internal/ports"
	"context"
	"fmt"
	"github.com/google/uuid"
	"time"
)

type RecoveryService struct {
	TransactionRepo ports.TransactionRepo
	AccountRepo     ports.AccountRepo
	CustomerRepo    ports.CustomerRepo
	EventRepo       ports.EventRepo
}

func NewRecoveryService(customerRepo ports.CustomerRepo, accountRepo ports.AccountRepo, transactionRepo ports.TransactionRepo, eventRepo ports.EventRepo) *RecoveryService {
	return &RecoveryService{
		TransactionRepo: transactionRepo,
		AccountRepo:     accountRepo,
		CustomerRepo:    customerRepo,
		EventRepo:       eventRepo,
	}
}

// Start begins the transaction recovery process
func (s *RecoveryService) Start(ctx context.Context) {
	logging.Logger.Info().Msg("Starting transaction recovery service")

	// Handle locked inconsistent states on startup (if money moved but transaction state not 'completed')
	if err := s.handleLockedInconsistentTransactions(); err != nil {
		logging.Logger.Error().Err(err).Msg("Failed to handle locked inconsistent transactions")
	}

	if err := s.RecoverStuckTransactions(ctx); err != nil {
		logging.Logger.Error().Err(err).Msg("Initial transaction recovery failed")
	}

	// Start periodic recovery
	ticker := time.NewTicker(config.Current().Recovery.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.handleLockedInconsistentTransactions(); err != nil {
				logging.Logger.Error().Err(err).Msg("Periodic locked inconsistent transactions check failed")
			}
			if err := s.RecoverStuckTransactions(ctx); err != nil {
				logging.Logger.Error().Err(err).Msg("Periodic transaction recovery failed")
			}
		case <-ctx.Done():
			logging.Logger.Info().Msg("Stopping transaction recovery service")
			return
		}
	}
}

// RecoverStuckTransactions recovers all stuck transactions
func (s *RecoveryService) RecoverStuckTransactions(ctx context.Context) error {
	logging.Logger.Debug().Msg("Starting stuck transaction recovery")

	// Get stuck transactions (past timeout)
	stuckTransactions, err := s.TransactionRepo.GetStuckTransactions()
	if err != nil {
		return fmt.Errorf("failed to get stuck transactions: %w", err)
	}

	if len(stuckTransactions) > 0 {
		logging.Logger.Info().Int("count", len(stuckTransactions)).Msg("Found stuck transactions")
	}

	for _, transaction := range stuckTransactions {
		if err := s.RecoverSingleTransaction(ctx, transaction); err != nil {
			logging.Logger.Error().
				Err(err).
				Str("transaction_id", transaction.ID).
				Msg("Failed to recover transaction")
		}
	}
	return nil
}

// RecoverSingleTransaction recovers a single stuck transaction
func (s *RecoveryService) RecoverSingleTransaction(ctx context.Context, transaction *entity.Transaction) error {
	logging.Logger.Info().
		Str("transaction_id", transaction.ID).
		Str("type", transaction.Type).
		Str("transaction_status", transaction.TransactionStatus).
		Time("timeout_at", transaction.TimeoutAt).
		Msg("Recovering transaction")

	// Get accounts locked by this transaction
	accounts, err := s.AccountRepo.GetAccountsInTransaction(transaction.ID)
	if err != nil {
		return fmt.Errorf("failed to get locked accounts: %w", err)
	}

	switch {
	case transaction.ShouldTimeout():
		logging.Logger.Warn().
			Str("transaction_id", transaction.ID).
			Msg("Transaction timed out, failing and unlocking accounts")

		return s.handleTimeout(ctx, transaction, accounts)

	case transaction.CanRetry():
		logging.Logger.Info().
			Str("transaction_id", transaction.ID).
			Int("retry_count", transaction.RetryCount).
			Msg("Retrying transaction")

		return s.retryTransaction(ctx, transaction)

	default:
		logging.Logger.Warn().
			Str("transaction_id", transaction.ID).
			Int("retry_count", transaction.RetryCount).
			Msg("Max retries exceeded, failing transaction")

		return s.handleMaxRetriesExceeded(ctx, transaction, accounts)
	}
}

// handleTimeout handles transactions that have timed out
func (s *RecoveryService) handleTimeout(ctx context.Context, transaction *entity.Transaction, accounts []*entity.Account) error {
	// Force unlock accounts first
	if err := s.TransactionRepo.ForceUnlockAccounts(transaction.ID); err != nil {
		return fmt.Errorf("failed to unlock accounts: %w", err)
	}

	// Mark transaction as failed
	transaction.TransactionStatus = entity.TransactionStatusFailed
	transaction.ErrorReason = "transaction timeout during recovery"
	transaction.UpdatedAt = time.Now()

	if err := s.TransactionRepo.UpdateTransaction(transaction); err != nil {
		return fmt.Errorf("failed to update transaction status: %w", err)
	}

	logging.Logger.Info().
		Str("transaction_id", transaction.ID).
		Int("accounts_unlocked", len(accounts)).
		Msg("Successfully handled timed out transaction")

	return nil
}

// retryTransaction attempts to retry a transaction
func (s *RecoveryService) retryTransaction(ctx context.Context, transaction *entity.Transaction) error {
	transaction.MarkForRetry()
	if err := s.TransactionRepo.UpdateTransactionOnRecovery(transaction); err != nil {
		return fmt.Errorf("failed to mark transaction for retry: %w", err)
	}

	// Attempt to commit the transaction
	commitTransactionService := apptx.NewCommitTransaction(s.AccountRepo, s.CustomerRepo, s.TransactionRepo, s.EventRepo)
	_, err := commitTransactionService.Execute(transaction.ID, "system", "system-"+uuid.NewString())
	if err != nil {
		logging.Logger.Error().
			Err(err).
			Str("transaction_id", transaction.ID).
			Msg("Transaction retry failed")

		return fmt.Errorf("transaction retry failed: %w", err)
	}

	logging.Logger.Info().
		Str("transaction_id", transaction.ID).
		Msg("Transaction retry completed successfully")

	return nil
}

// handleMaxRetriesExceeded handles transactions that exceeded max retries
func (s *RecoveryService) handleMaxRetriesExceeded(ctx context.Context, transaction *entity.Transaction, accounts []*entity.Account) error {
	if err := s.TransactionRepo.ForceUnlockAccounts(transaction.ID); err != nil {
		return fmt.Errorf("failed to unlock accounts: %w", err)
	}

	// Mark transaction as failed
	transaction.TransactionStatus = entity.TransactionStatusFailed
	transaction.ErrorReason = fmt.Sprintf("max retries exceeded (%d retries)", transaction.RetryCount)
	transaction.UpdatedAt = time.Now()

	if err := s.TransactionRepo.UpdateTransaction(transaction); err != nil {
		return fmt.Errorf("failed to update transaction status: %w", err)
	}

	logging.Logger.Warn().
		Str("transaction_id", transaction.ID).
		Int("retry_count", transaction.RetryCount).
		Int("accounts_unlocked", len(accounts)).
		Msg("Transaction failed due to max retries exceeded")

	return nil
}

// RecoverAllOnStartup recovers all pending transactions on service startup
func (s *RecoveryService) RecoverAllOnStartup(ctx context.Context) error {
	logging.Logger.Info().Msg("Starting comprehensive transaction recovery on startup")

	// Get all pending and recovering transactions
	transactions, err := s.TransactionRepo.GetPendingTransactions()
	if err != nil {
		return fmt.Errorf("failed to get pending transactions: %w", err)
	}

	logging.Logger.Info().Int("count", len(transactions)).Msg("Found transactions for recovery")

	recoveredCount := 0
	failedCount := 0

	for _, transaction := range transactions {
		if err := s.RecoverSingleTransaction(ctx, transaction); err != nil {
			logging.Logger.Error().
				Err(err).
				Str("transaction_id", transaction.ID).
				Msg("Failed to recover transaction on startup")
			failedCount++
		} else {
			recoveredCount++
		}
	}

	logging.Logger.Info().
		Int("recovered", recoveredCount).
		Int("failed", failedCount).
		Int("total", len(transactions)).
		Msg("Startup transaction recovery completed")

	return nil
}

func (s *RecoveryService) handleLockedInconsistentTransactions() error {
	// Find transactions where accounts are locked but transaction isn't completed
	lockedTransactions, err := s.TransactionRepo.GetLockedButIncompleteTransactions()
	if err != nil {
		return err
	}

	for _, transaction := range lockedTransactions {
		logging.Logger.Warn().
			Str("transaction_id", transaction.ID).
			Msg("Found locked inconsistent transaction - forcing unlock and fail")

		// Force unlock accounts
		if err := s.TransactionRepo.ForceUnlockAccounts(transaction.ID); err != nil {
			logging.Logger.Error().Err(err).Str("transaction_id", transaction.ID).Msg("Failed to unlock accounts")
			continue
		}

		// Mark as failed
		transaction.TransactionStatus = entity.TransactionStatusFailed
		transaction.ErrorReason = "recovery: forced unlock due to inconsistent state"
		if err := s.TransactionRepo.UpdateTransaction(transaction); err != nil {
			logging.Logger.Error().Err(err).Str("transaction_id", transaction.ID).Msg("Failed to update transaction")
		}
	}

	return nil
}
