package handlers

import (
	prototx "transaction-service/api/protogen/txservice/proto"
	apptx "transaction-service/internal/app"
)

type TransactionHandlerService struct {
	prototx.UnimplementedTransactionServiceServer
	InitTransactionService       *apptx.InitTransaction
	GetTransactionHistoryService *apptx.GetTransactionHistory
}

// NewAggregatedHandler creates a new AccountHandler.
func NewAggregatedHandler() *TransactionHandlerService {
	return &TransactionHandlerService{}
}
