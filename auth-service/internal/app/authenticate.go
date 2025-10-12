package app

import (
	"auth-service/internal/auth"
	"auth-service/internal/config"
	"auth-service/internal/domain/entity"
	"auth-service/internal/logging"
	"auth-service/internal/ports"
	"errors"
	"fmt"
	"time"
)

// Authenticate is the use case for validating user credentials and generating a JWT.
type Authenticate struct {
	EmployeeRepo ports.EmployeeRepo
	TokenSigner  *auth.TokenSigner
}

// NewAuthenticate creates a new Authenticate use-case instance.
func NewAuthenticate(employeeRepo ports.EmployeeRepo, tokenSigner *auth.TokenSigner) *Authenticate {
	return &Authenticate{
		EmployeeRepo: employeeRepo,
		TokenSigner:  tokenSigner,
	}
}

// Execute validates the user's credentials (username/password) and returns the JWT token.
func (a *Authenticate) Execute(username, password string) (string, string, error) {
	// Fetch the user by username
	employee, err := a.EmployeeRepo.GetEmployeeByUsername(username)
	if err != nil {
		logging.Logger.Warn().Err(err).Str("username", username).Msg("user not found")
		return "", "", errors.New("invalid credentials or user not found")
	}

	hashedPassword, err := auth.HashData(password)
	if err != nil {
		logging.Logger.Warn().Err(err).Str("username", username).Msg("unable to hashed secret")
		return "", "", err
	}

	// Validate password
	if employee.Status != entity.EmployeeStatusValid || employee.Password != hashedPassword {
		logging.Logger.Warn().Str("username", username).Msg("invalid password or user status invalid")
		return "", "", errors.New("invalid credentials")
	}

	// Generate JWT token
	token, err := a.TokenSigner.SignJWT(employee.Username, employee.Role, config.Current().Auth.JWTSecret, 5*time.Minute)
	if err != nil {
		logging.Logger.Warn().Err(err).Msg("failed to generate JWT token")
		return "", "", fmt.Errorf("failed to generate token: %w", err)
	}

	// Generate Refresh Token (optional)
	refreshToken, err := a.TokenSigner.SignJWTRefreshToken(employee.Username, config.Current().Auth.JWTSecret)
	if err != nil {
		logging.Logger.Warn().Err(err).Msg("failed to generate refresh token")
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Return the generated tokens
	return token, refreshToken, nil
}
