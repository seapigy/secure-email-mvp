// =============================================================================
// SECURE EMAIL MVP - AUDIT SERVICE TESTS
// =============================================================================
// Unit tests for the audit service functionality.
// =============================================================================

package audit

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

// createTestAuditDB creates an in-memory database with audit tables for testing
func createTestAuditDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Create audit tables
	auditSchema := `
		CREATE TABLE IF NOT EXISTS audit_log (
			log_id TEXT PRIMARY KEY,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			event_type TEXT NOT NULL,
			user_id TEXT,
			ip_address TEXT,
			user_agent TEXT,
			related_email_id TEXT,
			outcome TEXT NOT NULL CHECK (outcome IN ('success', 'failure', 'blocked')),
			details TEXT,
			severity TEXT DEFAULT 'info' CHECK (severity IN ('info', 'warning', 'error', 'critical')),
			session_id TEXT,
			request_id TEXT,
			country TEXT,
			city TEXT,
			device_type TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS audit_log_retention (
			retention_id TEXT PRIMARY KEY,
			event_type TEXT NOT NULL,
			retention_days INTEGER DEFAULT 90,
			auto_purge BOOLEAN DEFAULT TRUE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS audit_log_exports (
			export_id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			export_type TEXT NOT NULL CHECK (export_type IN ('csv', 'json')),
			date_from DATETIME,
			date_to DATETIME,
			event_types TEXT,
			filters TEXT,
			file_path TEXT,
			file_size INTEGER,
			status TEXT DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
			error_message TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			completed_at DATETIME,
			expires_at DATETIME
		);

		CREATE INDEX IF NOT EXISTS idx_audit_log_timestamp ON audit_log(timestamp);
		CREATE INDEX IF NOT EXISTS idx_audit_log_event_type ON audit_log(event_type);
		CREATE INDEX IF NOT EXISTS idx_audit_log_user_id ON audit_log(user_id);
		CREATE INDEX IF NOT EXISTS idx_audit_log_outcome ON audit_log(outcome);
		CREATE INDEX IF NOT EXISTS idx_audit_log_related_email_id ON audit_log(related_email_id);
		CREATE INDEX IF NOT EXISTS idx_audit_log_severity ON audit_log(severity);
		CREATE INDEX IF NOT EXISTS idx_audit_log_ip_address ON audit_log(ip_address);
		CREATE INDEX IF NOT EXISTS idx_audit_log_session_id ON audit_log(session_id);
		CREATE INDEX IF NOT EXISTS idx_audit_log_created_at ON audit_log(created_at);

		CREATE INDEX IF NOT EXISTS idx_audit_log_exports_user_id ON audit_log_exports(user_id);
		CREATE INDEX IF NOT EXISTS idx_audit_log_exports_status ON audit_log_exports(status);
		CREATE INDEX IF NOT EXISTS idx_audit_log_exports_created_at ON audit_log_exports(created_at);

		-- Insert default retention policies
		INSERT OR IGNORE INTO audit_log_retention (retention_id, event_type, retention_days, auto_purge) VALUES
			('retention_email_creation', 'email_creation', 365, TRUE),
			('retention_email_access', 'email_access', 90, TRUE),
			('retention_email_deletion', 'email_deletion', 365, TRUE),
			('retention_login_attempt', 'login_attempt', 90, TRUE),
			('retention_api_key_use', 'api_key_use', 90, TRUE),
			('retention_read_receipt', 'read_receipt', 90, TRUE),
			('retention_expiration_alert', 'expiration_alert', 90, TRUE),
			('retention_system_event', 'system_event', 180, TRUE);
	`

	if _, err := db.Exec(auditSchema); err != nil {
		t.Fatalf("Failed to create audit tables: %v", err)
	}

	return db
}

func TestNewAuditService(t *testing.T) {
	db := createTestAuditDB(t)
	defer db.Close()

	service := NewAuditService(db)
	if service == nil {
		t.Fatal("Expected audit service to be created")
	}
	if service.db != db {
		t.Fatal("Expected audit service to use provided database")
	}
}

func TestRecordEvent(t *testing.T) {
	db := createTestAuditDB(t)
	defer db.Close()

	service := NewAuditService(db)
	ctx := context.Background()

	// Test recording a simple event
	event := &AuditEvent{
		EventType: EventTypeEmailCreation,
		UserID:    stringPtr("user123"),
		IPAddress: stringPtr("192.168.1.1"),
		Outcome:   OutcomeSuccess,
		Severity:  SeverityInfo,
		Details: map[string]interface{}{
			"email_id": "email123",
			"subject":  "Test Email",
		},
	}

	err := service.RecordEvent(ctx, event)
	if err != nil {
		t.Fatalf("Failed to record event: %v", err)
	}

	if event.LogID == "" {
		t.Fatal("Expected LogID to be set")
	}

	// Verify event was recorded
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM audit_log WHERE log_id = ?", event.LogID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query audit log: %v", err)
	}
	if count != 1 {
		t.Fatal("Expected exactly one audit log entry")
	}
}

