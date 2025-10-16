package repo

import (
	"auth-service/internal/domain/entity"
	"github.com/stretchr/testify/mock"
)

type MockEmployeeRepo struct {
	mock.Mock
}

func (m *MockEmployeeRepo) CreateEmployee(employee *entity.Employee) (*entity.Employee, error) {
	args := m.Called(employee)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Employee), args.Error(1)
}

func (m *MockEmployeeRepo) GetEmployeeByUsername(username string) (*entity.Employee, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Employee), args.Error(1)
}

func (m *MockEmployeeRepo) UpdateEmployee(employee *entity.Employee) (*entity.Employee, error) {
	args := m.Called(employee)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Employee), args.Error(1)
}

func (m *MockEmployeeRepo) DeleteEmployee(username, requester string) error {
	args := m.Called(username, requester)
	return args.Error(0)
}

func (m *MockEmployeeRepo) ListEmployee(page, pageSize int, sortOrder string) ([]*entity.Employee, int64, error) {
	args := m.Called(page, pageSize, sortOrder)
	if args.Get(0) == nil {
		return []*entity.Employee{}, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entity.Employee), args.Get(1).(int64), args.Error(2)
}
