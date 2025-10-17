package account

import (
	"account-service/internal/domain/entity"
	custom_err "account-service/internal/domain/error"
	"account-service/internal/logging"
	"account-service/internal/observability/metrics"
	"account-service/internal/ports"
	"fmt"
	"strconv"
	"strings"
)

// ListAccount is a use-case for getting account list
type ListAccount struct {
	AccountRepo ports.AccountRepo
}

// NewListAccount creates a new ListAccount use-case
func NewListAccount(accountRepo ports.AccountRepo) *ListAccount {
	return &ListAccount{
		AccountRepo: accountRepo,
	}
}

func (a *ListAccount) Execute(customerID, minBalance, inTransaction string, page, pageSize int, setOrder, requester, requestId string) ([]*entity.Account, int64, int64, string, error) {
	metrics.IncRequestActive()
	defer metrics.DecRequestActive()

	var err error
	defer func() {
		metrics.RecordOperation("list_accounts", err)
	}()

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 100
	}

	var minBalanceAmount *float64
	if strings.TrimSpace(minBalance) != "" {
		amount, err := strconv.ParseFloat(strings.TrimSpace(minBalance), 64)
		if err != nil {
			err = fmt.Errorf("%w: customer ID is required for customer scope", custom_err.ErrValidationFailed)
			logging.Logger.Error().Err(err).Str("min_balance", minBalance).Msg("Invalid request - invalid balance id")
			return nil, 0, 0, "Invalid request - customer id missing", custom_err.ErrValidationFailed
		}

		if amount < 0 {
			err = fmt.Errorf("%w: valid balance amount is required for has_balance scope", custom_err.ErrValidationFailed)
			logging.Logger.Error().Err(err).Str("balance", minBalance).Msg("Invalid request - invalid balance")
			return nil, 0, 0, "Invalid request - invalid balance", err
		}
		minBalanceAmount = &amount
	}

	filters := make(map[string]interface{})
	filters["status"] = "valid"

	if strings.TrimSpace(customerID) != "" {
		filters["customer_id"] = strings.TrimSpace(customerID)
	}

	if minBalanceAmount != nil {
		filters["min_balance"] = *minBalanceAmount
	}

	if strings.TrimSpace(inTransaction) == "true" {
		filters["locked_for_tx"] = true
	} else if strings.TrimSpace(inTransaction) == "false" {
		filters["locked_for_tx"] = false
	}

	accounts, totalCount, err := a.AccountRepo.GetAccountsByFiltersWithPagination(filters, page, pageSize, setOrder)

	totalPages := int64(0)
	if totalCount > 0 {
		totalPages = (totalCount + int64(pageSize) - 1) / int64(pageSize)
	}

	if accounts == nil {
		accounts = []*entity.Account{}
	}

	return accounts, totalCount, totalPages, "Account List", nil
}
