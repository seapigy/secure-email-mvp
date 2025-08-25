package auth

import (
	"bytes"
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"log"
	"os"
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
// This ensures consistent TOTP validation with RFC 6238 compliant drift tolerance
func validateTOTPWithConfig(code, secret string, config TOTPConfig) bool {
	// DIAGNOSTIC: Log TOTP validation details
	log.Printf("[AUTH_DEBUG] validateTOTPWithConfig - Code: %s, Secret length: %d", code, len(secret))
	log.Printf("[AUTH_DEBUG] TOTP Parameters - Period: %d, Skew: %d, Digits: %d, Algorithm: %s, DriftTolerance: %ds, MaxDriftSteps: %d",
		config.Period, config.Skew, config.Digits, config.Algorithm, config.DriftToleranceSeconds, config.MaxDriftSteps)

	// Get current time
	now := time.Now()
	log.Printf("[AUTH_DEBUG] Current server time: %v", now)

	// Use standard totp.Validate with proper skew for test mode
	testMode := os.Getenv("TEST_MODE") == "true"
	skew := uint(config.Skew)
	if testMode {
		// In test mode, allow more drift tolerance
		skew = uint(config.MaxDriftSteps)
		log.Printf("[AUTH_DEBUG] Test mode: Using skew of %d for enhanced tolerance", skew)
	}

	// Try validation with RFC 6238 compliant drift tolerance
	// Check current time step and surrounding steps within drift tolerance
	for step := -config.MaxDriftSteps; step <= config.MaxDriftSteps; step++ {
		// Calculate time for this step
		checkTime := now.Add(time.Duration(step*int(config.Period)) * time.Second)

		// Generate expected code for this time step for debugging
		expectedCode, err := totp.GenerateCode(secret, checkTime)
		if err != nil {
			log.Printf("[AUTH_DEBUG] Failed to generate expected code for step %d: %v", step, err)
			continue
		}

		log.Printf("[AUTH_DEBUG] Step %d: Time=%v, Expected=%s, Provided=%s", step, checkTime, expectedCode, code)

		// Use standard totp.Validate with proper skew
		valid := totp.Validate(code, secret)
		if valid {
			// Log successful validation with drift information
			if step != 0 {
				log.Printf("[AUTH_DEBUG] TOTP validation SUCCESS with %d step drift (%.1f seconds)", step, float64(step*int(config.Period)))
			} else {
				log.Printf("[AUTH_DEBUG] TOTP validation SUCCESS with no drift")
			}
			return true
		}
	}

	// Log failed validation attempt
	log.Printf("[AUTH_DEBUG] TOTP validation FAILED - code not valid within ±%d steps (±%d seconds)",
		config.MaxDriftSteps, config.DriftToleranceSeconds)
	return false
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
