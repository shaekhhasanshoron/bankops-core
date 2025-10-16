package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	protoacc "gateway-service/api/protogen/accountservice/proto"
	mock_client "gateway-service/internal/ports/mocks/grpc_client"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupAccountRoutes(handler *AccountHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(func(c *gin.Context) {
		c.Set("username", "test-admin")
		c.Next()
	})

	router.POST("/api/v1/account", handler.CreateAccount)
	router.DELETE("/api/v1/account", handler.DeleteAccount)
	router.GET("/api/v1/account/:id/balance", handler.GetAccountBalance)
	router.GET("/api/v1/account", handler.ListAccounts)

	return router
}

// TestCreateAccount_Success tests account creation success with valid input
func TestCreateAccount_Success(t *testing.T) {
	mockClient := new(mock_client.MockAccountClient)
	accountHandler := &AccountHandler{AccountClient: mockClient}
	router := setupAccountRoutes(accountHandler)

	createReq := CreateAccountRequest{
		CustomerID:    "cust-12345",
		DepositAmount: 1000.0,
	}

	expectedResponse := &protoacc.CreateAccountResponse{
		AccountId: "acc-67890",
		Response: &protoacc.Response{
			Success: true,
			Message: "Account created successfully",
		},
	}

	mockClient.On("CreateAccount", mock.Anything, mock.MatchedBy(func(req *protoacc.CreateAccountRequest) bool {
		return req.CustomerId == "cust-12345" &&
			req.InitialDeposit == 1000.0 &&
			req.Metadata != nil &&
			req.Metadata.Requester == "test-admin" && // Match the username from middleware
			req.Metadata.RequestId == "test-request-id"
	})).Return(expectedResponse, nil)

	body, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/api/v1/account", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token")
	req.Header.Set("X-Request-ID", "test-request-id")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code, "Expected status 201 but got %d. Response: %s", w.Code, w.Body.String())

	var response CreateAccountResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err, "Failed to unmarshal response: %s", w.Body.String())
	assert.Equal(t, "acc-67890", response.AccountID)
	assert.Equal(t, "Account created successfully", response.Message)

	mockClient.AssertExpectations(t)
}

// TestCreateAccount_InvalidJSON test for invalid json  input
func TestCreateAccount_InvalidJSON(t *testing.T) {
	mockClient := new(mock_client.MockAccountClient)
	accountHandler := &AccountHandler{AccountClient: mockClient}
	router := setupAccountRoutes(accountHandler)

	// Send invalid JSON
	body := bytes.NewBufferString("{invalid json")
	req, _ := http.NewRequest("POST", "/api/v1/account", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code, "Expected status 400 but got %d", w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err, "Failed to unmarshal error response: %s", w.Body.String())
	assert.Equal(t, "Invalid request payload", response.Error)
}

// TestCreateAccount_MissingRequiredFields if required fields are missing
func TestCreateAccount_MissingRequiredFields(t *testing.T) {
	mockClient := new(mock_client.MockAccountClient)
	accountHandler := &AccountHandler{AccountClient: mockClient}
	router := setupAccountRoutes(accountHandler)

	createReq := map[string]interface{}{
		"customer_id": "cust-12345",
	}

	body, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/api/v1/account", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code, "Expected status 400 but got %d", w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err, "Failed to unmarshal error response: %s", w.Body.String())
	assert.Equal(t, "Invalid request payload", response.Error)
}

// TestCreateAccount_GrpcError tests  if grpc throws error
func TestCreateAccount_GrpcError(t *testing.T) {
	mockClient := new(mock_client.MockAccountClient)
	accountHandler := &AccountHandler{AccountClient: mockClient}
	router := setupAccountRoutes(accountHandler)

	createReq := CreateAccountRequest{
		CustomerID:    "cust-12345",
		DepositAmount: 1000.0,
	}

	mockClient.On("CreateAccount", mock.Anything, mock.MatchedBy(func(req *protoacc.CreateAccountRequest) bool {
		return req.CustomerId == "cust-12345" &&
			req.InitialDeposit == 1000.0 &&
			req.Metadata != nil &&
			req.Metadata.Requester == "test-admin" &&
			req.Metadata.RequestId == "test-request-id"
	})).Return((*protoacc.CreateAccountResponse)(nil), errors.New("gRPC connection error"))

	body, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/api/v1/account", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token")
	req.Header.Set("X-Request-ID", "test-request-id")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code, "Expected status 401 but got %d. Response: %s", w.Code, w.Body.String())

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err, "Failed to unmarshal error response: %s", w.Body.String())
	assert.Equal(t, "Invalid credentials", response.Error)

	mockClient.AssertExpectations(t)
}

