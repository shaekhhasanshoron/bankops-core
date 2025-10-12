package handlers

import (
	"errors"
	"gateway-service/api/proto"
	"gateway-service/internal/grpc/clients"
	"gateway-service/internal/logging"
	"github.com/gin-gonic/gin"
	"net/http"
)

type AuthHandler struct {
	AuthClient clients.AuthClient
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
	//RefreshToken string `json:"refresh_token"`
}

type CreateEmployeeRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role" binding:"required"`
}

type CreateEmployeeResponse struct {
	Username string `json:"username" binding:"required"`
	Message  string `json:"message" binding:"message"`
}

type DeleteEmployeeRequest struct {
	Username string `json:"username" binding:"required"`
}

type DeleteEmployeeResponse struct {
	Message string `json:"message" binding:"message"`
}

type UpdateEmployeeRequest struct {
	Role string `json:"role" binding:"required"`
}

type UpdateEmployeeResponse struct {
	Message string `json:"message" binding:"message"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewAuthHandler(authClient clients.AuthClient) *AuthHandler {
	return &AuthHandler{
		AuthClient: authClient,
	}
}

// Login api for employee login
// @Tags Authentication
// @Summary Login API
// @Description Login to get access token using username and password
// @Accept json
// @Produce json
// @Param login body LoginRequest true "Login credentials"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logging.Logger.Warn().Err(err).Msg("invalid request param")
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request payload"})
		return
	}

	grpcReq := &proto.AuthenticateRequest{
		Username: req.Username,
		Password: req.Password,
	}

	resp, err := h.AuthClient.Authenticate(c.Request.Context(), grpcReq)
	if err != nil {
		logging.Logger.Error().Err(err).Msg("login failed")
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid credentials"})
		return
	}

	loginResp := LoginResponse{
		AccessToken: resp.Token,
		//RefreshToken: resp.RefreshToken,
	}

	c.JSON(http.StatusOK, loginResp)
}

// CreateEmployee for creating new employee
// @Tags Employee
// @Summary Create Employee
// @Description Create employee - Username: lowercase + underscores only (in middle) | Role: viewer/admin/editor | Bearer token required
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token for authorization, include 'Bearer ' followed by access_token"
// @Param employee body CreateEmployeeRequest true "Employee details"
// @Success 201 {object} CreateEmployeeResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/employee [post]
func (h *AuthHandler) CreateEmployee(c *gin.Context) {
	var req CreateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logging.Logger.Warn().Err(err).Msg("invalid request param")
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request payload"})
		return
	}

	un, _ := c.Get("username") // middleware
	requester, ok := un.(string)
	if !ok {
		logging.Logger.Warn().Err(errors.New("unable to get requester username")).Msg("requester: " + requester)
	}

	grpcReq := &proto.CreateEmployeeRequest{
		Username:  req.Username,
		Password:  req.Password,
		Role:      req.Role,
		Requester: requester,
	}

	resp, err := h.AuthClient.CreateEmployee(c.Request.Context(), grpcReq)
	if err != nil {
		logging.Logger.Error().Err(err).Msg("failed to create new employee")
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid credentials"})
		return
	}

	if !resp.Success {
		logging.Logger.Error().Err(errors.New(resp.Message)).Msg("unable to create employee")
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: resp.Message})
		return
	}

	res := CreateEmployeeResponse{
		Username: req.Username,
		Message:  resp.Message,
	}

	c.JSON(http.StatusCreated, res)
}

// UpdateEmployee for update an employee
// @Tags Employee
// @Summary Delete Employee
// @Description Delete an employee by username
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token for authorization, include 'Bearer ' followed by access_token"
// @Param username path string true "Username of the employee"
// @Param employee body UpdateEmployeeRequest true "Employee details"
// @Success 200 {string} {object} UpdateEmployeeResponse
// @Failure 400 {string} {object} ErrorResponse
// @Failure 401 {string} {object} ErrorResponse
// @Router /api/v1/employee/{username} [put]
func (h *AuthHandler) UpdateEmployee(c *gin.Context) {
	username := c.Param("username")

	var req UpdateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logging.Logger.Warn().Err(err).Msg("invalid request param")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	un, _ := c.Get("username") // middleware
	requester, ok := un.(string)
	if !ok {
		logging.Logger.Warn().Err(errors.New("unable to get requester username")).Msg("requester: " + username)
	}

	grpcReq := &proto.UpdateRoleRequest{
		Username:  username,
		Role:      req.Role,
		Requester: requester,
	}

	resp, err := h.AuthClient.UpdateEmployee(c.Request.Context(), grpcReq)
	if err != nil {
		logging.Logger.Error().Err(err).Msg("failed to update employee")
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
		return
	}

	if !resp.Success {
		logging.Logger.Error().Err(errors.New(resp.Message)).Msg("unable to update employee")
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: resp.Message})
		return
	}

	res := DeleteEmployeeResponse{
		Message: resp.Message,
	}

	c.JSON(http.StatusOK, res)
}

// DeleteEmployee for delete an employee by username
// @Tags Employee
// @Summary Delete Employee
// @Description Delete an employee by username
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token for authorization, include 'Bearer ' followed by access_token"
// @Param username path string true "Username of the employee"
// @Success 200 {string} {object} DeleteEmployeeResponse
// @Failure 400 {string} {object} ErrorResponse
// @Failure 401 {string} {object} ErrorResponse
// @Router /api/v1/employee/{username} [delete]
func (h *AuthHandler) DeleteEmployee(c *gin.Context) {
	username := c.Param("username")

	un, _ := c.Get("username") // middleware
	requester, ok := un.(string)
	if !ok {
		logging.Logger.Warn().Err(errors.New("unable to get requester username")).Msg("requester: " + requester)
	}

	grpcReq := &proto.DeleteEmployeeRequest{
		Username:  username,
		Requester: requester,
	}

	resp, err := h.AuthClient.DeleteEmployee(c.Request.Context(), grpcReq)
	if err != nil {
		logging.Logger.Error().Err(err).Msg("failed to delete employee")
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
		return
	}

	if !resp.Success {
		logging.Logger.Error().Err(errors.New(resp.Message)).Msg("unable to delete employee")
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: resp.Message})
		return
	}

	res := DeleteEmployeeResponse{
		Message: resp.Message,
	}

	c.JSON(http.StatusOK, res)
}
