package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"secure-email-mvp/pkg/auth"

	"github.com/dgrijalva/jwt-go"
)

func TestJWTMiddleware(t *testing.T) {
	// Set JWT_SECRET for testing
	os.Setenv("JWT_SECRET", "test-secret-key-for-middleware")

	// Create a test handler that extracts email from context
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		email, ok := GetUserEmailFromContext(r)
		if !ok {
			http.Error(w, `{"error":"Email not found in context"}`, http.StatusInternalServerError)
			return
		}

		response := ProtectedResponse{
			Email:   email,
			Message: "Access granted",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Create middleware
	middleware := jwtMiddleware(testHandler)

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedError  string
		expectedEmail  string
	}{
		{
			name:           "Valid JWT token",
			authHeader:     "Bearer " + generateTestToken(t, "test@example.com"),
			expectedStatus: http.StatusOK,
			expectedEmail:  "test@example.com",
		},
		{
			name:           "Missing Authorization header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid or missing token",
		},
		{
			name:           "Invalid Authorization format",
			authHeader:     "InvalidFormat token",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid or missing token",
		},
		{
			name:           "Empty Bearer token",
			authHeader:     "Bearer ",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid or missing token",
		},
		{
			name:           "Invalid JWT token",
			authHeader:     "Bearer invalid.token.here",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid or missing token",
		},
		{
			name:           "Expired JWT token",
			authHeader:     "Bearer " + generateExpiredToken(t, "test@example.com"),
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid or missing token",
		},
		{
			name:           "JWT token with wrong secret",
			authHeader:     "Bearer " + generateTokenWithWrongSecret(t, "test@example.com"),
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid or missing token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call middleware
			middleware.ServeHTTP(rr, req)

			// Check status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			// Check error message if expected
			if tt.expectedError != "" {
				var errorResp map[string]string
				if err := json.Unmarshal(rr.Body.Bytes(), &errorResp); err != nil {
					t.Fatal(err)
				}
				if errorResp["error"] != tt.expectedError {
					t.Errorf("Expected error '%s', got '%s'", tt.expectedError, errorResp["error"])
				}
			}

			// Check success response if expected
			if tt.expectedEmail != "" {
				var successResp ProtectedResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &successResp); err != nil {
					t.Fatal(err)
				}
				if successResp.Email != tt.expectedEmail {
					t.Errorf("Expected email '%s', got '%s'", tt.expectedEmail, successResp.Email)
				}
				if successResp.Message != "Access granted" {
					t.Errorf("Expected message 'Access granted', got '%s'", successResp.Message)
				}
			}
		})
	}
}

func TestGetUserEmailFromContext(t *testing.T) {
	// Test with email in context
	email := "test@example.com"
	ctx := context.WithValue(context.Background(), UserEmailKey, email)
	req := httptest.NewRequest("GET", "/test", nil)
	req = req.WithContext(ctx)

	retrievedEmail, ok := GetUserEmailFromContext(req)
	if !ok {
		t.Error("Expected to find email in context")
	}
	if retrievedEmail != email {
		t.Errorf("Expected email %s, got %s", email, retrievedEmail)
	}

	// Test without email in context
	req = httptest.NewRequest("GET", "/test", nil)
	retrievedEmail, ok = GetUserEmailFromContext(req)
	if ok {
		t.Error("Expected not to find email in context")
	}
	if retrievedEmail != "" {
		t.Errorf("Expected empty email, got %s", retrievedEmail)
	}
}

func TestProtectedTestHandler(t *testing.T) {
	// Test with email in context
	email := "test@example.com"
	ctx := context.WithValue(context.Background(), UserEmailKey, email)
	req := httptest.NewRequest("GET", "/protected-test", nil)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	protectedTestHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var response ProtectedResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}

	if response.Email != email {
		t.Errorf("Expected email %s, got %s", email, response.Email)
	}
	if response.Message != "Access granted" {
		t.Errorf("Expected message 'Access granted', got '%s'", response.Message)
	}

	// Test without email in context
	req = httptest.NewRequest("GET", "/protected-test", nil)
	rr = httptest.NewRecorder()
	protectedTestHandler(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
}

// Helper functions for generating test tokens
func generateTestToken(t *testing.T, email string) string {
	token, err := auth.GenerateJWT(email)
	if err != nil {
		t.Fatalf("Failed to generate test token: %v", err)
	}
	return token
}

func generateExpiredToken(t *testing.T, email string) string {
	// Temporarily change JWT_SECRET to create an expired token
	originalSecret := os.Getenv("JWT_SECRET")
	os.Setenv("JWT_SECRET", "expired-secret-key")
	defer os.Setenv("JWT_SECRET", originalSecret)

	// Create a token that expires in the past
	secret := os.Getenv("JWT_SECRET")
	claims := &auth.Claims{
		Email: email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(-1 * time.Hour).Unix(), // Expired 1 hour ago
			IssuedAt:  time.Now().Add(-2 * time.Hour).Unix(),
			Issuer:    "secure-email-mvp",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("Failed to generate expired token: %v", err)
	}

	return tokenString
}

func generateTokenWithWrongSecret(t *testing.T, email string) string {
	// Create a token with a different secret
	claims := &auth.Claims{
		Email: email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(1 * time.Hour).Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "secure-email-mvp",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("wrong-secret-key"))
	if err != nil {
		t.Fatalf("Failed to generate token with wrong secret: %v", err)
	}

	return tokenString
}
