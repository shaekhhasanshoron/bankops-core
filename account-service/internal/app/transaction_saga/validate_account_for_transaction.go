package transaction_saga

import (
	"account-service/internal/domain/entity"
	custom_err "account-service/internal/domain/error"
	"account-service/internal/logging"
	"account-service/internal/observability/metrics"
	"account-service/internal/ports"
	"fmt"
)

// ValidateAccountForTransaction is a use-case for validating accounts before transaction
type ValidateAccountForTransaction struct {
	AccountRepo ports.AccountRepo
}

// NewValidateAccountForTransaction creates a new ValidateAccountForTransaction use-case
func NewValidateAccountForTransaction(accountRepo ports.AccountRepo) *ValidateAccountForTransaction {
	return &ValidateAccountForTransaction{
		AccountRepo: accountRepo,
	}
}

func (t *ValidateAccountForTransaction) Execute(transactionId string, accountsIds []string, requester, requestId string) ([]*entity.Account, string, error) {
	metrics.IncRequestActive()
	defer metrics.DecRequestActive()

	var err error
	defer func() {
		metrics.RecordOperation("validate_account_for_transaction", err)
	}()

	if len(accountsIds) == 0 {
		logging.Logger.Error().Err(custom_err.ErrMinimumOneAccountIdRequired).Msg("At least one account ID is required")
		return nil, "at least one account ID is required", custom_err.ErrMinimumOneAccountIdRequired
	}

	var accounts []*entity.Account
	for _, accountID := range accountsIds {
		account, err := t.AccountRepo.GetAccountByID(accountID)
		if err != nil {
			logging.Logger.Error().Err(custom_err.ErrDatabase).Str("account_id", accountID).Msg("Failed to validate account")
			return nil, fmt.Sprintf("Failed to validate account'%s'", accountID), custom_err.ErrDatabase
		}

		if account == nil {
			logging.Logger.Error().Err(custom_err.ErrAccountNotFound).Str("account_id", accountID).Msg("Account not found")
			return nil, fmt.Sprintf("Account '%s' not found", accountID), custom_err.ErrAccountNotFound
		}

		if !account.CanTransact() {
			logging.Logger.Error().Err(custom_err.ErrAccountLocked).Str("account_id", accountID).Msg("Account cannot transact")
			return nil, fmt.Sprintf("Account '%s' cannot transact", accountID), custom_err.ErrAccountLocked
		}
		accounts = append(accounts, account)
	}

	return accounts, "All accounts are valid for transaction", nil
}
