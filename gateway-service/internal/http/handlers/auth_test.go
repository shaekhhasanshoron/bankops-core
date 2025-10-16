package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	protoauth "gateway-service/api/protogen/authservice/proto"
	mock_client "gateway-service/internal/ports/mocks/grpc_client"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupAuthRoutes(authHandler *AuthHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(func(c *gin.Context) {
		c.Set("username", "test-admin")
		c.Next()
	})

	router.POST("/api/v1/auth/login", authHandler.Login)
	return router
}

// TestLogin_Success tests successful login
func TestLogin_Success(t *testing.T) {
	mockClient := new(mock_client.MockAuthClient)
	authHandler := NewAuthHandler(mockClient)
	router := setupAuthRoutes(authHandler)

	loginReq := LoginRequest{
		Username: "testuser",
		Password: "password123",
	}

	expectedResponse := &protoauth.AuthenticateResponse{
		Token: "jwt-token-here",
	}

	mockClient.On("Authenticate", mock.Anything, &protoauth.AuthenticateRequest{
		Username: "testuser",
		Password: "password123",
	}).Return(expectedResponse, nil)

	body, _ := json.Marshal(loginReq)
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response LoginResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "jwt-token-here", response.AccessToken)

	mockClient.AssertExpectations(t)
}

// TestLogin_InvalidPayload tests login with invalid JSON payload
func TestLogin_InvalidPayload(t *testing.T) {
	mockClient := new(mock_client.MockAuthClient)
	authHandler := NewAuthHandler(mockClient)
	router := setupAuthRoutes(authHandler)

	invalidJSON := `{"username": "test", "password": 123}`

	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString(invalidJSON))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid request payload", response.Error)
}

// TestLogin_InvalidCredentials tests login with invalid credentials
func TestLogin_InvalidCredentials(t *testing.T) {
	mockClient := new(mock_client.MockAuthClient)
	authHandler := NewAuthHandler(mockClient)
	router := setupAuthRoutes(authHandler)

	loginReq := LoginRequest{
		Username: "testuser",
		Password: "wrongpassword",
	}

	mockClient.On("Authenticate", mock.Anything, &protoauth.AuthenticateRequest{
		Username: "testuser",
		Password: "wrongpassword",
	}).Return(nil, errors.New("invalid credentials"))

	body, _ := json.Marshal(loginReq)
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid credentials", response.Error)

	mockClient.AssertExpectations(t)
}
