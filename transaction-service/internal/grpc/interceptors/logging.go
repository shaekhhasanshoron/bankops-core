package interceptors

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"time"
	"transaction-service/internal/logging"
)

// LoggingInterceptor is a logging interceptor for gRPC.
func LoggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()

	resp, err := handler(ctx, req)

	md, _ := metadata.FromIncomingContext(ctx)
	logging.Logger.Info().
		Str("method", info.FullMethod).
		Dur("duration", time.Since(start)).
		Int("status_code", getStatusCode(err)).
		Str("request_id", fmt.Sprintf("%v", md["request_id"])).
		Msg("gRPC request completed")

	return resp, err
}

// Helper to extract status code from error
func getStatusCode(err error) int {
	if err == nil {
		return 0
	}
	return 1 // return status based on error code
}
