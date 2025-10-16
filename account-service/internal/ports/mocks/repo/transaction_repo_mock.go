package repo

import (
	"account-service/internal/domain/entity"
	"github.com/stretchr/testify/mock"
	"time"
)

type MockTransactionRepo struct {
	mock.Mock
}

func (m *MockTransactionRepo) CreateTransaction(transaction *entity.Transaction) error {
	args := m.Called(transaction)
	return args.Error(0)
}

func (m *MockTransactionRepo) GetTransactionByID(id string) (*entity.Transaction, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Transaction), args.Error(1)
}

func (m *MockTransactionRepo) GetTransactionByReferenceID(referenceID string) (*entity.Transaction, error) {
	args := m.Called(referenceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Transaction), args.Error(1)
}

func (m *MockTransactionRepo) BeginTransactionLifecycle(transactionID string, accountIDs []string) error {
	args := m.Called(transactionID, accountIDs)
	return args.Error(0)
}

func (m *MockTransactionRepo) UpdateTransactionStatus(id string, transactionStatus string, errorReason string) error {
	args := m.Called(id, transactionStatus, errorReason)
	return args.Error(0)
}

func (m *MockTransactionRepo) CompleteTransactionLifecycle(transactionID string) error {
	args := m.Called(transactionID)
	return args.Error(0)
}

func (m *MockTransactionRepo) UpdateTransaction(transaction *entity.Transaction) error {
	args := m.Called(transaction)
	return args.Error(0)
}

func (m *MockTransactionRepo) GetTransactionHistory(accountID string, customerID string, startDate, endDate *time.Time, sortOrder string, page, pageSize int, types []string) ([]*entity.Transaction, int64, error) {
	args := m.Called(accountID, customerID, startDate, endDate, sortOrder, page, pageSize, types)
	return args.Get(0).([]*entity.Transaction), args.Get(1).(int64), args.Error(2)
}

func (m *MockTransactionRepo) GetPendingTransactions() ([]*entity.Transaction, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Transaction), args.Error(1)
}

func (m *MockTransactionRepo) GetStuckTransactions() ([]*entity.Transaction, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Transaction), args.Error(1)
}

func (m *MockTransactionRepo) GetLockedButIncompleteTransactions() ([]*entity.Transaction, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Transaction), args.Error(1)
}

func (m *MockTransactionRepo) ForceUnlockAccounts(transactionID string) error {
	args := m.Called(transactionID)
	return args.Error(0)
}

func (m *MockTransactionRepo) UpdateTransactionOnRecovery(transaction *entity.Transaction) error {
	args := m.Called(transaction)
	return args.Error(0)
}

func (m *MockTransactionRepo) GetTransactionWithAccounts(transactionID string) (*entity.Transaction, []*entity.Account, error) {
	args := m.Called(transactionID)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	if args.Get(1) == nil {
		return args.Get(0).(*entity.Transaction), nil, args.Error(2)
	}
	return args.Get(0).(*entity.Transaction), args.Get(1).([]*entity.Account), args.Error(2)
}
