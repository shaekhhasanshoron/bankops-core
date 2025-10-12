package httpserver

import (
	"auth-service/internal/config"
	"auth-service/internal/http/handlers"
	"auth-service/internal/http/middleware"
	"net/http"
)

func routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", handlers.HealthCheck)
	mux.HandleFunc("/readyz", handlers.ReadyzCheck)
	//mux.HandleFunc("/trace", handlers.Tracer)

	if config.Current().Observability.MetricsConfig.Enabled {
		mux.Handle("/metrics", middleware.Metrics(http.HandlerFunc(handlers.MetricsHandler)))
	}
	return mux
}
