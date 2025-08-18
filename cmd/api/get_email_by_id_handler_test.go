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
	"os"
	"strings"
	"testing"
	"time"

	"secure-email-mvp/pkg/auth"
	"secure-email-mvp/pkg/storage"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/pquerna/otp/totp"
)

// GetEmailByIdResponse represents the response structure for the get email by ID endpoint
type GetEmailByIdResponse struct {
	EmailID   string                 `json:"email_id"`
	SenderID  string                 `json:"sender_id,omitempty"`
	Sender    string                 `json:"sender,omitempty"`
	Recipient string                 `json:"recipient,omitempty"`
	Subject   string                 `json:"subject"`
	Body      string                 `json:"body"`
	CreatedAt string                 `json:"created_at"`
	Status    string                 `json:"status"`
	Security  map[string]interface{} `json:"security,omitempty"`
}

// TestGetEmailByIdHandler_Success tests successful email retrieval
func TestGetEmailByIdHandler_Success(t *testing.T) {
	// Load environment variables for R2 testing
	if err := godotenv.Load(); err != nil {
		t.Skip("Skipping test - .env file not found")
	}

	// Check if R2 credentials are available
	if os.Getenv("R2_ACCESS_KEY_ID") == "" || os.Getenv("R2_SECRET_ACCESS_KEY") == "" {
		t.Skip("Skipping test - R2 credentials not available")
	}

	// Setup test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := runMigrations(db); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Create test user
	userID := "1"
	userEmail := "test@example.com"
	_, err = db.Exec("INSERT INTO users (id, email, password_hash) VALUES (?, ?, ?)",
		userID, userEmail, "test-hash")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create test email
	emailID := "test-email-123"
	subject := "Test Subject"
	body := "This is a test email body"
	recipient := "recipient@example.com"

	// Encrypt and compress the email content
	encryptedContent, encryptedKey, nonce, authTag, err := encryptEmailContent(body)
	if err != nil {
		t.Fatalf("Failed to encrypt email content: %v", err)
	}

	// Upload to R2
	r2Client, err := storage.NewR2ClientFromEnv()
	if err != nil {
		t.Fatalf("Failed to create R2 client: %v", err)
	}

	ctx := context.Background()
	err = r2Client.UploadEmail(ctx, emailID, encryptedContent)
	if err != nil {
		t.Fatalf("Failed to upload email to R2: %v", err)
	}

	// Clean up R2 object after test
	t.Cleanup(func() {
		r2Client.DeleteEmail(ctx, emailID)
	})

	// Insert email record
	_, err = db.Exec(`
		INSERT INTO emails (
			email_id, sender_id, recipient, subject, encrypted_blob_url,
			encrypted_key, encryption_nonce, encryption_auth_tag, compression_algo,
			created_at, read_once, remote_revoke, self_destruct_threshold
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		emailID, userID, recipient, subject, emailID,
		encryptedKey, nonce, authTag, "gzip",
		time.Now().Unix(), false, false, 3)
	if err != nil {
		t.Fatalf("Failed to insert test email: %v", err)
	}

	// Create server with R2 client
	srv := &Server{
		db:       db,
		r2Client: r2Client,
	}

	// Create request
	req := httptest.NewRequest("GET", "/api/email/"+emailID, nil)

	// Add JWT context
	ctx = context.WithValue(req.Context(), UserIDKey, userID)
	req = req.WithContext(ctx)

	// Create response recorder
	w := httptest.NewRecorder()

	// Create router and add handler
	router := mux.NewRouter()
	router.HandleFunc("/api/email/{id}", srv.getEmailByIdHandler).Methods("GET")

	// Serve request
	router.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
		t.Logf("Response body: %s", w.Body.String())
	}

	// Parse response
	var response GetEmailByIdResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify response
	if response.EmailID != emailID {
		t.Errorf("Expected email ID %s, got %s", emailID, response.EmailID)
	}
	if response.Subject != subject {
		t.Errorf("Expected subject %s, got %s", subject, response.Subject)
	}
	if response.Body != body {
		t.Errorf("Expected body %s, got %s", body, response.Body)
	}
	if response.Recipient != recipient {
		t.Errorf("Expected recipient %s, got %s", recipient, response.Recipient)
	}
	if response.Status != "success" {
		t.Errorf("Expected status 'success', got %s", response.Status)
	}
}

// TestGetEmailByIdHandler_ReadOnceEnforcement tests read-once functionality
func TestGetEmailByIdHandler_ReadOnceEnforcement(t *testing.T) {
	// Load environment variables for R2 testing
	if err := godotenv.Load(); err != nil {
		t.Skip("Skipping test - .env file not found")
	}

	// Check if R2 credentials are available
	if os.Getenv("R2_ACCESS_KEY_ID") == "" || os.Getenv("R2_SECRET_ACCESS_KEY") == "" {
		t.Skip("Skipping test - R2 credentials not available")
	}

	// Setup test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := runMigrations(db); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Create test user
	userID := "1"
	userEmail := "test@example.com"
	_, err = db.Exec("INSERT INTO users (id, email, password_hash) VALUES (?, ?, ?)",
		userID, userEmail, "test-hash")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create test email with read-once enabled
	emailID := "test-read-once-email"
	subject := "Read-Once Test"
	body := "This is a read-once email"
	recipient := "recipient@example.com"

	// Encrypt and compress the email content
	encryptedContent, encryptedKey, nonce, authTag, err := encryptEmailContent(body)
	if err != nil {
		t.Fatalf("Failed to encrypt email content: %v", err)
	}

	// Upload to R2
	r2Client, err := storage.NewR2ClientFromEnv()
	if err != nil {
		t.Fatalf("Failed to create R2 client: %v", err)
	}

	ctx := context.Background()
	err = r2Client.UploadEmail(ctx, emailID, encryptedContent)
	if err != nil {
		t.Fatalf("Failed to upload email to R2: %v", err)
	}

	// Clean up R2 object after test
	t.Cleanup(func() {
		r2Client.DeleteEmail(ctx, emailID)
	})

	// Insert email record with read-once enabled
	_, err = db.Exec(`
		INSERT INTO emails (
			email_id, sender_id, recipient, subject, encrypted_blob_url,
			encrypted_key, encryption_nonce, encryption_auth_tag, compression_algo,
			created_at, read_once, remote_revoke, self_destruct_threshold
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		emailID, userID, recipient, subject, emailID,
		encryptedKey, nonce, authTag, "gzip",
		time.Now().Unix(), true, false, 3)
	if err != nil {
		t.Fatalf("Failed to insert test email: %v", err)
	}

	// Create server with R2 client
	srv := &Server{
		db:       db,
		r2Client: r2Client,
	}

	// First access should succeed
	req1 := httptest.NewRequest("GET", "/api/email/"+emailID, nil)
	ctx1 := context.WithValue(req1.Context(), UserIDKey, userID)
	req1 = req1.WithContext(ctx1)
	w1 := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/api/email/{id}", srv.getEmailByIdHandler).Methods("GET")
	router.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Errorf("First access expected status 200, got %d", w1.Code)
		t.Logf("First response body: %s", w1.Body.String())
	}

	// Second access should fail (read-once consumed)
	req2 := httptest.NewRequest("GET", "/api/email/"+emailID, nil)
	ctx2 := context.WithValue(req2.Context(), UserIDKey, userID)
	req2 = req2.WithContext(ctx2)
	w2 := httptest.NewRecorder()

	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusForbidden {
		t.Errorf("Second access expected status 403, got %d", w2.Code)
	}

	var errorResponse map[string]string
	if err := json.NewDecoder(w2.Body).Decode(&errorResponse); err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}

	if errorResponse["error"] != "Email has been revoked or cannot be accessed" {
		t.Errorf("Expected generic error message, got %s", errorResponse["error"])
	}
}

