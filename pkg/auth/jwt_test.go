package auth

import (
	"os"
	"testing"
	"time"
)

func TestGenerateJWT(t *testing.T) {
	// Set JWT_SECRET for testing
	os.Setenv("JWT_SECRET", "test-secret-key")

	// Test generating JWT token
	email := "test@example.com"
	token, err := GenerateJWT(email)
	if err != nil {
		t.Fatalf("GenerateJWT failed: %v", err)
	}

	if token == "" {
		t.Error("Expected non-empty JWT token, got empty")
	}

	if len(token) < 50 {
		t.Errorf("Expected JWT token to be longer, got length %d", len(token))
	}

	// Test token validation
	claims, err := ParseJWT(token)
	if err != nil {
		t.Fatalf("ParseJWT failed: %v", err)
	}

	if claims.Email != email {
		t.Errorf("Expected email %s, got %s", email, claims.Email)
	}

	// Check expiration time
	now := time.Now().Unix()
	if claims.ExpiresAt <= now {
		t.Error("Token should not be expired")
	}

	// Check that expiration is within 1 hour
	expectedExp := now + 3600              // 1 hour
	if claims.ExpiresAt > expectedExp+60 { // Allow 1 minute tolerance
		t.Error("Token expiration should be within 1 hour")
	}
}

func TestParseJWT(t *testing.T) {
	// Set JWT_SECRET for testing
	os.Setenv("JWT_SECRET", "test-secret-key")

	// Test valid token
	email := "test@example.com"
	token, err := GenerateJWT(email)
	if err != nil {
		t.Fatalf("GenerateJWT failed: %v", err)
	}

	claims, err := ParseJWT(token)
	if err != nil {
		t.Fatalf("ParseJWT failed: %v", err)
	}

	if claims.Email != email {
		t.Errorf("Expected email %s, got %s", email, claims.Email)
	}

	// Test invalid token
	_, err = ParseJWT("invalid.token.here")
	if err == nil {
		t.Error("Expected error for invalid token")
	}

	// Test empty token
	_, err = ParseJWT("")
	if err == nil {
		t.Error("Expected error for empty token")
	}

	// Test token with wrong secret
	os.Setenv("JWT_SECRET", "different-secret")
	_, err = ParseJWT(token)
	if err == nil {
		t.Error("Expected error for token with wrong secret")
	}
}

func TestGenerateJWTWithDefaultSecret(t *testing.T) {
	// Clear JWT_SECRET to test default behavior
	os.Unsetenv("JWT_SECRET")

	email := "test@example.com"
	token, err := GenerateJWT(email)
	if err != nil {
		t.Fatalf("GenerateJWT with default secret failed: %v", err)
	}

	if token == "" {
		t.Error("Expected non-empty JWT token, got empty")
	}

	// Test validation with default secret
	claims, err := ParseJWT(token)
	if err != nil {
		t.Fatalf("ParseJWT with default secret failed: %v", err)
	}

	if claims.Email != email {
		t.Errorf("Expected email %s, got %s", email, claims.Email)
	}
}
