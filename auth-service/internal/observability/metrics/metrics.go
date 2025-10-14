package metrics

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"auth-service/internal/config"
	"auth-service/internal/logging"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpReqTotal    *prometheus.CounterVec
	httpReqDuration *prometheus.HistogramVec
	grpcReqTotal    *prometheus.CounterVec
	grpcReqDuration *prometheus.HistogramVec
	dbConnections   prometheus.Gauge
	mu              sync.RWMutex
)

// Init sets up the Prometheus metrics for HTTP requests.
func Init() {
	if !config.Current().Observability.MetricsConfig.Enabled {
		logging.Logger.Info().Msg("metrics disabled")
		return
	}

	// Initialize the metrics
	grpcReqTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_requests_total",
			Help: "Total number of gRPC requests.",
		},
		[]string{"method", "status"},
	)

	grpcReqDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_request_duration_seconds",
			Help:    "Histogram of gRPC request durations",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)

	dbConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_active",
			Help: "Number of active database connections.",
		},
	)

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
	prometheus.MustRegister(
		httpReqTotal,
		httpReqDuration,
		grpcReqTotal,
		grpcReqDuration,
		dbConnections,
	)

	logging.Logger.Info().Msg("metrics initialized")
}

// Handler returns the /metrics handler for Prometheus to scrape.
func Handler() http.Handler {
	return promhttp.Handler()
}

// ObserveHTTP records HTTP request metrics.
func ObserveHTTP(method, path string, status int, dur time.Duration) {
	mu.Lock()
	defer mu.Unlock()

	if httpReqTotal != nil && httpReqDuration != nil {
		httpReqTotal.WithLabelValues(method, path, itoa(status)).Inc()
		httpReqDuration.WithLabelValues(method, path).Observe(dur.Seconds())
	}
}

func ObserveGRPC(method string, status string, duration time.Duration) {
	mu.Lock()
	defer mu.Unlock()

	if grpcReqTotal == nil || grpcReqDuration == nil {
		return
	}

	grpcReqTotal.WithLabelValues(method, status).Inc()
	grpcReqDuration.WithLabelValues(method).Observe(duration.Seconds())
}

func itoa(i int) string {
	return fmt.Sprintf("%d", i)
}
