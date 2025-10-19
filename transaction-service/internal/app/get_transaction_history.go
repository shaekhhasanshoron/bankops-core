package app

import (
	"fmt"
	"time"
	"transaction-service/internal/domain/entity"
	custom_err "transaction-service/internal/domain/error"
	"transaction-service/internal/logging"
	"transaction-service/internal/observability/metrics"
	"transaction-service/internal/ports"
)

// GetTransactionHistory is a use-case getting transaction history
type GetTransactionHistory struct {
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
		err = fmt.Errorf("%w: failed to get transaction history", custom_err.ErrDatabase)
		return nil, 0, err
	}

	return transactions, total, nil
}
