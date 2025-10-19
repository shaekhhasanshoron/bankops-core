package httpserver

import (
	"net/http"
	"transaction-service/internal/config"
	"transaction-service/internal/http/handlers"
	"transaction-service/internal/http/middleware"
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
