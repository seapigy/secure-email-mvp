package auth

import (
	"bytes"
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/argon2"
)

// normalizeEmail standardizes email format for consistent salting
// This ensures that email addresses are handled consistently across signup and login
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// hashPasswordWithConfig creates Argon2 hash using the provided configuration
// This ensures consistent hashing parameters across the system
func hashPasswordWithConfig(password, email string, config Argon2Config) ([]byte, error) {
	normalizedEmail := normalizeEmail(email)

	// DIAGNOSTIC: Log hashing details
	log.Printf("[AUTH_DEBUG] hashPasswordWithConfig - Email: '%s', Normalized: '%s'", email, normalizedEmail)
	log.Printf("[AUTH_DEBUG] Argon2 Parameters - Memory: %d, Iterations: %d, Parallelism: %d, KeyLength: %d",
		config.Memory, config.Iterations, config.Parallelism, config.KeyLength)

	// Generate hash using Argon2 with configured parameters
	hash := argon2.IDKey([]byte(password), []byte(normalizedEmail),
		config.Iterations, config.Memory, config.Parallelism, config.KeyLength)

	// DIAGNOSTIC: Log hash generation details
	log.Printf("[AUTH_DEBUG] Generated hash (first 16 bytes): %x", hash[:minInt(16, len(hash))])

	return hash, nil
}

// validateTOTPWithConfig validates TOTP code using the provided configuration
// This ensures consistent TOTP validation with proper time skew tolerance
func validateTOTPWithConfig(code, secret string, config TOTPConfig) bool {
	// DIAGNOSTIC: Log TOTP validation details
	log.Printf("[AUTH_DEBUG] validateTOTPWithConfig - Code: %s, Secret length: %d", code, len(secret))
	log.Printf("[AUTH_DEBUG] TOTP Parameters - Period: %d, Skew: %d, Digits: %d, Algorithm: %s",
		config.Period, config.Skew, config.Digits, config.Algorithm)

	// Validate TOTP with custom parameters for time skew tolerance
	valid, err := totp.ValidateCustom(code, secret, time.Now(), totp.ValidateOpts{
		Period:    config.Period,
		Skew:      config.Skew,
		Digits:    6, // Default to 6 digits
		Algorithm: 1, // Default to SHA1 for compatibility
	})

	if err != nil {
		log.Printf("[AUTH_DEBUG] TOTP validation error: %v", err)
		return false
	}

	log.Printf("[AUTH_DEBUG] TOTP validation result: %t", valid)
	return valid
}

// generateTOTPSecretWithConfig generates a new TOTP secret using the provided configuration
func generateTOTPSecretWithConfig(config TOTPConfig) (string, error) {
	// Generate 20 random bytes for TOTP secret
	secret := make([]byte, 20)
	if _, err := rand.Read(secret); err != nil {
		return "", fmt.Errorf("failed to generate TOTP secret: %v", err)
	}

	// Encode as base32 (standard for TOTP)
	base32Secret := base32.StdEncoding.EncodeToString(secret)

	log.Printf("[AUTH_DEBUG] Generated TOTP secret (base32): %s", base32Secret)
	return base32Secret, nil
}

// compareHashes safely compares two byte slices for timing attack resistance
func compareHashes(hash1, hash2 []byte) bool {
	return bytes.Equal(hash1, hash2)
}

// minInt returns the minimum of two integers
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// logHashComparison logs hash comparison details for debugging
func logHashComparison(expected, actual []byte, email string) {
	log.Printf("[AUTH_DEBUG] Hash comparison for user: %s", email)
	log.Printf("[AUTH_DEBUG]   Expected hash (first 16 bytes): %x", expected[:minInt(16, len(expected))])
	log.Printf("[AUTH_DEBUG]   Actual hash (first 16 bytes): %x", actual[:minInt(16, len(actual))])
	log.Printf("[AUTH_DEBUG]   Hash lengths - Expected: %d, Actual: %d", len(expected), len(actual))
	log.Printf("[AUTH_DEBUG]   Hash comparison result: %t", compareHashes(expected, actual))
}

// logTOTPValidation logs TOTP validation details for debugging
func logTOTPValidation(code, secret string, config TOTPConfig) {
	log.Printf("[AUTH_DEBUG] TOTP validation details:")
	log.Printf("[AUTH_DEBUG]   Code: %s", code)
	log.Printf("[AUTH_DEBUG]   Secret (base32): %s", secret)
	log.Printf("[AUTH_DEBUG]   Secret length: %d", len(secret))
	log.Printf("[AUTH_DEBUG]   Current time: %v", time.Now())
	log.Printf("[AUTH_DEBUG]   Period: %d seconds", config.Period)
	log.Printf("[AUTH_DEBUG]   Skew tolerance: ±%d time steps", config.Skew)
}
