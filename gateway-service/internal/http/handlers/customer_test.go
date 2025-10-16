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

func setupCustomerRoutes(accountHandler *AccountHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(func(c *gin.Context) {
		c.Set("username", "test-admin")
		c.Next()
	})

	router.POST("/api/v1/customer", accountHandler.CreateCustomer)
	router.DELETE("/api/v1/customer/:id", accountHandler.DeleteCustomer)
	router.GET("/api/v1/customer", accountHandler.ListCustomer)
	router.GET("/api/v1/customer/:id/account", accountHandler.ListCustomerAccounts)

	return router
}

// TestCreateCustomer_Success tests successful customer creation
func TestCreateCustomer_Success(t *testing.T) {
	mockClient := new(mock_client.MockAccountClient)
	accountHandler := NewAccountHandler(mockClient)
	router := setupCustomerRoutes(accountHandler)

	createReq := CreateCustomerRequest{
		Name: "Customer One",
	}

	expectedResponse := &protoacc.CreateCustomerResponse{
		CustomerId: "cust-12345",
		Response: &protoacc.Response{
			Success: true,
			Message: "Customer created successfully",
		},
	}

	// Use mock.MatchedBy to match the request with any RequestId
	mockClient.On("CreateCustomer", mock.Anything, mock.MatchedBy(func(req *protoacc.CreateCustomerRequest) bool {
		return req.Name == "Customer One" &&
			req.Metadata != nil &&
			req.Metadata.Requester == "test-admin" &&
			req.Metadata.RequestId == "test-request-id"
	})).Return(expectedResponse, nil)

	body, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/api/v1/customer", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-ID", "test-request-id")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response CreateCustomerResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "cust-12345", response.CustomerID)
	assert.Equal(t, "Customer created successfully", response.Message)

	mockClient.AssertExpectations(t)
}

// TestCreateCustomer_InvalidPayload tests customer creation with invalid payload
func TestCreateCustomer_InvalidPayload(t *testing.T) {
	mockClient := new(mock_client.MockAccountClient)
	accountHandler := NewAccountHandler(mockClient)
	router := setupCustomerRoutes(accountHandler)

	invalidJSON := `{"name": 123}`

	req, _ := http.NewRequest("POST", "/api/v1/customer", bytes.NewBufferString(invalidJSON))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid request payload", response.Error)
}

// TestCreateCustomer_MissingName tests customer creation with missing name
func TestCreateCustomer_MissingName(t *testing.T) {
	mockClient := new(mock_client.MockAccountClient)
	accountHandler := NewAccountHandler(mockClient)
	router := setupCustomerRoutes(accountHandler)

	createReq := map[string]interface{}{} // empty request

	body, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/api/v1/customer", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid request payload", response.Error)
}

