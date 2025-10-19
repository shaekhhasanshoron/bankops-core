package repo

import (
	"github.com/stretchr/testify/mock"
	"time"
	"transaction-service/internal/domain/entity"
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

func (m *MockTransactionRepo) UpdateTransactionStatus(id string, transactionStatus string, errorReason string) error {
	args := m.Called(id, transactionStatus, errorReason)
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
	return args.Get(0).([]*entity.Transaction), args.Error(1)
}

func (m *MockTransactionRepo) GetStuckTransactions() ([]*entity.Transaction, error) {
	args := m.Called()
	return args.Get(0).([]*entity.Transaction), args.Error(1)
}
