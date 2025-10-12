package interceptors

import (
	"auth-service/internal/observability/tracing"
	"context"
	"fmt"
	"go.opentelemetry.io/otel/codes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"time"
)

// TracingInterceptor adds tracing information to gRPC requests
func TracingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Start the tracing span for the incoming gRPC method
	span, ctx := tracing.StartSpan(ctx, info.FullMethod)
	defer tracing.EndSpan(span)

	start := time.Now()

	// Proceed with handling the request
	resp, err := handler(ctx, req)

	// Add tracing information (e.g., peer info, request_id, duration, etc.)
	md, _ := metadata.FromIncomingContext(ctx)
	p, _ := peer.FromContext(ctx)

	// Add request metadata and peer address to the span for better trace context
	tracing.AddAttributesToSpan(span, map[string]string{
		"request_id": fmt.Sprintf("%v", md["request_id"]),
		"peer":       p.Addr.String(),
		"duration":   fmt.Sprintf("%v", time.Since(start)),
	})

	// If the request failed, mark it in the span (useful for troubleshooting)
	if err != nil {
		span.SetStatus(codes.Error, fmt.Sprintf("Failed to process request: %v", err))
	}

	// Return the response and any error
	return resp, err
}