// TestCreateCustomer_GRPCFailure tests customer creation when gRPC call fails
func TestCreateCustomer_GRPCFailure(t *testing.T) {
	mockClient := new(mock_client.MockAccountClient)
	accountHandler := NewAccountHandler(mockClient)
	router := setupCustomerRoutes(accountHandler)

	createReq := CreateCustomerRequest{
		Name: "Customer One",
	}

	mockClient.On("CreateCustomer", mock.Anything, &protoacc.CreateCustomerRequest{
		Name: "Customer One",
		Metadata: &protoacc.Metadata{
			RequestId: "",
			Requester: "test-admin",
		},
	}).Return(nil, errors.New("gRPC connection failed"))

	body, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/api/v1/customer", bytes.NewBuffer(body))
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

// TestCreateCustomer_NilResponse tests customer creation with nil gRPC response
func TestCreateCustomer_NilResponse(t *testing.T) {
	mockClient := new(mock_client.MockAccountClient)
	accountHandler := NewAccountHandler(mockClient)
	router := setupCustomerRoutes(accountHandler)

	createReq := CreateCustomerRequest{
		Name: "Customer One",
	}

	mockClient.On("CreateCustomer", mock.Anything, &protoacc.CreateCustomerRequest{
		Name: "Customer One",
		Metadata: &protoacc.Metadata{
			RequestId: "",
			Requester: "test-admin",
		},
	}).Return(nil, nil)

	body, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/api/v1/customer", bytes.NewBuffer(body))
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

// TestCreateCustomer_UnsuccessfulResponse tests customer creation with unsuccessful gRPC response
func TestCreateCustomer_UnsuccessfulResponse(t *testing.T) {
	mockClient := new(mock_client.MockAccountClient)
	accountHandler := NewAccountHandler(mockClient)
	router := setupCustomerRoutes(accountHandler)

	createReq := CreateCustomerRequest{
		Name: "Customer One",
	}

	expectedResponse := &protoacc.CreateCustomerResponse{
		CustomerId: "",
		Response: &protoacc.Response{
			Success: false,
			Message: "Customer name already exists",
		},
	}

	mockClient.On("CreateCustomer", mock.Anything, &protoacc.CreateCustomerRequest{
		Name: "Customer One",
		Metadata: &protoacc.Metadata{
			RequestId: "",
			Requester: "test-admin",
		},
	}).Return(expectedResponse, nil)

	body, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/api/v1/customer", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Customer name already exists", response.Error)

	mockClient.AssertExpectations(t)
}

// TestDeleteCustomer_Success tests successful customer deletion
func TestDeleteCustomer_Success(t *testing.T) {
	mockClient := new(mock_client.MockAccountClient)
	accountHandler := NewAccountHandler(mockClient)
	router := setupCustomerRoutes(accountHandler)

	expectedResponse := &protoacc.DeleteCustomerResponse{
		Response: &protoacc.Response{
			Success: true,
			Message: "Customer deleted successfully",
		},
	}

	mockClient.On("DeleteCustomer", mock.Anything, mock.MatchedBy(func(req *protoacc.DeleteCustomerRequest) bool {
		return req.CustomerId == "cust-12345" &&
			req.Metadata != nil &&
			req.Metadata.Requester == "test-admin" &&
			req.Metadata.RequestId == "test-request-id"
	})).Return(expectedResponse, nil)

	req, _ := http.NewRequest("DELETE", "/api/v1/customer/cust-12345", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-ID", "test-request-id")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response DeleteCustomerResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Customer deleted successfully", response.Message)

	mockClient.AssertExpectations(t)
}

// TestDeleteCustomer_GRPCFailure tests customer deletion when gRPC call fails
func TestDeleteCustomer_GRPCFailure(t *testing.T) {
	mockClient := new(mock_client.MockAccountClient)
	accountHandler := NewAccountHandler(mockClient)
	router := setupCustomerRoutes(accountHandler)

	mockClient.On("DeleteCustomer", mock.Anything, &protoacc.DeleteCustomerRequest{
		CustomerId: "cust-12345",
		Metadata: &protoacc.Metadata{
			RequestId: "",
			Requester: "test-admin",
		},
	}).Return(nil, errors.New("gRPC connection failed"))

	req, _ := http.NewRequest("DELETE", "/api/v1/customer/cust-12345", nil)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "gRPC connection failed", response.Error)

	mockClient.AssertExpectations(t)
}

// TestDeleteCustomer_UnsuccessfulResponse tests customer deletion with unsuccessful gRPC response
func TestDeleteCustomer_UnsuccessfulResponse(t *testing.T) {
	mockClient := new(mock_client.MockAccountClient)
	accountHandler := NewAccountHandler(mockClient)
	router := setupCustomerRoutes(accountHandler)

	expectedResponse := &protoacc.DeleteCustomerResponse{
		Response: &protoacc.Response{
			Success: false,
			Message: "Customer not found",
		},
	}

	mockClient.On("DeleteCustomer", mock.Anything, &protoacc.DeleteCustomerRequest{
		CustomerId: "nonexistent",
		Metadata: &protoacc.Metadata{
			RequestId: "",
			Requester: "test-admin",
		},
	}).Return(expectedResponse, nil)

	req, _ := http.NewRequest("DELETE", "/api/v1/customer/nonexistent", nil)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Customer not found", response.Error)

	mockClient.AssertExpectations(t)
}

