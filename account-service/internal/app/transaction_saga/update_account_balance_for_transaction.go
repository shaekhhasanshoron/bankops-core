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
		err = custom_err.ErrInvalidRequest
		return nil, "at least one account balance update is required", err
	}

	for _, update := range accountBalanceUpdates {
		account, err := t.AccountRepo.GetAccountByID(update.AccountID)
		if err != nil {
			logging.Logger.Error().Err(err).Str("account_id", update.AccountID).Msg("Failed to lock accounts for transaction")
			err = custom_err.ErrDatabase
			return nil, "failed to lock accounts for transaction", err
		}

		if account == nil {
			logging.Logger.Error().Err(err).Str("account_id", update.AccountID).Msg("Account not found")
			err = custom_err.ErrAccountNotFound
			return nil, "Account not found", err
		}
	}

	resp, err := t.AccountRepo.UpdateAccountBalanceLifecycle(accountBalanceUpdates, requester)
	if err != nil {
		logging.Logger.Error().Err(err).Msg("Could not update account balance")
		err = custom_err.ErrDatabase
		return nil, "failed to update account balance", err
	}

	return resp, "account balances updated successfully", nil
}
