// =============================================================================
// SECURE EMAIL MVP - READ RECEIPTS TESTS
// =============================================================================
// Unit tests for read receipts and expiration alerts functionality.
// Micro-Iteration 4.19: Email Read Receipt & Expiration Alerts
// =============================================================================

package readreceipts

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

func TestReadReceiptService_RecordReadEvent(t *testing.T) {
	// Setup test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Create tables
	if err := createTestTables(db); err != nil {
		t.Fatalf("Failed to create test tables: %v", err)
	}

	// Create service
	service := NewReadReceiptService(db)

	// Create test email
	emailID := uuid.New().String()
	senderID := uuid.New().String()
	_, err = db.Exec(`
		INSERT INTO emails (
			email_id, sender_id, recipient, subject, encrypted_blob_url,
			encrypted_key, encryption_nonce, encryption_auth_tag, sha256_hash
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		emailID, senderID, "test@example.com", "Test Subject", "test-blob",
		"test-key", "test-nonce", "test-auth-tag", "test-hash",
	)
	if err != nil {
		t.Fatalf("Failed to insert test email: %v", err)
	}

	// Test first read event
	userID := uuid.New().String()
	readEvent := &ReadEvent{
		EmailID:   emailID,
		UserID:    userID,
		ReadAt:    time.Now(),
		IPAddress: "192.168.1.1",
		UserAgent: "test-agent",
	}

	err = service.RecordReadEvent(context.Background(), readEvent)
	if err != nil {
		t.Fatalf("Failed to record read event: %v", err)
	}

	// Verify first read was recorded
	if !readEvent.IsFirstRead {
		t.Error("Expected IsFirstRead to be true for first read")
	}

	// Verify email was updated
	var firstReadAt *time.Time
	var readCount int
	err = db.QueryRow(`SELECT first_read_at, read_count FROM emails WHERE email_id = ?`, emailID).Scan(&firstReadAt, &readCount)
	if err != nil {
		t.Fatalf("Failed to query email: %v", err)
	}

	if firstReadAt == nil {
		t.Error("Expected first_read_at to be set")
	}

	if readCount != 1 {
		t.Errorf("Expected read_count to be 1, got %d", readCount)
	}

	// Test second read event
	readEvent2 := &ReadEvent{
		EmailID:   emailID,
		UserID:    userID,
		ReadAt:    time.Now(),
		IPAddress: "192.168.1.2",
		UserAgent: "test-agent-2",
	}

	err = service.RecordReadEvent(context.Background(), readEvent2)
	if err != nil {
		t.Fatalf("Failed to record second read event: %v", err)
	}

	// Verify second read was not marked as first
	if readEvent2.IsFirstRead {
		t.Error("Expected IsFirstRead to be false for second read")
	}

	// Verify read count was incremented
	err = db.QueryRow(`SELECT read_count FROM emails WHERE email_id = ?`, emailID).Scan(&readCount)
	if err != nil {
		t.Fatalf("Failed to query email: %v", err)
	}

	if readCount != 2 {
		t.Errorf("Expected read_count to be 2, got %d", readCount)
	}
}

func TestReadReceiptService_GetReadReceiptPreferences(t *testing.T) {
	// Setup test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Create tables
	if err := createTestTables(db); err != nil {
		t.Fatalf("Failed to create test tables: %v", err)
	}

	// Create service
	service := NewReadReceiptService(db)

	// Test getting preferences for new user (should create defaults)
	userID := uuid.New().String()
	prefs, err := service.GetReadReceiptPreferences(context.Background(), userID)
	if err != nil {
		t.Fatalf("Failed to get preferences: %v", err)
	}

	if prefs.UserID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, prefs.UserID)
	}

	if !prefs.EnableReadReceipts {
		t.Error("Expected EnableReadReceipts to be true by default")
	}

	if !prefs.EnableExpirationAlerts {
		t.Error("Expected EnableExpirationAlerts to be true by default")
	}

	if prefs.ExpirationAlertHours != 24 {
		t.Errorf("Expected ExpirationAlertHours to be 24, got %d", prefs.ExpirationAlertHours)
	}

	// Test updating preferences
	prefs.EnableReadReceipts = false
	prefs.ExpirationAlertHours = 48
	prefs.DeliveryMethods = "email"

	err = service.UpdateReadReceiptPreferences(context.Background(), prefs)
	if err != nil {
		t.Fatalf("Failed to update preferences: %v", err)
	}

	// Verify preferences were updated
	updatedPrefs, err := service.GetReadReceiptPreferences(context.Background(), userID)
	if err != nil {
		t.Fatalf("Failed to get updated preferences: %v", err)
	}

	if updatedPrefs.EnableReadReceipts {
		t.Error("Expected EnableReadReceipts to be false after update")
	}

	if updatedPrefs.ExpirationAlertHours != 48 {
		t.Errorf("Expected ExpirationAlertHours to be 48, got %d", updatedPrefs.ExpirationAlertHours)
	}

	if updatedPrefs.DeliveryMethods != "email" {
		t.Errorf("Expected DeliveryMethods to be 'email', got %s", updatedPrefs.DeliveryMethods)
	}
}

func TestReadReceiptService_GetEmailReadReceiptInfo(t *testing.T) {
	// Setup test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Create tables
	if err := createTestTables(db); err != nil {
		t.Fatalf("Failed to create test tables: %v", err)
	}

	// Create service
	service := NewReadReceiptService(db)

	// Create test email
	emailID := uuid.New().String()
	senderID := uuid.New().String()
	_, err = db.Exec(`
		INSERT INTO emails (
			email_id, sender_id, recipient, subject, encrypted_blob_url,
			encrypted_key, encryption_nonce, encryption_auth_tag, sha256_hash,
			enable_read_receipts, enable_expiration_alerts, expiration_alert_hours
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		emailID, senderID, "test@example.com", "Test Subject", "test-blob",
		"test-key", "test-nonce", "test-auth-tag", "test-hash",
		true, true, 24,
	)
	if err != nil {
		t.Fatalf("Failed to insert test email: %v", err)
	}

	// Get read receipt info
	info, err := service.GetEmailReadReceiptInfo(context.Background(), emailID)
	if err != nil {
		t.Fatalf("Failed to get read receipt info: %v", err)
	}

	if info.EmailID != emailID {
		t.Errorf("Expected email ID %s, got %s", emailID, info.EmailID)
	}

	if !info.EnableReadReceipts {
		t.Error("Expected EnableReadReceipts to be true")
	}

	if info.ReadCount != 0 {
		t.Errorf("Expected ReadCount to be 0, got %d", info.ReadCount)
	}

	if info.ReadReceiptSent {
		t.Error("Expected ReadReceiptSent to be false")
	}
}

func TestReadReceiptService_GetReadEvents(t *testing.T) {
	// Setup test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Create tables
	if err := createTestTables(db); err != nil {
		t.Fatalf("Failed to create test tables: %v", err)
	}

	// Create service
	service := NewReadReceiptService(db)

	// Create test email
	emailID := uuid.New().String()
	senderID := uuid.New().String()
	_, err = db.Exec(`
		INSERT INTO emails (
			email_id, sender_id, recipient, subject, encrypted_blob_url,
			encrypted_key, encryption_nonce, encryption_auth_tag, sha256_hash
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		emailID, senderID, "test@example.com", "Test Subject", "test-blob",
		"test-key", "test-nonce", "test-auth-tag", "test-hash",
	)
	if err != nil {
		t.Fatalf("Failed to insert test email: %v", err)
	}

	// Record multiple read events
	userID := uuid.New().String()
	for i := 0; i < 3; i++ {
		readEvent := &ReadEvent{
			EmailID:   emailID,
			UserID:    userID,
			ReadAt:    time.Now().Add(time.Duration(i) * time.Hour),
			IPAddress: "192.168.1.1",
			UserAgent: "test-agent",
		}

		err = service.RecordReadEvent(context.Background(), readEvent)
		if err != nil {
			t.Fatalf("Failed to record read event %d: %v", i, err)
		}
	}

	// Get read events
	events, err := service.GetReadEvents(context.Background(), emailID, 10)
	if err != nil {
		t.Fatalf("Failed to get read events: %v", err)
	}

	if len(events) != 3 {
		t.Errorf("Expected 3 read events, got %d", len(events))
	}

	// Verify events are ordered by timestamp (newest first)
	for i := 0; i < len(events)-1; i++ {
		if events[i].ReadAt.Before(events[i+1].ReadAt) {
			t.Error("Events should be ordered by timestamp (newest first)")
		}
	}

	// Verify first event is marked as first read
	if !events[len(events)-1].IsFirstRead {
		t.Error("Expected last event (oldest) to be marked as first read")
	}
}

// Helper function to create test tables
func createTestTables(db *sql.DB) error {
	// Create users table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			user_id TEXT PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	// Create emails table with read receipt fields
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS emails (
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
			first_read_at DATETIME,
			read_count INTEGER DEFAULT 0,
			read_receipt_sent BOOLEAN DEFAULT FALSE,
			expiration_alert_sent BOOLEAN DEFAULT FALSE,
			final_expiration_alert_sent BOOLEAN DEFAULT FALSE,
			enable_read_receipts BOOLEAN DEFAULT TRUE,
			enable_expiration_alerts BOOLEAN DEFAULT TRUE,
			expiration_alert_hours INTEGER DEFAULT 24
		)
	`)
	if err != nil {
		return err
	}

	// Create read_events table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS read_events (
			event_id TEXT PRIMARY KEY,
			email_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			read_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			ip_address TEXT,
			user_agent TEXT,
			country TEXT,
			city TEXT,
			device_type TEXT,
			is_first_read BOOLEAN DEFAULT FALSE
		)
	`)
	if err != nil {
		return err
	}

	// Create read_receipts table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS read_receipts (
			receipt_id TEXT PRIMARY KEY,
			email_id TEXT NOT NULL,
			sender_id TEXT NOT NULL,
			recipient_id TEXT NOT NULL,
			sent_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			delivery_method TEXT NOT NULL,
			delivery_status TEXT DEFAULT 'pending',
			error_message TEXT,
			metadata TEXT
		)
	`)
	if err != nil {
		return err
	}

	// Create expiration_alerts table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS expiration_alerts (
			alert_id TEXT PRIMARY KEY,
			email_id TEXT NOT NULL,
			sender_id TEXT NOT NULL,
			alert_type TEXT NOT NULL,
			sent_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			delivery_method TEXT NOT NULL,
			delivery_status TEXT DEFAULT 'pending',
			error_message TEXT,
			metadata TEXT
		)
	`)
	if err != nil {
		return err
	}

	// Create read_receipt_preferences table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS read_receipt_preferences (
			user_id TEXT PRIMARY KEY,
			enable_read_receipts BOOLEAN DEFAULT TRUE,
			enable_expiration_alerts BOOLEAN DEFAULT TRUE,
			expiration_alert_hours INTEGER DEFAULT 24,
			delivery_methods TEXT DEFAULT 'email,sms',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	return err
}
