package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"secure-email-mvp/pkg/auth"
	"secure-email-mvp/pkg/email"

	_ "modernc.org/sqlite"
)

func createTestTablesForReadOnce(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create the emails table with all required columns
	_, err = db.Exec(`
		CREATE TABLE emails (
			email_id TEXT PRIMARY KEY,
			sender_id TEXT NOT NULL,
			recipient TEXT NOT NULL,
			subject TEXT,
			encrypted_blob_url TEXT,
			encrypted_key TEXT,
			encryption_nonce TEXT,
			encryption_auth_tag TEXT,
			compression_algo TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_access_at DATETIME,
			access_count INTEGER DEFAULT 0,
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
			read_once_consumed_at INTEGER,
			read_once_consumer_device TEXT,
			self_destruct_on_read_once BOOLEAN DEFAULT FALSE
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	return db
}

func generateReadOnceTestToken(userID, userEmail string) (string, error) {
	sessionManager, err := auth.NewSessionManager()
	if err != nil {
		return "", err
	}
	return sessionManager.GenerateAccessToken(userID, userEmail)
}

func TestReadOnceFlow_Success(t *testing.T) {
	db := createTestTablesForReadOnce(t)
	defer db.Close()

	// Create test email with read_once = true
	emailID := "test-read-once-email"
	senderID := "test-sender"
	recipient := "test@example.com"
	
	_, err := db.Exec(`
		INSERT INTO emails (
			email_id, sender_id, recipient, subject,
			encrypted_blob_url, encrypted_key, encryption_nonce, encryption_auth_tag,
			compression_algo, read_once, self_destruct_on_read_once
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, emailID, senderID, recipient, "Test Read-Once Email",
		"test-blob-url", "dGVzdC1rZXk=", "dGVzdC1ub25jZQ==", "dGVzdC1hdXRoLXRhZw==",
		"gzip", true, false)
	if err != nil {
		t.Fatalf("Failed to create test email: %v", err)
	}

	// Create server with test database
	srv := &Server{db: db}

	// Generate test token
	token, err := generateReadOnceTestToken(senderID, "sender@example.com")
	if err != nil {
		t.Fatalf("Failed to generate test token: %v", err)
	}

	// First request should succeed
	reqBody := GetEmailRequest{EmailID: emailID}
	reqJSON, _ := json.Marshal(reqBody)
	
	req := httptest.NewRequest("POST", "/api/email/get", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	
	// Apply JWT middleware and call handler
	handler := jwtMiddleware(http.HandlerFunc(srv.getEmailHandler))
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected first request to succeed, got status %d: %s", w.Code, w.Body.String())
	}

	// Verify email is marked as consumed
	emailSecurityDB := email.NewEmailSecurityDB(db)
	isConsumed, _, err := emailSecurityDB.IsReadOnceConsumed(emailID)
	if err != nil {
		t.Fatalf("Failed to check consumption status: %v", err)
	}

	if !isConsumed {
		t.Fatal("Expected email to be marked as consumed after first read")
	}

	// Second request should fail
	req2 := httptest.NewRequest("POST", "/api/email/get", bytes.NewBuffer(reqJSON))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer "+token)

	w2 := httptest.NewRecorder()
	
	// Apply JWT middleware and call handler
	handler2 := jwtMiddleware(http.HandlerFunc(srv.getEmailHandler))
	handler2.ServeHTTP(w2, req2)

	if w2.Code != http.StatusForbidden {
		t.Fatalf("Expected second request to fail, got status %d: %s", w2.Code, w2.Body.String())
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w2.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["error"] != "Email has been revoked by sender" {
		t.Fatalf("Expected generic error message, got: %v", response["error"])
	}
}

func TestReadOnce_DeletionOnRead(t *testing.T) {
	db := createTestTablesForReadOnce(t)
	defer db.Close()

	// Create test email with read_once = true and self_destruct_on_read_once = true
	emailID := "test-read-once-delete-email"
	senderID := "test-sender"
	recipient := "test@example.com"
	
	_, err := db.Exec(`
		INSERT INTO emails (
			email_id, sender_id, recipient, subject,
			encrypted_blob_url, encrypted_key, encryption_nonce, encryption_auth_tag,
			compression_algo, read_once, self_destruct_on_read_once
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, emailID, senderID, recipient, "Test Read-Once Delete Email",
		"test-blob-url", "dGVzdC1rZXk=", "dGVzdC1ub25jZQ==", "dGVzdC1hdXRoLXRhZw==",
		"gzip", true, true)
	if err != nil {
		t.Fatalf("Failed to create test email: %v", err)
	}

	// Create server with test database
	srv := &Server{db: db}

	// Generate test token
	token, err := generateReadOnceTestToken(senderID, "sender@example.com")
	if err != nil {
		t.Fatalf("Failed to generate test token: %v", err)
	}

	// First request should succeed
	reqBody := GetEmailRequest{EmailID: emailID}
	reqJSON, _ := json.Marshal(reqBody)
	
	req := httptest.NewRequest("POST", "/api/email/get", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	
	// Apply JWT middleware and call handler
	handler := jwtMiddleware(http.HandlerFunc(srv.getEmailHandler))
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected first request to succeed, got status %d: %s", w.Code, w.Body.String())
	}

	// Verify email is deleted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM emails WHERE email_id = ?", emailID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to check email existence: %v", err)
	}

	if count != 0 {
		t.Fatal("Expected email to be deleted after read-once consumption with self-destruct")
	}

	// Second request should fail (email doesn't exist)
	req2 := httptest.NewRequest("POST", "/api/email/get", bytes.NewBuffer(reqJSON))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer "+token)

	w2 := httptest.NewRecorder()
	
	// Apply JWT middleware and call handler
	handler2 := jwtMiddleware(http.HandlerFunc(srv.getEmailHandler))
	handler2.ServeHTTP(w2, req2)

	if w2.Code != http.StatusNotFound {
		t.Fatalf("Expected second request to fail with not found, got status %d: %s", w2.Code, w2.Body.String())
	}
}

func TestReadOnce_WithFailedAttempts(t *testing.T) {
	db := createTestTablesForReadOnce(t)
	defer db.Close()

	// Create test email with read_once = true and some failed attempts
	emailID := "test-read-once-failed-attempts"
	senderID := "test-sender"
	recipient := "test@example.com"
	
	_, err := db.Exec(`
		INSERT INTO emails (
			email_id, sender_id, recipient, subject,
			encrypted_blob_url, encrypted_key, encryption_nonce, encryption_auth_tag,
			compression_algo, read_once, failed_attempts, self_destruct_threshold
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, emailID, senderID, recipient, "Test Read-Once Failed Attempts",
		"test-blob-url", "dGVzdC1rZXk=", "dGVzdC1ub25jZQ==", "dGVzdC1hdXRoLXRhZw==",
		"gzip", true, 2, 5)
	if err != nil {
		t.Fatalf("Failed to create test email: %v", err)
	}

	// Create server with test database
	srv := &Server{db: db}

	// Generate test token
	token, err := generateReadOnceTestToken(senderID, "sender@example.com")
	if err != nil {
		t.Fatalf("Failed to generate test token: %v", err)
	}

	// Successful request should reset failed attempts and mark as consumed
	reqBody := GetEmailRequest{EmailID: emailID}
	reqJSON, _ := json.Marshal(reqBody)
	
	req := httptest.NewRequest("POST", "/api/email/get", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	
	// Apply JWT middleware and call handler
	handler := jwtMiddleware(http.HandlerFunc(srv.getEmailHandler))
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected request to succeed, got status %d: %s", w.Code, w.Body.String())
	}

	// Verify failed attempts are reset
	emailSecurityDB := email.NewEmailSecurityDB(db)
	failedAttempts, err := emailSecurityDB.GetFailedAttemptsCount(emailID)
	if err != nil {
		t.Fatalf("Failed to get failed attempts count: %v", err)
	}

	if failedAttempts != 0 {
		t.Fatalf("Expected failed attempts to be reset to 0, got %d", failedAttempts)
	}

	// Verify email is marked as consumed
	isConsumed, _, err := emailSecurityDB.IsReadOnceConsumed(emailID)
	if err != nil {
		t.Fatalf("Failed to check consumption status: %v", err)
	}

	if !isConsumed {
		t.Fatal("Expected email to be marked as consumed")
	}
}

func TestReadOnce_NonReadOnceEmail(t *testing.T) {
	db := createTestTablesForReadOnce(t)
	defer db.Close()

	// Create test email with read_once = false
	emailID := "test-non-read-once-email"
	senderID := "test-sender"
	recipient := "test@example.com"
	
	_, err := db.Exec(`
		INSERT INTO emails (
			email_id, sender_id, recipient, subject,
			encrypted_blob_url, encrypted_key, encryption_nonce, encryption_auth_tag,
			compression_algo, read_once
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, emailID, senderID, recipient, "Test Non-Read-Once Email",
		"test-blob-url", "dGVzdC1rZXk=", "dGVzdC1ub25jZQ==", "dGVzdC1hdXRoLXRhZw==",
		"gzip", false)
	if err != nil {
		t.Fatalf("Failed to create test email: %v", err)
	}

	// Create server with test database
	srv := &Server{db: db}

	// Generate test token
	token, err := generateReadOnceTestToken(senderID, "sender@example.com")
	if err != nil {
		t.Fatalf("Failed to generate test token: %v", err)
	}

	// First request should succeed
	reqBody := GetEmailRequest{EmailID: emailID}
	reqJSON, _ := json.Marshal(reqBody)
	
	req := httptest.NewRequest("POST", "/api/email/get", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	
	// Apply JWT middleware and call handler
	handler := jwtMiddleware(http.HandlerFunc(srv.getEmailHandler))
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected first request to succeed, got status %d: %s", w.Code, w.Body.String())
	}

	// Second request should also succeed (not read-once)
	req2 := httptest.NewRequest("POST", "/api/email/get", bytes.NewBuffer(reqJSON))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer "+token)

	w2 := httptest.NewRecorder()
	
	// Apply JWT middleware and call handler
	handler2 := jwtMiddleware(http.HandlerFunc(srv.getEmailHandler))
	handler2.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("Expected second request to succeed for non-read-once email, got status %d: %s", w2.Code, w2.Body.String())
	}

	// Verify email is not marked as consumed
	emailSecurityDB := email.NewEmailSecurityDB(db)
	isConsumed, _, err := emailSecurityDB.IsReadOnceConsumed(emailID)
	if err != nil {
		t.Fatalf("Failed to check consumption status: %v", err)
	}

	if isConsumed {
		t.Fatal("Expected non-read-once email to not be marked as consumed")
	}
}

// TestReadOnceDirectFunctionality tests the read-once functionality directly without HTTP handler
func TestReadOnceDirectFunctionality(t *testing.T) {
	db := createTestTablesForReadOnce(t)
	defer db.Close()

	// Create test email with read_once = true
	emailID := "test-read-once-direct"
	senderID := "test-sender"
	recipient := "test@example.com"
	
	_, err := db.Exec(`
		INSERT INTO emails (
			email_id, sender_id, recipient, subject,
			encrypted_blob_url, encrypted_key, encryption_nonce, encryption_auth_tag,
			compression_algo, read_once, self_destruct_on_read_once
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, emailID, senderID, recipient, "Test Read-Once Direct",
		"test-blob-url", "dGVzdC1rZXk=", "dGVzdC1ub25jZQ==", "dGVzdC1hdXRoLXRhZw==",
		"gzip", true, false)
	if err != nil {
		t.Fatalf("Failed to create test email: %v", err)
	}

	// Test the read-once functionality directly
	emailSecurityDB := email.NewEmailSecurityDB(db)

	// Verify initial state - not consumed
	isConsumed, _, err := emailSecurityDB.IsReadOnceConsumed(emailID)
	if err != nil {
		t.Fatalf("Failed to check initial consumption status: %v", err)
	}
	if isConsumed {
		t.Fatal("Expected email to not be consumed initially")
	}

	// Mark as consumed
	consumedAt, err := emailSecurityDB.MarkReadOnceConsumed(emailID, "test-device")
	if err != nil {
		t.Fatalf("Failed to mark email as consumed: %v", err)
	}

	// Verify it's now consumed
	isConsumed, actualConsumedAt, err := emailSecurityDB.IsReadOnceConsumed(emailID)
	if err != nil {
		t.Fatalf("Failed to check consumption status after marking: %v", err)
	}
	if !isConsumed {
		t.Fatal("Expected email to be consumed after marking")
	}

	// Verify consumed timestamp matches
	if !actualConsumedAt.Equal(consumedAt) {
		t.Fatalf("Expected consumed timestamp %v, got %v", consumedAt, actualConsumedAt)
	}

	// Try to mark as consumed again - should fail
	_, err = emailSecurityDB.MarkReadOnceConsumed(emailID, "test-device-2")
	if err == nil {
		t.Fatal("Expected second attempt to mark as consumed to fail")
	}

	// Verify the error is the expected type
	if _, ok := err.(email.ReadOnceConsumedError); !ok {
		t.Fatalf("Expected ReadOnceConsumedError, got %T: %v", err, err)
	}

	// Test with self-destruct on read
	emailID2 := "test-read-once-delete-direct"
	_, err = db.Exec(`
		INSERT INTO emails (
			email_id, sender_id, recipient, subject,
			encrypted_blob_url, encrypted_key, encryption_nonce, encryption_auth_tag,
			compression_algo, read_once, self_destruct_on_read_once
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, emailID2, senderID, recipient, "Test Read-Once Delete Direct",
		"test-blob-url-2", "dGVzdC1rZXk=", "dGVzdC1ub25jZQ==", "dGVzdC1hdXRoLXRhZw==",
		"gzip", true, true)
	if err != nil {
		t.Fatalf("Failed to create test email: %v", err)
	}

	// Mark as consumed with self-destruct
	_, err = emailSecurityDB.MarkReadOnceConsumed(emailID2, "test-device")
	if err != nil {
		t.Fatalf("Failed to mark email as consumed with self-destruct: %v", err)
	}

	// Check if email should be deleted after consumption (same logic as getEmailHandler)
	toggles, err := emailSecurityDB.GetEmailSecurityTogglesForAccess(emailID2)
	if err != nil {
		t.Fatalf("Failed to get email security toggles: %v", err)
	}
	if toggles == nil {
		t.Fatal("Expected email security toggles to be found")
	}

	if toggles.ShouldSelfDestructOnReadOnce() {
		log.Printf("Email %s configured for deletion after read, performing secure deletion", emailID2)
		if deleteErr := emailSecurityDB.DeleteEmailSecure(emailID2); deleteErr != nil {
			t.Fatalf("Failed to delete email after read-once consumption: %v", deleteErr)
		}
		log.Printf("Successfully deleted email %s after read-once consumption", emailID2)
	}

	// Verify email is deleted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM emails WHERE email_id = ?", emailID2).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to check email existence: %v", err)
	}

	if count != 0 {
		t.Fatal("Expected email to be deleted after read-once consumption with self-destruct")
	}
}
