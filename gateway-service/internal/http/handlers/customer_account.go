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

type CreateAccountRequest struct {
	CustomerID    string  `json:"customer_id" binding:"required"`
	DepositAmount float64 `json:"deposit_amount" binding:"required"`
}

type CreateAccountResponse struct {
	AccountID string `json:"account_id" binding:"required"`
	Message   string `json:"message" binding:"message"`
}

type DeleteAccountResponse struct {
	Message string `json:"message" binding:"message"`
}

type ListAccountResponse struct {
	Accounts   interface{} `json:"accounts"`
	Page       int         `json:"page"`
	PageSize   int         `json:"pageSize"`
	TotalCount int         `json:"totalCount"`
	TotalPages int         `json:"totalPages"`
	Message    string      `json:"message" binding:"message"`
}

type GetBalanceResponse struct {
	Balance float64 `json:"balance"`
	Message string  `json:"message" binding:"message"`
}

// CreateAccount creates a new account
// @Tags Account
// @Summary Create Account
// @Description
// @Description **Request Body:**
// @Description
// @Description Customer ID:
// @Description - Required
// @Description
// @Description Deposit Amount:
// @Description - Required
// @Description - Must be greater than zero
// @Description
// @Description **Header:**
// @Description
// @Description Authorization:
// @Description - Required
// @Description - Format: Bearer token
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token" default(Bearer )
// @Param account body CreateAccountRequest true "Account details"
// @Success 201 {object} CreateAccountResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/account [post]
func (h *AccountHandler) CreateAccount(c *gin.Context) {
	var req CreateAccountRequest
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

	grpcReq := &protoacc.CreateAccountRequest{
		CustomerId:     req.CustomerID,
		InitialDeposit: req.DepositAmount,
		Metadata: &protoacc.Metadata{
			RequestId: c.GetHeader("X-Request-ID"),
			Requester: requester,
		},
	}

	resp, err := h.AccountClient.CreateAccount(c.Request.Context(), grpcReq)
	if err != nil || resp == nil {
		logging.Logger.Error().Err(err).Msg("failed to create new account")
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid credentials"})
		return
	}

	if !resp.Response.Success {
		logging.Logger.Error().Err(errors.New(resp.Response.Message)).Msg("unable to create account")
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: resp.Response.Message})
		return
	}

	res := CreateAccountResponse{
		AccountID: resp.AccountId,
		Message:   resp.Response.Message,
	}

	c.JSON(http.StatusCreated, res)
}

// DeleteAccount deletes a single account or all accounts under a customer
// @Tags Account
// @Summary Delete Account (single/all)
// @Description
// @Description **Query Parameters:**
// @Description
// @Description scope:
// @Description - Required
// @Description - Options: **single**, **all**
// @Description - Default: single
// @Description - **single**: Delete one account (id = account_id)
// @Description - **all**: Delete all accounts for customer (id = customer_id)
// @Description
// @Description id:
// @Description - Required
// @Description - AccountID (if scope=single) or CustomerID (if scope=all)
// @Description
// @Description **Header:**
// @Description
// @Description Authorization:
// @Description - Required
// @Description - Format: Bearer token
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token" default(Bearer )
// @Param scope query string true "Scope (single/all)" default(single)
// @Param id query string true "AccountID or CustomerID"
// @Success 200 {object} DeleteAccountResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/account [delete]
func (h *AccountHandler) DeleteAccount(c *gin.Context) {
	scope := strings.TrimSpace(c.Query("scope"))
	id := strings.TrimSpace(c.Query("id"))

	if scope == "" || id == "" {
		logging.Logger.Error().Err(errors.New("Missing required parameters (scope and id)")).
			Str("scope", scope).
			Str("id", id).
			Msg("Invalid request")

		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Missing required parameters (scope and id)"})
		return
	}

	un, _ := c.Get("username") // middleware
	requester, ok := un.(string)
	if !ok {
		logging.Logger.Warn().Err(errors.New("unable to get requester username")).Msg("requester: " + requester)
	}

	grpcReq := &protoacc.DeleteAccountRequest{
		Scope: scope,
		Id:    id,
		Metadata: &protoacc.Metadata{
			RequestId: c.GetHeader("X-Request-ID"),
			Requester: requester,
		},
	}

	resp, err := h.AccountClient.DeleteAccount(c.Request.Context(), grpcReq)
	if err != nil {
		logging.Logger.Error().Err(err).Msg("failed to delete account")
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
		return
	}

	if !resp.Response.Success {
		logging.Logger.Error().Err(errors.New(resp.Response.Message)).Msg("unable to delete account")
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: resp.Response.Message})
		return
	}

	res := DeleteCustomerResponse{
		Message: resp.Response.Message,
	}

	c.JSON(http.StatusOK, res)
}

