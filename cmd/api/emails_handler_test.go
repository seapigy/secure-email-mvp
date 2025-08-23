package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	_ "modernc.org/sqlite"
)

// Helper function to generate JWT token for testing
func generateTestToken(t *testing.T, userID, email string) string {
	// Set JWT secrets for testing
	os.Setenv("JWT_SECRET", "test-secret-key-for-jwt-signing")
	os.Setenv("JWT_ACCESS_SECRET", "test-access-secret-key-for-testing-only")

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	})
	tokenString, err := token.SignedString([]byte("test-secret-key-for-jwt-signing"))
	if err != nil {
		t.Fatalf("Failed to generate JWT token: %v", err)
	}
	return tokenString
}

// TestEmailsHandlerAuthentication tests that JWT authentication is required
func TestEmailsHandlerAuthentication(t *testing.T) {
	// Setup test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Create test tables
	if err := createTestTablesForEmails(db); err != nil {
		t.Fatalf("Failed to create test tables: %v", err)
	}

	// Create server with test database
	srv := &Server{db: db}

	// Test without Authorization header
	req := httptest.NewRequest("GET", "/api/emails", nil)
	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// For authentication tests, don't set user context - let the handler check for authentication
		srv.emailsHandler(w, r)
	})
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401 without token, got %d", w.Code)
	}

	// Test with invalid Authorization header
	req = httptest.NewRequest("GET", "/api/emails", nil)
	req.Header.Set("Authorization", "Invalid")
	w = httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401 with invalid header, got %d", w.Code)
	}

	// Test with empty token
	req = httptest.NewRequest("GET", "/api/emails", nil)
	req.Header.Set("Authorization", "Bearer ")
	w = httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401 with empty token, got %d", w.Code)
	}
}

// TestEmailsHandlerUserIsolation tests that users only get their own emails
func TestEmailsHandlerUserIsolation(t *testing.T) {
	// Setup test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Create test tables
	if err := createTestTablesForEmails(db); err != nil {
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
		t.Fatalf("Failed to insert other user: %v", err)
	}

	// Insert test emails
	now := time.Now()

	// Emails sent by test user
	_, err = db.Exec(`
		INSERT INTO emails (email_id, sender_id, recipient, subject, encrypted_blob_url, 
			encrypted_key, encryption_nonce, encryption_auth_tag, sha256_hash, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"email-1", testUserID, "recipient1@example.com", "Test Email 1",
		"blob-1", "key-1", "nonce-1", "tag-1", "hash-1", now.Add(-1*time.Hour))
	if err != nil {
		t.Fatalf("Failed to insert test email 1: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO emails (email_id, sender_id, recipient, subject, encrypted_blob_url, 
			encrypted_key, encryption_nonce, encryption_auth_tag, sha256_hash, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"email-2", testUserID, "recipient2@example.com", "Test Email 2",
		"blob-2", "key-2", "nonce-2", "tag-2", "hash-2", now.Add(-2*time.Hour))
	if err != nil {
		t.Fatalf("Failed to insert test email 2: %v", err)
	}

	// Email received by test user
	_, err = db.Exec(`
		INSERT INTO emails (email_id, sender_id, recipient, subject, encrypted_blob_url, 
			encrypted_key, encryption_nonce, encryption_auth_tag, sha256_hash, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"email-3", otherUserID, testUserEmail, "Email to Test User",
		"blob-3", "key-3", "nonce-3", "tag-3", "hash-3", now.Add(-30*time.Minute))
	if err != nil {
		t.Fatalf("Failed to insert received email: %v", err)
	}

	// Email for other user (should not appear in results)
	_, err = db.Exec(`
		INSERT INTO emails (email_id, sender_id, recipient, subject, encrypted_blob_url, 
			encrypted_key, encryption_nonce, encryption_auth_tag, sha256_hash, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"email-4", otherUserID, "someone@example.com", "Other User Email",
		"blob-4", "key-4", "nonce-4", "tag-4", "hash-4", now.Add(-4*time.Hour))
	if err != nil {
		t.Fatalf("Failed to insert other user email: %v", err)
	}

	// Create server with test database
	srv := &Server{db: db}

	// Generate JWT token for test user
	token := generateTestToken(t, testUserID, testUserEmail)

	// Create request
	req := httptest.NewRequest("GET", "/api/emails", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Create response recorder
	w := httptest.NewRecorder()

	// Apply test middleware and call handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set proper user context for authenticated test
		ctx := context.WithValue(r.Context(), UserIDKey, testUserID)
		ctx = context.WithValue(ctx, EmailKey, testUserEmail)
		r = r.WithContext(ctx)
		srv.emailsHandler(w, r)
	})
	handler.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
		t.Logf("Response body: %s", w.Body.String())
		return
	}

	// Parse response
	var response EmailsResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify response structure
	if response.Status != "success" {
		t.Errorf("Expected status 'success', got '%s'", response.Status)
	}

	// Verify email count (should be 3 emails: 2 sent + 1 received)
	if len(response.Emails) != 3 {
		t.Errorf("Expected 3 emails, got %d", len(response.Emails))
	}

	// Verify emails belong to test user and are ordered by created_at DESC
	expectedEmails := []string{"email-3", "email-1", "email-2"} // Most recent first
	for i, email := range response.Emails {
		if email.ID != expectedEmails[i] {
			t.Errorf("Expected email ID '%s' at position %d, got '%s'", expectedEmails[i], i, email.ID)
		}
	}

	// Verify no emails from other user are included
	for _, email := range response.Emails {
		if email.ID == "email-4" {
			t.Error("Found email from other user that should not be included")
		}
	}

	// Verify sender email is populated correctly
	for _, email := range response.Emails {
		if email.SenderEmail == "" {
			t.Error("Expected sender_email to be populated")
		}
	}
}

// TestEmailsHandlerSorting tests that emails are sorted by created_at descending
func TestEmailsHandlerSorting(t *testing.T) {
	// Setup test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Create test tables
	if err := createTestTablesForEmails(db); err != nil {
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

	// Insert test emails with different timestamps
	now := time.Now()

	emails := []struct {
		id        string
		createdAt time.Time
	}{
		{"email-oldest", now.Add(-3 * time.Hour)},
		{"email-middle", now.Add(-2 * time.Hour)},
		{"email-newest", now.Add(-1 * time.Hour)},
	}

	for _, email := range emails {
		_, err = db.Exec(`
			INSERT INTO emails (email_id, sender_id, recipient, subject, encrypted_blob_url, 
				encrypted_key, encryption_nonce, encryption_auth_tag, sha256_hash, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			email.id, testUserID, "recipient@example.com", "Test Email",
			"blob-"+email.id, "key-"+email.id, "nonce-"+email.id, "tag-"+email.id,
			"hash-"+email.id, email.createdAt)
		if err != nil {
			t.Fatalf("Failed to insert email %s: %v", email.id, err)
		}
	}

	// Create server with test database
	srv := &Server{db: db}

	// Generate JWT token for test user
	token := generateTestToken(t, testUserID, testUserEmail)

	// Create request
	req := httptest.NewRequest("GET", "/api/emails", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Create response recorder
	w := httptest.NewRecorder()

	// Apply test middleware and call handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set proper user context for authenticated test
		ctx := context.WithValue(r.Context(), UserIDKey, testUserID)
		ctx = context.WithValue(ctx, EmailKey, testUserEmail)
		r = r.WithContext(ctx)
		srv.emailsHandler(w, r)
	})
	handler.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
		return
	}

	// Parse response
	var response EmailsResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify email count
	if len(response.Emails) != 3 {
		t.Errorf("Expected 3 emails, got %d", len(response.Emails))
		return
	}

	// Verify sorting order (newest first)
	expectedOrder := []string{"email-newest", "email-middle", "email-oldest"}
	for i, email := range response.Emails {
		if email.ID != expectedOrder[i] {
			t.Errorf("Expected email ID '%s' at position %d, got '%s'", expectedOrder[i], i, email.ID)
		}
	}

	// Verify timestamps are in descending order
	for i := 0; i < len(response.Emails)-1; i++ {
		if response.Emails[i].CreatedAt.Before(response.Emails[i+1].CreatedAt) {
			t.Errorf("Emails not sorted correctly: %s (%v) comes before %s (%v)",
				response.Emails[i].ID, response.Emails[i].CreatedAt,
				response.Emails[i+1].ID, response.Emails[i+1].CreatedAt)
		}
	}
}

