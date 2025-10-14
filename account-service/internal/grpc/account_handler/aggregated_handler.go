package handlers

import (
	protoacc "account-service/api/protogen/accountservice/proto"
	appaccount "account-service/internal/app/account"
	appcustomer "account-service/internal/app/customer"
	apptx "account-service/internal/app/transaction"
)

// AccountHandlerService implements the AccountServiceServer interface.
type AccountHandlerService struct {
	protoacc.UnimplementedAccountServiceServer
	CreateCustomerService        *appcustomer.CreateCustomer
	ListCustomerService          *appcustomer.ListCustomer
	DeleteCustomerService        *appcustomer.DeleteCustomer
	CreateAccountService         *appaccount.CreateAccount
	DeleteAccountService         *appaccount.DeleteAccount
	GetAccountBalanceService     *appaccount.GetAccountBalance
	ListAccountService           *appaccount.ListAccount
	InitTransactionService       *apptx.IniTransaction
	CommitTransactionService     *apptx.CommitTransaction
	GetTransactionHistoryService *apptx.GetTransactionHistory
}

// NewAggregatedHandler creates a new AccountHandler.
func NewAggregatedHandler() *AccountHandlerService {
	return &AccountHandlerService{}
}