// GetAccountBalance fetches the account balance
// @Tags Account
// @Summary Get Account balance
// @Description
// @Description **Path Parameter:**
// @Description
// @Description id:
// @Description - Required
// @Description - AccountID of a customer account
// @Description
// @Description **Header:**
// @Description
// @Description Authorization:
// @Description - Required
// @Description - Format: Bearer token
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token" default(Bearer )
// @Param id path string true "AccountID of a customer account"
// @Success 200 {object} GetBalanceResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/account/{id}/balance [get]
func (h *AccountHandler) GetAccountBalance(c *gin.Context) {
	accountId := strings.TrimSpace(c.Param("id"))

	un, _ := c.Get("username") // middleware
	requester, ok := un.(string)
	if !ok {
		logging.Logger.Warn().Err(errors.New("unable to get requester username")).Msg("requester: " + requester)
	}

	grpcReq := &protoacc.GetBalanceRequest{
		AccountId: accountId,
		Metadata: &protoacc.Metadata{
			RequestId: c.GetHeader("X-Request-ID"),
			Requester: requester,
		},
	}

	resp, err := h.AccountClient.GetBalance(c.Request.Context(), grpcReq)
	if err != nil {
		logging.Logger.Error().Err(err).Msg("failed to get account balance")
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
		return
	}

	if !resp.Response.Success {
		logging.Logger.Error().Err(errors.New(resp.Response.Message)).Msg("unable to get account balance")
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: resp.Response.Message})
		return
	}

	res := GetBalanceResponse{
		Balance: resp.Balance,
		Message: resp.Response.Message,
	}

	c.JSON(http.StatusOK, res)
}

// ListAccounts fetches the account list with optional filters
// @Tags Account
// @Summary Get Account List
// @Description
// @Description **Query Parameters:**
// @Description
// @Description customer_id:
// @Description - Optional
// @Description - Filter by Customer ID
// @Description
// @Description in_transaction:
// @Description - Optional
// @Description - Filter accounts currently in transaction
// @Description - Values: true/false
// @Description - Default: no filter applied
// @Description
// @Description min_balance:
// @Description - Optional
// @Description - Filter by minimum balance
// @Description
// @Description page:
// @Description - Optional
// @Description - Page number for pagination
// @Description - Default: 1
// @Description
// @Description pagesize:
// @Description - Optional
// @Description - Number of accounts per page
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
// @Param customer_id query string false "Customer ID"
// @Param in_transaction query string false "Account In Transaction (value: true/false; default won't affect the filter)"
// @Param min_balance query string false "Minimum balance"
// @Param page query int false "Page number for pagination" default(1)
// @Param pagesize query int false "Number of accounts per page" default(50)
// @Param order query string false "Sort order (asc/desc)" default(desc)
// @Success 200 {object} ListAccountResponse "Successfully retrieved account list"
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/account [get]
func (h *AccountHandler) ListAccounts(c *gin.Context) {
	pageStr := strings.TrimSpace(c.Query("page"))
	pageSizeStr := strings.TrimSpace(c.Query("pagesize"))
	order := strings.TrimSpace(c.Query("order"))
	customerID := strings.TrimSpace(c.Query("customer_id"))
	minBalanceStr := strings.TrimSpace(c.Query("min_balance"))
	inTransactionStr := strings.TrimSpace(c.Query("in_transaction"))

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
		CustomerId:    customerID,
		MinBalance:    minBalanceStr,
		InTransaction: inTransactionStr,
		SortOrder:     order,
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
