package ports

import (
	"context"
	protoauth "gateway-service/api/protogen/authservice/proto"
)

type AuthClient interface {
	Connect() error
	EnsureConnection() error
	Close()
	IsHealthy() bool
	Authenticate(ctx context.Context, req *protoauth.AuthenticateRequest) (*protoauth.AuthenticateResponse, error)
	CreateEmployee(ctx context.Context, req *protoauth.CreateEmployeeRequest) (*protoauth.CreateEmployeeResponse, error)
	DeleteEmployee(ctx context.Context, req *protoauth.DeleteEmployeeRequest) (*protoauth.DeleteEmployeeResponse, error)
	UpdateEmployee(ctx context.Context, req *protoauth.UpdateRoleRequest) (*protoauth.UpdateRoleResponse, error)
	ListEmployee(ctx context.Context, req *protoauth.ListEmployeeRequest) (*protoauth.ListEmployeeResponse, error)
	StartConnectionMonitor(ctx context.Context)
}
