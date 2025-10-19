package ports

import (
	"context"
	protoacc "gateway-service/api/protogen/accountservice/proto"
)

type AccountClient interface {
	Connect() error
	EnsureConnection() error
	Close()
	IsHealthy() bool
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
}
