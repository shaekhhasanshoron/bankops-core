package account

import (
	"account-service/internal/config"
	"account-service/internal/domain/entity"
	"account-service/internal/domain/value"
	"account-service/internal/logging"
	"account-service/internal/observability/metrics"
	"account-service/internal/ports"
	"fmt"
)

// CreateAccount is a use-case for creating a new account for customer
type CreateAccount struct {
	AccountRepo  ports.AccountRepo
	CustomerRepo ports.CustomerRepo
	EventRepo    ports.EventRepo
}

// NewCreateAccount creates a new CreateAccount use-case
func NewCreateAccount(accountRepo ports.AccountRepo, customerRepo ports.CustomerRepo, eventRepo ports.EventRepo) *CreateAccount {
	return &CreateAccount{
		AccountRepo:  accountRepo,
		CustomerRepo: customerRepo,
		EventRepo:    eventRepo,
	}
}

// Execute creates a new account for customer
func (a *CreateAccount) Execute(customerID string, initialDeposit float64, requester, requestId string) (*entity.Account, string, error) {
	metrics.IncRequestActive()
	defer metrics.DecRequestActive()

	var err error
	defer func() {
		metrics.RecordOperation("create_account", err)
	}()

	if customerID == "" {
		err = fmt.Errorf("%w: customer ID is required", value.ErrValidationFailed)
		logging.Logger.Error().Err(err).Msg("Required missing fields")
		return nil, "Required missing fields", err
	}
	if initialDeposit < 0 {
		err = fmt.Errorf("%w", value.ErrInvalidAmount)
		logging.Logger.Error().Err(err).Msg("Invalid request")
		return nil, "Invalid request", err
	}
	if initialDeposit < config.Current().AccountConfig.MinDepositAmount {
		err = fmt.Errorf("%w: minimum deposit amount - %.2f", value.ErrInvalidAmount, config.Current().AccountConfig.MinDepositAmount)
		logging.Logger.Error().Err(err).Msg(fmt.Sprintf("Minimum deposit amount %.2f", config.Current().AccountConfig.MinDepositAmount))
		return nil, fmt.Sprintf("Minimum deposit amount %.2f", config.Current().AccountConfig.MinDepositAmount), err
	}

	if requester == "" {
		err = fmt.Errorf("%w: requester is required", value.ErrValidationFailed)
		logging.Logger.Error().Err(err).Msg("Unknown requester")
		return nil, "Unknown requester", err
	}

	customerExists, err := a.CustomerRepo.Exists(customerID)
	if err != nil {
		err = fmt.Errorf("%w: failed to verify customer", value.ErrDatabase)
		logging.Logger.Error().Err(err).Msg("Failed to verify customer")
		return nil, "Failed to verify customer", err
	}

	if !customerExists {
		err = fmt.Errorf("%w", value.ErrCustomerNotFound)
		logging.Logger.Error().Err(err).Msg("Customer not found")
		return nil, "Customer not found", err
	}

	account, err := entity.NewAccount(customerID, entity.AccountTypeSavings, initialDeposit, requester)
	if err != nil {
		err = fmt.Errorf("%w", value.ErrInvalidAccount)
		logging.Logger.Error().Err(err).Msg("Failed to verify account")
		return nil, "Failed to verify account", err
	}

	if err = account.Validate(); err != nil {
		err = fmt.Errorf("%w", value.ErrInvalidAccount)
		logging.Logger.Error().Err(err).Msg("Failed to validate account")
		return nil, "Failed to validate account", err
	}

	account.CreatedBy = requester

	// Create account
	if err = a.AccountRepo.CreateAccount(account); err != nil {
		logging.Logger.Error().Err(err).Str("account_id", account.ID).Str("customer_id", customerID).Msg("Failed to create account")
		return nil, "Failed to create account", fmt.Errorf("%w: failed to create account", value.ErrDatabase)
	}

	eventData := map[string]interface{}{
		"account_id":      account.ID,
		"customer_id":     customerID,
		"initial_deposit": initialDeposit,
		"created_by":      requester,
		"request_id":      requestId,
	}

	event, eventErr := entity.NewEvent(entity.EventTypeAccountCreated, account.ID, entity.EventAggregateTypeAccount, requester, eventData)
	if eventErr == nil {
		if createErr := a.EventRepo.CreateEvent(event); createErr != nil {
			logging.Logger.Error().Err(createErr).Str("account_id", account.ID).Str("customer_id", customerID).Msg("Failed to create account create event")
		}
	}

	return account, "Account successfully created", nil
}