// TestListCustomer_Success tests successful customer listing
func TestListCustomer_Success(t *testing.T) {
	mockClient := new(mock_client.MockAccountClient)
	accountHandler := NewAccountHandler(mockClient)
	router := setupCustomerRoutes(accountHandler)

	// Create sample customers
	expectedCustomers := []*protoacc.Customer{
		{
			Id:   "cust-1",
			Name: "Customer One",
		},
		{
			Id:   "cust-2",
			Name: "Jane Smith",
		},
	}

	expectedResponse := &protoacc.ListCustomersResponse{
		Customers: expectedCustomers,
		Pagination: &protoacc.PaginationResponse{
			Page:       1,
			PageSize:   50,
			TotalCount: 2,
			TotalPages: 1,
		},
		Response: &protoacc.Response{
			Success: true,
			Message: "Customers retrieved successfully",
		},
	}

	mockClient.On("ListCustomer", mock.Anything, mock.MatchedBy(func(req *protoacc.ListCustomersRequest) bool {
		if req.SortOrder != "desc" {
			return false
		}
		if req.Pagination == nil {
			return false
		}
		if req.Pagination.Page != 1 || req.Pagination.PageSize != 50 {
			return false
		}
		if req.Metadata == nil || req.Metadata.Requester != "test-admin" {
			return false
		}
		return true
	})).Return(expectedResponse, nil)

	req, _ := http.NewRequest("GET", "/api/v1/customer?page=1&pagesize=50&order=desc", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-ID", "test-request-id")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response ListCustomerResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 1, response.Page)
	assert.Equal(t, 50, response.PageSize)
	assert.Equal(t, 2, response.TotalCount)
	assert.Equal(t, 1, response.TotalPages)
	assert.Equal(t, "Customers retrieved successfully", response.Message)

	customers, ok := response.Customers.([]interface{})
	assert.True(t, ok)
	assert.Len(t, customers, 2)

	mockClient.AssertExpectations(t)
}

// TestListCustomer_EmptyList tests customer listing with no customers
func TestListCustomer_EmptyList(t *testing.T) {
	mockClient := new(mock_client.MockAccountClient)
	accountHandler := NewAccountHandler(mockClient)
	router := setupCustomerRoutes(accountHandler)

	expectedResponse := &protoacc.ListCustomersResponse{
		Customers: []*protoacc.Customer{},
		Pagination: &protoacc.PaginationResponse{
			Page:       1,
			PageSize:   50,
			TotalCount: 0,
			TotalPages: 0,
		},
		Response: &protoacc.Response{
			Success: true,
			Message: "No customers found",
		},
	}

	mockClient.On("ListCustomer", mock.Anything, mock.MatchedBy(func(req *protoacc.ListCustomersRequest) bool {
		return req.SortOrder == "desc" &&
			req.Pagination != nil &&
			req.Pagination.Page == 1 &&
			req.Pagination.PageSize == 50 &&
			req.Metadata != nil &&
			req.Metadata.Requester == "test-admin"
	})).Return(expectedResponse, nil)

	req, _ := http.NewRequest("GET", "/api/v1/customer?page=1&pagesize=50&order=desc", nil)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response ListCustomerResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 1, response.Page)
	assert.Equal(t, 50, response.PageSize)
	assert.Equal(t, 0, response.TotalCount)
	assert.Equal(t, 0, response.TotalPages)
	assert.Equal(t, "No customers found", response.Message)

	mockClient.AssertExpectations(t)
}

// TestListCustomer_InvalidPagination tests customer listing with invalid pagination parameters
func TestListCustomer_InvalidPagination(t *testing.T) {
	mockClient := new(mock_client.MockAccountClient)
	accountHandler := NewAccountHandler(mockClient)
	router := setupCustomerRoutes(accountHandler)

	expectedResponse := &protoacc.ListCustomersResponse{
		Customers: []*protoacc.Customer{},
		Pagination: &protoacc.PaginationResponse{
			Page:       -1,
			PageSize:   -1,
			TotalCount: 0,
			TotalPages: 0,
		},
		Response: &protoacc.Response{
			Success: true,
			Message: "Customers retrieved successfully",
		},
	}

	mockClient.On("ListCustomer", mock.Anything, mock.MatchedBy(func(req *protoacc.ListCustomersRequest) bool {
		return req.SortOrder == "" &&
			req.Pagination != nil &&
			req.Pagination.Page == -1 &&
			req.Pagination.PageSize == -1 &&
			req.Metadata != nil &&
			req.Metadata.Requester == "test-admin"
	})).Return(expectedResponse, nil)

	req, _ := http.NewRequest("GET", "/api/v1/customer?page=invalid&pagesize=invalid", nil)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response ListCustomerResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, -1, response.Page)
	assert.Equal(t, -1, response.PageSize)

	mockClient.AssertExpectations(t)
}