// TestEmailsHandlerEmptyResults tests correct handling of empty results
func TestEmailsHandlerEmptyResults(t *testing.T) {
	// Setup test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Create test tables
	if err := createTestTablesForEmails(db); err != nil {
		t.Fatalf("Failed to create test tables: %v", err)
	}

	// Create test user with no emails
	testUserID := "user-1"
	testUserEmail := "test@securesystem.email"

	_, err = db.Exec(`INSERT INTO users (id, email, password_hash) VALUES (?, ?, ?)`,
		testUserID, testUserEmail, "hash1")
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	// Create server with test database
	srv := &Server{db: db}

	// Generate JWT token for test user
	token := generateTestToken(t, testUserID, testUserEmail)

	// Create request
	req := httptest.NewRequest("GET", "/api/emails", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Create response recorder
	w := httptest.NewRecorder()

	// Apply test middleware and call handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set proper user context for authenticated test
		ctx := context.WithValue(r.Context(), UserIDKey, testUserID)
		ctx = context.WithValue(ctx, EmailKey, testUserEmail)
		r = r.WithContext(ctx)
		srv.emailsHandler(w, r)
	})
	handler.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
		return
	}

	// Parse response
	var response EmailsResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify response structure
	if response.Status != "success" {
		t.Errorf("Expected status 'success', got '%s'", response.Status)
	}

	// Verify empty results
	if len(response.Emails) != 0 {
		t.Errorf("Expected 0 emails, got %d", len(response.Emails))
	}
}

// TestEmailsHandlerDatabaseError tests graceful handling of database errors
func TestEmailsHandlerDatabaseError(t *testing.T) {
	// Create server with nil database to simulate database error
	srv := &Server{db: nil}

	// Generate JWT token
	token := generateTestToken(t, "test-user-id", "test@securesystem.email")

	// Create request
	req := httptest.NewRequest("GET", "/api/emails", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Create response recorder
	w := httptest.NewRecorder()

	// Apply test middleware and call handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set proper user context for authenticated test
		ctx := context.WithValue(r.Context(), UserIDKey, "test-user-id")
		ctx = context.WithValue(ctx, EmailKey, "test@securesystem.email")
		r = r.WithContext(ctx)
		srv.emailsHandler(w, r)
	})
	handler.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	// Verify error message
	var errorResponse map[string]string
	if err := json.NewDecoder(w.Body).Decode(&errorResponse); err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}

	if errorResponse["error"] != "Database connection unavailable" {
		t.Errorf("Expected error 'Database connection unavailable', got '%s'", errorResponse["error"])
	}
}

// Helper function to create test tables for emails tests
func createTestTablesForEmails(db *sql.DB) error {
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

	// Create emails table with all required fields
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
			sha256_hash TEXT NOT NULL,
			access_count INTEGER DEFAULT 0,
			has_attachments INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (sender_id) REFERENCES users(id)
		)
	`)
	return err
}
