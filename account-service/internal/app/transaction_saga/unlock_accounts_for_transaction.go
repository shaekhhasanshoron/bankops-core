package transaction_saga

import (
	custom_err "account-service/internal/domain/error"
	"account-service/internal/logging"
	"account-service/internal/observability/metrics"
	"account-service/internal/ports"
	"strings"
)

// UnlockAccountsForTransaction is a use-case for unlocking accounts after a transaction
type UnlockAccountsForTransaction struct {
	AccountRepo ports.AccountRepo
}

// NewUnlockAccountsForTransaction creates a new UnlockAccountsForTransaction use-case
func NewUnlockAccountsForTransaction(accountRepo ports.AccountRepo) *UnlockAccountsForTransaction {
	return &UnlockAccountsForTransaction{
		AccountRepo: accountRepo,
	}
}

func (t *UnlockAccountsForTransaction) Execute(transactionId string, requester, requestId string) (string, error) {
	metrics.IncRequestActive()
	defer metrics.DecRequestActive()

	var err error
	defer func() {
		metrics.RecordOperation("update_account_for_transaction", err)
	}()

	if strings.TrimSpace(transactionId) == "" {
		logging.Logger.Error().Err(custom_err.ErrTransactionIdRequired).Msg("Transaction id is required")
		return "transaction id is required", custom_err.ErrTransactionIdRequired
	}

	err = t.AccountRepo.UnlockAccountsForTransaction(transactionId)
	if err != nil {
		logging.Logger.Error().Err(err).Str("transaction_id", transactionId).Msg("Failed to unlock accounts")
		return "failed to unlock accounts for transaction", custom_err.ErrDatabase
	}

	return "Accounts unlocked successfully", nil
}
