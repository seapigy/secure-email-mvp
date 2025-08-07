package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"secure-email-mvp/pkg/auth"

	"github.com/pquerna/otp/totp"
	_ "modernc.org/sqlite"
)

func TestLoginHandler(t *testing.T) {
	// Set JWT secrets for testing
	os.Setenv("JWT_SECRET", "test-secret-key")
	os.Setenv("JWT_ACCESS_SECRET", "test-access-secret-key-for-testing-only")
	os.Setenv("JWT_REFRESH_SECRET", "test-refresh-secret-key-for-testing-only")

	// Setup in-memory database for testing
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal("Failed to open test database:", err)
	}
	defer db.Close()

	// Create users table with current schema
	_, err = db.Exec(`CREATE TABLE users (
		id TEXT PRIMARY KEY,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		totp_secret TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		t.Fatal("Failed to create test table:", err)
	}

	// Create refresh_tokens table
	_, err = db.Exec(`CREATE TABLE refresh_tokens (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		token_hash TEXT NOT NULL,
		expires_at TIMESTAMP NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		last_used_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		is_revoked BOOLEAN DEFAULT FALSE,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	)`)
	if err != nil {
		t.Fatal("Failed to create refresh_tokens table:", err)
	}

	// Create a test user with hashed password and TOTP secret
	testEmail := "test@securesystem.email"
	testPassword := "password123"
	hashedPassword, err := auth.HashPassword(testPassword, testEmail)
	if err != nil {
		t.Fatal("Failed to hash password:", err)
	}

	// Generate a valid TOTP secret
	totpSecret, err := auth.GenerateTOTPSecret()
	if err != nil {
		t.Fatal("Failed to generate TOTP secret:", err)
	}

	// Insert test user
	_, err = db.Exec("INSERT INTO users (id, email, password_hash, totp_secret) VALUES (?, ?, ?, ?)",
		"test-user-123", testEmail, hashedPassword, totpSecret)
	if err != nil {
		t.Fatal("Failed to insert test user:", err)
	}

	// Generate a valid TOTP code for testing
	validTOTP, err := totp.GenerateCode(totpSecret, time.Now())
	if err != nil {
		t.Fatal("Failed to generate TOTP code:", err)
	}

	// Create handler with test database
	handler := loginHandler(db)

	tests := []struct {
		name           string
		method         string
		requestBody    LoginRequest
		expectedStatus int
		expectedError  string
		expectToken    bool
	}{
		{
			name:   "Valid login",
			method: "POST",
			requestBody: LoginRequest{
				Email:    testEmail,
				Password: testPassword,
				TOTPCode: validTOTP,
			},
			expectedStatus: http.StatusOK,
			expectToken:    true,
		},
		{
			name:   "Invalid password",
			method: "POST",
			requestBody: LoginRequest{
				Email:    testEmail,
				Password: "wrongpassword",
				TOTPCode: "123456",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid credentials",
			expectToken:    false,
		},
		{
			name:   "User not found",
			method: "POST",
			requestBody: LoginRequest{
				Email:    "nonexistent@securesystem.email",
				Password: "password123",
				TOTPCode: "123456",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid credentials",
			expectToken:    false,
		},
		{
			name:   "Empty email",
			method: "POST",
			requestBody: LoginRequest{
				Email:    "",
				Password: "password123",
				TOTPCode: "123456",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Email, password, and TOTP code are required",
			expectToken:    false,
		},
		{
			name:   "Empty password",
			method: "POST",
			requestBody: LoginRequest{
				Email:    testEmail,
				Password: "",
				TOTPCode: "123456",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Email, password, and TOTP code are required",
			expectToken:    false,
		},
		{
			name:   "Empty TOTP code",
			method: "POST",
			requestBody: LoginRequest{
				Email:    testEmail,
				Password: testPassword,
				TOTPCode: "",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Email, password, and TOTP code are required",
			expectToken:    false,
		},
		{
			name:   "Wrong HTTP method",
			method: "GET",
			requestBody: LoginRequest{
				Email:    testEmail,
				Password: testPassword,
				TOTPCode: "123456",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid credentials",
			expectToken:    false,
		},
		{
			name:   "Invalid JSON",
			method: "POST",
			requestBody: LoginRequest{
				Email:    testEmail,
				Password: testPassword,
				TOTPCode: "123456",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request format",
			expectToken:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			var err error

			// Handle special case for invalid JSON test
			if tt.name == "Invalid JSON" {
				body = []byte(`{"email": "test@example.com", "password": "password123"`) // Missing closing brace
			} else {
				body, err = json.Marshal(tt.requestBody)
				if err != nil {
					t.Fatal(err)
				}
			}

			// Create request
			req := httptest.NewRequest(tt.method, "/api/auth/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler
			handler.ServeHTTP(rr, req)

			// Check status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
				t.Errorf("Response body: %s", rr.Body.String())
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
			if tt.expectToken {
				var successResp LoginResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &successResp); err != nil {
					t.Fatal(err)
				}

				// Check for access token if expected
				if successResp.AccessToken == "" {
					t.Error("Expected non-empty access token, got empty")
				}
				if len(successResp.AccessToken) < 50 {
					t.Errorf("Expected access token to be longer, got length %d", len(successResp.AccessToken))
				}

				// Check for refresh token
				if successResp.RefreshToken == "" {
					t.Error("Expected non-empty refresh token, got empty")
				}

				// Check token type
				if successResp.TokenType != "Bearer" {
					t.Errorf("Expected token type 'Bearer', got '%s'", successResp.TokenType)
				}
			}
		})
	}
}

// TestGetUserByEmail is removed as the function is no longer available
// The functionality is now handled by the auth package

// Password verification is tested as part of the login authentication flow
// in TestLoginHandler above, so no separate test is needed here.
