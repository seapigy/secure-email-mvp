package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// TestJWTAuthentication tests JWT authentication middleware
func TestJWTAuthentication(t *testing.T) {
	// Test cases
	testCases := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Missing Authorization Header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid or missing token",
		},
		{
			name:           "Invalid Bearer Format",
			authHeader:     "InvalidFormat token123",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid or missing token",
		},
		{
			name:           "Empty Token",
			authHeader:     "Bearer ",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid or missing token",
		},
		{
			name:           "Invalid JWT Token",
			authHeader:     "Bearer invalid.jwt.token",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid or missing token",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest("POST", "/api/email/send", bytes.NewBuffer([]byte("{}")))
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Create a minimal server for testing
			srv := &Server{db: nil}

			// Apply JWT middleware to the handler
			handler := jwtMiddleware(http.HandlerFunc(srv.sendEmailHandler))
			handler.ServeHTTP(w, req)

			// Check status code
			if w.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, w.Code)
			}

			// Check error message
			var response map[string]string
			if err := json.NewDecoder(w.Body).Decode(&response); err == nil {
				if response["error"] != tc.expectedError {
					t.Errorf("Expected error '%s', got '%s'", tc.expectedError, response["error"])
				}
			}
		})
	}
}

// TestValidJWTAuthentication tests successful JWT authentication
func TestValidJWTAuthentication(t *testing.T) {
	// Set JWT secret for testing
	os.Setenv("JWT_SECRET", "test-secret-key-for-jwt-signing")
	defer os.Unsetenv("JWT_SECRET")

	// Generate a valid JWT token with both user_id and email claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": "123",
		"email":   "test@example.com",
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	})
	tokenString, err := token.SignedString([]byte("test-secret-key-for-jwt-signing"))
	if err != nil {
		t.Fatalf("Failed to generate JWT token: %v", err)
	}

	// Create request with valid token
	reqBody := `{"recipient":"test@example.com","subject":"Test","body":"Test body"}`
	req := httptest.NewRequest("POST", "/api/email/send", bytes.NewBuffer([]byte(reqBody)))
	req.Header.Set("Authorization", "Bearer "+tokenString)
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Create a minimal server for testing (no database needed for this test)
	srv := &Server{db: nil}

	// Apply JWT middleware to the handler
	handler := jwtMiddleware(http.HandlerFunc(srv.sendEmailHandler))
	handler.ServeHTTP(w, req)

	// Should not get 401 Unauthorized (might get other errors due to missing DB, but not auth error)
	if w.Code == http.StatusUnauthorized {
		t.Errorf("Expected not to get 401 Unauthorized with valid JWT token, got %d", w.Code)
	}
}

// TestGetUserIDFromContext tests the GetUserID helper function
func TestGetUserIDFromContext(t *testing.T) {
	// Test with valid user_id in context
	ctx := context.WithValue(context.Background(), UserIDKey, "test-user-123")
	userID, err := GetUserID(ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if userID != "test-user-123" {
		t.Errorf("Expected user_id 'test-user-123', got '%s'", userID)
	}

	// Test with missing user_id in context
	emptyCtx := context.Background()
	_, err = GetUserID(emptyCtx)
	if err == nil {
		t.Error("Expected error when user_id not found in context")
	}
}

// TestGetUserIDFromContextRequest tests the GetUserIDFromContext function
func TestGetUserIDFromContextRequest(t *testing.T) {
	// Test with valid user_id in request context
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(req.Context(), UserIDKey, "test-user-123")
	req = req.WithContext(ctx)

	userID, ok := GetUserIDFromContext(req)
	if !ok {
		t.Error("Expected to find user_id in context")
	}
	if userID != "test-user-123" {
		t.Errorf("Expected user_id 'test-user-123', got '%s'", userID)
	}

	// Test with missing user_id in request context
	emptyReq := httptest.NewRequest("GET", "/test", nil)
	_, ok = GetUserIDFromContext(emptyReq)
	if ok {
		t.Error("Expected not to find user_id in empty context")
	}
}
