package middleware

import (
	"gateway-service/internal/observability/metrics"
	"github.com/gin-gonic/gin"
	"strings"
	"time"
)

// Metrics records metrics for HTTP requests
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		if skipMetricsPath(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Start timer
		start := time.Now()

		// Process request
		c.Next()

		// Record metrics after request completes
		metrics.ObserveHTTP(c.Request.Method, c.Request.URL.Path, c.Writer.Status(), time.Since(start))
	}
}

func skipMetricsPath(path string) bool {
	skipMap := map[string]bool{
		"/health":  true,
		"/ready":   true,
		"/metrics": true,
	}

	if skipMap[path] || strings.HasPrefix(path, "/swagger") {
		return true
	}
	return false
}
