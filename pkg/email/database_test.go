package email

import (
	"database/sql"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

// createTestEmailWithBlob creates a test email with a specific blob URL
func createTestEmailWithBlob(t *testing.T, db *sql.DB, emailID, senderID string, readOnce bool, blobURL string) {
	_, err := db.Exec(`
		INSERT INTO emails (
			email_id, sender_id, recipient, subject,
			encrypted_blob_url, encrypted_key, encryption_nonce, encryption_auth_tag,
			compression_algo, read_once
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, emailID, senderID, "test@example.com", "Test Email",
		blobURL, "test-key", "test-nonce", "test-auth-tag",
		"gzip", readOnce)
	if err != nil {
		t.Fatalf("Failed to create test email: %v", err)
	}
}

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create the emails table with all required columns
	_, err = db.Exec(`
		CREATE TABLE emails (
			email_id TEXT PRIMARY KEY,
			sender_id TEXT NOT NULL,
			recipient TEXT NOT NULL,
			subject TEXT,
			encrypted_blob_url TEXT,
			encrypted_key TEXT,
			encryption_nonce TEXT,
			encryption_auth_tag TEXT,
			compression_algo TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_access_at DATETIME,
			access_count INTEGER DEFAULT 0,
			not_before INTEGER,
			expires_at INTEGER,
			read_once BOOLEAN DEFAULT FALSE,
			mfa_on_open BOOLEAN DEFAULT FALSE,
			decoy_secret TEXT,
			remote_revoke BOOLEAN DEFAULT FALSE,
			strip_metadata BOOLEAN DEFAULT FALSE,
			self_destruct_threshold INTEGER DEFAULT 3,
			geo_rules_ref TEXT,
			failed_attempts INTEGER DEFAULT 0,
			read_once_consumed_at INTEGER,
			read_once_consumer_device TEXT,
			self_destruct_on_read_once BOOLEAN DEFAULT FALSE
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	return db
}

// createTestEmail creates a test email in the database
func createTestEmail(t *testing.T, db *sql.DB, emailID, senderID string, readOnce bool) {
	_, err := db.Exec(`
		INSERT INTO emails (
			email_id, sender_id, recipient, subject, 
			encrypted_blob_url, encrypted_key, encryption_nonce, encryption_auth_tag,
			compression_algo, read_once, self_destruct_on_read_once
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, emailID, senderID, "test@example.com", "Test Subject",
		"test-blob-url", "test-key", "test-nonce", "test-auth-tag",
		"gzip", readOnce, false)
	if err != nil {
		t.Fatalf("Failed to create test email: %v", err)
	}
}

// createTestEmailWithFailedAttempts creates a test email with specific failed attempts count
func createTestEmailWithFailedAttempts(t *testing.T, db *sql.DB, emailID, senderID string, readOnce bool, failedAttempts int, selfDestructThreshold int) {
	_, err := db.Exec(`
		INSERT INTO emails (
			email_id, sender_id, recipient, subject, 
			encrypted_blob_url, encrypted_key, encryption_nonce, encryption_auth_tag,
			compression_algo, read_once, self_destruct_on_read_once,
			failed_attempts, self_destruct_threshold
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, emailID, senderID, "test@example.com", "Test Subject",
		"test-blob-url", "test-key", "test-nonce", "test-auth-tag",
		"gzip", readOnce, false, failedAttempts, selfDestructThreshold)
	if err != nil {
		t.Fatalf("Failed to create test email: %v", err)
	}
}

func TestIncrementFailedAttempts(t *testing.T) {
	tests := []struct {
		name                   string
		emailID                string
		initialFailedAttempts  int
		selfDestructThreshold  int
		expectedError          bool
		expectedSelfDestruct   bool
		expectedFailedAttempts int
	}{
		{
			name:                   "Increment failed attempts below threshold",
			emailID:                "email-1",
			initialFailedAttempts:  1,
			selfDestructThreshold:  3,
			expectedError:          false,
			expectedSelfDestruct:   false,
			expectedFailedAttempts: 2,
		},
		{
			name:                   "Reach threshold and trigger self-destruct",
			emailID:                "email-2",
			initialFailedAttempts:  2,
			selfDestructThreshold:  3,
			expectedError:          true,
			expectedSelfDestruct:   true,
			expectedFailedAttempts: 3,
		},
		{
			name:                   "Use default threshold when not set",
			emailID:                "email-3",
			initialFailedAttempts:  2,
			selfDestructThreshold:  0, // Will use default of 3
			expectedError:          true,
			expectedSelfDestruct:   true,
			expectedFailedAttempts: 3,
		},
		{
			name:                   "Non-existent email",
			emailID:                "non-existent",
			initialFailedAttempts:  0,
			selfDestructThreshold:  3,
			expectedError:          true,
			expectedSelfDestruct:   false,
			expectedFailedAttempts: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh database for each test
			db := setupTestDB(t)
			defer db.Close()

			esdb := NewEmailSecurityDB(db)

			// Create test email (except for non-existent test)
			if tt.emailID != "non-existent" {
				createTestEmailWithFailedAttempts(t, db, tt.emailID, "sender-1", false, tt.initialFailedAttempts, tt.selfDestructThreshold)
			}

			// Test increment failed attempts
			err := esdb.IncrementFailedAttempts(tt.emailID)

			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}

				if tt.expectedSelfDestruct {
					if _, ok := err.(SelfDestructError); !ok {
						t.Errorf("Expected SelfDestructError, got %T", err)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}
			}

			// Check final failed attempts count
			if !tt.expectedSelfDestruct && tt.emailID != "non-existent" {
				count, err := esdb.GetFailedAttemptsCount(tt.emailID)
				if err != nil {
					t.Errorf("Failed to get failed attempts count: %v", err)
					return
				}
				if count != tt.expectedFailedAttempts {
					t.Errorf("Expected failed attempts %d, got %d", tt.expectedFailedAttempts, count)
				}
			}
		})
	}
}

func TestResetFailedAttempts(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	esdb := NewEmailSecurityDB(db)

	// Create test email with failed attempts
	createTestEmail(t, db, "email-1", "sender-1", false) // Changed to false for failed attempts test

	// Test reset failed attempts
	err := esdb.ResetFailedAttempts("email-1")
	if err != nil {
		t.Errorf("Failed to reset failed attempts: %v", err)
	}

	// Verify reset
	count, err := esdb.GetFailedAttemptsCount("email-1")
	if err != nil {
		t.Errorf("Failed to get failed attempts count: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected failed attempts 0, got %d", count)
	}
}

func TestDeleteEmailSecure(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	esdb := NewEmailSecurityDB(db)

	// Create test email
	createTestEmail(t, db, "email-1", "sender-1", false) // Changed to false for failed attempts test

	// Verify email exists
	exists, err := esdb.ValidateEmailExists("email-1")
	if err != nil {
		t.Errorf("Failed to check email existence: %v", err)
	}
	if !exists {
		t.Error("Email should exist before deletion")
	}

	// Test secure deletion
	err = esdb.DeleteEmailSecure("email-1")
	if err != nil {
		t.Errorf("Failed to delete email securely: %v", err)
	}

	// Verify email is deleted
	exists, err = esdb.ValidateEmailExists("email-1")
	if err != nil {
		t.Errorf("Failed to check email existence: %v", err)
	}
	if exists {
		t.Error("Email should not exist after deletion")
	}
}

func TestDeleteEmailSecureWithBlobURL(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	esdb := NewEmailSecurityDB(db)

	// Create test email with R2 blob URL
	emailID := "email-with-blob"
	blobURL := "test-blob-123"
	createTestEmailWithBlob(t, db, emailID, "sender-1", false, blobURL)

	// Verify email exists in DB
	exists, err := esdb.ValidateEmailExists(emailID)
	if err != nil {
		t.Errorf("Failed to check email existence: %v", err)
	}
	if !exists {
		t.Error("Email should exist before deletion")
	}

	// Test secure deletion (without R2 client, should still work)
	err = esdb.DeleteEmailSecure(emailID)
	if err != nil {
		t.Errorf("Failed to delete email securely: %v", err)
	}

	// Verify email is deleted from DB
	exists, err = esdb.ValidateEmailExists(emailID)
	if err != nil {
		t.Errorf("Failed to check email existence: %v", err)
	}
	if exists {
		t.Error("Email should not exist after deletion")
	}
}

func TestGetFailedAttemptsCount(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	esdb := NewEmailSecurityDB(db)

	// Create test email with specific failed attempts
	createTestEmail(t, db, "email-1", "sender-1", false) // Changed to false for failed attempts test

	// Test get failed attempts count
	count, err := esdb.GetFailedAttemptsCount("email-1")
	if err != nil {
		t.Errorf("Failed to get failed attempts count: %v", err)
	}
	if count != 0 { // Changed to 0 for failed attempts test
		t.Errorf("Expected failed attempts 0, got %d", count)
	}

	// Test non-existent email
	_, err = esdb.GetFailedAttemptsCount("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent email")
	}
}

func TestMarkReadOnceConsumed_SucceedsOnce(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	emailID := "test-email-1"
	senderID := "test-sender"
	createTestEmail(t, db, emailID, senderID, true)

	esdb := NewEmailSecurityDB(db)

	// First call should succeed
	consumedAt, err := esdb.MarkReadOnceConsumed(emailID, "test-device")
	if err != nil {
		t.Fatalf("Expected first call to succeed, got error: %v", err)
	}

	if consumedAt.IsZero() {
		t.Fatal("Expected non-zero consumption timestamp")
	}

	// Second call should fail with ReadOnceConsumedError
	_, err = esdb.MarkReadOnceConsumed(emailID, "test-device-2")
	if err == nil {
		t.Fatal("Expected second call to fail")
	}

	if _, ok := err.(ReadOnceConsumedError); !ok {
		t.Fatalf("Expected ReadOnceConsumedError, got: %T", err)
	}
}

func TestMarkReadOnceConsumed_Race(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	emailID := "test-email-race"
	senderID := "test-sender"
	createTestEmail(t, db, emailID, senderID, true)

	esdb := NewEmailSecurityDB(db)

	// Simulate concurrent access by marking consumed directly in database
	_, err := db.Exec(`
		UPDATE emails 
		SET read_once_consumed_at = ?, read_once_consumer_device = ?
		WHERE email_id = ? AND read_once = TRUE AND read_once_consumed_at IS NULL
	`, time.Now().Unix(), "tx-device", emailID)
	if err != nil {
		t.Fatalf("Failed to mark consumed directly: %v", err)
	}

	// Try to mark as consumed using the function - should fail
	_, err = esdb.MarkReadOnceConsumed(emailID, "outside-device")
	if err == nil {
		t.Fatal("Expected concurrent marking to fail")
	}

	if _, ok := err.(ReadOnceConsumedError); !ok {
		t.Fatalf("Expected ReadOnceConsumedError, got: %T", err)
	}
}

func TestMarkReadOnceConsumed_NonReadOnce(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	emailID := "test-email-non-readonce"
	senderID := "test-sender"
	createTestEmail(t, db, emailID, senderID, false) // read_once = false

	esdb := NewEmailSecurityDB(db)

	// Should fail for non-read-once email
	_, err := esdb.MarkReadOnceConsumed(emailID, "test-device")
	if err == nil {
		t.Fatal("Expected error for non-read-once email")
	}

	if err.Error() != "email is not configured for read-once" {
		t.Fatalf("Expected specific error message, got: %v", err)
	}
}

func TestIsReadOnceConsumed(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	emailID := "test-email-consumed"
	senderID := "test-sender"
	createTestEmail(t, db, emailID, senderID, true)

	esdb := NewEmailSecurityDB(db)

	// Initially not consumed
	isConsumed, consumedAt, err := esdb.IsReadOnceConsumed(emailID)
	if err != nil {
		t.Fatalf("Failed to check consumption status: %v", err)
	}

	if isConsumed {
		t.Fatal("Expected email to not be consumed initially")
	}

	if !consumedAt.IsZero() {
		t.Fatal("Expected zero timestamp for unconsumed email")
	}

	// Mark as consumed
	expectedConsumedAt, err := esdb.MarkReadOnceConsumed(emailID, "test-device")
	if err != nil {
		t.Fatalf("Failed to mark as consumed: %v", err)
	}

	// Check consumption status
	isConsumed, consumedAt, err = esdb.IsReadOnceConsumed(emailID)
	if err != nil {
		t.Fatalf("Failed to check consumption status: %v", err)
	}

	if !isConsumed {
		t.Fatal("Expected email to be consumed")
	}

	if consumedAt.IsZero() {
		t.Fatal("Expected non-zero timestamp for consumed email")
	}

	// Timestamps should be close (within 1 second)
	if time.Duration(consumedAt.Sub(expectedConsumedAt).Abs()) > time.Second {
		t.Fatalf("Expected timestamps to be close, got difference: %v", consumedAt.Sub(expectedConsumedAt))
	}
}

func TestGetReadOnceInfo(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	emailID := "test-email-info"
	senderID := "test-sender"
	createTestEmail(t, db, emailID, senderID, true)

	esdb := NewEmailSecurityDB(db)

	// Get initial info
	info, err := esdb.GetReadOnceInfo(emailID)
	if err != nil {
		t.Fatalf("Failed to get read-once info: %v", err)
	}

	if info.IsConsumed {
		t.Fatal("Expected email to not be consumed initially")
	}

	if info.SelfDestructOnRead {
		t.Fatal("Expected self-destruct-on-read to be false initially")
	}

	// Mark as consumed
	_, err = esdb.MarkReadOnceConsumed(emailID, "test-device")
	if err != nil {
		t.Fatalf("Failed to mark as consumed: %v", err)
	}

	// Get updated info
	info, err = esdb.GetReadOnceInfo(emailID)
	if err != nil {
		t.Fatalf("Failed to get read-once info: %v", err)
	}

	if !info.IsConsumed {
		t.Fatal("Expected email to be consumed")
	}

	if info.ConsumedAt.IsZero() {
		t.Fatal("Expected non-zero consumption timestamp")
	}

	if info.ConsumerDevice != "test-device" {
		t.Fatalf("Expected consumer device 'test-device', got: %s", info.ConsumerDevice)
	}
}

func TestGetReadOnceInfo_NonExistentEmail(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	esdb := NewEmailSecurityDB(db)

	// Should fail for non-existent email
	_, err := esdb.GetReadOnceInfo("non-existent-email")
	if err == nil {
		t.Fatal("Expected error for non-existent email")
	}

	if err.Error() != "email not found" {
		t.Fatalf("Expected specific error message, got: %v", err)
	}
}
