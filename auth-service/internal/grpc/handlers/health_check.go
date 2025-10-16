package handlers

import (
	"auth-service/api/protogen/authservice/proto"
	"context"
	"time"
)

// HealthCheck checks ping request and replies it
func (h *AuthHandler) HealthCheck(ctx context.Context, req *proto.HealthCheckRequest) (*proto.HealthCheckResponse, error) {
	if req.Message != "ping" {
		return &proto.HealthCheckResponse{
			Message:   "invalid_request",
			Timestamp: time.Now().Unix(),
		}, nil
	}

	return &proto.HealthCheckResponse{
		Message:   "pong",
		Timestamp: time.Now().Unix(),
	}, nil
}
