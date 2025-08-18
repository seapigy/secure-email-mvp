package zkid

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func setupIntegrationTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}

	// Create all required tables
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
            updated_at DATETIME
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
            used_at DATETIME
        );`)
	if err != nil {
		t.Fatal(err)
	}

	return db
}

func TestZKIDIntegrationFlow(t *testing.T) {
	db := setupIntegrationTestDB(t)

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

	userID := "integration-test-user"
	email := "test@example.com"

	// Test 1: Create email mapping
	mapping, err := svc.CreateOrUpdateMapping(userID, email, nil)
	if err != nil {
		t.Fatalf("create mapping failed: %v", err)
	}
	if mapping.UserID != userID {
		t.Fatalf("expected user ID %s, got %s", userID, mapping.UserID)
	}

	// Test 2: Retrieve email mapping
	retrievedEmail, err := svc.GetEmailByUserID(userID)
	if err != nil {
		t.Fatalf("get email failed: %v", err)
	}
	if retrievedEmail != "test@example.com" {
		t.Fatalf("expected email test@example.com, got %s", retrievedEmail)
	}

	// Test 3: Generate recovery codes
	codes, err := svc.GenerateRecoveryCodes(userID, 3)
	if err != nil {
		t.Fatalf("generate recovery codes failed: %v", err)
	}
	if len(codes) != 3 {
		t.Fatalf("expected 3 codes, got %d", len(codes))
	}

	// Test 4: Validate recovery code
	valid, err := svc.ValidateAndConsumeRecoveryCode(userID, codes[0])
	if err != nil {
		t.Fatalf("validate recovery code failed: %v", err)
	}
	if !valid {
		t.Fatal("recovery code should be valid")
	}

	// Test 5: Get code ID for revocation
	var codeID string
	err = db.QueryRow("SELECT id FROM zkid_recovery_codes WHERE user_id = ? AND used = 0 LIMIT 1", userID).Scan(&codeID)
	if err != nil {
		t.Fatalf("get code ID failed: %v", err)
	}

	// Test 6: Revoke recovery code
	success, err := svc.RevokeRecoveryCode(userID, codeID)
	if err != nil {
		t.Fatalf("revoke recovery code failed: %v", err)
	}
	if !success {
		t.Fatal("revoke should succeed")
	}

	// Test 7: Get statistics
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
	if stats["total_recovery_codes"] != 3 {
		t.Fatalf("expected 3 recovery codes, got %v", stats["total_recovery_codes"])
	}
	if stats["used_recovery_codes"] != 2 { // 1 consumed + 1 revoked
		t.Fatalf("expected 2 used codes, got %v", stats["used_recovery_codes"])
	}
}

func TestZKIDDisabledMode(t *testing.T) {
	db := setupIntegrationTestDB(t)

	// Create service with ZKID disabled
	svc := NewService(db, &Config{Enabled: false})

	userID := "disabled-test-user"
	email := "test@example.com"

	// Test that operations fail when disabled
	_, err := svc.CreateOrUpdateMapping(userID, email, nil)
	if err == nil {
		t.Fatal("create mapping should fail when disabled")
	}

	_, err = svc.GetEmailByUserID(userID)
	if err == nil {
		t.Fatal("get email should fail when disabled")
	}

	_, err = svc.GenerateRecoveryCodes(userID, 3)
	if err == nil {
		t.Fatal("generate recovery codes should fail when disabled")
	}

	stats, err := svc.GetStats()
	if err != nil {
		t.Fatalf("get stats failed: %v", err)
	}
	if stats["enabled"] != false {
		t.Fatal("stats should show disabled")
	}
}
