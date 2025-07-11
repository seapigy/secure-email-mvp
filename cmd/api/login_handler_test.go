package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"golang.org/x/crypto/bcrypt"

	_ "modernc.org/sqlite"
)

func TestLoginHandler(t *testing.T) {
	// Set JWT_SECRET for testing
	os.Setenv("JWT_SECRET", "test-secret-key")

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
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		t.Fatal("Failed to create test table:", err)
	}

	// Create a test user with hashed password
	testEmail := "test@example.com"
	testPassword := "password123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal("Failed to hash password:", err)
	}

	// Insert test user with confirmed fallback
	_, err = db.Exec("INSERT INTO users (email, password, fallback_email, fallback_token, fallback_confirmed) VALUES (?, ?, ?, ?, ?)",
		testEmail, string(hashedPassword), "recovery@example.com", "test-token", true)
	if err != nil {
		t.Fatal("Failed to insert test user:", err)
	}

	// Create handler with test database
	handler := loginHandlerFactory(db)

	tests := []struct {
		name            string
		method          string
		requestBody     LoginRequest
		expectedStatus  int
		expectedError   string
		expectedMessage string
		expectToken     bool
	}{
		{
			name:   "Valid login",
			method: "POST",
			requestBody: LoginRequest{
				Email:    testEmail,
				Password: testPassword,
			},
			expectedStatus:  http.StatusOK,
			expectedMessage: "Login successful",
			expectToken:     true,
		},
		{
			name:   "Invalid password",
			method: "POST",
			requestBody: LoginRequest{
				Email:    testEmail,
				Password: "wrongpassword",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid email or password",
			expectToken:    false,
		},
		{
			name:   "User not found",
			method: "POST",
			requestBody: LoginRequest{
				Email:    "nonexistent@example.com",
				Password: "password123",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid email or password",
			expectToken:    false,
		},
		{
			name:   "Empty email",
			method: "POST",
			requestBody: LoginRequest{
				Email:    "",
				Password: "password123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Email and password are required",
			expectToken:    false,
		},
		{
			name:   "Empty password",
			method: "POST",
			requestBody: LoginRequest{
				Email:    testEmail,
				Password: "",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Email and password are required",
			expectToken:    false,
		},
		{
			name:   "Whitespace only email",
			method: "POST",
			requestBody: LoginRequest{
				Email:    "   ",
				Password: "password123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Email and password are required",
			expectToken:    false,
		},
		{
			name:   "Whitespace only password",
			method: "POST",
			requestBody: LoginRequest{
				Email:    testEmail,
				Password: "   ",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Email and password are required",
			expectToken:    false,
		},
		{
			name:   "Wrong HTTP method",
			method: "GET",
			requestBody: LoginRequest{
				Email:    testEmail,
				Password: testPassword,
			},
			expectedStatus: http.StatusMethodNotAllowed,
			expectToken:    false,
		},
		{
			name:   "Invalid JSON",
			method: "POST",
			requestBody: LoginRequest{
				Email:    testEmail,
				Password: testPassword,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid JSON format",
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
			req := httptest.NewRequest(tt.method, "/login", bytes.NewBuffer(body))
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

			// Check success message and token if expected
			if tt.expectedMessage != "" {
				var successResp LoginResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &successResp); err != nil {
					t.Fatal(err)
				}
				if successResp.Message != tt.expectedMessage {
					t.Errorf("Expected message '%s', got '%s'", tt.expectedMessage, successResp.Message)
				}

				// Check for token if expected
				if tt.expectToken {
					if successResp.Token == "" {
						t.Error("Expected non-empty JWT token, got empty")
					}
					if len(successResp.Token) < 50 {
						t.Errorf("Expected JWT token to be longer, got length %d", len(successResp.Token))
					}
				}
			}
		})
	}
}

func TestGetUserByEmail(t *testing.T) {
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
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		t.Fatal("Failed to create test table:", err)
	}

	// Test user not found
	user, err := getUserByEmail(db, "nonexistent@example.com")
	if err != sql.ErrNoRows {
		t.Errorf("Expected sql.ErrNoRows, got %v", err)
	}
	if user != nil {
		t.Error("Expected nil user, got non-nil")
	}

	// Insert a test user
	testEmail := "test@example.com"
	testPassword := "hashedpassword"
	_, err = db.Exec("INSERT INTO users (email, password, fallback_email, fallback_token, fallback_confirmed) VALUES (?, ?, ?, ?, ?)",
		testEmail, testPassword, "recovery@example.com", "test-token", true)
	if err != nil {
		t.Fatal("Failed to insert test user:", err)
	}

	// Test user found
	user, err = getUserByEmail(db, testEmail)
	if err != nil {
		t.Fatalf("getUserByEmail failed: %v", err)
	}
	if user == nil {
		t.Fatal("Expected non-nil user, got nil")
	}
	if user.Email != testEmail {
		t.Errorf("Expected email %s, got %s", testEmail, user.Email)
	}
	if user.Password != testPassword {
		t.Errorf("Expected password %s, got %s", testPassword, user.Password)
	}
}

func TestBcryptPasswordComparison(t *testing.T) {
	// Test password hashing and comparison
	password := "testpassword123"

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal("Failed to hash password:", err)
	}

	// Test correct password
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		t.Error("Password comparison failed for correct password")
	}

	// Test incorrect password
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte("wrongpassword"))
	if err == nil {
		t.Error("Password comparison should have failed for incorrect password")
	}
}
