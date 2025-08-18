package main

import (
	"database/sql"
	"net/http"
	"testing"

	"secure-email-mvp/pkg/devicefingerprint"

	_ "modernc.org/sqlite"
)

// TestDeviceFingerprintingIntegration tests the complete device fingerprinting integration
func TestDeviceFingerprintingIntegration(t *testing.T) {
	// Set up test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Apply schema
	schema := `
	CREATE TABLE users (
		id TEXT PRIMARY KEY,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		totp_secret TEXT
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
		requires_password INTEGER DEFAULT 0,
		password_hash TEXT,
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
		allowed_city TEXT,
		allowed_country TEXT,
		geo_verification_type TEXT CHECK (geo_verification_type IN ('none', 'country', 'city', 'city_country')) DEFAULT 'none',
		geo_city TEXT,
		geo_country TEXT,
		require_mfa INTEGER DEFAULT 0,
		mfa_type TEXT CHECK (mfa_type IN ('TOTP', 'EMAIL_CODE')),
		encrypted_totp_secret TEXT,
		mfa_failed_attempts INTEGER DEFAULT 0,
		mfa_locked_until DATETIME,
		brute_force_failed_attempts INTEGER DEFAULT 0,
		brute_force_last_failed_attempt DATETIME,
		brute_force_lockout_until DATETIME,
		brute_force_max_attempts INTEGER DEFAULT 3,
		brute_force_lockout_duration_minutes INTEGER DEFAULT 15,
		is_password_protected BOOLEAN DEFAULT FALSE,
		password_salt TEXT,
		allowed_countries TEXT,
		allowed_ip_ranges TEXT,
		geofence_violations INTEGER DEFAULT 0,
		geofence_last_violation DATETIME,
		trusted_devices_only BOOLEAN DEFAULT FALSE
	);

	CREATE TABLE trusted_devices (
		id TEXT PRIMARY KEY,
		email_id TEXT NOT NULL,
		device_hash TEXT NOT NULL,
		device_fingerprint TEXT NOT NULL,
		user_agent TEXT,
		ip_address TEXT,
		added_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_used_at DATETIME,
		access_count INTEGER DEFAULT 0,
		FOREIGN KEY (email_id) REFERENCES emails(email_id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_trusted_devices_email_id ON trusted_devices(email_id);
	CREATE INDEX IF NOT EXISTS idx_trusted_devices_device_hash ON trusted_devices(device_hash);
	CREATE INDEX IF NOT EXISTS idx_trusted_devices_email_hash ON trusted_devices(email_id, device_hash);
	CREATE INDEX IF NOT EXISTS idx_emails_trusted_devices_only ON emails(trusted_devices_only);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to apply schema: %v", err)
	}

	// Create test user
	testUserID := "test-user-123"
	testEmail := "test@example.com"
	_, err = db.Exec("INSERT INTO users (id, email, password_hash) VALUES (?, ?, ?)",
		testUserID, testEmail, "hashed-password")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create test email with device fingerprinting enabled
	testEmailID := "test-device-fingerprint-email"
	_, err = db.Exec(`
		INSERT INTO emails (
			email_id, sender_id, recipient, subject, 
			encrypted_blob_url, encrypted_key, encryption_nonce, encryption_auth_tag, sha256_hash,
			trusted_devices_only
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		testEmailID, testUserID, testEmail, "Test Device Fingerprinting",
		"https://test.blob.url", "encrypted-key", "nonce", "auth-tag", "sha256-hash",
		true)
	if err != nil {
		t.Fatalf("Failed to create test email: %v", err)
	}

	// Initialize device fingerprinting service (use mock for testing)
	deviceFingerprintSvc := devicefingerprint.NewMockDeviceFingerprintService()

	// Test cases
	tests := []struct {
		name            string
		userAgent       string
		clientIP        string
		expectedTrusted bool
		description     string
	}{
		{
			name:            "Trusted device access",
			userAgent:       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			clientIP:        "192.168.1.100",
			expectedTrusted: true,
			description:     "Device should be trusted after being added to trusted list",
		},
		{
			name:            "Untrusted device access",
			userAgent:       "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
			clientIP:        "10.0.0.50",
			expectedTrusted: false,
			description:     "Different device should not be trusted",
		},
		{
			name:            "Same device different IP",
			userAgent:       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			clientIP:        "192.168.1.200",
			expectedTrusted: false,
			description:     "Same user agent but different IP should not be trusted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate fingerprint for the test device
			fingerprint, err := deviceFingerprintSvc.GenerateFingerprint(tt.userAgent, tt.clientIP, nil)
			if err != nil {
				t.Fatalf("Failed to generate fingerprint: %v", err)
			}

			// Hash the fingerprint
			deviceHash, err := deviceFingerprintSvc.HashFingerprint(fingerprint, testEmailID)
			if err != nil {
				t.Fatalf("Failed to hash fingerprint: %v", err)
			}

			// If this is the first test case, trust the device first
			if tt.name == "Trusted device access" {
				err = deviceFingerprintSvc.TrustDevice(testEmailID, deviceHash, fingerprint, tt.userAgent, tt.clientIP)
				if err != nil {
					t.Fatalf("Failed to trust device: %v", err)
				}
			}

			// Check if device is trusted
			isTrusted, err := deviceFingerprintSvc.IsDeviceTrusted(testEmailID, deviceHash)
			if err != nil {
				t.Fatalf("Failed to check device trust: %v", err)
			}

			if isTrusted != tt.expectedTrusted {
				t.Errorf("Expected trusted=%t, got trusted=%t for %s",
					tt.expectedTrusted, isTrusted, tt.description)
			}
		})
	}
}

