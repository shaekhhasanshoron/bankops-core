package app

import (
	"auth-service/internal/config"
	"auth-service/internal/domain/entity"
	custom_err "auth-service/internal/domain/error"
	"auth-service/internal/logging"
	"auth-service/internal/ports"
	"errors"
	"fmt"
	"strings"
)

// Authenticate is the use case for validating user credentials and generating a JWT.
type Authenticate struct {
	EmployeeRepo ports.EmployeeRepo
	TokenSigner  ports.TokenSigner
	Hashing      ports.Hashing
}

// NewAuthenticate creates a new Authenticate use-case instance.
func NewAuthenticate(employeeRepo ports.EmployeeRepo, tokenSigner ports.TokenSigner, hashing ports.Hashing) *Authenticate {
	return &Authenticate{
		EmployeeRepo: employeeRepo,
		TokenSigner:  tokenSigner,
		Hashing:      hashing,
	}
}

// Execute validates the user's credentials (username/password) and returns the JWT token and refresh token.
func (a *Authenticate) Execute(username, password string) (string, string, error) {
	if strings.TrimSpace(username) == "" {
		logging.Logger.Warn().Err(custom_err.ErrInvalidUsername).Str("username", username).Msg("Invalid username")
		return "", "", custom_err.ErrInvalidUsername
	}

	employee, err := a.EmployeeRepo.GetEmployeeByUsername(username)
	if err != nil {
		logging.Logger.Warn().Err(err).Str("username", username).Msg("user not found")
		return "", "", errors.New("invalid credentials or user not found")
	}

	hashedPassword, err := a.Hashing.HashData(password)
	if err != nil {
		logging.Logger.Warn().Err(err).Str("username", username).Msg("unable to hashed secret")
		return "", "", err
	}

	// Validate password
	if employee.Status != entity.EmployeeStatusValid || employee.Password != hashedPassword {
		logging.Logger.Warn().Str("username", username).Msg("invalid password or user status invalid")
		return "", "", errors.New("invalid credentials")
	}

	token, err := a.TokenSigner.SignJWT(employee.Username, employee.Role, config.Current().Auth.JWTSecret, config.Current().Auth.JWTTokentDuration)
	if err != nil {
		logging.Logger.Warn().Err(err).Msg("failed to generate JWT token")
		return "", "", fmt.Errorf("failed to generate token: %w", err)
	}

	refreshToken, err := a.TokenSigner.SignJWTRefreshToken(employee.Username, config.Current().Auth.JWTSecret)
	if err != nil {
		logging.Logger.Warn().Err(err).Msg("failed to generate refresh token")
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return token, refreshToken, nil
}
