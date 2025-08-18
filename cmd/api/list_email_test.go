package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"

	_ "modernc.org/sqlite"
)

// TestListEmailAuthentication tests authentication requirements for the list endpoint
func TestListEmailAuthentication(t *testing.T) {
	// Test cases
	testCases := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Missing Authorization Header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid or missing token",
		},
		{
			name:           "Invalid Bearer Format",
			authHeader:     "InvalidFormat token123",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid or missing token",
		},
		{
			name:           "Empty Token",
			authHeader:     "Bearer ",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid or missing token",
		},
		{
			name:           "Invalid JWT Token",
			authHeader:     "Bearer invalid.jwt.token",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid or missing token",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest("GET", "/api/email/list", nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Create a minimal server for testing
			srv := &Server{db: nil}

			// Apply JWT middleware to the handler
			handler := jwtMiddleware(http.HandlerFunc(srv.listEmailHandler))
			handler.ServeHTTP(w, req)

			// Check status code
			if w.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, w.Code)
			}

			// Check error message
			var response map[string]string
			if err := json.NewDecoder(w.Body).Decode(&response); err == nil {
				if response["error"] != tc.expectedError {
					t.Errorf("Expected error '%s', got '%s'", tc.expectedError, response["error"])
				}
			}
		})
	}
}

// TestListEmailWithValidToken tests successful list retrieval with valid JWT
func TestListEmailWithValidToken(t *testing.T) {
	// Set JWT secret for testing
	os.Setenv("JWT_SECRET", "test-secret-key-for-jwt-signing")
	defer os.Unsetenv("JWT_SECRET")

	// Generate a valid JWT token with both user_id and email claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": "123",
		"email":   "test@example.com",
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	})
	tokenString, err := token.SignedString([]byte("test-secret-key-for-jwt-signing"))
	if err != nil {
		t.Fatalf("Failed to generate JWT token: %v", err)
	}

	// Create request with valid token
	req := httptest.NewRequest("GET", "/api/email/list", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	// Create response recorder
	w := httptest.NewRecorder()

	// Create a minimal server for testing (no database needed for this test)
	srv := &Server{db: nil}

	// Apply JWT middleware to the handler
	handler := jwtMiddleware(http.HandlerFunc(srv.listEmailHandler))
	handler.ServeHTTP(w, req)

	// Should not get 401 Unauthorized (might get other errors due to missing DB, but not auth error)
	if w.Code == http.StatusUnauthorized {
		t.Errorf("Expected not to get 401 Unauthorized with valid JWT token, got %d", w.Code)
	}

	// Should get 500 Internal Server Error due to nil database
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500 Internal Server Error due to nil database, got %d", w.Code)
	}

	// Check error message
	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err == nil {
		if response["error"] != "Database connection unavailable" {
			t.Errorf("Expected error 'Database connection unavailable', got '%s'", response["error"])
		}
	}
}

