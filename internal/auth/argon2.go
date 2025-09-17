package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Argon2Config holds the configuration for Argon2id hashing
type Argon2Config struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

// DefaultArgon2Config returns the default Argon2id configuration
func DefaultArgon2Config() *Argon2Config {
	return &Argon2Config{
		Memory:      getEnvUint32("ARGON2_MEMORY_KB", 1024) * 1024, // Convert KB to bytes - 1MB default for free tier
		Iterations:  getEnvUint32("ARGON2_ITERATIONS", 1),
		Parallelism: uint8(getEnvUint32("ARGON2_PARALLELISM", 1)),
		SaltLength:  getEnvUint32("ARGON2_SALT_LEN", 16),
		KeyLength:   getEnvUint32("ARGON2_KEY_LEN", 32),
	}
}

// HashPassword hashes a password using Argon2id
func HashPassword(password string) (hash string, err error) {
	config := DefaultArgon2Config()
	
	// Generate a random salt
	salt := make([]byte, config.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}
	
	// Hash the password
	key := argon2.IDKey([]byte(password), salt, config.Iterations, config.Memory, config.Parallelism, config.KeyLength)
	
	// Encode the hash with salt
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Key := base64.RawStdEncoding.EncodeToString(key)
	
	// Format: $argon2id$v=19$m=memory,t=time,p=parallelism$salt$key
	hash = fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, config.Memory, config.Iterations, config.Parallelism, b64Salt, b64Key)
	
	return hash, nil
}

// VerifyPassword verifies a password against its hash
func VerifyPassword(hash, password string) (bool, error) {
	// Parse the hash
	parts := strings.Split(hash, "$")
	if len(parts) != 6 {
		return false, fmt.Errorf("invalid hash format")
	}
	
	if parts[1] != "argon2id" {
		return false, fmt.Errorf("unsupported hash algorithm")
	}
	
	// Parse version
	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return false, fmt.Errorf("invalid version: %w", err)
	}
	
	// Parse parameters
	var memory, iterations uint32
	var parallelism uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism); err != nil {
		return false, fmt.Errorf("invalid parameters: %w", err)
	}
	
	// Decode salt and key
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("invalid salt: %w", err)
	}
	
	key, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("invalid key: %w", err)
	}
	
	// Hash the provided password with the same parameters
	otherKey := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, uint32(len(key)))
	
	// Compare keys using constant time comparison
	return subtle.ConstantTimeCompare(key, otherKey) == 1, nil
}

// getEnvUint32 gets an environment variable as uint32 with a default value
func getEnvUint32(key string, defaultValue uint32) uint32 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseUint(value, 10, 32); err == nil {
			return uint32(parsed)
		}
	}
	return defaultValue
}
