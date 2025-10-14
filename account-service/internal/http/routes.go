package httpserver

import (
	"account-service/internal/config"
	"account-service/internal/http/handlers"
	"account-service/internal/http/middleware"
	"net/http"
)

func routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handlers.HealthCheck)
	mux.HandleFunc("/ready", handlers.ReadyCheck)

	if config.Current().Observability.MetricsConfig.Enabled {
		mux.Handle("/metrics", middleware.Metrics(http.HandlerFunc(handlers.MetricsHandler)))
	}
	return mux
}
