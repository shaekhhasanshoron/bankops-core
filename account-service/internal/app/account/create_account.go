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
	"strings"
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

	customerID = strings.TrimSpace(customerID)
	if customerID == "" {
		err = fmt.Errorf("%w: customer ID is required", custom_err.ErrValidationFailed)
		logging.Logger.Error().Err(err).Msg("Required missing fields")
		err = custom_err.ErrValidationFailed
		return nil, fmt.Sprintf("%s: customer ID is required", custom_err.ErrValidationFailed), err
	}
	if initialDeposit < 0 {
		err = custom_err.ErrInvalidAmount
		logging.Logger.Error().Err(err).Msg("Invalid request")
		return nil, err.Error(), err
	}
	if initialDeposit < config.Current().AccountConfig.MinDepositAmount {
		err = fmt.Errorf("%w: minimum deposit amount - %.2f", custom_err.ErrInvalidAmount, config.Current().AccountConfig.MinDepositAmount)
		logging.Logger.Error().Err(err).Msg(fmt.Sprintf("Minimum deposit amount %.2f", config.Current().AccountConfig.MinDepositAmount))
		err = custom_err.ErrInvalidAmount
		return nil, fmt.Sprintf("%s: minimum deposit amount - %.2f", custom_err.ErrInvalidAmount, config.Current().AccountConfig.MinDepositAmount), err
	}

	if requester == "" {
		err = fmt.Errorf("%w: requester not found", custom_err.ErrUnauthorizedRequest)
		logging.Logger.Error().Err(err).Msg("Unknown requester")
		err = custom_err.ErrUnauthorizedRequest
		return nil, fmt.Sprintf("%s: requester not found", custom_err.ErrUnauthorizedRequest), err
	}

	customerExists, err := a.CustomerRepo.Exists(customerID)
	if err != nil {
		err = fmt.Errorf("%w: failed to verify customer", custom_err.ErrDatabase)
		logging.Logger.Error().Err(err).Msg("Failed to verify customer")
		err = custom_err.ErrDatabase
		return nil, "Failed to verify customer", err
	}

	if !customerExists {
		err = fmt.Errorf("%w", custom_err.ErrCustomerNotFound)
		logging.Logger.Error().Err(err).Msg("Customer not found")
		err = custom_err.ErrCustomerNotFound
		return nil, "Customer not found", err
	}

	account, err := entity.NewAccount(customerID, entity.AccountTypeSavings, initialDeposit, requester)
	if err != nil {
		err = fmt.Errorf("%w", custom_err.ErrInvalidAccount)
		logging.Logger.Error().Err(err).Msg("Failed to verify account")
		err = custom_err.ErrInvalidAccount
		return nil, "Failed to verify account", err
	}

	if err = account.Validate(); err != nil {
		err = fmt.Errorf("%w", custom_err.ErrInvalidAccount)
		logging.Logger.Error().Err(err).Msg("Failed to validate account")
		err = custom_err.ErrInvalidAccount
		return nil, "Failed to validate account", err
	}

	account.CreatedBy = requester

	// Create account
	if err = a.AccountRepo.CreateAccount(account); err != nil {
		err = fmt.Errorf("%w: failed to create account", custom_err.ErrDatabase)
		logging.Logger.Error().Err(err).Str("account_id", account.ID).Str("customer_id", customerID).Msg("Failed to create account")
		err = custom_err.ErrDatabase
		return nil, fmt.Sprintf("%s: failed to create account", custom_err.ErrDatabase), err
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
