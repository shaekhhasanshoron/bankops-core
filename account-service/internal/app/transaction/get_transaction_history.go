package transaction

import (
	"account-service/internal/domain/entity"
	custom_err "account-service/internal/domain/error"
	"account-service/internal/logging"
	"account-service/internal/observability/metrics"
	"account-service/internal/ports"
	"fmt"
	"time"
)

// GetTransactionHistory is a use-case getting transaction history
type GetTransactionHistory struct {
	AccountRepo     ports.AccountRepo
	CustomerRepo    ports.CustomerRepo
	EventRepo       ports.EventRepo
	TransactionRepo ports.TransactionRepo
}

// NewGetTransactionHistory creates a new GetTransactionHistory use-case
func NewGetTransactionHistory(transactionRepo ports.TransactionRepo) *GetTransactionHistory {
	return &GetTransactionHistory{
		TransactionRepo: transactionRepo,
	}
}

func (t *GetTransactionHistory) Execute(accountID string, companyID string, types []string, startDate, endDate *time.Time, sortOrder string, page, pageSize int, requester, requestId string) ([]*entity.Transaction, int64, error) {
	metrics.IncRequestActive()
	defer metrics.DecRequestActive()

	var err error
	defer func() {
		metrics.RecordOperation("get_transaction_history", err)
	}()

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 50
	}

	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	if startDate != nil && endDate != nil && startDate.After(*endDate) {
		err = fmt.Errorf("%w: start date cannot be after end date", custom_err.ErrValidationFailed)
		return nil, 0, err
	}

	transactions, total, err := t.TransactionRepo.GetTransactionHistory(accountID, companyID, startDate, endDate, sortOrder, page, pageSize, types)
	if err != nil {
		logging.Logger.Error().Err(err).
			Str("account_id", accountID).
			Int("page", page).
			Int("page_size", pageSize).
			Msg("Failed to get transaction history")
		return nil, 0, fmt.Errorf("%w: failed to get transaction history", custom_err.ErrDatabase)
	}

	return transactions, total, nil
}
