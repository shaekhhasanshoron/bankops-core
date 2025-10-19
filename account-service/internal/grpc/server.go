package grpc

import (
	protoacc "account-service/api/protogen/accountservice/proto"
	appaccount "account-service/internal/app/account"
	appcustomer "account-service/internal/app/customer"
	apptxsaga "account-service/internal/app/transaction_saga"
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
	CustomerRepo ports.CustomerRepo
	AccountRepo  ports.AccountRepo
	EventRepo    ports.EventRepo
}

func StartGRPCServer(repos ServiceRepos) {
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
	accountAggregatedHandler.ListCustomerService = appcustomer.NewListCustomer(repos.CustomerRepo)
	accountAggregatedHandler.DeleteCustomerService = appcustomer.NewDeleteCustomer(repos.CustomerRepo, repos.EventRepo)
	accountAggregatedHandler.CreateAccountService = appaccount.NewCreateAccount(repos.AccountRepo, repos.CustomerRepo, repos.EventRepo)
	accountAggregatedHandler.DeleteAccountService = appaccount.NewDeleteAccount(repos.AccountRepo, repos.CustomerRepo, repos.EventRepo)
	accountAggregatedHandler.GetAccountBalanceService = appaccount.NewGetAccountBalance(repos.AccountRepo)
	accountAggregatedHandler.ListAccountService = appaccount.NewListAccount(repos.AccountRepo)
	accountAggregatedHandler.ValidateAccountForTransactionService = apptxsaga.NewValidateAccountForTransaction(repos.AccountRepo)
	accountAggregatedHandler.LockAccountForTransaction = apptxsaga.NewLockAccountForTransaction(repos.AccountRepo)
	accountAggregatedHandler.UnlockAccountsForTransaction = apptxsaga.NewUnlockAccountsForTransaction(repos.AccountRepo)
	accountAggregatedHandler.UpdateAccountBalanceForTransaction = apptxsaga.NewUpdateAccountBalanceForTransaction(repos.AccountRepo)
	return accountAggregatedHandler
}