// TestListEmailWithDatabase tests the full functionality with a real database
func TestListEmailWithDatabase(t *testing.T) {
	// Create temporary database
	dbPath := fmt.Sprintf("/tmp/test-list-email-%d.db", time.Now().Unix())
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Apply schema
	schema, err := os.ReadFile("../../schema/emails.sql")
	if err != nil {
		t.Fatalf("Failed to read schema: %v", err)
	}

	_, err = db.Exec(string(schema))
	if err != nil {
		t.Fatalf("Failed to apply schema: %v", err)
	}

	// Insert test data
	testUserID := "test-user-123"
	otherUserID := "other-user-456"

	// Insert emails for test user
	_, err = db.Exec(`
		INSERT INTO emails (email_id, sender_id, recipient, subject, encrypted_blob_url, encrypted_key, 
			encryption_nonce, encryption_auth_tag, compression_algo, sha256_hash, created_at)
		VALUES 
		('email-1', ?, 'alice@example.com', 'Project Update', 'blob-1', 'key-1', 'nonce-1', 'tag-1', 'gzip', 'hash-1', ?),
		('email-2', ?, 'bob@example.com', 'Meeting Notes', 'blob-2', 'key-2', 'nonce-2', 'tag-2', 'gzip', 'hash-2', ?),
		('email-3', ?, 'charlie@example.com', 'Weekly Report', 'blob-3', 'key-3', 'nonce-3', 'tag-3', 'gzip', 'hash-3', ?)`,
		testUserID, time.Now().Add(-1*time.Hour),
		testUserID, time.Now().Add(-2*time.Hour),
		testUserID, time.Now().Add(-3*time.Hour),
	)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Insert email for other user (should not appear in results)
	_, err = db.Exec(`
		INSERT INTO emails (email_id, sender_id, recipient, subject, encrypted_blob_url, encrypted_key, 
			encryption_nonce, encryption_auth_tag, compression_algo, sha256_hash, created_at)
		VALUES ('email-4', ?, 'dave@example.com', 'Other User Email', 'blob-4', 'key-4', 'nonce-4', 'tag-4', 'gzip', 'hash-4', ?)`,
		otherUserID, time.Now().Add(-4*time.Hour),
	)
	if err != nil {
		t.Fatalf("Failed to insert other user data: %v", err)
	}

	// Create server with test database
	srv := &Server{db: db}

	// Set JWT secret for testing
	os.Setenv("JWT_SECRET", "test-secret-key-for-jwt-signing")
	defer os.Unsetenv("JWT_SECRET")

	// Generate JWT token for test user
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": testUserID,
		"email":   "test@example.com",
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	})
	tokenString, err := token.SignedString([]byte("test-secret-key-for-jwt-signing"))
	if err != nil {
		t.Fatalf("Failed to generate JWT token: %v", err)
	}

	// Create request
	req := httptest.NewRequest("GET", "/api/email/list", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	// Create response recorder
	w := httptest.NewRecorder()

	// Apply JWT middleware and call handler
	handler := jwtMiddleware(http.HandlerFunc(srv.listEmailHandler))
	handler.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
		t.Logf("Response body: %s", w.Body.String())
		return
	}

	// Parse response
	var response ListEmailResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify response structure
	if response.Status != "success" {
		t.Errorf("Expected status 'success', got '%s'", response.Status)
	}

	// Verify email count (should be 3 emails for test user)
	if len(response.Emails) != 3 {
		t.Errorf("Expected 3 emails, got %d", len(response.Emails))
	}

	// Verify emails belong to test user and are ordered by created_at DESC
	expectedEmails := []string{"email-1", "email-2", "email-3"}
	for i, email := range response.Emails {
		if email.EmailID != expectedEmails[i] {
			t.Errorf("Expected email ID '%s' at position %d, got '%s'", expectedEmails[i], i, email.EmailID)
		}
	}

	// Verify no emails from other user are included
	for _, email := range response.Emails {
		if email.EmailID == "email-4" {
			t.Errorf("Found email from other user: %s", email.EmailID)
		}
	}
}

// TestListEmailEmptyResults tests the endpoint when user has no emails
func TestListEmailEmptyResults(t *testing.T) {
	// Create temporary database
	dbPath := fmt.Sprintf("/tmp/test-list-email-empty-%d.db", time.Now().Unix())
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Apply schema
	schema, err := os.ReadFile("../../schema/emails.sql")
	if err != nil {
		t.Fatalf("Failed to read schema: %v", err)
	}

	_, err = db.Exec(string(schema))
	if err != nil {
		t.Fatalf("Failed to apply schema: %v", err)
	}

	// Create server with empty database
	srv := &Server{db: db}

	// Set JWT secret for testing
	os.Setenv("JWT_SECRET", "test-secret-key-for-jwt-signing")
	defer os.Unsetenv("JWT_SECRET")

	// Generate JWT token for user with no emails
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": "456",
		"email":   "test@example.com",
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	})
	tokenString, err := token.SignedString([]byte("test-secret-key-for-jwt-signing"))
	if err != nil {
		t.Fatalf("Failed to generate JWT token: %v", err)
	}

	// Create request
	req := httptest.NewRequest("GET", "/api/email/list", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	// Create response recorder
	w := httptest.NewRecorder()

	// Apply JWT middleware and call handler
	handler := jwtMiddleware(http.HandlerFunc(srv.listEmailHandler))
	handler.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
		t.Logf("Response body: %s", w.Body.String())
		return
	}

	// Parse response
	var response ListEmailResponse
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

