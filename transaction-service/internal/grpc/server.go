package grpc

import (
	"context"
	"fmt"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	prototx "transaction-service/api/protogen/txservice/proto"
	apptx "transaction-service/internal/app"
	"transaction-service/internal/config"
	"transaction-service/internal/grpc/handlers"
	"transaction-service/internal/grpc/interceptors"
	"transaction-service/internal/logging"
	"transaction-service/internal/ports"
)

type ServiceRepos struct {
	AccountClient   ports.AccountClient
	SagaRepo        ports.SagaRepo
	TransactionRepo ports.TransactionRepo
	EventRepo       ports.EventRepo
}

func StartGRPCServer(ctx context.Context, repos ServiceRepos) {
	var unaryInterceptors []grpc.UnaryServerInterceptor

	if config.Current().Observability.MetricsConfig.Enabled {
		unaryInterceptors = append(unaryInterceptors, interceptors.MetricsInterceptor)
	}

	if config.Current().Observability.TracingConfig.Enabled {
		unaryInterceptors = append(unaryInterceptors, interceptors.TracingInterceptor)
	}

	unaryInterceptors = append(unaryInterceptors, interceptors.RecoveryInterceptor, interceptors.LoggingInterceptor)
	var options []grpc.ServerOption
	if len(unaryInterceptors) > 0 {
		options = append(options, grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(unaryInterceptors...),
		))
	}
	grpcServer := grpc.NewServer(options...)

	// Register gRPC services
	prototx.RegisterTransactionServiceServer(grpcServer, generateAggregatedHandlers(repos))

	// Register reflection for gRPC introspection (optional, useful for testing)
	reflection.Register(grpcServer)

	// Listen on the configured address
	listener, err := net.Listen("tcp", config.Current().GRPC.Addr)
	if err != nil {
		logging.Logger.Fatal().Err(err).Msg("Failed to listen on TCP address")
	}

	logging.Logger.Info().Msg(fmt.Sprintf("gRPC server listening on %s", listener.Addr().String()))

	go func() {
		<-ctx.Done()
		logging.Logger.Info().Msg("Shutting down gRPC server...")
		grpcServer.GracefulStop()
	}()

	// Start the server
	if err := grpcServer.Serve(listener); err != nil {
		logging.Logger.Fatal().Err(err).Msg("Failed to serve gRPC server")
	}
}

func generateAggregatedHandlers(repos ServiceRepos) *handlers.TransactionHandlerService {
	accountAggregatedHandler := handlers.NewAggregatedHandler()
	accountAggregatedHandler.InitTransactionService = apptx.NewInitTransaction(
		repos.TransactionRepo,
		repos.AccountClient,
		repos.SagaRepo,
		repos.EventRepo,
	)

	accountAggregatedHandler.GetTransactionHistoryService = apptx.NewGetTransactionHistory(
		repos.TransactionRepo,
	)

	return accountAggregatedHandler
}