// TestGetEmailByIdHandler_SelfDestructEnforcement tests self-destruct functionality
func TestGetEmailByIdHandler_SelfDestructEnforcement(t *testing.T) {
	// Load environment variables for R2 testing
	if err := godotenv.Load(); err != nil {
		t.Skip("Skipping test - .env file not found")
	}

	// Check if R2 credentials are available
	if os.Getenv("R2_ACCESS_KEY_ID") == "" || os.Getenv("R2_SECRET_ACCESS_KEY") == "" {
		t.Skip("Skipping test - R2 credentials not available")
	}

	// Setup test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := runMigrations(db); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Create test users
	senderID := "1"
	senderEmail := "sender@example.com"
	_, err = db.Exec("INSERT INTO users (id, email, password_hash) VALUES (?, ?, ?)",
		senderID, senderEmail, "test-hash")
	if err != nil {
		t.Fatalf("Failed to create test sender: %v", err)
	}

	unauthorizedUserID := "2"
	unauthorizedEmail := "unauthorized@example.com"
	_, err = db.Exec("INSERT INTO users (id, email, password_hash) VALUES (?, ?, ?)",
		unauthorizedUserID, unauthorizedEmail, "test-hash")
	if err != nil {
		t.Fatalf("Failed to create test unauthorized user: %v", err)
	}

	// Create test email with low self-destruct threshold
	emailID := "test-self-destruct-email"
	subject := "Self-Destruct Test"
	body := "This is a self-destruct email"
	recipient := "recipient@example.com"

	// Encrypt and compress the email content
	encryptedContent, encryptedKey, nonce, authTag, err := encryptEmailContent(body)
	if err != nil {
		t.Fatalf("Failed to encrypt email content: %v", err)
	}

	// Upload to R2
	r2Client, err := storage.NewR2ClientFromEnv()
	if err != nil {
		t.Fatalf("Failed to create R2 client: %v", err)
	}

	ctx := context.Background()
	err = r2Client.UploadEmail(ctx, emailID, encryptedContent)
	if err != nil {
		t.Fatalf("Failed to upload email to R2: %v", err)
	}

	// Clean up R2 object after test
	t.Cleanup(func() {
		r2Client.DeleteEmail(ctx, emailID)
	})

	// Insert email record with low self-destruct threshold
	_, err = db.Exec(`
		INSERT INTO emails (
			email_id, sender_id, recipient, subject, encrypted_blob_url,
			encrypted_key, encryption_nonce, encryption_auth_tag, compression_algo,
			created_at, read_once, remote_revoke, self_destruct_threshold, failed_attempts
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		emailID, senderID, recipient, subject, emailID,
		encryptedKey, nonce, authTag, "gzip",
		time.Now().Unix(), false, false, 2, 1) // Threshold 2, already 1 failed attempt
	if err != nil {
		t.Fatalf("Failed to insert test email: %v", err)
	}

	// Create server with R2 client
	srv := &Server{
		db:       db,
		r2Client: r2Client,
	}

	// Attempt unauthorized access (should trigger self-destruct)
	req := httptest.NewRequest("GET", "/api/email/"+emailID, nil)
	ctx = context.WithValue(req.Context(), UserIDKey, unauthorizedUserID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/api/email/{id}", srv.getEmailByIdHandler).Methods("GET")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}

	var errorResponse map[string]string
	if err := json.NewDecoder(w.Body).Decode(&errorResponse); err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}

	if errorResponse["error"] != "Email has been revoked or cannot be accessed" {
		t.Errorf("Expected generic error message, got %s", errorResponse["error"])
	}

	// Verify email was deleted from database
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM emails WHERE email_id = ?", emailID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to check email existence: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected email to be deleted from database, but %d records found", count)
	}

	// Verify R2 object was deleted
	exists, err := r2Client.EmailExists(ctx, emailID)
	if err != nil {
		t.Fatalf("Failed to check R2 object existence: %v", err)
	}

	if exists {
		t.Errorf("Expected R2 object to be deleted, but it still exists")
	}
}

// TestGetEmailByIdHandler_RemoteRevoke tests remote revoke functionality
func TestGetEmailByIdHandler_RemoteRevoke(t *testing.T) {
	// Load environment variables for R2 testing
	if err := godotenv.Load(); err != nil {
		t.Skip("Skipping test - .env file not found")
	}

	// Check if R2 credentials are available
	if os.Getenv("R2_ACCESS_KEY_ID") == "" || os.Getenv("R2_SECRET_ACCESS_KEY") == "" {
		t.Skip("Skipping test - R2 credentials not available")
	}

	// Setup test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := runMigrations(db); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Create test user
	userID := "1"
	userEmail := "test@example.com"
	_, err = db.Exec("INSERT INTO users (id, email, password_hash) VALUES (?, ?, ?)",
		userID, userEmail, "test-hash")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create test email with remote revoke enabled
	emailID := "test-remote-revoke-email"
	subject := "Remote Revoke Test"
	body := "This is a remote revoked email"
	recipient := "recipient@example.com"

	// Encrypt and compress the email content
	encryptedContent, encryptedKey, nonce, authTag, err := encryptEmailContent(body)
	if err != nil {
		t.Fatalf("Failed to encrypt email content: %v", err)
	}

	// Upload to R2
	r2Client, err := storage.NewR2ClientFromEnv()
	if err != nil {
		t.Fatalf("Failed to create R2 client: %v", err)
	}

	ctx := context.Background()
	err = r2Client.UploadEmail(ctx, emailID, encryptedContent)
	if err != nil {
		t.Fatalf("Failed to upload email to R2: %v", err)
	}

	// Clean up R2 object after test
	t.Cleanup(func() {
		r2Client.DeleteEmail(ctx, emailID)
	})

	// Insert email record with remote revoke enabled
	_, err = db.Exec(`
		INSERT INTO emails (
			email_id, sender_id, recipient, subject, encrypted_blob_url,
			encrypted_key, encryption_nonce, encryption_auth_tag, compression_algo,
			created_at, read_once, remote_revoke, self_destruct_threshold
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		emailID, userID, recipient, subject, emailID,
		encryptedKey, nonce, authTag, "gzip",
		time.Now().Unix(), false, true, 3)
	if err != nil {
		t.Fatalf("Failed to insert test email: %v", err)
	}

	// Create server with R2 client
	srv := &Server{
		db:       db,
		r2Client: r2Client,
	}

	// Attempt access (should be denied due to remote revoke)
	req := httptest.NewRequest("GET", "/api/email/"+emailID, nil)
	ctx = context.WithValue(req.Context(), UserIDKey, userID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/api/email/{id}", srv.getEmailByIdHandler).Methods("GET")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}

	var errorResponse map[string]string
	if err := json.NewDecoder(w.Body).Decode(&errorResponse); err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}

	if errorResponse["error"] != "Email has been revoked or cannot be accessed" {
		t.Errorf("Expected generic error message, got %s", errorResponse["error"])
	}
}

// TestGetEmailByIdHandler_TimeLock tests time lock functionality
func TestGetEmailByIdHandler_TimeLock(t *testing.T) {
	// Load environment variables for R2 testing
	if err := godotenv.Load(); err != nil {
		t.Skip("Skipping test - .env file not found")
	}

	// Check if R2 credentials are available
	if os.Getenv("R2_ACCESS_KEY_ID") == "" || os.Getenv("R2_SECRET_ACCESS_KEY") == "" {
		t.Skip("Skipping test - R2 credentials not available")
	}

	// Setup test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := runMigrations(db); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Create test user
	userID := "1"
	userEmail := "test@example.com"
	_, err = db.Exec("INSERT INTO users (id, email, password_hash) VALUES (?, ?, ?)",
		userID, userEmail, "test-hash")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create test email with time lock (future time)
	emailID := "test-time-lock-email"
	subject := "Time Lock Test"
	body := "This is a time-locked email"
	recipient := "recipient@example.com"
	futureTime := time.Now().Add(1 * time.Hour).Unix() // Lock until 1 hour from now

	// Encrypt and compress the email content
	encryptedContent, encryptedKey, nonce, authTag, err := encryptEmailContent(body)
	if err != nil {
		t.Fatalf("Failed to encrypt email content: %v", err)
	}

	// Upload to R2
	r2Client, err := storage.NewR2ClientFromEnv()
	if err != nil {
		t.Fatalf("Failed to create R2 client: %v", err)
	}

	ctx := context.Background()
	err = r2Client.UploadEmail(ctx, emailID, encryptedContent)
	if err != nil {
		t.Fatalf("Failed to upload email to R2: %v", err)
	}

	// Clean up R2 object after test
	t.Cleanup(func() {
		r2Client.DeleteEmail(ctx, emailID)
	})

	// Insert email record with time lock
	_, err = db.Exec(`
		INSERT INTO emails (
			email_id, sender_id, recipient, subject, encrypted_blob_url,
			encrypted_key, encryption_nonce, encryption_auth_tag, compression_algo,
			created_at, read_once, remote_revoke, self_destruct_threshold, not_before
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		emailID, userID, recipient, subject, emailID,
		encryptedKey, nonce, authTag, "gzip",
		time.Now().Unix(), false, false, 3, futureTime)
	if err != nil {
		t.Fatalf("Failed to insert test email: %v", err)
	}

	// Create server with R2 client
	srv := &Server{
		db:       db,
		r2Client: r2Client,
	}

	// Attempt access (should be denied due to time lock)
	req := httptest.NewRequest("GET", "/api/email/"+emailID, nil)
	ctx = context.WithValue(req.Context(), UserIDKey, userID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/api/email/{id}", srv.getEmailByIdHandler).Methods("GET")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}

	var errorResponse map[string]string
	if err := json.NewDecoder(w.Body).Decode(&errorResponse); err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}

	if errorResponse["error"] != "Email has been revoked or cannot be accessed" {
		t.Errorf("Expected generic error message, got %s", errorResponse["error"])
	}
}

// TestGetEmailByIdHandler_Expired tests expiration functionality
func TestGetEmailByIdHandler_Expired(t *testing.T) {
	// Load environment variables for R2 testing
	if err := godotenv.Load(); err != nil {
		t.Skip("Skipping test - .env file not found")
	}

	// Check if R2 credentials are available
	if os.Getenv("R2_ACCESS_KEY_ID") == "" || os.Getenv("R2_SECRET_ACCESS_KEY") == "" {
		t.Skip("Skipping test - R2 credentials not available")
	}

	// Setup test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := runMigrations(db); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Create test user
	userID := "1"
	userEmail := "test@example.com"
	_, err = db.Exec("INSERT INTO users (id, email, password_hash) VALUES (?, ?, ?)",
		userID, userEmail, "test-hash")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create test email with expiration (past time)
	emailID := "test-expired-email"
	subject := "Expired Test"
	body := "This is an expired email"
	recipient := "recipient@example.com"
	pastTime := time.Now().Add(-1 * time.Hour).Unix() // Expired 1 hour ago

	// Encrypt and compress the email content
	encryptedContent, encryptedKey, nonce, authTag, err := encryptEmailContent(body)
	if err != nil {
		t.Fatalf("Failed to encrypt email content: %v", err)
	}

	// Upload to R2
	r2Client, err := storage.NewR2ClientFromEnv()
	if err != nil {
		t.Fatalf("Failed to create R2 client: %v", err)
	}

	ctx := context.Background()
	err = r2Client.UploadEmail(ctx, emailID, encryptedContent)
	if err != nil {
		t.Fatalf("Failed to upload email to R2: %v", err)
	}

	// Clean up R2 object after test
	t.Cleanup(func() {
		r2Client.DeleteEmail(ctx, emailID)
	})

	// Insert email record with expiration
	_, err = db.Exec(`
		INSERT INTO emails (
			email_id, sender_id, recipient, subject, encrypted_blob_url,
			encrypted_key, encryption_nonce, encryption_auth_tag, compression_algo,
			created_at, read_once, remote_revoke, self_destruct_threshold, expires_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		emailID, userID, recipient, subject, emailID,
		encryptedKey, nonce, authTag, "gzip",
		time.Now().Unix(), false, false, 3, pastTime)
	if err != nil {
		t.Fatalf("Failed to insert test email: %v", err)
	}

	// Create server with R2 client
	srv := &Server{
		db:       db,
		r2Client: r2Client,
	}

	// Attempt access (should be denied due to expiration)
	req := httptest.NewRequest("GET", "/api/email/"+emailID, nil)
	ctx = context.WithValue(req.Context(), UserIDKey, userID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/api/email/{id}", srv.getEmailByIdHandler).Methods("GET")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}

	var errorResponse map[string]string
	if err := json.NewDecoder(w.Body).Decode(&errorResponse); err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}

	if errorResponse["error"] != "Email has been revoked or cannot be accessed" {
		t.Errorf("Expected generic error message, got %s", errorResponse["error"])
	}
}

// TestGetEmailByIdHandler_SecurityTogglesResponse tests that security toggles are returned for sender
func TestGetEmailByIdHandler_SecurityTogglesResponse(t *testing.T) {
	// Load environment variables for R2 testing
	if err := godotenv.Load(); err != nil {
		t.Skip("Skipping test - .env file not found")
	}

	// Check if R2 credentials are available
	if os.Getenv("R2_ACCESS_KEY_ID") == "" || os.Getenv("R2_SECRET_ACCESS_KEY") == "" {
		t.Skip("Skipping test - R2 credentials not available")
	}

	// Setup test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := runMigrations(db); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Create test user
	userID := "1"
	userEmail := "test@example.com"
	_, err = db.Exec("INSERT INTO users (id, email, password_hash) VALUES (?, ?, ?)",
		userID, userEmail, "test-hash")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create test email with security toggles
	emailID := "test-security-toggles-email"
	subject := "Security Toggles Test"
	body := "This is a test email with security toggles"
	recipient := "recipient@example.com"

	// Encrypt and compress the email content
	encryptedContent, encryptedKey, nonce, authTag, err := encryptEmailContent(body)
	if err != nil {
		t.Fatalf("Failed to encrypt email content: %v", err)
	}

	// Upload to R2
	r2Client, err := storage.NewR2ClientFromEnv()
	if err != nil {
		t.Fatalf("Failed to create R2 client: %v", err)
	}

	ctx := context.Background()
	err = r2Client.UploadEmail(ctx, emailID, encryptedContent)
	if err != nil {
		t.Fatalf("Failed to upload email to R2: %v", err)
	}

	// Clean up R2 object after test
	t.Cleanup(func() {
		r2Client.DeleteEmail(ctx, emailID)
	})

	// Insert email record with security toggles
	_, err = db.Exec(`
		INSERT INTO emails (
			email_id, sender_id, recipient, subject, encrypted_blob_url,
			encrypted_key, encryption_nonce, encryption_auth_tag, compression_algo,
			created_at, read_once, remote_revoke, self_destruct_threshold, mfa_on_open
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		emailID, userID, recipient, subject, emailID,
		encryptedKey, nonce, authTag, "gzip",
		time.Now().Unix(), true, false, 5, true)
	if err != nil {
		t.Fatalf("Failed to insert test email: %v", err)
	}

	// Create server with R2 client
	srv := &Server{
		db:       db,
		r2Client: r2Client,
	}

	// Access as sender (should include security toggles)
	req := httptest.NewRequest("GET", "/api/email/"+emailID, nil)
	ctx = context.WithValue(req.Context(), UserIDKey, userID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/api/email/{id}", srv.getEmailByIdHandler).Methods("GET")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
		t.Logf("Response body: %s", w.Body.String())
	}

	// Parse response
	var response GetEmailByIdResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify security toggles are included
	if response.Security == nil {
		t.Errorf("Expected security toggles to be included in response")
	} else {
		if !response.Security["readOnce"].(bool) {
			t.Errorf("Expected ReadOnce to be true")
		}
		if response.Security["remoteRevoke"].(bool) {
			t.Errorf("Expected RemoteRevoke to be false")
		}
		if response.Security["selfDestructThreshold"].(int) != 5 {
			t.Errorf("Expected SelfDestructThreshold to be 5, got %d", response.Security["selfDestructThreshold"].(int))
		}
		if response.Security["requiresMFA"].(bool) {
			t.Errorf("Expected MFAOnOpen to be true")
		}
	}
}

// TestGetEmailByIdHandler_NoAuth tests that the handler properly rejects requests without authentication
func TestGetEmailByIdHandler_NoAuth(t *testing.T) {
	// Setup test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := runMigrations(db); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Create server without R2 client
	srv := &Server{
		db: db,
	}

	// Create request without JWT context
	req := httptest.NewRequest("GET", "/api/email/test-email", nil)
	w := httptest.NewRecorder()

	// Create router and add handler
	router := mux.NewRouter()
	router.HandleFunc("/api/email/{id}", srv.getEmailByIdHandler).Methods("GET")

	// Serve request
	router.ServeHTTP(w, req)

	// Check response - should be 401 Unauthorized
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
		t.Logf("Response body: %s", w.Body.String())
	}

	var errorResponse map[string]string
	if err := json.NewDecoder(w.Body).Decode(&errorResponse); err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}

	if errorResponse["error"] != "Authentication required" {
		t.Errorf("Expected error 'Authentication required', got %s", errorResponse["error"])
	}
}

// TestGetEmailByIdHandler_InvalidEmailID tests that the handler properly handles invalid email IDs
func TestGetEmailByIdHandler_InvalidEmailID(t *testing.T) {
	// Setup test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := runMigrations(db); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Create server without R2 client
	srv := &Server{
		db: db,
	}

	// Create request with JWT context but empty email ID
	req := httptest.NewRequest("GET", "/api/email/", nil) // Empty email ID
	ctx := context.WithValue(req.Context(), UserIDKey, "1")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	// Create router and add handler
	router := mux.NewRouter()
	router.HandleFunc("/api/email/{id}", srv.getEmailByIdHandler).Methods("GET")

	// Serve request
	router.ServeHTTP(w, req)

	// Check response - should be 404 Not Found (router doesn't match empty path)
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
		t.Logf("Response body: %s", w.Body.String())
	}
}

// TestGetEmailByIdHandler_MFARequired tests MFA enforcement functionality
func TestGetEmailByIdHandler_MFARequired(t *testing.T) {
	// Setup test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := runMigrations(db); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Create test user
	userID := "1"
	userEmail := "test@example.com"
	_, err = db.Exec("INSERT INTO users (id, email, password_hash) VALUES (?, ?, ?)",
		userID, userEmail, "test-hash")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create test email with MFA required
	emailID := "test-mfa-email"
	subject := "MFA Test"
	recipient := "recipient@example.com"
	totpSecret := "JBSWY3DPEHPK3PXP" // Test TOTP secret

	// Insert email record with MFA enabled
	_, err = db.Exec(`
		INSERT INTO emails (
			email_id, sender_id, recipient, subject, encrypted_blob_url,
			encrypted_key, encryption_nonce, encryption_auth_tag, compression_algo,
			created_at, mfa_on_open, totp_secret
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		emailID, userID, recipient, subject, emailID,
		"test-key", "test-nonce", "test-auth-tag", "gzip",
		time.Now().Unix(), true, totpSecret)
	if err != nil {
		t.Fatalf("Failed to insert test email: %v", err)
	}

	// Create server without R2 client
	srv := &Server{
		db: db,
	}

	// Test 1: Access without TOTP code (should fail)
	req1 := httptest.NewRequest("GET", "/api/email/"+emailID, nil)
	ctx1 := context.WithValue(req1.Context(), UserIDKey, userID)
	req1 = req1.WithContext(ctx1)
	w1 := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/api/email/{id}", srv.getEmailByIdHandler).Methods("GET")
	router.ServeHTTP(w1, req1)

	if w1.Code != http.StatusForbidden {
		t.Errorf("Expected status 403 for missing TOTP, got %d", w1.Code)
		t.Logf("Response body: %s", w1.Body.String())
	}

	// Test 2: Access with invalid TOTP code (should fail)
	req2 := httptest.NewRequest("GET", "/api/email/"+emailID+"?totp_code=123456", nil)
	ctx2 := context.WithValue(req2.Context(), UserIDKey, userID)
	req2 = req2.WithContext(ctx2)
	w2 := httptest.NewRecorder()

	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusForbidden {
		t.Errorf("Expected status 403 for invalid TOTP, got %d", w2.Code)
		t.Logf("Response body: %s", w2.Body.String())
	}

	// Test 3: Access with valid TOTP code (should succeed)
	validTOTP, err := totp.GenerateCode(totpSecret, time.Now())
	if err != nil {
		t.Fatalf("Failed to generate valid TOTP: %v", err)
	}

	req3 := httptest.NewRequest("GET", "/api/email/"+emailID+"?totp_code="+validTOTP, nil)
	ctx3 := context.WithValue(req3.Context(), UserIDKey, userID)
	req3 = req3.WithContext(ctx3)
	w3 := httptest.NewRecorder()

	router.ServeHTTP(w3, req3)

	// Should fail because we don't have R2 setup, but MFA validation should pass
	if w3.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500 for R2 error, got %d", w3.Code)
		t.Logf("Response body: %s", w3.Body.String())
	}
}

// TestGetEmailByIdHandler_DecoyMessage tests decoy message functionality
func TestGetEmailByIdHandler_DecoyMessage(t *testing.T) {
	// Setup test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := runMigrations(db); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Create test user
	userID := "1"
	userEmail := "test@example.com"
	_, err = db.Exec("INSERT INTO users (id, email, password_hash) VALUES (?, ?, ?)",
		userID, userEmail, "test-hash")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create test email with decoy secret
	emailID := "test-decoy-email"
	subject := "Decoy Test"
	recipient := "recipient@example.com"
	totpSecret := "JBSWY3DPEHPK3PXP"

	// Hash the decoy secret using Argon2id (same as the handler expects)
	decoySecretHash, err := auth.HashPassword("decoy123", emailID)
	if err != nil {
		t.Fatalf("Failed to hash decoy secret: %v", err)
	}

	// Insert email record with MFA and decoy enabled
	_, err = db.Exec(`
		INSERT INTO emails (
			email_id, sender_id, recipient, subject, encrypted_blob_url,
			encrypted_key, encryption_nonce, encryption_auth_tag, compression_algo,
			created_at, mfa_on_open, totp_secret, decoy_secret
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		emailID, userID, recipient, subject, emailID,
		"test-key", "test-nonce", "test-auth-tag", "gzip",
		time.Now().Unix(), true, totpSecret, decoySecretHash)
	if err != nil {
		t.Fatalf("Failed to insert test email: %v", err)
	}

	// Create server without R2 client
	srv := &Server{
		db: db,
	}

	// Test: Access with decoy TOTP code (should return decoy email)
	req := httptest.NewRequest("GET", "/api/email/"+emailID+"?totp_code=decoy123", nil)
	ctx := context.WithValue(req.Context(), UserIDKey, userID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/api/email/{id}", srv.getEmailByIdHandler).Methods("GET")
	router.ServeHTTP(w, req)

	// Should return decoy email (status 200 with decoy content)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for decoy email, got %d", w.Code)
		t.Logf("Response body: %s", w.Body.String())
	}

	// Parse response to verify it's a decoy
	var response GetEmailByIdResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify decoy email structure
	if !strings.HasPrefix(response.EmailID, "decoy-") {
		t.Errorf("Expected decoy email ID, got %s", response.EmailID)
	}
	if response.Sender != "system@securesystem.email" {
		t.Errorf("Expected decoy sender, got %s", response.Sender)
	}
	if response.Subject != "System Notification" {
		t.Errorf("Expected decoy subject, got %s", response.Subject)
	}
}

// Helper function to encrypt email content for testing
func encryptEmailContent(content string) ([]byte, string, string, string, error) {
	// Generate a random AES key
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, "", "", "", err
	}

	// Compress the content
	var compressed bytes.Buffer
	gzWriter := gzip.NewWriter(&compressed)
	if _, err := gzWriter.Write([]byte(content)); err != nil {
		return nil, "", "", "", err
	}
	gzWriter.Close()

	// Encrypt the compressed content
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, "", "", "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, "", "", "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, "", "", "", err
	}

	ciphertext := gcm.Seal(nil, nonce, compressed.Bytes(), nil)

	// Extract auth tag (last 16 bytes)
	authTag := ciphertext[len(ciphertext)-16:]
	ciphertextOnly := ciphertext[:len(ciphertext)-16]

	// Combine ciphertext and auth tag for storage
	encryptedBlob := append(ciphertextOnly, authTag...)

	// Encode as base64
	encryptedKeyB64 := base64.StdEncoding.EncodeToString(key)
	nonceB64 := base64.StdEncoding.EncodeToString(nonce)
	authTagB64 := base64.StdEncoding.EncodeToString(authTag)

	return encryptedBlob, encryptedKeyB64, nonceB64, authTagB64, nil
}

// Helper function to run migrations
func runMigrations(db *sql.DB) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
		)`,
		`CREATE TABLE IF NOT EXISTS emails (
			email_id TEXT PRIMARY KEY,
			sender_id TEXT NOT NULL,
			recipient TEXT NOT NULL,
			subject TEXT NOT NULL,
			encrypted_blob_url TEXT NOT NULL,
			encrypted_key TEXT NOT NULL,
			encryption_nonce TEXT NOT NULL,
			encryption_auth_tag TEXT NOT NULL,
			compression_algo TEXT NOT NULL DEFAULT 'gzip',
			created_at INTEGER NOT NULL,
			last_access_at INTEGER,
			access_count INTEGER DEFAULT 0,
			fail_count INTEGER DEFAULT 0,
			not_before INTEGER,
			expires_at INTEGER,
			read_once BOOLEAN DEFAULT FALSE,
			mfa_on_open BOOLEAN DEFAULT FALSE,
			decoy_secret TEXT,
			totp_secret TEXT,
			remote_revoke BOOLEAN DEFAULT FALSE,
			strip_metadata BOOLEAN DEFAULT FALSE,
			self_destruct_threshold INTEGER DEFAULT 3,
			geo_rules_ref TEXT,
			self_destruct_on_read_once BOOLEAN DEFAULT FALSE,
			failed_attempts INTEGER DEFAULT 0,
			read_once_consumed_at INTEGER,
			read_once_consumer_device TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS audit_log (
			log_id TEXT PRIMARY KEY,
			timestamp INTEGER NOT NULL,
			event_type TEXT NOT NULL,
			user_id TEXT,
			ip_address TEXT,
			user_agent TEXT,
			related_email_id TEXT,
			outcome TEXT NOT NULL,
			details TEXT,
			severity TEXT DEFAULT 'info',
			country TEXT,
			city TEXT
		)`,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return err
		}
	}

	return nil
}
