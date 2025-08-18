package notification

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func setupDailyDigestTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create required tables
	queries := []string{
		`CREATE TABLE users (
			user_id TEXT PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE emails (
			email_id TEXT PRIMARY KEY,
			sender_id TEXT NOT NULL,
			recipient TEXT NOT NULL,
			subject TEXT NOT NULL,
			encrypted_blob_url TEXT,
			encrypted_key TEXT,
			encryption_nonce TEXT,
			encryption_auth_tag TEXT,
			sha256_hash TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (sender_id) REFERENCES users(user_id)
		)`,
		`CREATE TABLE notification_preferences (
			user_id TEXT PRIMARY KEY,
			email_notifications BOOLEAN DEFAULT TRUE,
			sms_notifications BOOLEAN DEFAULT FALSE,
			notify_on_success BOOLEAN DEFAULT TRUE,
			notify_on_failure BOOLEAN DEFAULT TRUE,
			notify_on_blocked BOOLEAN DEFAULT TRUE,
			include_geolocation BOOLEAN DEFAULT TRUE,
			include_device_info BOOLEAN DEFAULT TRUE,
			delivery_frequency TEXT DEFAULT 'immediate',
			threshold_attempts INTEGER DEFAULT 3,
			rate_limit_window_minutes INTEGER DEFAULT 15,
			rate_limit_max_notifications INTEGER DEFAULT 5,
			digest_delivery_time TEXT DEFAULT '08:00',
			digest_email_enabled BOOLEAN DEFAULT TRUE,
			digest_sms_enabled BOOLEAN DEFAULT FALSE,
			push_notifications_enabled BOOLEAN DEFAULT FALSE,
			signal_enabled BOOLEAN DEFAULT FALSE,
			matrix_enabled BOOLEAN DEFAULT FALSE,
			telegram_enabled BOOLEAN DEFAULT FALSE,
			discord_enabled BOOLEAN DEFAULT FALSE,
			push_device_token TEXT,
			signal_phone TEXT,
			matrix_user_id TEXT,
			matrix_homeserver TEXT,
			telegram_chat_id TEXT,
			discord_webhook_url TEXT,
			high_risk_channels TEXT DEFAULT 'email,sms',
			high_risk_threshold INTEGER DEFAULT 3,
			high_risk_timeout_minutes INTEGER DEFAULT 30,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE access_events (
			event_id TEXT PRIMARY KEY,
			email_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			event_type TEXT NOT NULL,
			ip_address TEXT,
			user_agent TEXT,
			country TEXT,
			city TEXT,
			device_type TEXT,
			failure_reason TEXT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (email_id) REFERENCES emails(email_id),
			FOREIGN KEY (user_id) REFERENCES users(user_id)
		)`,
		`CREATE TABLE notification_suppressions (
			suppression_id TEXT PRIMARY KEY,
			email_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			event_id TEXT NOT NULL,
			suppression_reason TEXT NOT NULL,
			suppressed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			original_event_type TEXT NOT NULL,
			ip_address TEXT,
			user_agent TEXT,
			country TEXT,
			city TEXT,
			device_type TEXT,
			failure_reason TEXT,
			FOREIGN KEY (email_id) REFERENCES emails(email_id),
			FOREIGN KEY (user_id) REFERENCES users(user_id),
			FOREIGN KEY (event_id) REFERENCES access_events(event_id)
		)`,
		`CREATE TABLE daily_digest_deliveries (
			delivery_id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			digest_date DATE NOT NULL,
			email_sent BOOLEAN DEFAULT FALSE,
			sms_sent BOOLEAN DEFAULT FALSE,
			email_sent_at DATETIME,
			sms_sent_at DATETIME,
			event_count INTEGER DEFAULT 0,
			email_count INTEGER DEFAULT 0,
			success_count INTEGER DEFAULT 0,
			failure_count INTEGER DEFAULT 0,
			blocked_count INTEGER DEFAULT 0,
			suppression_count INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(user_id),
			UNIQUE(user_id, digest_date)
		)`,
		`CREATE TABLE daily_digest_content (
			content_id TEXT PRIMARY KEY,
			delivery_id TEXT NOT NULL,
			email_id TEXT NOT NULL,
			email_subject TEXT,
			recipient TEXT,
			success_count INTEGER DEFAULT 0,
			failure_count INTEGER DEFAULT 0,
			blocked_count INTEGER DEFAULT 0,
			last_access_at DATETIME,
			last_ip_address TEXT,
			last_device_type TEXT,
			last_country TEXT,
			last_city TEXT,
			suppression_count INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (delivery_id) REFERENCES daily_digest_deliveries(delivery_id),
			FOREIGN KEY (email_id) REFERENCES emails(email_id)
		)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			t.Fatalf("Failed to create table: %v", err)
		}
	}

	return db
}

func TestGenerateDailyDigest(t *testing.T) {
	db := setupDailyDigestTestDB(t)
	defer db.Close()

	service := NewNotificationService(db)
	ctx := context.Background()

	// Create test user
	_, err := db.Exec("INSERT INTO users (user_id, email, password_hash) VALUES (?, ?, ?)", "user1", "test@example.com", "dummy_hash_for_testing")
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	// Create test email
	_, err = db.Exec("INSERT INTO emails (email_id, sender_id, recipient, subject) VALUES (?, ?, ?, ?)",
		"email1", "user1", "recipient@example.com", "Test Email")
	if err != nil {
		t.Fatalf("Failed to insert test email: %v", err)
	}

	// Set up notification preferences with daily digest
	prefs := &NotificationPreferences{
		UserID:                    "user1",
		EmailNotifications:        true,
		SMSNotifications:          false,
		NotifyOnSuccess:           true,
		NotifyOnFailure:           true,
		NotifyOnBlocked:           true,
		IncludeGeolocation:        true,
		IncludeDeviceInfo:         true,
		DeliveryFrequency:         DeliveryFrequencyDailyDigest,
		ThresholdAttempts:         3,
		RateLimitWindowMinutes:    15,
		RateLimitMaxNotifications: 5,
		DigestDeliveryTime:        "08:00",
		DigestEmailEnabled:        true,
		DigestSMSEnabled:          false,
		CreatedAt:                 time.Now(),
		UpdatedAt:                 time.Now(),
	}

	err = service.UpdateNotificationPreferences(ctx, prefs)
	if err != nil {
		t.Fatalf("Failed to update notification preferences: %v", err)
	}

	// Create test access events
	digestDate := time.Now().UTC().Truncate(24 * time.Hour)
	event1 := &AccessEvent{
		EventID:   "event1",
		EmailID:   "email1",
		UserID:    "user1",
		EventType: AccessEventTypeSuccess,
		IPAddress: "192.168.1.1",
		Timestamp: digestDate.Add(2 * time.Hour),
	}

	event2 := &AccessEvent{
		EventID:   "event2",
		EmailID:   "email1",
		UserID:    "user1",
		EventType: AccessEventTypeFailure,
		IPAddress: "192.168.1.2",
		Timestamp: digestDate.Add(4 * time.Hour),
	}

	event3 := &AccessEvent{
		EventID:   "event3",
		EmailID:   "email1",
		UserID:    "user1",
		EventType: AccessEventTypeBlocked,
		IPAddress: "192.168.1.3",
		Timestamp: digestDate.Add(6 * time.Hour),
	}

	// Record events
	err = service.RecordAccessEvent(ctx, event1)
	if err != nil {
		t.Fatalf("Failed to record event1: %v", err)
	}

	err = service.RecordAccessEvent(ctx, event2)
	if err != nil {
		t.Fatalf("Failed to record event2: %v", err)
	}

	err = service.RecordAccessEvent(ctx, event3)
	if err != nil {
		t.Fatalf("Failed to record event3: %v", err)
	}

	// Generate digest
	summary, err := service.GenerateDailyDigest(ctx, "user1", digestDate)
	if err != nil {
		t.Fatalf("Failed to generate daily digest: %v", err)
	}

	// Verify summary
	if summary.UserID != "user1" {
		t.Errorf("Expected user ID 'user1', got %s", summary.UserID)
	}

	if summary.TotalEvents != 3 {
		t.Errorf("Expected 3 total events, got %d", summary.TotalEvents)
	}

	if summary.TotalEmails != 1 {
		t.Errorf("Expected 1 total email, got %d", summary.TotalEmails)
	}

	if summary.SuccessCount != 1 {
		t.Errorf("Expected 1 success event, got %d", summary.SuccessCount)
	}

	if summary.FailureCount != 1 {
		t.Errorf("Expected 1 failure event, got %d", summary.FailureCount)
	}

	if summary.BlockedCount != 1 {
		t.Errorf("Expected 1 blocked event, got %d", summary.BlockedCount)
	}

	if len(summary.EmailSummaries) != 1 {
		t.Errorf("Expected 1 email summary, got %d", len(summary.EmailSummaries))
	}

	emailSummary := summary.EmailSummaries[0]
	if emailSummary.EmailID != "email1" {
		t.Errorf("Expected email ID 'email1', got %s", emailSummary.EmailID)
	}

	if emailSummary.SuccessCount != 1 {
		t.Errorf("Expected 1 success event in email summary, got %d", emailSummary.SuccessCount)
	}

	if emailSummary.FailureCount != 1 {
		t.Errorf("Expected 1 failure event in email summary, got %d", emailSummary.FailureCount)
	}

	if emailSummary.BlockedCount != 1 {
		t.Errorf("Expected 1 blocked event in email summary, got %d", emailSummary.BlockedCount)
	}
}

func TestGenerateDailyDigestNoEvents(t *testing.T) {
	db := setupDailyDigestTestDB(t)
	defer db.Close()

	service := NewNotificationService(db)
	ctx := context.Background()

	// Create test user
	_, err := db.Exec("INSERT INTO users (user_id, email, password_hash) VALUES (?, ?, ?)", "user1", "test@example.com", "dummy_hash_for_testing")
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	// Set up notification preferences with daily digest
	prefs := &NotificationPreferences{
		UserID:                    "user1",
		EmailNotifications:        true,
		SMSNotifications:          false,
		NotifyOnSuccess:           true,
		NotifyOnFailure:           true,
		NotifyOnBlocked:           true,
		IncludeGeolocation:        true,
		IncludeDeviceInfo:         true,
		DeliveryFrequency:         DeliveryFrequencyDailyDigest,
		ThresholdAttempts:         3,
		RateLimitWindowMinutes:    15,
		RateLimitMaxNotifications: 5,
		DigestDeliveryTime:        "08:00",
		DigestEmailEnabled:        true,
		DigestSMSEnabled:          false,
		CreatedAt:                 time.Now(),
		UpdatedAt:                 time.Now(),
	}

	err = service.UpdateNotificationPreferences(ctx, prefs)
	if err != nil {
		t.Fatalf("Failed to update notification preferences: %v", err)
	}

	// Generate digest for a date with no events
	digestDate := time.Now().UTC().Truncate(24 * time.Hour)
	summary, err := service.GenerateDailyDigest(ctx, "user1", digestDate)
	if err != nil {
		t.Fatalf("Failed to generate daily digest: %v", err)
	}

	// Verify empty summary
	if summary.TotalEvents != 0 {
		t.Errorf("Expected 0 total events, got %d", summary.TotalEvents)
	}

	if summary.TotalEmails != 0 {
		t.Errorf("Expected 0 total emails, got %d", summary.TotalEmails)
	}

	if len(summary.EmailSummaries) != 0 {
		t.Errorf("Expected 0 email summaries, got %d", len(summary.EmailSummaries))
	}
}

func TestGenerateDailyDigestWrongFrequency(t *testing.T) {
	db := setupDailyDigestTestDB(t)
	defer db.Close()

	service := NewNotificationService(db)
	ctx := context.Background()

	// Create test user
	_, err := db.Exec("INSERT INTO users (user_id, email, password_hash) VALUES (?, ?, ?)", "user1", "test@example.com", "dummy_hash_for_testing")
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	// Set up notification preferences with immediate frequency (not daily digest)
	prefs := &NotificationPreferences{
		UserID:                    "user1",
		EmailNotifications:        true,
		SMSNotifications:          false,
		NotifyOnSuccess:           true,
		NotifyOnFailure:           true,
		NotifyOnBlocked:           true,
		IncludeGeolocation:        true,
		IncludeDeviceInfo:         true,
		DeliveryFrequency:         DeliveryFrequencyImmediate,
		ThresholdAttempts:         3,
		RateLimitWindowMinutes:    15,
		RateLimitMaxNotifications: 5,
		DigestDeliveryTime:        "08:00",
		DigestEmailEnabled:        true,
		DigestSMSEnabled:          false,
		CreatedAt:                 time.Now(),
		UpdatedAt:                 time.Now(),
	}

	err = service.UpdateNotificationPreferences(ctx, prefs)
	if err != nil {
		t.Fatalf("Failed to update notification preferences: %v", err)
	}

	// Try to generate digest for user with wrong frequency
	digestDate := time.Now().UTC().Truncate(24 * time.Hour)
	_, err = service.GenerateDailyDigest(ctx, "user1", digestDate)
	if err == nil {
		t.Error("Expected error when generating digest for user without daily digest enabled")
	}
}

func TestGetDailyDigestHistory(t *testing.T) {
	db := setupDailyDigestTestDB(t)
	defer db.Close()

	service := NewNotificationService(db)
	ctx := context.Background()

	// Create test user
	_, err := db.Exec("INSERT INTO users (user_id, email, password_hash) VALUES (?, ?, ?)", "user1", "test@example.com", "dummy_hash_for_testing")
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	// Insert test digest deliveries
	delivery1 := &DailyDigestDelivery{
		DeliveryID:       "delivery1",
		UserID:           "user1",
		DigestDate:       time.Now().UTC().AddDate(0, 0, -1),
		EmailSent:        true,
		SMSSent:          false,
		EventCount:       5,
		EmailCount:       2,
		SuccessCount:     3,
		FailureCount:     1,
		BlockedCount:     1,
		SuppressionCount: 0,
		CreatedAt:        time.Now(),
	}

	delivery2 := &DailyDigestDelivery{
		DeliveryID:       "delivery2",
		UserID:           "user1",
		DigestDate:       time.Now().UTC().AddDate(0, 0, -2),
		EmailSent:        true,
		SMSSent:          true,
		EventCount:       3,
		EmailCount:       1,
		SuccessCount:     2,
		FailureCount:     1,
		BlockedCount:     0,
		SuppressionCount: 1,
		CreatedAt:        time.Now(),
	}

	// Store deliveries
	err = service.storeDigestDelivery(ctx, delivery1)
	if err != nil {
		t.Fatalf("Failed to store delivery1: %v", err)
	}

	err = service.storeDigestDelivery(ctx, delivery2)
	if err != nil {
		t.Fatalf("Failed to store delivery2: %v", err)
	}

	// Get digest history
	deliveries, err := service.GetDailyDigestHistory(ctx, "user1", 10)
	if err != nil {
		t.Fatalf("Failed to get digest history: %v", err)
	}

	// Verify history
	if len(deliveries) != 2 {
		t.Errorf("Expected 2 deliveries, got %d", len(deliveries))
	}

	// Check first delivery (most recent)
	if deliveries[0].DeliveryID != "delivery1" {
		t.Errorf("Expected delivery1 as first delivery, got %s", deliveries[0].DeliveryID)
	}

	if deliveries[0].EventCount != 5 {
		t.Errorf("Expected 5 events in first delivery, got %d", deliveries[0].EventCount)
	}

	// Check second delivery
	if deliveries[1].DeliveryID != "delivery2" {
		t.Errorf("Expected delivery2 as second delivery, got %s", deliveries[1].DeliveryID)
	}

	if deliveries[1].EventCount != 3 {
		t.Errorf("Expected 3 events in second delivery, got %d", deliveries[1].EventCount)
	}
}

func TestGetUsersWithDailyDigestEnabled(t *testing.T) {
	db := setupDailyDigestTestDB(t)
	defer db.Close()

	service := NewNotificationService(db)
	ctx := context.Background()

	// Create test users
	_, err := db.Exec("INSERT INTO users (user_id, email, password_hash) VALUES (?, ?, ?)", "user1", "test1@example.com", "dummy_hash_for_testing")
	if err != nil {
		t.Fatalf("Failed to insert test user1: %v", err)
	}

	_, err = db.Exec("INSERT INTO users (user_id, email, password_hash) VALUES (?, ?, ?)", "user2", "test2@example.com", "dummy_hash_for_testing")
	if err != nil {
		t.Fatalf("Failed to insert test user2: %v", err)
	}

	_, err = db.Exec("INSERT INTO users (user_id, email, password_hash) VALUES (?, ?, ?)", "user3", "test3@example.com", "dummy_hash_for_testing")
	if err != nil {
		t.Fatalf("Failed to insert test user3: %v", err)
	}

	// Set up notification preferences
	prefs1 := &NotificationPreferences{
		UserID:             "user1",
		DeliveryFrequency:  DeliveryFrequencyDailyDigest,
		DigestDeliveryTime: "08:00",
		DigestEmailEnabled: true,
		DigestSMSEnabled:   false,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	prefs2 := &NotificationPreferences{
		UserID:             "user2",
		DeliveryFrequency:  DeliveryFrequencyImmediate,
		DigestDeliveryTime: "08:00",
		DigestEmailEnabled: true,
		DigestSMSEnabled:   false,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	prefs3 := &NotificationPreferences{
		UserID:             "user3",
		DeliveryFrequency:  DeliveryFrequencyDailyDigest,
		DigestDeliveryTime: "09:00",
		DigestEmailEnabled: true,
		DigestSMSEnabled:   true,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	// Update preferences
	err = service.UpdateNotificationPreferences(ctx, prefs1)
	if err != nil {
		t.Fatalf("Failed to update preferences1: %v", err)
	}

	err = service.UpdateNotificationPreferences(ctx, prefs2)
	if err != nil {
		t.Fatalf("Failed to update preferences2: %v", err)
	}

	err = service.UpdateNotificationPreferences(ctx, prefs3)
	if err != nil {
		t.Fatalf("Failed to update preferences3: %v", err)
	}

	// Get users with daily digest enabled
	userIDs, err := service.GetUsersWithDailyDigestEnabled(ctx)
	if err != nil {
		t.Fatalf("Failed to get users with daily digest enabled: %v", err)
	}

	// Verify results
	if len(userIDs) != 2 {
		t.Errorf("Expected 2 users with daily digest enabled, got %d", len(userIDs))
	}

	// Check that user1 and user3 are in the list, but user2 is not
	foundUser1 := false
	foundUser3 := false
	foundUser2 := false

	for _, userID := range userIDs {
		switch userID {
		case "user1":
			foundUser1 = true
		case "user2":
			foundUser2 = true
		case "user3":
			foundUser3 = true
		}
	}

	if !foundUser1 {
		t.Error("Expected to find user1 in daily digest enabled users")
	}

	if !foundUser3 {
		t.Error("Expected to find user3 in daily digest enabled users")
	}

	if foundUser2 {
		t.Error("Expected user2 to NOT be in daily digest enabled users")
	}
}
