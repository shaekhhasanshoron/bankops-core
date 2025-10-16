package account

import (
	custom_err "account-service/internal/domain/error"
	"account-service/internal/logging"
	"account-service/internal/observability/metrics"
	"account-service/internal/ports"
	"fmt"
	"strings"
)

// GetAccountBalance is a use-case for getting balance of an account
type GetAccountBalance struct {
	AccountRepo ports.AccountRepo
}

// NewGetAccountBalance creates a new GetAccountBalance use-case
func NewGetAccountBalance(accountRepo ports.AccountRepo) *GetAccountBalance {
	return &GetAccountBalance{
		AccountRepo: accountRepo,
	}
}

func (a *GetAccountBalance) Execute(id, requester, requestId string) (float64, string, error) {
	metrics.IncRequestActive()
	defer metrics.DecRequestActive()

	var err error
	defer func() {
		metrics.RecordOperation("get_balance", err)
	}()

	if strings.TrimSpace(id) == "" {
		err = fmt.Errorf("%w: 'id' - account id required in param", custom_err.ErrValidationFailed)
		logging.Logger.Error().Err(err).Msg("Invalid request - 'id' account id missing")
		return 0, "Invalid request - 'id' account id missing", err
	}

	account, err := a.AccountRepo.GetAccountByID(id)
	if err != nil {
		logging.Logger.Error().Err(err).Str("account_id", id).Msg("Failed to verify account")
		return 0, "Failed to verify account", fmt.Errorf("%v: failed to verify account", custom_err.ErrDatabase)
	}

	if account == nil {
		err = fmt.Errorf("%v", custom_err.ErrAccountNotFound)
		logging.Logger.Error().Err(err).Str("account_id", id).Msg("Account not found")
		return 0, "Account not found", err
	}
	return account.Balance, "Account balance", nil
}
