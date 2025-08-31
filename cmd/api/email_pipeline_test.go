package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"secure-email-mvp/pkg/email"

	_ "modernc.org/sqlite"
)

// TestEmailSendingPipeline tests the complete email sending pipeline with security features
func TestEmailSendingPipeline(t *testing.T) {
	// Setup test database
	dbPath := fmt.Sprintf("/tmp/test-email-pipeline-%d.db", time.Now().Unix())
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer os.Remove(dbPath)
	defer db.Close()

	// Create test tables
	if err := createPipelineTestTables(db); err != nil {
		t.Fatalf("Failed to create test tables: %v", err)
	}

	// Create test user
	userID := "test-user-123"
	if err := createTestUser(db, userID); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Initialize server
	srv := &Server{
		db:                    db,
		emailSecurityService: email.NewEmailSecurityService(db),
	}

	// Test cases
	testCases := []struct {
		name           string
		request        SendEmailRequest
		expectedStatus int
		description    string
	}{
		{
			name: "Basic Email with Password Protection",
			request: SendEmailRequest{
				Recipient: "test@example.com",
				Subject:   "Test Email with Password",
				Body:      "This is a test email with password protection",
				Password:  "securepass123",
			},
			expectedStatus: http.StatusOK,
			description:    "Should successfully send email with password protection",
		},
		{
			name: "Email with All Security Features",
			request: SendEmailRequest{
				Recipient:                 "secure@example.com",
				Subject:                   "Test Email with All Security",
				Body:                      "This is a test email with all security features",
				Password:                  "securepass123",
				BurnAfterRead:             true,
				SelfDestructAfterAttempts: true,
				MaxFailedAttempts:         3,
				TimeLock:                  true,
				UnlockAfter:               time.Now().Add(time.Hour).Format(time.RFC3339),
				ExpiresAt:                 time.Now().Add(24 * time.Hour).Format(time.RFC3339),
				GeoVerificationType:       "country",
				GeoCountry:                "US",
				RequireMFA:                true,
				MFAType:                   "TOTP",
				RemoteRevoke:              true,
				DecoyMessage:              true,
				StripMetadata:             true,
				TamperAlerts:              true,
			},
			expectedStatus: http.StatusOK,
			description:    "Should successfully send email with all security features",
		},
		{
			name: "Invalid Email Format",
			request: SendEmailRequest{
				Recipient: "invalid-email",
				Subject:   "Test Invalid Email",
				Body:      "This should fail validation",
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Should reject invalid email format",
		},
		{
			name: "Missing Required Fields",
			request: SendEmailRequest{
				Recipient: "test@example.com",
				// Missing subject and body
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Should reject missing required fields",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create request body
			reqBody, err := json.Marshal(tc.request)
			if err != nil {
				t.Fatalf("Failed to marshal request: %v", err)
			}

			// Create HTTP request
			req := httptest.NewRequest("POST", "/api/email/send", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			// Add user context
			ctx := context.WithValue(req.Context(), "userID", userID)
			req = req.WithContext(ctx)

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			srv.sendEmailHandler(w, req)

			// Check status code
			if w.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, w.Code)
				t.Logf("Response body: %s", w.Body.String())
			}

			// If successful, verify response structure
			if tc.expectedStatus == http.StatusOK {
				var response SendEmailResponse
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				// Verify response fields
				if response.Status != "success" {
					t.Errorf("Expected status 'success', got '%s'", response.Status)
				}

				if response.BlobID == "" {
					t.Error("Expected blob_id to be present")
				}

				// Log success details
				t.Logf("✅ Email sent successfully")
				t.Logf("   Blob ID: %s", response.BlobID)
				t.Logf("   Status: %s", response.Status)
				if response.SecureLinkURL != nil {
					t.Logf("   Secure Link: %s", *response.SecureLinkURL)
				}
			}
		})
	}
}

// TestEmailSecurityService tests the EmailSecurityService directly
func TestEmailSecurityService(t *testing.T) {
	// Setup test database
	dbPath := fmt.Sprintf("/tmp/test-security-service-%d.db", time.Now().Unix())
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer os.Remove(dbPath)
	defer db.Close()

	// Create test tables
	if err := createPipelineTestTables(db); err != nil {
		t.Fatalf("Failed to create test tables: %v", err)
	}

	// Initialize service
	service := email.NewEmailSecurityService(db)

	// Test security configuration validation
	validConfig := email.SecurityFeatureConfig{
		PasswordProtection: true,
		Password:          "securepass123",
		BurnAfterRead:     true,
		RequireMFA:        true,
		MFAType:           "TOTP",
	}

	if err := service.ValidateSecurityConfig(validConfig); err != nil {
		t.Errorf("Valid config should pass validation: %v", err)
	}

	// Test invalid configuration
	invalidConfig := email.SecurityFeatureConfig{
		PasswordProtection: true,
		Password:          "", // Missing password
	}

	if err := service.ValidateSecurityConfig(invalidConfig); err == nil {
		t.Error("Invalid config should fail validation")
	}

	t.Logf("✅ EmailSecurityService validation tests passed")
}

// Helper functions

func createPipelineTestTables(db *sql.DB) error {
	// Create users table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			totp_secret TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	// Create emails table with all security fields
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS emails (
			email_id TEXT PRIMARY KEY,
			sender_id TEXT NOT NULL,
			recipient TEXT NOT NULL,
			subject TEXT NOT NULL,
			body TEXT NOT NULL,
			encrypted_blob_url TEXT,
			encrypted_key TEXT,
			encryption_nonce TEXT,
			encryption_auth_tag TEXT,
			sha256_hash TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			time_lock BOOLEAN DEFAULT FALSE,
			unlock_after TEXT,
			expires_at TEXT,
			burn_after_read BOOLEAN DEFAULT FALSE,
			self_destruct_after_attempts BOOLEAN DEFAULT FALSE,
			max_attempts INTEGER DEFAULT 3,
			require_mfa BOOLEAN DEFAULT FALSE,
			mfa_type TEXT DEFAULT 'TOTP',
			mfa_on_open BOOLEAN DEFAULT FALSE,
			mfa_on_reply BOOLEAN DEFAULT FALSE,
			mfa_on_forward BOOLEAN DEFAULT FALSE,
			remote_revoke BOOLEAN DEFAULT FALSE,
			decoy_message BOOLEAN DEFAULT FALSE,
			decoy_secret TEXT,
			strip_metadata BOOLEAN DEFAULT FALSE,
			tamper_alerts BOOLEAN DEFAULT FALSE,
			geolocation_json TEXT,
			is_password_protected BOOLEAN DEFAULT FALSE,
			password_hash TEXT,
			password_salt TEXT
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create emails table: %w", err)
	}

	// Create audit_log table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS audit_log (
			log_id TEXT PRIMARY KEY,
			timestamp DATETIME,
			event_type TEXT,
			user_id TEXT,
			ip_address TEXT,
			user_agent TEXT,
			related_email_id TEXT,
			outcome TEXT,
			details TEXT,
			severity TEXT,
			session_id TEXT,
			request_id TEXT,
			country TEXT,
			city TEXT,
			device_type TEXT,
			created_at DATETIME
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create audit_log table: %w", err)
	}

	return nil
}

func createTestUser(db *sql.DB, userID string) error {
	_, err := db.Exec(`
		INSERT OR REPLACE INTO users (id, email, password_hash, totp_secret)
		VALUES (?, ?, ?, ?)
	`, userID, "test@example.com", "hashed_password", "test_totp_secret")
	return err
}
