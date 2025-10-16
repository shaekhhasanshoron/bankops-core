package handlers

import (
	"gateway-service/internal/ports"
)

type AccountHandler struct {
	AccountClient ports.AccountClient
}

func NewAccountHandler(authClient ports.AccountClient) *AccountHandler {
	return &AccountHandler{
		AccountClient: authClient,
	}
}
