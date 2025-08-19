package auth

import (
	"os"
	"strconv"
)

// Argon2Config defines Argon2 password hashing parameters
type Argon2Config struct {
	Memory      uint32 // Memory cost in KB
	Iterations  uint32 // Number of iterations
	Parallelism uint8  // Number of parallel threads
	KeyLength   uint32 // Length of the derived key
}

// TOTPConfig defines TOTP validation parameters
type TOTPConfig struct {
	Period    uint   // Time period in seconds (default: 30)
	Skew      uint   // Time skew tolerance (default: 1)
	Digits    int    // Number of digits (default: 6)
	Algorithm string // Hash algorithm (default: "SHA1")
	// RFC 6238 compliant drift tolerance settings
	DriftToleranceSeconds int // ±30 seconds tolerance (default: 30)
	MaxDriftSteps         int // Maximum drift steps to check (default: 2)
}

// AuthConfig defines the global authentication configuration
type AuthConfig struct {
	Argon2     Argon2Config
	TOTP       TOTPConfig
	UseNewFlow bool // Feature flag for new authentication flow
}

// GlobalAuthConfig is the global authentication configuration instance
var GlobalAuthConfig = LoadAuthConfig()

// LoadAuthConfig loads authentication configuration from environment variables
// with sensible defaults for production use
func LoadAuthConfig() AuthConfig {
	config := AuthConfig{
		Argon2: Argon2Config{
			Memory:      getEnvUint32("ARGON2_MEMORY", 64*1024), // 64MB
			Iterations:  getEnvUint32("ARGON2_ITERATIONS", 3),
			Parallelism: getEnvUint8("ARGON2_PARALLELISM", 2),
			KeyLength:   getEnvUint32("ARGON2_KEY_LENGTH", 32),
		},
		TOTP: TOTPConfig{
			Period:    getEnvUint("TOTP_PERIOD", 30),
			Skew:      getEnvUint("TOTP_SKEW", 1),
			Digits:    6,      // Fixed at 6 for compatibility
			Algorithm: "SHA1", // Fixed for compatibility
			// RFC 6238 compliant drift tolerance
			DriftToleranceSeconds: getEnvInt("TOTP_DRIFT_TOLERANCE_SECONDS", 30),
			MaxDriftSteps:         getEnvInt("TOTP_MAX_DRIFT_STEPS", 2),
		},
		UseNewFlow: getEnvBool("AUTH_USE_NEW_FLOW", true),
	}

	return config
}

// Helper functions to load environment variables with defaults

func getEnvUint32(key string, defaultValue uint32) uint32 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseUint(value, 10, 32); err == nil {
			return uint32(parsed)
		}
	}
	return defaultValue
}

func getEnvUint8(key string, defaultValue uint8) uint8 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseUint(value, 10, 8); err == nil {
			return uint8(parsed)
		}
	}
	return defaultValue
}

func getEnvUint(key string, defaultValue uint) uint {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseUint(value, 10, 32); err == nil {
			return uint(parsed)
		}
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseInt(value, 10, 64); err == nil {
			return int(parsed)
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// DefaultArgon2Config returns the default Argon2 configuration for testing
func DefaultArgon2Config() Argon2Config {
	return Argon2Config{
		Memory:      64 * 1024, // 64MB
		Iterations:  1,         // Reduced for testing
		Parallelism: 4,
		KeyLength:   32,
	}
}

// DefaultTOTPConfig returns the default TOTP configuration for testing
func DefaultTOTPConfig() TOTPConfig {
	return TOTPConfig{
		Period:                30,
		Skew:                  1,
		Digits:                6,
		Algorithm:             "SHA1",
		DriftToleranceSeconds: 30,
		MaxDriftSteps:         2,
	}
}
