package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"secure-email-mvp/pkg/auth"
	"secure-email-mvp/pkg/email"

	_ "modernc.org/sqlite"
)

// Helper function to generate JWT token for self-destruct tests
func generateSelfDestructTestToken(t *testing.T, userID, userEmail string) string {
	sessionManager, err := auth.NewSessionManager()
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}

	token, err := sessionManager.GenerateAccessToken(userID, userEmail)
	if err != nil {
		t.Fatalf("Failed to generate JWT token: %v", err)
	}

	return token
}

// Helper function to create test tables for self-destruct tests
func createTestTablesForSelfDestruct(db *sql.DB) error {
	// Create users table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	// Create emails table with all required columns including failed_attempts
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS emails (
			email_id TEXT PRIMARY KEY,
			sender_id TEXT NOT NULL,
			recipient TEXT NOT NULL,
			subject TEXT,
			encrypted_blob_url TEXT NOT NULL,
			encrypted_key TEXT NOT NULL,
			encryption_nonce TEXT NOT NULL,
			encryption_auth_tag TEXT NOT NULL,
			compression_algo TEXT DEFAULT 'gzip',
			sha256_hash TEXT NOT NULL,
			not_before INTEGER,
			expires_at INTEGER,
			read_once BOOLEAN DEFAULT FALSE,
			mfa_on_open BOOLEAN DEFAULT FALSE,
			decoy_secret TEXT,
			remote_revoke BOOLEAN DEFAULT FALSE,
			strip_metadata BOOLEAN DEFAULT FALSE,
			self_destruct_threshold INTEGER DEFAULT 3,
			geo_rules_ref TEXT,
			failed_attempts INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (sender_id) REFERENCES users(id)
		)
	`)
	return err
}

// TestSelfDestructEnforcement tests that emails are destroyed when failed attempts reach threshold
func TestSelfDestructEnforcement(t *testing.T) {
	// Setup test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Create test tables
	if err := createTestTablesForSelfDestruct(db); err != nil {
		t.Fatalf("Failed to create test tables: %v", err)
	}

	// Create test user
	testUserID := "user-1"
	testUserEmail := "test@securesystem.email"
	otherUserID := "user-2"
	otherUserEmail := "other@securesystem.email"

	_, err = db.Exec(`INSERT INTO users (id, email, password_hash) VALUES (?, ?, ?)`,
		testUserID, testUserEmail, "hash1")
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	_, err = db.Exec(`INSERT INTO users (id, email, password_hash) VALUES (?, ?, ?)`,
		otherUserID, otherUserEmail, "hash2")
	if err != nil {
		t.Fatalf("Failed to insert other test user: %v", err)
	}

	// Create server with test database
	srv := &Server{db: db}

	tests := []struct {
		name                  string
		emailID               string
		senderID              string
		initialFailedAttempts int
		selfDestructThreshold int
		attemptingUserID      string
		attemptingUserEmail   string
		expectedStatus        int
		expectedError         string
		expectedEmailDeleted  bool
	}{
		{
			name:                  "Fail access N times and verify deletion",
			emailID:               "email-1",
			senderID:              testUserID,
			initialFailedAttempts: 2, // Start with 2 failed attempts
			selfDestructThreshold: 3, // Threshold is 3
			attemptingUserID:      otherUserID,
			attemptingUserEmail:   otherUserEmail,
			expectedStatus:        http.StatusForbidden,
			expectedError:         "Email has been revoked by sender",
			expectedEmailDeleted:  true,
		},
		{
			name:                  "Counter persists between requests",
			emailID:               "email-2",
			senderID:              testUserID,
			initialFailedAttempts: 1,
			selfDestructThreshold: 3,
			attemptingUserID:      otherUserID,
			attemptingUserEmail:   otherUserEmail,
			expectedStatus:        http.StatusForbidden,
			expectedError:         "Access denied",
			expectedEmailDeleted:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test email
			_, err = db.Exec(`
				INSERT INTO emails (
					email_id, sender_id, recipient, subject, encrypted_blob_url,
					encrypted_key, encryption_nonce, encryption_auth_tag, sha256_hash,
					failed_attempts, self_destruct_threshold
				) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				tt.emailID, tt.senderID, "recipient@example.com", "Test Email",
				"blob-1", "key-1", "nonce-1", "tag-1", "hash-1",
				tt.initialFailedAttempts, tt.selfDestructThreshold,
			)
			if err != nil {
				t.Fatalf("Failed to insert test email: %v", err)
			}

			// Generate JWT token for attempting user
			token := generateSelfDestructTestToken(t, tt.attemptingUserID, tt.attemptingUserEmail)

			// Create request to access email
			requestBody := map[string]string{"email_id": tt.emailID}
			bodyBytes, _ := json.Marshal(requestBody)

			req := httptest.NewRequest("POST", "/api/email/get", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			// Apply JWT middleware and call handler
			handler := jwtMiddleware(http.HandlerFunc(srv.getEmailHandler))
			handler.ServeHTTP(w, req)

			// Check response status
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
				t.Logf("Response body: %s", w.Body.String())
				return
			}

			// Check error message if expected
			if tt.expectedError != "" {
				var response map[string]interface{}
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				if response["error"] != tt.expectedError {
					t.Errorf("Expected error '%s', got '%v'", tt.expectedError, response["error"])
				}
			}

			// Check if email was deleted
			if tt.expectedEmailDeleted {
				// Verify email no longer exists
				var exists int
				err := db.QueryRow("SELECT 1 FROM emails WHERE email_id = ?", tt.emailID).Scan(&exists)
				if err != sql.ErrNoRows {
					t.Error("Email should have been deleted but still exists")
				}
			} else {
				// Verify email still exists
				var exists int
				err := db.QueryRow("SELECT 1 FROM emails WHERE email_id = ?", tt.emailID).Scan(&exists)
				if err != nil {
					t.Errorf("Email should still exist but was deleted: %v", err)
				}
			}

			// Clean up test email
			db.Exec("DELETE FROM emails WHERE email_id = ?", tt.emailID)
		})
	}
}

