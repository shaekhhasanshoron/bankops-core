package handlers

import (
	"context"
	"google.golang.org/protobuf/types/known/timestamppb"
	"strings"
	"time"
	prototx "transaction-service/api/protogen/txservice/proto"
	"transaction-service/internal/domain/entity"
)

func (s *TransactionHandlerService) InitTransaction(ctx context.Context, req *prototx.InitTransactionRequest) (*prototx.InitTransactionResponse, error) {
	var destId *string
	if req.DestinationAccountId != "" {
		destId = &req.DestinationAccountId
	}
	// Execute the transaction
	transaction, message, err := s.InitTransactionService.Execute(
		ctx,
		req.SourceAccountId,
		destId,
		req.Amount,
		req.Type,
		req.Reference,
		req.Metadata.Requester,
		req.Metadata.RequestId,
	)

	if err != nil {
		return &prototx.InitTransactionResponse{
			TransactionId:     "",
			TransactionStatus: entity.TransactionStatusFailed,
			Response: &prototx.Response{
				Message: message,
				Success: false,
			},
		}, nil
	}

	return &prototx.InitTransactionResponse{
		TransactionId:     transaction.ID,
		TransactionStatus: transaction.TransactionStatus,
		Response: &prototx.Response{
			Message: message,
			Success: true,
		},
	}, nil
}

func (h *TransactionHandlerService) GetTransactionHistory(ctx context.Context, req *prototx.GetTransactionHistoryRequest) (*prototx.GetTransactionHistoryResponse, error) {
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
		req.CustomerId,
		transactionTypes,
		startDate,
		endDate,
		sortOrder,
		page,
		pageSize,
		req.GetMetadata().GetRequester(),
		req.GetMetadata().GetRequestId())

	if err != nil {
		return &prototx.GetTransactionHistoryResponse{
			Transactions: nil,
			Pagination: &prototx.PaginationResponse{
				Page:       req.GetPagination().GetPage(),
				PageSize:   req.GetPagination().GetPageSize(),
				TotalCount: 0,
				TotalPages: 0,
			},
			Response: &prototx.Response{
				Message: "Failed to get transaction history",
				Success: false,
			},
		}, nil
	}

	// Convert to proto response
	protoTransactions := make([]*prototx.Transaction, len(transactions))
	for i, tx := range transactions {
		protoTx := &prototx.Transaction{
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

	return &prototx.GetTransactionHistoryResponse{
		Transactions: protoTransactions,
		Pagination: &prototx.PaginationResponse{
			Page:       int32(page),
			PageSize:   int32(pageSize),
			TotalCount: int32(total),
			TotalPages: totalPages,
		},
		Response: &prototx.Response{
			Message: "Transaction history response",
			Success: true,
		},
	}, nil
}

func (h *TransactionHandlerService) toString(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}
