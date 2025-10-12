package handlers

import "github.com/gin-gonic/gin"

// Health returns the health status of the service
func Health(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "ok",
	})
}

// Ready returns the readiness status of the service
func Ready(c *gin.Context) {
	// Check if necessary services (like DB) are ready here
	c.JSON(200, gin.H{
		"status": "ready",
	})
}
