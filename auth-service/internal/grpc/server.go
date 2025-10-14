package grpc

import (
	"auth-service/api/protogen/authservice/proto"
	"auth-service/internal/app"
	"auth-service/internal/auth"
	"auth-service/internal/config"
	"auth-service/internal/grpc/handlers"
	"auth-service/internal/grpc/interceptors"
	"auth-service/internal/logging"
	"auth-service/internal/ports"
	"fmt"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
)

func StartGRPCServer(employeeRepo ports.EmployeeRepo, tokenSigner *auth.TokenSigner) {
	//// Create gRPC server options including interceptors
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				interceptors.MetricsInterceptor,
				interceptors.TracingInterceptor,
				interceptors.LoggingInterceptor,
				interceptors.RecoveryInterceptor,
			),
		),
	)

	// Register gRPC services
	authHandler := handlers.NewAuthHandler(
		app.NewAuthenticate(employeeRepo, tokenSigner),
		app.NewCreateEmployee(employeeRepo, tokenSigner),
		app.NewUpdateEmployee(employeeRepo),
		app.NewDeleteEmployee(employeeRepo),
	)

	proto.RegisterAuthServiceServer(grpcServer, authHandler)

	// Register reflection for gRPC introspection (optional, useful for testing)
	reflection.Register(grpcServer)

	// Listen on the configured address
	listener, err := net.Listen("tcp", config.Current().GRPC.Addr)
	if err != nil {
		logging.Logger.Fatal().Err(err).Msg("Failed to listen on TCP address")
	}

	logging.Logger.Info().Msg(fmt.Sprintf("gRPC server listening on %s", listener.Addr().String()))

	// Start the server
	if err := grpcServer.Serve(listener); err != nil {
		logging.Logger.Fatal().Err(err).Msg("Failed to serve gRPC server")
	}

}
