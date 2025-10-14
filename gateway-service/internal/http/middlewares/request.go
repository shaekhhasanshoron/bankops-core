package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const headerRequestID = "X-Request-ID"

// RequestID ensures each request has an X-Request-ID
func RequestID(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")
	if requestID == "" {
		requestID = uuid.NewString()
		c.Request.Header.Set("X-Request-ID", requestID)
	}

	// Set the header for the response
	c.Header("X-Request-ID", requestID)

	c.Next()
}
