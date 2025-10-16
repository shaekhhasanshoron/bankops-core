package repo

import (
	"account-service/internal/domain/entity"
	"github.com/stretchr/testify/mock"
)

// MockCustomerRepo implements ports.CustomerRepo for testing
type MockCustomerRepo struct {
	mock.Mock
}

func (m *MockCustomerRepo) CreateCustomer(customer *entity.Customer) (*entity.Customer, error) {
	args := m.Called(customer)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Customer), args.Error(1)
}

func (m *MockCustomerRepo) GetCustomerByName(name string) (*entity.Customer, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Customer), args.Error(1)
}

func (m *MockCustomerRepo) GetCustomerByID(id string) (*entity.Customer, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Customer), args.Error(1)
}

func (m *MockCustomerRepo) ListCustomer(page, pageSize int) ([]*entity.Customer, int64, error) {
	args := m.Called(page, pageSize)
	return args.Get(0).([]*entity.Customer), args.Get(1).(int64), args.Error(2)
}

func (m *MockCustomerRepo) DeleteCustomerByID(id, requester string) error {
	args := m.Called(id, requester)
	return args.Error(0)
}

func (m *MockCustomerRepo) CheckModificationAllowed(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockCustomerRepo) Exists(id string) (bool, error) {
	args := m.Called(id)
	return args.Bool(0), args.Error(1)
}