// TestSelfDestructCounterReset tests that the ResetFailedAttempts function works correctly
func TestSelfDestructCounterReset(t *testing.T) {
	// Setup test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Create test tables
	if err := createTestTablesForSelfDestruct(db); err != nil {
		t.Fatalf("Failed to create test tables: %v", err)
	}

	// Create test user
	testUserID := "user-1"
	testUserEmail := "test@securesystem.email"

	_, err = db.Exec(`INSERT INTO users (id, email, password_hash) VALUES (?, ?, ?)`,
		testUserID, testUserEmail, "hash1")
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	// Create test email with failed attempts
	emailID := "email-reset"
	_, err = db.Exec(`
		INSERT INTO emails (
			email_id, sender_id, recipient, subject, encrypted_blob_url,
			encrypted_key, encryption_nonce, encryption_auth_tag, sha256_hash,
			failed_attempts, self_destruct_threshold
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		emailID, testUserID, "recipient@example.com", "Test Email",
		"blob-1", "key-1", "nonce-1", "tag-1", "hash-1",
		2, 3, // 2 failed attempts, threshold of 3
	)
	if err != nil {
		t.Fatalf("Failed to insert test email: %v", err)
	}

	// Test the ResetFailedAttempts function directly
	emailSecurityDB := email.NewEmailSecurityDB(db)

	// Verify initial count
	count, err := emailSecurityDB.GetFailedAttemptsCount(emailID)
	if err != nil {
		t.Errorf("Failed to get initial failed attempts count: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected initial failed attempts 2, got %d", count)
	}

	// Reset failed attempts
	err = emailSecurityDB.ResetFailedAttempts(emailID)
	if err != nil {
		t.Errorf("Failed to reset failed attempts: %v", err)
	}

	// Verify reset
	count, err = emailSecurityDB.GetFailedAttemptsCount(emailID)
	if err != nil {
		t.Errorf("Failed to get failed attempts count after reset: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected failed attempts to be reset to 0, got %d", count)
	}
}

// TestSelfDestructCounterPersistence tests that the failed attempts counter persists across requests
func TestSelfDestructCounterPersistence(t *testing.T) {
	// Setup test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Create test tables
	if err := createTestTablesForSelfDestruct(db); err != nil {
		t.Fatalf("Failed to create test tables: %v", err)
	}

	// Create test users
	testUserID := "user-1"
	testUserEmail := "test@securesystem.email"
	otherUserID := "user-2"
	otherUserEmail := "other@securesystem.email"

	_, err = db.Exec(`INSERT INTO users (id, email, password_hash) VALUES (?, ?, ?)`,
		testUserID, testUserEmail, "hash1")
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	_, err = db.Exec(`INSERT INTO users (id, email, password_hash) VALUES (?, ?, ?)`,
		otherUserID, otherUserEmail, "hash2")
	if err != nil {
		t.Fatalf("Failed to insert other test user: %v", err)
	}

	// Create test email with 1 failed attempt
	emailID := "email-persistence"
	_, err = db.Exec(`
		INSERT INTO emails (
			email_id, sender_id, recipient, subject, encrypted_blob_url,
			encrypted_key, encryption_nonce, encryption_auth_tag, sha256_hash,
			failed_attempts, self_destruct_threshold
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		emailID, testUserID, "recipient@example.com", "Test Email",
		"blob-1", "key-1", "nonce-1", "tag-1", "hash-1",
		1, 3, // 1 failed attempt, threshold of 3
	)
	if err != nil {
		t.Fatalf("Failed to insert test email: %v", err)
	}

	// Create server with test database
	srv := &Server{db: db}

	// Generate JWT token for unauthorized user
	token := generateSelfDestructTestToken(t, otherUserID, otherUserEmail)

	// First failed attempt - should increment to 2
	requestBody := map[string]string{"email_id": emailID}
	bodyBytes, _ := json.Marshal(requestBody)

	req1 := httptest.NewRequest("POST", "/api/email/get", bytes.NewBuffer(bodyBytes))
	req1.Header.Set("Authorization", "Bearer "+token)
	req1.Header.Set("Content-Type", "application/json")

	w1 := httptest.NewRecorder()
	handler := jwtMiddleware(http.HandlerFunc(srv.getEmailHandler))
	handler.ServeHTTP(w1, req1)

	if w1.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w1.Code)
	}

	// Second failed attempt - should increment to 3 and trigger deletion
	req2 := httptest.NewRequest("POST", "/api/email/get", bytes.NewBuffer(bodyBytes))
	req2.Header.Set("Authorization", "Bearer "+token)
	req2.Header.Set("Content-Type", "application/json")

	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)

	if w2.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w2.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w2.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["error"] != "Email has been revoked by sender" {
		t.Errorf("Expected error 'Email has been revoked by sender', got '%v'", response["error"])
	}

	// Verify email was deleted
	var exists int
	err = db.QueryRow("SELECT 1 FROM emails WHERE email_id = ?", emailID).Scan(&exists)
	if err != sql.ErrNoRows {
		t.Error("Email should have been deleted but still exists")
	}
}

