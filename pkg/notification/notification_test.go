// =============================================================================
// SECURE EMAIL MVP - NOTIFICATION SYSTEM TESTS
// =============================================================================
// Unit tests for the notification package.
// =============================================================================

package notification

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create test tables
	queries := []string{
		`CREATE TABLE users (
			user_id TEXT PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			phone_number TEXT
		)`,
		`CREATE TABLE emails (
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
			requires_password INTEGER DEFAULT 0,
			password_hash TEXT,
			geolocation_json TEXT,
			expires_at DATETIME,
			burn_after_read INTEGER DEFAULT 0,
			failed_attempts INTEGER DEFAULT 0,
			max_attempts INTEGER DEFAULT 3,
			self_destruct_after_attempts INTEGER DEFAULT 0,
			reply_enabled INTEGER DEFAULT 0,
			reply_requires_password INTEGER DEFAULT 1,
			allow_forwarding INTEGER DEFAULT 0,
			show_sender_metadata INTEGER DEFAULT 0,
			metadata_stripped INTEGER DEFAULT 1,
			is_honeytoken INTEGER DEFAULT 0,
			secure_link_id TEXT UNIQUE,
			link_created_at DATETIME,
			last_access_at DATETIME,
			access_count INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME,
			self_destructed INTEGER DEFAULT 0,
			allowed_city TEXT,
			allowed_country TEXT,
			geo_verification_type TEXT CHECK (geo_verification_type IN ('none', 'country', 'city', 'city_country')) DEFAULT 'none',
			geo_city TEXT,
			geo_country TEXT,
			require_mfa INTEGER DEFAULT 0,
			mfa_type TEXT CHECK (mfa_type IN ('TOTP', 'EMAIL_CODE')),
			encrypted_totp_secret TEXT,
			mfa_failed_attempts INTEGER DEFAULT 0,
			mfa_locked_until DATETIME,
			brute_force_failed_attempts INTEGER DEFAULT 0,
			brute_force_last_failed_attempt DATETIME,
			brute_force_lockout_until DATETIME,
			brute_force_max_attempts INTEGER DEFAULT 3,
			brute_force_lockout_duration_minutes INTEGER DEFAULT 15,
			is_password_protected BOOLEAN DEFAULT FALSE,
			password_salt TEXT,
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
			delivery_frequency TEXT DEFAULT 'immediate' CHECK (delivery_frequency IN ('immediate', 'daily_digest', 'first_attempt_only', 'threshold_trigger')),
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
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
		)`,
		`CREATE TABLE email_notification_preferences (
			email_id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			delivery_frequency TEXT DEFAULT 'immediate' CHECK (delivery_frequency IN ('immediate', 'daily_digest', 'first_attempt_only', 'threshold_trigger')),
			threshold_attempts INTEGER DEFAULT 3,
			rate_limit_window_minutes INTEGER DEFAULT 15,
			rate_limit_max_notifications INTEGER DEFAULT 5,
			inherit_global_settings BOOLEAN DEFAULT TRUE,
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
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (email_id) REFERENCES emails(email_id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
		)`,
		`CREATE TABLE access_events (
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
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (email_id) REFERENCES emails(email_id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
		)`,
		`CREATE TABLE notification_suppressions (
			suppression_id TEXT PRIMARY KEY,
			email_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			event_id TEXT NOT NULL,
			suppression_reason TEXT NOT NULL CHECK (suppression_reason IN ('rate_limited', 'frequency_controlled', 'threshold_not_met', 'first_attempt_only')),
			suppressed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			original_event_type TEXT NOT NULL,
			ip_address TEXT,
			user_agent TEXT,
			country TEXT,
			city TEXT,
			device_type TEXT,
			failure_reason TEXT,
			FOREIGN KEY (email_id) REFERENCES emails(email_id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
			FOREIGN KEY (event_id) REFERENCES access_events(event_id) ON DELETE CASCADE
		)`,
		`CREATE TABLE notification_rate_limits (
			rate_limit_id TEXT PRIMARY KEY,
			email_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			ip_address TEXT NOT NULL,
			notification_count INTEGER DEFAULT 1,
			window_start DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_notification_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (email_id) REFERENCES emails(email_id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
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
			FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
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
			FOREIGN KEY (delivery_id) REFERENCES daily_digest_deliveries(delivery_id) ON DELETE CASCADE,
			FOREIGN KEY (email_id) REFERENCES emails(email_id) ON DELETE CASCADE
		)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			t.Fatalf("Failed to create test table: %v", err)
		}
	}

	return db
}

func TestNotificationService_ShouldSendNotification_Immediate(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewNotificationService(db)
	ctx := context.Background()

	// Create test user and email
	_, err := db.Exec("INSERT INTO users (user_id, email) VALUES (?, ?)", "user1", "test@example.com")
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	_, err = db.Exec("INSERT INTO emails (email_id, sender_id, recipient, subject, encrypted_blob_url, encrypted_key, encryption_nonce, encryption_auth_tag, sha256_hash) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		"email1", "user1", "recipient@example.com", "Test", "url", "key", "nonce", "tag", "hash")
	if err != nil {
		t.Fatalf("Failed to insert test email: %v", err)
	}

	// Create preferences with immediate delivery
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
		CreatedAt:                 time.Now(),
		UpdatedAt:                 time.Now(),
	}

	err = service.UpdateNotificationPreferences(ctx, prefs)
	if err != nil {
		t.Fatalf("Failed to update preferences: %v", err)
	}

	// Test immediate delivery
	event := &AccessEvent{
		EventID:   "event1",
		EmailID:   "email1",
		UserID:    "user1",
		EventType: AccessEventTypeSuccess,
		IPAddress: "192.168.1.1",
		Timestamp: time.Now(),
	}

	shouldSend, reason, err := service.ShouldSendNotification(ctx, event, prefs, nil)
	if err != nil {
		t.Fatalf("ShouldSendNotification failed: %v", err)
	}

	if !shouldSend {
		t.Errorf("Expected notification to be sent for immediate delivery, but it was suppressed: %s", reason)
	}
}

func TestNotificationService_ShouldSendNotification_FirstAttemptOnly(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewNotificationService(db)
	ctx := context.Background()

	// Create test user and email
	_, err := db.Exec("INSERT INTO users (user_id, email) VALUES (?, ?)", "user1", "test@example.com")
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	_, err = db.Exec("INSERT INTO emails (email_id, sender_id, recipient, subject, encrypted_blob_url, encrypted_key, encryption_nonce, encryption_auth_tag, sha256_hash) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		"email1", "user1", "recipient@example.com", "Test", "url", "key", "nonce", "tag", "hash")
	if err != nil {
		t.Fatalf("Failed to insert test email: %v", err)
	}

	// Create preferences with first attempt only
	prefs := &NotificationPreferences{
		UserID:                    "user1",
		EmailNotifications:        true,
		SMSNotifications:          false,
		NotifyOnSuccess:           true,
		NotifyOnFailure:           true,
		NotifyOnBlocked:           true,
		IncludeGeolocation:        true,
		IncludeDeviceInfo:         true,
		DeliveryFrequency:         DeliveryFrequencyFirstAttemptOnly,
		ThresholdAttempts:         3,
		RateLimitWindowMinutes:    15,
		RateLimitMaxNotifications: 5,
		CreatedAt:                 time.Now(),
		UpdatedAt:                 time.Now(),
	}

	err = service.UpdateNotificationPreferences(ctx, prefs)
	if err != nil {
		t.Fatalf("Failed to update preferences: %v", err)
	}

	// First attempt should be sent
	event1 := &AccessEvent{
		EventID:   "event1",
		EmailID:   "email1",
		UserID:    "user1",
		EventType: AccessEventTypeSuccess,
		IPAddress: "192.168.1.1",
		Timestamp: time.Now(),
	}

	shouldSend, reason, err := service.ShouldSendNotification(ctx, event1, prefs, nil)
	if err != nil {
		t.Fatalf("ShouldSendNotification failed: %v", err)
	}

	if !shouldSend {
		t.Errorf("Expected first attempt to be sent, but it was suppressed: %s", reason)
	}

	// Record the first event
	err = service.RecordAccessEvent(ctx, event1)
	if err != nil {
		t.Fatalf("Failed to record access event: %v", err)
	}

	// Second attempt from same IP should be suppressed
	event2 := &AccessEvent{
		EventID:   "event2",
		EmailID:   "email1",
		UserID:    "user1",
		EventType: AccessEventTypeSuccess,
		IPAddress: "192.168.1.1",
		Timestamp: time.Now(),
	}

	shouldSend, reason, err = service.ShouldSendNotification(ctx, event2, prefs, nil)
	if err != nil {
		t.Fatalf("ShouldSendNotification failed: %v", err)
	}

	if shouldSend {
		t.Error("Expected second attempt to be suppressed, but it was sent")
	}

	if reason != SuppressionReasonFirstAttemptOnly {
		t.Errorf("Expected suppression reason to be first_attempt_only, got: %s", reason)
	}
}

func TestNotificationService_ShouldSendNotification_ThresholdTrigger(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewNotificationService(db)
	ctx := context.Background()

	// Create test user and email
	_, err := db.Exec("INSERT INTO users (user_id, email) VALUES (?, ?)", "user1", "test@example.com")
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	_, err = db.Exec("INSERT INTO emails (email_id, sender_id, recipient, subject, encrypted_blob_url, encrypted_key, encryption_nonce, encryption_auth_tag, sha256_hash) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		"email1", "user1", "recipient@example.com", "Test", "url", "key", "nonce", "tag", "hash")
	if err != nil {
		t.Fatalf("Failed to insert test email: %v", err)
	}

	// Create preferences with threshold trigger
	prefs := &NotificationPreferences{
		UserID:                    "user1",
		EmailNotifications:        true,
		SMSNotifications:          false,
		NotifyOnSuccess:           true,
		NotifyOnFailure:           true,
		NotifyOnBlocked:           true,
		IncludeGeolocation:        true,
		IncludeDeviceInfo:         true,
		DeliveryFrequency:         DeliveryFrequencyThresholdTrigger,
		ThresholdAttempts:         3,
		RateLimitWindowMinutes:    15,
		RateLimitMaxNotifications: 5,
		CreatedAt:                 time.Now(),
		UpdatedAt:                 time.Now(),
	}

	err = service.UpdateNotificationPreferences(ctx, prefs)
	if err != nil {
		t.Fatalf("Failed to update preferences: %v", err)
	}

	// First failure should be suppressed
	event1 := &AccessEvent{
		EventID:   "event1",
		EmailID:   "email1",
		UserID:    "user1",
		EventType: AccessEventTypeFailure,
		IPAddress: "192.168.1.1",
		Timestamp: time.Now(),
	}

	shouldSend, reason, err := service.ShouldSendNotification(ctx, event1, prefs, nil)
	if err != nil {
		t.Fatalf("ShouldSendNotification failed: %v", err)
	}

	if shouldSend {
		t.Error("Expected first failure to be suppressed, but it was sent")
	}

	if reason != SuppressionReasonThresholdNotMet {
		t.Errorf("Expected suppression reason to be threshold_not_met, got: %s", reason)
	}

	// Record first failure
	err = service.RecordAccessEvent(ctx, event1)
	if err != nil {
		t.Fatalf("Failed to record access event: %v", err)
	}

	// Second failure should be suppressed
	event2 := &AccessEvent{
		EventID:   "event2",
		EmailID:   "email1",
		UserID:    "user1",
		EventType: AccessEventTypeFailure,
		IPAddress: "192.168.1.1",
		Timestamp: time.Now(),
	}

	err = service.RecordAccessEvent(ctx, event2)
	if err != nil {
		t.Fatalf("Failed to record access event: %v", err)
	}

	// Third failure should trigger notification
	event3 := &AccessEvent{
		EventID:   "event3",
		EmailID:   "email1",
		UserID:    "user1",
		EventType: AccessEventTypeFailure,
		IPAddress: "192.168.1.1",
		Timestamp: time.Now(),
	}

	shouldSend, reason, err = service.ShouldSendNotification(ctx, event3, prefs, nil)
	if err != nil {
		t.Fatalf("ShouldSendNotification failed: %v", err)
	}

	if !shouldSend {
		t.Errorf("Expected third failure to trigger notification, but it was suppressed: %s", reason)
	}
}

func TestNotificationService_ShouldSendNotification_RateLimiting(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewNotificationService(db)
	ctx := context.Background()

	// Create test user and email
	_, err := db.Exec("INSERT INTO users (user_id, email) VALUES (?, ?)", "user1", "test@example.com")
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	_, err = db.Exec("INSERT INTO emails (email_id, sender_id, recipient, subject, encrypted_blob_url, encrypted_key, encryption_nonce, encryption_auth_tag, sha256_hash) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		"email1", "user1", "recipient@example.com", "Test", "url", "key", "nonce", "tag", "hash")
	if err != nil {
		t.Fatalf("Failed to insert test email: %v", err)
	}

	// Create preferences with rate limiting
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
		RateLimitMaxNotifications: 2, // Only allow 2 notifications
		CreatedAt:                 time.Now(),
		UpdatedAt:                 time.Now(),
	}

	err = service.UpdateNotificationPreferences(ctx, prefs)
	if err != nil {
		t.Fatalf("Failed to update preferences: %v", err)
	}

	// First notification should be sent
	event1 := &AccessEvent{
		EventID:   "event1",
		EmailID:   "email1",
		UserID:    "user1",
		EventType: AccessEventTypeSuccess,
		IPAddress: "192.168.1.1",
		Timestamp: time.Now(),
	}

	shouldSend, reason, err := service.ShouldSendNotification(ctx, event1, prefs, nil)
	if err != nil {
		t.Fatalf("ShouldSendNotification failed: %v", err)
	}

	if !shouldSend {
		t.Errorf("Expected first notification to be sent, but it was suppressed: %s", reason)
	}

	// Second notification should be sent
	event2 := &AccessEvent{
		EventID:   "event2",
		EmailID:   "email1",
		UserID:    "user1",
		EventType: AccessEventTypeSuccess,
		IPAddress: "192.168.1.1",
		Timestamp: time.Now(),
	}

	shouldSend, reason, err = service.ShouldSendNotification(ctx, event2, prefs, nil)
	if err != nil {
		t.Fatalf("ShouldSendNotification failed: %v", err)
	}

	if !shouldSend {
		t.Errorf("Expected second notification to be sent, but it was suppressed: %s", reason)
	}

	// Third notification should be rate limited
	event3 := &AccessEvent{
		EventID:   "event3",
		EmailID:   "email1",
		UserID:    "user1",
		EventType: AccessEventTypeSuccess,
		IPAddress: "192.168.1.1",
		Timestamp: time.Now(),
	}

	shouldSend, reason, err = service.ShouldSendNotification(ctx, event3, prefs, nil)
	if err != nil {
		t.Fatalf("ShouldSendNotification failed: %v", err)
	}

	if shouldSend {
		t.Error("Expected third notification to be rate limited, but it was sent")
	}

	if reason != SuppressionReasonRateLimited {
		t.Errorf("Expected suppression reason to be rate_limited, got: %s", reason)
	}
}

func TestNotificationService_EmailSpecificPreferences(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewNotificationService(db)
	ctx := context.Background()

	// Create test user and email
	_, err := db.Exec("INSERT INTO users (user_id, email) VALUES (?, ?)", "user1", "test@example.com")
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	_, err = db.Exec("INSERT INTO emails (email_id, sender_id, recipient, subject, encrypted_blob_url, encrypted_key, encryption_nonce, encryption_auth_tag, sha256_hash) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		"email1", "user1", "recipient@example.com", "Test", "url", "key", "nonce", "tag", "hash")
	if err != nil {
		t.Fatalf("Failed to insert test email: %v", err)
	}

	// Create global preferences
	globalPrefs := &NotificationPreferences{
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
		CreatedAt:                 time.Now(),
		UpdatedAt:                 time.Now(),
	}

	err = service.UpdateNotificationPreferences(ctx, globalPrefs)
	if err != nil {
		t.Fatalf("Failed to update global preferences: %v", err)
	}

	// Create email-specific preferences
	emailPrefs := &EmailNotificationPreferences{
		EmailID:                   "email1",
		UserID:                    "user1",
		DeliveryFrequency:         DeliveryFrequencyFirstAttemptOnly,
		ThresholdAttempts:         5,
		RateLimitWindowMinutes:    30,
		RateLimitMaxNotifications: 10,
		InheritGlobalSettings:     false,
		CreatedAt:                 time.Now(),
		UpdatedAt:                 time.Now(),
	}

	err = service.UpdateEmailNotificationPreferences(ctx, emailPrefs)
	if err != nil {
		t.Fatalf("Failed to update email preferences: %v", err)
	}

	// Test that email-specific preferences are used
	event := &AccessEvent{
		EventID:   "event1",
		EmailID:   "email1",
		UserID:    "user1",
		EventType: AccessEventTypeSuccess,
		IPAddress: "192.168.1.1",
		Timestamp: time.Now(),
	}

	shouldSend, reason, err := service.ShouldSendNotification(ctx, event, globalPrefs, emailPrefs)
	if err != nil {
		t.Fatalf("ShouldSendNotification failed: %v", err)
	}

	if !shouldSend {
		t.Errorf("Expected first attempt to be sent, but it was suppressed: %s", reason)
	}

	// Record first event
	err = service.RecordAccessEvent(ctx, event)
	if err != nil {
		t.Fatalf("Failed to record access event: %v", err)
	}

	// Second attempt should be suppressed due to first_attempt_only
	event2 := &AccessEvent{
		EventID:   "event2",
		EmailID:   "email1",
		UserID:    "user1",
		EventType: AccessEventTypeSuccess,
		IPAddress: "192.168.1.1",
		Timestamp: time.Now(),
	}

	shouldSend, reason, err = service.ShouldSendNotification(ctx, event2, globalPrefs, emailPrefs)
	if err != nil {
		t.Fatalf("ShouldSendNotification failed: %v", err)
	}

	if shouldSend {
		t.Error("Expected second attempt to be suppressed, but it was sent")
	}

	if reason != SuppressionReasonFirstAttemptOnly {
		t.Errorf("Expected suppression reason to be first_attempt_only, got: %s", reason)
	}
}

func TestNotificationService_RecordNotificationSuppression(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewNotificationService(db)
	ctx := context.Background()

	// Create test user and email
	_, err := db.Exec("INSERT INTO users (user_id, email) VALUES (?, ?)", "user1", "test@example.com")
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	_, err = db.Exec("INSERT INTO emails (email_id, sender_id, recipient, subject, encrypted_blob_url, encrypted_key, encryption_nonce, encryption_auth_tag, sha256_hash) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		"email1", "user1", "recipient@example.com", "Test", "url", "key", "nonce", "tag", "hash")
	if err != nil {
		t.Fatalf("Failed to insert test email: %v", err)
	}

	// Create test event
	event := &AccessEvent{
		EventID:   "event1",
		EmailID:   "email1",
		UserID:    "user1",
		EventType: AccessEventTypeSuccess,
		IPAddress: "192.168.1.1",
		Timestamp: time.Now(),
	}

	// Record the event first
	err = service.RecordAccessEvent(ctx, event)
	if err != nil {
		t.Fatalf("Failed to record access event: %v", err)
	}

	// Record suppression
	err = service.RecordNotificationSuppression(ctx, event, SuppressionReasonRateLimited)
	if err != nil {
		t.Fatalf("Failed to record notification suppression: %v", err)
	}

	// Verify suppression was recorded
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM notification_suppressions WHERE email_id = ? AND user_id = ?", "email1", "user1").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query suppressions: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 suppression record, got %d", count)
	}
}

func TestNotificationService_GetSuppressionStats(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	service := NewNotificationService(db)
	ctx := context.Background()

	// Create test user and email
	_, err := db.Exec("INSERT INTO users (user_id, email) VALUES (?, ?)", "user1", "test@example.com")
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	_, err = db.Exec("INSERT INTO emails (email_id, sender_id, recipient, subject, encrypted_blob_url, encrypted_key, encryption_nonce, encryption_auth_tag, sha256_hash) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		"email1", "user1", "recipient@example.com", "Test", "url", "key", "nonce", "tag", "hash")
	if err != nil {
		t.Fatalf("Failed to insert test email: %v", err)
	}

	// Create test events
	events := []*AccessEvent{
		{EventID: "event1", EmailID: "email1", UserID: "user1", EventType: AccessEventTypeSuccess, IPAddress: "192.168.1.1", Timestamp: time.Now()},
		{EventID: "event2", EmailID: "email1", UserID: "user1", EventType: AccessEventTypeSuccess, IPAddress: "192.168.1.1", Timestamp: time.Now()},
		{EventID: "event3", EmailID: "email1", UserID: "user1", EventType: AccessEventTypeSuccess, IPAddress: "192.168.1.1", Timestamp: time.Now()},
	}

	// Record events and suppressions
	for _, event := range events {
		err = service.RecordAccessEvent(ctx, event)
		if err != nil {
			t.Fatalf("Failed to record access event: %v", err)
		}
	}

	// Record suppressions with different reasons
	err = service.RecordNotificationSuppression(ctx, events[0], SuppressionReasonRateLimited)
	if err != nil {
		t.Fatalf("Failed to record suppression: %v", err)
	}

	err = service.RecordNotificationSuppression(ctx, events[1], SuppressionReasonRateLimited)
	if err != nil {
		t.Fatalf("Failed to record suppression: %v", err)
	}

	err = service.RecordNotificationSuppression(ctx, events[2], SuppressionReasonFrequencyControlled)
	if err != nil {
		t.Fatalf("Failed to record suppression: %v", err)
	}

	// Get suppression stats
	stats, err := service.GetSuppressionStats(ctx, "user1")
	if err != nil {
		t.Fatalf("Failed to get suppression stats: %v", err)
	}

	// Verify stats
	if stats["rate_limited"] != 2 {
		t.Errorf("Expected 2 rate_limited suppressions, got %d", stats["rate_limited"])
	}

	if stats["frequency_controlled"] != 1 {
		t.Errorf("Expected 1 frequency_controlled suppression, got %d", stats["frequency_controlled"])
	}
}
