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
	"google.golang.org/protobuf/types/known/timestamppb"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupEmployeeRoutes(authHandler *AuthHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(func(c *gin.Context) {
		c.Set("username", "test-admin")
		c.Next()
	})

	router.POST("/api/v1/employee", authHandler.CreateEmployee)
	router.PUT("/api/v1/employee/:username", authHandler.UpdateEmployee)
	router.DELETE("/api/v1/employee/:username", authHandler.DeleteEmployee)
	router.GET("/api/v1/employee", authHandler.ListEmployee)

	return router
}

// TestCreateEmployee_Success tests successful employee creation
func TestCreateEmployee_Success(t *testing.T) {
	mockClient := new(mock_client.MockAuthClient)
	authHandler := NewAuthHandler(mockClient)
	router := setupEmployeeRoutes(authHandler)

	createReq := CreateEmployeeRequest{
		Username: "newemployee",
		Password: "password123",
		Role:     "viewer",
	}

	expectedResponse := &protoauth.CreateEmployeeResponse{
		Success: true,
		Message: "Employee created successfully",
	}

	mockClient.On("CreateEmployee", mock.Anything, &protoauth.CreateEmployeeRequest{
		Username:  "newemployee",
		Password:  "password123",
		Role:      "viewer",
		Requester: "test-admin",
	}).Return(expectedResponse, nil)

	body, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/api/v1/employee", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response CreateEmployeeResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "newemployee", response.Username)
	assert.Equal(t, "Employee created successfully", response.Message)

	mockClient.AssertExpectations(t)
}

// TestCreateEmployee_InvalidPayload tests employee creation with invalid payload
func TestCreateEmployee_InvalidPayload(t *testing.T) {
	mockClient := new(mock_client.MockAuthClient)
	authHandler := NewAuthHandler(mockClient)
	router := setupEmployeeRoutes(authHandler)

	invalidJSON := `{"username": "test"}`

	req, _ := http.NewRequest("POST", "/api/v1/employee", bytes.NewBufferString(invalidJSON))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid request payload", response.Error)
}

// TestCreateEmployee_GRPCFailure tests employee creation when gRPC call fails
func TestCreateEmployee_GRPCFailure(t *testing.T) {
	mockClient := new(mock_client.MockAuthClient)
	authHandler := NewAuthHandler(mockClient)
	router := setupEmployeeRoutes(authHandler)

	createReq := CreateEmployeeRequest{
		Username: "newemployee",
		Password: "password123",
		Role:     "viewer",
	}

	mockClient.On("CreateEmployee", mock.Anything, &protoauth.CreateEmployeeRequest{
		Username:  "newemployee",
		Password:  "password123",
		Role:      "viewer",
		Requester: "test-admin",
	}).Return(nil, errors.New("gRPC connection failed"))

	body, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/api/v1/employee", bytes.NewBuffer(body))
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

// TestCreateEmployee_UnsuccessfulResponse tests employee creation with unsuccessful gRPC response
func TestCreateEmployee_UnsuccessfulResponse(t *testing.T) {
	mockClient := new(mock_client.MockAuthClient)
	authHandler := NewAuthHandler(mockClient)
	router := setupEmployeeRoutes(authHandler)

	createReq := CreateEmployeeRequest{
		Username: "newemployee",
		Password: "password123",
		Role:     "viewer",
	}

	expectedResponse := &protoauth.CreateEmployeeResponse{
		Success: false,
		Message: "Username already exists",
	}

	mockClient.On("CreateEmployee", mock.Anything, &protoauth.CreateEmployeeRequest{
		Username:  "newemployee",
		Password:  "password123",
		Role:      "viewer",
		Requester: "test-admin",
	}).Return(expectedResponse, nil)

	body, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/api/v1/employee", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Username already exists", response.Error)

	mockClient.AssertExpectations(t)
}

// TestDeleteEmployee_Success tests successful employee deletion
func TestDeleteEmployee_Success(t *testing.T) {
	mockClient := new(mock_client.MockAuthClient)
	authHandler := NewAuthHandler(mockClient)
	router := setupEmployeeRoutes(authHandler)

	expectedResponse := &protoauth.DeleteEmployeeResponse{
		Success: true,
		Message: "Employee deleted successfully",
	}

	mockClient.On("DeleteEmployee", mock.Anything, &protoauth.DeleteEmployeeRequest{
		Username:  "usertodelete",
		Requester: "test-admin",
	}).Return(expectedResponse, nil)

	req, _ := http.NewRequest("DELETE", "/api/v1/employee/usertodelete", nil)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response DeleteEmployeeResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Employee deleted successfully", response.Message)

	mockClient.AssertExpectations(t)
}

