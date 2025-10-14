package http

import (
	"fmt"
	"gateway-service/internal/config"
	"gateway-service/internal/grpc/clients"
	middleware "gateway-service/internal/http/middlewares"
	"gateway-service/internal/logging"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"net/http"
)

type GrpcClients struct {
	AuthClient    *clients.AuthClient
	AccountClient *clients.AccountClient
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
	r.Use(gin.LoggerWithWriter(gin.DefaultWriter, "/health", "/ready", "/metrics"), cors.New(corsConfig), middleware.Metrics(), middleware.Tracing(), gin.Recovery())

	// Setup routes
	setRoutes(r, gRPCClients)

	logging.Logger.Info().Msg(fmt.Sprintf("server listening on %s", config.Current().HTTP.Addr))

	// Run the server
	err := r.Run(config.Current().HTTP.Addr)
	if err != nil {
		logging.Logger.Error().Err(err).Msg("Failed to start server")
	}
}
