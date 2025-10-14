package handlers

import (
	protoacc "account-service/api/protogen/accountservice/proto"
	appcustomer "account-service/internal/app/customer"
)

// AccountHandlerService implements the AccountServiceServer interface.
type AccountHandlerService struct {
	protoacc.UnimplementedAccountServiceServer
	CreateCustomerService *appcustomer.CreateCustomer
	ListCustomerService   *appcustomer.ListCustomer
	DeleteCustomerService *appcustomer.DeleteCustomer
}

// NewAggregatedHandler creates a new AccountHandler.
func NewAggregatedHandler() *AccountHandlerService {
	return &AccountHandlerService{}
}
