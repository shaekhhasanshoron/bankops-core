package mocks

import (
	"github.com/stretchr/testify/mock"
	"transaction-service/internal/domain/entity"
)

type MockEventRepo struct {
	mock.Mock
}

func (m *MockEventRepo) CreateEvent(event *entity.Event) error {
	args := m.Called(event)
	return args.Error(0)
}
