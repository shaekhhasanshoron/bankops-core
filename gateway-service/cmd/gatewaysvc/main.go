package main

import (
	"context"
	"gateway-service/internal/adapter/grpc/clients"
	clients2 "gateway-service/internal/adapter/grpc/clients"
	"gateway-service/internal/config"
	"gateway-service/internal/http"
	"gateway-service/internal/logging"
	"gateway-service/internal/observability/tracing"
	"log"
	"os"
	"time"
)

func main() {
	// Loading env variables and configurations
	_, err := config.LoadConfig()
	if err != nil {
		_, _ = os.Stderr.WriteString("failed to load config: " + err.Error() + "\n")
		os.Exit(1)
	}

	// Initiating Logger
	if err = logging.InitiateLogger(); err != nil {
		_, _ = os.Stderr.WriteString("failed to initiate logger: " + err.Error() + "\n")
		os.Exit(1)
	}

	// Initialize tracing
	traceShutdown, err := tracing.Init(context.Background(), config.ServiceName)
	if err != nil {
		log.Fatalf("failed to initialize tracing: %v", err)
	}
	defer func() {
		if config.Current().Observability.TracingConfig.Enabled {
			_ = traceShutdown(context.Background())
		}
	}()

	// Initialize auth client
	authClient := clients2.NewAuthClient(30*time.Second, config.Current().GRPC.AuthServiceAddr)
	if err := authClient.Connect(); err != nil {
		logging.Logger.Fatal().Err(err).Msg("failed to connect to auth service")
	}
	defer authClient.Close()

	accountClient := clients.NewAccountClient(30*time.Second, config.Current().GRPC.AccountServiceAddr)
	if err := accountClient.Connect(); err != nil {
		logging.Logger.Fatal().Err(err).Msg("failed to connect to account service")
	}
	defer accountClient.Close()

	transactionClient := clients.NewTransactionClient(30*time.Second, config.Current().GRPC.TransactionServiceAddr)
	if err := accountClient.Connect(); err != nil {
		logging.Logger.Fatal().Err(err).Msg("failed to connect to transaction service")
	}
	defer transactionClient.Close()

	// Start connection monitor
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go accountClient.StartConnectionMonitor(ctx)
	go authClient.StartConnectionMonitor(ctx)
	go transactionClient.StartConnectionMonitor(ctx)

	http.StartServer(http.GrpcClients{
		AuthClient:        &authClient,
		AccountClient:     &accountClient,
		TransactionClient: &transactionClient,
	})
}
