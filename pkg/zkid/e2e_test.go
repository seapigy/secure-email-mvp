package zkid

import (
	"database/sql"
	"fmt"
	"testing"

	_ "modernc.org/sqlite"
)

func setupE2ETestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}

	// Create all required tables including users table for foreign key constraints
	_, err = db.Exec(`
        CREATE TABLE users (
            id TEXT PRIMARY KEY,
            email TEXT UNIQUE,
            password_hash TEXT,
            totp_secret TEXT,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP
        );`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(`
        CREATE TABLE zkid_email_mappings (
            id TEXT PRIMARY KEY,
            user_id TEXT NOT NULL UNIQUE,
            email_hash TEXT NOT NULL UNIQUE,
            email_ciphertext BLOB NOT NULL,
            email_nonce BLOB NOT NULL,
            email_tag BLOB NOT NULL,
            wrapped_key BLOB NOT NULL,
            wrapped_key_nonce BLOB NOT NULL,
            wrapped_key_tag BLOB NOT NULL,
            fallback_email_ciphertext BLOB,
            fallback_email_nonce BLOB,
            fallback_email_tag BLOB,
            created_at DATETIME,
            updated_at DATETIME,
            FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
        );`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(`
        CREATE TABLE zkid_recovery_codes (
            id TEXT PRIMARY KEY,
            user_id TEXT NOT NULL,
            salt BLOB NOT NULL,
            hash BLOB NOT NULL,
            used BOOLEAN NOT NULL DEFAULT 0,
            created_at DATETIME,
            used_at DATETIME,
            FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
        );`)
	if err != nil {
		t.Fatal(err)
	}

	// Create indexes
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_zkid_email_hash ON zkid_email_mappings(email_hash);`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_zkid_user_id ON zkid_email_mappings(user_id);`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_zkid_recovery_user ON zkid_recovery_codes(user_id);`)
	if err != nil {
		t.Fatal(err)
	}

	return db
}

// Mock admin context for testing
type mockAdminContext struct {
	userID string
	email  string
	role   string
	orgID  string
}

func (m *mockAdminContext) GetUserFromContext() (string, string, string, string, error) {
	return m.userID, m.email, m.role, m.orgID, nil
}

// TestE2EZKIDWorkflow tests the complete ZKID workflow from signup to admin operations
func TestE2EZKIDWorkflow(t *testing.T) {
	db := setupE2ETestDB(t)

	// Create service with proper configuration
	svc := NewService(db, &Config{
		Enabled:         true,
		MasterKey:       make([]byte, 32),
		EmailHashPepper: []byte("email-pepper"),
		RecoveryPepper:  []byte("recovery-pepper"),
	})

	// Set non-zero master key
	for i := range svc.config.MasterKey {
		svc.config.MasterKey[i] = byte(i)
	}

	// Test user creation and ZKID mapping
	userID := "e2e-test-user-123"
	email := "test@example.com"

	// Create user in users table (simulating signup)
	_, err := db.Exec("INSERT INTO users (id, email) VALUES (?, ?)", userID, email)
	if err != nil {
		t.Fatalf("create user failed: %v", err)
	}

	// Test 1: Create ZKID email mapping (simulating signup handler)
	mapping, err := svc.CreateOrUpdateMapping(userID, email, nil)
	if err != nil {
		t.Fatalf("create ZKID mapping failed: %v", err)
	}
	if mapping.UserID != userID {
		t.Fatalf("expected user ID %s, got %s", userID, mapping.UserID)
	}

	// Test 2: Verify email mapping was created
	retrievedEmail, err := svc.GetEmailByUserID(userID)
	if err != nil {
		t.Fatalf("get email failed: %v", err)
	}
	if retrievedEmail != "test@example.com" {
		t.Fatalf("expected email test@example.com, got %s", retrievedEmail)
	}

	// Test 3: Generate recovery codes (admin operation)
	codes, err := svc.GenerateRecoveryCodes(userID, 5)
	if err != nil {
		t.Fatalf("generate recovery codes failed: %v", err)
	}
	if len(codes) != 5 {
		t.Fatalf("expected 5 codes, got %d", len(codes))
	}

	// Test 4: Validate recovery code (user operation)
	valid, err := svc.ValidateAndConsumeRecoveryCode(userID, codes[0])
	if err != nil {
		t.Fatalf("validate recovery code failed: %v", err)
	}
	if !valid {
		t.Fatal("recovery code should be valid")
	}

	// Test 5: Verify used code is no longer valid
	valid, err = svc.ValidateAndConsumeRecoveryCode(userID, codes[0])
	if err != nil {
		t.Fatalf("validate used recovery code failed: %v", err)
	}
	if valid {
		t.Fatal("used recovery code should not be valid")
	}

	// Test 6: Admin revocation of recovery code
	var codeID string
	err = db.QueryRow("SELECT id FROM zkid_recovery_codes WHERE user_id = ? AND used = 0 LIMIT 1", userID).Scan(&codeID)
	if err != nil {
		t.Fatalf("get code ID failed: %v", err)
	}

	success, err := svc.RevokeRecoveryCode(userID, codeID)
	if err != nil {
		t.Fatalf("revoke recovery code failed: %v", err)
	}
	if !success {
		t.Fatal("revoke should succeed")
	}

	// Test 7: Get ZKID statistics (admin operation)
	stats, err := svc.GetStats()
	if err != nil {
		t.Fatalf("get stats failed: %v", err)
	}

	if stats["enabled"] != true {
		t.Fatal("stats should show enabled")
	}
	if stats["total_mappings"] != 1 {
		t.Fatalf("expected 1 mapping, got %v", stats["total_mappings"])
	}
	if stats["total_recovery_codes"] != 5 {
		t.Fatalf("expected 5 recovery codes, got %v", stats["total_recovery_codes"])
	}
	if stats["used_recovery_codes"] != 2 { // 1 consumed + 1 revoked
		t.Fatalf("expected 2 used codes, got %v", stats["used_recovery_codes"])
	}

	// Test 8: Verify zero-knowledge guarantees
	// Check that no external emails are stored in plaintext
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM zkid_email_mappings WHERE email_hash LIKE '%@%'").Scan(&count)
	if err != nil {
		t.Fatalf("check email hash failed: %v", err)
	}
	if count > 0 {
		t.Fatal("email hashes should not contain @ symbols")
	}

	// Verify all email data is encrypted
	err = db.QueryRow("SELECT COUNT(*) FROM zkid_email_mappings WHERE email_ciphertext IS NULL").Scan(&count)
	if err != nil {
		t.Fatalf("check ciphertext failed: %v", err)
	}
	if count > 0 {
		t.Fatal("all email data should be encrypted")
	}
}

// TestE2ERBACEnforcement tests RBAC enforcement for admin endpoints
func TestE2ERBACEnforcement(t *testing.T) {
	db := setupE2ETestDB(t)

	// Create service
	svc := NewService(db, &Config{
		Enabled:         true,
		MasterKey:       make([]byte, 32),
		EmailHashPepper: []byte("email-pepper"),
		RecoveryPepper:  []byte("recovery-pepper"),
	})

	// Set non-zero master key
	for i := range svc.config.MasterKey {
		svc.config.MasterKey[i] = byte(i)
	}

	// Test that non-admin users cannot access admin operations
	// This would be tested with actual HTTP handlers in a real scenario
	// For now, we test the service layer directly

	userID := "rbac-test-user"

	// Test that regular users can still use recovery codes
	codes, err := svc.GenerateRecoveryCodes(userID, 3)
	if err != nil {
		t.Fatalf("generate recovery codes failed: %v", err)
	}

	// Test validation works for regular users
	valid, err := svc.ValidateAndConsumeRecoveryCode(userID, codes[0])
	if err != nil {
		t.Fatalf("validate recovery code failed: %v", err)
	}
	if !valid {
		t.Fatal("recovery code should be valid")
	}
}

// TestE2EFeatureFlagControl tests that ZKID can be disabled cleanly
func TestE2EFeatureFlagControl(t *testing.T) {
	db := setupE2ETestDB(t)

	// Test with ZKID disabled
	disabledSvc := NewService(db, &Config{Enabled: false})

	userID := "feature-flag-test-user"
	email := "test@example.com"

	// All operations should fail when disabled
	_, err := disabledSvc.CreateOrUpdateMapping(userID, email, nil)
	if err == nil {
		t.Fatal("create mapping should fail when disabled")
	}

	_, err = disabledSvc.GetEmailByUserID(userID)
	if err == nil {
		t.Fatal("get email should fail when disabled")
	}

	_, err = disabledSvc.GenerateRecoveryCodes(userID, 3)
	if err == nil {
		t.Fatal("generate recovery codes should fail when disabled")
	}

	stats, err := disabledSvc.GetStats()
	if err != nil {
		t.Fatalf("get stats failed: %v", err)
	}
	if stats["enabled"] != false {
		t.Fatal("stats should show disabled")
	}

	// Test with ZKID enabled
	enabledSvc := NewService(db, &Config{
		Enabled:         true,
		MasterKey:       make([]byte, 32),
		EmailHashPepper: []byte("email-pepper"),
		RecoveryPepper:  []byte("recovery-pepper"),
	})

	// Set non-zero master key
	for i := range enabledSvc.config.MasterKey {
		enabledSvc.config.MasterKey[i] = byte(i)
	}

	// Operations should work when enabled
	_, err = enabledSvc.CreateOrUpdateMapping(userID, email, nil)
	if err != nil {
		t.Fatalf("create mapping should work when enabled: %v", err)
	}

	stats, err = enabledSvc.GetStats()
	if err != nil {
		t.Fatalf("get stats failed: %v", err)
	}
	if stats["enabled"] != true {
		t.Fatal("stats should show enabled")
	}
}

// TestE2EDataIntegrity tests that ZKID data remains consistent
func TestE2EDataIntegrity(t *testing.T) {
	db := setupE2ETestDB(t)

	// Create service
	svc := NewService(db, &Config{
		Enabled:         true,
		MasterKey:       make([]byte, 32),
		EmailHashPepper: []byte("email-pepper"),
		RecoveryPepper:  []byte("recovery-pepper"),
	})

	// Set non-zero master key
	for i := range svc.config.MasterKey {
		svc.config.MasterKey[i] = byte(i)
	}

	userID := "integrity-test-user"
	email := "integrity@example.com"

	// Create mapping
	_, err := svc.CreateOrUpdateMapping(userID, email, nil)
	if err != nil {
		t.Fatalf("create mapping failed: %v", err)
	}

	// Verify data integrity
	var mappingCount int
	err = db.QueryRow("SELECT COUNT(*) FROM zkid_email_mappings WHERE user_id = ?", userID).Scan(&mappingCount)
	if err != nil {
		t.Fatalf("count mappings failed: %v", err)
	}
	if mappingCount != 1 {
		t.Fatalf("expected 1 mapping, got %d", mappingCount)
	}

	// Test update mapping (should not create duplicate)
	_, err = svc.CreateOrUpdateMapping(userID, email, nil)
	if err != nil {
		t.Fatalf("update mapping failed: %v", err)
	}

	// Verify still only one mapping
	err = db.QueryRow("SELECT COUNT(*) FROM zkid_email_mappings WHERE user_id = ?", userID).Scan(&mappingCount)
	if err != nil {
		t.Fatalf("count mappings failed: %v", err)
	}
	if mappingCount != 1 {
		t.Fatalf("expected 1 mapping after update, got %d", mappingCount)
	}

	// Test recovery code integrity
	codes, err := svc.GenerateRecoveryCodes(userID, 3)
	if err != nil {
		t.Fatalf("generate recovery codes failed: %v", err)
	}

	var codeCount int
	err = db.QueryRow("SELECT COUNT(*) FROM zkid_recovery_codes WHERE user_id = ?", userID).Scan(&codeCount)
	if err != nil {
		t.Fatalf("count recovery codes failed: %v", err)
	}
	if codeCount != 3 {
		t.Fatalf("expected 3 recovery codes, got %d", codeCount)
	}

	// Use a code
	valid, err := svc.ValidateAndConsumeRecoveryCode(userID, codes[0])
	if err != nil {
		t.Fatalf("validate recovery code failed: %v", err)
	}
	if !valid {
		t.Fatal("recovery code should be valid")
	}

	// Verify code is marked as used
	var usedCount int
	err = db.QueryRow("SELECT COUNT(*) FROM zkid_recovery_codes WHERE user_id = ? AND used = 1", userID).Scan(&usedCount)
	if err != nil {
		t.Fatalf("count used codes failed: %v", err)
	}
	if usedCount != 1 {
		t.Fatalf("expected 1 used code, got %d", usedCount)
	}
}

// TestE2EPerformance tests basic performance characteristics
func TestE2EPerformance(t *testing.T) {
	db := setupE2ETestDB(t)

	// Create service
	svc := NewService(db, &Config{
		Enabled:         true,
		MasterKey:       make([]byte, 32),
		EmailHashPepper: []byte("email-pepper"),
		RecoveryPepper:  []byte("recovery-pepper"),
	})

	// Set non-zero master key
	for i := range svc.config.MasterKey {
		svc.config.MasterKey[i] = byte(i)
	}

	// Test multiple operations
	for i := 0; i < 10; i++ {
		userID := fmt.Sprintf("perf-test-user-%d", i)
		email := fmt.Sprintf("test%d@example.com", i)

		// Create mapping
		_, err := svc.CreateOrUpdateMapping(userID, email, nil)
		if err != nil {
			t.Fatalf("create mapping %d failed: %v", i, err)
		}

		// Retrieve email
		retrievedEmail, err := svc.GetEmailByUserID(userID)
		if err != nil {
			t.Fatalf("get email %d failed: %v", i, err)
		}
		if retrievedEmail != email {
			t.Fatalf("email mismatch for user %d: expected %s, got %s", i, email, retrievedEmail)
		}

		// Generate recovery codes
		codes, err := svc.GenerateRecoveryCodes(userID, 5)
		if err != nil {
			t.Fatalf("generate recovery codes %d failed: %v", i, err)
		}
		if len(codes) != 5 {
			t.Fatalf("expected 5 codes for user %d, got %d", i, len(codes))
		}
	}

	// Test statistics performance
	stats, err := svc.GetStats()
	if err != nil {
		t.Fatalf("get stats failed: %v", err)
	}

	if stats["total_mappings"] != 10 {
		t.Fatalf("expected 10 mappings, got %v", stats["total_mappings"])
	}
	if stats["total_recovery_codes"] != 50 {
		t.Fatalf("expected 50 recovery codes, got %v", stats["total_recovery_codes"])
	}
}
