package repo

import (
	"account-service/internal/domain/entity"
	"github.com/stretchr/testify/mock"
)

// MockEventRepo implements ports.EventRepo for testing
type MockEventRepo struct {
	mock.Mock
}

func (m *MockEventRepo) CreateEvent(event *entity.Event) error {
	args := m.Called(event)
	return args.Error(0)
}
