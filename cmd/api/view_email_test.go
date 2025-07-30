package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"secure-email-mvp/pkg/auth"

	"github.com/gorilla/mux"
)

// TestViewEmailAuthentication tests authentication requirements for the view endpoint
func TestViewEmailAuthentication(t *testing.T) {
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
			req := httptest.NewRequest("GET", "/api/email/view/"+tc.emailID, nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Create a minimal server for testing
			srv := &Server{db: nil}

			// Apply JWT middleware to the handler
			handler := jwtMiddleware(http.HandlerFunc(srv.viewEmailHandler))
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

// TestViewEmailWithValidToken tests successful view retrieval with valid JWT
func TestViewEmailWithValidToken(t *testing.T) {
	// Generate a valid JWT token
	token, err := auth.GenerateJWT("test-user-123")
	if err != nil {
		t.Fatalf("Failed to generate JWT token: %v", err)
	}

	// Create request with valid token
	req := httptest.NewRequest("GET", "/api/email/view/test-email-123", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Set up mux vars manually for testing
	req = mux.SetURLVars(req, map[string]string{"id": "test-email-123"})

	// Create response recorder
	w := httptest.NewRecorder()

	// Create a minimal server for testing (no database needed for this test)
	srv := &Server{db: nil}

	// Apply JWT middleware to the handler
	handler := jwtMiddleware(http.HandlerFunc(srv.viewEmailHandler))
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

// TestViewEmailMissingID tests the endpoint when email_id is missing from URL
func TestViewEmailMissingID(t *testing.T) {
	// Generate a valid JWT token
	token, err := auth.GenerateJWT("test-user-123")
	if err != nil {
		t.Fatalf("Failed to generate JWT token: %v", err)
	}

	// Create request with valid token but missing email_id
	req := httptest.NewRequest("GET", "/api/email/view/", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Create response recorder
	w := httptest.NewRecorder()

	// Create a minimal server for testing
	srv := &Server{db: nil}

	// Apply JWT middleware to the handler
	handler := jwtMiddleware(http.HandlerFunc(srv.viewEmailHandler))
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

// TestViewEmailDatabaseError tests error handling when database query fails
func TestViewEmailDatabaseError(t *testing.T) {
	// Create server with nil database to simulate database error
	srv := &Server{db: nil}

	// Generate JWT token
	token, err := auth.GenerateJWT("test-user-123")
	if err != nil {
		t.Fatalf("Failed to generate JWT token: %v", err)
	}

	// Create request
	req := httptest.NewRequest("GET", "/api/email/view/test-email-123", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Set up mux vars manually for testing
	req = mux.SetURLVars(req, map[string]string{"id": "test-email-123"})

	// Create response recorder
	w := httptest.NewRecorder()

	// Apply JWT middleware and call handler
	handler := jwtMiddleware(http.HandlerFunc(srv.viewEmailHandler))
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

// TestViewEmailResponseStructure tests the response structure when successful
func TestViewEmailResponseStructure(t *testing.T) {
	// This test verifies the response structure without requiring a real database
	// It's a unit test for the response format

	// Create a mock response
	response := ViewEmailResponse{
		EmailID:   "test-email-123",
		Recipient: "alice@example.com",
		Subject:   "Test Email Subject",
		Body:      "This is the decrypted body of the email.",
		CreatedAt: time.Now(),
		Status:    "success",
	}

	// Verify response structure
	if response.Status != "success" {
		t.Errorf("Expected status 'success', got '%s'", response.Status)
	}

	if response.EmailID != "test-email-123" {
		t.Errorf("Expected email ID 'test-email-123', got '%s'", response.EmailID)
	}

	if response.Recipient != "alice@example.com" {
		t.Errorf("Expected recipient 'alice@example.com', got '%s'", response.Recipient)
	}

	if response.Subject != "Test Email Subject" {
		t.Errorf("Expected subject 'Test Email Subject', got '%s'", response.Subject)
	}

	if response.Body != "This is the decrypted body of the email." {
		t.Errorf("Expected body 'This is the decrypted body of the email.', got '%s'", response.Body)
	}
}
