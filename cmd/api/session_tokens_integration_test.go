package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"secure-email-mvp/pkg/sessiontokens"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"

	_ "modernc.org/sqlite"
)

func TestSessionTokensIntegration(t *testing.T) {
	// Setup test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Apply session tokens migration
	migration := `
		CREATE TABLE IF NOT EXISTS emails (
			email_id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			one_time_link_only BOOLEAN DEFAULT FALSE
		);
		
		CREATE TABLE IF NOT EXISTS email_sessions (
			session_id TEXT PRIMARY KEY,
			email_id TEXT NOT NULL,
			token_hash TEXT NOT NULL,
			expires_at DATETIME NOT NULL,
			used BOOLEAN DEFAULT FALSE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			user_agent TEXT,
			ip_address TEXT
		);
		
		CREATE INDEX IF NOT EXISTS idx_email_sessions_email_id ON email_sessions(email_id);
		CREATE INDEX IF NOT EXISTS idx_email_sessions_token_hash ON email_sessions(token_hash);
	`

	if _, err := db.Exec(migration); err != nil {
		t.Fatalf("Failed to apply migration: %v", err)
	}

	// Initialize session token service (use mock for testing)
	sessionTokenSvc := sessiontokens.NewMockSessionTokenService()

	// Create test email
	emailID := "test-email-123"
	userID := "test-user-456"
	_, err = db.Exec("INSERT INTO emails (email_id, user_id) VALUES (?, ?)", emailID, userID)
	if err != nil {
		t.Fatalf("Failed to create test email: %v", err)
	}

	t.Run("Generate and validate session token", func(t *testing.T) {
		// Generate session token
		sessionToken, err := sessionTokenSvc.GenerateSessionToken(emailID, "Mozilla/5.0", "192.168.1.1")
		if err != nil {
			t.Fatalf("Failed to generate session token: %v", err)
		}

		if sessionToken == "" {
			t.Error("Generated session token should not be empty")
		}

		// Validate session token
		isValid, err := sessionTokenSvc.ValidateSessionToken(emailID, sessionToken)
		if err != nil {
			t.Fatalf("Failed to validate session token: %v", err)
		}

		if !isValid {
			t.Error("Generated session token should be valid")
		}
	})

	t.Run("One-time link functionality", func(t *testing.T) {
		// Set email to one-time link mode
		_, err := db.Exec("UPDATE emails SET one_time_link_only = TRUE WHERE email_id = ?", emailID)
		if err != nil {
			t.Fatalf("Failed to update email to one-time link mode: %v", err)
		}

		// Generate session token
		sessionToken, err := sessionTokenSvc.GenerateSessionToken(emailID, "Chrome/91.0", "192.168.1.2")
		if err != nil {
			t.Fatalf("Failed to generate session token: %v", err)
		}

		// Validate session token (should be valid)
		isValid, err := sessionTokenSvc.ValidateSessionToken(emailID, sessionToken)
		if err != nil {
			t.Fatalf("Failed to validate session token: %v", err)
		}

		if !isValid {
			t.Error("Session token should be valid before marking as used")
		}

		// Mark session token as used
		err = sessionTokenSvc.MarkSessionTokenUsed(emailID, sessionToken)
		if err != nil {
			t.Fatalf("Failed to mark session token as used: %v", err)
		}

		// Validate session token again (should be invalid now)
		isValid, err = sessionTokenSvc.ValidateSessionToken(emailID, sessionToken)
		if err != nil {
			t.Fatalf("Failed to validate session token: %v", err)
		}

		if isValid {
			t.Error("Session token should be invalid after being marked as used")
		}
	})

	t.Run("Expired session token", func(t *testing.T) {
		// Create an expired session manually using mock
		expiredToken := "expired-token"
		expiredSession := &sessiontokens.SessionInfo{
			SessionID: "expired-session",
			EmailID:   emailID,
			TokenHash: "expired-hash",
			ExpiresAt: time.Now().Add(-1 * time.Hour),
			Used:      false,
			CreatedAt: time.Now().Add(-2 * time.Hour),
			UserAgent: "Safari/14.0",
			IPAddress: "192.168.1.3",
		}

		sessionTokenSvc.SetSession(emailID, expiredToken, expiredSession)

		// Validate expired session token (should be invalid)
		isValid, err := sessionTokenSvc.ValidateSessionToken(emailID, expiredToken)
		if err != nil {
			t.Fatalf("Failed to validate expired session token: %v", err)
		}

		if isValid {
			t.Error("Expired session token should be invalid")
		}
	})

	t.Run("Cleanup expired sessions", func(t *testing.T) {
		// Create both valid and expired sessions
		validToken, _ := sessionTokenSvc.GenerateSessionToken(emailID, "Firefox/89.0", "192.168.1.4")

		expiredToken := "another-expired-token"
		expiredSession2 := &sessiontokens.SessionInfo{
			SessionID: "another-expired-session",
			EmailID:   emailID,
			TokenHash: "another-expired-hash",
			ExpiresAt: time.Now().Add(-1 * time.Hour),
			Used:      false,
			CreatedAt: time.Now().Add(-2 * time.Hour),
			UserAgent: "Edge/91.0",
			IPAddress: "192.168.1.5",
		}

		sessionTokenSvc.SetSession(emailID, expiredToken, expiredSession2)

		// Verify both sessions exist before cleanup
		validBefore, _ := sessionTokenSvc.ValidateSessionToken(emailID, validToken)
		expiredBefore, _ := sessionTokenSvc.ValidateSessionToken(emailID, expiredToken)

		if !validBefore {
			t.Error("Valid token should be valid before cleanup")
		}
		if expiredBefore {
			t.Error("Expired token should not be valid before cleanup")
		}

		// Run cleanup
		err = sessionTokenSvc.CleanupExpiredSessions()
		if err != nil {
			t.Fatalf("Failed to cleanup expired sessions: %v", err)
		}

		// Verify valid session still exists, expired session is removed
		validAfter, _ := sessionTokenSvc.ValidateSessionToken(emailID, validToken)
		expiredAfter, _ := sessionTokenSvc.ValidateSessionToken(emailID, expiredToken)

		if !validAfter {
			t.Error("Valid token should still be valid after cleanup")
		}
		if expiredAfter {
			t.Error("Expired token should not be valid after cleanup")
		}
	})
}

