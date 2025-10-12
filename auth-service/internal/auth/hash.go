package auth

import (
	"auth-service/internal/config"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

// HashData generates a hashed data using HMAC-SHA256 and a key from config
func HashData(data string) (string, error) {
	// Retrieve the hash key from the config
	hashKey := config.Current().Auth.HashKey
	if hashKey == "" {
		return "", fmt.Errorf("hash key is not configured")
	}

	// Create a new HMAC SHA256 hash using the key from the config
	hmacHash := hmac.New(sha256.New, []byte(hashKey))
	hmacHash.Write([]byte(data))
	hashedPassword := base64.URLEncoding.EncodeToString(hmacHash.Sum(nil))

	return hashedPassword, nil
}
