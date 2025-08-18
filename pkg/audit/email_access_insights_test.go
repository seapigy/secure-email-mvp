package audit

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func createTestEmailAccessAuditor(t *testing.T) (*EmailAccessAuditor, *sql.DB) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Create email_access_logs table
	_, err = db.Exec(`
		CREATE TABLE email_access_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email_id TEXT NOT NULL,
			user_id TEXT,
			ip_address TEXT NOT NULL,
			user_agent TEXT,
			status TEXT NOT NULL,
			attempt_count INTEGER DEFAULT 1,
			result TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create email_access_logs table: %v", err)
	}

	auditor := NewEmailAccessAuditor(db, DefaultRateLimitConfig)
	return auditor, db
}

func TestGetSenderAccessInsights(t *testing.T) {
	auditor, db := createTestEmailAccessAuditor(t)
	defer db.Close()

	ctx := context.Background()

	// Insert test data
	_, err := db.Exec(`
		INSERT INTO email_access_logs (email_id, user_id, ip_address, user_agent, status, attempt_count, result, created_at)
		VALUES 
			('email1', 'user1', '192.168.1.100', 'Mozilla/5.0', 'success', 1, 'success', datetime('now', '-1 hour')),
			('email1', 'user2', '10.0.0.50', 'Chrome/90.0', 'fail', 2, 'failed_password', datetime('now', '-30 minutes')),
			('email2', 'user1', '172.16.0.25', 'Safari/14.0', 'success', 1, 'success', datetime('now', '-15 minutes'))
	`)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Test getting insights for email1
	insights, err := auditor.GetSenderAccessInsights(ctx, "email1")
	if err != nil {
		t.Fatalf("Failed to get sender access insights: %v", err)
	}

	// Verify insights
	if insights["email_id"] != "email1" {
		t.Errorf("Expected email_id to be 'email1', got %v", insights["email_id"])
	}

	if insights["total_access_count"] != 1 {
		t.Errorf("Expected total_access_count to be 1, got %v", insights["total_access_count"])
	}

	// Check that last access information is present
	if insights["last_accessed_at"] == nil {
		t.Error("Expected last_accessed_at to be present")
	}

	if insights["last_access_ip"] == nil {
		t.Error("Expected last_access_ip to be present")
	}

	if insights["last_access_result"] == nil {
		t.Error("Expected last_access_result to be present")
	}

	// Verify IP is anonymized
	lastAccessIP, ok := insights["last_access_ip"].(string)
	if !ok {
		t.Error("Expected last_access_ip to be a string")
	} else {
		// Should be anonymized (either 192.168.1.0/24 or 10.0.0.0/24)
		if lastAccessIP != "192.168.1.0/24" && lastAccessIP != "10.0.0.0/24" {
			t.Errorf("Expected anonymized IP, got %s", lastAccessIP)
		}
	}

	// Test getting insights for email with no access logs
	insights, err = auditor.GetSenderAccessInsights(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("Failed to get sender access insights for nonexistent email: %v", err)
	}

	if insights["total_access_count"] != 0 {
		t.Errorf("Expected total_access_count to be 0 for nonexistent email, got %v", insights["total_access_count"])
	}

	if insights["last_accessed_at"] != nil {
		t.Error("Expected last_accessed_at to be nil for nonexistent email")
	}
}

func TestGetAccessLogsForAdmin(t *testing.T) {
	auditor, db := createTestEmailAccessAuditor(t)
	defer db.Close()

	ctx := context.Background()

	// Insert test data
	_, err := db.Exec(`
		INSERT INTO email_access_logs (email_id, user_id, ip_address, user_agent, status, attempt_count, result, created_at)
		VALUES 
			('email1', 'user1', '192.168.1.100', 'Mozilla/5.0', 'success', 1, 'success', datetime('now', '-1 hour')),
			('email1', 'user2', '10.0.0.50', 'Chrome/90.0', 'fail', 2, 'failed_password', datetime('now', '-30 minutes')),
			('email2', 'user1', '172.16.0.25', 'Safari/14.0', 'success', 1, 'success', datetime('now', '-15 minutes')),
			('email3', 'user3', '8.8.8.8', 'Firefox/88.0', 'fail', 1, 'expired', datetime('now', '-5 minutes'))
	`)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Test filtering by email_id
	filters := map[string]string{"email_id": "email1"}
	logs, err := auditor.GetAccessLogsForAdmin(ctx, filters, 10, 0)
	if err != nil {
		t.Fatalf("Failed to get access logs for admin: %v", err)
	}

	if len(logs) != 2 {
		t.Errorf("Expected 2 logs for email1, got %d", len(logs))
	}

	// Test filtering by result
	filters = map[string]string{"result": "success"}
	logs, err = auditor.GetAccessLogsForAdmin(ctx, filters, 10, 0)
	if err != nil {
		t.Fatalf("Failed to get access logs for admin: %v", err)
	}

	if len(logs) != 2 {
		t.Errorf("Expected 2 successful logs, got %d", len(logs))
	}

	// Test filtering by user_id
	filters = map[string]string{"user_id": "user1"}
	logs, err = auditor.GetAccessLogsForAdmin(ctx, filters, 10, 0)
	if err != nil {
		t.Fatalf("Failed to get access logs for admin: %v", err)
	}

	if len(logs) != 2 {
		t.Errorf("Expected 2 logs for user1, got %d", len(logs))
	}

	// Test pagination
	logs, err = auditor.GetAccessLogsForAdmin(ctx, map[string]string{}, 2, 0)
	if err != nil {
		t.Fatalf("Failed to get access logs for admin: %v", err)
	}

	if len(logs) != 2 {
		t.Errorf("Expected 2 logs with limit 2, got %d", len(logs))
	}

	// Test offset
	logs, err = auditor.GetAccessLogsForAdmin(ctx, map[string]string{}, 2, 2)
	if err != nil {
		t.Fatalf("Failed to get access logs for admin: %v", err)
	}

	if len(logs) != 2 {
		t.Errorf("Expected 2 logs with offset 2, got %d", len(logs))
	}
}

func TestGetAccessLogsCountForAdmin(t *testing.T) {
	auditor, db := createTestEmailAccessAuditor(t)
	defer db.Close()

	ctx := context.Background()

	// Insert test data
	_, err := db.Exec(`
		INSERT INTO email_access_logs (email_id, user_id, ip_address, user_agent, status, attempt_count, result, created_at)
		VALUES 
			('email1', 'user1', '192.168.1.100', 'Mozilla/5.0', 'success', 1, 'success', datetime('now', '-1 hour')),
			('email1', 'user2', '10.0.0.50', 'Chrome/90.0', 'fail', 2, 'failed_password', datetime('now', '-30 minutes')),
			('email2', 'user1', '172.16.0.25', 'Safari/14.0', 'success', 1, 'success', datetime('now', '-15 minutes'))
	`)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Test total count
	count, err := auditor.GetAccessLogsCountForAdmin(ctx, map[string]string{})
	if err != nil {
		t.Fatalf("Failed to get access logs count: %v", err)
	}

	if count != 3 {
		t.Errorf("Expected total count to be 3, got %d", count)
	}

	// Test filtering by email_id
	filters := map[string]string{"email_id": "email1"}
	count, err = auditor.GetAccessLogsCountForAdmin(ctx, filters)
	if err != nil {
		t.Fatalf("Failed to get access logs count: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected count for email1 to be 2, got %d", count)
	}

	// Test filtering by result
	filters = map[string]string{"result": "success"}
	count, err = auditor.GetAccessLogsCountForAdmin(ctx, filters)
	if err != nil {
		t.Fatalf("Failed to get access logs count: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected count for successful logs to be 2, got %d", count)
	}
}
