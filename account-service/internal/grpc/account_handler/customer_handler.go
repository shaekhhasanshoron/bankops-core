package handlers

import (
	protoacc "account-service/api/protogen/accountservice/proto"
	"account-service/internal/logging"
	"context"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CreateCustomer handles the creation of a new customer
func (h *AccountHandlerService) CreateCustomer(ctx context.Context, req *protoacc.CreateCustomerRequest) (*protoacc.CreateCustomerResponse, error) {
	customer, message, err := h.CreateCustomerService.Execute(req.GetName(), req.GetMetadata().GetRequester(), req.GetMetadata().GetRequestId())
	if err != nil {
		logging.Logger.Warn().Err(err).Str("customer", req.GetName()).Msg("create customer failed")
		return &protoacc.CreateCustomerResponse{
			CustomerId: "",
			Response: &protoacc.Response{
				Message: message,
				Success: false,
			},
		}, nil
	}

	return &protoacc.CreateCustomerResponse{
		CustomerId: customer.ID,
		Response: &protoacc.Response{
			Message: message,
			Success: true,
		},
	}, nil
}

func (h *AccountHandlerService) ListCustomers(ctx context.Context, req *protoacc.ListCustomersRequest) (*protoacc.ListCustomersResponse, error) {
	customers, totalCount, totalPage, message, err := h.ListCustomerService.Execute(int(req.GetPagination().GetPage()), int(req.GetPagination().GetPageSize()), req.GetSortOrder(), req.GetMetadata().GetRequestId())
	if err != nil {
		logging.Logger.Warn().Err(err).Msg("list customer failed")
		return &protoacc.ListCustomersResponse{
			Customers: nil,
			Pagination: &protoacc.PaginationResponse{
				Page:       req.GetPagination().GetPage(),
				PageSize:   req.GetPagination().GetPageSize(),
				TotalCount: 0,
				TotalPages: 0,
			},
			Response: &protoacc.Response{
				Message: message,
				Success: false,
			},
		}, nil
	}

	protoCustomers := make([]*protoacc.Customer, len(customers))
	for i, customer := range customers {
		accounts := make([]*protoacc.Account, len(customer.Accounts))
		for j, acc := range customer.Accounts {
			accounts[j] = &protoacc.Account{
				Id:           acc.ID,
				CustomerId:   acc.CustomerID,
				Balance:      acc.Balance,
				ActiveStatus: acc.ActiveStatus,
				CreatedAt:    timestamppb.New(acc.CreatedAt),
			}
		}

		protoCustomers[i] = &protoacc.Customer{
			Id:        customer.ID,
			Name:      customer.Name,
			Accounts:  accounts,
			CreatedAt: timestamppb.New(customer.CreatedAt),
			UpdatedAt: timestamppb.New(customer.UpdatedAt),
		}
	}

	return &protoacc.ListCustomersResponse{
		Customers: protoCustomers,
		Pagination: &protoacc.PaginationResponse{
			Page:       req.GetPagination().GetPage(),
			PageSize:   req.GetPagination().GetPageSize(),
			TotalCount: int32(totalCount),
			TotalPages: int32(totalPage),
		},
		Response: &protoacc.Response{
			Message: message,
			Success: true,
		},
	}, nil
}

func (h *AccountHandlerService) DeleteCustomer(ctx context.Context, req *protoacc.DeleteCustomerRequest) (*protoacc.DeleteCustomerResponse, error) {
	message, err := h.DeleteCustomerService.Execute(req.CustomerId, req.GetMetadata().GetRequester(), req.GetMetadata().GetRequestId())
	if err != nil {
		logging.Logger.Warn().Err(err).Str("customer_id", req.CustomerId).Msg("delete customer failed")
		return &protoacc.DeleteCustomerResponse{
			Response: &protoacc.Response{
				Message: message,
				Success: false,
			},
		}, nil
	}

	return &protoacc.DeleteCustomerResponse{
		Response: &protoacc.Response{
			Message: message,
			Success: true,
		},
	}, nil
}