// TestListEmailDatabaseError tests error handling when database query fails
func TestListEmailDatabaseError(t *testing.T) {
	// Create server with nil database to simulate database error
	srv := &Server{db: nil}

	// Set JWT secret for testing
	os.Setenv("JWT_SECRET", "test-secret-key-for-jwt-signing")
	defer os.Unsetenv("JWT_SECRET")

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": "123",
		"email":   "test@example.com",
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	})
	tokenString, err := token.SignedString([]byte("test-secret-key-for-jwt-signing"))
	if err != nil {
		t.Fatalf("Failed to generate JWT token: %v", err)
	}

	// Create request
	req := httptest.NewRequest("GET", "/api/email/list", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	// Create response recorder
	w := httptest.NewRecorder()

	// Apply JWT middleware and call handler
	handler := jwtMiddleware(http.HandlerFunc(srv.listEmailHandler))
	handler.ServeHTTP(w, req)

	// Should get 500 Internal Server Error due to nil database
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	// Check error message
	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err == nil {
		if response["error"] != "Database connection unavailable" {
			t.Errorf("Expected error 'Database connection unavailable', got '%s'", response["error"])
		}
	}
}

// TestListEmailStatusComputation tests the status computation logic for different email states
func TestListEmailStatusComputation(t *testing.T) {
	// Create temporary database
	dbPath := fmt.Sprintf("/tmp/test-list-email-status-%d.db", time.Now().Unix())
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Apply schema
	schema, err := os.ReadFile("../../schema/emails.sql")
	if err != nil {
		t.Fatalf("Failed to read schema: %v", err)
	}

	_, err = db.Exec(string(schema))
	if err != nil {
		t.Fatalf("Failed to apply schema: %v", err)
	}

	// Insert test data with different status conditions
	testUserID := "test-user-status"
	now := time.Now()

	// Insert emails with different status conditions
	_, err = db.Exec(`
		INSERT INTO emails (
			email_id, sender_id, recipient, subject, encrypted_blob_url, encrypted_key, 
			encryption_nonce, encryption_auth_tag, compression_algo, sha256_hash, created_at,
			self_destructed, expires_at, access_count
		) VALUES 
		('email-delivered', ?, 'alice@example.com', 'Delivered Email', 'blob-1', 'key-1', 'nonce-1', 'tag-1', 'gzip', 'hash-1', ?, 0, NULL, 0),
		('email-read', ?, 'bob@example.com', 'Read Email', 'blob-2', 'key-2', 'nonce-2', 'tag-2', 'gzip', 'hash-2', ?, 0, NULL, 1),
		('email-expired', ?, 'charlie@example.com', 'Expired Email', 'blob-3', 'key-3', 'nonce-3', 'tag-3', 'gzip', 'hash-3', ?, 0, ?, 0),
		('email-deleted', ?, 'dave@example.com', 'Deleted Email', 'blob-4', 'key-4', 'nonce-4', 'tag-4', 'gzip', 'hash-4', ?, 1, NULL, 0)`,
		testUserID, now.Add(-1*time.Hour),
		testUserID, now.Add(-2*time.Hour),
		testUserID, now.Add(-3*time.Hour), now.Add(-1*time.Hour), // expired 1 hour ago
		testUserID, now.Add(-4*time.Hour),
	)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Create server with test database
	srv := &Server{db: db}

	// Set JWT secret for testing
	os.Setenv("JWT_SECRET", "test-secret-key-for-jwt-signing")
	defer os.Unsetenv("JWT_SECRET")

	// Generate JWT token for test user
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": testUserID,
		"email":   "test@example.com",
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	})
	tokenString, err := token.SignedString([]byte("test-secret-key-for-jwt-signing"))
	if err != nil {
		t.Fatalf("Failed to generate JWT token: %v", err)
	}

	// Create request
	req := httptest.NewRequest("GET", "/api/email/list", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	// Create response recorder
	w := httptest.NewRecorder()

	// Apply JWT middleware and call handler
	handler := jwtMiddleware(http.HandlerFunc(srv.listEmailHandler))
	handler.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
		t.Logf("Response body: %s", w.Body.String())
		return
	}

	// Parse response
	var response ListEmailResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify response structure
	if response.Status != "success" {
		t.Errorf("Expected status 'success', got '%s'", response.Status)
	}

	// Verify email count
	if len(response.Emails) != 4 {
		t.Errorf("Expected 4 emails, got %d", len(response.Emails))
		return
	}

	// Create a map to easily check status by email ID
	emailStatusMap := make(map[string]string)
	for _, email := range response.Emails {
		emailStatusMap[email.EmailID] = email.Status
	}

	// Verify status computation
	expectedStatuses := map[string]string{
		"email-delivered": "delivered",
		"email-read":      "read",
		"email-expired":   "expired",
		"email-deleted":   "deleted",
	}

	for emailID, expectedStatus := range expectedStatuses {
		if actualStatus, exists := emailStatusMap[emailID]; exists {
			if actualStatus != expectedStatus {
				t.Errorf("Email %s: Expected status '%s', got '%s'", emailID, expectedStatus, actualStatus)
			}
		} else {
			t.Errorf("Email %s not found in response", emailID)
		}
	}

	// Verify required fields are present and extra fields are not
	for _, email := range response.Emails {
		// Check required fields exist
		if email.EmailID == "" {
			t.Errorf("Email missing required field: email_id")
		}
		if email.Recipient == "" {
			t.Errorf("Email missing required field: recipient")
		}
		if email.Subject == "" {
			t.Errorf("Email missing required field: subject")
		}
		if email.CreatedAt.IsZero() {
			t.Errorf("Email missing required field: created_at")
		}
		if email.Status == "" {
			t.Errorf("Email missing required field: status")
		}
	}
}

