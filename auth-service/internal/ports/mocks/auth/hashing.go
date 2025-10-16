package auth

import (
	"github.com/stretchr/testify/mock"
)

type MockHashing struct {
	mock.Mock
}

func (m *MockHashing) HashData(data string) (string, error) {
	args := m.Called(data)
	return args.String(0), args.Error(1)
}
