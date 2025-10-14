package metrics

import (
	"account-service/internal/config"
	"account-service/internal/logging"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpReqTotal    *prometheus.CounterVec
	httpReqDuration *prometheus.HistogramVec
	grpcReqTotal    *prometheus.CounterVec
	grpcReqDuration *prometheus.HistogramVec
	operationTotal  *prometheus.CounterVec
	operationErrors *prometheus.CounterVec
	dbConnections   prometheus.Gauge
	activeRequests  prometheus.Gauge
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

	activeRequests = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_active_requests",
			Help: "Number of active HTTP requests.",
		},
	)

	operationTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "account_service_operations_total",
			Help: "Total number of business operations.",
		},
		[]string{"operation", "type"},
	)

	operationErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "account_service_operation_errors_total",
			Help: "Total number of business operation errors.",
		},
		[]string{"operation", "error_type"},
	)

	// Register the metrics with Prometheus
	prometheus.MustRegister(
		httpReqTotal,
		httpReqDuration,
		grpcReqTotal,
		grpcReqDuration,
		dbConnections,
		activeRequests,
		operationTotal,
		operationErrors,
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

func IncRequestActive() {
	mu.Lock()
	defer mu.Unlock()

	if !config.Current().Observability.MetricsConfig.Enabled || activeRequests == nil {
		return
	}
	activeRequests.Inc()
}

func DecRequestActive() {
	mu.Lock()
	defer mu.Unlock()

	if !config.Current().Observability.MetricsConfig.Enabled || activeRequests == nil {
		return
	}
	activeRequests.Dec()
}

func RecordOperation(operation string, err error) {
	if !config.Current().Observability.MetricsConfig.Enabled {
		return
	}

	mu.Lock()
	defer mu.Unlock()

	labels := make(prometheus.Labels)
	labels["operation"] = operation

	if err != nil {
		labels["type"] = "error"
		operationTotal.With(labels).Inc()

		errorLabels := make(prometheus.Labels)
		errorLabels["operation"] = operation
		errorLabels["error_type"] = classifyError(err)
		operationErrors.With(errorLabels).Inc()
	} else {
		labels["type"] = "success"
		operationTotal.With(labels).Inc()
	}
}

func classifyError(err error) string {
	if err == nil {
		return "none"
	}

	errStr := err.Error()
	switch {
	case contains(errStr, "not found"):
		return "not_found"
	case contains(errStr, "already exists"), contains(errStr, "duplicate"):
		return "duplicate"
	case contains(errStr, "validation"), contains(errStr, "invalid"):
		return "validation"
	case contains(errStr, "timeout"), contains(errStr, "deadline"):
		return "timeout"
	case contains(errStr, "circuit breaker"):
		return "circuit_breaker"
	case contains(errStr, "database"), contains(errStr, "sql"):
		return "database"
	case contains(errStr, "insufficient balance"):
		return "insufficient_balance"
	case contains(errStr, "locked"):
		return "locked"
	case contains(errStr, "concurrent"):
		return "concurrent_modification"
	default:
		return "general"
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func itoa(i int) string {
	return fmt.Sprintf("%d", i)
}
