package customer

import (
	"account-service/internal/domain/entity"
	"account-service/internal/domain/value"
	"account-service/internal/logging"
	"account-service/internal/observability/metrics"
	"account-service/internal/ports"
	"errors"
	"fmt"
	"gorm.io/gorm"
)

// DeleteCustomer is a use-case for delete a customers
type DeleteCustomer struct {
	CustomerRepo ports.CustomerRepo
	EventRepo    ports.EventRepo
}

// NewDeleteCustomer creates a new DeleteCustomer use-case
func NewDeleteCustomer(customerRepo ports.CustomerRepo, eventRepo ports.EventRepo) *DeleteCustomer {
	return &DeleteCustomer{
		CustomerRepo: customerRepo,
		EventRepo:    eventRepo,
	}
}

func (c *DeleteCustomer) Execute(id, requester, requestId string) (string, error) {
	metrics.IncRequestActive()
	defer metrics.DecRequestActive()

	var err error
	defer func() {
		metrics.RecordOperation("delete_customer", err)
	}()

	if id == "" {
		err = fmt.Errorf("%w: customer ID is required", value.ErrValidationFailed)
		logging.Logger.Error().Err(err).Msg("missing required value")
		return "Missing required data", err
	}

	if requester == "" {
		err = fmt.Errorf("%w: requester is required", value.ErrValidationFailed)
		logging.Logger.Error().Err(err).Msg("Unknown requester")
		return "Unknown requester", err
	}

	customer, err := c.CustomerRepo.GetCustomerByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = value.ErrCustomerNotFound
			return "Customer not found", err
		}
		logging.Logger.Error().Err(err).Str("customer_id", id).Msg("Failed to get customer")
		return "Failed to get customer", err
	}

	if customer == nil {
		err = value.ErrCustomerNotFound
		return "Customer not found", err
	}

	if err = c.CustomerRepo.CheckModificationAllowed(id); err != nil {
		logging.Logger.Warn().Err(err).Msg("Customer deletion not allowed")
		return "Customer deletion not allowed right now", err
	}

	for _, account := range customer.Accounts {
		if account.ActiveStatus == entity.AccountActiveStatusActive && account.Balance > 0 {
			err = fmt.Errorf("cannot delete customer with active accounts having balance: account %s has %.2f", account.ID, account.Balance)
			logging.Logger.Warn().Err(err).Str("customer_id", id).Msg("Customer deletion blocked")
			return "Customer deletion blocked", err
		}
	}

	// Delete customer
	if err = c.CustomerRepo.DeleteCustomerByID(id, requester); err != nil {
		err = fmt.Errorf("%w: failed to delete customer", value.ErrDatabase)
		logging.Logger.Error().Err(err).Str("customer_id", id).Msg("Failed to  deletion customer")
		return "Customer deletion failed", err
	}

	eventData := map[string]interface{}{
		"customer_id":   id,
		"deleted_by":    requester,
		"account_count": len(customer.Accounts),
		"request_id":    requestId,
	}

	event, eventErr := entity.NewEvent(entity.EventTypeCustomerDeleted, id, "customer", requester, eventData)
	if eventErr == nil {
		if createErr := c.EventRepo.CreateEvent(event); createErr != nil {
			logging.Logger.Error().Err(createErr).Str("customer_id", customer.ID).Msg("Failed to create customer deletion event")
		}
	}

	logging.Logger.Debug().Str("customer_id", customer.ID).Str("requester", requester).Msg("Customer deleted successfully")
	return "Customer delete successfully", nil
}