// TestSelfDestructGenericError tests that generic error messages are returned to avoid info leakage
func TestSelfDestructGenericError(t *testing.T) {
	// Setup test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Create test tables
	if err := createTestTablesForSelfDestruct(db); err != nil {
		t.Fatalf("Failed to create test tables: %v", err)
	}

	// Create test users
	testUserID := "user-1"
	testUserEmail := "test@securesystem.email"
	otherUserID := "user-2"
	otherUserEmail := "other@securesystem.email"

	_, err = db.Exec(`INSERT INTO users (id, email, password_hash) VALUES (?, ?, ?)`,
		testUserID, testUserEmail, "hash1")
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	_, err = db.Exec(`INSERT INTO users (id, email, password_hash) VALUES (?, ?, ?)`,
		otherUserID, otherUserEmail, "hash2")
	if err != nil {
		t.Fatalf("Failed to insert other test user: %v", err)
	}

	// Create test email with threshold already reached
	emailID := "email-generic-error"
	_, err = db.Exec(`
		INSERT INTO emails (
			email_id, sender_id, recipient, subject, encrypted_blob_url,
			encrypted_key, encryption_nonce, encryption_auth_tag, sha256_hash,
			failed_attempts, self_destruct_threshold
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		emailID, testUserID, "recipient@example.com", "Test Email",
		"blob-1", "key-1", "nonce-1", "tag-1", "hash-1",
		3, 3, // 3 failed attempts, threshold of 3
	)
	if err != nil {
		t.Fatalf("Failed to insert test email: %v", err)
	}

	// Create server with test database
	srv := &Server{db: db}

	// Generate JWT token for unauthorized user
	token := generateSelfDestructTestToken(t, otherUserID, otherUserEmail)

	// Attempt to access email - should trigger deletion and return generic error
	requestBody := map[string]string{"email_id": emailID}
	bodyBytes, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/api/email/get", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler := jwtMiddleware(http.HandlerFunc(srv.getEmailHandler))
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Should return generic error, not specific details about self-destruct
	if response["error"] != "Email has been revoked by sender" {
		t.Errorf("Expected generic error 'Email has been revoked by sender', got '%v'", response["error"])
	}

	// Verify email was deleted
	var exists int
	err = db.QueryRow("SELECT 1 FROM emails WHERE email_id = ?", emailID).Scan(&exists)
	if err != sql.ErrNoRows {
		t.Error("Email should have been deleted but still exists")
	}
}
