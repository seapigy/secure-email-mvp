package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
)

// TestInboxSecurity tests Zero Visibility compliance and security features
func TestInboxSecurity(t *testing.T) {
	// Setup test database
	db, err := setupTestDatabaseForSecurity()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer db.Close()

	// Create test server
	srv := &Server{db: db}

	// Create test users
	user1ID := "user-1-uuid"
	user2ID := "user-2-uuid"
	user1Email := "user1@test.com"
	user2Email := "user2@test.com"

	// Insert test users
	_, err = db.Exec("INSERT INTO users (id, email, password_hash, totp_secret) VALUES (?, ?, ?, ?)",
		user1ID, user1Email, "hashed_password", "totp_secret_1")
	if err != nil {
		t.Fatalf("Failed to insert test user 1: %v", err)
	}

	_, err = db.Exec("INSERT INTO users (id, email, password_hash, totp_secret) VALUES (?, ?, ?, ?)",
		user2ID, user2Email, "hashed_password", "totp_secret_2")
	if err != nil {
		t.Fatalf("Failed to insert test user 2: %v", err)
	}

	// Insert test emails
	email1ID := "email-1-uuid"
	email2ID := "email-2-uuid"
	_, err = db.Exec("INSERT INTO emails (email_id, sender_id, recipient, subject, encrypted_blob_url, encrypted_key, encryption_nonce, encryption_auth_tag, sha256_hash, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		email1ID, "sender-1", user1Email, "Test Email 1", "https://test.blob.url/1", "encrypted-key-1", "nonce-1", "auth-tag-1", "sha256-hash-1", time.Now())
	if err != nil {
		t.Fatalf("Failed to insert test email 1: %v", err)
	}

	_, err = db.Exec("INSERT INTO emails (email_id, sender_id, recipient, subject, encrypted_blob_url, encrypted_key, encryption_nonce, encryption_auth_tag, sha256_hash, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		email2ID, "sender-2", user2Email, "Test Email 2", "https://test.blob.url/2", "encrypted-key-2", "nonce-2", "auth-tag-2", "sha256-hash-2", time.Now())
	if err != nil {
		t.Fatalf("Failed to insert test email 2: %v", err)
	}

	// Insert inbox messages
	_, err = db.Exec("INSERT INTO inbox_messages (id, user_id, email_id, is_read, is_deleted, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		"inbox-1-uuid", user1ID, email1ID, false, false, time.Now(), time.Now())
	if err != nil {
		t.Fatalf("Failed to insert inbox message 1: %v", err)
	}

	_, err = db.Exec("INSERT INTO inbox_messages (id, user_id, email_id, is_read, is_deleted, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		"inbox-2-uuid", user2ID, email2ID, false, false, time.Now(), time.Now())
	if err != nil {
		t.Fatalf("Failed to insert inbox message 2: %v", err)
	}

	t.Run("TestUserIsolation", func(t *testing.T) {
		testUserIsolation(t, srv, user1ID, user2ID, email1ID, email2ID)
	})

	t.Run("TestZeroVisibilityLogging", func(t *testing.T) {
		testZeroVisibilityLogging(t, srv, user1ID, email1ID)
	})

	t.Run("TestGenericErrorMessages", func(t *testing.T) {
		testGenericErrorMessages(t, srv, user1ID)
	})

	t.Run("TestAuthenticationEnforcement", func(t *testing.T) {
		testAuthenticationEnforcement(t, srv)
	})
}

// testUserIsolation verifies that users cannot access each other's inbox
func testUserIsolation(t *testing.T, srv *Server, user1ID, user2ID, email1ID, email2ID string) {
	// Test 1: User 1 cannot access User 2's email
	req := httptest.NewRequest("GET", "/api/inbox/"+email2ID, nil)
	req = mux.SetURLVars(req, map[string]string{"id": email2ID})
	req = setUserContext(req, user1ID)

	w := httptest.NewRecorder()
	srv.getInboxEmailHandler(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}

	// Test 2: User 2 cannot access User 1's email
	req = httptest.NewRequest("GET", "/api/inbox/"+email1ID, nil)
	req = mux.SetURLVars(req, map[string]string{"id": email1ID})
	req = setUserContext(req, user2ID)

	w = httptest.NewRecorder()
	srv.getInboxEmailHandler(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}

	// Test 3: User 1 cannot delete User 2's email
	req = httptest.NewRequest("DELETE", "/api/inbox/"+email2ID, nil)
	req = mux.SetURLVars(req, map[string]string{"id": email2ID})
	req = setUserContext(req, user1ID)

	w = httptest.NewRecorder()
	srv.deleteInboxEmailHandler(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}

	// Test 4: User 2 cannot delete User 1's email
	req = httptest.NewRequest("DELETE", "/api/inbox/"+email1ID, nil)
	req = mux.SetURLVars(req, map[string]string{"id": email1ID})
	req = setUserContext(req, user2ID)

	w = httptest.NewRecorder()
	srv.deleteInboxEmailHandler(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

// testZeroVisibilityLogging verifies that no PII appears in logs
func testZeroVisibilityLogging(t *testing.T, srv *Server, userID, emailID string) {
	// Capture log output
	var logBuffer bytes.Buffer
	log.SetOutput(&logBuffer)
	defer log.SetOutput(os.Stderr)

	// Make a request that should generate logs
	req := httptest.NewRequest("GET", "/api/inbox/"+emailID, nil)
	req = mux.SetURLVars(req, map[string]string{"id": emailID})
	req = setUserContext(req, userID)

	w := httptest.NewRecorder()
	srv.getInboxEmailHandler(w, req)

	// Get log output
	logOutput := logBuffer.String()

	// Check for PII patterns that should NOT appear in logs
	piiPatterns := []string{
		"user1@test.com",
		"user2@test.com",
		"@test.com",
		"@",
		"password",
		"totp_secret",
		"hashed_password",
	}

	for _, pattern := range piiPatterns {
		if strings.Contains(logOutput, pattern) {
			t.Errorf("PII pattern '%s' found in logs, violating Zero Visibility rule", pattern)
		}
	}

	// Check that structured logging is working
	if !strings.Contains(logOutput, "[STRUCTURED]") {
		t.Error("Structured logging not found in output")
	}

	// Check that user_id is properly sanitized (should be UUID or "anonymous")
	if strings.Contains(logOutput, "user_id") {
		// Should contain UUID format or "anonymous", not email addresses
		if strings.Contains(logOutput, "user1@test.com") || strings.Contains(logOutput, "user2@test.com") {
			t.Error("Email addresses found in user_id field, violating Zero Visibility rule")
		}
	}
}

// testGenericErrorMessages verifies that error messages are generic and safe
func testGenericErrorMessages(t *testing.T, srv *Server, userID string) {
	// Test 1: Invalid email ID should return generic error
	req := httptest.NewRequest("GET", "/api/inbox/invalid-email-id", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "invalid-email-id"})
	req = setUserContext(req, userID)

	w := httptest.NewRecorder()
	srv.getInboxEmailHandler(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	errorMsg, ok := response["error"].(string)
	if !ok {
		t.Fatal("Error message not found in response")
	}

	// Check that error message is generic
	genericErrors := []string{
		"Email not found",
		"Authentication required",
		"Service temporarily unavailable",
		"Missing email ID",
	}

	isGeneric := false
	for _, genericError := range genericErrors {
		if errorMsg == genericError {
			isGeneric = true
			break
		}
	}

	if !isGeneric {
		t.Errorf("Error message '%s' is not generic and safe", errorMsg)
	}

	// Test 2: Missing email ID should return generic error
	req = httptest.NewRequest("GET", "/api/inbox/", nil)
	req = mux.SetURLVars(req, map[string]string{"id": ""})
	req = setUserContext(req, userID)

	w = httptest.NewRecorder()
	srv.getInboxEmailHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	errorMsg, ok = response["error"].(string)
	if !ok {
		t.Fatal("Error message not found in response")
	}

	if errorMsg != "Missing email ID" {
		t.Errorf("Expected generic error message, got '%s'", errorMsg)
	}
}

// testAuthenticationEnforcement verifies that all endpoints require authentication
func testAuthenticationEnforcement(t *testing.T, srv *Server) {
	endpoints := []struct {
		method  string
		path    string
		urlVars map[string]string
	}{
		{"GET", "/api/inbox/list", nil},
		{"GET", "/api/inbox/test-email-id", map[string]string{"id": "test-email-id"}},
		{"DELETE", "/api/inbox/test-email-id", map[string]string{"id": "test-email-id"}},
	}

	for _, endpoint := range endpoints {
		t.Run(fmt.Sprintf("TestAuthRequired_%s_%s", endpoint.method, endpoint.path), func(t *testing.T) {
			req := httptest.NewRequest(endpoint.method, endpoint.path, nil)
			if endpoint.urlVars != nil {
				req = mux.SetURLVars(req, endpoint.urlVars)
			}

			w := httptest.NewRecorder()

			switch endpoint.path {
			case "/api/inbox/list":
				srv.listInboxHandler(w, req)
			case "/api/inbox/test-email-id":
				if endpoint.method == "GET" {
					srv.getInboxEmailHandler(w, req)
				} else {
					srv.deleteInboxEmailHandler(w, req)
				}
			}

			if w.Code != http.StatusUnauthorized {
				t.Errorf("Expected status 401 for %s %s, got %d", endpoint.method, endpoint.path, w.Code)
			}

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}

			errorMsg, ok := response["error"].(string)
			if !ok {
				t.Fatal("Error message not found in response")
			}

			if errorMsg != "Authentication required" {
				t.Errorf("Expected 'Authentication required' error, got '%s'", errorMsg)
			}
		})
	}
}

// setupTestDatabaseForSecurity creates an in-memory test database for security tests
func setupTestDatabaseForSecurity() (*sql.DB, error) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, err
	}

	// Read and apply schema files
	schemaFiles := []string{
		"../../schema/users.sql",
		"../../schema/emails.sql",
		"../../schema/migrate_add_inbox_indexes.sql",
		"../../schema/migrate_add_inbox_messages_table.sql",
	}

	for _, schemaFile := range schemaFiles {
		schema, err := os.ReadFile(schemaFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read schema file %s: %v", schemaFile, err)
		}

		_, err = db.Exec(string(schema))
		if err != nil {
			return nil, fmt.Errorf("failed to apply schema %s: %v", schemaFile, err)
		}
	}

	return db, nil
}

// setUserContext adds user ID to request context for testing
func setUserContext(req *http.Request, userID string) *http.Request {
	ctx := req.Context()
	ctx = context.WithValue(ctx, UserIDKey, userID)
	return req.WithContext(ctx)
}
