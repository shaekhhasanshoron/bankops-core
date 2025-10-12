package auth

import (
	"fmt"
	"github.com/golang-jwt/jwt/v4"
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

func (t *TokenSigner) ParseJWT(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// Return the secret key for validation
		return []byte(t.secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse the token: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, fmt.Errorf("invalid token")
	}
}
