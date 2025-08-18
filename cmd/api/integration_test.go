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

	_ "modernc.org/sqlite"
)

func TestSignupLoginIntegration(t *testing.T) {
	// Set JWT secrets for testing
	os.Setenv("JWT_SECRET", "test-secret-key")
	os.Setenv("JWT_ACCESS_SECRET", "test-access-secret-key")

	// Setup in-memory database for testing
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal("Failed to open test database:", err)
	}
	defer db.Close()

	// Create users table with all required fields for signup handler
	_, err = db.Exec(`CREATE TABLE users (
		id INTEGER PRIMARY KEY,
		email TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		password_hash TEXT NOT NULL,
		totp_secret TEXT NOT NULL,
		fallback_email TEXT,
		fallback_token TEXT,
		fallback_confirmed BOOLEAN DEFAULT FALSE,
		fallback_token_expiration TIMESTAMP,
		failed_login_attempts INTEGER DEFAULT 0,
		last_failed_login TIMESTAMP,
		account_locked_until TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		t.Fatal("Failed to create users table:", err)
	}

	// Create refresh_tokens table for session management
	_, err = db.Exec(`CREATE TABLE refresh_tokens (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		token_hash TEXT NOT NULL,
		expires_at TIMESTAMP NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		last_used_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		is_revoked BOOLEAN DEFAULT FALSE
	)`)
	if err != nil {
		t.Fatal("Failed to create refresh_tokens table:", err)
	}

	// Create handlers
	signupHandler := signupHandlerFactory(db)
	loginHandler := loginHandler(db)
	confirmHandler := confirmFallbackHandlerFactory(db)

	// Test data
	testEmail := "integration@example.com"
	testPassword := "SecurePassword123!"

	// Step 1: Sign up a new user
	t.Run("Signup", func(t *testing.T) {
		signupReq := SignupRequest{
			Email:         testEmail,
			Password:      testPassword,
			FallbackEmail: "recovery@example.com",
		}

		body, err := json.Marshal(signupReq)
		if err != nil {
			t.Fatal(err)
		}

		req := httptest.NewRequest("POST", "/signup", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		signupHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d", http.StatusCreated, rr.Code)
		}

		var response SignupResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
			t.Fatal(err)
		}

		if response.Message != "User created" {
			t.Errorf("Expected message 'User created', got '%s'", response.Message)
		}
	})

	// Step 2: Confirm fallback email
	t.Run("Confirm_fallback", func(t *testing.T) {
		// Get the fallback token from the database
		var fallbackToken string
		var fallbackExpiration time.Time
		err := db.QueryRow("SELECT fallback_token, fallback_token_expiration FROM users WHERE email = ?", testEmail).Scan(&fallbackToken, &fallbackExpiration)
		if err != nil {
			t.Fatalf("Failed to get fallback token: %v", err)
		}
		if time.Now().After(fallbackExpiration) {
			t.Fatalf("Token should not be expired in test setup")
		}

		// Confirm fallback email
		req := httptest.NewRequest("GET", "/confirm-fallback?token="+fallbackToken, nil)
		rr := httptest.NewRecorder()

		confirmHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
		}

		var response map[string]string
		if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
			t.Fatal(err)
		}

		if response["message"] != "Fallback email confirmed successfully. You may now log in." {
			t.Errorf("Expected message 'Fallback email confirmed successfully. You may now log in.', got '%s'", response["message"])
		}
	})

	// Step 3: Login with the created user
	t.Run("Login", func(t *testing.T) {
		// Get the TOTP secret from the database to generate the correct code
		var totpSecret string
		err := db.QueryRow("SELECT totp_secret FROM users WHERE email = ?", testEmail).Scan(&totpSecret)
		if err != nil {
			t.Fatalf("Failed to get TOTP secret: %v", err)
		}

		// Generate the correct TOTP code for the current time
		totpCode, err := auth.GenerateTOTPCode(totpSecret)
		if err != nil {
			t.Fatalf("Failed to generate TOTP code: %v", err)
		}

		loginReq := LoginRequest{
			Email:    testEmail,
			Password: testPassword,
			TOTPCode: totpCode,
		}

		body, err := json.Marshal(loginReq)
		if err != nil {
			t.Fatal(err)
		}

		req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		loginHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
		}

		var response LoginResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
			t.Fatal(err)
		}

		// Check for access token
		if response.AccessToken == "" {
			t.Error("Expected non-empty access token, got empty")
		}
		if len(response.AccessToken) < 50 {
			t.Errorf("Expected access token to be longer, got length %d", len(response.AccessToken))
		}

		// Check for refresh token
		if response.RefreshToken == "" {
			t.Error("Expected non-empty refresh token, got empty")
		}
	})

	// Step 4: Try to login with wrong password
	t.Run("Login with wrong password", func(t *testing.T) {
		loginReq := LoginRequest{
			Email:    testEmail,
			Password: "wrongpassword",
			TOTPCode: "123456",
		}

		body, err := json.Marshal(loginReq)
		if err != nil {
			t.Fatal(err)
		}

		req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		loginHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rr.Code)
		}

		var errorResp map[string]string
		if err := json.Unmarshal(rr.Body.Bytes(), &errorResp); err != nil {
			t.Fatal(err)
		}

		if errorResp["error"] != "Invalid credentials" {
			t.Errorf("Expected error 'Invalid credentials', got '%s'", errorResp["error"])
		}
	})

	// Step 5: Try to signup with same email (should fail)
	t.Run("Duplicate signup", func(t *testing.T) {
		signupReq := SignupRequest{
			Email:         testEmail,
			Password:      testPassword,
			FallbackEmail: "recovery@example.com",
		}

		body, err := json.Marshal(signupReq)
		if err != nil {
			t.Fatal(err)
		}

		req := httptest.NewRequest("POST", "/signup", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		signupHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
		}

		var errorResp map[string]string
		if err := json.Unmarshal(rr.Body.Bytes(), &errorResp); err != nil {
			t.Fatal(err)
		}

		if errorResp["error"] != "User already exists" {
			t.Errorf("Expected error 'User already exists', got '%s'", errorResp["error"])
		}
	})
}
