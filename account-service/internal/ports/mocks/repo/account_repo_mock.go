package repo

import (
	"account-service/internal/domain/entity"
	"github.com/stretchr/testify/mock"
)

// MockAccountRepo implements ports.AccountRepo for testing
type MockAccountRepo struct {
	mock.Mock
}

func (m *MockAccountRepo) CreateAccount(account *entity.Account) error {
	args := m.Called(account)
	return args.Error(0)
}

func (m *MockAccountRepo) GetAccountByID(id string) (*entity.Account, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Account), args.Error(1)
}

func (m *MockAccountRepo) GetAccountByCustomerID(id string) ([]*entity.Account, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Account), args.Error(1)
}

func (m *MockAccountRepo) GetAccountByIDForUpdate(id string) (*entity.Account, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Account), args.Error(1)
}

func (m *MockAccountRepo) GetAccountsByCustomerID(customerID string) ([]*entity.Account, error) {
	args := m.Called(customerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Account), args.Error(1)
}

func (m *MockAccountRepo) UpdateAccount(account *entity.Account) error {
	args := m.Called(account)
	return args.Error(0)
}

func (m *MockAccountRepo) UpdateAccountBalance(id string, newBalance float64, currentVersion int, requester string) error {
	args := m.Called(id, newBalance, currentVersion, requester)
	return args.Error(0)
}

func (m *MockAccountRepo) LockAccountForTransaction(id string, transactionID string) error {
	args := m.Called(id, transactionID)
	return args.Error(0)
}

func (m *MockAccountRepo) UnlockAccountFromTransaction(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockAccountRepo) CheckTransactionLock(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockAccountRepo) ForceUnlockAllAccounts(transactionID string) error {
	args := m.Called(transactionID)
	return args.Error(0)
}

func (m *MockAccountRepo) IncrementVersion(id string, currentVersion int) error {
	args := m.Called(id, currentVersion)
	return args.Error(0)
}

func (m *MockAccountRepo) GetAccountsInTransaction(transactionID string) ([]*entity.Account, error) {
	args := m.Called(transactionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Account), args.Error(1)
}

func (m *MockAccountRepo) GetCustomerAccountsInTransaction(customerID string) ([]*entity.Account, error) {
	args := m.Called(customerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Account), args.Error(1)
}

func (m *MockAccountRepo) GetCustomerAccountsInTransactionOrHasBalance(customerID string) ([]*entity.Account, error) {
	args := m.Called(customerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Account), args.Error(1)
}

func (m *MockAccountRepo) DeleteAccount(id, requester string) error {
	args := m.Called(id, requester)
	return args.Error(0)
}

func (m *MockAccountRepo) GetAccountsByFiltersWithPagination(filters map[string]interface{}, page, pageSize int, setOrder string) ([]*entity.Account, int64, error) {
	args := m.Called(filters, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.Account), args.Get(1).(int64), args.Error(2)
}

func (m *MockAccountRepo) DeleteAllAccountsByCustomerID(customerID, requester string) error {
	args := m.Called(customerID, requester)
	return args.Error(0)
}
