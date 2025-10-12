package httpserver

import (
	"auth-service/internal/http/middleware"
	"net/http"
	"time"
)

type ServerConfig struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// chain applies middlewares in the given order (outer → inner).
func chain(h http.Handler, mws ...func(http.Handler) http.Handler) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

// NewServerHTTP generates a new http server
func NewServerHTTP(cfg ServerConfig) *http.Server {
	base := routes()

	h := chain(
		base,
		middleware.RequestID,
		middleware.AccessLog,
		middleware.Metrics,
	)

	return &http.Server{
		Addr:         cfg.Addr,
		Handler:      h,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}
}