func TestQueryEvents(t *testing.T) {
	db := createTestAuditDB(t)
	defer db.Close()

	service := NewAuditService(db)
	ctx := context.Background()

	// Record some test events
	events := []*AuditEvent{
		{
			EventType: EventTypeEmailCreation,
			UserID:    stringPtr("user1"),
			Outcome:   OutcomeSuccess,
			Severity:  SeverityInfo,
		},
		{
			EventType: EventTypeEmailAccess,
			UserID:    stringPtr("user1"),
			Outcome:   OutcomeSuccess,
			Severity:  SeverityInfo,
		},
		{
			EventType: EventTypeLoginAttempt,
			UserID:    stringPtr("user2"),
			Outcome:   OutcomeFailure,
			Severity:  SeverityWarning,
		},
	}

	for _, event := range events {
		if err := service.RecordEvent(ctx, event); err != nil {
			t.Fatalf("Failed to record event: %v", err)
		}
	}

	// Test querying all events
	result, err := service.QueryEvents(ctx, AuditLogFilter{}, 1, 10)
	if err != nil {
		t.Fatalf("Failed to query events: %v", err)
	}

	if result.Total != 3 {
		t.Fatalf("Expected 3 events, got %d", result.Total)
	}

	if len(result.Events) != 3 {
		t.Fatalf("Expected 3 events in result, got %d", len(result.Events))
	}

	// Test filtering by user
	filter := AuditLogFilter{
		UserIDs: []string{"user1"},
	}
	result, err = service.QueryEvents(ctx, filter, 1, 10)
	if err != nil {
		t.Fatalf("Failed to query events with filter: %v", err)
	}

	if result.Total != 2 {
		t.Fatalf("Expected 2 events for user1, got %d", result.Total)
	}

	// Test filtering by event type
	filter = AuditLogFilter{
		EventTypes: []EventType{EventTypeEmailCreation},
	}
	result, err = service.QueryEvents(ctx, filter, 1, 10)
	if err != nil {
		t.Fatalf("Failed to query events with event type filter: %v", err)
	}

	if result.Total != 1 {
		t.Fatalf("Expected 1 email_creation event, got %d", result.Total)
	}

	// Test filtering by outcome
	filter = AuditLogFilter{
		Outcomes: []Outcome{OutcomeFailure},
	}
	result, err = service.QueryEvents(ctx, filter, 1, 10)
	if err != nil {
		t.Fatalf("Failed to query events with outcome filter: %v", err)
	}

	if result.Total != 1 {
		t.Fatalf("Expected 1 failure event, got %d", result.Total)
	}
}

func TestGetEventTypes(t *testing.T) {
	db := createTestAuditDB(t)
	defer db.Close()

	service := NewAuditService(db)
	ctx := context.Background()

	// Record events with different types
	eventTypes := []EventType{
		EventTypeEmailCreation,
		EventTypeEmailAccess,
		EventTypeLoginAttempt,
	}

	for _, eventType := range eventTypes {
		event := &AuditEvent{
			EventType: eventType,
			UserID:    stringPtr("user123"),
			Outcome:   OutcomeSuccess,
			Severity:  SeverityInfo,
		}
		if err := service.RecordEvent(ctx, event); err != nil {
			t.Fatalf("Failed to record event: %v", err)
		}
	}

	// Get event types
	types, err := service.GetEventTypes(ctx)
	if err != nil {
		t.Fatalf("Failed to get event types: %v", err)
	}

	if len(types) != 3 {
		t.Fatalf("Expected 3 event types, got %d", len(types))
	}

	// Verify all expected types are present
	expectedTypes := map[string]bool{
		"email_creation": true,
		"email_access":   true,
		"login_attempt":  true,
	}

	for _, eventType := range types {
		if !expectedTypes[eventType] {
			t.Fatalf("Unexpected event type: %s", eventType)
		}
	}
}

func TestGetRetentionPolicies(t *testing.T) {
	db := createTestAuditDB(t)
	defer db.Close()

	service := NewAuditService(db)
	ctx := context.Background()

	policies, err := service.GetRetentionPolicies(ctx)
	if err != nil {
		t.Fatalf("Failed to get retention policies: %v", err)
	}

	if len(policies) != 8 {
		t.Fatalf("Expected 8 retention policies, got %d", len(policies))
	}

	// Verify some expected policies
	expectedPolicies := map[string]bool{
		"email_creation": true,
		"email_access":   true,
		"login_attempt":  true,
	}

	for _, policy := range policies {
		if expectedPolicies[policy.EventType] {
			if policy.RetentionDays <= 0 {
				t.Fatalf("Expected positive retention days for %s", policy.EventType)
			}
		}
	}
}

