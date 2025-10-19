package ports

import (
	"context"
	prototx "gateway-service/api/protogen/txservice/proto"
)

type TransactionClient interface {
	Connect() error
	EnsureConnection() error
	Close()
	StartConnectionMonitor(ctx context.Context)
	IsHealthy() bool
	InitTransaction(ctx context.Context, req *prototx.InitTransactionRequest) (*prototx.InitTransactionResponse, error)
	GetTransactionHistory(ctx context.Context, req *prototx.GetTransactionHistoryRequest) (*prototx.GetTransactionHistoryResponse, error)
}
