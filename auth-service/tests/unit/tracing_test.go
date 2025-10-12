package unit

import (
	"context"
	"testing"

	"auth-service/internal/config"
	"auth-service/internal/observability/tracing"
)

// TestTracing_NoCollector_NoCrash test if no OTEL collector exists the program should not crash
func TestTracing_NoCollector_NoCrash(t *testing.T) {
	t.Setenv("AUTH_ENV", "prod")
	t.Setenv("AUTH_HTTP__ADDR", "127.0.0.1:0")
	t.Setenv("AUTH_OBSERVABILITY__TRACING__ENABLED", "true")

	if _, err := config.LoadConfig(); err != nil {
		t.Fatalf("config: %v", err)
	}

	shutdown, err := tracing.Init(context.Background(), "auth-service")
	if err != nil {
		t.Fatalf("tracing init returned error: %v", err)
	}
	if err := shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown err: %v", err)
	}
}
