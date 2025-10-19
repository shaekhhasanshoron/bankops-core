package runtime

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"
	"transaction-service/internal/logging"
)

// SignalContext checks OS signals and returns a context on system close or container termination
func SignalContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
}

// ServeHTTP starts the server on the provided listener and shuts it down gracefully
func ServeHTTP(ctx context.Context, srv *http.Server, listener net.Listener) error {
	errCh := make(chan error, 1)

	// Watching the listener, if any error occurs then close it.
	go func() {
		if err := srv.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		logging.Logger.Info().Msg("gracefully shutting down server")
		_ = srv.Shutdown(shutdownCtx)
		return <-errCh
	case err := <-errCh:
		return err
	}
}
