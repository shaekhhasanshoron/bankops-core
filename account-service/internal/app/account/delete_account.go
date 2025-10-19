package account

import (
	"account-service/internal/domain/entity"
	custom_err "account-service/internal/domain/error"
	"account-service/internal/logging"
	"account-service/internal/messaging"
	"account-service/internal/observability/metrics"
	"account-service/internal/ports"
	"fmt"
	"strings"
)

// DeleteAccount is a use-case for deleting an account or accounts of a customer
type DeleteAccount struct {
	AccountRepo  ports.AccountRepo
	CustomerRepo ports.CustomerRepo
	EventRepo    ports.EventRepo
}

// NewDeleteAccount creates a new DeleteAccount use-case
func NewDeleteAccount(accountRepo ports.AccountRepo, customerRepo ports.CustomerRepo, eventRepo ports.EventRepo) *DeleteAccount {
	return &DeleteAccount{
		AccountRepo:  accountRepo,
		CustomerRepo: customerRepo,
		EventRepo:    eventRepo,
	}
}

// Execute delete a account  or delete all accounts for a customer
func (a *DeleteAccount) Execute(scope, id, requester, requestId string) (string, error) {
	metrics.IncRequestActive()
	defer metrics.DecRequestActive()

	var err error
	defer func() {
		metrics.RecordOperation("delete_account", err)
	}()

	scope = strings.TrimSpace(scope)
	id = strings.TrimSpace(id)

	if scope == "" || (scope != "single" && scope != "all") {
		err = fmt.Errorf("%w: scope required or invalid", custom_err.ErrValidationFailed)
		logging.Logger.Error().Err(err).Msg("Invalid request - 'scope' missing or invalid")
		return "Invalid request - 'scope' missing or invalid", err
	}

	if id == "" {
		if scope == "all" {
			err = fmt.Errorf("%w: 'id' - customer id required in param", custom_err.ErrValidationFailed)
			logging.Logger.Error().Err(err).Msg("Invalid request - 'id' customer id missing")
			return "Invalid request - 'id' customer id missing", err
		} else {
			err = fmt.Errorf("%w: 'id' - account id required in param", custom_err.ErrValidationFailed)
			logging.Logger.Error().Err(err).Msg("Invalid request - 'id' account id missing")
			return "Invalid request - 'id' account id missing", err
		}
	}

	if requester == "" {
		err = fmt.Errorf("%w: requester is required", custom_err.ErrValidationFailed)
		logging.Logger.Error().Err(err).Msg("Unknown requester")
		return "Unknown requester", err
	}

	var accountIDs []string
	var customerId string

	if scope == "all" {
		customerExists, err := a.CustomerRepo.Exists(id)
		if err != nil {
			err = fmt.Errorf("%w: failed to verify customer", custom_err.ErrDatabase)
			logging.Logger.Error().Err(err).Str("customer_id", id).Msg("Failed to verify customer")
			return "Failed to verify customer", err
		}

		if !customerExists {
			err = fmt.Errorf("%w", custom_err.ErrCustomerNotFound)
			logging.Logger.Error().Err(err).Str("customer_id", id).Msg("Customer not found")
			return "Customer not found", err
		}

		accounts, err := a.AccountRepo.GetCustomerAccountsInTransactionOrHasBalance(id)
		if err != nil {
			err = fmt.Errorf("%w: failed to verify accounts", custom_err.ErrDatabase)
			logging.Logger.Error().Err(err).Str("customer_id", id).Msg("Failed to verify accounts")
			return "Failed to verify accounts", err
		}

		if accounts != nil && len(accounts) > 0 {
			err = fmt.Errorf("%w: either accounts are in transaction or has balance", custom_err.ErrAccountLocked)
			logging.Logger.Error().Err(err).Str("customer_id", id).Msg("Account deletion blocked - either accounts are in transaction or has balance")
			return "Account deletion blocked. Some accounts are in transaction or has balance", err
		}

		if err = a.AccountRepo.DeleteAllAccountsByCustomerID(id, requester); err != nil {
			err = fmt.Errorf("%w: failed to delete accounts", custom_err.ErrDatabase)
			logging.Logger.Error().Err(err).Str("customer_id", id).Msg("Failed to delete accounts")
			return "Failed to delete accounts", err
		}

		for _, account := range accounts {
			accountIDs = append(accountIDs, account.ID)
		}
		customerId = id
	} else {
		account, err := a.AccountRepo.GetAccountByID(id)
		if err != nil {
			logging.Logger.Error().Err(err).Str("account_id", id).Msg("Failed to verify account")
			return "Failed to verify account", fmt.Errorf("%v: failed to verify account", custom_err.ErrDatabase)
		}

		if account == nil {
			err = fmt.Errorf("%v", custom_err.ErrAccountNotFound)
			logging.Logger.Error().Err(err).Str("account_id", id).Msg("Account not found")
			return "Account not found", err
		}

		if account.Balance > 0 {
			err = fmt.Errorf("cannot delete account with positive balance: %.2f", account.Balance)
			logging.Logger.Error().Err(err).Str("account_id", id).Msg("Account deletion blocked")
			return "Account deletion blocked. Account has balance", err
		}

		if err = a.AccountRepo.CheckTransactionLock(id); err != nil {
			logging.Logger.Error().Err(err).Str("account_id", id).Msg("Failed to verify account")
			return "Failed to verify accounts", err
		}

		if err = a.AccountRepo.DeleteAccount(id, requester); err != nil {
			logging.Logger.Error().Err(err).Str("account_id", id).Msg("Failed to delete account")
			err = fmt.Errorf("%w: failed to delete account", custom_err.ErrDatabase)
			return "Failed to delete account", err
		}

		accountIDs = append(accountIDs, id)
		customerId = account.CustomerID
	}

	accountIdsStr := strings.Join(accountIDs, ",")
	eventData := map[string]interface{}{
		"account_ids": accountIdsStr,
		"customer_id": customerId,
		"deleted_by":  requester,
		"request_id":  requestId,
	}

	event, eventErr := entity.NewEvent(entity.EventTypeAccountDeleted, accountIdsStr, entity.EventAggregateTypeAccount, requester, eventData)
	if eventErr == nil {
		if createErr := a.EventRepo.CreateEvent(event); createErr != nil {
			logging.Logger.Error().Err(createErr).Str("account_ids", accountIdsStr).Str("customer_id", customerId).Msg("Failed to create account delete event")
		}
	}

	_ = messaging.GetService().PublishToDefaultTopic(messaging.Message{Content: accountIdsStr, Status: true, Type: messaging.MessageTypeDeleteAccount})
	return "Accounts deleted successfully", nil
}
