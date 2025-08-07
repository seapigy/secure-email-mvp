package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"secure-email-mvp/pkg/auth"
)

func TestMeHandler(t *testing.T) {
	// Set up test environment
	os.Setenv("JWT_SECRET", "test-secret-key-for-me-handler")
	os.Setenv("JWT_ACCESS_SECRET", "test-access-secret-key-for-me-handler")
	os.Setenv("JWT_REFRESH_SECRET", "test-refresh-secret-key-for-me-handler")

	// Create test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Apply schema
	schema := `
	CREATE TABLE users (
		id TEXT PRIMARY KEY,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		totp_secret TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE TABLE refresh_tokens (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		token_hash TEXT NOT NULL,
		expires_at TIMESTAMP NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		last_used_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		is_revoked BOOLEAN DEFAULT FALSE,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);
	`
	_, err = db.Exec(schema)
	if err != nil {
		t.Fatalf("Failed to apply schema: %v", err)
	}

	// Create test user
	userID := "test-user-123"
	email := "test@securesystem.email"
	passwordHash := "test-hash"
	totpSecret := "test-totp-secret"

	_, err = db.Exec("INSERT INTO users (id, email, password_hash, totp_secret) VALUES (?, ?, ?, ?)",
		userID, email, passwordHash, totpSecret)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create session manager and generate access token
	sessionManager, err := auth.NewSessionManager()
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}

	accessToken, err := sessionManager.GenerateAccessToken(userID, email)
	if err != nil {
		t.Fatalf("Failed to generate access token: %v", err)
	}

	// Test cases
	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedUserID string
		expectedEmail  string
		expectError    bool
	}{
		{
			name:           "Valid token",
			authHeader:     "Bearer " + accessToken,
			expectedStatus: http.StatusOK,
			expectedUserID: userID,
			expectedEmail:  email,
			expectError:    false,
		},
		{
			name:           "No authorization header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectError:    true,
		},
		{
			name:           "Invalid authorization format",
			authHeader:     "Invalid " + accessToken,
			expectedStatus: http.StatusUnauthorized,
			expectError:    true,
		},
		{
			name:           "Empty token",
			authHeader:     "Bearer ",
			expectedStatus: http.StatusUnauthorized,
			expectError:    true,
		},
		{
			name:           "Invalid token",
			authHeader:     "Bearer invalid-token",
			expectedStatus: http.StatusUnauthorized,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest("GET", "/api/auth/me", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Apply JWT middleware and call me handler
			handler := EnhancedJWTMiddleware(db)(meHandler())
			handler.ServeHTTP(w, req)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Check response
			if tt.expectError {
				var errorResp map[string]string
				if err := json.Unmarshal(w.Body.Bytes(), &errorResp); err != nil {
					t.Fatalf("Failed to parse error response: %v", err)
				}
				if errorResp["error"] == "" {
					t.Error("Expected error message, got empty")
				}
			} else {
				var response MeResponse
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to parse response: %v", err)
				}

				if response.UserID != tt.expectedUserID {
					t.Errorf("Expected user ID '%s', got '%s'", tt.expectedUserID, response.UserID)
				}
				if response.Email != tt.expectedEmail {
					t.Errorf("Expected email '%s', got '%s'", tt.expectedEmail, response.Email)
				}
			}
		})
	}
}

func TestMeHandlerWithoutMiddleware(t *testing.T) {
	// Test that me handler requires middleware
	req := httptest.NewRequest("GET", "/api/auth/me", nil)
	w := httptest.NewRecorder()

	// Call me handler directly without middleware
	handler := meHandler()
	handler.ServeHTTP(w, req)

	// Should return internal server error because no user context
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	var errorResp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &errorResp); err != nil {
		t.Fatalf("Failed to parse error response: %v", err)
	}
	if errorResp["error"] != "User information not found in context" {
		t.Errorf("Expected specific error message, got '%s'", errorResp["error"])
	}
}
