package auth

import (
	"auth-service/internal/ports"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

// Hashing implements ports.Hashing.
type Hashing struct {
	HashKey string
}

// NewHashing creates a new instance of Hashing.
func NewHashing(hashKey string) ports.Hashing {
	return &Hashing{
		HashKey: hashKey,
	}
}

// HashData generates a hashed data using HMAC-SHA256 and a key from config
func (a *Hashing) HashData(data string) (string, error) {
	// Create a new HMAC SHA256 hash using the key from the config
	hmacHash := hmac.New(sha256.New, []byte(a.HashKey))
	hmacHash.Write([]byte(data))
	hashedPassword := base64.URLEncoding.EncodeToString(hmacHash.Sum(nil))

	return hashedPassword, nil
}
