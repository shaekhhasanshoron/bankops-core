package handlers

import (
	"errors"
	protoauth "gateway-service/api/protogen/authservice/proto"
	"gateway-service/internal/logging"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
)

type CreateEmployeeRequest struct {
	Username string `json:"username" binding:"required,max=50"`
	Password string `json:"password" binding:"required,max=50"`
	Role     string `json:"role" binding:"required,oneof=admin viewer editor"`
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

// CreateEmployee creates a new employee
// @Tags Employee
// @Summary Create Employee
// @Description
// @Description **Request Body:**
// @Description
// @Description Username:
// @Description - Required
// @Description - Max 50 characters
// @Description - Lowercase letters only
// @Description - Underscores allowed only in middle
// @Description
// @Description Password:
// @Description - Required
// @Description - Max 50 characters
// @Description - Supports only A-Z, a-z, 0-9, and these special characters: ! - _ & $ @ # [ ]
// @Description
// @Description Role:
// @Description - Required
// @Description - Options: **admin**, **viewer**, **editor**
// @Description
// @Description **Header:**
// @Description
// @Description Authorization:
// @Description - Required
// @Description - Format: Bearer token
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token" default(Bearer )
// @Param employee body CreateEmployeeRequest true "Employee creation data"
// @Success 201 {object} CreateEmployeeResponse "Employee created successfully"
// @Failure 400 {object} ErrorResponse "Invalid input data"
// @Failure 401 {object} ErrorResponse "Unauthorized - Missing or invalid token"
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

// DeleteEmployee deletes an employee by username
// @Tags Employee
// @Summary Delete Employee
// @Description
// @Description **Header:**
// @Description
// @Description Authorization:
// @Description - Required
// @Description - Format: Bearer token
// @Description
// @Description **Path Parameter:**
// @Description
// @Description username:
// @Description - Required
// @Description - Username of the employee to delete
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token" default(Bearer )
// @Param username path string true "Username of the employee"
// @Success 200 {object} DeleteEmployeeResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
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

// ListEmployee fetches the employee list
// @Tags Employee
// @Summary Get Employee List
// @Description
// @Description **Query Parameters:**
// @Description
// @Description page:
// @Description - Optional
// @Description - Page number for pagination
// @Description - Default: 1
// @Description
// @Description pagesize:
// @Description - Optional
// @Description - Number of employees per page
// @Description - Default: 50
// @Description
// @Description order:
// @Description - Optional
// @Description - Sort order (asc/desc)
// @Description - Default: desc
// @Description
// @Description **Header:**
// @Description
// @Description Authorization:
// @Description - Required
// @Description - Format: Bearer token
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token" default(Bearer )
// @Param page query int false "Page number for pagination" default(1)
// @Param pagesize query int false "Number of employee per page" default(50)
// @Param order query string false "Sort order (asc/desc)" default(desc)
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
