// =============================================================================
// SECURE EMAIL MVP - NOTIFICATION SYSTEM TESTS
// =============================================================================
// Unit tests for the notification package.
// =============================================================================

package notification

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

// setupTestDB creates a test database with notification tables
func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create notification_preferences table
	_, err = db.Exec(`
		CREATE TABLE notification_preferences (
			user_id TEXT PRIMARY KEY,
			email_notifications BOOLEAN DEFAULT TRUE,
			sms_notifications BOOLEAN DEFAULT FALSE,
			notify_on_success BOOLEAN DEFAULT TRUE,
			notify_on_failure BOOLEAN DEFAULT TRUE,
			notify_on_blocked BOOLEAN DEFAULT TRUE,
			include_geolocation BOOLEAN DEFAULT TRUE,
			include_device_info BOOLEAN DEFAULT TRUE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create notification_preferences table: %v", err)
	}

	// Create access_events table
	_, err = db.Exec(`
		CREATE TABLE access_events (
			event_id TEXT PRIMARY KEY,
			email_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			event_type TEXT NOT NULL CHECK (event_type IN ('success', 'failure', 'blocked')),
			ip_address TEXT NOT NULL,
			user_agent TEXT,
			country TEXT,
			city TEXT,
			device_type TEXT,
			failure_reason TEXT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create access_events table: %v", err)
	}

	// Create users table for testing
	_, err = db.Exec(`
		CREATE TABLE users (
			user_id TEXT PRIMARY KEY,
			email TEXT NOT NULL,
			phone_number TEXT
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	// Insert test user
	_, err = db.Exec(`INSERT INTO users (user_id, email, phone_number) VALUES (?, ?, ?)`,
		"test-user-123", "test@example.com", "+1234567890")
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	return db
}

func TestNewNotificationService(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewNotificationService(db)
	if service == nil {
		t.Fatal("NewNotificationService returned nil")
	}
	if service.db != db {
		t.Fatal("NotificationService database not set correctly")
	}
}

func TestGetNotificationPreferences_Default(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewNotificationService(db)
	ctx := context.Background()

	prefs, err := service.GetNotificationPreferences(ctx, "test-user-123")
	if err != nil {
		t.Fatalf("Failed to get notification preferences: %v", err)
	}

	// Check default values
	if prefs.UserID != "test-user-123" {
		t.Errorf("Expected user ID 'test-user-123', got '%s'", prefs.UserID)
	}
	if !prefs.EmailNotifications {
		t.Error("Expected email notifications to be enabled by default")
	}
	if prefs.SMSNotifications {
		t.Error("Expected SMS notifications to be disabled by default")
	}
	if !prefs.NotifyOnSuccess {
		t.Error("Expected success notifications to be enabled by default")
	}
	if !prefs.NotifyOnFailure {
		t.Error("Expected failure notifications to be enabled by default")
	}
	if !prefs.NotifyOnBlocked {
		t.Error("Expected blocked notifications to be enabled by default")
	}
}

func TestUpdateNotificationPreferences(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewNotificationService(db)
	ctx := context.Background()

	prefs := &NotificationPreferences{
		UserID:             "test-user-123",
		EmailNotifications: false,
		SMSNotifications:   true,
		NotifyOnSuccess:    false,
		NotifyOnFailure:    true,
		NotifyOnBlocked:    true,
		IncludeGeolocation: false,
		IncludeDeviceInfo:  true,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	err := service.UpdateNotificationPreferences(ctx, prefs)
	if err != nil {
		t.Fatalf("Failed to update notification preferences: %v", err)
	}

	// Retrieve and verify
	retrievedPrefs, err := service.GetNotificationPreferences(ctx, "test-user-123")
	if err != nil {
		t.Fatalf("Failed to get updated notification preferences: %v", err)
	}

	if retrievedPrefs.EmailNotifications {
		t.Error("Expected email notifications to be disabled")
	}
	if !retrievedPrefs.SMSNotifications {
		t.Error("Expected SMS notifications to be enabled")
	}
	if retrievedPrefs.NotifyOnSuccess {
		t.Error("Expected success notifications to be disabled")
	}
	if !retrievedPrefs.NotifyOnFailure {
		t.Error("Expected failure notifications to be enabled")
	}
}

func TestRecordAccessEvent(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewNotificationService(db)
	ctx := context.Background()

	event := &AccessEvent{
		EventID:       "test-event-123",
		EmailID:       "test-email-456",
		UserID:        "test-user-123",
		EventType:     AccessEventTypeSuccess,
		IPAddress:     "192.168.1.1",
		UserAgent:     "Mozilla/5.0 (Test Browser)",
		Country:       "US",
		City:          "New York",
		DeviceType:    "Desktop",
		FailureReason: "",
		Timestamp:     time.Now(),
	}

	err := service.RecordAccessEvent(ctx, event)
	if err != nil {
		t.Fatalf("Failed to record access event: %v", err)
	}

	// Verify the event was recorded
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM access_events WHERE event_id = ?", "test-event-123").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query access events: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 access event, got %d", count)
	}
}

func TestShouldSendNotification(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewNotificationService(db)

	prefs := &NotificationPreferences{
		NotifyOnSuccess: true,
		NotifyOnFailure: false,
		NotifyOnBlocked: true,
	}

	// Test success notification
	if !service.ShouldSendNotification(prefs, AccessEventTypeSuccess) {
		t.Error("Expected success notification to be sent")
	}

	// Test failure notification
	if service.ShouldSendNotification(prefs, AccessEventTypeFailure) {
		t.Error("Expected failure notification to not be sent")
	}

	// Test blocked notification
	if !service.ShouldSendNotification(prefs, AccessEventTypeBlocked) {
		t.Error("Expected blocked notification to be sent")
	}
}

func TestGetAccessEventHistory(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewNotificationService(db)
	ctx := context.Background()

	// Record multiple events
	events := []*AccessEvent{
		{
			EventID:   "event-1",
			EmailID:   "email-1",
			UserID:    "test-user-123",
			EventType: AccessEventTypeSuccess,
			IPAddress: "192.168.1.1",
			Timestamp: time.Now().Add(-time.Hour),
		},
		{
			EventID:   "event-2",
			EmailID:   "email-2",
			UserID:    "test-user-123",
			EventType: AccessEventTypeFailure,
			IPAddress: "192.168.1.2",
			Timestamp: time.Now(),
		},
	}

	for _, event := range events {
		err := service.RecordAccessEvent(ctx, event)
		if err != nil {
			t.Fatalf("Failed to record access event: %v", err)
		}
	}

	// Get history
	history, err := service.GetAccessEventHistory(ctx, "test-user-123", 10)
	if err != nil {
		t.Fatalf("Failed to get access event history: %v", err)
	}

	if len(history) != 2 {
		t.Errorf("Expected 2 events in history, got %d", len(history))
	}

	// Verify events are ordered by timestamp (newest first)
	if history[0].EventID != "event-2" {
		t.Error("Expected newest event first")
	}
}

func TestDetectDeviceType(t *testing.T) {
	tests := []struct {
		userAgent    string
		expectedType string
	}{
		{"Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)", "Mobile"},
		{"Mozilla/5.0 (Android 10; Mobile)", "Mobile"},
		{"Mozilla/5.0 (iPad; CPU OS 14_0 like Mac OS X)", "Tablet"},
		{"Mozilla/5.0 (Windows NT 10.0; Win64; x64)", "Desktop"},
		{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15)", "Desktop"},
		{"Mozilla/5.0 (Linux; x86_64)", "Desktop"},
		{"Unknown User Agent", "Unknown"},
	}

	for _, test := range tests {
		result := DetectDeviceType(test.userAgent)
		if result != test.expectedType {
			t.Errorf("For user agent '%s', expected '%s', got '%s'",
				test.userAgent, test.expectedType, result)
		}
	}
}

func TestBuildEmailSubject(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewNotificationService(db)

	tests := []struct {
		eventType AccessEventType
		expected  string
	}{
		{AccessEventTypeSuccess, "Secure Email Accessed Successfully"},
		{AccessEventTypeFailure, "Secure Email Access Attempt Failed"},
		{AccessEventTypeBlocked, "Secure Email Access Blocked"},
	}

	for _, test := range tests {
		event := &AccessEvent{EventType: test.eventType}
		subject := service.buildEmailSubject(event)
		if subject != test.expected {
			t.Errorf("For event type %s, expected subject '%s', got '%s'",
				test.eventType, test.expected, subject)
		}
	}
}

func TestBuildEmailBody(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewNotificationService(db)

	event := &AccessEvent{
		EventType:  AccessEventTypeSuccess,
		IPAddress:  "192.168.1.1",
		Country:    "US",
		City:       "New York",
		DeviceType: "Desktop",
		Timestamp:  time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
	}

	prefs := &NotificationPreferences{
		IncludeGeolocation: true,
		IncludeDeviceInfo:  true,
	}

	body := service.buildEmailBody(event, prefs)

	// Check that body contains expected content
	if !strings.Contains(body, "success") {
		t.Error("Email body should contain event type")
	}
	if !strings.Contains(body, "US") {
		t.Error("Email body should contain country")
	}
	if !strings.Contains(body, "New York") {
		t.Error("Email body should contain city")
	}
	if !strings.Contains(body, "Desktop") {
		t.Error("Email body should contain device type")
	}
}

func TestBuildSMSMessage(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewNotificationService(db)

	event := &AccessEvent{
		EventType: AccessEventTypeSuccess,
		Country:   "US",
		Timestamp: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
	}

	prefs := &NotificationPreferences{
		IncludeGeolocation: true,
	}

	message := service.buildSMSMessage(event, prefs)

	// Check that message contains expected content
	if !strings.Contains(message, "Secure email accessed successfully") {
		t.Error("SMS message should contain event type")
	}
	if !strings.Contains(message, "12:00 UTC") {
		t.Error("SMS message should contain timestamp")
	}
	if !strings.Contains(message, "from US") {
		t.Error("SMS message should contain location")
	}
}
