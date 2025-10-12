package integration

import (
	"auth-service/internal/runtime"
	"context"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	httpserver "auth-service/internal/http"
)

// TestHttpHealthEndpoints (Integration Test) tests health check endpoints
func TestHttpHealthEndpoints(t *testing.T) {
	// Listener with a random port
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := "http://" + ln.Addr().String()

	// Generating new server and context
	srv := httpserver.NewServerHTTP(httpserver.ServerConfig{
		Addr:         ln.Addr().String(),
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
		IdleTimeout:  5 * time.Second,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- runtime.ServeHTTP(ctx, srv, ln)
	}()

	time.Sleep(100 * time.Millisecond) // Waiting time

	// wait until server is actually up to avoid flakiness
	client := &http.Client{Timeout: 1 * time.Second}
	waitOK := func(url string) error {
		deadline := time.Now().Add(2 * time.Second)
		for {
			if time.Now().After(deadline) {
				return context.DeadlineExceeded
			}
			resp, err := client.Get(url)
			if err == nil {
				resp.Body.Close()
				return nil
			}
			time.Sleep(25 * time.Millisecond)
		}
	}

	if err := waitOK(addr + "/healthz"); err != nil {
		t.Fatalf("server did not become healthy in time: %v", err)
	}

	// Check health apis
	healthResp, err := http.Get(addr + "/healthz")
	if err != nil {
		t.Fatalf("GET /healthz: %v", err)
	}
	defer healthResp.Body.Close()
	if healthResp.StatusCode != http.StatusOK {
		t.Fatalf("healthz status = %d", healthResp.StatusCode)
	}

	b, _ := io.ReadAll(healthResp.Body)
	if string(b) != "ok" {
		t.Fatalf("healthz body = %q", string(b))
	}

	// Check readiness
	readinessResp, err := http.Get(addr + "/readyz")
	if err != nil {
		t.Fatalf("GET /readyz: %v", err)
	}
	defer readinessResp.Body.Close()
	if readinessResp.StatusCode != http.StatusOK {
		t.Fatalf("readyz status = %d", readinessResp.StatusCode)
	}

	// Shutdown and assert the server exits cleanly
	cancel()
	select {
	case e := <-errCh:
		if e != nil {
			t.Fatalf("server exit error: %v", e)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("server did not shut down in time")
	}
}
