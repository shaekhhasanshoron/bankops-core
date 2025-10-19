package mocks

import (
	"github.com/stretchr/testify/mock"
	"transaction-service/internal/domain/entity"
)

// MockSagaRepo implements ports.SagaRepo for testing
type MockSagaRepo struct {
	mock.Mock
}

func (m *MockSagaRepo) CreateSaga(saga *entity.TransactionSaga) error {
	args := m.Called(saga)
	return args.Error(0)
}

func (m *MockSagaRepo) GetSagaByID(id string) (*entity.TransactionSaga, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.TransactionSaga), args.Error(1)
}

func (m *MockSagaRepo) GetSagaByTransactionID(transactionID string) (*entity.TransactionSaga, error) {
	args := m.Called(transactionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.TransactionSaga), args.Error(1)
}

func (m *MockSagaRepo) UpdateSaga(saga *entity.TransactionSaga) error {
	args := m.Called(saga)
	return args.Error(0)
}

func (m *MockSagaRepo) GetStuckSagas() ([]*entity.TransactionSaga, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.TransactionSaga), args.Error(1)
}

func (m *MockSagaRepo) GetSagasForRetry() ([]*entity.TransactionSaga, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.TransactionSaga), args.Error(1)
}

func (m *MockSagaRepo) GetSagasByState(state string) ([]*entity.TransactionSaga, error) {
	args := m.Called(state)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.TransactionSaga), args.Error(1)
}
