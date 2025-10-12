package main

import (
	"auth-service/internal/adapters/repo/sqlite"
	"auth-service/internal/auth"
	"auth-service/internal/bootstrap"
	"auth-service/internal/config"
	"auth-service/internal/db"
	"auth-service/internal/grpc"
	httpserver "auth-service/internal/http"
	"auth-service/internal/logging"
	"auth-service/internal/observability/metrics"
	"auth-service/internal/observability/tracing"
	"auth-service/internal/runtime"
	"context"
	"fmt"
	"log"
	"net"
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

	// Initialize database
	dbInstance, err := db.InitDatabase()
	if err != nil {
		logging.Logger.Fatal().Err(err).Msg("failed to connect to database")
		os.Exit(1)
	}

	// Ensure initial admin user is created
	if err := bootstrap.CreateAdmin(dbInstance); err != nil {
		logging.Logger.Fatal().Err(err).Msg("failed to create admin user")
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
		_ = traceShutdown(context.Background())
	}()

	// JWT Token Signer
	tokenSigner := auth.NewTokenSigner(config.Current().Auth.JWTSecret)

	go grpc.StartGRPCServer(sqlite.NewEmployeeRepo(dbInstance), tokenSigner)

	// Creating new http server for liveness and readiness checking
	srv := httpserver.NewServerHTTP(httpserver.ServerConfig{
		Addr:         config.Current().HTTP.Addr,
		ReadTimeout:  time.Duration(config.Current().HTTP.ReadTimeoutSeconds) * time.Second,
		WriteTimeout: time.Duration(config.Current().HTTP.WriteTimeoutSeconds) * time.Second,
		IdleTimeout:  time.Duration(config.Current().HTTP.IdleTimeoutSeconds) * time.Second,
	})

	// Listener for test
	listener, err := net.Listen("tcp", config.Current().HTTP.Addr)
	if err != nil {
		logging.Logger.Error().Err(err).Str("addr", config.Current().HTTP.Addr).Msg("listen failed")
		os.Exit(1)
	}

	// Getting context for receiving OS signals for initiate graceful shutdown.
	ctx := context.Background()
	ctx, stop := runtime.SignalContext(ctx)
	defer stop()

	logging.Logger.Info().Msg(fmt.Sprintf("server listening on %s", listener.Addr().String()))

	// Start the http server and watch for OS signals
	if err := runtime.ServeHTTP(ctx, srv, listener); err != nil {
		logging.Logger.Error().Err(err).Str("addr", config.Current().HTTP.Addr).Msg("server exit")
	}

	logging.Logger.Info().Str("addr", config.Current().HTTP.Addr).Msg("server closed")
}
