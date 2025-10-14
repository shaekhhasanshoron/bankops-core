package account

import (
	"account-service/internal/domain/value"
	"account-service/internal/logging"
	"account-service/internal/observability/metrics"
	"account-service/internal/ports"
	"fmt"
)

// GetAccountBalance is a use-case for getting balance of an account
type GetAccountBalance struct {
	AccountRepo  ports.AccountRepo
	CustomerRepo ports.CustomerRepo
	EventRepo    ports.EventRepo
}

// NewGetAccountBalance creates a new GetAccountBalance use-case
func NewGetAccountBalance(accountRepo ports.AccountRepo, customerRepo ports.CustomerRepo, eventRepo ports.EventRepo) *GetAccountBalance {
	return &GetAccountBalance{
		AccountRepo:  accountRepo,
		CustomerRepo: customerRepo,
		EventRepo:    eventRepo,
	}
}

func (a *GetAccountBalance) Execute(id, requester, requestId string) (float64, string, error) {
	metrics.IncRequestActive()
	defer metrics.DecRequestActive()

	var err error
	defer func() {
		metrics.RecordOperation("get_balance", err)
	}()

	if id == "" {
		err = fmt.Errorf("%w: 'id' - account id required in param", value.ErrValidationFailed)
		logging.Logger.Error().Err(err).Msg("Invalid request - 'id' account id missing")
		return 0, "Invalid request - 'id' account id missing", err
	}

	account, err := a.AccountRepo.GetAccountByID(id)
	if err != nil {
		logging.Logger.Error().Err(err).Str("account_id", id).Msg("Failed to verify account")
		return 0, "Failed to verify account", fmt.Errorf("%v: failed to verify account", value.ErrDatabase)
	}

	if account == nil {
		err = fmt.Errorf("%v", value.ErrAccountNotFound)
		logging.Logger.Error().Err(err).Str("account_id", id).Msg("Account not found")
		return 0, "Account not found", err
	}
	return account.Balance, "Account balance", nil
}