// TestListCustomer_GRPCFailure tests customer listing when gRPC call fails
func TestListCustomer_GRPCFailure(t *testing.T) {
	mockClient := new(mock_client.MockAccountClient)
	accountHandler := NewAccountHandler(mockClient)
	router := setupCustomerRoutes(accountHandler)

	mockClient.On("ListCustomer", mock.Anything, mock.MatchedBy(func(req *protoacc.ListCustomersRequest) bool {
		return req.SortOrder == "desc" &&
			req.Pagination != nil &&
			req.Pagination.Page == 1 &&
			req.Pagination.PageSize == 50 &&
			req.Metadata != nil &&
			req.Metadata.Requester == "test-admin"
	})).Return(nil, errors.New("gRPC connection failed"))

	req, _ := http.NewRequest("GET", "/api/v1/customer?page=1&pagesize=50&order=desc", nil)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "gRPC connection failed", response.Error)

	mockClient.AssertExpectations(t)
}

// TestListCustomerAccounts_Success tests successful customer accounts listing
func TestListCustomerAccounts_Success(t *testing.T) {
	mockClient := new(mock_client.MockAccountClient)
	accountHandler := NewAccountHandler(mockClient)
	router := setupCustomerRoutes(accountHandler)

	// Create sample accounts
	expectedAccounts := []*protoacc.Account{
		{
			Id:         "acc-1",
			CustomerId: "cust-123",
			Balance:    1000.50,
		},
		{
			Id:         "acc-2",
			CustomerId: "cust-123",
			Balance:    2500.75,
		},
	}

	expectedResponse := &protoacc.ListAccountsResponse{
		Accounts: expectedAccounts,
		Pagination: &protoacc.PaginationResponse{
			Page:       1,
			PageSize:   50,
			TotalCount: 2,
			TotalPages: 1,
		},
		Response: &protoacc.Response{
			Success: true,
			Message: "Accounts retrieved successfully",
		},
	}

	mockClient.On("ListAccount", mock.Anything, mock.MatchedBy(func(req *protoacc.ListAccountsRequest) bool {
		return req.CustomerId == "cust-123" &&
			req.Metadata != nil &&
			req.Metadata.Requester == "test-admin" &&
			req.Pagination != nil &&
			req.Pagination.Page == 1 &&
			req.Pagination.PageSize == 50
	})).Return(expectedResponse, nil)

	req, _ := http.NewRequest("GET", "/api/v1/customer/cust-123/account?page=1&pagesize=50", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-ID", "test-request-id")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response ListAccountResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 1, response.Page)
	assert.Equal(t, 50, response.PageSize)
	assert.Equal(t, 2, response.TotalCount)
	assert.Equal(t, 1, response.TotalPages)
	assert.Equal(t, "Accounts retrieved successfully", response.Message)

	// Verify accounts data is properly marshaled
	accounts, ok := response.Accounts.([]interface{})
	assert.True(t, ok)
	assert.Len(t, accounts, 2)

	mockClient.AssertExpectations(t)
}

// TestListCustomerAccounts_NoAccounts tests customer accounts listing with no accounts
func TestListCustomerAccounts_NoAccounts(t *testing.T) {
	mockClient := new(mock_client.MockAccountClient)
	accountHandler := NewAccountHandler(mockClient)
	router := setupCustomerRoutes(accountHandler)

	expectedResponse := &protoacc.ListAccountsResponse{
		Accounts: []*protoacc.Account{},
		Pagination: &protoacc.PaginationResponse{
			Page:       1,
			PageSize:   50,
			TotalCount: 0,
			TotalPages: 0,
		},
		Response: &protoacc.Response{
			Success: true,
			Message: "No accounts found for customer",
		},
	}

	mockClient.On("ListAccount", mock.Anything, mock.MatchedBy(func(req *protoacc.ListAccountsRequest) bool {
		return req.CustomerId == "cust-123" &&
			req.Metadata != nil &&
			req.Metadata.Requester == "test-admin" &&
			req.Pagination != nil &&
			req.Pagination.Page == 1 &&
			req.Pagination.PageSize == 50
	})).Return(expectedResponse, nil)

	req, _ := http.NewRequest("GET", "/api/v1/customer/cust-123/account?page=1&pagesize=50", nil)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response ListAccountResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 1, response.Page)
	assert.Equal(t, 50, response.PageSize)
	assert.Equal(t, 0, response.TotalCount)
	assert.Equal(t, 0, response.TotalPages)
	assert.Equal(t, "No accounts found for customer", response.Message)

	mockClient.AssertExpectations(t)
}