// TestDeleteEmployee_GRPCFailure tests employee deletion when gRPC call fails
func TestDeleteEmployee_GRPCFailure(t *testing.T) {
	mockClient := new(mock_client.MockAuthClient)
	authHandler := NewAuthHandler(mockClient)
	router := setupEmployeeRoutes(authHandler)

	mockClient.On("DeleteEmployee", mock.Anything, &protoauth.DeleteEmployeeRequest{
		Username:  "usertodelete",
		Requester: "test-admin",
	}).Return(nil, errors.New("gRPC connection failed"))

	req, _ := http.NewRequest("DELETE", "/api/v1/employee/usertodelete", nil)
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

// TestDeleteEmployee_UnsuccessfulResponse tests employee deletion with unsuccessful gRPC response
func TestDeleteEmployee_UnsuccessfulResponse(t *testing.T) {
	mockClient := new(mock_client.MockAuthClient)
	authHandler := NewAuthHandler(mockClient)
	router := setupEmployeeRoutes(authHandler)

	expectedResponse := &protoauth.DeleteEmployeeResponse{
		Success: false,
		Message: "Employee not found",
	}

	mockClient.On("DeleteEmployee", mock.Anything, &protoauth.DeleteEmployeeRequest{
		Username:  "nonexistent",
		Requester: "test-admin",
	}).Return(expectedResponse, nil)

	req, _ := http.NewRequest("DELETE", "/api/v1/employee/nonexistent", nil)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Employee not found", response.Error)

	mockClient.AssertExpectations(t)
}

// TestListEmployee_Success tests successful employee listing
func TestListEmployee_Success(t *testing.T) {
	mockClient := new(mock_client.MockAuthClient)
	authHandler := NewAuthHandler(mockClient)
	router := setupEmployeeRoutes(authHandler)

	expectedEmployees := []*protoauth.Employee{
		{
			UserName:  "user1",
			Role:      "admin",
			CreatedAt: timestamppb.Now(),
		},
		{
			UserName:  "user2",
			Role:      "viewer",
			CreatedAt: timestamppb.Now(),
		},
	}

	expectedResponse := &protoauth.ListEmployeeResponse{
		Success:    true,
		Employees:  expectedEmployees,
		Page:       1,
		PageSize:   50,
		TotalCount: 2,
		TotalPages: 1,
		Message:    "Employees retrieved successfully",
	}

	mockClient.On("ListEmployee", mock.Anything, &protoauth.ListEmployeeRequest{
		SortOrder: "desc",
		Page:      1,
		PageSize:  50,
	}).Return(expectedResponse, nil)

	req, _ := http.NewRequest("GET", "/api/v1/employee?page=1&pagesize=50&order=desc", nil)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response ListEmployeeResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 1, response.Page)
	assert.Equal(t, 50, response.PageSize)
	assert.Equal(t, 2, response.TotalCount)
	assert.Equal(t, 1, response.TotalPages)
	assert.Equal(t, "Employees retrieved successfully", response.Message)
	assert.Len(t, response.Employees, 2)

	mockClient.AssertExpectations(t)
}

// TestListEmployee_InvalidPagination tests employee listing with invalid pagination parameters
func TestListEmployee_InvalidPagination(t *testing.T) {
	mockClient := new(mock_client.MockAuthClient)
	authHandler := NewAuthHandler(mockClient)
	router := setupEmployeeRoutes(authHandler)

	expectedResponse := &protoauth.ListEmployeeResponse{
		Success:    true,
		Employees:  []*protoauth.Employee{},
		Page:       -1,
		PageSize:   -1,
		TotalCount: 0,
		TotalPages: 0,
		Message:    "Employees retrieved successfully",
	}

	mockClient.On("ListEmployee", mock.Anything, &protoauth.ListEmployeeRequest{
		SortOrder: "",
		Page:      -1,
		PageSize:  -1,
	}).Return(expectedResponse, nil)

	req, _ := http.NewRequest("GET", "/api/v1/employee?page=invalid&pagesize=invalid", nil)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response ListEmployeeResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, -1, response.Page)
	assert.Equal(t, -1, response.PageSize)

	mockClient.AssertExpectations(t)
}

// TestListEmployee_GRPCFailure tests employee listing when gRPC call fails
func TestListEmployee_GRPCFailure(t *testing.T) {
	mockClient := new(mock_client.MockAuthClient)
	authHandler := NewAuthHandler(mockClient)
	router := setupEmployeeRoutes(authHandler)

	mockClient.On("ListEmployee", mock.Anything, &protoauth.ListEmployeeRequest{
		SortOrder: "desc",
		Page:      1,
		PageSize:  50,
	}).Return(nil, errors.New("gRPC connection failed"))

	req, _ := http.NewRequest("GET", "/api/v1/employee?page=1&pagesize=50&order=desc", nil)
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
