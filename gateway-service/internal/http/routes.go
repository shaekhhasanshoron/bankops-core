package http

import (
	"gateway-service/api/docs"
	"gateway-service/internal/config"
	"gateway-service/internal/http/handlers"
	middleware "gateway-service/internal/http/middlewares"
	"gateway-service/internal/observability/metrics"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"net/http"
)

// setRoutes sets up all the routes for the API Gateway.
func setRoutes(router *gin.Engine, gRPCClients GrpcClients) {
	docs.SwaggerInfo.Title = "BankOps Core - API Docs"
	docs.SwaggerInfo.Description = "API documentation for the BankOps Core"
	docs.SwaggerInfo.Version = "1.0"
	// Serve static files (for swagger, images, etc.)
	router.Static("/static", "./static")

	// Load templates
	router.LoadHTMLGlob("static/*.html")

	// Route for serving the index page
	router.GET("/", func(c *gin.Context) {
		// Define the Swagger URL dynamically
		swaggerURL := "/swagger/index.html"

		// Render the index.html file with the dynamic Swagger URL
		c.HTML(http.StatusOK, "index.html", gin.H{
			"SwaggerURL": swaggerURL,
		})
	})

	router.GET("/swagger", func(c *gin.Context) {
		c.Redirect(302, "/swagger/index.html")
	})

	router.GET("/swagger/*any", ginSwagger.WrapHandler(
		swaggerFiles.Handler,
		ginSwagger.DefaultModelsExpandDepth(-1),
		ginSwagger.DeepLinking(true),
	))

	// Health check routes
	router.GET("/health", handlers.Health)
	router.GET("/ready", handlers.Ready)

	if config.Current().Observability.MetricsConfig.Enabled {
		router.GET("/metrics", gin.WrapH(metrics.Handler()))
	}

	// Initialize handlers with clients
	authHandler := handlers.NewAuthHandler(*gRPCClients.AuthClient)
	accountHandler := handlers.NewAccountHandler(*gRPCClients.AccountClient)

	authGroup := router.Group("/api/v1/auth")
	{
		authGroup.POST("/login", authHandler.Login)
	}

	protectedGroup := router.Group("/api/v1")
	protectedGroup.Use(middleware.AuthMiddleware(gRPCClients.AuthClient), middleware.RequestID)
	{
		// Employee API
		protectedGroup.POST("/employee", authHandler.CreateEmployee)
		protectedGroup.DELETE("/employee/:username", authHandler.DeleteEmployee)
		protectedGroup.PUT("/employee/:username", authHandler.UpdateEmployee)
		protectedGroup.GET("/employee", authHandler.ListEmployee)
		// Customer API
		protectedGroup.POST("/customer", accountHandler.CreateCustomer)
		//protectedGroup.PUT("/customer/:id", accountHandler.UpdateCustomer)
		//protectedGroup.GET("/customer/:id", accountHandler.GetCustomer)
		protectedGroup.GET("/customer", accountHandler.ListCustomer)
		protectedGroup.GET("/customer/:id/account", accountHandler.ListCustomerAccounts)
		protectedGroup.DELETE("/customer/:id", accountHandler.DeleteCustomer)
		// Account API
		protectedGroup.POST("/account", accountHandler.CreateAccount)
		//protectedGroup.PUT("/account/:id", accountHandler.UpdateCustomer)
		//protectedGroup.GET("/account/:id", accountHandler.GetCustomer)
		protectedGroup.GET("/account/:id/balance", accountHandler.GetAccountBalance)
		protectedGroup.GET("/account", accountHandler.ListAccounts)
		protectedGroup.DELETE("/account", accountHandler.DeleteAccount)
		//Transaction API
		protectedGroup.POST("/transaction/init", accountHandler.InitTransaction)
		protectedGroup.GET("/transaction", accountHandler.ListTransactions)
	}
}
