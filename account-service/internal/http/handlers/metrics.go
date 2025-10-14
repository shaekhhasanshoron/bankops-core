package handlers

import (
	"account-service/internal/observability/metrics"
	"net/http"
)

// MetricsHandler serves the /metrics endpoint for Prometheus scraping.
func MetricsHandler(w http.ResponseWriter, r *http.Request) {
	metrics.Handler().ServeHTTP(w, r)
}
