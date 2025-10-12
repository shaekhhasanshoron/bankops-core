package metrics

import (
	"fmt"
	"gateway-service/internal/config"
	"gateway-service/internal/logging"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpReqTotal    *prometheus.CounterVec
	httpReqDuration *prometheus.HistogramVec
)

// Init sets up the Prometheus metrics for HTTP requests.
func Init() {
	if !config.Current().Observability.MetricsConfig.Enabled {
		logging.Logger.Info().Msg("metrics disabled")
		return
	}

	// Initialize the metrics
	httpReqTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method", "path", "status"},
	)

	httpReqDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Histogram of HTTP request durations.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// Register the metrics with Prometheus
	prometheus.MustRegister(httpReqTotal, httpReqDuration)

	logging.Logger.Info().Msg("metrics initialized")
}

// Handler returns the /metrics handler for Prometheus to scrape.
func Handler() http.Handler {
	return promhttp.Handler()
}

// ObserveHTTP records HTTP request metrics.
func ObserveHTTP(method, path string, status int, dur time.Duration) {
	if httpReqTotal != nil && httpReqDuration != nil {
		httpReqTotal.WithLabelValues(method, path, itoa(status)).Inc()
		httpReqDuration.WithLabelValues(method, path).Observe(dur.Seconds())
	}
}

func itoa(i int) string {
	return fmt.Sprintf("%d", i)
}
