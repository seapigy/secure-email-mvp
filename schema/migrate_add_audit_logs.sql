-- Migration: Add audit_logs table
-- This migration creates the audit_logs table for tracking all important system events

CREATE TABLE IF NOT EXISTS audit_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    user_id TEXT,
    action TEXT NOT NULL,
    entity TEXT NOT NULL,
    details TEXT,
    severity TEXT CHECK(severity IN ('low','medium','high','critical')) DEFAULT 'medium'
);

-- Create indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_audit_logs_timestamp ON audit_logs(timestamp);
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_entity ON audit_logs(entity);
CREATE INDEX IF NOT EXISTS idx_audit_logs_severity ON audit_logs(severity);

-- Insert some sample audit logs for testing
INSERT OR IGNORE INTO audit_logs (user_id, action, entity, details, severity, timestamp) VALUES
    ('admin@securesystem.email', 'login', 'admin', 'Admin login successful', 'low', datetime('now', '-1 hour')),
    ('user@example.com', 'dlp_scan', 'secure_link', 'Detected credit card number in content', 'high', datetime('now', '-30 minutes')),
    ('admin@securesystem.email', 'update_policy', 'system_security_policy', 'Updated password policy to require 12 characters', 'high', datetime('now', '-15 minutes')),
    ('user2@example.com', 'create_secure_link', 'secure_link', 'Created new secure email link', 'medium', datetime('now', '-10 minutes')),
    ('admin@securesystem.email', 'view_audit_logs', 'audit_logs', 'Viewed audit log entries', 'low', datetime('now', '-5 minutes'));






