package main

import (
	"database/sql"
	"testing"

	"secure-email-mvp/pkg/geofencing"
	"secure-email-mvp/pkg/geolocation"

	_ "modernc.org/sqlite"
)

// TestGeofencingIntegration tests the complete geofencing integration
func TestGeofencingIntegration(t *testing.T) {
	// Set up test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Apply schema
	schema := `
	CREATE TABLE users (
		id INTEGER PRIMARY KEY,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		totp_secret TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE emails (
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
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		allowed_countries TEXT,
		allowed_ip_ranges TEXT,
		geofence_violations INTEGER DEFAULT 0,
		geofence_last_violation DATETIME
	);

	INSERT INTO users (id, email, password_hash) VALUES (1, 'sender@test.com', 'hash');
	INSERT INTO users (id, email, password_hash) VALUES (2, 'recipient@test.com', 'hash');
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to apply schema: %v", err)
	}

	// Create test email with geofencing restrictions
	emailID := "test-geofence-email"
	allowedCountries := `["US", "CA"]`
	allowedIPRanges := `["192.168.1.0/24"]`

	_, err = db.Exec(`
		INSERT INTO emails (
			email_id, sender_id, recipient, subject, 
			encrypted_blob_url, encrypted_key, encryption_nonce, encryption_auth_tag, sha256_hash,
			allowed_countries, allowed_ip_ranges
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		emailID, "1", "recipient@test.com", "Test Subject",
		"test-blob", "test-key", "test-nonce", "test-auth-tag", "test-hash",
		allowedCountries, allowedIPRanges,
	)
	if err != nil {
		t.Fatalf("Failed to insert test email: %v", err)
	}

	// Initialize services
	geolocationSvc := geolocation.NewMockGeolocationService()
	geofencingSvc := geofencing.NewGeofencingService(db, geolocationSvc)

	// Set up test locations
	geolocationSvc.SetLocation("192.168.1.100", &geolocation.Location{
		Country: "US",
		City:    "New York",
		IP:      "192.168.1.100",
	})

	geolocationSvc.SetLocation("10.0.0.1", &geolocation.Location{
		Country: "CA",
		City:    "Toronto",
		IP:      "10.0.0.1",
	})

	geolocationSvc.SetLocation("203.0.113.1", &geolocation.Location{
		Country: "AU",
		City:    "Sydney",
		IP:      "203.0.113.1",
	})

	// Test cases
	testCases := []struct {
		name            string
		clientIP        string
		expectedAllowed bool
		description     string
	}{
		{
			name:            "Allowed US IP in range",
			clientIP:        "192.168.1.100",
			expectedAllowed: true,
			description:     "US IP within allowed CIDR range should be allowed",
		},
		{
			name:            "Allowed CA IP",
			clientIP:        "10.0.0.1",
			expectedAllowed: true,
			description:     "CA IP should be allowed based on country restriction",
		},
		{
			name:            "Blocked AU IP",
			clientIP:        "203.0.113.1",
			expectedAllowed: false,
			description:     "AU IP should be blocked (not in allowed countries or IP ranges)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := geofencingSvc.CheckGeofenceAccess(emailID, tc.clientIP)
			if err != nil {
				t.Fatalf("Geofencing check failed: %v", err)
			}

			if result.Allowed != tc.expectedAllowed {
				t.Errorf("Expected allowed=%v, got allowed=%v for %s",
					tc.expectedAllowed, result.Allowed, tc.description)
			}

			// Check violation counter for blocked attempts
			if !tc.expectedAllowed {
				violations, err := geofencingSvc.GetGeofenceViolations(emailID)
				if err != nil {
					t.Fatalf("Failed to get geofence violations: %v", err)
				}
				if violations == 0 {
					t.Error("Expected violation counter to be incremented for blocked access")
				}
			}
		})
	}
}

