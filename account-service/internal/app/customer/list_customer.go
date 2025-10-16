package customer

import (
	"account-service/internal/domain/entity"
	"account-service/internal/domain/value"
	"account-service/internal/logging"
	"account-service/internal/observability/metrics"
	"account-service/internal/ports"
	"fmt"
)

// ListCustomer is a use-case for list of customers
type ListCustomer struct {
	CustomerRepo ports.CustomerRepo
}

// NewListCustomer creates a new ListCustomer use-case
func NewListCustomer(customerRepo ports.CustomerRepo) *ListCustomer {
	return &ListCustomer{
		CustomerRepo: customerRepo,
	}
}

// Execute list customers
func (c *ListCustomer) Execute(page, pageSize int, requestId string) ([]*entity.Customer, int64, int64, string, error) {
	metrics.IncRequestActive()
	defer metrics.DecRequestActive()

	var err error
	defer func() {
		metrics.RecordOperation("list_customers", err)
	}()

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 100
	}

	customers, totalCount, err := c.CustomerRepo.ListCustomer(page, pageSize)
	if err != nil {
		logging.Logger.Error().Err(err).Int("page", page).Int("page_size", pageSize).Msg("Failed to list customers")
		return nil, 0, 0, "Failed to list customer", fmt.Errorf("%w: failed to list customers", value.ErrDatabase)
	}

	totalPages := int64(0)
	if totalCount > 0 {
		totalPages = (totalCount + int64(pageSize) - 1) / int64(pageSize)
	}

	return customers, totalCount, totalPages, "Customer List", nil
}
