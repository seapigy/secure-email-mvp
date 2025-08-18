package models

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func setupComplianceTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create organizations table
	_, err = db.Exec(`
		CREATE TABLE organizations (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create organizations table: %v", err)
	}

	// Create users table
	_, err = db.Exec(`
		CREATE TABLE users (
			id TEXT PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			totp_secret TEXT NOT NULL,
			organization_id TEXT,
			role TEXT CHECK (role IN ('system_admin', 'enterprise_admin', 'enterprise_user')) DEFAULT 'enterprise_user',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE SET NULL
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	// Create organization_compliance_logs table
	_, err = db.Exec(`
		CREATE TABLE organization_compliance_logs (
			id TEXT PRIMARY KEY,
			organization_id TEXT NOT NULL,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			action TEXT NOT NULL,
			details TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create organization_compliance_logs table: %v", err)
	}

	// Create indexes
	_, err = db.Exec(`CREATE INDEX idx_compliance_logs_org_id ON organization_compliance_logs(organization_id)`)
	if err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}

	_, err = db.Exec(`CREATE INDEX idx_compliance_logs_timestamp ON organization_compliance_logs(timestamp)`)
	if err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}

	_, err = db.Exec(`CREATE INDEX idx_compliance_logs_action ON organization_compliance_logs(action)`)
	if err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}

	// Create compliance summary view
	_, err = db.Exec(`
		CREATE VIEW organization_compliance_summary AS
		SELECT 
			o.id as organization_id,
			o.name as organization_name,
			COUNT(DISTINCT u.id) as total_users,
			COUNT(CASE WHEN cl.action = 'policy_violation' THEN 1 END) as policy_violations,
			COUNT(CASE WHEN cl.action = 'user_data_retained' THEN 1 END) as data_retention_events,
			COUNT(CASE WHEN cl.action = 'export_requested' THEN 1 END) as export_requests,
			COUNT(CASE WHEN cl.action = 'access_denied' THEN 1 END) as access_denials,
			COUNT(CASE WHEN cl.action = 'data_deleted' THEN 1 END) as data_deletions,
			COUNT(CASE WHEN cl.timestamp >= datetime('now', '-30 days') THEN 1 END) as last_30d_activity,
			MAX(cl.timestamp) as last_activity_timestamp
		FROM organizations o
		LEFT JOIN users u ON o.id = u.organization_id
		LEFT JOIN organization_compliance_logs cl ON o.id = cl.organization_id
		GROUP BY o.id, o.name
	`)
	if err != nil {
		t.Fatalf("Failed to create compliance summary view: %v", err)
	}

	// Create recent activity view
	_, err = db.Exec(`
		CREATE VIEW organization_recent_compliance_activity AS
		SELECT 
			organization_id,
			action,
			COUNT(*) as count,
			MAX(timestamp) as last_occurrence
		FROM organization_compliance_logs
		WHERE timestamp >= datetime('now', '-30 days')
		GROUP BY organization_id, action
		ORDER BY last_occurrence DESC
	`)
	if err != nil {
		t.Fatalf("Failed to create recent activity view: %v", err)
	}

	return db
}

func TestLogComplianceEvent(t *testing.T) {
	db := setupComplianceTestDB(t)
	defer db.Close()

	// Create test organization
	org, err := CreateOrganization(db, "Test Organization")
	if err != nil {
		t.Fatalf("Failed to create test organization: %v", err)
	}

	// Test logging a compliance event
	action := "policy_violation"
	details := map[string]interface{}{
		"user_id":  "user123",
		"policy":   "data_retention",
		"severity": "high",
	}

	err = LogComplianceEvent(db, org.ID, action, details)
	if err != nil {
		t.Fatalf("Failed to log compliance event: %v", err)
	}

	// Verify the event was logged
	var count int
	err = db.QueryRow(`
		SELECT COUNT(*) 
		FROM organization_compliance_logs 
		WHERE organization_id = ? AND action = ?
	`, org.ID, action).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query logged event: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 logged event, got %d", count)
	}
}

func TestValidateComplianceAction(t *testing.T) {
	// Test valid actions
	validActions := []string{
		"policy_violation",
		"user_data_retained",
		"export_requested",
		"access_denied",
		"data_deleted",
		"compliance_audit",
		"data_breach",
		"retention_policy_applied",
	}

	for _, action := range validActions {
		if !ValidateComplianceAction(action) {
			t.Errorf("Expected action '%s' to be valid", action)
		}
	}

	// Test invalid actions
	invalidActions := []string{
		"invalid_action",
		"",
		"policy_violations", // plural
		"data_retention",    // missing prefix
		"export_request",    // missing suffix
	}

	for _, action := range invalidActions {
		if ValidateComplianceAction(action) {
			t.Errorf("Expected action '%s' to be invalid", action)
		}
	}
}
