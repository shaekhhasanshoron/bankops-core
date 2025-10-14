package interceptors

import (
	"auth-service/internal/observability/metrics"
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"time"
)

// MetricsInterceptor is an interceptor for recording metrics for gRPC calls.
func MetricsInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()

	resp, err := handler(ctx, req)

	// Record metrics
	statusCode := "success"
	if err != nil {
		if st, ok := status.FromError(err); ok {
			statusCode = st.Code().String()
		} else {
			statusCode = "unknown"
		}
	}

	metrics.ObserveGRPC(info.FullMethod, statusCode, time.Since(start))

	return resp, err
}