func TestUpdateRetentionPolicy(t *testing.T) {
	db := createTestAuditDB(t)
	defer db.Close()

	service := NewAuditService(db)
	ctx := context.Background()

	// Update a retention policy
	policy := RetentionPolicy{
		RetentionID:   "retention_email_creation",
		EventType:     "email_creation",
		RetentionDays: 500,
		AutoPurge:     false,
	}

	err := service.UpdateRetentionPolicy(ctx, policy)
	if err != nil {
		t.Fatalf("Failed to update retention policy: %v", err)
	}

	// Verify the update
	var retentionDays int
	var autoPurge bool
	err = db.QueryRow("SELECT retention_days, auto_purge FROM audit_log_retention WHERE retention_id = ?", policy.RetentionID).Scan(&retentionDays, &autoPurge)
	if err != nil {
		t.Fatalf("Failed to query updated policy: %v", err)
	}

	if retentionDays != 500 {
		t.Fatalf("Expected retention days 500, got %d", retentionDays)
	}
	if autoPurge {
		t.Fatal("Expected auto_purge to be false")
	}
}

func TestGetUserEvents(t *testing.T) {
	db := createTestAuditDB(t)
	defer db.Close()

	service := NewAuditService(db)
	ctx := context.Background()

	// Record events for different users
	user1Events := []*AuditEvent{
		{
			EventType: EventTypeEmailCreation,
			UserID:    stringPtr("user1"),
			Outcome:   OutcomeSuccess,
			Severity:  SeverityInfo,
		},
		{
			EventType: EventTypeEmailAccess,
			UserID:    stringPtr("user1"),
			Outcome:   OutcomeSuccess,
			Severity:  SeverityInfo,
		},
	}

	user2Events := []*AuditEvent{
		{
			EventType: EventTypeLoginAttempt,
			UserID:    stringPtr("user2"),
			Outcome:   OutcomeFailure,
			Severity:  SeverityWarning,
		},
	}

	for _, event := range user1Events {
		if err := service.RecordEvent(ctx, event); err != nil {
			t.Fatalf("Failed to record user1 event: %v", err)
		}
	}

	for _, event := range user2Events {
		if err := service.RecordEvent(ctx, event); err != nil {
			t.Fatalf("Failed to record user2 event: %v", err)
		}
	}

	// Get user1 events
	events, err := service.GetUserEvents(ctx, "user1", 10)
	if err != nil {
		t.Fatalf("Failed to get user1 events: %v", err)
	}

	if len(events) != 2 {
		t.Fatalf("Expected 2 events for user1, got %d", len(events))
	}

	// Verify all events belong to user1
	for _, event := range events {
		if *event.UserID != "user1" {
			t.Fatalf("Expected user1 event, got user %s", *event.UserID)
		}
	}
}

func TestPurgeExpiredLogs(t *testing.T) {
	db := createTestAuditDB(t)
	defer db.Close()

	service := NewAuditService(db)
	ctx := context.Background()

	// Record an old event (should be purged)
	oldEvent := &AuditEvent{
		EventType: EventTypeEmailCreation,
		UserID:    stringPtr("user123"),
		Outcome:   OutcomeSuccess,
		Severity:  SeverityInfo,
		CreatedAt: time.Now().AddDate(0, 0, -400), // 400 days ago
	}
	if err := service.RecordEvent(ctx, oldEvent); err != nil {
		t.Fatalf("Failed to record old event: %v", err)
	}

	// Record a recent event (should not be purged)
	recentEvent := &AuditEvent{
		EventType: EventTypeEmailAccess,
		UserID:    stringPtr("user123"),
		Outcome:   OutcomeSuccess,
		Severity:  SeverityInfo,
		CreatedAt: time.Now().AddDate(0, 0, -30), // 30 days ago
	}
	if err := service.RecordEvent(ctx, recentEvent); err != nil {
		t.Fatalf("Failed to record recent event: %v", err)
	}

	// Purge expired logs
	err := service.PurgeExpiredLogs(ctx)
	if err != nil {
		t.Fatalf("Failed to purge expired logs: %v", err)
	}

	// Verify old event was purged
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM audit_log WHERE log_id = ?", oldEvent.LogID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query old event: %v", err)
	}
	if count != 0 {
		t.Fatal("Expected old event to be purged")
	}

	// Verify recent event was not purged
	err = db.QueryRow("SELECT COUNT(*) FROM audit_log WHERE log_id = ?", recentEvent.LogID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query recent event: %v", err)
	}
	if count != 1 {
		t.Fatal("Expected recent event to remain")
	}
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}





















