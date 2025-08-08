package cleanup

import (
	"database/sql"
	"os"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

// TestEmailCleanupWorker_ExpiredEmailDeletion tests deletion of expired emails
func TestEmailCleanupWorker_ExpiredEmailDeletion(t *testing.T) {
	// Create temporary database
	dbPath := createTempDB(t)
	defer os.Remove(dbPath)

	// Create worker with short interval for testing
	mockR2Client := NewMockR2Client()
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	worker, err := NewEmailCleanupWorkerWithR2Client(db, mockR2Client, 1)
	if err != nil {
		t.Fatalf("Failed to create worker: %v", err)
	}
	defer worker.Stop()

	// Insert test data: expired email
	expiredTime := time.Now().Add(-1 * time.Hour) // 1 hour ago
	insertTestEmail(t, worker.db, "expired-email-1", "blob-1", &expiredTime, false, 0)

	// Run cleanup
	err = worker.RunCleanupOnce()
	if err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}

	// Verify email was marked as deleted (soft delete)
	var blobURL sql.NullString
	err = worker.db.QueryRow("SELECT encrypted_blob_url FROM emails WHERE email_id = ?", "expired-email-1").Scan(&blobURL)
	if err != nil {
		t.Fatalf("Failed to query email: %v", err)
	}

	if blobURL.Valid {
		t.Errorf("Expected email to be soft-deleted, but encrypted_blob_url is still set: %s", blobURL.String)
	}
}

// TestEmailCleanupWorker_BurnAfterReadDeletion tests deletion of burn-after-read emails
func TestEmailCleanupWorker_BurnAfterReadDeletion(t *testing.T) {
	// Create temporary database
	dbPath := createTempDB(t)
	defer os.Remove(dbPath)

	// Create worker
	mockR2Client := NewMockR2Client()
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	worker, err := NewEmailCleanupWorkerWithR2Client(db, mockR2Client, 1)
	if err != nil {
		t.Fatalf("Failed to create worker: %v", err)
	}
	defer worker.Stop()

	// Insert test data: burn-after-read email that has been accessed
	insertTestEmail(t, worker.db, "burn-email-1", "blob-2", nil, true, 1)

	// Run cleanup
	err = worker.RunCleanupOnce()
	if err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}

	// Verify email was marked as deleted
	var blobURL sql.NullString
	err = worker.db.QueryRow("SELECT encrypted_blob_url FROM emails WHERE email_id = ?", "burn-email-1").Scan(&blobURL)
	if err != nil {
		t.Fatalf("Failed to query email: %v", err)
	}

	if blobURL.Valid {
		t.Errorf("Expected burn-after-read email to be soft-deleted, but encrypted_blob_url is still set: %s", blobURL.String)
	}
}

// TestEmailCleanupWorker_NoDeletionForValidEmails tests that valid emails are not deleted
func TestEmailCleanupWorker_NoDeletionForValidEmails(t *testing.T) {
	// Create temporary database
	dbPath := createTempDB(t)
	defer os.Remove(dbPath)

	// Create worker
	mockR2Client := NewMockR2Client()
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	worker, err := NewEmailCleanupWorkerWithR2Client(db, mockR2Client, 1)
	if err != nil {
		t.Fatalf("Failed to create worker: %v", err)
	}
	defer worker.Stop()

	// Insert test data: valid emails (not expired, not burn-after-read)
	insertTestEmail(t, worker.db, "valid-email-1", "blob-3", nil, false, 0)
	insertTestEmail(t, worker.db, "valid-email-2", "blob-4", nil, false, 0)

	// Run cleanup
	err = worker.RunCleanupOnce()
	if err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}

	// Verify emails were NOT deleted
	var blobURL sql.NullString
	err = worker.db.QueryRow("SELECT encrypted_blob_url FROM emails WHERE email_id = ?", "valid-email-1").Scan(&blobURL)
	if err != nil {
		t.Fatalf("Failed to query email: %v", err)
	}
	if !blobURL.Valid || blobURL.String != "blob-3" {
		t.Errorf("Expected valid email to remain, but got: %v", blobURL)
	}

	err = worker.db.QueryRow("SELECT encrypted_blob_url FROM emails WHERE email_id = ?", "valid-email-2").Scan(&blobURL)
	if err != nil {
		t.Fatalf("Failed to query email: %v", err)
	}
	if !blobURL.Valid || blobURL.String != "blob-4" {
		t.Errorf("Expected valid email to remain, but got: %v", blobURL)
	}
}

// TestEmailCleanupWorker_BurnAfterReadNotAccessed tests that burn-after-read emails are not deleted if not accessed
func TestEmailCleanupWorker_BurnAfterReadNotAccessed(t *testing.T) {
	// Create temporary database
	dbPath := createTempDB(t)
	defer os.Remove(dbPath)

	// Create worker
	mockR2Client := NewMockR2Client()
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	worker, err := NewEmailCleanupWorkerWithR2Client(db, mockR2Client, 1)
	if err != nil {
		t.Fatalf("Failed to create worker: %v", err)
	}
	defer worker.Stop()

	// Insert test data: burn-after-read email that has NOT been accessed
	insertTestEmail(t, worker.db, "burn-email-2", "blob-5", nil, true, 0)

	// Run cleanup
	err = worker.RunCleanupOnce()
	if err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}

	// Verify email was NOT deleted (not accessed yet)
	var blobURL sql.NullString
	err = worker.db.QueryRow("SELECT encrypted_blob_url FROM emails WHERE email_id = ?", "burn-email-2").Scan(&blobURL)
	if err != nil {
		t.Fatalf("Failed to query email: %v", err)
	}

	if !blobURL.Valid || blobURL.String != "blob-5" {
		t.Errorf("Expected burn-after-read email to remain (not accessed), but got: %v", blobURL)
	}
}

