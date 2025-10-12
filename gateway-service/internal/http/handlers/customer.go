package handlers

import "github.com/gin-gonic/gin"

// GetCustomer returns customer list
// @Summary Get a customer
// @Description Retrieve a customer's details
// @Tags customers
// @Accept  json
// @Produce  json
// @Param id path int true "Customer ID"
// @Router /api/v1/customer/{id} [get]
func GetCustomer(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "customer",
		"status":  "ready",
	})
}
