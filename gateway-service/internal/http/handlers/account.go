package handlers

import "gateway-service/internal/grpc/clients"

type AccountHandler struct {
	AccountClient clients.AccountClient
}

func NewAccountHandler(authClient clients.AccountClient) *AccountHandler {
	return &AccountHandler{
		AccountClient: authClient,
	}
}
