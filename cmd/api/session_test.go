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
)

// TestSessionManagement tests the complete session management flow
func TestSessionManagement(t *testing.T) {
	// Set up test environment
	os.Setenv("JWT_SECRET", "test-secret-key-for-testing-only")
	os.Setenv("JWT_ACCESS_SECRET", "test-access-secret-key-for-testing-only")
	os.Setenv("JWT_REFRESH_SECRET", "test-refresh-secret-key-for-testing-only")

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

	// Create test user with proper Argon2 hash
	userID := "test-user-123"
	email := "test@securesystem.email"
	testPassword := "test-password"
	hashedPassword, err := auth.HashPassword(testPassword, email)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Generate a valid TOTP secret
	totpSecret, err := auth.GenerateTOTPSecret()
	if err != nil {
		t.Fatalf("Failed to generate TOTP secret: %v", err)
	}

	_, err = db.Exec("INSERT INTO users (id, email, password_hash, totp_secret) VALUES (?, ?, ?, ?)",
		userID, email, hashedPassword, totpSecret)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Test 1: Login and get token pair
	t.Run("Login and get token pair", func(t *testing.T) {
		// Generate valid TOTP code
		validTOTP, err := totp.GenerateCode(totpSecret, time.Now())
		if err != nil {
			t.Fatalf("Failed to generate TOTP code: %v", err)
		}

		reqBody := map[string]string{
			"email":     email,
			"password":  "test-password",
			"totp_code": validTOTP,
		}
		reqBodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(reqBodyBytes))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		handler := loginHandler(db)
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response LoginResponse
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response.AccessToken == "" {
			t.Error("Expected access token, got empty")
		}
		if response.RefreshToken == "" {
			t.Error("Expected refresh token, got empty")
		}
		if response.TokenType != "Bearer" {
			t.Errorf("Expected token type 'Bearer', got '%s'", response.TokenType)
		}
		if response.ExpiresIn <= 0 {
			t.Error("Expected positive expiry time")
		}
		if response.UserID != userID {
			t.Errorf("Expected user ID '%s', got '%s'", userID, response.UserID)
		}
		if response.Email != email {
			t.Errorf("Expected email '%s', got '%s'", email, response.Email)
		}
	})

	// Test 2: Refresh token
	t.Run("Refresh token", func(t *testing.T) {
		// Create session manager and generate tokens
		sessionManager, err := auth.NewSessionManager()
		if err != nil {
			t.Fatalf("Failed to create session manager: %v", err)
		}

		tokenPair, err := sessionManager.GenerateTokenPair(userID, email, db)
		if err != nil {
			t.Fatalf("Failed to generate token pair: %v", err)
		}

		// Test refresh
		reqBody := map[string]string{
			"refresh_token": tokenPair.RefreshToken,
		}
		reqBodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/auth/refresh", bytes.NewBuffer(reqBodyBytes))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		handler := refreshHandler(db)
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response RefreshResponse
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response.AccessToken == "" {
			t.Error("Expected access token, got empty")
		}
		if response.TokenType != "Bearer" {
			t.Errorf("Expected token type 'Bearer', got '%s'", response.TokenType)
		}
		if response.ExpiresIn <= 0 {
			t.Error("Expected positive expiry time")
		}
	})

	// Test 3: Logout
	t.Run("Logout", func(t *testing.T) {
		// Create session manager and generate tokens
		sessionManager, err := auth.NewSessionManager()
		if err != nil {
			t.Fatalf("Failed to create session manager: %v", err)
		}

		tokenPair, err := sessionManager.GenerateTokenPair(userID, email, db)
		if err != nil {
			t.Fatalf("Failed to generate token pair: %v", err)
		}

		// Test logout
		reqBody := map[string]string{
			"refresh_token": tokenPair.RefreshToken,
		}
		reqBodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/auth/logout", bytes.NewBuffer(reqBodyBytes))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		handler := logoutHandler(db)
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response LogoutResponse
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response.Message != "Successfully logged out" {
			t.Errorf("Expected message 'Successfully logged out', got '%s'", response.Message)
		}

		// Verify token is revoked by trying to refresh it
		refreshReqBody := map[string]string{
			"refresh_token": tokenPair.RefreshToken,
		}
		refreshReqBodyBytes, _ := json.Marshal(refreshReqBody)

		refreshReq := httptest.NewRequest("POST", "/api/auth/refresh", bytes.NewBuffer(refreshReqBodyBytes))
		refreshReq.Header.Set("Content-Type", "application/json")

		refreshW := httptest.NewRecorder()
		refreshHandler(db).ServeHTTP(refreshW, refreshReq)

		if refreshW.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401 after logout, got %d", refreshW.Code)
		}
	})

	// Test 4: JWT middleware
	t.Run("JWT middleware", func(t *testing.T) {
		// Create session manager and generate access token
		sessionManager, err := auth.NewSessionManager()
		if err != nil {
			t.Fatalf("Failed to create session manager: %v", err)
		}

		accessToken, err := sessionManager.GenerateAccessToken(userID, email)
		if err != nil {
			t.Fatalf("Failed to generate access token: %v", err)
		}

		// Test protected endpoint with valid token
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)

		w := httptest.NewRecorder()
		handler := EnhancedJWTMiddleware(db)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := GetUserIDFromContext(r)
			if !ok {
				http.Error(w, "User ID not found", http.StatusInternalServerError)
				return
			}
			email, ok := GetEmailFromContext(r)
			if !ok {
				http.Error(w, "Email not found", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"user_id":"` + userID + `","email":"` + email + `"}`))
		}))
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Test protected endpoint without token
		reqNoToken := httptest.NewRequest("GET", "/protected", nil)
		wNoToken := httptest.NewRecorder()
		handler.ServeHTTP(wNoToken, reqNoToken)

		if wNoToken.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401 without token, got %d", wNoToken.Code)
		}
	})
}

// TestSessionManager tests the session manager directly
func TestSessionManager(t *testing.T) {
	// Set up test environment
	os.Setenv("JWT_SECRET", "test-secret-key-for-testing-only")

	// Create test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Apply schema
	schema := `
	CREATE TABLE refresh_tokens (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		token_hash TEXT NOT NULL,
		expires_at TIMESTAMP NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		last_used_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		is_revoked BOOLEAN DEFAULT FALSE
	);
	`
	_, err = db.Exec(schema)
	if err != nil {
		t.Fatalf("Failed to apply schema: %v", err)
	}

	// Create session manager
	sessionManager, err := auth.NewSessionManager()
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}

	userID := "test-user-123"
	email := "test@securesystem.email"

	// Test token generation
	t.Run("Generate token pair", func(t *testing.T) {
		tokenPair, err := sessionManager.GenerateTokenPair(userID, email, db)
		if err != nil {
			t.Fatalf("Failed to generate token pair: %v", err)
		}

		if tokenPair.AccessToken == "" {
			t.Error("Expected access token, got empty")
		}
		if tokenPair.RefreshToken == "" {
			t.Error("Expected refresh token, got empty")
		}
		if tokenPair.TokenType != "Bearer" {
			t.Errorf("Expected token type 'Bearer', got '%s'", tokenPair.TokenType)
		}
		if tokenPair.ExpiresIn <= 0 {
			t.Error("Expected positive expiry time")
		}
	})

	// Test access token validation
	t.Run("Validate access token", func(t *testing.T) {
		accessToken, err := sessionManager.GenerateAccessToken(userID, email)
		if err != nil {
			t.Fatalf("Failed to generate access token: %v", err)
		}

		claims, err := sessionManager.ValidateAccessToken(accessToken)
		if err != nil {
			t.Fatalf("Failed to validate access token: %v", err)
		}

		if claims.UserID != userID {
			t.Errorf("Expected user ID '%s', got '%s'", userID, claims.UserID)
		}
		if claims.Email != email {
			t.Errorf("Expected email '%s', got '%s'", email, claims.Email)
		}
	})

	// Test refresh token validation
	t.Run("Validate refresh token", func(t *testing.T) {
		tokenPair, err := sessionManager.GenerateTokenPair(userID, email, db)
		if err != nil {
			t.Fatalf("Failed to generate token pair: %v", err)
		}

		validatedUserID, err := sessionManager.ValidateRefreshToken(tokenPair.RefreshToken, db)
		if err != nil {
			t.Fatalf("Failed to validate refresh token: %v", err)
		}

		if validatedUserID != userID {
			t.Errorf("Expected user ID '%s', got '%s'", userID, validatedUserID)
		}
	})

	// Test token revocation
	t.Run("Revoke refresh token", func(t *testing.T) {
		tokenPair, err := sessionManager.GenerateTokenPair(userID, email, db)
		if err != nil {
			t.Fatalf("Failed to generate token pair: %v", err)
		}

		// Revoke token
		err = sessionManager.RevokeRefreshToken(tokenPair.RefreshToken, db)
		if err != nil {
			t.Fatalf("Failed to revoke refresh token: %v", err)
		}

		// Try to validate revoked token
		_, err = sessionManager.ValidateRefreshToken(tokenPair.RefreshToken, db)
		if err == nil {
			t.Error("Expected error when validating revoked token")
		}
	})
}
