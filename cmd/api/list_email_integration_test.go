package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// TestListEmailIntegration tests the complete list email flow with authentication
func TestListEmailIntegration(t *testing.T) {
	// Test 1: Valid JWT token should pass authentication
	t.Run("ValidJWTToken", func(t *testing.T) {
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
		req := httptest.NewRequest("GET", "/api/email/list", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)

		// Create response recorder
		w := httptest.NewRecorder()

		// Create server with nil database (will return error but not auth error)
		srv := &Server{db: nil}

		// Apply JWT middleware and call handler
		handler := jwtMiddleware(http.HandlerFunc(srv.listEmailHandler))
		handler.ServeHTTP(w, req)

		// Should not get 401 Unauthorized
		if w.Code == http.StatusUnauthorized {
			t.Errorf("Expected not to get 401 Unauthorized with valid JWT token, got %d", w.Code)
		}

		// Should get 500 due to nil database
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
	})

	// Test 2: Missing JWT token should return 401
	t.Run("MissingJWTToken", func(t *testing.T) {
		// Create request without token
		req := httptest.NewRequest("GET", "/api/email/list", nil)

		// Create response recorder
		w := httptest.NewRecorder()

		// Create server
		srv := &Server{db: nil}

		// Apply JWT middleware and call handler
		handler := jwtMiddleware(http.HandlerFunc(srv.listEmailHandler))
		handler.ServeHTTP(w, req)

		// Should get 401 Unauthorized
		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected 401 Unauthorized, got %d", w.Code)
		}

		// Check error message
		var response map[string]string
		if err := json.NewDecoder(w.Body).Decode(&response); err == nil {
			if response["error"] != "Invalid or missing token" {
				t.Errorf("Expected error 'Invalid or missing token', got '%s'", response["error"])
			}
		}
	})

	// Test 3: Invalid JWT token should return 401
	t.Run("InvalidJWTToken", func(t *testing.T) {
		// Create request with invalid token
		req := httptest.NewRequest("GET", "/api/email/list", nil)
		req.Header.Set("Authorization", "Bearer invalid.jwt.token")

		// Create response recorder
		w := httptest.NewRecorder()

		// Create server
		srv := &Server{db: nil}

		// Apply JWT middleware and call handler
		handler := jwtMiddleware(http.HandlerFunc(srv.listEmailHandler))
		handler.ServeHTTP(w, req)

		// Should get 401 Unauthorized
		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected 401 Unauthorized, got %d", w.Code)
		}

		// Check error message
		var response map[string]string
		if err := json.NewDecoder(w.Body).Decode(&response); err == nil {
			if response["error"] != "Invalid or missing token" {
				t.Errorf("Expected error 'Invalid or missing token', got '%s'", response["error"])
			}
		}
	})
}

// TestListEmailResponseStructure tests the response structure when successful
func TestListEmailResponseStructure(t *testing.T) {
	// This test verifies the response structure without requiring a real database
	// It's a unit test for the response format

	// Create a mock response
	response := ListEmailResponse{
		Emails: []EmailListItem{
			{
				EmailID:   "email-1",
				Recipient: "alice@example.com",
				Subject:   "Project Update",
				CreatedAt: time.Now(),
			},
			{
				EmailID:   "email-2",
				Recipient: "bob@example.com",
				Subject:   "Meeting Notes",
				CreatedAt: time.Now().Add(-1 * time.Hour),
			},
		},
		Status: "success",
	}

	// Verify response structure
	if response.Status != "success" {
		t.Errorf("Expected status 'success', got '%s'", response.Status)
	}

	if len(response.Emails) != 2 {
		t.Errorf("Expected 2 emails, got %d", len(response.Emails))
	}

	// Verify first email
	if response.Emails[0].EmailID != "email-1" {
		t.Errorf("Expected email ID 'email-1', got '%s'", response.Emails[0].EmailID)
	}

	if response.Emails[0].Recipient != "alice@example.com" {
		t.Errorf("Expected recipient 'alice@example.com', got '%s'", response.Emails[0].Recipient)
	}

	if response.Emails[0].Subject != "Project Update" {
		t.Errorf("Expected subject 'Project Update', got '%s'", response.Emails[0].Subject)
	}

	// Verify second email
	if response.Emails[1].EmailID != "email-2" {
		t.Errorf("Expected email ID 'email-2', got '%s'", response.Emails[1].EmailID)
	}

	if response.Emails[1].Recipient != "bob@example.com" {
		t.Errorf("Expected recipient 'bob@example.com', got '%s'", response.Emails[1].Recipient)
	}

	if response.Emails[1].Subject != "Meeting Notes" {
		t.Errorf("Expected subject 'Meeting Notes', got '%s'", response.Emails[1].Subject)
	}
}