func TestSessionTokensWithServer(t *testing.T) {
	// Setup test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Set JWT secret for testing
	os.Setenv("JWT_SECRET", "test-secret-key-for-jwt-signing")
	defer os.Unsetenv("JWT_SECRET")

	// Apply migrations
	migration := `
		CREATE TABLE IF NOT EXISTS users (
			user_id TEXT PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			totp_secret TEXT NOT NULL
		);
		
		CREATE TABLE IF NOT EXISTS emails (
			email_id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			one_time_link_only BOOLEAN DEFAULT FALSE
		);
		
		CREATE TABLE IF NOT EXISTS email_sessions (
			session_id TEXT PRIMARY KEY,
			email_id TEXT NOT NULL,
			token_hash TEXT NOT NULL,
			expires_at DATETIME NOT NULL,
			used BOOLEAN DEFAULT FALSE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			user_agent TEXT,
			ip_address TEXT
		);
		
		CREATE INDEX IF NOT EXISTS idx_email_sessions_email_id ON email_sessions(email_id);
		CREATE INDEX IF NOT EXISTS idx_email_sessions_token_hash ON email_sessions(token_hash);
		CREATE INDEX IF NOT EXISTS idx_email_sessions_expires_at ON email_sessions(expires_at);
		CREATE INDEX IF NOT EXISTS idx_email_sessions_used ON email_sessions(used);
		
		CREATE TABLE IF NOT EXISTS audit_log (
			id TEXT PRIMARY KEY,
			user_id TEXT,
			email_id TEXT,
			event_type TEXT NOT NULL,
			event_data TEXT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			ip_address TEXT,
			user_agent TEXT
		);
	`

	if _, err := db.Exec(migration); err != nil {
		t.Fatalf("Failed to apply migration: %v", err)
	}

	// Create test user and email
	userID := "test-user-123"
	emailID := "test-email-456"

	_, err = db.Exec("INSERT INTO users (user_id, email, password_hash, totp_secret) VALUES (?, ?, ?, ?)",
		userID, "test@example.com", "hashed_password", "totp_secret")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	_, err = db.Exec("INSERT INTO emails (email_id, user_id) VALUES (?, ?)", emailID, userID)
	if err != nil {
		t.Fatalf("Failed to create test email: %v", err)
	}

	// Initialize server with session token service
	sessionTokenSvc := sessiontokens.NewSessionTokenService(db)
	srv := &Server{
		db:                  db,
		sessionTokenService: sessionTokenSvc,
	}

	// Generate JWT token for testing
	// Create JWT with both user_id and email claims like in the login process
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "test-secret-key-for-jwt-signing"
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"email":   "test@example.com",
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	})
	jwtToken, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		t.Fatalf("Failed to generate JWT token: %v", err)
	}

	t.Run("Generate session token endpoint", func(t *testing.T) {
		// Create request
		req, err := http.NewRequest("POST", "/api/email/"+emailID+"/session", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		req.Header.Set("Authorization", "Bearer "+jwtToken)
		req.Header.Set("User-Agent", "Test-Agent/1.0")
		req.RemoteAddr = "192.168.1.100:12345"

		// Create response recorder
		rr := httptest.NewRecorder()

		// Create a router to properly set up URL variables
		router := mux.NewRouter()
		router.Handle("/api/email/{id}/session", http.HandlerFunc(srv.generateSessionTokenHandler))

		router.ServeHTTP(rr, req)

		// Check response
		if rr.Code != http.StatusOK {
			// Parse error response
			var errorResponse map[string]interface{}
			if err := json.NewDecoder(rr.Body).Decode(&errorResponse); err != nil {
				t.Fatalf("Failed to decode error response: %v", err)
			}
			t.Errorf("Expected status 200, got %d. Error: %v", rr.Code, errorResponse)
			return
		}

		// Parse response
		var response map[string]interface{}
		if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		// Verify response fields
		if response["session_token"] == "" {
			t.Error("Response should contain session_token")
		}

		if response["expires_in"] != float64(300) {
			t.Error("Response should contain expires_in = 300")
		}

		if response["one_time_only"] != false {
			t.Error("Response should contain one_time_only = false")
		}

		// Verify session token is valid
		sessionToken, ok := response["session_token"].(string)
		if !ok {
			t.Fatalf("session_token is not a string: %v", response["session_token"])
		}
		isValid, err := sessionTokenSvc.ValidateSessionToken(emailID, sessionToken)
		if err != nil {
			t.Fatalf("Failed to validate generated session token: %v", err)
		}

		if !isValid {
			t.Error("Generated session token should be valid")
		}
	})

	t.Run("Generate session token with one-time link", func(t *testing.T) {
		// Set email to one-time link mode
		_, err := db.Exec("UPDATE emails SET one_time_link_only = TRUE WHERE email_id = ?", emailID)
		if err != nil {
			t.Fatalf("Failed to update email to one-time link mode: %v", err)
		}

		// Create request
		req, err := http.NewRequest("POST", "/api/email/"+emailID+"/session", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		req.Header.Set("Authorization", "Bearer "+jwtToken)
		req.Header.Set("User-Agent", "Test-Agent/2.0")
		req.RemoteAddr = "192.168.1.101:12346"

		// Create response recorder
		rr := httptest.NewRecorder()

		// Create a router to properly set up URL variables
		router := mux.NewRouter()
		router.Handle("/api/email/{id}/session", http.HandlerFunc(srv.generateSessionTokenHandler))

		router.ServeHTTP(rr, req)

		// Check response
		if rr.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rr.Code)
		}

		// Parse response
		var response map[string]interface{}
		if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		// Verify one-time link flag
		if response["one_time_only"] != true {
			t.Error("Response should contain one_time_only = true")
		}
	})
}
