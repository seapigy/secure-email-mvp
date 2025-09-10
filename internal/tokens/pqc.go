package tokens

import (
	"crypto/rand"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
)

// GenerateToken generates a cryptographically random token
func GenerateToken(n int) (string, error) {
	if n < 1 {
		return "", fmt.Errorf("token length must be at least 1")
	}
	
	// Generate random bytes
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}
	
	// Encode as base64 URL-safe string
	token := base64.URLEncoding.EncodeToString(bytes)
	return token, nil
}

// HashToken hashes a token using SHA-512
func HashToken(token string) (string, error) {
	hash := sha512.Sum512([]byte(token))
	return fmt.Sprintf("%x", hash), nil
}

// VerifyTokenHash verifies a token against its hash using constant time comparison
func VerifyTokenHash(token, hash string) bool {
	computedHash, err := HashToken(token)
	if err != nil {
		return false
	}
	
	return subtle.ConstantTimeCompare([]byte(computedHash), []byte(hash)) == 1
}

// GenerateSecureToken generates a secure token suitable for verification or recovery
func GenerateSecureToken() (string, error) {
	// Generate 32 bytes (256 bits) of random data
	return GenerateToken(32)
}
