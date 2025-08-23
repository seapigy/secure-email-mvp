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

	"github.com/gorilla/mux"
	_ "modernc.org/sqlite"
)

// setupTestDB creates a test database with the required schema
func setupTestDB(t *testing.T) *sql.DB {
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
		geolocation_json TEXT,
		expires_at DATETIME,
		burn_after_read INTEGER DEFAULT 0,
		failed_attempts INTEGER DEFAULT 0,
		max_attempts INTEGER DEFAULT 3,
		self_destruct_after_attempts INTEGER DEFAULT 0,
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
		-- Simple geolocation fields
		allowed_city TEXT,
		allowed_country TEXT,
		-- Geofencing fields
		allowed_countries TEXT,
		allowed_ip_ranges TEXT,
		geofence_violations INTEGER DEFAULT 0,
		geofence_last_violation DATETIME,
		-- Enhanced geolocation verification fields
		geo_verification_type TEXT CHECK (geo_verification_type IN ('none', 'country', 'city', 'city_country')) DEFAULT 'none',
		geo_city TEXT,
		geo_country TEXT,
		-- MFA fields
		require_mfa INTEGER DEFAULT 0,
		mfa_type TEXT CHECK (mfa_type IN ('TOTP', 'EMAIL_CODE')),
		encrypted_totp_secret TEXT,
		mfa_failed_attempts INTEGER DEFAULT 0,
		mfa_locked_until DATETIME,
		-- Password protection fields
		is_password_protected BOOLEAN DEFAULT FALSE,
		password_hash TEXT,
		password_salt TEXT,
		-- Enhanced geo restriction fields
		geo_restriction_rules TEXT,
		geo_restriction_config TEXT,
		geo_restriction_enabled INTEGER DEFAULT 1,
		geo_restriction_violations INTEGER DEFAULT 0,
		geo_restriction_last_violation DATETIME,
		-- Brute force protection fields
		brute_force_failed_attempts INTEGER DEFAULT 0,
		brute_force_last_failed_attempt DATETIME,
		brute_force_lockout_until DATETIME,
		brute_force_max_attempts INTEGER DEFAULT 3,
		brute_force_lockout_duration_minutes INTEGER DEFAULT 15
	);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to apply schema: %v", err)
	}

	// Create IP access attempts table for IP tracking
	ipTrackingSchema := `
	CREATE TABLE IF NOT EXISTS ip_access_attempts (
		ip_address TEXT PRIMARY KEY,
		failed_attempts INTEGER DEFAULT 0,
		last_attempt_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		lockout_until TIMESTAMP NULL
	);
	`

	if _, err := db.Exec(ipTrackingSchema); err != nil {
		t.Fatalf("Failed to create ip_access_attempts table: %v", err)
	}

	return db
}

