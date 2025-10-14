package handlers

import (
	protoacc "account-service/api/protogen/accountservice/proto"
	"account-service/internal/logging"
	"context"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *AccountHandlerService) CreateAccount(ctx context.Context, req *protoacc.CreateAccountRequest) (*protoacc.CreateAccountResponse, error) {
	account, message, err := s.CreateAccountService.Execute(req.CustomerId, req.InitialDeposit, req.GetMetadata().GetRequester(), req.GetMetadata().GetRequestId())
	if err != nil {
		logging.Logger.Warn().Err(err).Str("customer_id", req.CustomerId).Msg("create account failed")
		return &protoacc.CreateAccountResponse{
			AccountId: "",
			Response: &protoacc.Response{
				Message: message,
				Success: false,
			},
		}, nil
	}

	return &protoacc.CreateAccountResponse{
		AccountId: account.ID,
		Response: &protoacc.Response{
			Message: message,
			Success: true,
		},
	}, nil
}

func (s *AccountHandlerService) DeleteAccount(ctx context.Context, req *protoacc.DeleteAccountRequest) (*protoacc.DeleteAccountResponse, error) {
	message, err := s.DeleteAccountService.Execute(req.Scope, req.Id, req.GetMetadata().GetRequester(), req.GetMetadata().GetRequestId())
	if err != nil {
		logging.Logger.Warn().Err(err).Str("scope", req.Scope).Str("id", req.Id).Msg("delete account failed")
		return &protoacc.DeleteAccountResponse{
			Response: &protoacc.Response{
				Message: message,
				Success: false,
			},
		}, nil
	}

	return &protoacc.DeleteAccountResponse{
		Response: &protoacc.Response{
			Message: message,
			Success: true,
		},
	}, nil
}

func (s *AccountHandlerService) GetBalance(ctx context.Context, req *protoacc.GetBalanceRequest) (*protoacc.GetBalanceResponse, error) {
	amount, message, err := s.GetAccountBalanceService.Execute(req.AccountId, req.GetMetadata().GetRequester(), req.GetMetadata().GetRequestId())
	if err != nil {
		logging.Logger.Warn().Err(err).Str("account_id", req.AccountId).Msg("get account failed")
		return &protoacc.GetBalanceResponse{
			Balance: 0,
			Response: &protoacc.Response{
				Message: message,
				Success: false,
			},
		}, nil
	}

	return &protoacc.GetBalanceResponse{
		Balance: amount,
		Response: &protoacc.Response{
			Message: message,
			Success: true,
		},
	}, nil
}

func (h *AccountHandlerService) ListAccount(ctx context.Context, req *protoacc.ListAccountsRequest) (*protoacc.ListAccountsResponse, error) {
	accounts, totalCount, totalPages, message, err := h.ListAccountService.Execute(req.Scopes, req.CustomerId, req.MinBalance, int(req.GetPagination().GetPage()), int(req.GetPagination().GetPageSize()), req.GetMetadata().GetRequester(), req.GetMetadata().GetRequestId())
	if err != nil {
		logging.Logger.Warn().Err(err).Msg("list customer failed")
		return &protoacc.ListAccountsResponse{
			Accounts: nil,
			Pagination: &protoacc.PaginationResponse{
				Page:       req.GetPagination().GetPage(),
				PageSize:   req.GetPagination().GetPageSize(),
				TotalCount: 0,
			},
			Response: &protoacc.Response{
				Message: message,
				Success: false,
			},
		}, nil
	}

	protoAccounts := make([]*protoacc.Account, len(accounts))
	for i, account := range accounts {
		protoAccounts[i] = &protoacc.Account{
			Id:           account.ID,
			CustomerId:   account.CustomerID,
			Balance:      account.Balance,
			ActiveStatus: account.ActiveStatus,
			CreatedAt:    timestamppb.New(account.CreatedAt),
		}
	}

	return &protoacc.ListAccountsResponse{
		Accounts: protoAccounts,
		Pagination: &protoacc.PaginationResponse{
			Page:       req.GetPagination().GetPage(),
			PageSize:   req.GetPagination().GetPageSize(),
			TotalCount: int32(totalCount),
			TotalPages: int32(totalPages),
		},
		Response: &protoacc.Response{
			Message: message,
			Success: true,
		},
	}, nil
}
