package main

import (
	"context"
	"gateway-service/internal/config"
	"gateway-service/internal/grpc/clients"
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
	authClient := clients.NewAuthClient(30 * time.Second)
	if err := authClient.Connect(); err != nil {
		logging.Logger.Fatal().Err(err).Msg("failed to connect to auth service")
	}
	defer authClient.Close()

	accountClient := clients.NewAccountClient(30 * time.Second)
	if err := accountClient.Connect(); err != nil {
		logging.Logger.Fatal().Err(err).Msg("failed to connect to account service")
	}
	defer accountClient.Close()

	// Start connection monitor
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go accountClient.StartConnectionMonitor(ctx)
	go authClient.StartConnectionMonitor(ctx)

	http.StartServer(http.GrpcClients{AuthClient: &authClient, AccountClient: &accountClient})
}
