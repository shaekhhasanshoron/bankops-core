package grpc

import (
	protoacc "account-service/api/protogen/accountservice/proto"
	appcustomer "account-service/internal/app/customer"
	"account-service/internal/config"
	handlers "account-service/internal/grpc/account_handler"
	"account-service/internal/grpc/interceptors"
	"account-service/internal/logging"
	"account-service/internal/ports"
	"fmt"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
)

type ServiceRepos struct {
	CustomerRepo    ports.CustomerRepo
	AccountRepo     ports.AccountRepo
	TransactionRepo ports.TransactionRepo
	EventRepo       ports.EventRepo
}

func StartGRPCServer(repos ServiceRepos) {
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
	protoacc.RegisterAccountServiceServer(grpcServer, generateAggregatedHandlers(repos))

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

func generateAggregatedHandlers(repos ServiceRepos) *handlers.AccountHandlerService {
	accountAggregatedHandler := handlers.NewAggregatedHandler()
	accountAggregatedHandler.CreateCustomerService = appcustomer.NewCreateCustomer(repos.CustomerRepo, repos.EventRepo)
	accountAggregatedHandler.ListCustomerService = appcustomer.NewListCustomer(repos.CustomerRepo, repos.EventRepo)
	accountAggregatedHandler.DeleteCustomerService = appcustomer.NewDeleteCustomer(repos.CustomerRepo, repos.EventRepo)
	return accountAggregatedHandler
}
