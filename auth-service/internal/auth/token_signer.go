package auth

import (
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"time"
)

// TokenSigner is responsible for signing JWT tokens.
type TokenSigner struct {
	secretKey string
}

// Claims defines the structure of JWT claims.
type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// NewTokenSigner creates a new instance of TokenSigner.
func NewTokenSigner(secretKey string) *TokenSigner {
	return &TokenSigner{
		secretKey: secretKey,
	}
}

// Sign generates a SignJWT token for the user.
func (t *TokenSigner) SignJWT(username, role, secretKey string, expiryTime time.Duration) (string, error) {
	expirationTime := time.Now().Add(expiryTime)
	claims := &Claims{
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign the token: %w", err)
	}

	return signedToken, nil
}

// SignJWTRefreshToken generates a refresh token for the user.
func (t *TokenSigner) SignJWTRefreshToken(username, secretKey string) (string, error) {
	expirationTime := time.Now().Add(7 * 24 * time.Hour)
	claims := &Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign the refresh token: %w", err)
	}

	return signedToken, nil
}