// TestDeleteAccount_SuccessSingleScope test delete account with scope=single and id=account_id
func TestDeleteAccount_SuccessSingleScope(t *testing.T) {
	mockClient := new(mock_client.MockAccountClient)
	accountHandler := &AccountHandler{AccountClient: mockClient}
	router := setupAccountRoutes(accountHandler)

	expectedResponse := &protoacc.DeleteAccountResponse{
		Response: &protoacc.Response{
			Success: true,
			Message: "Account deleted successfully",
		},
	}

	mockClient.On("DeleteAccount", mock.Anything, mock.MatchedBy(func(req *protoacc.DeleteAccountRequest) bool {
		return req.Scope == "single" &&
			req.Id == "acc-12345" &&
			req.Metadata != nil &&
			req.Metadata.Requester == "test-admin" &&
			req.Metadata.RequestId == "test-request-id"
	})).Return(expectedResponse, nil)

	req, _ := http.NewRequest("DELETE", "/api/v1/account?scope=single&id=acc-12345", nil)
	req.Header.Set("Authorization", "Bearer token")
	req.Header.Set("X-Request-ID", "test-request-id")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Expected status 200 but got %d. Response: %s", w.Code, w.Body.String())

	var response DeleteAccountResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err, "Failed to unmarshal response: %s", w.Body.String())
	assert.Equal(t, "Account deleted successfully", response.Message)

	mockClient.AssertExpectations(t)
}

// TestDeleteAccount_EmptyParameters tests if empty params provided
func TestDeleteAccount_EmptyParameters(t *testing.T) {
	mockClient := new(mock_client.MockAccountClient)
	accountHandler := &AccountHandler{AccountClient: mockClient}
	router := setupAccountRoutes(accountHandler)

	req, _ := http.NewRequest("DELETE", "/api/v1/account?scope=&id=", nil)
	req.Header.Set("Authorization", "Bearer token")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code, "Response: %s", w.Code, w.Body.String())

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err, "Failed to unmarshal error response: %s", w.Body.String())
	assert.Contains(t, response.Error, "Missing required parameters")
}

// TestGetAccountBalance_Success tests success of balance get
func TestGetAccountBalance_Success(t *testing.T) {
	mockClient := new(mock_client.MockAccountClient)
	accountHandler := &AccountHandler{AccountClient: mockClient}
	router := setupAccountRoutes(accountHandler)

	expectedResponse := &protoacc.GetBalanceResponse{
		Balance: 1500.75,
		Response: &protoacc.Response{
			Success: true,
			Message: "Balance retrieved successfully",
		},
	}

	mockClient.On("GetBalance", mock.Anything, mock.MatchedBy(func(req *protoacc.GetBalanceRequest) bool {
		return req.AccountId == "acc-12345" &&
			req.Metadata != nil &&
			req.Metadata.Requester == "test-admin" &&
			req.Metadata.RequestId == "test-request-id"
	})).Return(expectedResponse, nil)

	req, _ := http.NewRequest("GET", "/api/v1/account/acc-12345/balance", nil)
	req.Header.Set("Authorization", "Bearer token")
	req.Header.Set("X-Request-ID", "test-request-id")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Expected status 200 but got %d. Response: %s", w.Code, w.Body.String())

	var response GetBalanceResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err, "Failed to unmarshal response: %s", w.Body.String())
	assert.Equal(t, 1500.75, response.Balance)
	assert.Equal(t, "Balance retrieved successfully", response.Message)

	mockClient.AssertExpectations(t)
}

