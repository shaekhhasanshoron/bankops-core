package unit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	protoauth "gateway-service/api/protogen/authservice/proto"
	"gateway-service/internal/http/handlers"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
)

type MockAuthClient struct {
	mock.Mock
}

func (m *MockAuthClient) Connect() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockAuthClient) EnsureConnection() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockAuthClient) Authenticate(ctx context.Context, req *protoauth.AuthenticateRequest) (*protoauth.AuthenticateResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*protoauth.AuthenticateResponse), args.Error(1)
}

func (m *MockAuthClient) CreateEmployee(ctx context.Context, req *protoauth.CreateEmployeeRequest) (*protoauth.CreateEmployeeResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*protoauth.CreateEmployeeResponse), args.Error(1)
}

func (m *MockAuthClient) DeleteEmployee(ctx context.Context, req *protoauth.DeleteEmployeeRequest) (*protoauth.DeleteEmployeeResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*protoauth.DeleteEmployeeResponse), args.Error(1)
}

func (m *MockAuthClient) UpdateEmployee(ctx context.Context, req *protoauth.UpdateRoleRequest) (*protoauth.UpdateRoleResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*protoauth.UpdateRoleResponse), args.Error(1)
}

func (m *MockAuthClient) Close() {
	m.Called()
}

func (m *MockAuthClient) StartConnectionMonitor(ctx context.Context) {
	m.Called(ctx)
}

// TestLoginAPI_Success simply test login success result
func TestLoginAPI_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockClient := &MockAuthClient{}
	handler := handlers.NewAuthHandler(mockClient)

	// Test request
	loginRequest := map[string]string{
		"username": "admin",
		"password": "password123",
	}

	expectedResponse := &protoauth.AuthenticateResponse{
		Token:        "access-token-123",
		RefreshToken: "refresh-token-456",
	}

	mockClient.On("Authenticate", mock.Anything, mock.MatchedBy(func(req *protoauth.AuthenticateRequest) bool {
		return req.Username == "admin" && req.Password == "password123"
	})).Return(expectedResponse, nil)

	body, _ := json.Marshal(loginRequest)
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/login", handler.Login)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	fmt.Println(response["access_token"])

	assert.Equal(t, "access-token-123", response["access_token"])

	mockClient.AssertExpectations(t)
}
