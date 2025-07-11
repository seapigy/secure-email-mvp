package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "modernc.org/sqlite"
)

func TestSignupHandler(t *testing.T) {
	// Setup in-memory database for testing
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal("Failed to open test database:", err)
	}
	defer db.Close()

	// Create users table
	_, err = db.Exec(`CREATE TABLE users (
		id INTEGER PRIMARY KEY,
		email TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		fallback_email TEXT,
		fallback_token TEXT,
		fallback_confirmed BOOLEAN DEFAULT FALSE,
		fallback_token_expiration TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		t.Fatal("Failed to create test table:", err)
	}

	// Create handler with test database
	handler := signupHandlerFactory(db)

	tests := []struct {
		name            string
		method          string
		requestBody     SignupRequest
		expectedStatus  int
		expectedError   string
		expectedMessage string
	}{
		{
			name:   "Valid signup request",
			method: "POST",
			requestBody: SignupRequest{
				Email:         "test@example.com",
				Password:      "password123",
				FallbackEmail: "recovery@example.com",
			},
			expectedStatus:  http.StatusCreated,
			expectedMessage: "User created",
		},
		{
			name:   "Invalid email format",
			method: "POST",
			requestBody: SignupRequest{
				Email:         "invalid-email",
				Password:      "password123",
				FallbackEmail: "recovery@example.com",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid email format",
		},
		{
			name:   "Password too short",
			method: "POST",
			requestBody: SignupRequest{
				Email:         "test@example.com",
				Password:      "short",
				FallbackEmail: "recovery@example.com",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Password must be at least 8 characters long",
		},
		{
			name:   "Wrong HTTP method",
			method: "GET",
			requestBody: SignupRequest{
				Email:         "test@example.com",
				Password:      "password123",
				FallbackEmail: "recovery@example.com",
			},
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:   "Empty email",
			method: "POST",
			requestBody: SignupRequest{
				Email:         "",
				Password:      "password123",
				FallbackEmail: "recovery@example.com",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid email format",
		},
		{
			name:   "Empty password",
			method: "POST",
			requestBody: SignupRequest{
				Email:         "test@example.com",
				Password:      "",
				FallbackEmail: "recovery@example.com",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Password must be at least 8 characters long",
		},
		{
			name:   "Duplicate email",
			method: "POST",
			requestBody: SignupRequest{
				Email:         "duplicate@example.com",
				Password:      "password123",
				FallbackEmail: "recovery@example.com",
			},
			expectedStatus:  http.StatusCreated,
			expectedMessage: "User created",
		},
		{
			name:   "Duplicate email - second attempt",
			method: "POST",
			requestBody: SignupRequest{
				Email:         "duplicate@example.com",
				Password:      "password123",
				FallbackEmail: "recovery@example.com",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "User already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request body
			body, err := json.Marshal(tt.requestBody)
			if err != nil {
				t.Fatal(err)
			}

			// Create request
			req := httptest.NewRequest(tt.method, "/signup", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler
			handler.ServeHTTP(rr, req)

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

			// Check success message if expected
			if tt.expectedMessage != "" {
				var successResp SignupResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &successResp); err != nil {
					t.Fatal(err)
				}
				if successResp.Message != tt.expectedMessage {
					t.Errorf("Expected message '%s', got '%s'", tt.expectedMessage, successResp.Message)
				}
			}
		})
	}
}

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		email    string
		expected bool
	}{
		{"test@example.com", true},
		{"user.name@domain.co.uk", true},
		{"user+tag@example.org", true},
		{"invalid-email", false},
		{"@example.com", false},
		{"test@", false},
		{"", false},
		{"test@example", false},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			result := isValidEmail(tt.email)
			if result != tt.expected {
				t.Errorf("isValidEmail(%s) = %v, expected %v", tt.email, result, tt.expected)
			}
		})
	}
}

func TestIsValidPassword(t *testing.T) {
	tests := []struct {
		password string
		expected bool
	}{
		{"password123", true},
		{"12345678", true},
		{"short", false},
		{"", false},
		{"1234567", false},
		{"   password123   ", true}, // Should trim whitespace
	}

	for _, tt := range tests {
		t.Run(tt.password, func(t *testing.T) {
			result := isValidPassword(tt.password)
			if result != tt.expected {
				t.Errorf("isValidPassword(%s) = %v, expected %v", tt.password, result, tt.expected)
			}
		})
	}
}

func TestCreateUser(t *testing.T) {
	// Setup in-memory database for testing
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal("Failed to open test database:", err)
	}
	defer db.Close()

	// Create users table
	_, err = db.Exec(`CREATE TABLE users (
		id INTEGER PRIMARY KEY,
		email TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		fallback_email TEXT,
		fallback_token TEXT,
		fallback_confirmed BOOLEAN DEFAULT FALSE,
		fallback_token_expiration TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		t.Fatal("Failed to create test table:", err)
	}

	// Test creating a user
	email := "test@example.com"
	hashedPassword := "$2a$10$hashedpasswordstring"

	err = createUser(db, email, hashedPassword)
	if err != nil {
		t.Fatalf("createUser failed: %v", err)
	}

	// Verify user was created
	var storedEmail, storedPassword string
	err = db.QueryRow("SELECT email, password FROM users WHERE email = ?", email).Scan(&storedEmail, &storedPassword)
	if err != nil {
		t.Fatalf("Failed to query created user: %v", err)
	}

	if storedEmail != email {
		t.Errorf("Expected email %s, got %s", email, storedEmail)
	}
	if storedPassword != hashedPassword {
		t.Errorf("Expected password %s, got %s", hashedPassword, storedPassword)
	}

	// Test duplicate user creation
	err = createUser(db, email, hashedPassword)
	if err == nil {
		t.Error("Expected error for duplicate user")
	}
}
