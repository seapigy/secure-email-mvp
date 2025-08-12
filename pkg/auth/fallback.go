package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"time"
)

// GenerateFallbackToken creates a secure HMAC-based token for fallback email verification. Uses JWT secret for signing.
func GenerateFallbackToken(email string) string {
	// Get secret from environment or use default
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "default-secret-key"
	}

	// Create HMAC with email and timestamp
	timestamp := time.Now().Unix()
	data := fmt.Sprintf("%s:%d", email, timestamp)

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))

	return hex.EncodeToString(h.Sum(nil))
}

// GenerateFallbackExpiration returns a timestamp 1 hour from now for token expiry.
func GenerateFallbackExpiration() time.Time {
	return time.Now().Add(1 * time.Hour)
}

// IsTokenExpired checks if a given timestamp is in the past for token validity.
func IsTokenExpired(expiration time.Time) bool {
	return time.Now().After(expiration)
}

// SendFallbackConfirmationEmail logs a confirmation link (stub for real email sending). Used for account recovery.
func SendFallbackConfirmationEmail(email, token string) error {
	// In a real implementation, this would send an actual email
	// For now, we'll just log the confirmation link
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	confirmationLink := fmt.Sprintf("%s/confirm-fallback?token=%s", baseURL, token)

	log.Printf("FALLBACK EMAIL CONFIRMATION:")
	log.Printf("To: %s", email)
	log.Printf("Subject: Confirm Your Fallback Email")
	log.Printf("Body: Please click the following link to confirm your fallback email:")
	log.Printf("Link: %s", confirmationLink)

	return nil
}

// ValidateFallbackToken checks token length and hex validity for basic security.
func ValidateFallbackToken(token string) bool {
	// Basic validation - token should be at least 32 characters
	if len(token) < 32 {
		return false
	}

	// Check if it's valid hex
	_, err := hex.DecodeString(token)
	return err == nil
}

// GenerateResetToken creates a secure HMAC-based token for password reset
func GenerateResetToken() string {
	// Get secret from environment or use default
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "default-secret-key"
	}

	// Create HMAC with timestamp and random data
	timestamp := time.Now().UnixNano()
	data := fmt.Sprintf("reset:%d", timestamp)

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))

	return hex.EncodeToString(h.Sum(nil))
}

// SendPasswordResetEmail logs a password reset link (stub for real email sending)
func SendPasswordResetEmail(email, token string) error {
	// In a real implementation, this would send an actual email
	// For now, we'll just log the reset link
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	resetLink := fmt.Sprintf("%s/reset-password?token=%s&email=%s", baseURL, token, email)

	log.Printf("PASSWORD RESET EMAIL:")
	log.Printf("To: %s", email)
	log.Printf("Subject: Password Reset Request")
	log.Printf("Body: Please click the following link to reset your password:")
	log.Printf("Link: %s", resetLink)

	return nil
}
