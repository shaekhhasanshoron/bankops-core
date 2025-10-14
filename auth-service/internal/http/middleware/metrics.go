package middleware

import (
	"net/http"
	"time"

	"auth-service/internal/observability/metrics"
)

// Metrics middleware tracks the metrics for HTTP requests.
func Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" || r.URL.Path == "/ready" || r.URL.Path == "/metrics" {
			next.ServeHTTP(w, r)
			return
		}

		// Measure request duration
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w}
		next.ServeHTTP(sw, r)
		duration := time.Since(start)

		// Record the metrics
		metrics.ObserveHTTP(r.Method, r.URL.Path, sw.status, duration)
	})
}
