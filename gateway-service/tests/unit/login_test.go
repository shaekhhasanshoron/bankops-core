package unit

import (
	"bytes"
	"encoding/json"
	"fmt"
	protoauth "gateway-service/api/protogen/authservice/proto"
	"gateway-service/internal/http/handlers"
	mock_grpc_client "gateway-service/internal/ports/mocks/grpc_client"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestLoginAPI_Success simply test login success result
func TestLoginAPI_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockClient := &mock_grpc_client.MockAuthClient{}
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
