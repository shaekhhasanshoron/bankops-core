package handlers

import (
	"errors"
	protoacc "gateway-service/api/protogen/accountservice/proto"
	"gateway-service/internal/logging"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
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

// CreateAccount for creating new customer
// @Tags Account
// @Summary Create Account
// @Description Create account - Bearer token required
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token for authorization, include 'Bearer ' followed by access_token"
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

// DeleteAccount for delete account by scope (scope=single; id = account_id / scope=all; id = customer_id)
// @Tags Account
// @Summary Delete Account (single/all)
// @Description Delete an account or all accounts for a customer by scope (scope=single; id = account_id / scope=all; id = customer_id)
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token for authorization, include 'Bearer ' followed by access_token"
// @Param scope query string true "Scope (single/all)"
// @Param id query string true "AccountID or CustomerID"
// @Success 200 {string} {object} DeleteAccountResponse
// @Failure 400 {string} {object} ErrorResponse
// @Failure 401 {string} {object} ErrorResponse
// @Router /api/v1/account [delete]
func (h *AccountHandler) DeleteAccount(c *gin.Context) {
	scope := c.Query("scope")
	id := c.Query("id")

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

// GetAccountBalance for fetching account balance
// @Tags Account
// @Summary Get Account balance
// @Description Get account balance
// @Accept json
// @Produce json
// @Param id path string true "AccountID of a customer account"
// @Param Authorization header string true "Bearer token for authorization, include 'Bearer ' followed by access_token"
// @Success 200 {object} GetBalanceResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/account/{id}/balance [get]
func (h *AccountHandler) GetAccountBalance(c *gin.Context) {
	accountId := c.Param("id")

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

// ListAccounts for fetching account list
// @Tags Account
// @Summary Get Account List
// @Description Get account list based on scopes. scopes is optional, value of scope (customer/in_transaction/has_balance).
// if scope is customer then provide customer_id in query, if scope=has_balance provide min_balance in query
// @Accept json
// @Produce json
// @Param scopes query string false "Scope filter (comma separated) (customer,in_transaction,has_balance)" default(all)
// @Param customer_id query string false "Customer ID (required when adding 'customer' to scopes)"
// @Param min_balance query number false "Minimum balance (required when adding 'has_balance' to scopes)"
// @Param page query int false "Page number for pagination" default(1)
// @Param pagesize query int false "Number of accounts per page" default(50)
// @Param Authorization header string true "Bearer token for authorization, include 'Bearer ' followed by access_token"
// @Success 200 {object} ListAccountResponse "Successfully retrieved account list"
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/account [get]
func (h *AccountHandler) ListAccounts(c *gin.Context) {
	pageStr := c.Query("page")
	pageSizeStr := c.Query("pagesize")
	scopes := c.Query("scopes")
	customerID := c.Query("customer_id")
	minBalanceStr := c.Query("min_balance")

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

	var minBalance int
	if minBalanceStr != "" {
		minBalance, err = strconv.Atoi(minBalanceStr)
		if err != nil {
			logging.Logger.Warn().Err(err).Msg("Invalid minimum balance: " + minBalanceStr)
			minBalance = -1
		}
	}

	un, _ := c.Get("username") // middleware
	requester, ok := un.(string)
	if !ok {
		logging.Logger.Warn().Err(errors.New("unable to get requester username")).Msg("requester: " + requester)
	}

	grpcReq := &protoacc.ListAccountsRequest{
		Scopes:     scopes,
		CustomerId: customerID,
		MinBalance: float64(minBalance),
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
