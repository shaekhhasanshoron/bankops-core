package transaction_saga

import (
	custom_err "account-service/internal/domain/error"
	"account-service/internal/grpc/types"
	"account-service/internal/logging"
	"account-service/internal/observability/metrics"
	"account-service/internal/ports"
)

// UpdateAccountBalanceForTransaction is a use-case for update account balance
type UpdateAccountBalanceForTransaction struct {
	AccountRepo ports.AccountRepo
}

// NewUpdateAccountBalanceForTransaction creates a new UpdateAccountBalanceForTransaction use-case
func NewUpdateAccountBalanceForTransaction(accountRepo ports.AccountRepo) *UpdateAccountBalanceForTransaction {
	return &UpdateAccountBalanceForTransaction{
		AccountRepo: accountRepo,
	}
}

func (t *UpdateAccountBalanceForTransaction) Execute(accountBalanceUpdates []types.AccountBalance, requester string) ([]types.AccountBalanceResponse, string, error) {
	metrics.IncRequestActive()
	defer metrics.DecRequestActive()

	var err error
	defer func() {
		metrics.RecordOperation("update_balance_account_for_transaction", err)
	}()

	if accountBalanceUpdates == nil || len(accountBalanceUpdates) == 0 {
		logging.Logger.Error().Err(custom_err.ErrInvalidRequest).Msg("At least one account balance update is required")
		return nil, "at least one account balance update is required", custom_err.ErrInvalidRequest
	}

	for _, update := range accountBalanceUpdates {
		account, err := t.AccountRepo.GetAccountByID(update.AccountID)
		if err != nil {
			logging.Logger.Error().Err(err).Str("account_id", update.AccountID).Msg("Failed to lock accounts for transaction")
			return nil, "failed to lock accounts for transaction", custom_err.ErrDatabase
		}

		if account == nil {
			logging.Logger.Error().Err(err).Str("account_id", update.AccountID).Msg("Account not found")
			return nil, "Account not found", custom_err.ErrAccountNotFound
		}
	}

	resp, err := t.AccountRepo.UpdateAccountBalanceLifecycle(accountBalanceUpdates, requester)
	if err != nil {
		logging.Logger.Error().Err(err).Msg("Could not update account balance")
		return nil, "failed to update account balance", custom_err.ErrDatabase
	}

	return resp, "account balances updated successfully", nil
}
