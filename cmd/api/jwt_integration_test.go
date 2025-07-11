package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"secure-email-mvp/pkg/auth"
)

func TestJWTTokenStructure(t *testing.T) {
	// Set JWT_SECRET for testing
	os.Setenv("JWT_SECRET", "test-secret-key-for-jwt-structure")

	// Generate a JWT token
	email := "test@example.com"
	token, err := auth.GenerateJWT(email)
	if err != nil {
		t.Fatalf("GenerateJWT failed: %v", err)
	}

	// Check token structure (JWT tokens have 3 parts separated by dots)
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Errorf("Expected JWT token to have 3 parts, got %d", len(parts))
	}

	// Check that all parts are non-empty
	for i, part := range parts {
		if part == "" {
			t.Errorf("JWT token part %d is empty", i)
		}
	}

	// Parse the token to verify claims
	claims, err := auth.ParseJWT(token)
	if err != nil {
		t.Fatalf("ParseJWT failed: %v", err)
	}

	// Verify claims
	if claims.Email != email {
		t.Errorf("Expected email %s, got %s", email, claims.Email)
	}

	if claims.Issuer != "secure-email-mvp" {
		t.Errorf("Expected issuer 'secure-email-mvp', got '%s'", claims.Issuer)
	}

	// Check that token is not expired
	if claims.ExpiresAt <= 0 {
		t.Error("Token should have a valid expiration time")
	}

	t.Logf("Generated JWT token: %s", token)
	t.Logf("Token claims: email=%s, issuer=%s, expires=%d", claims.Email, claims.Issuer, claims.ExpiresAt)
}

func TestJWTResponseFormat(t *testing.T) {
	// Set JWT_SECRET for testing
	os.Setenv("JWT_SECRET", "test-secret-key-for-response-format")

	// Simulate a login response
	email := "user@example.com"
	token, err := auth.GenerateJWT(email)
	if err != nil {
		t.Fatalf("GenerateJWT failed: %v", err)
	}

	// Create response
	response := LoginResponse{
		Token:   token,
		Message: "Login successful",
	}

	// Marshal to JSON to verify format
	jsonData, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("JSON marshal failed: %v", err)
	}

	// Verify JSON structure
	var parsedResponse LoginResponse
	err = json.Unmarshal(jsonData, &parsedResponse)
	if err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	if parsedResponse.Token != token {
		t.Error("Token was not preserved in JSON marshaling")
	}

	if parsedResponse.Message != "Login successful" {
		t.Error("Message was not preserved in JSON marshaling")
	}

	t.Logf("Login response JSON: %s", string(jsonData))
}
