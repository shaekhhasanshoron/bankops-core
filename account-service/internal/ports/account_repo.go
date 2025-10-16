package ports

import "account-service/internal/domain/entity"

type AccountRepo interface {
	CreateAccount(account *entity.Account) error
	GetAccountByID(id string) (*entity.Account, error)
	GetAccountByIDForUpdate(id string) (*entity.Account, error)
	GetAccountByCustomerID(customerID string) ([]*entity.Account, error)
	UpdateAccount(account *entity.Account) error
	UpdateAccountBalance(id string, newBalance float64, currentVersion int, requester string) error
	LockAccountForTransaction(id string, transactionID string) error
	UnlockAccountFromTransaction(id string) error
	CheckTransactionLock(id string) error
	ForceUnlockAllAccounts(transactionID string) error
	IncrementVersion(id string, currentVersion int) error
	GetAccountsInTransaction(transactionID string) ([]*entity.Account, error)
	GetCustomerAccountsInTransaction(customerID string) ([]*entity.Account, error)
	GetCustomerAccountsInTransactionOrHasBalance(customerID string) ([]*entity.Account, error)
	DeleteAccount(id, requester string) error
	GetAccountsByFiltersWithPagination(filters map[string]interface{}, page, pageSize int, setOrder string) ([]*entity.Account, int64, error)
	DeleteAllAccountsByCustomerID(customerID, requester string) error
}
