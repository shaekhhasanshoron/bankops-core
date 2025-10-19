package ports

import (
	"context"
)

type AccountInfo struct {
	AccountID  string
	CustomerID string
	Balance    float64
	Version    int
}

type AccountBalanceUpdate struct {
	AccountID  string
	NewBalance float64
	Version    int
}

type AccountBalanceUpdateResponse struct {
	AccountID string
	Version   int
}

type AccountClient interface {
	Connect() error
	EnsureConnection() error
	Close()
	StartConnectionMonitor(ctx context.Context)
	ValidateAndGetAccounts(ctx context.Context, accountIDs []string, requester, requestId string) ([]AccountInfo, string, error)
	LockAccounts(ctx context.Context, accountIDs []string, transactionID string, requester, requestId string) (string, error)
	UnlockAccounts(ctx context.Context, transactionID string, requester, requestId string) (string, error)
	UpdateAccountsBalance(ctx context.Context, updates []AccountBalanceUpdate, requester, requestId string) ([]AccountBalanceUpdateResponse, string, error)
	GetBalance(ctx context.Context, accountID string) (float64, int, error)
	IsHealthy() bool
}
