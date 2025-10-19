package grpc_client

import (
	"context"
	prototx "gateway-service/api/protogen/txservice/proto"
	"github.com/stretchr/testify/mock"
)

type MockTransactionClient struct {
	mock.Mock
}

func (m *MockTransactionClient) Connect() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockTransactionClient) EnsureConnection() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockTransactionClient) Close() {
	m.Called()
}

func (m *MockTransactionClient) StartConnectionMonitor(ctx context.Context) {
	m.Called(ctx)
}

func (m *MockTransactionClient) IsHealthy() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockTransactionClient) InitTransaction(ctx context.Context, req *prototx.InitTransactionRequest) (*prototx.InitTransactionResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*prototx.InitTransactionResponse), args.Error(1)
}

func (m *MockTransactionClient) GetTransactionHistory(ctx context.Context, req *prototx.GetTransactionHistoryRequest) (*prototx.GetTransactionHistoryResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*prototx.GetTransactionHistoryResponse), args.Error(1)
}
