package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"secure-email-mvp/pkg/email"

	"github.com/dgrijalva/jwt-go"
	_ "modernc.org/sqlite"
)

// Helper function to generate JWT token for email security tests
func generateEmailSecurityTestToken(t *testing.T, userID, userEmail string) string {
	// Set JWT secrets for testing
	os.Setenv("JWT_SECRET", "test-secret-key-for-jwt-signing")
	os.Setenv("JWT_ACCESS_SECRET", "test-access-secret-key-for-testing-only")

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"email":   userEmail,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	})
	tokenString, err := token.SignedString([]byte("test-secret-key-for-jwt-signing"))
	if err != nil {
		t.Fatalf("Failed to generate JWT token: %v", err)
	}
	return tokenString
}

// TestUpdateEmailSecurityHandler tests the POST /api/email/security/{id} endpoint
func TestUpdateEmailSecurityHandler(t *testing.T) {
	// Setup test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Create test tables
	if err := createTestTablesForEmailSecurity(db); err != nil {
		t.Fatalf("Failed to create test tables: %v", err)
	}

	// Create test user and email
	testUserID := "user-1"
	testUserEmail := "test@securesystem.email"
	emailID := "email-1"

	_, err = db.Exec(`INSERT INTO users (id, email, password_hash) VALUES (?, ?, ?)`,
		testUserID, testUserEmail, "hash1")
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO emails (email_id, sender_id, recipient, subject, encrypted_blob_url, 
			encrypted_key, encryption_nonce, encryption_auth_tag, compression_algo, sha256_hash)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		emailID, testUserID, "recipient@example.com", "Test Email",
		"blob-1", "key-1", "nonce-1", "tag-1", "gzip", "hash-1")
	if err != nil {
		t.Fatalf("Failed to insert test email: %v", err)
	}

	// Create server with test database
	srv := &Server{db: db}

	// Generate JWT token for test user
	token := generateTestToken(t, testUserID, testUserEmail)

	tests := []struct {
		name           string
		emailID        string
		requestBody    email.EmailSecurityRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name:    "Valid security toggles update",
			emailID: emailID,
			requestBody: email.EmailSecurityRequest{
				EmailID: emailID,
				Toggles: email.EmailSecurityToggles{
					ReadOnce:              true,
					MFAOnOpen:             true,
					RemoteRevoke:          false,
					StripMetadata:         true,
					SelfDestructThreshold: intPtr(5),
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "Invalid time window",
			emailID: emailID,
			requestBody: email.EmailSecurityRequest{
				EmailID: emailID,
				Toggles: email.EmailSecurityToggles{
					NotBefore: int64Ptr(time.Now().Add(time.Hour).Unix()),
					ExpiresAt: int64Ptr(time.Now().Unix()),
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "not_before must be before expires_at",
		},
		{
			name:    "Invalid self-destruct threshold",
			emailID: emailID,
			requestBody: email.EmailSecurityRequest{
				EmailID: emailID,
				Toggles: email.EmailSecurityToggles{
					SelfDestructThreshold: intPtr(0),
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "self_destruct_threshold must be at least 1",
		},
		{
			name:    "Invalid geo rules JSON",
			emailID: emailID,
			requestBody: email.EmailSecurityRequest{
				EmailID: emailID,
				Toggles: email.EmailSecurityToggles{
					GeoRulesRef: stringPtr(`{"type": "circle", "lat": 40.7128`),
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "geo_rules_ref must be valid JSON: unexpected end of JSON input",
		},
		{
			name:    "Email ID mismatch",
			emailID: emailID,
			requestBody: email.EmailSecurityRequest{
				EmailID: "different-email-id",
				Toggles: email.EmailSecurityToggles{},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Email ID mismatch",
		},
		{
			name:    "Non-existent email",
			emailID: "non-existent-email",
			requestBody: email.EmailSecurityRequest{
				EmailID: "non-existent-email",
				Toggles: email.EmailSecurityToggles{},
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Failed to update security settings",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request body
			requestBody, err := json.Marshal(tt.requestBody)
			if err != nil {
				t.Fatalf("Failed to marshal request body: %v", err)
			}

			// Create request
			req := httptest.NewRequest("POST", "/api/email/security/"+tt.emailID, bytes.NewBuffer(requestBody))
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Apply test middleware and call handler
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx := context.WithValue(r.Context(), UserIDKey, testUserID)
				ctx = context.WithValue(ctx, EmailKey, testUserEmail)
				r = r.WithContext(ctx)
				srv.updateEmailSecurityHandler(w, r)
			})
			handler.ServeHTTP(w, req)

			// Check response status
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
				t.Logf("Response body: %s", w.Body.String())
				return
			}

			// Check error message if expected
			if tt.expectedError != "" {
				var response email.EmailSecurityResponse
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				if response.Error != tt.expectedError {
					t.Errorf("Expected error '%s', got '%s'", tt.expectedError, response.Error)
				}
			}
		})
	}
}

// TestGetEmailSecurityHandler tests the GET /api/email/security/{id} endpoint
func TestGetEmailSecurityHandler(t *testing.T) {
	// Setup test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Create test tables
	if err := createTestTablesForEmailSecurity(db); err != nil {
		t.Fatalf("Failed to create test tables: %v", err)
	}

	// Create test user and email
	testUserID := "user-1"
	testUserEmail := "test@securesystem.email"
	emailID := "email-1"

	_, err = db.Exec(`INSERT INTO users (id, email, password_hash) VALUES (?, ?, ?)`,
		testUserID, testUserEmail, "hash1")
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	// Insert email with security toggles
	_, err = db.Exec(`
		INSERT INTO emails (email_id, sender_id, recipient, subject, encrypted_blob_url, 
			encrypted_key, encryption_nonce, encryption_auth_tag, sha256_hash,
			read_once, mfa_on_open, remote_revoke, strip_metadata, self_destruct_threshold)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		emailID, testUserID, "recipient@example.com", "Test Email",
		"blob-1", "key-1", "nonce-1", "tag-1", "hash-1",
		true, true, false, true, 5)
	if err != nil {
		t.Fatalf("Failed to insert test email: %v", err)
	}

	// Create server with test database
	srv := &Server{db: db}

	// Generate JWT token for test user
	token := generateTestToken(t, testUserID, testUserEmail)

	tests := []struct {
		name           string
		emailID        string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Valid security settings retrieval",
			emailID:        emailID,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Non-existent email",
			emailID:        "non-existent-email",
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Failed to retrieve security settings",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest("GET", "/api/email/security/"+tt.emailID, nil)
			req.Header.Set("Authorization", "Bearer "+token)

			// Create response recorder
			w := httptest.NewRecorder()

			// Apply test middleware and call handler
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx := context.WithValue(r.Context(), UserIDKey, testUserID)
				ctx = context.WithValue(ctx, EmailKey, testUserEmail)
				r = r.WithContext(ctx)
				srv.getEmailSecurityHandler(w, r)
			})
			handler.ServeHTTP(w, req)

			// Check response status
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
				t.Logf("Response body: %s", w.Body.String())
				return
			}

			// Check response for valid case
			if tt.expectedStatus == http.StatusOK {
				var response email.EmailSecurityInfo
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if response.EmailID != tt.emailID {
					t.Errorf("Expected email ID '%s', got '%s'", tt.emailID, response.EmailID)
				}

				// Verify security toggles
				if !response.Toggles.ReadOnce {
					t.Error("Expected read_once to be true")
				}
				if !response.Toggles.MFAOnOpen {
					t.Error("Expected mfa_on_open to be true")
				}
				if response.Toggles.RemoteRevoke {
					t.Error("Expected remote_revoke to be false")
				}
				if !response.Toggles.StripMetadata {
					t.Error("Expected strip_metadata to be true")
				}
				if response.Toggles.GetSelfDestructThreshold() != 5 {
					t.Errorf("Expected self_destruct_threshold to be 5, got %d", response.Toggles.GetSelfDestructThreshold())
				}
			}

			// Check error message if expected
			if tt.expectedError != "" {
				var response email.EmailSecurityResponse
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				if response.Error != tt.expectedError {
					t.Errorf("Expected error '%s', got '%s'", tt.expectedError, response.Error)
				}
			}
		})
	}
}

// TestEmailSecurityEnforcement tests that security toggles are enforced during email access
func TestEmailSecurityEnforcement(t *testing.T) {
	// Setup test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Create test tables
	if err := createTestTablesForEmailSecurity(db); err != nil {
		t.Fatalf("Failed to create test tables: %v", err)
	}

	// Create test user and email
	testUserID := "user-1"
	testUserEmail := "test@securesystem.email"
	emailID := "email-1"

	_, err = db.Exec(`INSERT INTO users (id, email, password_hash) VALUES (?, ?, ?)`,
		testUserID, testUserEmail, "hash1")
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	// Insert email with revoked security toggle
	_, err = db.Exec(`
		INSERT INTO emails (email_id, sender_id, recipient, subject, encrypted_blob_url, 
			encrypted_key, encryption_nonce, encryption_auth_tag, compression_algo, sha256_hash, remote_revoke)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		emailID, testUserID, "recipient@example.com", "Test Email",
		"blob-1", "key-1", "nonce-1", "tag-1", "gzip", "hash-1", true)
	if err != nil {
		t.Fatalf("Failed to insert test email: %v", err)
	}

	// Create server with test database
	srv := &Server{db: db}

	// Generate JWT token for test user
	token := generateTestToken(t, testUserID, testUserEmail)

	// Create request to access email
	requestBody := map[string]string{"email_id": emailID}
	bodyBytes, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/api/email/get", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	// Apply test middleware and call handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), UserIDKey, testUserID)
		ctx = context.WithValue(ctx, EmailKey, testUserEmail)
		r = r.WithContext(ctx)
		srv.getEmailHandler(w, r)
	})
	handler.ServeHTTP(w, req)

	// Check that access is denied due to remote revoke
	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
		t.Logf("Response body: %s", w.Body.String())
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["error"] != "Email has been revoked by sender" {
		t.Errorf("Expected error 'Email has been revoked by sender', got '%v'", response["error"])
	}
}

// Helper function to create test tables for email security tests
func createTestTablesForEmailSecurity(db *sql.DB) error {
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

	// Create emails table with all required columns for getEmailHandler
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
			-- New security toggle fields (Micro-Iteration 4.7)
			not_before INTEGER,
			expires_at INTEGER,
			read_once BOOLEAN DEFAULT FALSE,
			read_once_consumed_at DATETIME,
			read_once_consumer_device TEXT,
			self_destruct_on_read_once BOOLEAN DEFAULT FALSE,
			mfa_on_open BOOLEAN DEFAULT FALSE,
			decoy_secret TEXT,
			remote_revoke BOOLEAN DEFAULT FALSE,
			strip_metadata BOOLEAN DEFAULT FALSE,
			self_destruct_threshold INTEGER DEFAULT 3,
			geo_rules_ref TEXT,
			-- Required columns for getEmailHandler
			fail_count INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (sender_id) REFERENCES users(id)
		)
	`)
	return err
}

// Helper functions
func intPtr(v int) *int {
	return &v
}

func int64Ptr(v int64) *int64 {
	return &v
}

func stringPtr(v string) *string {
	return &v
}
