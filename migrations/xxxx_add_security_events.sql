-- Migration: Add security_events table for comprehensive security logging
-- This table stores all security-related events for audit and forensic analysis

-- Create security_events table
CREATE TABLE IF NOT EXISTS security_events (
    id TEXT PRIMARY KEY,
    event_type TEXT NOT NULL,
    severity TEXT NOT NULL CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    user_id TEXT,
    organization_id TEXT,
    ip_address TEXT NOT NULL,
    user_agent TEXT,
    endpoint TEXT NOT NULL,
    method TEXT NOT NULL,
    details TEXT, -- JSON data for structured event details
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE SET NULL
);

-- Create indexes for performance and querying
CREATE INDEX IF NOT EXISTS idx_security_events_timestamp ON security_events(timestamp);
CREATE INDEX IF NOT EXISTS idx_security_events_type ON security_events(event_type);
CREATE INDEX IF NOT EXISTS idx_security_events_severity ON security_events(severity);
CREATE INDEX IF NOT EXISTS idx_security_events_user_id ON security_events(user_id);
CREATE INDEX IF NOT EXISTS idx_security_events_org_id ON security_events(organization_id);
CREATE INDEX IF NOT EXISTS idx_security_events_ip_address ON security_events(ip_address);
CREATE INDEX IF NOT EXISTS idx_security_events_endpoint ON security_events(endpoint);

-- Create composite indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_security_events_org_timestamp ON security_events(organization_id, timestamp);
CREATE INDEX IF NOT EXISTS idx_security_events_user_timestamp ON security_events(user_id, timestamp);
CREATE INDEX IF NOT EXISTS idx_security_events_type_severity ON security_events(event_type, severity);

-- Create a view for high-severity security events (for monitoring)
CREATE VIEW IF NOT EXISTS high_severity_security_events AS
SELECT 
    id,
    event_type,
    severity,
    user_id,
    organization_id,
    ip_address,
    endpoint,
    method,
    timestamp
FROM security_events
WHERE severity IN ('high', 'critical')
ORDER BY timestamp DESC;

-- Create a view for recent security events (last 24 hours)
CREATE VIEW IF NOT EXISTS recent_security_events AS
SELECT 
    id,
    event_type,
    severity,
    user_id,
    organization_id,
    ip_address,
    endpoint,
    method,
    timestamp
FROM security_events
WHERE timestamp >= datetime('now', '-1 day')
ORDER BY timestamp DESC;

-- Create a view for security event statistics by organization
CREATE VIEW IF NOT EXISTS security_event_stats_by_org AS
SELECT 
    organization_id,
    event_type,
    severity,
    COUNT(*) as count,
    MAX(timestamp) as last_occurrence
FROM security_events
WHERE organization_id IS NOT NULL
GROUP BY organization_id, event_type, severity
ORDER BY organization_id, count DESC;

-- Create a view for security event statistics by user
CREATE VIEW IF NOT EXISTS security_event_stats_by_user AS
SELECT 
    user_id,
    event_type,
    severity,
    COUNT(*) as count,
    MAX(timestamp) as last_occurrence
FROM security_events
WHERE user_id IS NOT NULL
GROUP BY user_id, event_type, severity
ORDER BY user_id, count DESC;

-- Create a view for suspicious IP addresses (multiple high-severity events)
CREATE VIEW IF NOT EXISTS suspicious_ip_addresses AS
SELECT 
    ip_address,
    COUNT(*) as event_count,
    COUNT(CASE WHEN severity IN ('high', 'critical') THEN 1 END) as high_severity_count,
    COUNT(DISTINCT event_type) as unique_event_types,
    MAX(timestamp) as last_activity
FROM security_events
WHERE timestamp >= datetime('now', '-1 day')
GROUP BY ip_address
HAVING high_severity_count > 5 OR event_count > 20
ORDER BY high_severity_count DESC, event_count DESC;

-- Insert sample security events for testing (optional - can be removed in production)
INSERT INTO security_events (id, event_type, severity, user_id, organization_id, ip_address, user_agent, endpoint, method, details, timestamp) VALUES
('test-sec-001', 'failed_login', 'low', NULL, NULL, '192.168.1.100', 'Mozilla/5.0', '/api/auth/login', 'POST', '{"email":"test@example.com","failed_attempts":1,"error_code":"AUTH_FAILED"}', datetime('now', '-1 hour')),
('test-sec-002', 'invalid_jwt', 'medium', 'user-123', 'org-456', '192.168.1.101', 'Mozilla/5.0', '/api/admin/organizations', 'GET', '{"jwt_token":"invalid","error_code":"JWT_INVALID"}', datetime('now', '-30 minutes')),
('test-sec-003', 'privilege_escalation', 'high', 'user-789', 'org-789', '192.168.1.102', 'Mozilla/5.0', '/api/admin/organizations', 'GET', '{"current_role":"enterprise_user","requested_role":"system_admin","error_code":"PRIVILEGE_ESCALATION"}', datetime('now', '-15 minutes'));

-- Add trigger to automatically clean up old security events (older than 90 days)
CREATE TRIGGER IF NOT EXISTS cleanup_old_security_events
AFTER INSERT ON security_events
BEGIN
    DELETE FROM security_events 
    WHERE timestamp < datetime('now', '-90 days');
END;

-- Add trigger to log when security events table is accessed (for audit)
CREATE TRIGGER IF NOT EXISTS log_security_events_access
AFTER SELECT ON security_events
BEGIN
    -- This is a placeholder for potential future audit logging
    -- In a real implementation, you might want to log access to security events
    SELECT 'Security events accessed' as audit_note;
END;