// createTestEmail creates a test email in the database
func createTestEmail(t *testing.T, db *sql.DB, emailID, senderID string, selfDestructAfterAttempts bool, maxAttempts int) {
	query := `
	INSERT INTO emails (
		email_id, sender_id, recipient, subject, encrypted_blob_url, encrypted_key,
		sha256_hash, self_destruct_after_attempts, max_attempts, encryption_nonce, encryption_auth_tag,
		allowed_city, allowed_country, geo_verification_type, geo_city, geo_country,
		require_mfa, mfa_type, encrypted_totp_secret, is_password_protected, geo_restriction_enabled,
		geo_restriction_rules, geo_restriction_config
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	selfDestructInt := 0
	if selfDestructAfterAttempts {
		selfDestructInt = 1
	}

	_, err := db.Exec(query,
		emailID, senderID, "test@example.com", "Test Subject",
		"test-blob-id", "test-key", "test-hash", selfDestructInt, maxAttempts,
		"test-nonce", "test-auth-tag",
		"", "", "none", "", "", // geolocation fields with defaults
		0, "TOTP", "", 0, 1, // MFA and password protection fields with defaults
		"", "", // geo restriction fields with defaults
	)
	if err != nil {
		t.Fatalf("Failed to create test email: %v", err)
	}
}

func TestHandleFailedAccessAttempt(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	srv := &Server{db: db}

	// Test case 1: Increment failed attempts without triggering self-destruct
	emailID := "test-email-1"
	createTestEmail(t, db, emailID, "test-sender", true, 3)

	// First failed attempt
	err := srv.handleFailedAccessAttempt(emailID, "test-blob", 0, 3)
	if err != nil {
		t.Errorf("Failed to handle first failed attempt: %v", err)
	}

	// Check that failed attempts were incremented
	var failedAttempts int
	err = db.QueryRow("SELECT failed_access_attempts FROM emails WHERE email_id = ?", emailID).Scan(&failedAttempts)
	if err != nil {
		t.Errorf("Failed to get failed attempts: %v", err)
	}
	if failedAttempts != 1 {
		t.Errorf("Expected failed attempts to be 1, got %d", failedAttempts)
	}

	// Test case 2: Trigger self-destruct on third attempt
	emailID2 := "test-email-2"
	createTestEmail(t, db, emailID2, "test-sender", true, 3)

	// First failed attempt
	err = srv.handleFailedAccessAttempt(emailID2, "test-blob-2", 0, 3)
	if err != nil {
		t.Errorf("Failed to handle first failed attempt: %v", err)
	}

	// Second failed attempt
	err = srv.handleFailedAccessAttempt(emailID2, "test-blob-2", 1, 3)
	if err != nil {
		t.Errorf("Failed to handle second failed attempt: %v", err)
	}

	// Third failed attempt (should trigger self-destruct)
	err = srv.handleFailedAccessAttempt(emailID2, "test-blob-2", 2, 3)
	if err != nil {
		t.Errorf("Failed to handle third failed attempt: %v", err)
	}

	// Check that email was marked as self-destructed
	var selfDestructed int
	err = db.QueryRow("SELECT self_destructed FROM emails WHERE email_id = ?", emailID2).Scan(&selfDestructed)
	if err != nil {
		t.Errorf("Failed to get self_destructed status: %v", err)
	}
	if selfDestructed != 1 {
		t.Errorf("Expected self_destructed to be 1, got %d", selfDestructed)
	}

	// Check that sensitive fields were cleared
	var encryptedBlobURL sql.NullString
	err = db.QueryRow("SELECT encrypted_blob_url FROM emails WHERE email_id = ?", emailID2).Scan(&encryptedBlobURL)
	if err != nil {
		t.Errorf("Failed to get encrypted_blob_url: %v", err)
	}
	if encryptedBlobURL.Valid {
		t.Errorf("Expected encrypted_blob_url to be NULL after self-destruct, got %s", encryptedBlobURL.String)
	}
}

func TestViewEmailHandlerSelfDestruct(t *testing.T) {
	// Set environment variable for testing
	os.Setenv("DEFAULT_MAX_FAILED_ATTEMPTS", "3")
	defer os.Unsetenv("DEFAULT_MAX_FAILED_ATTEMPTS")

	db := setupTestDB(t)
	defer db.Close()

	srv := &Server{db: db}

	// Create test email with self-destruct enabled
	emailID := "test-email-3"
	createTestEmail(t, db, emailID, "test-sender", true, 3)

	// Set up failed attempts to trigger self-destruct
	_, err := db.Exec("UPDATE emails SET failed_access_attempts = 3 WHERE email_id = ?", emailID)
	if err != nil {
		t.Fatalf("Failed to set failed attempts: %v", err)
	}

	// Create request
	req := httptest.NewRequest("GET", "/api/email/view/"+emailID, nil)
	req.Header.Set("Authorization", "Bearer test-token")

	// Create response recorder
	w := httptest.NewRecorder()

	// Set up router
	router := mux.NewRouter()
	router.HandleFunc("/api/email/view/{id}", srv.viewEmailHandler).Methods("GET")

	// Mock JWT middleware by setting user context
	ctx := context.WithValue(req.Context(), UserIDKey, "test-sender")
	req = req.WithContext(ctx)

	// Make request
	router.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusGone {
		t.Errorf("Expected status 410 Gone, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if response["status"] != "self_destructed" {
		t.Errorf("Expected status 'self_destructed', got %v", response["status"])
	}
}

func TestTestSelfDestructHandler(t *testing.T) {
	// Enable simulation
	os.Setenv("SIMULATE_SELF_DESTRUCT", "1")
	defer os.Unsetenv("SIMULATE_SELF_DESTRUCT")

	db := setupTestDB(t)
	defer db.Close()

	srv := &Server{db: db}

	// Create test email
	emailID := "test-email-4"
	createTestEmail(t, db, emailID, "test-sender", true, 3)

	// Test increment failed attempts
	reqBody := map[string]interface{}{
		"email_id": emailID,
		"action":   "increment_failed",
	}
	reqBodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/test/self-destruct", bytes.NewBuffer(reqBodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	// Set up router
	router := mux.NewRouter()
	router.HandleFunc("/test/self-destruct", srv.testSelfDestructHandler).Methods("POST")

	// Make request
	router.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 OK, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if response["failed_attempts"].(float64) != 1 {
		t.Errorf("Expected failed_attempts to be 1, got %v", response["failed_attempts"])
	}

	// Test reset
	reqBody = map[string]interface{}{
		"email_id": emailID,
		"action":   "reset",
	}
	reqBodyBytes, _ = json.Marshal(reqBody)

	req = httptest.NewRequest("POST", "/test/self-destruct", bytes.NewBuffer(reqBodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 OK, got %d", w.Code)
	}

	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if response["failed_attempts"].(float64) != 0 {
		t.Errorf("Expected failed_attempts to be 0, got %v", response["failed_attempts"])
	}
}
