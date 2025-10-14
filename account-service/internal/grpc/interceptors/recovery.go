package interceptors

import (
	"account-service/internal/logging"
	"context"
	"fmt"
	"google.golang.org/grpc"
)

// RecoveryInterceptor handles panics in the gRPC server and returns a status error.
func RecoveryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	defer func() {
		if r := recover(); r != nil {
			logging.Logger.Error().Msg(fmt.Sprintf("panic occurred in gRPC method %s: %v", info.FullMethod, r))
		}
	}()

	return handler(ctx, req)
}
