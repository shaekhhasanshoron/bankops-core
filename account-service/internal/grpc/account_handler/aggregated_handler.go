package handlers

import (
	protoacc "account-service/api/protogen/accountservice/proto"
	appaccount "account-service/internal/app/account"
	appcustomer "account-service/internal/app/customer"
	apptxsaga "account-service/internal/app/transaction_saga"
)

// AccountHandlerService implements the AccountServiceServer interface.
type AccountHandlerService struct {
	protoacc.UnimplementedAccountServiceServer
	CreateCustomerService                *appcustomer.CreateCustomer
	ListCustomerService                  *appcustomer.ListCustomer
	DeleteCustomerService                *appcustomer.DeleteCustomer
	CreateAccountService                 *appaccount.CreateAccount
	DeleteAccountService                 *appaccount.DeleteAccount
	GetAccountBalanceService             *appaccount.GetAccountBalance
	ListAccountService                   *appaccount.ListAccount
	ValidateAccountForTransactionService *apptxsaga.ValidateAccountForTransaction
	LockAccountForTransaction            *apptxsaga.LockAccountForTransaction
	UnlockAccountsForTransaction         *apptxsaga.UnlockAccountsForTransaction
	UpdateAccountBalanceForTransaction   *apptxsaga.UpdateAccountBalanceForTransaction
}

// NewAggregatedHandler creates a new AccountHandler.
func NewAggregatedHandler() *AccountHandlerService {
	return &AccountHandlerService{}
}
