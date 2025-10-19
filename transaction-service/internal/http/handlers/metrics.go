package handlers

import (
	"net/http"
	"transaction-service/internal/observability/metrics"
)

// MetricsHandler serves the /metrics endpoint for Prometheus scraping.
func MetricsHandler(w http.ResponseWriter, r *http.Request) {
	metrics.Handler().ServeHTTP(w, r)
}
