// R2-backed HTTP Integration Tests for Read-Once Functionality
//
// This file contains end-to-end integration tests that require R2 storage.
// Tests will be skipped if R2 environment variables are not configured:
//   - R2_ACCESS_KEY_ID
//   - R2_SECRET_ACCESS_KEY
//   - R2_BUCKET
//   - R2_ENDPOINT
//   - R2_REGION
//
// To run these tests, create a .env file with valid R2 credentials.
// Tests will auto-skip if credentials are missing to ensure CI/CD passes.

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
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"secure-email-mvp/pkg/auth"
	"secure-email-mvp/pkg/email"
	"secure-email-mvp/pkg/storage"

	"github.com/joho/godotenv"
	_ "modernc.org/sqlite"
)

// TestConfig holds R2 configuration for tests
type TestConfig struct {
	R2Client *storage.R2Client
	Bucket   string
}

// setupR2IntegrationTest initializes R2 client and returns test configuration
// Returns nil if R2 credentials are not available (tests will be skipped)
func setupR2IntegrationTest(t *testing.T) *TestConfig {
	// Load environment variables from .env file
	// Try current directory first, then project root
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Could not load .env file from current directory: %v", err)
		// Try loading from project root (two directories up from cmd/api)
		if err := godotenv.Load("../../.env"); err != nil {
			log.Printf("Warning: Could not load .env file from project root: %v", err)
		} else {
			log.Printf("Successfully loaded .env from project root")
		}
	} else {
		log.Printf("Successfully loaded .env from current directory")
	}

	// Check for required R2 environment variables
	requiredVars := []string{
		"R2_ACCESS_KEY_ID",
		"R2_SECRET_ACCESS_KEY",
		"R2_BUCKET",
		"R2_ENDPOINT",
		"R2_REGION",
	}

	for _, varName := range requiredVars {
		value := os.Getenv(varName)
		if value == "" {
			t.Skipf("Skipping R2 integration tests: missing %s", varName)
		}
		// Log the first few characters of each value for debugging (don't log full secrets)
		if varName == "R2_SECRET_ACCESS_KEY" {
			if len(value) > 4 {
				log.Printf("Found %s: %s...", varName, value[:4])
			} else {
				log.Printf("Found %s: [too short]", varName)
			}
		} else {
			log.Printf("Found %s: %s", varName, value)
		}
	}

	// Initialize R2 client
	r2Client, err := storage.NewR2ClientFromEnv()
	if err != nil {
		t.Fatalf("Failed to initialize R2 client: %v", err)
	}

	// Test R2 connectivity by checking if bucket exists
	bucket := os.Getenv("R2_BUCKET")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test R2 connectivity by trying to check if a test object exists
	// This will fail if the bucket doesn't exist or credentials are invalid
	_, err = r2Client.EmailExists(ctx, "test-connectivity-check")
	if err != nil && !strings.Contains(err.Error(), "not found") {
		t.Fatalf("Failed to connect to R2 bucket %s: %v", bucket, err)
	}

	log.Printf("Successfully connected to R2 bucket: %s", bucket)

	return &TestConfig{
		R2Client: r2Client,
		Bucket:   bucket,
	}
}

// createTestEmailData creates encrypted test email data and uploads to R2
func createTestEmailData(t *testing.T, config *TestConfig, blobID string) (string, string, string, string) {
	// Create test email content
	testContent := "This is a test email for read-once functionality"

	// Compress the content using gzip (same as the real application)
	var compressed bytes.Buffer
	gzWriter := gzip.NewWriter(&compressed)
	if _, err := gzWriter.Write([]byte(testContent)); err != nil {
		t.Fatalf("Failed to compress test content: %v", err)
	}
	if err := gzWriter.Close(); err != nil {
		t.Fatalf("Failed to close gzip writer: %v", err)
	}

	// Generate encryption key and nonce
	key := make([]byte, 32)   // AES-256 key
	nonce := make([]byte, 12) // GCM nonce

	if _, err := rand.Read(key); err != nil {
		t.Fatalf("Failed to generate encryption key: %v", err)
	}
	if _, err := rand.Read(nonce); err != nil {
		t.Fatalf("Failed to generate nonce: %v", err)
	}

	// Encrypt the content using AES-256-GCM (same as the real application)
	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatalf("Failed to create AES cipher: %v", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatalf("Failed to create GCM: %v", err)
	}

	// Encrypt the data (this includes the auth tag at the end)
	fullCiphertext := aesGCM.Seal(nil, nonce, compressed.Bytes(), nil)

	// Extract the authentication tag (last 16 bytes of ciphertext)
	authTag := fullCiphertext[len(fullCiphertext)-16:]

	// Upload the full ciphertext (including auth tag) to R2
	// The getEmailHandler expects ciphertext + auth tag to be combined
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = config.R2Client.UploadEmail(ctx, blobID, fullCiphertext)
	if err != nil {
		t.Fatalf("Failed to upload test email to R2: %v", err)
	}

	// Return base64-encoded key, nonce, and auth tag
	keyB64 := base64.StdEncoding.EncodeToString(key)
	nonceB64 := base64.StdEncoding.EncodeToString(nonce)
	authTagB64 := base64.StdEncoding.EncodeToString(authTag)

	return keyB64, nonceB64, authTagB64, ""
}

