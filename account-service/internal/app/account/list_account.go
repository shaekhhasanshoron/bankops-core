package account

import (
	"account-service/internal/domain/entity"
	"account-service/internal/domain/value"
	"account-service/internal/logging"
	"account-service/internal/observability/metrics"
	"account-service/internal/ports"
	"fmt"
	"strings"
)

// ListAccount is a use-case for getting account list
type ListAccount struct {
	AccountRepo  ports.AccountRepo
	CustomerRepo ports.CustomerRepo
	EventRepo    ports.EventRepo
}

// NewListAccount creates a new ListAccount use-case
func NewListAccount(accountRepo ports.AccountRepo, customerRepo ports.CustomerRepo, eventRepo ports.EventRepo) *ListAccount {
	return &ListAccount{
		AccountRepo:  accountRepo,
		CustomerRepo: customerRepo,
		EventRepo:    eventRepo,
	}
}

func (a *ListAccount) Execute(scopes, customerID string, minBalance float64, page, pageSize int, requester, requestId string) ([]*entity.Account, int64, int64, string, error) {
	metrics.IncRequestActive()
	defer metrics.DecRequestActive()

	var err error
	defer func() {
		metrics.RecordOperation("list_accounts", err)
	}()

	scopeArr := strings.Split(scopes, ",")
	if len(scopeArr) == 0 {
		scopeArr = append(scopeArr, "all")
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 100
	}

	var hasCustomer, hasBalance, hasTransaction bool

	for _, scope := range scopeArr {
		switch strings.TrimSpace(scope) {
		case "customer":
			hasCustomer = true
		case "has_balance":
			hasBalance = true
		case "in_transaction":
			hasTransaction = true
		}
	}

	if hasCustomer && customerID == "" {
		err = fmt.Errorf("%w: customer ID is required for customer scope", value.ErrValidationFailed)
		logging.Logger.Error().Err(err).Msg("Invalid request - customer ID is required for customer scope")
		return nil, 0, 0, "Invalid request - customer id missing", err
	}

	if hasBalance && minBalance < 0 {
		err = fmt.Errorf("%w: valid balance amount is required for has_balance scope", value.ErrValidationFailed)
		logging.Logger.Error().Err(err).Str("balance", fmt.Sprintf("%.2f", minBalance)).Msg("Invalid request - invalid balance")
		return nil, 0, 0, "Invalid request - invalid balance", err
	}

	filters := make(map[string]interface{})
	filters["status"] = "valid"

	if hasCustomer {
		filters["customer_id"] = customerID
	}

	if hasBalance {
		filters["min_balance"] = minBalance
	}

	if hasTransaction {
		filters["locked_for_tx"] = true
	}

	accounts, totalCount, err := a.AccountRepo.GetAccountsByFiltersWithPagination(filters, page, pageSize)

	totalPages := int64(0)
	if totalCount > 0 {
		totalPages = (totalCount + int64(pageSize) - 1) / int64(pageSize)
	}

	return accounts, totalCount, totalPages, "Account List", nil
}
