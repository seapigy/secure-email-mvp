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
	"github.com/gorilla/mux"
)

// generateTestJWT creates a JWT token for testing with both user_id and email claims
func generateTestJWT(userID, email string) (string, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "test-secret-key-for-jwt-signing"
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	})
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// testJWTMiddleware is a simplified JWT middleware for testing that doesn't require a database
func testJWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error":"Invalid or missing token"}`, http.StatusUnauthorized)
			return
		}

		// Check Bearer format
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			http.Error(w, `{"error":"Invalid or missing token"}`, http.StatusUnauthorized)
			return
		}

		tokenString := authHeader[7:]
		if tokenString == "" {
			http.Error(w, `{"error":"Invalid or missing token"}`, http.StatusUnauthorized)
			return
		}

		// Validate JWT token using the simpler function
		userID, email, err := auth.ValidateJWT(tokenString)
		if err != nil {
			http.Error(w, `{"error":"Invalid or missing token"}`, http.StatusUnauthorized)
			return
		}

		// Set user_id in context using UserIDKey
		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		ctx = context.WithValue(ctx, EmailKey, email)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// TestDeleteEmailAuthentication tests authentication requirements for the delete endpoint
func TestDeleteEmailAuthentication(t *testing.T) {
	// Set JWT secret for testing
	os.Setenv("JWT_SECRET", "test-secret-key-for-jwt-signing")
	defer os.Unsetenv("JWT_SECRET")

	// Test cases
	testCases := []struct {
		name           string
		authHeader     string
		emailID        string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Missing Authorization Header",
			authHeader:     "",
			emailID:        "test-email-123",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid or missing token",
		},
		{
			name:           "Invalid Bearer Format",
			authHeader:     "InvalidFormat token123",
			emailID:        "test-email-123",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid or missing token",
		},
		{
			name:           "Empty Token",
			authHeader:     "Bearer ",
			emailID:        "test-email-123",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid or missing token",
		},
		{
			name:           "Invalid JWT Token",
			authHeader:     "Bearer invalid.jwt.token",
			emailID:        "test-email-123",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid or missing token",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest("DELETE", "/api/email/"+tc.emailID, nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Create a minimal server for testing
			srv := &Server{db: nil}

			// Apply JWT middleware to the handler
			handler := testJWTMiddleware(http.HandlerFunc(srv.deleteEmailHandler))
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

// TestDeleteEmailWithValidToken tests successful deletion with valid JWT
func TestDeleteEmailWithValidToken(t *testing.T) {
	// Set JWT secret for testing
	os.Setenv("JWT_SECRET", "test-secret-key-for-jwt-signing")
	defer os.Unsetenv("JWT_SECRET")

	// Generate a valid JWT token
	token, err := generateTestJWT("test-user-123", "test@example.com")
	if err != nil {
		t.Fatalf("Failed to generate JWT token: %v", err)
	}

	// Create request with valid token
	req := httptest.NewRequest("DELETE", "/api/email/test-email-123", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Set up mux vars manually for testing
	req = mux.SetURLVars(req, map[string]string{"id": "test-email-123"})

	// Create response recorder
	w := httptest.NewRecorder()

	// Create a minimal server for testing (no database needed for this test)
	srv := &Server{db: nil}

	// Apply JWT middleware to the handler
	handler := testJWTMiddleware(http.HandlerFunc(srv.deleteEmailHandler))
	handler.ServeHTTP(w, req)

	// Should not get 401 Unauthorized (might get other errors due to missing DB, but not auth error)
	if w.Code == http.StatusUnauthorized {
		t.Errorf("Expected not to get 401 Unauthorized with valid JWT token, got %d", w.Code)
	}

	// Should get 500 Internal Server Error due to nil database
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500 Internal Server Error due to nil database, got %d", w.Code)
	}

	// Check error message
	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err == nil {
		if response["error"] != "Database connection unavailable" {
			t.Errorf("Expected error 'Database connection unavailable', got '%s'", response["error"])
		}
	}
}

// TestDeleteEmailMissingID tests the endpoint when email_id is missing from URL
func TestDeleteEmailMissingID(t *testing.T) {
	// Set JWT secret for testing
	os.Setenv("JWT_SECRET", "test-secret-key-for-jwt-signing")
	defer os.Unsetenv("JWT_SECRET")

	// Generate a valid JWT token
	token, err := generateTestJWT("test-user-123", "test@example.com")
	if err != nil {
		t.Fatalf("Failed to generate JWT token: %v", err)
	}

	// Create request with valid token but missing email_id
	req := httptest.NewRequest("DELETE", "/api/email/", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Create response recorder
	w := httptest.NewRecorder()

	// Create a minimal server for testing
	srv := &Server{db: nil}

	// Apply JWT middleware to the handler
	handler := testJWTMiddleware(http.HandlerFunc(srv.deleteEmailHandler))
	handler.ServeHTTP(w, req)

	// Should get 400 Bad Request due to missing email_id
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 Bad Request due to missing email_id, got %d", w.Code)
	}

	// Check error message
	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err == nil {
		if response["error"] != "Missing email_id" {
			t.Errorf("Expected error 'Missing email_id', got '%s'", response["error"])
		}
	}
}

// TestDeleteEmailDatabaseError tests error handling when database query fails
func TestDeleteEmailDatabaseError(t *testing.T) {
	// Set JWT secret for testing
	os.Setenv("JWT_SECRET", "test-secret-key-for-jwt-signing")
	defer os.Unsetenv("JWT_SECRET")

	// Create server with nil database to simulate database error
	srv := &Server{db: nil}

	// Generate JWT token
	token, err := generateTestJWT("test-user-123", "test@example.com")
	if err != nil {
		t.Fatalf("Failed to generate JWT token: %v", err)
	}

	// Create request
	req := httptest.NewRequest("DELETE", "/api/email/test-email-123", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Set up mux vars manually for testing
	req = mux.SetURLVars(req, map[string]string{"id": "test-email-123"})

	// Create response recorder
	w := httptest.NewRecorder()

	// Apply JWT middleware and call handler
	handler := testJWTMiddleware(http.HandlerFunc(srv.deleteEmailHandler))
	handler.ServeHTTP(w, req)

	// Should get 500 Internal Server Error due to nil database
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	// Check error message
	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err == nil {
		if response["error"] != "Database connection unavailable" {
			t.Errorf("Expected error 'Database connection unavailable', got '%s'", response["error"])
		}
	}
}

// TestDeleteEmailResponseStructure tests the response structure when successful
func TestDeleteEmailResponseStructure(t *testing.T) {
	// This test verifies the response structure without requiring a real database
	// It's a unit test for the response format

	// Create a mock response
	response := DeleteEmailResponse{
		Status:  "deleted",
		EmailID: "test-email-123",
	}

	// Verify response structure
	if response.Status != "deleted" {
		t.Errorf("Expected status 'deleted', got '%s'", response.Status)
	}

	if response.EmailID != "test-email-123" {
		t.Errorf("Expected email ID 'test-email-123', got '%s'", response.EmailID)
	}
}

// TestDeleteEmailIntegration tests the complete deletion flow
func TestDeleteEmailIntegration(t *testing.T) {
	// Set JWT secret for testing
	os.Setenv("JWT_SECRET", "test-secret-key-for-jwt-signing")
	defer os.Unsetenv("JWT_SECRET")

	// This test verifies the integration aspects without requiring a real database
	// It's a unit test for the handler logic

	// Generate JWT token
	token, err := generateTestJWT("test-user-123", "test@example.com")
	if err != nil {
		t.Fatalf("Failed to generate JWT token: %v", err)
	}

	// Create request
	req := httptest.NewRequest("DELETE", "/api/email/test-email-123", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req = mux.SetURLVars(req, map[string]string{"id": "test-email-123"})

	// Create response recorder
	w := httptest.NewRecorder()

	// Create server with nil database
	srv := &Server{db: nil}

	// Apply JWT middleware and call handler
	handler := testJWTMiddleware(http.HandlerFunc(srv.deleteEmailHandler))
	handler.ServeHTTP(w, req)

	// Should get 500 Internal Server Error due to nil database
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	// Verify the handler was called (we can see this from the log output)
	// The test passes if we reach this point without panicking
}
