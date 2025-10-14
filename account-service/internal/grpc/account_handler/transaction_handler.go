package handlers

import (
	protoacc "account-service/api/protogen/accountservice/proto"
	"account-service/internal/domain/entity"
	"context"
	"google.golang.org/protobuf/types/known/timestamppb"
	"strings"
	"time"
)

// InitTransaction initiates a transaction
func (h *AccountHandlerService) InitTransaction(ctx context.Context, req *protoacc.InitTransactionRequest) (*protoacc.InitTransactionResponse, error) {
	var destinationAccountID *string
	if req.DestinationAccountId != "" {
		destinationAccountID = &req.DestinationAccountId
	}

	transaction, message, err := h.InitTransactionService.Execute(
		req.SourceAccountId,
		destinationAccountID,
		req.Amount,
		req.Type,
		req.Reference,
		req.GetMetadata().GetRequester(),
		req.GetMetadata().GetRequestId(),
	)

	if err != nil {
		return &protoacc.InitTransactionResponse{
			TransactionStatus: entity.TransactionStatusFailed,
			Response: &protoacc.Response{
				Message: message,
				Success: false,
			},
		}, nil
	}

	if req.Type == entity.TransactionTypeWithdrawFull {
		commitMessage, commitErr := h.CommitTransactionService.Execute(transaction.ID, req.GetMetadata().GetRequester(), req.GetMetadata().GetRequestId())
		if commitErr != nil {
			return &protoacc.InitTransactionResponse{
				TransactionStatus: entity.TransactionStatusFailed,
				Response: &protoacc.Response{
					Message: commitMessage,
					Success: false,
				},
			}, nil
		}

		updatedTx, err := h.InitTransactionService.TransactionRepo.GetTransactionByID(transaction.ID)
		if err == nil && updatedTx != nil {
			return &protoacc.InitTransactionResponse{
				TransactionId:     transaction.ID,
				TransactionStatus: entity.TransactionStatusCompleted,
				Response: &protoacc.Response{
					Message: "Transaction completed successfully",
					Success: true,
				},
			}, nil
		}
	} else {
		commitMessage, commitErr := h.CommitTransactionService.Execute(transaction.ID, req.GetMetadata().GetRequester(), req.GetMetadata().GetRequestId())
		if commitErr != nil {
			return &protoacc.InitTransactionResponse{
				TransactionStatus: entity.TransactionStatusFailed,
				Response: &protoacc.Response{
					Message: commitMessage,
					Success: false,
				},
			}, nil
		}
	}

	return &protoacc.InitTransactionResponse{
		TransactionId:     transaction.ID,
		TransactionStatus: entity.TransactionStatusCompleted,
		Response: &protoacc.Response{
			Message: "Transaction completed successfully",
			Success: true,
		},
	}, nil
}

func (h *AccountHandlerService) GetTransactionHistory(ctx context.Context, req *protoacc.GetTransactionHistoryRequest) (*protoacc.GetTransactionHistoryResponse, error) {
	var startDate, endDate *time.Time
	if req.StartDate != nil {
		sd := req.StartDate.AsTime()
		startDate = &sd
	}
	if req.EndDate != nil {
		ed := req.EndDate.AsTime()
		endDate = &ed
	}

	var transactionTypes []string
	if strings.TrimSpace(req.Types) != "" {
		transactionTypes = strings.Split(strings.TrimSpace(req.Types), ",")
	}

	// Get pagination parameters
	page := 1
	pageSize := 50
	if req.Pagination != nil {
		if req.Pagination.Page > 0 {
			page = int(req.Pagination.Page)
		}
		if req.Pagination.PageSize > 0 && req.Pagination.PageSize <= 100 {
			pageSize = int(req.Pagination.PageSize)
		}
	}

	// Get sort order
	sortOrder := "desc"
	if req.SortOrder != "" {
		sortOrder = req.SortOrder
	}

	// Call service
	transactions, total, err := h.GetTransactionHistoryService.Execute(
		req.AccountId,
		req.CompanyId,
		transactionTypes,
		startDate,
		endDate,
		sortOrder,
		page,
		pageSize,
		req.GetMetadata().GetRequester(),
		req.GetMetadata().GetRequestId())

	if err != nil {
		return &protoacc.GetTransactionHistoryResponse{
			Transactions: nil,
			Pagination: &protoacc.PaginationResponse{
				Page:       req.GetPagination().GetPage(),
				PageSize:   req.GetPagination().GetPageSize(),
				TotalCount: 0,
				TotalPages: 0,
			},
			Response: &protoacc.Response{
				Message: "Failed to get transaction history",
				Success: false,
			},
		}, nil
	}

	// Convert to proto response
	protoTransactions := make([]*protoacc.Transaction, len(transactions))
	for i, tx := range transactions {
		protoTx := &protoacc.Transaction{
			Id:                   tx.ID,
			SourceAccountId:      tx.SourceAccountID,
			DestinationAccountId: h.toString(tx.DestinationAccountID),
			Amount:               tx.Amount,
			Type:                 tx.Type,
			TransactionStatus:    tx.TransactionStatus,
			Reference:            tx.ReferenceID,
			CreatedAt:            timestamppb.New(tx.CreatedAt),
			UpdatedAt:            timestamppb.New(tx.UpdatedAt),
			CreatedBy:            tx.CreatedBy,
			ErrorReason:          tx.ErrorReason,
			RetryCount:           int32(tx.RetryCount),
			Version:              int32(tx.Version),
		}

		// Add optional fields if they exist
		if tx.LastRetryAt != nil {
			protoTx.LastRetryAt = timestamppb.New(*tx.LastRetryAt)
		}
		if !tx.TimeoutAt.IsZero() {
			protoTx.TimeoutAt = timestamppb.New(tx.TimeoutAt)
		}

		protoTransactions[i] = protoTx
	}

	// Calculate pagination info
	totalPages := int32(0)
	if total > 0 && pageSize > 0 {
		totalPages = int32((total + int64(pageSize) - 1) / int64(pageSize))
	}

	return &protoacc.GetTransactionHistoryResponse{
		Transactions: protoTransactions,
		Pagination: &protoacc.PaginationResponse{
			Page:       int32(page),
			PageSize:   int32(pageSize),
			TotalCount: int32(total),
			TotalPages: totalPages,
		},
		Response: &protoacc.Response{
			Message: "Transaction history response",
			Success: true,
		},
	}, nil
}

func (h *AccountHandlerService) toString(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}
