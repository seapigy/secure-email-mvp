-- Migration: Add Organization Compliance Logs Table
-- Micro-Iteration 4.34: Enterprise Compliance Dashboards & Reporting

-- Create organization_compliance_logs table
CREATE TABLE IF NOT EXISTS organization_compliance_logs (
    id TEXT PRIMARY KEY,
    organization_id TEXT NOT NULL,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    action TEXT NOT NULL,
    details TEXT, -- JSON data for structured event details
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_compliance_logs_org_id ON organization_compliance_logs(organization_id);
CREATE INDEX IF NOT EXISTS idx_compliance_logs_timestamp ON organization_compliance_logs(timestamp);
CREATE INDEX IF NOT EXISTS idx_compliance_logs_action ON organization_compliance_logs(action);
CREATE INDEX IF NOT EXISTS idx_compliance_logs_org_action ON organization_compliance_logs(organization_id, action);

-- Create a view for compliance summaries
CREATE VIEW IF NOT EXISTS organization_compliance_summary AS
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
GROUP BY o.id, o.name;

-- Create a view for recent compliance activity (last 30 days)
CREATE VIEW IF NOT EXISTS organization_recent_compliance_activity AS
SELECT 
    organization_id,
    action,
    COUNT(*) as count,
    DATE(timestamp) as activity_date
FROM organization_compliance_logs
WHERE timestamp >= datetime('now', '-30 days')
GROUP BY organization_id, action, DATE(timestamp)
ORDER BY activity_date DESC, count DESC;

-- Insert some sample compliance logs for testing
INSERT INTO organization_compliance_logs (id, organization_id, action, details) VALUES
-- Sample logs for system-default organization
('cl-001', 'system-default', 'user_data_retained', '{"user_id": "user-001", "data_type": "email", "retention_days": 30, "reason": "legal_hold"}'),
('cl-002', 'system-default', 'policy_violation', '{"user_id": "user-002", "violation_type": "unauthorized_access", "severity": "medium", "description": "Attempted access to restricted data"}'),
('cl-003', 'system-default', 'export_requested', '{"user_id": "user-003", "export_type": "compliance_report", "data_range": "last_30_days", "approved": true}'),
('cl-004', 'system-default', 'access_denied', '{"user_id": "user-004", "resource": "admin_panel", "reason": "insufficient_permissions", "ip_address": "192.168.1.100"}'),
('cl-005', 'system-default', 'data_deleted', '{"user_id": "user-005", "data_type": "expired_email", "deletion_reason": "retention_policy", "records_count": 150}'),

-- Sample logs for a test organization (if it exists)
('cl-006', 'test-org-001', 'user_data_retained', '{"user_id": "test-user-001", "data_type": "email", "retention_days": 90, "reason": "investigation"}'),
('cl-007', 'test-org-001', 'policy_violation', '{"user_id": "test-user-002", "violation_type": "data_export", "severity": "high", "description": "Unauthorized data export attempt"}'),
('cl-008', 'test-org-001', 'export_requested', '{"user_id": "test-user-003", "export_type": "audit_log", "data_range": "last_7_days", "approved": false}'),
('cl-009', 'test-org-001', 'access_denied', '{"user_id": "test-user-004", "resource": "compliance_dashboard", "reason": "role_restriction", "ip_address": "10.0.0.50"}'),
('cl-010', 'test-org-001', 'data_deleted', '{"user_id": "test-user-005", "data_type": "temp_files", "deletion_reason": "cleanup", "records_count": 75}');

-- Create a trigger to automatically update the compliance summary view
CREATE TRIGGER IF NOT EXISTS update_compliance_summary_trigger
AFTER INSERT ON organization_compliance_logs
BEGIN
    -- The view will automatically reflect the new data
    -- This trigger ensures the view is refreshed
    SELECT 1;
END;

-- Add comments for documentation
COMMENT ON TABLE organization_compliance_logs IS 'Stores compliance-related events and activities for each organization';
COMMENT ON COLUMN organization_compliance_logs.action IS 'Type of compliance event (policy_violation, user_data_retained, export_requested, access_denied, data_deleted)';
COMMENT ON COLUMN organization_compliance_logs.details IS 'JSON-formatted details about the compliance event';
COMMENT ON VIEW organization_compliance_summary IS 'Provides aggregated compliance metrics for each organization';
COMMENT ON VIEW organization_recent_compliance_activity IS 'Shows recent compliance activity (last 30 days) by organization and action type';
