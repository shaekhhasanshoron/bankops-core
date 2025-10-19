package handlers

import (
	protoacc "account-service/api/protogen/accountservice/proto"
	"account-service/internal/grpc/types"
	"context"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (h *AccountHandlerService) ValidateAccounts(ctx context.Context, req *protoacc.ValidateAccountsRequest) (*protoacc.ValidateAccountsResponse, error) {
	accounts, message, err := h.ValidateAccountForTransactionService.Execute(
		req.TransactionId,
		req.AccountIds,
		req.GetMetadata().GetRequester(),
		req.GetMetadata().GetRequestId(),
	)

	if err != nil {
		return &protoacc.ValidateAccountsResponse{
			Valid:   false,
			Message: message,
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
			Version:      int32(account.Version),
			ActiveStatus: account.ActiveStatus,
			CreatedAt:    timestamppb.New(account.CreatedAt),
		}
	}

	return &protoacc.ValidateAccountsResponse{
		Valid:    true,
		Accounts: protoAccounts,
		Message:  message,
		Response: &protoacc.Response{
			Message: message,
			Success: true,
		},
	}, nil
}

func (h *AccountHandlerService) LockAccounts(ctx context.Context, req *protoacc.LockAccountsRequest) (*protoacc.LockAccountsResponse, error) {
	message, err := h.LockAccountForTransaction.Execute(
		req.TransactionId,
		req.AccountIds,
		req.GetMetadata().GetRequester(),
		req.GetMetadata().GetRequestId(),
	)

	if err != nil {
		return &protoacc.LockAccountsResponse{
			Locked:  false,
			Message: message,
			Response: &protoacc.Response{
				Message: message,
				Success: false,
			},
		}, nil
	}

	return &protoacc.LockAccountsResponse{
		Locked:  true,
		Message: "accounts locked successfully",
		Response: &protoacc.Response{
			Message: "accounts locked successfully",
			Success: true,
		},
	}, nil
}

func (h *AccountHandlerService) UnlockAccounts(ctx context.Context, req *protoacc.UnlockAccountsRequest) (*protoacc.UnlockAccountsResponse, error) {
	message, err := h.UnlockAccountsForTransaction.Execute(
		req.TransactionId,
		req.GetMetadata().GetRequester(),
		req.GetMetadata().GetRequestId(),
	)

	if err != nil {
		return &protoacc.UnlockAccountsResponse{
			Unlocked: false,
			Message:  message,
			Response: &protoacc.Response{
				Message: message,
				Success: false,
			},
		}, nil
	}

	return &protoacc.UnlockAccountsResponse{
		Unlocked: true,
		Message:  message,
		Response: &protoacc.Response{
			Message: message,
			Success: true,
		},
	}, nil
}

func (h *AccountHandlerService) UpdateAccountsBalance(ctx context.Context, req *protoacc.UpdateAccountsBalanceRequest) (*protoacc.UpdateAccountsBalanceResponse, error) {
	if len(req.Updates) == 0 {
		return &protoacc.UpdateAccountsBalanceResponse{
			Success: false,
			Message: "at least one balance update is required",
			Response: &protoacc.Response{
				Message: "at least one balance update is required",
				Success: false,
			},
		}, nil
	}

	var accountUpdateBalanceDetails []types.AccountBalance
	for _, update := range req.Updates {
		accountUpdateBalanceDetails = append(accountUpdateBalanceDetails, types.AccountBalance{
			AccountID: update.AccountId,
			Balance:   update.NewBalance,
			Version:   int(update.Version),
		})
	}

	accounts, message, err := h.UpdateAccountBalanceForTransaction.Execute(
		accountUpdateBalanceDetails,
		req.GetMetadata().Requester,
	)

	if err != nil {
		return &protoacc.UpdateAccountsBalanceResponse{
			Success: true,
			Message: message,
			Response: &protoacc.Response{
				Message: message,
				Success: false,
			},
		}, nil
	}

	var newVersions []*protoacc.AccountVersion
	if accounts != nil {
		for _, account := range accounts {
			newVersions = append(newVersions, &protoacc.AccountVersion{
				AccountId: account.AccountID,
				Version:   int32(account.Version + 1),
			})
		}
	}

	return &protoacc.UpdateAccountsBalanceResponse{
		Success:     true,
		Message:     message,
		NewVersions: newVersions,
		Response: &protoacc.Response{
			Message: message,
			Success: true,
		},
	}, nil
}
