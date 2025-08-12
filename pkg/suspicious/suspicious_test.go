package suspicious

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

// createTestSuspiciousDB creates a test database with suspicious detection schema
func createTestSuspiciousDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Create users table
	_, err = db.Exec(`
		CREATE TABLE users (
			user_id TEXT PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	// Create emails table
	_, err = db.Exec(`
		CREATE TABLE emails (
			email_id TEXT PRIMARY KEY,
			sender_id TEXT NOT NULL,
			recipient TEXT NOT NULL,
			subject TEXT,
			suspicious_flag BOOLEAN DEFAULT FALSE,
			suspicious_flag_set_at DATETIME,
			suspicious_flag_cleared_at DATETIME,
			suspicious_flag_cleared_by TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (sender_id) REFERENCES users(user_id)
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create emails table: %v", err)
	}

	// Create access_events table
	_, err = db.Exec(`
		CREATE TABLE access_events (
			event_id TEXT PRIMARY KEY,
			email_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			event_type TEXT NOT NULL,
			ip_address TEXT NOT NULL,
			user_agent TEXT,
			country TEXT,
			city TEXT,
			device_type TEXT,
			failure_reason TEXT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (email_id) REFERENCES emails(email_id),
			FOREIGN KEY (user_id) REFERENCES users(user_id)
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create access_events table: %v", err)
	}

	// Create suspicious_access_events table
	_, err = db.Exec(`
		CREATE TABLE suspicious_access_events (
			detection_id TEXT PRIMARY KEY,
			email_id TEXT NOT NULL,
			detection_type TEXT NOT NULL,
			detection_rule TEXT NOT NULL,
			severity TEXT NOT NULL,
			triggered_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			resolved_at DATETIME,
			resolved_by TEXT,
			resolution_notes TEXT,
			detection_metadata TEXT,
			FOREIGN KEY (email_id) REFERENCES emails(email_id),
			FOREIGN KEY (resolved_by) REFERENCES users(user_id)
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create suspicious_access_events table: %v", err)
	}

	// Create detection_rules table
	_, err = db.Exec(`
		CREATE TABLE detection_rules (
			rule_id TEXT PRIMARY KEY,
			rule_name TEXT NOT NULL UNIQUE,
			rule_type TEXT NOT NULL,
			is_enabled BOOLEAN DEFAULT TRUE,
			threshold_value INTEGER NOT NULL,
			time_window_minutes INTEGER NOT NULL,
			severity TEXT NOT NULL,
			description TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create detection_rules table: %v", err)
	}

	// Create user_suspicious_activity_preferences table
	_, err = db.Exec(`
		CREATE TABLE user_suspicious_activity_preferences (
			user_id TEXT PRIMARY KEY,
			enable_suspicious_detection BOOLEAN DEFAULT TRUE,
			notify_on_suspicious_activity BOOLEAN DEFAULT TRUE,
			auto_flag_suspicious_emails BOOLEAN DEFAULT TRUE,
			minimum_severity_for_notification TEXT DEFAULT 'medium',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(user_id)
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create user_suspicious_activity_preferences table: %v", err)
	}

	// Insert test data
	_, err = db.Exec(`
		INSERT INTO users (user_id, email, password_hash) VALUES 
		('user1', 'test1@example.com', 'hash1'),
		('user2', 'test2@example.com', 'hash2')
	`)
	if err != nil {
		t.Fatalf("Failed to insert test users: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO emails (email_id, sender_id, recipient, subject) VALUES 
		('email1', 'user1', 'recipient1@example.com', 'Test Email 1'),
		('email2', 'user1', 'recipient2@example.com', 'Test Email 2')
	`)
	if err != nil {
		t.Fatalf("Failed to insert test emails: %v", err)
	}

	// Insert default detection rules
	_, err = db.Exec(`
		INSERT INTO detection_rules (rule_id, rule_name, rule_type, threshold_value, time_window_minutes, severity, description) VALUES 
		('rule_001', 'Multiple Failed Attempts', 'multiple_failed_attempts', 3, 5, 'high', 'Flag email if 3 or more failed access attempts within 5 minutes'),
		('rule_002', 'Unusual Geolocation', 'unusual_geolocation', 1, 60, 'medium', 'Flag email if access from geolocation not seen in previous successful accesses'),
		('rule_003', 'Rapid Multiple IPs', 'rapid_multiple_ips', 2, 10, 'high', 'Flag email if accessed from 2 or more different IPs within 10 minutes'),
		('rule_004', 'Impossible Travel', 'impossible_travel', 1, 5, 'critical', 'Flag email if access from geographically impossible locations within 5 minutes')
	`)
	if err != nil {
		t.Fatalf("Failed to insert detection rules: %v", err)
	}

	return db
}

func TestNewSuspiciousDetectionService(t *testing.T) {
	db := createTestSuspiciousDB(t)
	defer db.Close()

	service := NewSuspiciousDetectionService(db)
	if service == nil {
		t.Fatal("Expected service to be created, got nil")
	}
	if service.db != db {
		t.Fatal("Expected service to have the provided database connection")
	}
}

func TestGetUserPreferences(t *testing.T) {
	db := createTestSuspiciousDB(t)
	defer db.Close()

	service := NewSuspiciousDetectionService(db)

	// Test getting preferences for user that doesn't exist (should return defaults)
	prefs, err := service.GetUserPreferences(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("Expected no error for nonexistent user, got: %v", err)
	}
	if prefs.UserID != "nonexistent" {
		t.Fatalf("Expected user ID to be 'nonexistent', got: %s", prefs.UserID)
	}
	if !prefs.EnableSuspiciousDetection {
		t.Fatal("Expected suspicious detection to be enabled by default")
	}
	if prefs.MinimumSeverityForNotification != SeverityMedium {
		t.Fatalf("Expected minimum severity to be medium, got: %s", prefs.MinimumSeverityForNotification)
	}
}

func TestGetEnabledDetectionRules(t *testing.T) {
	db := createTestSuspiciousDB(t)
	defer db.Close()

	service := NewSuspiciousDetectionService(db)

	rules, err := service.GetEnabledDetectionRules(context.Background())
	if err != nil {
		t.Fatalf("Failed to get detection rules: %v", err)
	}

	if len(rules) != 4 {
		t.Fatalf("Expected 4 detection rules, got: %d", len(rules))
	}

	// Check that all rules are enabled
	for _, rule := range rules {
		if !rule.IsEnabled {
			t.Fatalf("Expected rule %s to be enabled", rule.RuleName)
		}
	}
}

func TestMultipleFailedAttemptsDetection(t *testing.T) {
	db := createTestSuspiciousDB(t)
	defer db.Close()

	service := NewSuspiciousDetectionService(db)

	// Insert some failed access events
	_, err := db.Exec(`
		INSERT INTO access_events (event_id, email_id, user_id, event_type, ip_address, timestamp) VALUES 
		('event1', 'email1', 'user1', 'failure', '192.168.1.1', datetime('now', '-2 minutes')),
		('event2', 'email1', 'user1', 'failure', '192.168.1.1', datetime('now', '-1 minute')),
		('event3', 'email1', 'user1', 'failure', '192.168.1.1', datetime('now'))
	`)
	if err != nil {
		t.Fatalf("Failed to insert test access events: %v", err)
	}

	// Process an access event to trigger detection
	err = service.ProcessAccessEvent(context.Background(), "email1", "user1", "192.168.1.1", "test-agent", "US", "New York", "desktop", "wrong password", "failure")
	if err != nil {
		t.Fatalf("Failed to process access event: %v", err)
	}

	// Check that suspicious flag was set
	var suspiciousFlag bool
	err = db.QueryRow("SELECT suspicious_flag FROM emails WHERE email_id = 'email1'").Scan(&suspiciousFlag)
	if err != nil {
		t.Fatalf("Failed to check suspicious flag: %v", err)
	}
	if !suspiciousFlag {
		t.Fatal("Expected suspicious flag to be set")
	}

	// Check that detection event was recorded
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM suspicious_access_events WHERE email_id = 'email1'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count detection events: %v", err)
	}
	if count == 0 {
		t.Fatal("Expected detection event to be recorded")
	}
}

func TestUnusualGeolocationDetection(t *testing.T) {
	db := createTestSuspiciousDB(t)
	defer db.Close()

	service := NewSuspiciousDetectionService(db)

	// Insert some successful access events from one location
	_, err := db.Exec(`
		INSERT INTO access_events (event_id, email_id, user_id, event_type, ip_address, country, city, timestamp) VALUES 
		('event1', 'email1', 'user1', 'success', '192.168.1.1', 'US', 'New York', datetime('now', '-2 hours')),
		('event2', 'email1', 'user1', 'success', '192.168.1.1', 'US', 'New York', datetime('now', '-1 hour'))
	`)
	if err != nil {
		t.Fatalf("Failed to insert test access events: %v", err)
	}

	// Process an access event from a different location
	err = service.ProcessAccessEvent(context.Background(), "email1", "user1", "192.168.1.2", "test-agent", "UK", "London", "desktop", "", "success")
	if err != nil {
		t.Fatalf("Failed to process access event: %v", err)
	}

	// Check that detection event was recorded
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM suspicious_access_events WHERE email_id = 'email1' AND detection_type = 'unusual_geolocation'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count detection events: %v", err)
	}
	if count == 0 {
		t.Fatal("Expected unusual geolocation detection event to be recorded")
	}
}

func TestRapidMultipleIPsDetection(t *testing.T) {
	db := createTestSuspiciousDB(t)
	defer db.Close()

	service := NewSuspiciousDetectionService(db)

	// Insert access events from different IPs within the time window
	_, err := db.Exec(`
		INSERT INTO access_events (event_id, email_id, user_id, event_type, ip_address, timestamp) VALUES 
		('event1', 'email1', 'user1', 'success', '192.168.1.1', datetime('now', '-5 minutes')),
		('event2', 'email1', 'user1', 'success', '192.168.1.2', datetime('now', '-3 minutes')),
		('event3', 'email1', 'user1', 'success', '192.168.1.3', datetime('now', '-1 minute'))
	`)
	if err != nil {
		t.Fatalf("Failed to insert test access events: %v", err)
	}

	// Process another access event to trigger detection
	err = service.ProcessAccessEvent(context.Background(), "email1", "user1", "192.168.1.4", "test-agent", "US", "New York", "desktop", "", "success")
	if err != nil {
		t.Fatalf("Failed to process access event: %v", err)
	}

	// Check that detection event was recorded
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM suspicious_access_events WHERE email_id = 'email1' AND detection_type = 'rapid_multiple_ips'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count detection events: %v", err)
	}
	if count == 0 {
		t.Fatal("Expected rapid multiple IPs detection event to be recorded")
	}
}

func TestClearSuspiciousFlag(t *testing.T) {
	db := createTestSuspiciousDB(t)
	defer db.Close()

	service := NewSuspiciousDetectionService(db)

	// Set suspicious flag
	_, err := db.Exec("UPDATE emails SET suspicious_flag = TRUE WHERE email_id = 'email1'")
	if err != nil {
		t.Fatalf("Failed to set suspicious flag: %v", err)
	}

	// Clear suspicious flag
	err = service.ClearSuspiciousFlag(context.Background(), "email1", "user1")
	if err != nil {
		t.Fatalf("Failed to clear suspicious flag: %v", err)
	}

	// Check that flag was cleared
	var suspiciousFlag bool
	err = db.QueryRow("SELECT suspicious_flag FROM emails WHERE email_id = 'email1'").Scan(&suspiciousFlag)
	if err != nil {
		t.Fatalf("Failed to check suspicious flag: %v", err)
	}
	if suspiciousFlag {
		t.Fatal("Expected suspicious flag to be cleared")
	}
}

func TestGetSuspiciousAccessEvents(t *testing.T) {
	db := createTestSuspiciousDB(t)
	defer db.Close()

	service := NewSuspiciousDetectionService(db)

	// Insert a detection event
	_, err := db.Exec(`
		INSERT INTO suspicious_access_events (detection_id, email_id, detection_type, detection_rule, severity, detection_metadata) VALUES 
		('detection1', 'email1', 'multiple_failed_attempts', 'rule_001', 'high', '{"failed_attempts": 3}')
	`)
	if err != nil {
		t.Fatalf("Failed to insert test detection event: %v", err)
	}

	// Get suspicious access events
	events, err := service.GetSuspiciousAccessEvents(context.Background(), "email1", 10)
	if err != nil {
		t.Fatalf("Failed to get suspicious access events: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("Expected 1 detection event, got: %d", len(events))
	}

	event := events[0]
	if event.EmailID != "email1" {
		t.Fatalf("Expected email ID to be 'email1', got: %s", event.EmailID)
	}
	if event.DetectionType != DetectionTypeMultipleFailedAttempts {
		t.Fatalf("Expected detection type to be 'multiple_failed_attempts', got: %s", event.DetectionType)
	}
	if event.Severity != SeverityHigh {
		t.Fatalf("Expected severity to be 'high', got: %s", event.Severity)
	}
}

func TestResolveDetectionEvent(t *testing.T) {
	db := createTestSuspiciousDB(t)
	defer db.Close()

	service := NewSuspiciousDetectionService(db)

	// Insert a detection event
	_, err := db.Exec(`
		INSERT INTO suspicious_access_events (detection_id, email_id, detection_type, detection_rule, severity) VALUES 
		('detection1', 'email1', 'multiple_failed_attempts', 'rule_001', 'high')
	`)
	if err != nil {
		t.Fatalf("Failed to insert test detection event: %v", err)
	}

	// Resolve the detection event
	err = service.ResolveDetectionEvent(context.Background(), "detection1", "user1", "False positive")
	if err != nil {
		t.Fatalf("Failed to resolve detection event: %v", err)
	}

	// Check that event was resolved
	var resolvedAt *time.Time
	var resolvedBy, resolutionNotes string
	err = db.QueryRow("SELECT resolved_at, resolved_by, resolution_notes FROM suspicious_access_events WHERE detection_id = 'detection1'").Scan(&resolvedAt, &resolvedBy, &resolutionNotes)
	if err != nil {
		t.Fatalf("Failed to check resolved event: %v", err)
	}
	if resolvedAt == nil {
		t.Fatal("Expected resolved_at to be set")
	}
	if resolvedBy != "user1" {
		t.Fatalf("Expected resolved_by to be 'user1', got: %s", resolvedBy)
	}
	if resolutionNotes != "False positive" {
		t.Fatalf("Expected resolution_notes to be 'False positive', got: %s", resolutionNotes)
	}
}
