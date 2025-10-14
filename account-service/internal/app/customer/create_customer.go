package customer

import (
	"account-service/internal/domain/entity"
	"account-service/internal/domain/value"
	"account-service/internal/logging"
	"account-service/internal/observability/metrics"
	"account-service/internal/ports"
	"fmt"
)

// CreateCustomer is a use-case for creating a new customer
type CreateCustomer struct {
	CustomerRepo ports.CustomerRepo
	EventRepo    ports.EventRepo
}

// NewCreateCustomer creates a new CreateCustomer use-case
func NewCreateCustomer(customerRepo ports.CustomerRepo, eventRepo ports.EventRepo) *CreateCustomer {
	return &CreateCustomer{
		CustomerRepo: customerRepo,
		EventRepo:    eventRepo,
	}
}

// Execute creates a new customer if they don't already exist
func (c *CreateCustomer) Execute(name, requester, requestId string) (*entity.Customer, string, error) {
	metrics.IncRequestActive()
	defer metrics.DecRequestActive()

	var err error
	defer func() {
		metrics.RecordOperation("create_customer", err)
	}()

	if name == "" {
		err = fmt.Errorf("%w: customer name is required", value.ErrValidationFailed)
		logging.Logger.Error().Err(err).Msg("Required missing fields")
		return nil, "Required missing fields", err
	}

	if requester == "" {
		err = fmt.Errorf("%w: requester is required", value.ErrValidationFailed)
		logging.Logger.Error().Err(err).Msg("Unknown requester")
		return nil, "Unknown requester", err
	}

	existingCustomer, err := c.CustomerRepo.GetCustomerByName(name)
	if err == nil && existingCustomer != nil {
		err = fmt.Errorf("%w", value.ErrCustomerExists)
		logging.Logger.Error().Err(err).Msg("Customer already exists")
		return nil, "Customer already exists", err
	}

	customerDTO := ports.CustomerDTO{
		CustomerName: name,
		Requester:    requester,
	}

	customer, err := c.CustomerRepo.CreateCustomer(&customerDTO)
	if err != nil {
		err = fmt.Errorf("%w", value.ErrDatabase)
		logging.Logger.Error().Err(err).Msg("Failed to create customer")
		return nil, "Failed to create customer", err
	}

	eventData := map[string]interface{}{
		"customer_id": customer.ID,
		"name":        customer.Name,
		"created_by":  requester,
		"request_id":  requestId,
	}

	event, eventErr := entity.NewEvent(entity.EventTypeCustomerCreated, customer.ID, entity.EventAggregateType, requester, eventData)
	if eventErr == nil {
		if createErr := c.EventRepo.CreateEvent(event); createErr != nil {
			logging.Logger.Error().Err(createErr).Str("customer_id", customer.ID).Msg("Failed to create customer create event")
		}
	}
	logging.Logger.Debug().Str("customer_id", customer.ID).Msg("Customer created successfully")
	return customer, "Customer created successfully", nil
}
