package middleware

import (
	"gateway-service/internal/observability/tracing"
	"github.com/gin-gonic/gin"
	"strings"
)

// Tracing tracks trace information for each request
func Tracing() gin.HandlerFunc {
	return func(c *gin.Context) {
		if skipTrace(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Start tracing for the request
		span, ctx := tracing.StartSpan(c.Request.Context(), c.FullPath())
		defer tracing.EndSpan(span)

		// Set the context for further use
		c.Set("trace-context", ctx)

		// Add TraceID to the response headers
		c.Header("X-Trace-ID", span.SpanContext().TraceID().String())

		// Proceed with handling the request
		c.Next()
	}
}

func skipTrace(path string) bool {
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
