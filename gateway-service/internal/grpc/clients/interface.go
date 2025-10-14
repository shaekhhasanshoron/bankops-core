package clients

import (
	"context"
	protoacc "gateway-service/api/protogen/accountservice/proto"
	protoauth "gateway-service/api/protogen/authservice/proto"
)

type AuthClient interface {
	Connect() error
	EnsureConnection() error
	Close()
	Authenticate(ctx context.Context, req *protoauth.AuthenticateRequest) (*protoauth.AuthenticateResponse, error)
	CreateEmployee(ctx context.Context, req *protoauth.CreateEmployeeRequest) (*protoauth.CreateEmployeeResponse, error)
	DeleteEmployee(ctx context.Context, req *protoauth.DeleteEmployeeRequest) (*protoauth.DeleteEmployeeResponse, error)
	UpdateEmployee(ctx context.Context, req *protoauth.UpdateRoleRequest) (*protoauth.UpdateRoleResponse, error)
	StartConnectionMonitor(ctx context.Context)
}

type AccountClient interface {
	Connect() error
	EnsureConnection() error
	Close()
	StartConnectionMonitor(ctx context.Context)
	CreateCustomer(ctx context.Context, req *protoacc.CreateCustomerRequest) (*protoacc.CreateCustomerResponse, error)
	UpdateCustomer(ctx context.Context, req *protoacc.UpdateCustomerRequest) (*protoacc.UpdateCustomerResponse, error)
	DeleteCustomer(ctx context.Context, req *protoacc.DeleteCustomerRequest) (*protoacc.DeleteCustomerResponse, error)
	GetCustomer(ctx context.Context, req *protoacc.GetCustomerRequest) (*protoacc.GetCustomerResponse, error)
	ListCustomer(ctx context.Context, req *protoacc.ListCustomersRequest) (*protoacc.ListCustomersResponse, error)
	CreateAccount(ctx context.Context, req *protoacc.CreateAccountRequest) (*protoacc.CreateAccountResponse, error)
	DeleteAccount(ctx context.Context, req *protoacc.DeleteAccountRequest) (*protoacc.DeleteAccountResponse, error)
	ListAccount(ctx context.Context, req *protoacc.ListAccountsRequest) (*protoacc.ListAccountsResponse, error)
	GetBalance(ctx context.Context, req *protoacc.GetBalanceRequest) (*protoacc.GetBalanceResponse, error)
	InitTransaction(ctx context.Context, req *protoacc.InitTransactionRequest) (*protoacc.InitTransactionResponse, error)
	GetTransactionHistory(ctx context.Context, req *protoacc.GetTransactionHistoryRequest) (*protoacc.GetTransactionHistoryResponse, error)
}
