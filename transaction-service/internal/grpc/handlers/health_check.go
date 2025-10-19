package handlers

import (
	"context"
	"time"
	prototx "transaction-service/api/protogen/txservice/proto"
)

// HealthCheck checks ping request and replies it
func (h *TransactionHandlerService) HealthCheck(ctx context.Context, req *prototx.HealthCheckRequest) (*prototx.HealthCheckResponse, error) {
	if req.Message != "ping" {
		return &prototx.HealthCheckResponse{
			Message:   "invalid_request",
			Timestamp: time.Now().Unix(),
		}, nil
	}

	return &prototx.HealthCheckResponse{
		Message:   "pong",
		Timestamp: time.Now().Unix(),
	}, nil
}
