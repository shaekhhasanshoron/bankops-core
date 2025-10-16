package handlers

import (
	protoacc "account-service/api/protogen/accountservice/proto"
	"context"
	"time"
)

// HealthCheck checks ping request and replies it
func (h *AccountHandlerService) HealthCheck(ctx context.Context, req *protoacc.HealthCheckRequest) (*protoacc.HealthCheckResponse, error) {
	if req.Message != "ping" {
		return &protoacc.HealthCheckResponse{
			Message:   "invalid_request",
			Timestamp: time.Now().Unix(),
		}, nil
	}

	return &protoacc.HealthCheckResponse{
		Message:   "pong",
		Timestamp: time.Now().Unix(),
	}, nil
}
