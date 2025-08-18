package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"secure-email-mvp/pkg/audit"

	_ "modernc.org/sqlite"
)

// TestSenderAccessInsights tests the sender-side access insights feature
func TestSenderAccessInsights(t *testing.T) {
	// Create test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Create required tables
	if err := createTestTablesForMicroIteration423(t, db); err != nil {
		t.Fatalf("Failed to create test tables: %v", err)
	}

	// Create server instance
	srv := &Server{
		db:                 db,
		emailAccessAuditor: audit.NewEmailAccessAuditor(db, audit.DefaultRateLimitConfig),
	}

	// Insert test data
	emailID := "test-email-123"
	userID := "test-user-456"

	// Insert email record
	_, err = db.Exec(`
		INSERT INTO emails (email_id, sender_id, recipient, subject, encrypted_blob_url, encrypted_key, encryption_nonce, encryption_auth_tag, compression_algo, sha256_hash, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, emailID, userID, "recipient@test.com", "Test Subject", "blob-url", "encrypted-key", "nonce", "auth-tag", "gzip", "hash", time.Now().Unix())
	if err != nil {
		t.Fatalf("Failed to insert test email: %v", err)
	}

	// Insert access log entries
	_, err = db.Exec(`
		INSERT INTO email_access_logs (email_id, user_id, ip_address, user_agent, status, attempt_count, result, created_at)
		VALUES 
			(?, ?, '192.168.1.100', 'Mozilla/5.0', 'success', 1, 'success', datetime('now', '-1 hour')),
			(?, ?, '10.0.0.50', 'Chrome/90.0', 'fail', 2, 'failed_password', datetime('now', '-30 minutes'))
	`, emailID, userID, emailID, "other-user")
	if err != nil {
		t.Fatalf("Failed to insert test access logs: %v", err)
	}

	// Create test request
	req := httptest.NewRequest("GET", "/api/email/detail/"+emailID, nil)
	req.Header.Set("Authorization", "Bearer test-token")

	// Mock JWT context
	ctx := context.WithValue(req.Context(), UserIDKey, userID)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	// Call the handler
	srv.getEmailDetailHandler(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response EmailDetailResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify access insights are present
	if response.AccessInsights == nil {
		t.Error("Expected access insights to be present")
		return
	}

	insights := *response.AccessInsights

	// Verify insights fields
	if insights["email_id"] != emailID {
		t.Errorf("Expected email_id to be %s, got %v", emailID, insights["email_id"])
	}

	if insights["total_access_count"] != 1 {
		t.Errorf("Expected total_access_count to be 1, got %v", insights["total_access_count"])
	}

	// Verify IP is anonymized
	lastAccessIP, ok := insights["last_access_ip"].(string)
	if !ok {
		t.Error("Expected last_access_ip to be a string")
	} else {
		// Should be anonymized
		if !strings.Contains(lastAccessIP, "/24") && !strings.Contains(lastAccessIP, "/64") {
			t.Errorf("Expected anonymized IP, got %s", lastAccessIP)
		}
	}
}

// TestAdminAccessLogsEndpoint tests the admin access log query endpoint
func TestAdminAccessLogsEndpoint(t *testing.T) {
	// Create test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Create required tables
	if err := createTestTablesForMicroIteration423(t, db); err != nil {
		t.Fatalf("Failed to create test tables: %v", err)
	}

	// Create server instance
	srv := &Server{
		db:                 db,
		emailAccessAuditor: audit.NewEmailAccessAuditor(db, audit.DefaultRateLimitConfig),
	}

	// Insert test access log data
	_, err = db.Exec(`
		INSERT INTO email_access_logs (email_id, user_id, ip_address, user_agent, status, attempt_count, result, created_at)
		VALUES 
			('email1', 'user1', '192.168.1.100', 'Mozilla/5.0', 'success', 1, 'success', datetime('now', '-1 hour')),
			('email1', 'user2', '10.0.0.50', 'Chrome/90.0', 'fail', 2, 'failed_password', datetime('now', '-30 minutes')),
			('email2', 'user1', '172.16.0.25', 'Safari/14.0', 'success', 1, 'success', datetime('now', '-15 minutes'))
	`)
	if err != nil {
		t.Fatalf("Failed to insert test access logs: %v", err)
	}

	// Test cases
	testCases := []struct {
		name           string
		queryParams    string
		expectedCount  int
		expectedStatus int
	}{
		{
			name:           "Get all logs",
			queryParams:    "",
			expectedCount:  3,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Filter by email_id",
			queryParams:    "?email_id=email1",
			expectedCount:  2,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Filter by user_id",
			queryParams:    "?user_id=user1",
			expectedCount:  2,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Filter by result",
			queryParams:    "?result=success",
			expectedCount:  2,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Pagination",
			queryParams:    "?limit=2&offset=0",
			expectedCount:  2,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create test request
			req := httptest.NewRequest("GET", "/api/admin/email/access-logs"+tc.queryParams, nil)
			req.Header.Set("Authorization", "Bearer test-token")

			// Mock JWT context
			ctx := context.WithValue(req.Context(), UserIDKey, "admin-user")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()

			// Call the handler
			srv.adminAccessLogsHandler(w, req)

			// Check response status
			if w.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, w.Code)
				return
			}

			if tc.expectedStatus != http.StatusOK {
				return
			}

			// Parse response
			var response AdminAccessLogsResponse
			if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			// Verify log count
			if len(response.Logs) != tc.expectedCount {
				t.Errorf("Expected %d logs, got %d", tc.expectedCount, len(response.Logs))
			}

			// Verify pagination info
			if response.TotalCount < tc.expectedCount {
				t.Errorf("Expected total count to be at least %d, got %d", tc.expectedCount, response.TotalCount)
			}
		})
	}
}

// TestAdminAccessLogsEndpointUnauthorized tests unauthorized access to admin endpoint
func TestAdminAccessLogsEndpointUnauthorized(t *testing.T) {
	// Create test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Create required tables
	if err := createTestTablesForMicroIteration423(t, db); err != nil {
		t.Fatalf("Failed to create test tables: %v", err)
	}

	// Create server instance
	srv := &Server{
		db:                 db,
		emailAccessAuditor: audit.NewEmailAccessAuditor(db, audit.DefaultRateLimitConfig),
	}

	// Test without authentication
	req := httptest.NewRequest("GET", "/api/admin/email/access-logs", nil)
	w := httptest.NewRecorder()

	srv.adminAccessLogsHandler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

// Helper function to create test tables for Micro-Iteration 4.23
func createTestTablesForMicroIteration423(t *testing.T, db *sql.DB) error {
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

	// Create emails table
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
			created_at INTEGER NOT NULL,
			FOREIGN KEY (sender_id) REFERENCES users(id)
		)
	`)
	if err != nil {
		return err
	}

	// Create email_access_logs table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS email_access_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email_id TEXT NOT NULL,
			user_id TEXT,
			ip_address TEXT NOT NULL,
			user_agent TEXT,
			status TEXT NOT NULL,
			attempt_count INTEGER DEFAULT 1,
			result TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (email_id) REFERENCES emails(email_id) ON DELETE CASCADE
		)
	`)
	return err
}




