package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	_ "modernc.org/sqlite"
)

// setupTestDB creates a test database with the required schema
func setupTestDBForGetEmail(t *testing.T) *sql.DB {
	// Create temporary database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Apply schema
	schema := `
	CREATE TABLE IF NOT EXISTS emails (
		email_id TEXT PRIMARY KEY,
		sender_id TEXT NOT NULL,
		recipient TEXT NOT NULL,
		subject TEXT,
		encrypted_blob_url TEXT,
		encrypted_key TEXT,
		compression_algo TEXT DEFAULT 'gzip',
		sha256_hash TEXT,
		requires_password INTEGER DEFAULT 0,
		password_hash TEXT,
		geolocation_json TEXT,
		expires_at DATETIME,
		burn_after_read INTEGER DEFAULT 0,
		failed_attempts INTEGER DEFAULT 0,
		max_attempts INTEGER DEFAULT 3,
		self_destruct_after_attempts INTEGER DEFAULT 0,
		self_destruct_threshold INTEGER DEFAULT 3,
		reply_enabled INTEGER DEFAULT 0,
		reply_requires_password INTEGER DEFAULT 1,
		allow_forwarding INTEGER DEFAULT 0,
		show_sender_metadata INTEGER DEFAULT 0,
		metadata_stripped INTEGER DEFAULT 1,
		is_honeytoken INTEGER DEFAULT 0,
		secure_link_id TEXT UNIQUE,
		link_created_at DATETIME,
		last_access_at DATETIME,
		access_count INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME,
		self_destructed INTEGER DEFAULT 0,
		deleted_at DATETIME,
		failed_access_attempts INTEGER DEFAULT 0,
		encryption_nonce TEXT,
		encryption_auth_tag TEXT,
		fail_count INTEGER DEFAULT 0
	);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to apply schema: %v", err)
	}

	return db
}

// createTestEmailForGet creates a test email in the database
func createTestEmailForGet(t *testing.T, db *sql.DB, emailID, senderID string, failCount int) {
	query := `
	INSERT INTO emails (
		email_id, sender_id, recipient, subject, encrypted_blob_url, encrypted_key,
		sha256_hash, encryption_nonce, encryption_auth_tag, failed_attempts
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := db.Exec(query,
		emailID, senderID, "test@example.com", "Test Subject",
		"test-blob-id", "test-key", "test-hash", "test-nonce", "test-auth-tag", failCount,
	)
	if err != nil {
		t.Fatalf("Failed to create test email: %v", err)
	}
}

func TestHandleFailedAccess(t *testing.T) {
	db := setupTestDBForGetEmail(t)
	defer db.Close()

	srv := &Server{db: db}

	// Test case 1: Increment fail count without reaching limit
	emailID := "test-email-1"
	createTestEmailForGet(t, db, emailID, "test-sender", 0)

	err := srv.handleFailedAccess(emailID, "test-blob", 0, "test failure")
	if err != nil {
		t.Errorf("Failed to handle failed access: %v", err)
	}

	// Check that fail count was incremented
	var failCount int
	err = db.QueryRow("SELECT failed_attempts FROM emails WHERE email_id = ?", emailID).Scan(&failCount)
	if err != nil {
		t.Errorf("Failed to get fail count: %v", err)
	}
	if failCount != 1 {
		t.Errorf("Expected fail count to be 1, got %d", failCount)
	}

	// Test case 2: Reach limit and trigger deletion
	emailID2 := "test-email-2"
	createTestEmailForGet(t, db, emailID2, "test-sender", 2) // Start with 2 failures

	err = srv.handleFailedAccess(emailID2, "test-blob-2", 2, "test failure")
	if err == nil {
		t.Errorf("Expected EmailDeletedError, got nil")
	}

	if _, ok := err.(EmailDeletedError); !ok {
		t.Errorf("Expected EmailDeletedError, got %T", err)
	}

	// Check that email was deleted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM emails WHERE email_id = ?", emailID2).Scan(&count)
	if err != nil {
		t.Errorf("Failed to check email existence: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected email to be deleted, but it still exists")
	}
}

func TestGetEmailHandlerFailCount(t *testing.T) {
	db := setupTestDBForGetEmail(t)
	defer db.Close()

	srv := &Server{db: db}

	// Create test email with high fail count
	emailID := "test-email-3"
	createTestEmailForGet(t, db, emailID, "test-sender", 2) // 2 failures already

	// Create request
	reqBody := GetEmailRequest{EmailID: emailID}
	reqBodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/email/get", bytes.NewBuffer(reqBodyBytes))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Set up router
	router := mux.NewRouter()
	router.HandleFunc("/api/email/get", srv.getEmailHandler).Methods("POST")

	// Mock JWT middleware by setting user context
	ctx := context.WithValue(req.Context(), UserIDKey, "wrong-sender") // Wrong sender to trigger failure
	req = req.WithContext(ctx)

	// Make request
	router.ServeHTTP(w, req)

	// Check response - should be 410 Gone after the third failed attempt
	if w.Code != http.StatusGone {
		t.Errorf("Expected status 410 Gone, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if response["error"] != "Email deleted due to too many failed attempts" {
		t.Errorf("Expected error about email deletion, got %v", response["error"])
	}

	// Verify email was deleted from database
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM emails WHERE email_id = ?", emailID).Scan(&count)
	if err != nil {
		t.Errorf("Failed to check email existence: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected email to be deleted, but it still exists")
	}
}
