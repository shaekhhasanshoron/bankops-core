package auth

import (
	"github.com/stretchr/testify/mock"
	"time"
)

type MockTokenSigner struct {
	mock.Mock
}

func (m *MockTokenSigner) SignJWT(username, role, secretKey string, expiryTime time.Duration) (string, error) {
	args := m.Called(username, role, secretKey, expiryTime)
	return args.String(0), args.Error(1)
}

func (m *MockTokenSigner) SignJWTRefreshToken(username, secretKey string) (string, error) {
	args := m.Called(username, secretKey)
	return args.String(0), args.Error(1)
}
