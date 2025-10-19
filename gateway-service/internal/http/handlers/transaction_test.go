package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	prototx "gateway-service/api/protogen/txservice/proto"
	mock_client "gateway-service/internal/ports/mocks/grpc_client"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupTransactionRoutes(handler *TransactionHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(func(c *gin.Context) {
		c.Set("username", "test-admin")
		c.Next()
	})

	router.POST("/api/v1/transaction/init", handler.InitTransaction)
	router.GET("/api/v1/transaction", handler.ListTransactions)

	return router
}

// TestInitTransaction_SuccessTransfer tests successful transaction
func TestInitTransaction_SuccessTransfer(t *testing.T) {
	mockClient := new(mock_client.MockTransactionClient)
	accountHandler := &TransactionHandler{TransactionClient: mockClient}
	router := setupTransactionRoutes(accountHandler)

	request := InitTransactionRequest{
		SourceAccountID:      "acc-12345",
		DestinationAccountID: "acc-67890",
		TransactionType:      "transfer",
		Amount:               1000.0,
		Reference:            "test-transfer",
	}

	expectedResponse := &prototx.InitTransactionResponse{
		TransactionId:     "txn-12345",
		TransactionStatus: "pending",
		Response: &prototx.Response{
			Success: true,
			Message: "Transaction initiated successfully",
		},
	}

	mockClient.On("InitTransaction", mock.Anything, mock.MatchedBy(func(req *prototx.InitTransactionRequest) bool {
		return req.SourceAccountId == "acc-12345" &&
			req.DestinationAccountId == "acc-67890" &&
			req.Amount == 1000.0 &&
			req.Type == "transfer" &&
			req.Reference == "test-transfer" &&
			req.Metadata != nil &&
			req.Metadata.Requester == "test-admin" &&
			req.Metadata.RequestId == "test-request-id"
	})).Return(expectedResponse, nil)

	body, _ := json.Marshal(request)
	req, _ := http.NewRequest("POST", "/api/v1/transaction/init", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token")
	req.Header.Set("X-Request-ID", "test-request-id")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code, "Response: %s", w.Code, w.Body.String())

	var response InitTransactionResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err, "Failed to unmarshal response: %s", w.Body.String())
	assert.Equal(t, "txn-12345", response.TransactionID)
	assert.Equal(t, "pending", response.TransactionStatus)
	assert.Equal(t, "Transaction initiated successfully", response.Message)

	mockClient.AssertExpectations(t)
}

// TestInitTransaction_SuccessWithdrawFull tests for successful withdraw of full amount
func TestInitTransaction_SuccessWithdrawFull(t *testing.T) {
	mockClient := new(mock_client.MockTransactionClient)
	accountHandler := &TransactionHandler{TransactionClient: mockClient}
	router := setupTransactionRoutes(accountHandler)

	request := InitTransactionRequest{
		SourceAccountID: "acc-12345",
		TransactionType: "withdraw_full",
		Amount:          0,
		Reference:       "close-account",
	}

	expectedResponse := &prototx.InitTransactionResponse{
		TransactionId:     "txn-67890",
		TransactionStatus: "completed",
		Response: &prototx.Response{
			Success: true,
			Message: "Full withdrawal completed",
		},
	}

	mockClient.On("InitTransaction", mock.Anything, mock.MatchedBy(func(req *prototx.InitTransactionRequest) bool {
		return req.SourceAccountId == "acc-12345" &&
			req.DestinationAccountId == "" && // Should be empty for withdraw_full
			req.Amount == 0.0 &&
			req.Type == "withdraw_full" &&
			req.Reference == "close-account" &&
			req.Metadata != nil
	})).Return(expectedResponse, nil)

	body, _ := json.Marshal(request)
	req, _ := http.NewRequest("POST", "/api/v1/transaction/init", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token")
	req.Header.Set("X-Request-ID", "test-request-id")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response InitTransactionResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "txn-67890", response.TransactionID)
	assert.Equal(t, "completed", response.TransactionStatus)

	mockClient.AssertExpectations(t)
}

