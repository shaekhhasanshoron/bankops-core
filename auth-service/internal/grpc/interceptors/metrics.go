package interceptors

import (
	"auth-service/internal/observability/metrics"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"time"
)

// MetricsInterceptor is an interceptor for recording metrics for gRPC calls.
func MetricsInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()

	resp, err := handler(ctx, req)

	// Record metrics
	md, _ := metadata.FromIncomingContext(ctx)
	metrics.ObserveHTTP(info.FullMethod, fmt.Sprintf("%v", md["request_id"]), getStatusCode(err), time.Since(start))

	return resp, err
}
