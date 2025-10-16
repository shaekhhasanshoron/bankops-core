package grpc_client

import (
	"context"
	protoauth "gateway-service/api/protogen/authservice/proto"
	"github.com/stretchr/testify/mock"
)

type MockAuthClient struct {
	mock.Mock
}

func (m *MockAuthClient) Connect() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockAuthClient) EnsureConnection() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockAuthClient) Close() {
	m.Called()
}

func (m *MockAuthClient) Authenticate(ctx context.Context, req *protoauth.AuthenticateRequest) (*protoauth.AuthenticateResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*protoauth.AuthenticateResponse), args.Error(1)
}

func (m *MockAuthClient) CreateEmployee(ctx context.Context, req *protoauth.CreateEmployeeRequest) (*protoauth.CreateEmployeeResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*protoauth.CreateEmployeeResponse), args.Error(1)
}

func (m *MockAuthClient) DeleteEmployee(ctx context.Context, req *protoauth.DeleteEmployeeRequest) (*protoauth.DeleteEmployeeResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*protoauth.DeleteEmployeeResponse), args.Error(1)
}

func (m *MockAuthClient) UpdateEmployee(ctx context.Context, req *protoauth.UpdateRoleRequest) (*protoauth.UpdateRoleResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*protoauth.UpdateRoleResponse), args.Error(1)
}

func (m *MockAuthClient) ListEmployee(ctx context.Context, req *protoauth.ListEmployeeRequest) (*protoauth.ListEmployeeResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*protoauth.ListEmployeeResponse), args.Error(1)
}

func (m *MockAuthClient) StartConnectionMonitor(ctx context.Context) {
	m.Called(ctx)
}
