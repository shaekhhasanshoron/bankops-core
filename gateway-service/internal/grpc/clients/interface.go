package clients

import (
	"context"
	"gateway-service/api/proto"
)

type AuthClient interface {
	Connect() error
	EnsureConnection() error
	Close()
	Authenticate(ctx context.Context, req *proto.AuthenticateRequest) (*proto.AuthenticateResponse, error)
	CreateEmployee(ctx context.Context, req *proto.CreateEmployeeRequest) (*proto.CreateEmployeeResponse, error)
	DeleteEmployee(ctx context.Context, req *proto.DeleteEmployeeRequest) (*proto.DeleteEmployeeResponse, error)
	UpdateEmployee(ctx context.Context, req *proto.UpdateRoleRequest) (*proto.UpdateRoleResponse, error)
	StartConnectionMonitor(ctx context.Context)
}

//// Ensure GRPCAuthClient implements AuthClient interface
//var _ AuthClient = (*GRPCAuthClient)(nil)
