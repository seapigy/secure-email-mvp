package mfa

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

// Mock database for testing
func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create emails table with MFA fields
	_, err = db.Exec(`
		CREATE TABLE emails (
			email_id TEXT PRIMARY KEY,
			require_mfa INTEGER DEFAULT 0,
			mfa_type TEXT CHECK (mfa_type IN ('TOTP', 'EMAIL_CODE')),
			encrypted_totp_secret TEXT,
			mfa_failed_attempts INTEGER DEFAULT 0,
			mfa_locked_until DATETIME
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	return db
}

func TestGenerateEmailCode(t *testing.T) {
	mfaService := &MFAService{db: nil} // db not needed for this test

	code, err := mfaService.GenerateEmailCode()
	if err != nil {
		t.Fatalf("Failed to generate email code: %v", err)
	}

	// Check code length
	if len(code) != 6 {
		t.Errorf("Expected 6-digit code, got %d digits", len(code))
	}

	// Check that it's numeric
	for _, char := range code {
		if char < '0' || char > '9' {
			t.Errorf("Code contains non-numeric character: %c", char)
		}
	}

	// Generate multiple codes to ensure randomness
	codes := make(map[string]bool)
	for i := 0; i < 100; i++ {
		code, err := mfaService.GenerateEmailCode()
		if err != nil {
			t.Fatalf("Failed to generate email code %d: %v", i, err)
		}
		if codes[code] {
			t.Errorf("Duplicate code generated: %s", code)
		}
		codes[code] = true
	}
}

func TestStoreAndValidateEmailCode(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	mfaService := NewMFAService(db)
	emailID := "test-email-123"

	// Insert email record first
	_, err := db.Exec(`
		INSERT INTO emails (email_id) 
		VALUES (?)
	`, emailID)
	if err != nil {
		t.Fatalf("Failed to insert email record: %v", err)
	}

	// Store a test code
	testCode := "123456"
	err = mfaService.StoreEmailCode(emailID, testCode)
	if err != nil {
		t.Fatalf("Failed to store email code: %v", err)
	}

	// Validate the correct code
	valid, err := mfaService.ValidateEmailCode(emailID, testCode)
	if err != nil {
		t.Fatalf("Failed to validate email code: %v", err)
	}
	if !valid {
		t.Error("Valid email code was rejected")
	}

	// Validate incorrect code
	valid, err = mfaService.ValidateEmailCode(emailID, "000000")
	if err != nil {
		t.Fatalf("Failed to validate incorrect email code: %v", err)
	}
	if valid {
		t.Error("Invalid email code was accepted")
	}

	// Test case-insensitive validation
	valid, err = mfaService.ValidateEmailCode(emailID, "123456")
	if err != nil {
		t.Fatalf("Failed to validate email code: %v", err)
	}
	if !valid {
		t.Error("Valid email code was rejected (case-insensitive)")
	}
}

func TestCheckMFALockout(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	mfaService := NewMFAService(db)
	emailID := "test-email-123"

	// Insert email record first
	_, err := db.Exec(`
		INSERT INTO emails (email_id) 
		VALUES (?)
	`, emailID)
	if err != nil {
		t.Fatalf("Failed to insert email record: %v", err)
	}

	// Initially should not be locked
	locked, lockedUntil, err := mfaService.CheckMFALockout(emailID)
	if err != nil {
		t.Fatalf("Failed to check MFA lockout: %v", err)
	}
	if locked {
		t.Error("Email should not be locked initially")
	}
	if lockedUntil != nil {
		t.Error("Locked until should be nil initially")
	}

	// Set a future lockout time
	futureTime := time.Now().Add(30 * time.Minute)
	_, err = db.Exec(`
		INSERT INTO emails (email_id, mfa_locked_until) 
		VALUES (?, ?) 
		ON CONFLICT(email_id) DO UPDATE SET mfa_locked_until = ?
	`, emailID, futureTime, futureTime)
	if err != nil {
		t.Fatalf("Failed to set lockout time: %v", err)
	}

	// Should be locked
	locked, lockedUntil, err = mfaService.CheckMFALockout(emailID)
	if err != nil {
		t.Fatalf("Failed to check MFA lockout: %v", err)
	}
	if !locked {
		t.Error("Email should be locked")
	}
	if lockedUntil == nil {
		t.Error("Locked until should not be nil")
	}

	// Set a past lockout time
	pastTime := time.Now().Add(-30 * time.Minute)
	_, err = db.Exec(`
		UPDATE emails SET mfa_locked_until = ? WHERE email_id = ?
	`, pastTime, emailID)
	if err != nil {
		t.Fatalf("Failed to update lockout time: %v", err)
	}

	// Should not be locked
	locked, lockedUntil, err = mfaService.CheckMFALockout(emailID)
	if err != nil {
		t.Fatalf("Failed to check MFA lockout: %v", err)
	}
	if locked {
		t.Error("Email should not be locked (past lockout time)")
	}
}

func TestIncrementFailedAttempts(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	mfaService := NewMFAService(db)
	emailID := "test-email-123"

	// Insert initial record
	_, err := db.Exec(`
		INSERT INTO emails (email_id, mfa_failed_attempts) 
		VALUES (?, 0)
	`, emailID)
	if err != nil {
		t.Fatalf("Failed to insert initial record: %v", err)
	}

	// Increment failed attempts
	err = mfaService.IncrementFailedAttempts(emailID)
	if err != nil {
		t.Fatalf("Failed to increment failed attempts: %v", err)
	}

	// Check that attempts were incremented
	var failedAttempts int
	err = db.QueryRow("SELECT mfa_failed_attempts FROM emails WHERE email_id = ?", emailID).Scan(&failedAttempts)
	if err != nil {
		t.Fatalf("Failed to get failed attempts: %v", err)
	}
	if failedAttempts != 1 {
		t.Errorf("Expected 1 failed attempt, got %d", failedAttempts)
	}

	// Increment more times to trigger lockout
	for i := 0; i < 4; i++ {
		err = mfaService.IncrementFailedAttempts(emailID)
		if err != nil {
			t.Fatalf("Failed to increment failed attempts: %v", err)
		}
	}

	// Check that lockout was triggered
	var lockedUntil *time.Time
	err = db.QueryRow("SELECT mfa_locked_until FROM emails WHERE email_id = ?", emailID).Scan(&lockedUntil)
	if err != nil {
		t.Fatalf("Failed to get locked until: %v", err)
	}
	if lockedUntil == nil {
		t.Error("Lockout should have been triggered")
	}

	// Check that it's locked
	locked, _, err := mfaService.CheckMFALockout(emailID)
	if err != nil {
		t.Fatalf("Failed to check MFA lockout: %v", err)
	}
	if !locked {
		t.Error("Email should be locked after 5 failed attempts")
	}
}

func TestResetFailedAttempts(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	mfaService := NewMFAService(db)
	emailID := "test-email-123"

	// Insert record with failed attempts and lockout
	lockoutTime := time.Now().Add(30 * time.Minute)
	_, err := db.Exec(`
		INSERT INTO emails (email_id, mfa_failed_attempts, mfa_locked_until) 
		VALUES (?, 5, ?)
	`, emailID, lockoutTime)
	if err != nil {
		t.Fatalf("Failed to insert record: %v", err)
	}

	// Reset failed attempts
	err = mfaService.ResetFailedAttempts(emailID)
	if err != nil {
		t.Fatalf("Failed to reset failed attempts: %v", err)
	}

	// Check that attempts were reset
	var failedAttempts int
	var lockedUntil *time.Time
	err = db.QueryRow("SELECT mfa_failed_attempts, mfa_locked_until FROM emails WHERE email_id = ?", emailID).Scan(&failedAttempts, &lockedUntil)
	if err != nil {
		t.Fatalf("Failed to get reset values: %v", err)
	}
	if failedAttempts != 0 {
		t.Errorf("Expected 0 failed attempts, got %d", failedAttempts)
	}
	if lockedUntil != nil {
		t.Error("Lockout should have been cleared")
	}

	// Check that it's not locked
	locked, _, err := mfaService.CheckMFALockout(emailID)
	if err != nil {
		t.Fatalf("Failed to check MFA lockout: %v", err)
	}
	if locked {
		t.Error("Email should not be locked after reset")
	}
}

func TestGetMFAConfig(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	mfaService := NewMFAService(db)
	emailID := "test-email-123"

	// Insert test record
	lockoutTime := time.Now().Add(30 * time.Minute)
	_, err := db.Exec(`
		INSERT INTO emails (email_id, require_mfa, mfa_type, mfa_failed_attempts, mfa_locked_until) 
		VALUES (?, 1, 'TOTP', 3, ?)
	`, emailID, lockoutTime)
	if err != nil {
		t.Fatalf("Failed to insert test record: %v", err)
	}

	// Get MFA config
	config, err := mfaService.GetMFAConfig(emailID)
	if err != nil {
		t.Fatalf("Failed to get MFA config: %v", err)
	}

	// Verify config
	if !config.RequireMFA {
		t.Error("RequireMFA should be true")
	}
	if config.MFAType != MFATypeTOTP {
		t.Errorf("Expected MFA type TOTP, got %s", config.MFAType)
	}
	if config.FailedAttempts != 3 {
		t.Errorf("Expected 3 failed attempts, got %d", config.FailedAttempts)
	}
	if config.LockedUntil == nil {
		t.Error("LockedUntil should not be nil")
	}
}

func TestGenerateTOTPSecret(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	mfaService := NewMFAService(db)
	emailID := "test-email-123"

	// Generate TOTP secret
	totpConfig, err := mfaService.GenerateTOTPSecret(emailID)
	if err != nil {
		t.Fatalf("Failed to generate TOTP secret: %v", err)
	}

	// Verify TOTP config
	if totpConfig.Secret == "" {
		t.Error("TOTP secret should not be empty")
	}
	if totpConfig.QRCodeURL == "" {
		t.Error("QR code URL should not be empty")
	}
	if totpConfig.Issuer != "Secure Email MVP" {
		t.Errorf("Expected issuer 'Secure Email MVP', got %s", totpConfig.Issuer)
	}
	if totpConfig.Account != emailID {
		t.Errorf("Expected account %s, got %s", emailID, totpConfig.Account)
	}

	// Verify QR code URL format
	if len(totpConfig.QRCodeURL) < 50 {
		t.Error("QR code URL seems too short")
	}
}

func TestValidateTOTP(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	mfaService := NewMFAService(db)
	emailID := "test-email-123"

	// Generate TOTP secret
	totpConfig, err := mfaService.GenerateTOTPSecret(emailID)
	if err != nil {
		t.Fatalf("Failed to generate TOTP secret: %v", err)
	}

	// Verify TOTP config was generated
	if totpConfig.Secret == "" {
		t.Fatal("TOTP secret should not be empty")
	}

	// Store encrypted TOTP secret in database
	encryptedComponents := map[string]string{
		"ciphertext": "test-ciphertext",
		"key":        "test-key",
		"nonce":      "test-nonce",
		"auth_tag":   "test-auth-tag",
	}
	encryptedJSON, err := json.Marshal(encryptedComponents)
	if err != nil {
		t.Fatalf("Failed to marshal encrypted components: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO emails (email_id, encrypted_totp_secret) 
		VALUES (?, ?)
	`, emailID, string(encryptedJSON))
	if err != nil {
		t.Fatalf("Failed to insert encrypted TOTP secret: %v", err)
	}

	// This test would require a more complex setup with actual encryption/decryption
	// For now, we'll just test that the function exists and doesn't panic
	_, err = mfaService.ValidateTOTP(emailID, "123456")
	// We expect an error because we're using fake encrypted data
	if err == nil {
		t.Error("Expected error with fake encrypted data")
	}
}

func TestMFATypeConstants(t *testing.T) {
	// Test MFA type constants
	if MFATypeTOTP != "TOTP" {
		t.Errorf("Expected MFATypeTOTP to be 'TOTP', got %s", MFATypeTOTP)
	}
	if MFATypeEmailCode != "EMAIL_CODE" {
		t.Errorf("Expected MFATypeEmailCode to be 'EMAIL_CODE', got %s", MFATypeEmailCode)
	}
}
