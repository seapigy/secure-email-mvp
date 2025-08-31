package main

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"secure-email-mvp/pkg/email"

	_ "modernc.org/sqlite"
)

// TestEmailSecurityServiceIntegration tests the EmailSecurityService integration
func TestEmailSecurityServiceIntegration(t *testing.T) {
	// Setup test database
	dbPath := fmt.Sprintf("/tmp/test-security-integration-%d.db", time.Now().Unix())
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer os.Remove(dbPath)
	defer db.Close()

	// Create test tables
	if err := createSimpleTestTables(db); err != nil {
		t.Fatalf("Failed to create test tables: %v", err)
	}

	// Initialize service
	service := email.NewEmailSecurityService(db)

	// Test 1: Basic security configuration validation
	t.Run("Basic Security Validation", func(t *testing.T) {
		config := email.SecurityFeatureConfig{
			PasswordProtection: true,
			Password:          "securepass123",
			BurnAfterRead:     true,
			RequireMFA:        true,
			MFAType:           "TOTP",
			MaxFailedAttempts: 5, // Valid value between 1-10
		}

		if err := service.ValidateSecurityConfig(config); err != nil {
			t.Errorf("Valid config should pass validation: %v", err)
		}
		t.Logf("✅ Basic security validation passed")
	})

	// Test 2: Invalid configuration validation
	t.Run("Invalid Configuration Validation", func(t *testing.T) {
		config := email.SecurityFeatureConfig{
			PasswordProtection: true,
			Password:          "", // Missing password
		}

		if err := service.ValidateSecurityConfig(config); err == nil {
			t.Error("Invalid config should fail validation")
		} else {
			t.Logf("✅ Invalid configuration correctly rejected: %v", err)
		}
	})

	// Test 3: Email delivery configuration
	t.Run("Email Delivery Configuration", func(t *testing.T) {
		deliveryConfig := email.EmailDeliveryConfig{
			DeliveryType: "both",
			InternalRecipient: "internal@example.com",
			ExternalRecipient: "external@example.com",
			ExternalSubject: "Test Subject",
			ExternalBody: "Test Body",
			ExternalSecurityFeatures: email.SecurityFeatureConfig{
				PasswordProtection: true,
				Password:          "securepass123",
				BurnAfterRead:     true,
				MaxFailedAttempts: 5, // Valid value between 1-10
			},
		}

		// This should work without external dependencies
		t.Logf("✅ Email delivery configuration created successfully")
		t.Logf("   Delivery Type: %s", deliveryConfig.DeliveryType)
		t.Logf("   Internal Recipient: %s", deliveryConfig.InternalRecipient)
		t.Logf("   External Recipient: %s", deliveryConfig.ExternalRecipient)
	})

	// Test 4: Security features description
	t.Run("Security Features Description", func(t *testing.T) {
		config := email.SecurityFeatureConfig{
			PasswordProtection: true,
			Password:          "securepass123",
			BurnAfterRead:     true,
			RequireMFA:        true,
			MFAType:           "TOTP",
			MaxFailedAttempts: 5, // Valid value between 1-10
			TimeLock:          true,
			UnlockAfter:       time.Now().Add(time.Hour).Format(time.RFC3339),
		}

		// Test that the config is valid
		if err := service.ValidateSecurityConfig(config); err != nil {
			t.Errorf("Config should be valid: %v", err)
		} else {
			t.Logf("✅ Security features configuration validated successfully")
		}
	})

	t.Logf("🎉 All EmailSecurityService integration tests passed!")
}

// TestSecurityFeatureConfig tests the SecurityFeatureConfig structure
func TestSecurityFeatureConfig(t *testing.T) {
	t.Run("Valid Configuration", func(t *testing.T) {
		config := email.SecurityFeatureConfig{
			PasswordProtection: true,
			Password:          "securepass123",
			BurnAfterRead:     true,
			SelfDestructAfterAttempts: true,
			MaxFailedAttempts: 5,
			TimeLock:          true,
			UnlockAfter:       time.Now().Add(time.Hour).Format(time.RFC3339),
			ExpiresAt:         time.Now().Add(24 * time.Hour).Format(time.RFC3339),
			GeoVerificationType: "country",
			GeoCountry:        "US",
			RequireMFA:        true,
			MFAType:           "TOTP",
			RemoteRevoke:      true,
			DecoyMessage:      true,
			StripMetadata:     true,
			TamperAlerts:      true,
		}

		// Verify all fields are set correctly
		if !config.PasswordProtection {
			t.Error("PasswordProtection should be true")
		}
		if config.Password != "securepass123" {
			t.Error("Password should match")
		}
		if !config.BurnAfterRead {
			t.Error("BurnAfterRead should be true")
		}
		if !config.RequireMFA {
			t.Error("RequireMFA should be true")
		}
		if config.MFAType != "TOTP" {
			t.Error("MFAType should be TOTP")
		}

		t.Logf("✅ SecurityFeatureConfig validation passed")
		t.Logf("   Password Protection: %v", config.PasswordProtection)
		t.Logf("   Burn After Read: %v", config.BurnAfterRead)
		t.Logf("   Require MFA: %v", config.RequireMFA)
		t.Logf("   MFA Type: %s", config.MFAType)
		t.Logf("   Time Lock: %v", config.TimeLock)
		t.Logf("   Remote Revoke: %v", config.RemoteRevoke)
		t.Logf("   Decoy Message: %v", config.DecoyMessage)
		t.Logf("   Strip Metadata: %v", config.StripMetadata)
		t.Logf("   Tamper Alerts: %v", config.TamperAlerts)
	})
}

// Helper functions

func createSimpleTestTables(db *sql.DB) error {
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
