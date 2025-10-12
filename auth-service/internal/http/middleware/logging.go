package middleware

import (
	"net/http"
	"strings"
	"time"

	"auth-service/internal/logging"
)

func isProbePath(reqPath string) bool {
	return reqPath == "/healthz" || reqPath == "/readyz" || reqPath == "/metrics"
}

// AccessLog logs method, path, status, duration
func AccessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isProbePath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}
		sw := &statusWriter{ResponseWriter: w}
		start := time.Now()
		next.ServeHTTP(sw, r)
		dur := time.Since(start)

		status := sw.status
		if status == 0 {
			status = http.StatusOK
		}

		logging.Logger.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", status).
			Dur("dur", dur).
			Str("req_id", r.Header.Get("X-Request-ID")).
			Str("remote", remoteIP(r.RemoteAddr)).
			Msg("http_access")
	})
}

func remoteIP(hostport string) string {
	if i := strings.LastIndex(hostport, ":"); i > 0 {
		return hostport[:i]
	}
	return hostport
}
