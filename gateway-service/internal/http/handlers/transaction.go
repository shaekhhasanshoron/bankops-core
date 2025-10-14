package handlers

import (
	"errors"
	protoacc "gateway-service/api/protogen/accountservice/proto"
	"gateway-service/internal/logging"
	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/types/known/timestamppb"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type InitTransactionRequest struct {
	SourceAccountID      string  `json:"source_account_id" binding:"required"`
	DestinationAccountID string  `json:"destination_account_id"`
	TransactionType      string  `json:"transaction_type" binding:"required"`
	Amount               float64 `json:"amount"`
	Reference            string  `json:"reference" binding:"required"`
}

type InitTransactionResponse struct {
	TransactionID     string `json:"transaction_id"`
	TransactionStatus string `json:"transaction_status"`
	Message           string `json:"message" binding:"message"`
}

type ListTransactionResponse struct {
	Transactions interface{} `json:"transactions"`
	Page         int         `json:"page"`
	PageSize     int         `json:"pageSize"`
	TotalCount   int         `json:"totalCount"`
	TotalPages   int         `json:"totalPages"`
	Message      string      `json:"message" binding:"message"`
}

// InitTransaction for initiating new transaction
// @Tags Transaction
// @Summary Create new transaction
// @Description Create Transaction - Bearer token required. Transaction type can be: transfer/withdraw_full/withdraw_amount/add_amount. Amount must be greater than zero for all transaction types except 'withdraw_full'. Destination account ID is only required when 'transaction_type=transfer'. Reference is required for all transactions.
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token for authorization, include 'Bearer ' followed by access_token"
// @Param transaction body InitTransactionRequest true "Transaction details"
// @Success 201 {object} InitTransactionResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/transaction/init [post]
func (h *AccountHandler) InitTransaction(c *gin.Context) {
	var req InitTransactionRequest
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

	grpcReq := &protoacc.InitTransactionRequest{
		SourceAccountId:      req.SourceAccountID,
		DestinationAccountId: req.DestinationAccountID,
		Amount:               req.Amount,
		Type:                 req.TransactionType,
		Reference:            req.Reference,
		Metadata: &protoacc.Metadata{
			RequestId: c.GetHeader("X-Request-ID"),
			Requester: requester,
		},
	}

	resp, err := h.AccountClient.InitTransaction(c.Request.Context(), grpcReq)
	if err != nil || resp == nil {
		logging.Logger.Error().Err(err).Msg("failed to init transaction")
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid request"})
		return
	}

	if !resp.Response.Success {
		logging.Logger.Error().Err(errors.New(resp.Response.Message)).Msg("unable to initialize transaction")
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: resp.Response.Message})
		return
	}

	res := InitTransactionResponse{
		TransactionID:     resp.TransactionId,
		TransactionStatus: resp.TransactionStatus,
		Message:           resp.Response.Message,
	}

	c.JSON(http.StatusCreated, res)
}

// ListTransactions for fetching transaction history
// @Tags Transaction
// @Summary Get Transaction History
// @Description Get transaction history for all accounts with optional filtering
// @Accept json
// @Produce json
// @Param account_id query string false "Transaction history by account id"
// @Param customer_id query string false "Transaction history by customer id"
// @Param types query string false "Comma separated transaction types (transfer/withdraw_full/withdraw_amount/add_amount)"
// @Param start_date query string false "Start date for filtering (format: DD-MM-YYYY)"
// @Param end_date query string false "End date for filtering (format: DD-MM-YYYY)"
// @Param page query int false "Page number for pagination" default(1)
// @Param pagesize query int false "Number of transactions per page" default(50)
// @Param order query string false "Sort order (asc/desc)" default(desc)
// @Param Authorization header string true "Bearer token for authorization, include 'Bearer ' followed by access_token"
// @Success 200 {object} ListTransactionResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/transaction [get]
func (h *AccountHandler) ListTransactions(c *gin.Context) {
	pageStr := c.Query("page")
	pageSizeStr := c.Query("pagesize")
	order := c.Query("order")
	types := c.Query("types")
	accountId := c.Query("account_id")
	customerId := c.Query("customer_id")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	var startDate, endDate *timestamppb.Timestamp
	if startDateStr != "" {
		if t, err := time.Parse("02-01-2006", startDateStr); err == nil {
			startDate = timestamppb.New(t)
		} else {
			logging.Logger.Warn().Err(err).Msg("Invalid start_date format, expected DD-MM-YYYY: " + startDateStr)
		}
	}

	if endDateStr != "" {
		if t, err := time.Parse("02-01-2006", endDateStr); err == nil {
			// Set to end of day for end_date
			t = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			endDate = timestamppb.New(t)
		} else {
			logging.Logger.Warn().Err(err).Msg("Invalid end_date format, expected DD-MM-YYYY: " + endDateStr)
		}
	}

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

	grpcReq := &protoacc.GetTransactionHistoryRequest{
		AccountId: accountId,
		CompanyId: customerId,
		StartDate: startDate,
		EndDate:   endDate,
		SortOrder: order,
		Types:     strings.TrimSpace(types),
		Pagination: &protoacc.PaginationRequest{
			Page:     int32(pageNo),
			PageSize: int32(pageSize),
		},
		Metadata: &protoacc.Metadata{
			RequestId: c.GetHeader("X-Request-ID"),
			Requester: requester,
		},
	}

	resp, err := h.AccountClient.GetTransactionHistory(c.Request.Context(), grpcReq)
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

	res := ListTransactionResponse{
		Transactions: resp.Transactions,
		Page:         int(resp.Pagination.Page),
		PageSize:     int(resp.Pagination.PageSize),
		TotalCount:   int(resp.Pagination.TotalCount),
		TotalPages:   int(resp.Pagination.TotalPages),
		Message:      resp.Response.Message,
	}

	c.JSON(http.StatusOK, res)
}
