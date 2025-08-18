package zkid

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
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
            updated_at DATETIME
        );`)
	if err != nil {
		t.Fatal(err)
	}

	// Add recovery codes table
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

func TestMappingRoundTrip(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db, &Config{Enabled: true, MasterKey: make([]byte, 32), EmailHashPepper: []byte("pepper")})
	// set a non-zero master key
	for i := range svc.config.MasterKey {
		svc.config.MasterKey[i] = byte(i)
	}

	userID := "test-user-id"
	email := "User@Example.com"
	if _, err := svc.CreateOrUpdateMapping(userID, email, nil); err != nil {
		t.Fatalf("create mapping failed: %v", err)
	}
	got, err := svc.GetEmailByUserID(userID)
	if err != nil {
		t.Fatalf("get email failed: %v", err)
	}
	if got != "user@example.com" {
		t.Fatalf("expected normalized email, got %q", got)
	}
}

func TestRecoveryCodeGeneration(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db, &Config{Enabled: true, RecoveryPepper: []byte("recovery-pepper")})

	userID := "test-user-id"
	codes, err := svc.GenerateRecoveryCodes(userID, 5)
	if err != nil {
		t.Fatalf("generate recovery codes failed: %v", err)
	}

	if len(codes) != 5 {
		t.Fatalf("expected 5 codes, got %d", len(codes))
	}

	// Test validation
	valid, err := svc.ValidateAndConsumeRecoveryCode(userID, codes[0])
	if err != nil {
		t.Fatalf("validate recovery code failed: %v", err)
	}
	if !valid {
		t.Fatal("recovery code should be valid")
	}

	// Test that used code is no longer valid
	valid, err = svc.ValidateAndConsumeRecoveryCode(userID, codes[0])
	if err != nil {
		t.Fatalf("validate used recovery code failed: %v", err)
	}
	if valid {
		t.Fatal("used recovery code should not be valid")
	}
}

func TestRevokeRecoveryCode(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db, &Config{Enabled: true, RecoveryPepper: []byte("recovery-pepper")})

	userID := "test-user-id"
	_, err := svc.GenerateRecoveryCodes(userID, 3)
	if err != nil {
		t.Fatalf("generate recovery codes failed: %v", err)
	}

	// Get the code ID from the database
	var codeID string
	err = db.QueryRow("SELECT id FROM zkid_recovery_codes WHERE user_id = ? LIMIT 1", userID).Scan(&codeID)
	if err != nil {
		t.Fatalf("get code ID failed: %v", err)
	}

	// Revoke the code
	success, err := svc.RevokeRecoveryCode(userID, codeID)
	if err != nil {
		t.Fatalf("revoke recovery code failed: %v", err)
	}
	if !success {
		t.Fatal("revoke should succeed")
	}

	// Try to revoke again (should fail)
	success, err = svc.RevokeRecoveryCode(userID, codeID)
	if err != nil {
		t.Fatalf("revoke used recovery code failed: %v", err)
	}
	if success {
		t.Fatal("revoke should fail for already used code")
	}
}

func TestGetStats(t *testing.T) {
	db := setupTestDB(t)
	svc := NewService(db, &Config{
		Enabled:         true,
		RecoveryPepper:  []byte("recovery-pepper"),
		MasterKey:       make([]byte, 32),
		EmailHashPepper: []byte("email-pepper"),
	})

	// Set non-zero master key
	for i := range svc.config.MasterKey {
		svc.config.MasterKey[i] = byte(i)
	}

	// Test stats when disabled
	disabledSvc := NewService(db, &Config{Enabled: false})
	stats, err := disabledSvc.GetStats()
	if err != nil {
		t.Fatalf("get stats failed: %v", err)
	}
	if stats["enabled"] != false {
		t.Fatal("stats should show disabled")
	}

	// Create some test data
	userID := "test-user-id"
	email := "test@example.com"
	_, err = svc.CreateOrUpdateMapping(userID, email, nil)
	if err != nil {
		t.Fatalf("create mapping failed: %v", err)
	}

	_, err = svc.GenerateRecoveryCodes(userID, 3)
	if err != nil {
		t.Fatalf("generate recovery codes failed: %v", err)
	}

	// Get stats
	stats, err = svc.GetStats()
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
	if stats["available_recovery_codes"] != 3 {
		t.Fatalf("expected 3 available codes, got %v", stats["available_recovery_codes"])
	}
}