// TestGetAccountBalance_ZeroBalance tests with zero balance
func TestGetAccountBalance_ZeroBalance(t *testing.T) {
	mockClient := new(mock_client.MockAccountClient)
	accountHandler := &AccountHandler{AccountClient: mockClient}
	router := setupAccountRoutes(accountHandler)

	expectedResponse := &protoacc.GetBalanceResponse{
		Balance: 0.0,
		Response: &protoacc.Response{
			Success: true,
			Message: "Balance retrieved successfully",
		},
	}

	mockClient.On("GetBalance", mock.Anything, mock.MatchedBy(func(req *protoacc.GetBalanceRequest) bool {
		return req.AccountId == "acc-zero" &&
			req.Metadata != nil &&
			req.Metadata.Requester == "test-admin" &&
			req.Metadata.RequestId == "test-request-id"
	})).Return(expectedResponse, nil)

	req, _ := http.NewRequest("GET", "/api/v1/account/acc-zero/balance", nil)
	req.Header.Set("Authorization", "Bearer token")
	req.Header.Set("X-Request-ID", "test-request-id")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Expected status 200 but got %d. Response: %s", w.Code, w.Body.String())

	var response GetBalanceResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err, "Failed to unmarshal response: %s", w.Body.String())
	assert.Equal(t, 0.0, response.Balance)
	assert.Equal(t, "Balance retrieved successfully", response.Message)

	mockClient.AssertExpectations(t)
}

// TestGetAccountBalance_AccountNotFound tests if account not found
func TestGetAccountBalance_AccountNotFound(t *testing.T) {
	mockClient := new(mock_client.MockAccountClient)
	accountHandler := &AccountHandler{AccountClient: mockClient}
	router := setupAccountRoutes(accountHandler)

	expectedResponse := &protoacc.GetBalanceResponse{
		Response: &protoacc.Response{
			Success: false,
			Message: "Account not found",
		},
	}

	mockClient.On("GetBalance", mock.Anything, mock.MatchedBy(func(req *protoacc.GetBalanceRequest) bool {
		return req.AccountId == "acc-notfound" &&
			req.Metadata != nil &&
			req.Metadata.Requester == "test-admin" &&
			req.Metadata.RequestId == "test-request-id"
	})).Return(expectedResponse, nil)

	req, _ := http.NewRequest("GET", "/api/v1/account/acc-notfound/balance", nil)
	req.Header.Set("Authorization", "Bearer token")
	req.Header.Set("X-Request-ID", "test-request-id")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code, "Expected status 400 but got %d. Response: %s", w.Code, w.Body.String())

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err, "Failed to unmarshal error response: %s", w.Body.String())
	assert.Equal(t, "Account not found", response.Error)

	mockClient.AssertExpectations(t)
}

// TestGetAccountBalance_GrpcConnectionError test if gRPC request error
func TestGetAccountBalance_GrpcConnectionError(t *testing.T) {
	mockClient := new(mock_client.MockAccountClient)
	accountHandler := &AccountHandler{AccountClient: mockClient}
	router := setupAccountRoutes(accountHandler)

	mockClient.On("GetBalance", mock.Anything, mock.MatchedBy(func(req *protoacc.GetBalanceRequest) bool {
		return req.AccountId == "acc-12345" &&
			req.Metadata != nil &&
			req.Metadata.Requester == "test-admin" &&
			req.Metadata.RequestId == "test-request-id"
	})).Return((*protoacc.GetBalanceResponse)(nil), errors.New("gRPC service unavailable"))

	req, _ := http.NewRequest("GET", "/api/v1/account/acc-12345/balance", nil)
	req.Header.Set("Authorization", "Bearer token")
	req.Header.Set("X-Request-ID", "test-request-id")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code, "Expected status 401 but got %d. Response: %s", w.Code, w.Body.String())

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err, "Failed to unmarshal error response: %s", w.Body.String())
	assert.Equal(t, "gRPC service unavailable", response.Error)

	mockClient.AssertExpectations(t)
}
