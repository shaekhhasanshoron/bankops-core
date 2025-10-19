package http

import (
	"fmt"
	"gateway-service/internal/config"
	middleware "gateway-service/internal/http/middlewares"
	"gateway-service/internal/logging"
	"gateway-service/internal/ports"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"net/http"
)

type GrpcClients struct {
	AuthClient        *ports.AuthClient
	AccountClient     *ports.AccountClient
	TransactionClient *ports.TransactionClient
}

func StartServer(gRPCClients GrpcClients) {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"*", "http://localhost:4200"}
	corsConfig.AllowMethods = []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete}
	corsConfig.AllowCredentials = true
	corsConfig.AllowHeaders = []string{"Origin", "*"}
	corsConfig.AddAllowMethods("OPTIONS")

	// setting middleware
	r.Use(gin.LoggerWithWriter(gin.DefaultWriter, "/health", "/ready", "/metrics"), cors.New(corsConfig), gin.Recovery())
	if config.Current().Observability.MetricsConfig.Enabled {
		r.Use(middleware.Metrics())
	}

	if config.Current().Observability.TracingConfig.Enabled {
		r.Use(middleware.Tracing())
	}

	// Setup routes
	setRoutes(r, gRPCClients)

	logging.Logger.Info().Msg(fmt.Sprintf("server listening on %s", config.Current().HTTP.Addr))

	// Run the server
	err := r.Run(config.Current().HTTP.Addr)
	if err != nil {
		logging.Logger.Error().Err(err).Msg("Failed to start server")
	}
}
