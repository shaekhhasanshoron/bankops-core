package handlers

import (
	protoauth "gateway-service/api/protogen/authservice/proto"
	"gateway-service/internal/logging"
	"gateway-service/internal/ports"
	"github.com/gin-gonic/gin"
	"net/http"
)

type AuthHandler struct {
	AuthClient ports.AuthClient
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
}

func NewAuthHandler(authClient ports.AuthClient) *AuthHandler {
	return &AuthHandler{
		AuthClient: authClient,
	}
}

// Login allows an employee to log in and get an access token
// @Tags Authentication
// @Summary Login API
// @Description
// @Description **Request Body:**
// @Description
// @Description username:
// @Description - Required
// @Description
// @Description password:
// @Description - Required
// @Accept json
// @Produce json
// @Param login body LoginRequest true "Login credentials"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logging.Logger.Warn().Err(err).Msg("invalid request param")
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request payload"})
		return
	}

	grpcReq := &protoauth.AuthenticateRequest{
		Username: req.Username,
		Password: req.Password,
	}

	resp, err := h.AuthClient.Authenticate(c.Request.Context(), grpcReq)
	if err != nil {
		logging.Logger.Error().Err(err).Msg("login failed")
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid credentials"})
		return
	}

	loginResp := LoginResponse{
		AccessToken: resp.Token,
	}

	c.JSON(http.StatusOK, loginResp)
}
