package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"
	"transaction-service/internal/adapter/grpc/clients"
	repo "transaction-service/internal/adapter/repo/sqlite"
	"transaction-service/internal/config"
	"transaction-service/internal/db"
	"transaction-service/internal/grpc"
	httpserver "transaction-service/internal/http"
	"transaction-service/internal/jobs"
	"transaction-service/internal/logging"
	"transaction-service/internal/messaging"
	"transaction-service/internal/observability/metrics"
	"transaction-service/internal/observability/tracing"
	"transaction-service/internal/runtime"
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

	// Initialize database
	dbInstance, err := db.InitDB()
	if err != nil {
		logging.Logger.Fatal().Err(err).Msg("failed to connect to database")
		os.Exit(1)
	}

	// Initialize Prometheus metrics
	metrics.Init()

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

	accountClient := clients.NewAccountClient(30*time.Second, config.Current().GRPC.AccountServiceAddr)
	if err := accountClient.Connect(); err != nil {
		logging.Logger.Fatal().Err(err).Msg("failed to connect to account service")
	}
	defer accountClient.Close()

	transactionRepo := repo.NewTransactionRepo(dbInstance)
	sagaRepo := repo.NewSagaRepo(dbInstance)
	eventRepo := repo.NewEventRepo(dbInstance)

	// Initiating message queues
	messaging.SetupMessaging()
	defer func() {
		if config.Current().MessagePublisher.Enabled {
			_ = messaging.GetService().Close()
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctx, stop := runtime.SignalContext(ctx)
	defer stop()

	go grpc.StartGRPCServer(ctx, grpc.ServiceRepos{
		AccountClient:   accountClient,
		SagaRepo:        sagaRepo,
		TransactionRepo: transactionRepo,
		EventRepo:       eventRepo,
	})

	recoveryJob := jobs.NewTransactionReconciliationJob(
		transactionRepo,
		accountClient,
		sagaRepo,
	)

	// Start recovery job
	go recoveryJob.Start(ctx)

	go accountClient.StartConnectionMonitor(ctx)

	// Listener for test
	listener, err := net.Listen("tcp", config.Current().HTTP.Addr)
	if err != nil {
		logging.Logger.Error().Err(err).Str("addr", config.Current().HTTP.Addr).Msg("listen failed")
		os.Exit(1)
	}

	// Creating new http server for liveness and readiness checking
	srv := httpserver.NewServerHTTP(httpserver.ServerConfig{
		Addr:         config.Current().HTTP.Addr,
		ReadTimeout:  time.Duration(config.Current().HTTP.ReadTimeoutSeconds) * time.Second,
		WriteTimeout: time.Duration(config.Current().HTTP.WriteTimeoutSeconds) * time.Second,
		IdleTimeout:  time.Duration(config.Current().HTTP.IdleTimeoutSeconds) * time.Second,
	})

	logging.Logger.Info().Msg(fmt.Sprintf("server listening on %s", listener.Addr().String()))

	// Start the http server and watch for OS signals
	if err := runtime.ServeHTTP(ctx, srv, listener); err != nil {
		logging.Logger.Error().Err(err).Str("addr", config.Current().HTTP.Addr).Msg("server exit")
	}

	logging.Logger.Info().Str("addr", config.Current().HTTP.Addr).Msg("server closed")
}