// cleanupTestEmail removes test email from R2
func cleanupTestEmail(t *testing.T, config *TestConfig, blobID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := config.R2Client.DeleteEmail(ctx, blobID)
	if err != nil {
		t.Logf("Warning: Failed to cleanup test email %s: %v", blobID, err)
	}
}

func createTestTablesForR2ReadOnce(t *testing.T) *sql.DB {
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

func generateR2ReadOnceTestToken(userID, userEmail string) (string, error) {
	sessionManager, err := auth.NewSessionManager()
	if err != nil {
		return "", err
	}
	return sessionManager.GenerateAccessToken(userID, userEmail)
}

// TestReadOnceR2Flow_Success tests the complete read-once flow with R2 storage
func TestReadOnceR2Flow_Success(t *testing.T) {
	config := setupR2IntegrationTest(t)
	if config == nil {
		return // Test was skipped
	}

	db := createTestTablesForR2ReadOnce(t)
	defer db.Close()

	// Create test email with read_once = true
	emailID := fmt.Sprintf("test-read-once-r2-%d", time.Now().Unix())
	senderID := "test-sender"
	recipient := "test@example.com"
	blobID := emailID

	// Create and upload test email data to R2
	keyB64, nonceB64, authTagB64, _ := createTestEmailData(t, config, blobID)

	// Clean up test data after test
	t.Cleanup(func() {
		cleanupTestEmail(t, config, blobID)
	})

	// Insert email metadata into database
	_, err := db.Exec(`
		INSERT INTO emails (
			email_id, sender_id, recipient, subject,
			encrypted_blob_url, encrypted_key, encryption_nonce, encryption_auth_tag,
			compression_algo, read_once, self_destruct_on_read_once
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, emailID, senderID, recipient, "Test Read-Once R2 Email",
		blobID, keyB64, nonceB64, authTagB64,
		"gzip", true, false)
	if err != nil {
		t.Fatalf("Failed to create test email: %v", err)
	}

	// Create server with test database and R2 client
	srv := &Server{db: db, r2Client: config.R2Client}

	// Generate test token
	token, err := generateR2ReadOnceTestToken(senderID, "sender@example.com")
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
	emailSecurityDB := email.NewEmailSecurityDBWithR2(db, config.R2Client)
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

// TestReadOnceR2_DeletionOnRead tests read-once with self-destruct on read
func TestReadOnceR2_DeletionOnRead(t *testing.T) {
	config := setupR2IntegrationTest(t)
	if config == nil {
		return // Test was skipped
	}

	db := createTestTablesForR2ReadOnce(t)
	defer db.Close()

	// Create test email with read_once = true and self_destruct_on_read_once = true
	emailID := fmt.Sprintf("test-read-once-delete-r2-%d", time.Now().Unix())
	senderID := "test-sender"
	recipient := "test@example.com"
	blobID := emailID

	// Create and upload test email data to R2
	keyB64, nonceB64, authTagB64, _ := createTestEmailData(t, config, blobID)

	// Clean up test data after test (in case deletion fails)
	t.Cleanup(func() {
		cleanupTestEmail(t, config, blobID)
	})

	// Insert email metadata into database
	_, err := db.Exec(`
		INSERT INTO emails (
			email_id, sender_id, recipient, subject,
			encrypted_blob_url, encrypted_key, encryption_nonce, encryption_auth_tag,
			compression_algo, read_once, self_destruct_on_read_once
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, emailID, senderID, recipient, "Test Read-Once Delete R2 Email",
		blobID, keyB64, nonceB64, authTagB64,
		"gzip", true, true)
	if err != nil {
		t.Fatalf("Failed to create test email: %v", err)
	}

	// Create server with test database and R2 client
	srv := &Server{db: db, r2Client: config.R2Client}

	// Generate test token
	token, err := generateR2ReadOnceTestToken(senderID, "sender@example.com")
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

	// Verify email is deleted from database
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM emails WHERE email_id = ?", emailID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to check email existence: %v", err)
	}

	if count != 0 {
		t.Fatal("Expected email to be deleted from database after read-once consumption with self-destruct")
	}

	// Verify email is deleted from R2 storage
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	exists, err := config.R2Client.EmailExists(ctx, blobID)
	if err != nil {
		t.Fatalf("Failed to check R2 email existence: %v", err)
	}

	if exists {
		t.Fatal("Expected email to be deleted from R2 after read-once consumption with self-destruct")
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
