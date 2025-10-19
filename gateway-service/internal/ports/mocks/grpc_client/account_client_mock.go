package grpc_client

import (
	"context"
	protoacc "gateway-service/api/protogen/accountservice/proto"
	"github.com/stretchr/testify/mock"
)

// MockAccountClient is a mock implementation of AccountClient
type MockAccountClient struct {
	mock.Mock
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

func (m *MockAccountClient) IsHealthy() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockAccountClient) StartConnectionMonitor(ctx context.Context) {
	m.Called(ctx)
}

func (m *MockAccountClient) CreateCustomer(ctx context.Context, req *protoacc.CreateCustomerRequest) (*protoacc.CreateCustomerResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*protoacc.CreateCustomerResponse), args.Error(1)
}

func (m *MockAccountClient) UpdateCustomer(ctx context.Context, req *protoacc.UpdateCustomerRequest) (*protoacc.UpdateCustomerResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*protoacc.UpdateCustomerResponse), args.Error(1)
}

func (m *MockAccountClient) DeleteCustomer(ctx context.Context, req *protoacc.DeleteCustomerRequest) (*protoacc.DeleteCustomerResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*protoacc.DeleteCustomerResponse), args.Error(1)
}

func (m *MockAccountClient) GetCustomer(ctx context.Context, req *protoacc.GetCustomerRequest) (*protoacc.GetCustomerResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*protoacc.GetCustomerResponse), args.Error(1)
}

func (m *MockAccountClient) ListCustomer(ctx context.Context, req *protoacc.ListCustomersRequest) (*protoacc.ListCustomersResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*protoacc.ListCustomersResponse), args.Error(1)
}

func (m *MockAccountClient) CreateAccount(ctx context.Context, req *protoacc.CreateAccountRequest) (*protoacc.CreateAccountResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*protoacc.CreateAccountResponse), args.Error(1)
}

func (m *MockAccountClient) DeleteAccount(ctx context.Context, req *protoacc.DeleteAccountRequest) (*protoacc.DeleteAccountResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*protoacc.DeleteAccountResponse), args.Error(1)
}

func (m *MockAccountClient) ListAccount(ctx context.Context, req *protoacc.ListAccountsRequest) (*protoacc.ListAccountsResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*protoacc.ListAccountsResponse), args.Error(1)
}

func (m *MockAccountClient) GetBalance(ctx context.Context, req *protoacc.GetBalanceRequest) (*protoacc.GetBalanceResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*protoacc.GetBalanceResponse), args.Error(1)
}
