package middleware

import (
	"encoding/json"
	"errors"
	"fmt"
	"gateway-service/internal/auth"
	"gateway-service/internal/config"
	"gateway-service/internal/grpc/clients"
	"gateway-service/internal/logging"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"time"
)

func AuthMiddleware(authClient *clients.AuthClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		token := parts[1]

		claim, httpCode, err := validateToken(c, token)
		if err != nil {
			c.JSON(httpCode, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("username", claim.Username)
		c.Set("role", claim.Role)

		// Proceed with the request
		c.Next()
	}
}

// validateToken validates incoming token and returns Claims obj, httpStatusCode, err
func validateToken(c *gin.Context, token string) (*auth.Claims, int, error) {
	claim, err := auth.NewTokenSigner(config.Current().Auth.JWTSecret).ParseJWT(token)
	if err != nil {
		logging.Logger.Err(err).Msg("Invalid token")
		return nil, http.StatusUnauthorized, errors.New("Invalid or expired token")
	}

	jsonBody, _ := json.Marshal(claim)
	fmt.Println("Claims: " + string(jsonBody))

	// Check token expiry
	if time.Now().After(claim.ExpiresAt.Time) {
		logging.Logger.Err(errors.New("token expired")).Msg("Invalid token")
		return nil, http.StatusUnauthorized, errors.New("Token has expired")
	}

	// Define route-to-role mapping with method-specific permissions
	permissions := map[string]map[string]map[string]bool{
		"/api/v1/employee": {
			"GET":    {"admin": true, "editor": false, "viewer": false},
			"POST":   {"admin": true, "editor": false, "viewer": false},
			"DELETE": {"admin": true, "editor": false, "viewer": false},
		},
		"/api/v1/employee/:username": {
			"GET":    {"admin": true, "editor": true, "viewer": true},
			"DELETE": {"admin": true, "editor": false, "viewer": false},
			"PUT":    {"admin": true, "editor": false, "viewer": false},
		},
		"/api/v1/customer": {
			"POST": {"admin": true, "editor": true, "viewer": false},
			"GET":  {"admin": true, "editor": true, "viewer": true},
		},
		"/api/v1/customer/:id": {
			"PUT":    {"admin": true, "editor": true, "viewer": false},
			"DELETE": {"admin": true, "editor": true, "viewer": false},
			"GET":    {"admin": true, "editor": true, "viewer": true},
		},
		"/api/v1/customer/:id/account": {
			"GET": {"admin": true, "editor": true, "viewer": true},
		},
		"/api/v1/account": {
			"POST":   {"admin": true, "editor": true, "viewer": false},
			"GET":    {"admin": true, "editor": true, "viewer": true},
			"DELETE": {"admin": true, "editor": true, "viewer": false},
		},
		"/api/v1/account/:id": {
			"PUT":    {"admin": true, "editor": true, "viewer": false},
			"DELETE": {"admin": true, "editor": true, "viewer": false},
			"GET":    {"admin": true, "editor": true, "viewer": true},
		},
		"/api/v1/account/:id/balance": {
			"GET": {"admin": true, "editor": true, "viewer": true},
		},
		"/api/v1/transaction/init": {
			"POST": {"admin": true, "editor": true, "viewer": false},
		},
		"/api/v1/transaction": {
			"GET": {"admin": true, "editor": true, "viewer": true},
		},
	}

	// Get the requested URL path and HTTP method
	fullPath := c.FullPath()
	method := c.Request.Method
	//fmt.Println(fullPath, method, claim.Role)
	if permissions[fullPath][method][claim.Role] == false {
		logging.Logger.Err(errors.New("Permission Denied")).Msg("User dont have permission")
		return nil, http.StatusForbidden, errors.New("Permission Denied")
	}
	return claim, http.StatusOK, nil
}
