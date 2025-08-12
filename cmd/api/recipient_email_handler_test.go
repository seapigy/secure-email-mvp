// =============================================================================
// SECURE EMAIL MVP - RECIPIENT EMAIL HANDLER TESTS
// =============================================================================
// Unit tests for recipient-based email access with forwarding prevention.
// Micro-Iteration 4.18: Secure Email Forwarding Prevention
// =============================================================================

package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"secure-email-mvp/pkg/auth"

	"github.com/google/uuid"
)

// TestGetRecipientEmailHandler_CorrectRecipient tests that the correct recipient can fetch and decrypt successfully
func TestGetRecipientEmailHandler_CorrectRecipient(t *testing.T) {
	// Setup test database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Create tables
	if err := createTestTables(db); err != nil {
		t.Fatalf("Failed to create test tables: %v", err)
	}

	// Create test server
	srv := &Server{db: db}

	// Create test users
	senderID := uuid.New().String()
	recipientID := uuid.New().String()
	emailID := uuid.New().String()

	// Insert test users
	_, err = db.Exec(`INSERT INTO users (user_id, email, password_hash) VALUES (?, ?, ?)`, senderID, "sender@test.com", "hash")
	if err != nil {
		t.Fatalf("Failed to insert sender: %v", err)
	}
	_, err = db.Exec(`INSERT INTO users (user_id, email, password_hash) VALUES (?, ?, ?)`, recipientID, "recipient@test.com", "hash")
	if err != nil {
		t.Fatalf("Failed to insert recipient: %v", err)
	}

	// Create test email content
	originalContent := "This is a test email content"
	compressed := compressContent(originalContent)
	encryptedData := encryptContent(compressed)

	// Upload to R2 (mock)
	blobID := uuid.New().String() + ".blob"
	// Note: In a real test, you'd mock the R2 storage

	// Insert test email with recipient_id
	_, err = db.Exec(`
		INSERT INTO emails (
			email_id, sender_id, recipient, recipient_id, subject, encrypted_blob_url, 
			encrypted_key, encryption_nonce, encryption_auth_tag, compression_algo, 
			sha256_hash, created_at, fail_count
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		emailID, senderID, "recipient@test.com", recipientID, "Test Subject", blobID,
		base64.StdEncoding.EncodeToString(encryptedData.Key),
		base64.StdEncoding.EncodeToString(encryptedData.Nonce),
		base64.StdEncoding.EncodeToString(encryptedData.AuthTag),
		"gzip", "test-hash", time.Now(), 0,
	)
	if err != nil {
		t.Fatalf("Failed to insert test email: %v", err)
	}

	// Create request
	req := httptest.NewRequest("GET", "/api/email/"+emailID+"/content", nil)
	w := httptest.NewRecorder()

	// Add user context (simulating JWT middleware)
	ctx := context.WithValue(req.Context(), "user_id", recipientID)
	req = req.WithContext(ctx)

	// Call handler
	srv.getRecipientEmailHandler(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response RecipientEmailResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Status != "success" {
		t.Errorf("Expected status 'success', got '%s'", response.Status)
	}

	if response.Body != originalContent {
		t.Errorf("Expected body '%s', got '%s'", originalContent, response.Body)
	}

	if response.Recipient != "recipient@test.com" {
		t.Errorf("Expected recipient 'recipient@test.com', got '%s'", response.Recipient)
	}
}

// TestGetRecipientEmailHandler_UnauthorizedUser tests that any other user_id gets blocked
func TestGetRecipientEmailHandler_UnauthorizedUser(t *testing.T) {
	// Setup test database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Create tables
	if err := createTestTables(db); err != nil {
		t.Fatalf("Failed to create test tables: %v", err)
	}

	// Create test server
	srv := &Server{db: db}

	// Create test users
	senderID := uuid.New().String()
	recipientID := uuid.New().String()
	unauthorizedUserID := uuid.New().String()
	emailID := uuid.New().String()

	// Insert test users
	_, err = db.Exec(`INSERT INTO users (user_id, email, password_hash) VALUES (?, ?, ?)`, senderID, "sender@test.com", "hash")
	if err != nil {
		t.Fatalf("Failed to insert sender: %v", err)
	}
	_, err = db.Exec(`INSERT INTO users (user_id, email, password_hash) VALUES (?, ?, ?)`, recipientID, "recipient@test.com", "hash")
	if err != nil {
		t.Fatalf("Failed to insert recipient: %v", err)
	}
	_, err = db.Exec(`INSERT INTO users (user_id, email, password_hash) VALUES (?, ?, ?)`, unauthorizedUserID, "unauthorized@test.com", "hash")
	if err != nil {
		t.Fatalf("Failed to insert unauthorized user: %v", err)
	}

	// Create test email content
	originalContent := "This is a test email content"
	compressed := compressContent(originalContent)
	encryptedData := encryptContent(compressed)

	// Upload to R2 (mock)
	blobID := uuid.New().String() + ".blob"

	// Insert test email with recipient_id
	_, err = db.Exec(`
		INSERT INTO emails (
			email_id, sender_id, recipient, recipient_id, subject, encrypted_blob_url, 
			encrypted_key, encryption_nonce, encryption_auth_tag, compression_algo, 
			sha256_hash, created_at, fail_count
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		emailID, senderID, "recipient@test.com", recipientID, "Test Subject", blobID,
		base64.StdEncoding.EncodeToString(encryptedData.Key),
		base64.StdEncoding.EncodeToString(encryptedData.Nonce),
		base64.StdEncoding.EncodeToString(encryptedData.AuthTag),
		"gzip", "test-hash", time.Now(), 0,
	)
	if err != nil {
		t.Fatalf("Failed to insert test email: %v", err)
	}

	// Create request with unauthorized user
	req := httptest.NewRequest("GET", "/api/email/"+emailID+"/content", nil)
	w := httptest.NewRecorder()

	// Add unauthorized user context
	ctx := context.WithValue(req.Context(), "user_id", unauthorizedUserID)
	req = req.WithContext(ctx)

	// Call handler
	srv.getRecipientEmailHandler(w, req)

	// Check response
	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["error"] != "Access denied" {
		t.Errorf("Expected error 'Access denied', got '%v'", response["error"])
	}

	// Check that fail count was incremented
	var failCount int
	err = db.QueryRow(`SELECT fail_count FROM emails WHERE email_id = ?`, emailID).Scan(&failCount)
	if err != nil {
		t.Fatalf("Failed to query fail count: %v", err)
	}

	if failCount != 1 {
		t.Errorf("Expected fail count 1, got %d", failCount)
	}
}

// TestGetRecipientEmailHandler_FailureCountIncrement tests that failure count increments on each unauthorized attempt
func TestGetRecipientEmailHandler_FailureCountIncrement(t *testing.T) {
	// Setup test database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Create tables
	if err := createTestTables(db); err != nil {
		t.Fatalf("Failed to create test tables: %v", err)
	}

	// Create test server
	srv := &Server{db: db}

	// Create test users
	senderID := uuid.New().String()
	recipientID := uuid.New().String()
	unauthorizedUserID := uuid.New().String()
	emailID := uuid.New().String()

	// Insert test users
	_, err = db.Exec(`INSERT INTO users (user_id, email, password_hash) VALUES (?, ?, ?)`, senderID, "sender@test.com", "hash")
	if err != nil {
		t.Fatalf("Failed to insert sender: %v", err)
	}
	_, err = db.Exec(`INSERT INTO users (user_id, email, password_hash) VALUES (?, ?, ?)`, recipientID, "recipient@test.com", "hash")
	if err != nil {
		t.Fatalf("Failed to insert recipient: %v", err)
	}
	_, err = db.Exec(`INSERT INTO users (user_id, email, password_hash) VALUES (?, ?, ?)`, unauthorizedUserID, "unauthorized@test.com", "hash")
	if err != nil {
		t.Fatalf("Failed to insert unauthorized user: %v", err)
	}

	// Create test email content
	originalContent := "This is a test email content"
	compressed := compressContent(originalContent)
	encryptedData := encryptContent(compressed)

	// Upload to R2 (mock)
	blobID := uuid.New().String() + ".blob"

	// Insert test email with recipient_id
	_, err = db.Exec(`
		INSERT INTO emails (
			email_id, sender_id, recipient, recipient_id, subject, encrypted_blob_url, 
			encrypted_key, encryption_nonce, encryption_auth_tag, compression_algo, 
			sha256_hash, created_at, fail_count
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		emailID, senderID, "recipient@test.com", recipientID, "Test Subject", blobID,
		base64.StdEncoding.EncodeToString(encryptedData.Key),
		base64.StdEncoding.EncodeToString(encryptedData.Nonce),
		base64.StdEncoding.EncodeToString(encryptedData.AuthTag),
		"gzip", "test-hash", time.Now(), 0,
	)
	if err != nil {
		t.Fatalf("Failed to insert test email: %v", err)
	}

	// Make multiple unauthorized attempts
	for i := 1; i <= 3; i++ {
		req := httptest.NewRequest("GET", "/api/email/"+emailID+"/content", nil)
		w := httptest.NewRecorder()

		// Add unauthorized user context
		ctx := context.WithValue(req.Context(), "user_id", unauthorizedUserID)
		req = req.WithContext(ctx)

		// Call handler
		srv.getRecipientEmailHandler(w, req)

		// Check response
		if w.Code != http.StatusForbidden {
			t.Errorf("Attempt %d: Expected status 403, got %d", i, w.Code)
		}

		// Check that fail count was incremented
		var failCount int
		err = db.QueryRow(`SELECT fail_count FROM emails WHERE email_id = ?`, emailID).Scan(&failCount)
		if err != nil {
			t.Fatalf("Attempt %d: Failed to query fail count: %v", i, err)
		}

		if failCount != i {
			t.Errorf("Attempt %d: Expected fail count %d, got %d", i, i, failCount)
		}
	}
}

// TestGetRecipientEmailHandler_AutoDeleteAfterThreeAttempts tests that email is auto-deleted after 3 failed attempts
func TestGetRecipientEmailHandler_AutoDeleteAfterThreeAttempts(t *testing.T) {
	// Setup test database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Create tables
	if err := createTestTables(db); err != nil {
		t.Fatalf("Failed to create test tables: %v", err)
	}

	// Create test server
	srv := &Server{db: db}

	// Create test users
	senderID := uuid.New().String()
	recipientID := uuid.New().String()
	unauthorizedUserID := uuid.New().String()
	emailID := uuid.New().String()

	// Insert test users
	_, err = db.Exec(`INSERT INTO users (user_id, email, password_hash) VALUES (?, ?, ?)`, senderID, "sender@test.com", "hash")
	if err != nil {
		t.Fatalf("Failed to insert sender: %v", err)
	}
	_, err = db.Exec(`INSERT INTO users (user_id, email, password_hash) VALUES (?, ?, ?)`, recipientID, "recipient@test.com", "hash")
	if err != nil {
		t.Fatalf("Failed to insert recipient: %v", err)
	}
	_, err = db.Exec(`INSERT INTO users (user_id, email, password_hash) VALUES (?, ?, ?)`, unauthorizedUserID, "unauthorized@test.com", "hash")
	if err != nil {
		t.Fatalf("Failed to insert unauthorized user: %v", err)
	}

	// Create test email content
	originalContent := "This is a test email content"
	compressed := compressContent(originalContent)
	encryptedData := encryptContent(compressed)

	// Upload to R2 (mock)
	blobID := uuid.New().String() + ".blob"

	// Insert test email with recipient_id
	_, err = db.Exec(`
		INSERT INTO emails (
			email_id, sender_id, recipient, recipient_id, subject, encrypted_blob_url, 
			encrypted_key, encryption_nonce, encryption_auth_tag, compression_algo, 
			sha256_hash, created_at, fail_count
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		emailID, senderID, "recipient@test.com", recipientID, "Test Subject", blobID,
		base64.StdEncoding.EncodeToString(encryptedData.Key),
		base64.StdEncoding.EncodeToString(encryptedData.Nonce),
		base64.StdEncoding.EncodeToString(encryptedData.AuthTag),
		"gzip", "test-hash", time.Now(), 0,
	)
	if err != nil {
		t.Fatalf("Failed to insert test email: %v", err)
	}

	// Make 3 unauthorized attempts
	for i := 1; i <= 3; i++ {
		req := httptest.NewRequest("GET", "/api/email/"+emailID+"/content", nil)
		w := httptest.NewRecorder()

		// Add unauthorized user context
		ctx := context.WithValue(req.Context(), "user_id", unauthorizedUserID)
		req = req.WithContext(ctx)

		// Call handler
		srv.getRecipientEmailHandler(w, req)

		// Check response for the 3rd attempt
		if i == 3 {
			if w.Code != http.StatusGone {
				t.Errorf("Expected status 410 (Gone), got %d", w.Code)
			}

			var response map[string]interface{}
			if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			if response["error"] != "Email deleted due to too many failed attempts" {
				t.Errorf("Expected error 'Email deleted due to too many failed attempts', got '%v'", response["error"])
			}
		}
	}

	// Verify email was deleted from database
	var count int
	err = db.QueryRow(`SELECT COUNT(*) FROM emails WHERE email_id = ?`, emailID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query email count: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected email to be deleted, but found %d records", count)
	}
}

// TestGetRecipientEmailHandler_NoRecipientID tests that emails without recipient_id are blocked
func TestGetRecipientEmailHandler_NoRecipientID(t *testing.T) {
	// Setup test database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Create tables
	if err := createTestTables(db); err != nil {
		t.Fatalf("Failed to create test tables: %v", err)
	}

	// Create test server
	srv := &Server{db: db}

	// Create test users
	senderID := uuid.New().String()
	userID := uuid.New().String()
	emailID := uuid.New().String()

	// Insert test users
	_, err = db.Exec(`INSERT INTO users (user_id, email, password_hash) VALUES (?, ?, ?)`, senderID, "sender@test.com", "hash")
	if err != nil {
		t.Fatalf("Failed to insert sender: %v", err)
	}
	_, err = db.Exec(`INSERT INTO users (user_id, email, password_hash) VALUES (?, ?, ?)`, userID, "user@test.com", "hash")
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}

	// Create test email content
	originalContent := "This is a test email content"
	compressed := compressContent(originalContent)
	encryptedData := encryptContent(compressed)

	// Upload to R2 (mock)
	blobID := uuid.New().String() + ".blob"

	// Insert test email WITHOUT recipient_id (NULL)
	_, err = db.Exec(`
		INSERT INTO emails (
			email_id, sender_id, recipient, recipient_id, subject, encrypted_blob_url, 
			encrypted_key, encryption_nonce, encryption_auth_tag, compression_algo, 
			sha256_hash, created_at, fail_count
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		emailID, senderID, "recipient@test.com", nil, "Test Subject", blobID,
		base64.StdEncoding.EncodeToString(encryptedData.Key),
		base64.StdEncoding.EncodeToString(encryptedData.Nonce),
		base64.StdEncoding.EncodeToString(encryptedData.AuthTag),
		"gzip", "test-hash", time.Now(), 0,
	)
	if err != nil {
		t.Fatalf("Failed to insert test email: %v", err)
	}

	// Create request
	req := httptest.NewRequest("GET", "/api/email/"+emailID+"/content", nil)
	w := httptest.NewRecorder()

	// Add user context
	ctx := context.WithValue(req.Context(), "user_id", userID)
	req = req.WithContext(ctx)

	// Call handler
	srv.getRecipientEmailHandler(w, req)

	// Check response
	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["error"] != "Access denied" {
		t.Errorf("Expected error 'Access denied', got '%v'", response["error"])
	}

	// Check that fail count was incremented
	var failCount int
	err = db.QueryRow(`SELECT fail_count FROM emails WHERE email_id = ?`, emailID).Scan(&failCount)
	if err != nil {
		t.Fatalf("Failed to query fail count: %v", err)
	}

	if failCount != 1 {
		t.Errorf("Expected fail count 1, got %d", failCount)
	}
}

// Helper functions

func createTestTables(db *sql.DB) error {
	// Create users table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			user_id TEXT PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	// Create emails table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS emails (
			email_id TEXT PRIMARY KEY,
			sender_id TEXT NOT NULL,
			recipient TEXT NOT NULL,
			recipient_id TEXT,
			subject TEXT,
			encrypted_blob_url TEXT NOT NULL,
			encrypted_key TEXT NOT NULL,
			encryption_nonce TEXT NOT NULL,
			encryption_auth_tag TEXT NOT NULL,
			compression_algo TEXT DEFAULT 'gzip',
			sha256_hash TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			fail_count INTEGER DEFAULT 0,
			access_count INTEGER DEFAULT 0,
			last_access_at DATETIME
		)
	`)
	if err != nil {
		return err
	}

	// Create access_events table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS access_events (
			event_id TEXT PRIMARY KEY,
			email_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			event_type TEXT NOT NULL,
			ip_address TEXT,
			user_agent TEXT,
			country TEXT,
			city TEXT,
			device_type TEXT,
			failure_reason TEXT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	return err
}

func compressContentHelper(content string) []byte {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	gz.Write([]byte(content))
	gz.Close()
	return buf.Bytes()
}

func encryptContent(data []byte) *auth.EncryptedData {
	// Generate a random key
	key := make([]byte, 32)
	rand.Read(key)

	// Generate a random nonce
	nonce := make([]byte, 12)
	rand.Read(nonce)

	// Create cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		panic(err)
	}

	// Encrypt
	ciphertext := aesGCM.Seal(nil, nonce, data, nil)

	// Extract auth tag (last 16 bytes)
	authTag := ciphertext[len(ciphertext)-16:]
	ciphertext = ciphertext[:len(ciphertext)-16]

	return &auth.EncryptedData{
		Ciphertext: ciphertext,
		Key:        key,
		Nonce:      nonce,
		AuthTag:    authTag,
	}
}
