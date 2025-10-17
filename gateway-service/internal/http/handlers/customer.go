package handlers

import (
	"errors"
	protoacc "gateway-service/api/protogen/accountservice/proto"
	"gateway-service/internal/logging"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
)

type Customer struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type CreateCustomerRequest struct {
	Name string `json:"name" binding:"required"`
}

type CreateCustomerResponse struct {
	CustomerID string `json:"customer_id" binding:"required"`
	Message    string `json:"message" binding:"message"`
}

type UpdateCustomerRequest struct {
	Name string `json:"name" binding:"required"`
}

type UpdateCustomerResponse struct {
	Message string `json:"message" binding:"message"`
}

type DeleteCustomerResponse struct {
	Message string `json:"message" binding:"message"`
}

type GetCustomerResponse struct {
	Customer Customer `json:"customer"`
	Message  string   `json:"message" binding:"message"`
}

type ListCustomerResponse struct {
	Customers  interface{} `json:"customers"`
	Page       int         `json:"page"`
	PageSize   int         `json:"pageSize"`
	TotalCount int         `json:"totalCount"`
	TotalPages int         `json:"totalPages"`
	Message    string      `json:"message" binding:"message"`
}

// CreateCustomer for creating new customer
// @Tags Customer
// @Summary Create Customer
// @Description Create employee - Name: required | Bearer token required
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token for authorization, include 'Bearer ' followed by access_token"
// @Param customer body CreateCustomerRequest true "Customer details"
// @Success 201 {object} CreateCustomerResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/customer [post]
func (h *AccountHandler) CreateCustomer(c *gin.Context) {
	var req CreateCustomerRequest
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

	grpcReq := &protoacc.CreateCustomerRequest{
		Name: req.Name,
		Metadata: &protoacc.Metadata{
			RequestId: c.GetHeader("X-Request-ID"),
			Requester: requester,
		},
	}

	resp, err := h.AccountClient.CreateCustomer(c.Request.Context(), grpcReq)
	if err != nil || resp == nil {
		logging.Logger.Error().Err(err).Msg("failed to create new customer")
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid credentials"})
		return
	}

	if !resp.Response.Success {
		logging.Logger.Error().Err(errors.New(resp.Response.Message)).Msg("unable to create customer")
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: resp.Response.Message})
		return
	}

	res := CreateCustomerResponse{
		CustomerID: resp.CustomerId,
		Message:    resp.Response.Message,
	}

	c.JSON(http.StatusCreated, res)
}

// DeleteCustomer for delete a customer by customer id
// @Tags Customer
// @Summary Delete Customer
// @Description Delete a customer by customer id
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token for authorization, include 'Bearer ' followed by access_token"
// @Param id path string true "CustomerID of the customer"
// @Success 200 {string} {object} DeleteCustomerResponse
// @Failure 400 {string} {object} ErrorResponse
// @Failure 401 {string} {object} ErrorResponse
// @Router /api/v1/customer/{id} [delete]
func (h *AccountHandler) DeleteCustomer(c *gin.Context) {
	customerId := c.Param("id")

	un, _ := c.Get("username") // middleware
	requester, ok := un.(string)
	if !ok {
		logging.Logger.Warn().Err(errors.New("unable to get requester username")).Msg("requester: " + requester)
	}

	grpcReq := &protoacc.DeleteCustomerRequest{
		CustomerId: customerId,
		Metadata: &protoacc.Metadata{
			RequestId: c.GetHeader("X-Request-ID"),
			Requester: requester,
		},
	}

	resp, err := h.AccountClient.DeleteCustomer(c.Request.Context(), grpcReq)
	if err != nil {
		logging.Logger.Error().Err(err).Msg("failed to delete customer")
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
		return
	}

	if !resp.Response.Success {
		logging.Logger.Error().Err(errors.New(resp.Response.Message)).Msg("unable to delete customer")
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: resp.Response.Message})
		return
	}

	res := DeleteCustomerResponse{
		Message: resp.Response.Message,
	}

	c.JSON(http.StatusOK, res)
}

