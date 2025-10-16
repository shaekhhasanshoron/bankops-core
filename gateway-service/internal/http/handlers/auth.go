package handlers

import (
	"errors"
	protoauth "gateway-service/api/protogen/authservice/proto"
	"gateway-service/internal/logging"
	"gateway-service/internal/ports"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
)

type AuthHandler struct {
	AuthClient ports.AuthClient
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

type ListEmployeeResponse struct {
	Employees  interface{} `json:"employees"`
	Page       int         `json:"page"`
	PageSize   int         `json:"pageSize"`
	TotalCount int         `json:"totalCount"`
	TotalPages int         `json:"totalPages"`
	Message    string      `json:"message" binding:"message"`
}

func NewAuthHandler(authClient ports.AuthClient) *AuthHandler {
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

	grpcReq := &protoauth.AuthenticateRequest{
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

	grpcReq := &protoauth.CreateEmployeeRequest{
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
// @Summary Update Employee
// @Description Update an employee by username
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

	grpcReq := &protoauth.UpdateRoleRequest{
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

	res := UpdateCustomerResponse{
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

	grpcReq := &protoauth.DeleteEmployeeRequest{
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

// ListEmployee for fetching employee list
// @Tags Employee
// @Summary Get Employee List
// @Description Get employee list.
// @Accept json
// @Produce json
// @Param page query int false "Page number for pagination" default(1)
// @Param pagesize query int false "Number of employee per page" default(50)
// @Param order query string false "Sort order (asc/desc)" default(desc)
// @Param Authorization header string true "Bearer token for authorization, include 'Bearer ' followed by access_token"
// @Success 200 {object} ListEmployeeResponse "Successfully retrieved employee list"
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/employee [get]
func (h *AuthHandler) ListEmployee(c *gin.Context) {
	pageStr := strings.TrimSpace(c.Query("page"))
	pageSizeStr := strings.TrimSpace(c.Query("pagesize"))
	order := strings.TrimSpace(c.Query("order"))

	var pageNo int = -1
	var err error
	if pageStr != "" {
		pageNo, err = strconv.Atoi(pageStr)
		if err != nil {
			logging.Logger.Warn().Err(err).Msg("Invalid page no: " + pageStr)
			pageNo = -1
		}
	}

	var pageSize int = -1
	if pageSizeStr != "" {
		pageSize, err = strconv.Atoi(pageSizeStr)
		if err != nil {
			logging.Logger.Warn().Err(err).Msg("Invalid page size: " + pageSizeStr)
			pageSize = -1
		}
	}

	un, _ := c.Get("username") // middleware
	requester, ok := un.(string)
	if !ok {
		logging.Logger.Warn().Err(errors.New("unable to get requester username")).Msg("requester: " + requester)
	}

	grpcReq := &protoauth.ListEmployeeRequest{
		SortOrder: order,
		Page:      int32(pageNo),
		PageSize:  int32(pageSize),
	}

	resp, err := h.AuthClient.ListEmployee(c.Request.Context(), grpcReq)
	if err != nil {
		logging.Logger.Error().Err(err).Msg("failed to get employee")
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
		return
	}

	if !resp.Success {
		logging.Logger.Error().Err(errors.New(resp.Message)).Msg("unable to get employee")
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: resp.Message})
		return
	}

	res := ListEmployeeResponse{
		Employees:  resp.Employees,
		Page:       int(resp.Page),
		PageSize:   int(resp.PageSize),
		TotalCount: int(resp.TotalCount),
		TotalPages: int(resp.TotalPages),
		Message:    resp.Message,
	}

	c.JSON(http.StatusOK, res)
}
