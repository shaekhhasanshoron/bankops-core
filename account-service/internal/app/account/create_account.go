package account

import (
	"account-service/internal/config"
	"account-service/internal/domain/entity"
	custom_err "account-service/internal/domain/error"
	"account-service/internal/logging"
	"account-service/internal/messaging"
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
		err = fmt.Errorf("%w: customer ID is required", custom_err.ErrValidationFailed)
		logging.Logger.Error().Err(err).Msg("Required missing fields")
		return nil, err.Error(), custom_err.ErrValidationFailed
	}
	if initialDeposit < 0 {
		err = fmt.Errorf("%w", custom_err.ErrInvalidAmount)
		logging.Logger.Error().Err(err).Msg("Invalid request")
		return nil, err.Error(), custom_err.ErrInvalidAmount
	}
	if initialDeposit < config.Current().AccountConfig.MinDepositAmount {
		err = fmt.Errorf("%w: minimum deposit amount - %.2f", custom_err.ErrInvalidAmount, config.Current().AccountConfig.MinDepositAmount)
		logging.Logger.Error().Err(err).Msg(fmt.Sprintf("Minimum deposit amount %.2f", config.Current().AccountConfig.MinDepositAmount))
		return nil, err.Error(), custom_err.ErrInvalidAmount
	}

	if requester == "" {
		err = fmt.Errorf("%w: requester not found", custom_err.ErrUnauthorizedRequest)
		logging.Logger.Error().Err(err).Msg("Unknown requester")
		return nil, err.Error(), custom_err.ErrUnauthorizedRequest
	}

	customerExists, err := a.CustomerRepo.Exists(customerID)
	if err != nil {
		err = fmt.Errorf("%w: failed to verify customer", custom_err.ErrDatabase)
		logging.Logger.Error().Err(err).Msg("Failed to verify customer")
		return nil, "Failed to verify customer", custom_err.ErrDatabase
	}

	if !customerExists {
		err = fmt.Errorf("%w", custom_err.ErrCustomerNotFound)
		logging.Logger.Error().Err(err).Msg("Customer not found")
		return nil, "Customer not found", custom_err.ErrCustomerNotFound
	}

	account, err := entity.NewAccount(customerID, entity.AccountTypeSavings, initialDeposit, requester)
	if err != nil {
		err = fmt.Errorf("%w", custom_err.ErrInvalidAccount)
		logging.Logger.Error().Err(err).Msg("Failed to verify account")
		return nil, "Failed to verify account", custom_err.ErrInvalidAccount
	}

	if err = account.Validate(); err != nil {
		err = fmt.Errorf("%w", custom_err.ErrInvalidAccount)
		logging.Logger.Error().Err(err).Msg("Failed to validate account")
		return nil, "Failed to validate account", custom_err.ErrInvalidAccount
	}

	account.CreatedBy = requester

	// Create account
	if err = a.AccountRepo.CreateAccount(account); err != nil {
		err = fmt.Errorf("%w: failed to create account", custom_err.ErrDatabase)
		logging.Logger.Error().Err(err).Str("account_id", account.ID).Str("customer_id", customerID).Msg("Failed to create account")
		return nil, err.Error(), custom_err.ErrDatabase
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
	_ = messaging.GetService().PublishToDefaultTopic(messaging.Message{Content: account.ToString(), Status: true, Type: messaging.MessageTypeCreateAccount})
	return account, "Account successfully created", nil
}
