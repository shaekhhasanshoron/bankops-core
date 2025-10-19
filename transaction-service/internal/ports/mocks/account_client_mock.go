package mocks

import (
	"context"
	"github.com/stretchr/testify/mock"
	"transaction-service/internal/ports"
)

// MockAccountClient implements ports.AccountClient for testing
type MockAccountClient struct {
	mock.Mock
}

func (m *MockAccountClient) ValidateAndGetAccounts(ctx context.Context, accountIDs []string, requester, requestId string) ([]ports.AccountInfo, string, error) {
	args := m.Called(ctx, accountIDs, requester, requestId)

	// Handle the case where no accounts are returned
	if args.Get(0) == nil {
		return []ports.AccountInfo{}, args.String(1), args.Error(2)
	}
	return args.Get(0).([]ports.AccountInfo), args.String(1), args.Error(2)
}

func (m *MockAccountClient) LockAccounts(ctx context.Context, accountIDs []string, transactionID string, requester, requestId string) (string, error) {
	args := m.Called(ctx, accountIDs, transactionID, requester, requestId)
	return args.String(0), args.Error(1)
}

func (m *MockAccountClient) UnlockAccounts(ctx context.Context, transactionID string, requester, requestId string) (string, error) {
	args := m.Called(ctx, transactionID, requester, requestId)
	return args.String(0), args.Error(1)
}

func (m *MockAccountClient) UpdateAccountsBalance(ctx context.Context, updates []ports.AccountBalanceUpdate, requester, requestId string) ([]ports.AccountBalanceUpdateResponse, string, error) {
	args := m.Called(ctx, updates, requester, requestId)
	if args.Get(0) == nil {
		return []ports.AccountBalanceUpdateResponse{}, args.String(1), args.Error(2)
	}
	return args.Get(0).([]ports.AccountBalanceUpdateResponse), args.String(1), args.Error(2)
}

func (m *MockAccountClient) GetBalance(ctx context.Context, accountID string) (float64, int, error) {
	args := m.Called(ctx, accountID)
	return args.Get(0).(float64), args.Get(1).(int), args.Error(2)
}

func (m *MockAccountClient) Connect() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockAccountClient) EnsureConnection() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockAccountClient) Close() {
	m.Called()
}

func (m *MockAccountClient) StartConnectionMonitor(ctx context.Context) {
	m.Called(ctx)
}

func (m *MockAccountClient) IsHealthy() bool {
	args := m.Called()
	return args.Bool(0)
}