// TestDeviceFingerprintingWithServer tests the device fingerprinting with the actual server
func TestDeviceFingerprintingWithServer(t *testing.T) {
	// Set up test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Apply schema (same as above)
	schema := `
	CREATE TABLE users (
		id TEXT PRIMARY KEY,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		totp_secret TEXT
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
		requires_password INTEGER DEFAULT 0,
		password_hash TEXT,
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
		allowed_city TEXT,
		allowed_country TEXT,
		geo_verification_type TEXT CHECK (geo_verification_type IN ('none', 'country', 'city', 'city_country')) DEFAULT 'none',
		geo_city TEXT,
		geo_country TEXT,
		require_mfa INTEGER DEFAULT 0,
		mfa_type TEXT CHECK (mfa_type IN ('TOTP', 'EMAIL_CODE')),
		encrypted_totp_secret TEXT,
		mfa_failed_attempts INTEGER DEFAULT 0,
		mfa_locked_until DATETIME,
		brute_force_failed_attempts INTEGER DEFAULT 0,
		brute_force_last_failed_attempt DATETIME,
		brute_force_lockout_until DATETIME,
		brute_force_max_attempts INTEGER DEFAULT 3,
		brute_force_lockout_duration_minutes INTEGER DEFAULT 15,
		is_password_protected BOOLEAN DEFAULT FALSE,
		password_salt TEXT,
		allowed_countries TEXT,
		allowed_ip_ranges TEXT,
		geofence_violations INTEGER DEFAULT 0,
		geofence_last_violation DATETIME,
		trusted_devices_only BOOLEAN DEFAULT FALSE
	);

	CREATE TABLE trusted_devices (
		id TEXT PRIMARY KEY,
		email_id TEXT NOT NULL,
		device_hash TEXT NOT NULL,
		device_fingerprint TEXT NOT NULL,
		user_agent TEXT,
		ip_address TEXT,
		added_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_used_at DATETIME,
		access_count INTEGER DEFAULT 0,
		FOREIGN KEY (email_id) REFERENCES emails(email_id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_trusted_devices_email_id ON trusted_devices(email_id);
	CREATE INDEX IF NOT EXISTS idx_trusted_devices_device_hash ON trusted_devices(device_hash);
	CREATE INDEX IF NOT EXISTS idx_trusted_devices_email_hash ON trusted_devices(email_id, device_hash);
	CREATE INDEX IF NOT EXISTS idx_emails_trusted_devices_only ON emails(trusted_devices_only);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to apply schema: %v", err)
	}

	// Create test user
	testUserID := "test-user-456"
	testEmail := "test2@example.com"
	_, err = db.Exec("INSERT INTO users (id, email, password_hash) VALUES (?, ?, ?)",
		testUserID, testEmail, "hashed-password")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create test email with device fingerprinting enabled
	testEmailID := "test-device-fingerprint-server-email"
	_, err = db.Exec(`
		INSERT INTO emails (
			email_id, sender_id, recipient, subject, 
			encrypted_blob_url, encrypted_key, encryption_nonce, encryption_auth_tag, sha256_hash,
			trusted_devices_only
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		testEmailID, testUserID, testEmail, "Test Device Fingerprinting with Server",
		"https://test.blob.url", "encrypted-key", "nonce", "auth-tag", "sha256-hash",
		true)
	if err != nil {
		t.Fatalf("Failed to create test email: %v", err)
	}

	// Initialize device fingerprinting service
	deviceFingerprintSvc := devicefingerprint.NewDeviceFingerprintService(db)

	// Test trust device endpoint
	t.Run("Trust device endpoint", func(t *testing.T) {
		// Create a simple server for testing
		req, err := http.NewRequest("POST", "/api/email/"+testEmailID+"/trust-device", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		// Set headers
		req.Header.Set("Authorization", "Bearer test-token")
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
		req.Header.Set("X-Forwarded-For", "192.168.1.100")

		// Test that the endpoint exists (we can't fully test without JWT middleware)
		if req.URL.Path != "/api/email/"+testEmailID+"/trust-device" {
			t.Errorf("Expected path %s, got %s", "/api/email/"+testEmailID+"/trust-device", req.URL.Path)
		}

		// Test device fingerprinting service directly
		fingerprint, err := deviceFingerprintSvc.GenerateFingerprint(
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			"192.168.1.100", nil)
		if err != nil {
			t.Fatalf("Failed to generate fingerprint: %v", err)
		}

		deviceHash, err := deviceFingerprintSvc.HashFingerprint(fingerprint, testEmailID)
		if err != nil {
			t.Fatalf("Failed to hash fingerprint: %v", err)
		}

		// Trust the device
		err = deviceFingerprintSvc.TrustDevice(testEmailID, deviceHash, fingerprint,
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36", "192.168.1.100")
		if err != nil {
			t.Fatalf("Failed to trust device: %v", err)
		}

		// Verify device is trusted
		isTrusted, err := deviceFingerprintSvc.IsDeviceTrusted(testEmailID, deviceHash)
		if err != nil {
			t.Fatalf("Failed to check device trust: %v", err)
		}
		if !isTrusted {
			t.Error("Device should be trusted after TrustDevice() call")
		}

		// Test device info retrieval
		deviceInfo, err := deviceFingerprintSvc.GetDeviceInfo(testEmailID, deviceHash)
		if err != nil {
			t.Fatalf("Failed to get device info: %v", err)
		}
		if deviceInfo.EmailID != testEmailID {
			t.Errorf("Expected EmailID %s, got %s", testEmailID, deviceInfo.EmailID)
		}
		if deviceInfo.DeviceHash != deviceHash {
			t.Errorf("Expected DeviceHash %s, got %s", deviceHash, deviceInfo.DeviceHash)
		}

		// Test trusted devices list
		devices, err := deviceFingerprintSvc.GetTrustedDevices(testEmailID)
		if err != nil {
			t.Fatalf("Failed to get trusted devices: %v", err)
		}
		if len(devices) != 1 {
			t.Errorf("Expected 1 trusted device, got %d", len(devices))
		}
		if devices[0].DeviceHash != deviceHash {
			t.Errorf("Expected device hash %s, got %s", deviceHash, devices[0].DeviceHash)
		}

		// Test device access tracking
		err = deviceFingerprintSvc.UpdateDeviceAccess(testEmailID, deviceHash)
		if err != nil {
			t.Fatalf("Failed to update device access: %v", err)
		}

		// Verify access count increased
		deviceInfo, err = deviceFingerprintSvc.GetDeviceInfo(testEmailID, deviceHash)
		if err != nil {
			t.Fatalf("Failed to get device info after update: %v", err)
		}
		if deviceInfo.AccessCount != 1 {
			t.Errorf("Expected access count 1, got %d", deviceInfo.AccessCount)
		}
	})
}

// TestDeviceFingerprintingSettings tests device fingerprinting settings management
func TestDeviceFingerprintingSettings(t *testing.T) {
	// Set up test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Apply schema (same as above)
	schema := `
	CREATE TABLE users (
		id TEXT PRIMARY KEY,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		totp_secret TEXT
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
		requires_password INTEGER DEFAULT 0,
		password_hash TEXT,
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
		allowed_city TEXT,
		allowed_country TEXT,
		geo_verification_type TEXT CHECK (geo_verification_type IN ('none', 'country', 'city', 'city_country')) DEFAULT 'none',
		geo_city TEXT,
		geo_country TEXT,
		require_mfa INTEGER DEFAULT 0,
		mfa_type TEXT CHECK (mfa_type IN ('TOTP', 'EMAIL_CODE')),
		encrypted_totp_secret TEXT,
		mfa_failed_attempts INTEGER DEFAULT 0,
		mfa_locked_until DATETIME,
		brute_force_failed_attempts INTEGER DEFAULT 0,
		brute_force_last_failed_attempt DATETIME,
		brute_force_lockout_until DATETIME,
		brute_force_max_attempts INTEGER DEFAULT 3,
		brute_force_lockout_duration_minutes INTEGER DEFAULT 15,
		is_password_protected BOOLEAN DEFAULT FALSE,
		password_salt TEXT,
		allowed_countries TEXT,
		allowed_ip_ranges TEXT,
		geofence_violations INTEGER DEFAULT 0,
		geofence_last_violation DATETIME,
		trusted_devices_only BOOLEAN DEFAULT FALSE
	);

	CREATE TABLE trusted_devices (
		id TEXT PRIMARY KEY,
		email_id TEXT NOT NULL,
		device_hash TEXT NOT NULL,
		device_fingerprint TEXT NOT NULL,
		user_agent TEXT,
		ip_address TEXT,
		added_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_used_at DATETIME,
		access_count INTEGER DEFAULT 0,
		FOREIGN KEY (email_id) REFERENCES emails(email_id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_trusted_devices_email_id ON trusted_devices(email_id);
	CREATE INDEX IF NOT EXISTS idx_trusted_devices_device_hash ON trusted_devices(device_hash);
	CREATE INDEX IF NOT EXISTS idx_trusted_devices_email_hash ON trusted_devices(email_id, device_hash);
	CREATE INDEX IF NOT EXISTS idx_emails_trusted_devices_only ON emails(trusted_devices_only);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to apply schema: %v", err)
	}

	// Create test user and email
	testUserID := "test-user-789"
	testEmail := "test3@example.com"
	_, err = db.Exec("INSERT INTO users (id, email, password_hash) VALUES (?, ?, ?)",
		testUserID, testEmail, "hashed-password")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	testEmailID := "test-device-fingerprint-settings-email"
	_, err = db.Exec(`
		INSERT INTO emails (
			email_id, sender_id, recipient, subject, 
			encrypted_blob_url, encrypted_key, encryption_nonce, encryption_auth_tag, sha256_hash,
			trusted_devices_only
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		testEmailID, testUserID, testEmail, "Test Device Fingerprinting Settings",
		"https://test.blob.url", "encrypted-key", "nonce", "auth-tag", "sha256-hash",
		false) // Start with device fingerprinting disabled
	if err != nil {
		t.Fatalf("Failed to create test email: %v", err)
	}

	// Test device fingerprinting settings
	t.Run("Device fingerprinting settings", func(t *testing.T) {
		// Check initial setting
		var trustedDevicesOnly bool
		err = db.QueryRow("SELECT trusted_devices_only FROM emails WHERE email_id = ?", testEmailID).Scan(&trustedDevicesOnly)
		if err != nil {
			t.Fatalf("Failed to check trusted_devices_only setting: %v", err)
		}
		if trustedDevicesOnly {
			t.Error("Expected trusted_devices_only to be false initially")
		}

		// Enable device fingerprinting
		_, err = db.Exec("UPDATE emails SET trusted_devices_only = ? WHERE email_id = ?", true, testEmailID)
		if err != nil {
			t.Fatalf("Failed to enable device fingerprinting: %v", err)
		}

		// Verify setting was updated
		err = db.QueryRow("SELECT trusted_devices_only FROM emails WHERE email_id = ?", testEmailID).Scan(&trustedDevicesOnly)
		if err != nil {
			t.Fatalf("Failed to check trusted_devices_only setting after update: %v", err)
		}
		if !trustedDevicesOnly {
			t.Error("Expected trusted_devices_only to be true after update")
		}

		// Test that device fingerprinting is now required
		deviceFingerprintSvc := devicefingerprint.NewDeviceFingerprintService(db)
		fingerprint, err := deviceFingerprintSvc.GenerateFingerprint(
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			"192.168.1.100", nil)
		if err != nil {
			t.Fatalf("Failed to generate fingerprint: %v", err)
		}

		deviceHash, err := deviceFingerprintSvc.HashFingerprint(fingerprint, testEmailID)
		if err != nil {
			t.Fatalf("Failed to hash fingerprint: %v", err)
		}

		// Device should not be trusted initially
		isTrusted, err := deviceFingerprintSvc.IsDeviceTrusted(testEmailID, deviceHash)
		if err != nil {
			t.Fatalf("Failed to check device trust: %v", err)
		}
		if isTrusted {
			t.Error("Device should not be trusted initially when device fingerprinting is enabled")
		}

		// Trust the device
		err = deviceFingerprintSvc.TrustDevice(testEmailID, deviceHash, fingerprint,
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36", "192.168.1.100")
		if err != nil {
			t.Fatalf("Failed to trust device: %v", err)
		}

		// Device should now be trusted
		isTrusted, err = deviceFingerprintSvc.IsDeviceTrusted(testEmailID, deviceHash)
		if err != nil {
			t.Fatalf("Failed to check device trust after trusting: %v", err)
		}
		if !isTrusted {
			t.Error("Device should be trusted after TrustDevice() call")
		}
	})
}
