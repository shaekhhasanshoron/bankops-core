package ports

import "time"

// TokenSigner is responsible for signing JWT tokens.
type TokenSigner interface {
	SignJWT(username, role, secretKey string, expiryTime time.Duration) (string, error)
	SignJWTRefreshToken(username, secretKey string) (string, error)
}