// TestListCustomerAccounts_InvalidPagination tests customer accounts listing with invalid pagination
func TestListCustomerAccounts_InvalidPagination(t *testing.T) {
	mockClient := new(mock_client.MockAccountClient)
	accountHandler := NewAccountHandler(mockClient)
	router := setupCustomerRoutes(accountHandler)

	expectedResponse := &protoacc.ListAccountsResponse{
		Accounts: []*protoacc.Account{},
		Pagination: &protoacc.PaginationResponse{
			Page:       -1,
			PageSize:   -1,
			TotalCount: 0,
			TotalPages: 0,
		},
		Response: &protoacc.Response{
			Success: true,
			Message: "Accounts retrieved successfully",
		},
	}

	mockClient.On("ListAccount", mock.Anything, mock.MatchedBy(func(req *protoacc.ListAccountsRequest) bool {
		return req.CustomerId == "cust-123" &&
			req.Metadata != nil &&
			req.Metadata.Requester == "test-admin" &&
			req.Pagination != nil &&
			req.Pagination.Page == -1 &&
			req.Pagination.PageSize == -1
	})).Return(expectedResponse, nil)

	req, _ := http.NewRequest("GET", "/api/v1/customer/cust-123/account?page=invalid&pagesize=invalid", nil)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response ListAccountResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, -1, response.Page)
	assert.Equal(t, -1, response.PageSize)

	mockClient.AssertExpectations(t)
}

// TestListCustomerAccounts_GRPCFailure tests customer accounts listing when gRPC call fails
func TestListCustomerAccounts_GRPCFailure(t *testing.T) {
	mockClient := new(mock_client.MockAccountClient)
	accountHandler := NewAccountHandler(mockClient)
	router := setupCustomerRoutes(accountHandler)

	mockClient.On("ListAccount", mock.Anything, mock.MatchedBy(func(req *protoacc.ListAccountsRequest) bool {
		return req.CustomerId == "cust-123" &&
			req.Metadata != nil &&
			req.Metadata.Requester == "test-admin" &&
			req.Pagination != nil &&
			req.Pagination.Page == 1 &&
			req.Pagination.PageSize == 50
	})).Return(nil, errors.New("gRPC connection failed"))

	req, _ := http.NewRequest("GET", "/api/v1/customer/cust-123/account?page=1&pagesize=50", nil)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "gRPC connection failed", response.Error)

	mockClient.AssertExpectations(t)
}

// TestListAccounts_SuccessDefaultPagination tests success with default pagination
func TestListAccounts_SuccessDefaultPagination(t *testing.T) {
	mockClient := new(mock_client.MockAccountClient)
	accountHandler := &AccountHandler{AccountClient: mockClient}
	router := setupAccountRoutes(accountHandler)

	expectedResponse := &protoacc.ListAccountsResponse{
		Accounts: []*protoacc.Account{
			{Id: "acc-1", CustomerId: "cust-123", Balance: 1000.0, ActiveStatus: "active"},
			{Id: "acc-2", CustomerId: "cust-123", Balance: 2000.0, ActiveStatus: "active"},
		},
		Pagination: &protoacc.PaginationResponse{
			Page:       1,
			PageSize:   50,
			TotalCount: 2,
			TotalPages: 1,
		},
		Response: &protoacc.Response{
			Success: true,
			Message: "Accounts retrieved successfully",
		},
	}

	mockClient.On("ListAccount", mock.Anything, mock.MatchedBy(func(req *protoacc.ListAccountsRequest) bool {
		return req.CustomerId == "" &&
			req.MinBalance == "" &&
			req.InTransaction == "" &&
			req.SortOrder == "" &&
			req.Pagination.Page == -1 && // Default when no pagination provided
			req.Pagination.PageSize == -1 &&
			req.Metadata != nil &&
			req.Metadata.Requester == "test-admin" &&
			req.Metadata.RequestId == "test-request-id"
	})).Return(expectedResponse, nil)

	req, _ := http.NewRequest("GET", "/api/v1/account", nil)
	req.Header.Set("Authorization", "Bearer token")
	req.Header.Set("X-Request-ID", "test-request-id")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Response: %s", w.Code, w.Body.String())

	var response ListAccountResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err, "Failed to unmarshal response: %s", w.Body.String())
	assert.Equal(t, 1, response.Page)
	assert.Equal(t, 50, response.PageSize)
	assert.Equal(t, 2, response.TotalCount)
	assert.Equal(t, 1, response.TotalPages)
	assert.Equal(t, "Accounts retrieved successfully", response.Message)
	assert.NotNil(t, response.Accounts)

	mockClient.AssertExpectations(t)
}