// TestGeofencingWithServer tests geofencing with a simplified server setup
func TestGeofencingWithServer(t *testing.T) {
	// Set up test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Apply schema
	schema := `
	CREATE TABLE emails (
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
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		allowed_countries TEXT,
		allowed_ip_ranges TEXT,
		geofence_violations INTEGER DEFAULT 0,
		geofence_last_violation DATETIME
	);

	INSERT INTO emails (
		email_id, sender_id, recipient, subject, 
		encrypted_blob_url, encrypted_key, encryption_nonce, encryption_auth_tag, sha256_hash,
		allowed_countries
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	emailID := "test-server-email"
	allowedCountries := `["US"]`

	_, err = db.Exec(schema, emailID, "1", "recipient@test.com", "Test Subject",
		"test-blob", "test-key", "test-nonce", "test-auth-tag", "test-hash",
		allowedCountries)
	if err != nil {
		t.Fatalf("Failed to apply schema: %v", err)
	}

	// Initialize services
	geolocationSvc := geolocation.NewMockGeolocationService()
	geofencingSvc := geofencing.NewGeofencingService(db, geolocationSvc)

	// Set up test location (blocked country)
	geolocationSvc.SetLocation("203.0.113.1", &geolocation.Location{
		Country: "AU",
		City:    "Sydney",
		IP:      "203.0.113.1",
	})

	// Test geofencing service directly
	result, err := geofencingSvc.CheckGeofenceAccess(emailID, "203.0.113.1")
	if err != nil {
		t.Fatalf("Geofencing check failed: %v", err)
	}

	if result.Allowed {
		t.Error("Expected geofencing to block AU IP when only US is allowed")
	}

	// Verify violation counter was incremented
	violations, err := geofencingSvc.GetGeofenceViolations(emailID)
	if err != nil {
		t.Fatalf("Failed to get geofence violations: %v", err)
	}
	if violations == 0 {
		t.Error("Expected violation counter to be incremented for blocked access")
	}
}

// TestGeofencingViolationCounter tests that geofence violations are tracked
func TestGeofencingViolationCounter(t *testing.T) {
	// Set up test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Apply schema
	schema := `
	CREATE TABLE emails (
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
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		allowed_countries TEXT,
		allowed_ip_ranges TEXT,
		geofence_violations INTEGER DEFAULT 0,
		geofence_last_violation DATETIME
	);

	INSERT INTO emails (
		email_id, sender_id, recipient, subject, 
		encrypted_blob_url, encrypted_key, encryption_nonce, encryption_auth_tag, sha256_hash,
		allowed_countries
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	emailID := "test-violation-email"
	allowedCountries := `["US"]`

	_, err = db.Exec(schema, emailID, "1", "recipient@test.com", "Test Subject",
		"test-blob", "test-key", "test-nonce", "test-auth-tag", "test-hash",
		allowedCountries)
	if err != nil {
		t.Fatalf("Failed to apply schema: %v", err)
	}

	// Initialize services
	geolocationSvc := geolocation.NewMockGeolocationService()
	geofencingSvc := geofencing.NewGeofencingService(db, geolocationSvc)

	// Set up blocked location
	geolocationSvc.SetLocation("203.0.113.1", &geolocation.Location{
		Country: "AU",
		City:    "Sydney",
		IP:      "203.0.113.1",
	})

	// Test multiple violations
	for i := 0; i < 3; i++ {
		result, err := geofencingSvc.CheckGeofenceAccess(emailID, "203.0.113.1")
		if err != nil {
			t.Fatalf("Geofencing check failed: %v", err)
		}

		if result.Allowed {
			t.Error("Expected geofencing to block AU IP")
		}
	}

	// Verify violation counter
	violations, err := geofencingSvc.GetGeofenceViolations(emailID)
	if err != nil {
		t.Fatalf("Failed to get geofence violations: %v", err)
	}

	if violations != 3 {
		t.Errorf("Expected 3 violations, got %d", violations)
	}

	// Test reset functionality
	err = geofencingSvc.ResetGeofenceViolations(emailID)
	if err != nil {
		t.Fatalf("Failed to reset geofence violations: %v", err)
	}

	violations, err = geofencingSvc.GetGeofenceViolations(emailID)
	if err != nil {
		t.Fatalf("Failed to get geofence violations after reset: %v", err)
	}

	if violations != 0 {
		t.Errorf("Expected 0 violations after reset, got %d", violations)
	}
}

// TestGeofencingSettings tests geofencing settings management
func TestGeofencingSettings(t *testing.T) {
	// Set up test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Apply schema
	schema := `
	CREATE TABLE emails (
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
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		allowed_countries TEXT,
		allowed_ip_ranges TEXT,
		geofence_violations INTEGER DEFAULT 0,
		geofence_last_violation DATETIME
	);

	INSERT INTO emails (
		email_id, sender_id, recipient, subject, 
		encrypted_blob_url, encrypted_key, encryption_nonce, encryption_auth_tag, sha256_hash
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	emailID := "test-settings-email"

	_, err = db.Exec(schema, emailID, "1", "recipient@test.com", "Test Subject",
		"test-blob", "test-key", "test-nonce", "test-auth-tag", "test-hash")
	if err != nil {
		t.Fatalf("Failed to apply schema: %v", err)
	}

	// Initialize services
	geolocationSvc := geolocation.NewMockGeolocationService()
	geofencingSvc := geofencing.NewGeofencingService(db, geolocationSvc)

	// Test setting geofencing restrictions
	allowedCountries := []string{"US", "CA"}
	allowedIPRanges := []string{"192.168.1.0/24"}

	err = geofencingSvc.SetGeofencingSettings(emailID, allowedCountries, allowedIPRanges)
	if err != nil {
		t.Fatalf("Failed to set geofencing settings: %v", err)
	}

	// Test retrieving geofencing settings
	retrievedCountries, retrievedRanges, err := geofencingSvc.GetGeofencingSettings(emailID)
	if err != nil {
		t.Fatalf("Failed to get geofencing settings: %v", err)
	}

	if len(retrievedCountries) != len(allowedCountries) {
		t.Errorf("Expected %d countries, got %d", len(allowedCountries), len(retrievedCountries))
	}

	if len(retrievedRanges) != len(allowedIPRanges) {
		t.Errorf("Expected %d IP ranges, got %d", len(allowedIPRanges), len(retrievedRanges))
	}

	// Test validation
	err = geofencingSvc.ValidateGeofencingSettings([]string{"US", "INVALID"}, []string{"192.168.1.0/24"})
	if err == nil {
		t.Error("Expected validation error for invalid country code")
	}

	err = geofencingSvc.ValidateGeofencingSettings([]string{"US"}, []string{"invalid-cidr"})
	if err == nil {
		t.Error("Expected validation error for invalid CIDR range")
	}

	// Test description formatting
	description := geofencingSvc.FormatGeofencingDescription(allowedCountries, allowedIPRanges)
	expectedDescription := "Countries: US, CA; IP Ranges: 192.168.1.0/24"
	if description != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, description)
	}
}
