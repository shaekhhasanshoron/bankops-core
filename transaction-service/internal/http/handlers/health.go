package handlers

import (
	"net/http"
)

// HealthCheck For liveness probe: k8s process will check
func HealthCheck(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

// ReadyCheck For readiness probe: k8s process will check
func ReadyCheck(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ready"))
}