// TestInitTransaction_InvalidJSON tests if invalid json in provided
func TestInitTransaction_InvalidJSON(t *testing.T) {
	mockClient := new(mock_client.MockTransactionClient)
	accountHandler := &TransactionHandler{TransactionClient: mockClient}
	router := setupTransactionRoutes(accountHandler)

	body := bytes.NewBufferString("{invalid json")
	req, _ := http.NewRequest("POST", "/api/v1/transaction/init", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid request payload", response.Error)
}

// TestInitTransaction_MissingRequiredFields tests if there is any missing value
func TestInitTransaction_MissingRequiredFields(t *testing.T) {
	mockClient := new(mock_client.MockTransactionClient)
	accountHandler := &TransactionHandler{TransactionClient: mockClient}
	router := setupTransactionRoutes(accountHandler)

	request := map[string]interface{}{
		"transaction_type": "transfer",
		"amount":           1000.0,
	}

	body, _ := json.Marshal(request)
	req, _ := http.NewRequest("POST", "/api/v1/transaction/init", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid request payload", response.Error)
}

// TestInitTransaction_GrpcConnectionError tests grpc error
func TestInitTransaction_GrpcConnectionError(t *testing.T) {
	mockClient := new(mock_client.MockTransactionClient)
	accountHandler := &TransactionHandler{TransactionClient: mockClient}
	router := setupTransactionRoutes(accountHandler)

	request := InitTransactionRequest{
		SourceAccountID:      "acc-12345",
		DestinationAccountID: "acc-67890",
		TransactionType:      "transfer",
		Amount:               1000.0,
		Reference:            "test-transfer",
	}

	mockClient.On("InitTransaction", mock.Anything, mock.MatchedBy(func(req *prototx.InitTransactionRequest) bool {
		return req.Type == "transfer" && req.SourceAccountId == "acc-12345"
	})).Return((*prototx.InitTransactionResponse)(nil), errors.New("gRPC service unavailable"))

	body, _ := json.Marshal(request)
	req, _ := http.NewRequest("POST", "/api/v1/transaction/init", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token")
	req.Header.Set("X-Request-ID", "test-request-id")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid request", response.Error)

	mockClient.AssertExpectations(t)
}

// TestInitTransaction_InvalidTransactionType if invalid type provided
func TestInitTransaction_InvalidTransactionType(t *testing.T) {
	mockClient := new(mock_client.MockTransactionClient)
	accountHandler := &TransactionHandler{TransactionClient: mockClient}
	router := setupTransactionRoutes(accountHandler)

	request := InitTransactionRequest{
		SourceAccountID: "acc-12345",
		TransactionType: "invalid_type",
		Amount:          1000.0,
		Reference:       "test-invalid",
	}

	expectedResponse := &prototx.InitTransactionResponse{
		Response: &prototx.Response{
			Success: false,
			Message: "Invalid transaction type",
		},
	}

	mockClient.On("InitTransaction", mock.Anything, mock.MatchedBy(func(req *prototx.InitTransactionRequest) bool {
		return req.Type == "invalid_type"
	})).Return(expectedResponse, nil)

	body, _ := json.Marshal(request)
	req, _ := http.NewRequest("POST", "/api/v1/transaction/init", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token")
	req.Header.Set("X-Request-ID", "test-request-id")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid transaction type", response.Error)

	mockClient.AssertExpectations(t)
}

// TestListTransactions_SuccessDefaultPagination tests transaction history list for default pagination
func TestListTransactions_SuccessDefaultPagination(t *testing.T) {
	mockClient := new(mock_client.MockTransactionClient)
	accountHandler := &TransactionHandler{TransactionClient: mockClient}
	router := setupTransactionRoutes(accountHandler)

	expectedResponse := &prototx.GetTransactionHistoryResponse{
		Transactions: []*prototx.Transaction{
			{
				Id:                   "txn-1",
				SourceAccountId:      "acc-123",
				DestinationAccountId: "acc-147",
				Type:                 "transfer",
				Amount:               1000.0,
				TransactionStatus:    "completed",
				Reference:            "test-ref-1",
			},
			{
				Id:                   "txn-2",
				SourceAccountId:      "acc-123",
				DestinationAccountId: "acc-234",
				Type:                 "transfer",
				Amount:               1000.0,
				TransactionStatus:    "completed",
				Reference:            "test-ref-1",
			},
		},
		Pagination: &prototx.PaginationResponse{
			Page:       1,
			PageSize:   50,
			TotalCount: 2,
			TotalPages: 1,
		},
		Response: &prototx.Response{
			Success: true,
			Message: "Transactions retrieved successfully",
		},
	}

	mockClient.On("GetTransactionHistory", mock.Anything, mock.MatchedBy(func(req *prototx.GetTransactionHistoryRequest) bool {
		return req.AccountId == "" &&
			req.CustomerId == "" &&
			req.Types == "" &&
			req.StartDate == nil &&
			req.EndDate == nil &&
			req.SortOrder == "" &&
			req.Pagination.Page == -1 &&
			req.Pagination.PageSize == -1 &&
			req.Metadata != nil &&
			req.Metadata.Requester == "test-admin" &&
			req.Metadata.RequestId == "test-request-id"
	})).Return(expectedResponse, nil)

	req, _ := http.NewRequest("GET", "/api/v1/transaction", nil)
	req.Header.Set("Authorization", "Bearer token")
	req.Header.Set("X-Request-ID", "test-request-id")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Expected status 200 but got %d. Response: %s", w.Code, w.Body.String())

	var response ListTransactionResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err, "Failed to unmarshal response: %s", w.Body.String())
	assert.Equal(t, 1, response.Page)
	assert.Equal(t, 50, response.PageSize)
	assert.Equal(t, 2, response.TotalCount)
	assert.Equal(t, 1, response.TotalPages)
	assert.Equal(t, "Transactions retrieved successfully", response.Message)
	assert.Len(t, response.Transactions, 2)

	mockClient.AssertExpectations(t)
}

// TestListTransactions_SuccessWithPagination tests success with valid pagination
func TestListTransactions_SuccessWithPagination(t *testing.T) {
	mockClient := new(mock_client.MockTransactionClient)
	accountHandler := &TransactionHandler{TransactionClient: mockClient}
	router := setupTransactionRoutes(accountHandler)

	expectedResponse := &prototx.GetTransactionHistoryResponse{
		Transactions: []*prototx.Transaction{
			{
				Id:                   "txn-2",
				SourceAccountId:      "acc-123",
				DestinationAccountId: "acc-234",
				Type:                 "transfer",
				Amount:               1000.0,
				TransactionStatus:    "completed",
				Reference:            "test-ref-1",
			},
		},
		Pagination: &prototx.PaginationResponse{
			Page:       2,
			PageSize:   10,
			TotalCount: 25,
			TotalPages: 3,
		},
		Response: &prototx.Response{
			Success: true,
			Message: "Transactions retrieved successfully",
		},
	}

	mockClient.On("GetTransactionHistory", mock.Anything, mock.MatchedBy(func(req *prototx.GetTransactionHistoryRequest) bool {
		return req.Pagination.Page == 2 &&
			req.Pagination.PageSize == 10 &&
			req.SortOrder == "desc" &&
			req.Metadata != nil
	})).Return(expectedResponse, nil)

	req, _ := http.NewRequest("GET", "/api/v1/transaction?page=2&pagesize=10&order=desc", nil)
	req.Header.Set("Authorization", "Bearer token")
	req.Header.Set("X-Request-ID", "test-request-id")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response ListTransactionResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 2, response.Page)
	assert.Equal(t, 10, response.PageSize)
	assert.Equal(t, 25, response.TotalCount)
	assert.Equal(t, 3, response.TotalPages)

	mockClient.AssertExpectations(t)
}

// TestListTransactions_SuccessWithCustomerFilter tests if customer id is provided for filtering
func TestListTransactions_SuccessWithCustomerFilter(t *testing.T) {
	mockClient := new(mock_client.MockTransactionClient)
	accountHandler := &TransactionHandler{TransactionClient: mockClient}
	router := setupTransactionRoutes(accountHandler)

	expectedResponse := &prototx.GetTransactionHistoryResponse{
		Transactions: []*prototx.Transaction{
			{
				Id:                   "txn-1",
				SourceAccountId:      "acc-123",
				DestinationAccountId: "acc-456",
				Type:                 "transfer",
				Amount:               1000.0,
				TransactionStatus:    "completed",
				Reference:            "test-ref-1",
			},
			{
				Id:                   "txn-2",
				SourceAccountId:      "acc-123",
				DestinationAccountId: "acc-234",
				Type:                 "transfer",
				Amount:               1000.0,
				TransactionStatus:    "completed",
				Reference:            "test-ref-1",
			},
		},
		Pagination: &prototx.PaginationResponse{
			Page:       1,
			PageSize:   50,
			TotalCount: 2,
			TotalPages: 1,
		},
		Response: &prototx.Response{
			Success: true,
			Message: "Customer transactions retrieved",
		},
	}

	mockClient.On("GetTransactionHistory", mock.Anything, mock.MatchedBy(func(req *prototx.GetTransactionHistoryRequest) bool {
		return req.CustomerId == "cust-123" &&
			req.Metadata != nil
	})).Return(expectedResponse, nil)

	req, _ := http.NewRequest("GET", "/api/v1/transaction?customer_id=cust-123", nil)
	req.Header.Set("Authorization", "Bearer token")
	req.Header.Set("X-Request-ID", "test-request-id")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response ListTransactionResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 2, response.TotalCount)
	assert.Equal(t, "Customer transactions retrieved", response.Message)

	mockClient.AssertExpectations(t)
}

// TestListTransactions_SuccessWithAllFilters tests success with all filters
func TestListTransactions_SuccessWithAllFilters(t *testing.T) {
	mockClient := new(mock_client.MockTransactionClient)
	accountHandler := &TransactionHandler{TransactionClient: mockClient}
	router := setupTransactionRoutes(accountHandler)

	expectedResponse := &prototx.GetTransactionHistoryResponse{
		Transactions: []*prototx.Transaction{
			{
				Id:                   "txn-2",
				SourceAccountId:      "acc-123",
				DestinationAccountId: "acc-214",
				Type:                 "transfer",
				Amount:               1000.0,
				TransactionStatus:    "completed",
				Reference:            "test-ref-1",
			},
		},
		Pagination: &prototx.PaginationResponse{
			Page:       1,
			PageSize:   20,
			TotalCount: 1,
			TotalPages: 1,
		},
		Response: &prototx.Response{
			Success: true,
			Message: "Filtered transactions retrieved",
		},
	}

	mockClient.On("GetTransactionHistory", mock.Anything, mock.MatchedBy(func(req *prototx.GetTransactionHistoryRequest) bool {
		return req.AccountId == "acc-123" &&
			req.CustomerId == "cust-123" &&
			req.Types == "transfer" &&
			req.StartDate != nil &&
			req.EndDate != nil &&
			req.Pagination.Page == 1 &&
			req.Pagination.PageSize == 20 &&
			req.SortOrder == "asc" &&
			req.Metadata != nil
	})).Return(expectedResponse, nil)

	req, _ := http.NewRequest("GET", "/api/v1/transaction?account_id=acc-123&customer_id=cust-123&types=transfer&start_date=01-01-2024&end_date=31-01-2024&page=1&pagesize=20&order=asc", nil)
	req.Header.Set("Authorization", "Bearer token")
	req.Header.Set("X-Request-ID", "test-request-id")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response ListTransactionResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 1, response.TotalCount)

	mockClient.AssertExpectations(t)
}

// TestListTransactions_GrpcConnectionError tests if grpc throws error
func TestListTransactions_GrpcConnectionError(t *testing.T) {
	mockClient := new(mock_client.MockTransactionClient)
	accountHandler := &TransactionHandler{TransactionClient: mockClient}
	router := setupTransactionRoutes(accountHandler)

	mockClient.On("GetTransactionHistory", mock.Anything, mock.MatchedBy(func(req *prototx.GetTransactionHistoryRequest) bool {
		return req.Metadata != nil
	})).Return((*prototx.GetTransactionHistoryResponse)(nil), errors.New("gRPC service unavailable"))

	req, _ := http.NewRequest("GET", "/api/v1/transaction", nil)
	req.Header.Set("Authorization", "Bearer token")
	req.Header.Set("X-Request-ID", "test-request-id")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "gRPC service unavailable", response.Error)

	mockClient.AssertExpectations(t)
}