// ListCustomer for fetching a customer list
// @Tags Customer
// @Summary Get Customer List
// @Description Get customer list
// @Accept json
// @Produce json
// @Param page query int false "Page number for pagination" default(1)
// @Param pagesize query int false "Number of customers per page" default(50)
// @Param order query string false "Sort order (asc/desc)" default(desc)
// @Param Authorization header string true "Bearer token for authorization, include 'Bearer ' followed by access_token"
// @Success 200 {string} {object} ListCustomerResponse
// @Failure 400 {string} {object} ErrorResponse
// @Failure 401 {string} {object} ErrorResponse
// @Router /api/v1/customer [get]
func (h *AccountHandler) ListCustomer(c *gin.Context) {
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

	grpcReq := &protoacc.ListCustomersRequest{
		SortOrder: order,
		Pagination: &protoacc.PaginationRequest{
			Page:     int32(pageNo),
			PageSize: int32(pageSize),
		},
		Metadata: &protoacc.Metadata{
			RequestId: c.GetHeader("X-Request-ID"),
			Requester: requester,
		},
	}

	resp, err := h.AccountClient.ListCustomer(c.Request.Context(), grpcReq)
	if err != nil {
		logging.Logger.Error().Err(err).Msg("failed to get customer")
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
		return
	}

	if !resp.Response.Success {
		logging.Logger.Error().Err(errors.New(resp.Response.Message)).Msg("unable to get customer")
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: resp.Response.Message})
		return
	}

	res := ListCustomerResponse{
		Customers:  resp.Customers,
		Page:       int(resp.Pagination.Page),
		PageSize:   int(resp.Pagination.PageSize),
		TotalCount: int(resp.Pagination.TotalCount),
		TotalPages: int(resp.Pagination.TotalPages),
		Message:    resp.Response.Message,
	}

	c.JSON(http.StatusOK, res)
}

// ListCustomerAccounts for fetching account list of customer
// @Tags Customer
// @Summary Get Account List of customer
// @Description Get account list of customers
// @Accept json
// @Produce json
// @Param id path string true "CustomerID of the customer"
// @Param page query int false "Page number for pagination" default(1)
// @Param pagesize query int false "Number of accounts per page" default(50)
// @Param Authorization header string true "Bearer token for authorization, include 'Bearer ' followed by access_token"
// @Success 200 {object} ListAccountResponse "Successfully retrieved account list"
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/customer/{id}/account [get]
func (h *AccountHandler) ListCustomerAccounts(c *gin.Context) {
	customerId := c.Param("id")
	pageStr := c.Query("page")
	pageSizeStr := c.Query("pagesize")

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

	grpcReq := &protoacc.ListAccountsRequest{
		CustomerId: customerId,
		Metadata: &protoacc.Metadata{
			RequestId: c.GetHeader("X-Request-ID"),
			Requester: requester,
		},
		Pagination: &protoacc.PaginationRequest{
			Page:     int32(pageNo),
			PageSize: int32(pageSize),
		},
	}

	resp, err := h.AccountClient.ListAccount(c.Request.Context(), grpcReq)
	if err != nil {
		logging.Logger.Error().Err(err).Msg("failed to get accounts")
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
		return
	}

	if !resp.Response.Success {
		logging.Logger.Error().Err(errors.New(resp.Response.Message)).Msg("unable to get accounts")
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: resp.Response.Message})
		return
	}

	res := ListAccountResponse{
		Accounts:   resp.Accounts,
		Page:       int(resp.Pagination.Page),
		PageSize:   int(resp.Pagination.PageSize),
		TotalCount: int(resp.Pagination.TotalCount),
		TotalPages: int(resp.Pagination.TotalPages),
		Message:    resp.Response.Message,
	}

	c.JSON(http.StatusOK, res)
}
