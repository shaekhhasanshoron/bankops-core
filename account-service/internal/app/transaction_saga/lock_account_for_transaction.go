package transaction_saga

import (
	custom_err "account-service/internal/domain/error"
	"account-service/internal/logging"
	"account-service/internal/observability/metrics"
	"account-service/internal/ports"
	"strings"
)

// LockAccountForTransaction is a use-case for locking accounts in transaction
type LockAccountForTransaction struct {
	AccountRepo ports.AccountRepo
}

// NewLockAccountForTransaction creates a new LockAccountForTransaction use-case
func NewLockAccountForTransaction(accountRepo ports.AccountRepo) *LockAccountForTransaction {
	return &LockAccountForTransaction{
		AccountRepo: accountRepo,
	}
}

func (t *LockAccountForTransaction) Execute(transactionId string, accountsIds []string, requester, requestId string) (string, error) {
	metrics.IncRequestActive()
	defer metrics.DecRequestActive()

	var err error
	defer func() {
		metrics.RecordOperation("lock_account_for_transaction", err)
	}()

	if len(accountsIds) == 0 {
		logging.Logger.Error().Err(custom_err.ErrMinimumOneAccountIdRequired).Msg("At least one account ID is required")
		return "at least one account ID is required", custom_err.ErrMinimumOneAccountIdRequired
	}

	if strings.TrimSpace(transactionId) == "" {
		logging.Logger.Error().Err(custom_err.ErrTransactionIdRequired).Msg("Transaction id is required")
		return "transaction id is required", custom_err.ErrTransactionIdRequired
	}

	err = t.AccountRepo.LockAccountsForTransaction(transactionId, accountsIds)
	if err != nil {
		logging.Logger.Error().Err(err).Str("transaction_id", transactionId).Msg("Failed to lock accounts for transaction")
		return "failed to lock accounts for transaction", custom_err.ErrDatabase
	}

	return "Accounts locked successfully", nil
}