// TestListAccounts_SuccessWithCustomerFilter with custom filters
func TestListAccounts_SuccessWithCustomerFilter(t *testing.T) {
	mockClient := new(mock_client.MockAccountClient)
	accountHandler := &AccountHandler{AccountClient: mockClient}
	router := setupAccountRoutes(accountHandler)

	expectedResponse := &protoacc.ListAccountsResponse{
		Accounts: []*protoacc.Account{
			{Id: "acc-1", CustomerId: "cust-123", Balance: 1000.0, ActiveStatus: "active"},
			{Id: "acc-2", CustomerId: "cust-123", Balance: 2000.0, ActiveStatus: "active"},
		},
		Pagination: &protoacc.PaginationResponse{
			Page:       1,
			PageSize:   50,
			TotalCount: 2,
			TotalPages: 1,
		},
		Response: &protoacc.Response{
			Success: true,
			Message: "Customer accounts retrieved successfully",
		},
	}

	mockClient.On("ListAccount", mock.Anything, mock.MatchedBy(func(req *protoacc.ListAccountsRequest) bool {
		return req.CustomerId == "cust-123" &&
			req.Metadata != nil &&
			req.Metadata.Requester == "test-admin" &&
			req.Metadata.RequestId == "test-request-id"
	})).Return(expectedResponse, nil)

	req, _ := http.NewRequest("GET", "/api/v1/account?customer_id=cust-123", nil)
	req.Header.Set("Authorization", "Bearer token")
	req.Header.Set("X-Request-ID", "test-request-id")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Expected status 200 but got %d. Response: %s", w.Code, w.Body.String())

	var response ListAccountResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err, "Failed to unmarshal response: %s", w.Body.String())
	assert.Equal(t, "cust-123", expectedResponse.Accounts[0].CustomerId)
	assert.Equal(t, "Customer accounts retrieved successfully", response.Message)

	mockClient.AssertExpectations(t)
}

// TestListAccounts_InvalidPageNumber if invalid page number or pages size is provided
func TestListAccounts_InvalidPageNumber(t *testing.T) {
	mockClient := new(mock_client.MockAccountClient)
	accountHandler := &AccountHandler{AccountClient: mockClient}
	router := setupAccountRoutes(accountHandler)

	expectedResponse := &protoacc.ListAccountsResponse{
		Accounts: []*protoacc.Account{},
		Pagination: &protoacc.PaginationResponse{
			Page:       -1,
			PageSize:   -1,
			TotalCount: 0,
			TotalPages: 0,
		},
		Response: &protoacc.Response{
			Success: true,
			Message: "Accounts retrieved successfully",
		},
	}

	mockClient.On("ListAccount", mock.Anything, mock.MatchedBy(func(req *protoacc.ListAccountsRequest) bool {
		return req.Pagination.Page == -1 &&
			req.Pagination.PageSize == -1 &&
			req.Metadata != nil &&
			req.Metadata.Requester == "test-admin" &&
			req.Metadata.RequestId == "test-request-id"
	})).Return(expectedResponse, nil)

	req, _ := http.NewRequest("GET", "/api/v1/account?page=invalid&pagesize=abc", nil)
	req.Header.Set("Authorization", "Bearer token")
	req.Header.Set("X-Request-ID", "test-request-id")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Response: %s", w.Code, w.Body.String())

	mockClient.AssertExpectations(t)
}