// TestEmailCleanupWorker_AlreadyDeletedEmails tests that already deleted emails are not processed
func TestEmailCleanupWorker_AlreadyDeletedEmails(t *testing.T) {
	// Create temporary database
	dbPath := createTempDB(t)
	defer os.Remove(dbPath)

	// Create worker
	mockR2Client := NewMockR2Client()
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	worker, err := NewEmailCleanupWorkerWithR2Client(db, mockR2Client, 1)
	if err != nil {
		t.Fatalf("Failed to create worker: %v", err)
	}
	defer worker.Stop()

	// Insert test data: expired email that's already soft-deleted
	expiredTime := time.Now().Add(-1 * time.Hour)
	insertTestEmail(t, worker.db, "already-deleted-1", "", &expiredTime, false, 0)

	// Manually soft-delete the email
	_, err = worker.db.Exec(`
		UPDATE emails SET 
			encrypted_blob_url = NULL,
			encrypted_key = NULL,
			encryption_nonce = NULL,
			encryption_auth_tag = NULL,
			sha256_hash = NULL
		WHERE email_id = ?`,
		"already-deleted-1",
	)
	if err != nil {
		t.Fatalf("Failed to soft-delete email: %v", err)
	}

	// Run cleanup
	err = worker.RunCleanupOnce()
	if err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}

	// Verify no errors occurred (email was already deleted)
	// The cleanup should handle this gracefully
}

// TestEmailCleanupWorker_GetCleanupStats tests the statistics function
func TestEmailCleanupWorker_GetCleanupStats(t *testing.T) {
	// Create temporary database
	dbPath := createTempDB(t)
	defer os.Remove(dbPath)

	// Create worker
	mockR2Client := NewMockR2Client()
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	worker, err := NewEmailCleanupWorkerWithR2Client(db, mockR2Client, 5)
	if err != nil {
		t.Fatalf("Failed to create worker: %v", err)
	}
	defer worker.Stop()

	// Insert test data
	expiredTime := time.Now().Add(-1 * time.Hour)
	insertTestEmail(t, worker.db, "expired-1", "blob-6", &expiredTime, false, 0)
	insertTestEmail(t, worker.db, "expired-2", "blob-7", &expiredTime, false, 0)
	insertTestEmail(t, worker.db, "burn-1", "blob-8", nil, true, 1)
	insertTestEmail(t, worker.db, "valid-1", "blob-9", nil, false, 0)

	// Get stats
	stats, err := worker.GetCleanupStats()
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}

	// Verify stats
	if stats["expired_emails"] != 2 {
		t.Errorf("Expected 2 expired emails, got %v", stats["expired_emails"])
	}
	if stats["burn_after_read_emails"] != 1 {
		t.Errorf("Expected 1 burn-after-read email, got %v", stats["burn_after_read_emails"])
	}
	if stats["total_emails_with_content"] != 4 {
		t.Errorf("Expected 4 total emails with content, got %v", stats["total_emails_with_content"])
	}
	if stats["cleanup_interval_minutes"] != 5 {
		t.Errorf("Expected cleanup interval of 5 minutes, got %v", stats["cleanup_interval_minutes"])
	}
}

// Helper functions

// createTempDB creates a temporary database with the emails schema
func createTempDB(t *testing.T) string {
	// Create temporary file
	tmpFile, err := os.CreateTemp("", "test_cleanup_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()

	dbPath := tmpFile.Name()

	// Open database and create schema
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create emails table
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
			compression_algo TEXT DEFAULT 'gzip',
			sha256_hash TEXT,
			expires_at DATETIME,
			burn_after_read INTEGER DEFAULT 0,
			access_count INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create emails table: %v", err)
	}

	return dbPath
}

// insertTestEmail inserts a test email into the database
func insertTestEmail(t *testing.T, db *sql.DB, emailID, blobID string, expiresAt *time.Time, burnAfterRead bool, accessCount int) {
	burnAfterReadInt := 0
	if burnAfterRead {
		burnAfterReadInt = 1
	}

	var expiresAtStr interface{}
	if expiresAt != nil {
		expiresAtStr = expiresAt.UTC().Format("2006-01-02 15:04:05")
	} else {
		expiresAtStr = nil
	}

	_, err := db.Exec(`
		INSERT INTO emails (
			email_id, sender_id, recipient, subject, encrypted_blob_url,
			encrypted_key, encryption_nonce, encryption_auth_tag, sha256_hash,
			expires_at, burn_after_read, access_count
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		emailID, "test-sender", "test@example.com", "Test Subject", blobID,
		"test-key", "test-nonce", "test-auth-tag", "test-hash",
		expiresAtStr, burnAfterReadInt, accessCount,
	)
	if err != nil {
		t.Fatalf("Failed to insert test email: %v", err)
	}
}
