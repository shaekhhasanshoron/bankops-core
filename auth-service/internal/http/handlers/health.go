package handlers

import (
	"net/http"
)

// HealthCheck For liveness probe: k8s process will check
func HealthCheck(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

// ReadyzCheck For readiness probe: k8s process will check
func ReadyzCheck(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ready"))
}

//func Tracer(w http.ResponseWriter, r *http.Request) {
//	// Get the global tracer
//	tracer := otel.Tracer("auth-service-handler")
//
//	// Start a new span for the request
//	_, span := tracer.Start(r.Context(), "auth-request")
//	defer span.End()
//
//	// Log attributes to the span
//	span.SetAttributes(
//		semconv.HTTPMethodKey.String(r.Method),
//		semconv.HTTPURLKey.String(r.URL.Path),
//		attribute.String("custom-attribute", "value"), // Example of a custom attribute
//	)
//
//	// Simulate some work (e.g., sleep for 2 seconds)
//	time.Sleep(2 * time.Second)
//
//	// Respond to the HTTP request
//	_, err := w.Write([]byte("ok"))
//	if err != nil {
//		logging.Logger.Error().Err(err).Msg("Error writing response")
//	}
//}