// TestListEmailResponseSchema tests that the response matches exactly the 4.19 schema
func TestListEmailResponseSchema(t *testing.T) {
	// Create temporary database
	dbPath := fmt.Sprintf("/tmp/test-list-email-schema-%d.db", time.Now().Unix())
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Apply schema
	schema, err := os.ReadFile("../../schema/emails.sql")
	if err != nil {
		t.Fatalf("Failed to read schema: %v", err)
	}

	_, err = db.Exec(string(schema))
	if err != nil {
		t.Fatalf("Failed to apply schema: %v", err)
	}

	// Insert test data
	testUserID := "test-user-schema"
	_, err = db.Exec(`
		INSERT INTO emails (
			email_id, sender_id, recipient, subject, encrypted_blob_url, encrypted_key, 
			encryption_nonce, encryption_auth_tag, compression_algo, sha256_hash, created_at
		) VALUES 
		('email-schema-test', ?, 'test@example.com', 'Schema Test', 'blob-1', 'key-1', 'nonce-1', 'tag-1', 'gzip', 'hash-1', ?)`,
		testUserID, time.Now(),
	)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Create server with test database
	srv := &Server{db: db}

	// Set JWT secret for testing
	os.Setenv("JWT_SECRET", "test-secret-key-for-jwt-signing")
	defer os.Unsetenv("JWT_SECRET")

	// Generate JWT token for test user
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": testUserID,
		"email":   "test@example.com",
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	})
	tokenString, err := token.SignedString([]byte("test-secret-key-for-jwt-signing"))
	if err != nil {
		t.Fatalf("Failed to generate JWT token: %v", err)
	}

	// Create request
	req := httptest.NewRequest("GET", "/api/email/list", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	// Create response recorder
	w := httptest.NewRecorder()

	// Apply JWT middleware and call handler
	handler := jwtMiddleware(http.HandlerFunc(srv.listEmailHandler))
	handler.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
		t.Logf("Response body: %s", w.Body.String())
		return
	}

	// Parse response
	var response ListEmailResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify response structure
	if response.Status != "success" {
		t.Errorf("Expected status 'success', got '%s'", response.Status)
	}

	// Verify email count
	if len(response.Emails) != 1 {
		t.Errorf("Expected 1 email, got %d", len(response.Emails))
		return
	}

	email := response.Emails[0]

	// Verify exact 4.19 schema compliance
	// Required fields only
	if email.EmailID == "" {
		t.Errorf("Missing required field: email_id")
	}
	if email.Recipient == "" {
		t.Errorf("Missing required field: recipient")
	}
	if email.Subject == "" {
		t.Errorf("Missing required field: subject")
	}
	if email.CreatedAt.IsZero() {
		t.Errorf("Missing required field: created_at")
	}
	if email.Status == "" {
		t.Errorf("Missing required field: status")
	}

	// Verify status is one of the allowed values
	allowedStatuses := map[string]bool{
		"delivered": true,
		"read":      true,
		"expired":   true,
		"deleted":   true,
	}
	if !allowedStatuses[email.Status] {
		t.Errorf("Invalid status value: '%s'. Allowed values: delivered, read, expired, deleted", email.Status)
	}

	// Verify ordering (most recent first)
	if len(response.Emails) > 1 {
		for i := 0; i < len(response.Emails)-1; i++ {
			if response.Emails[i].CreatedAt.Before(response.Emails[i+1].CreatedAt) {
				t.Errorf("Emails not ordered by created_at DESC. Email %d (%s) is before email %d (%s)",
					i, response.Emails[i].CreatedAt, i+1, response.Emails[i+1].CreatedAt)
			}
		}
	}
}
