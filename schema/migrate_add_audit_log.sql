-- =============================================================================
-- SECURE EMAIL MVP - ADVANCED AUDIT LOG & EXPORT MIGRATION
-- =============================================================================
-- This migration adds comprehensive audit logging support for Micro-Iteration 4.20:
-- Advanced Audit Log & Export
-- =============================================================================

-- Create audit_log table for comprehensive system event tracking
CREATE TABLE IF NOT EXISTS audit_log (
    log_id TEXT PRIMARY KEY,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    event_type TEXT NOT NULL,
    user_id TEXT, -- NULL for system events
    ip_address TEXT,
    user_agent TEXT,
    related_email_id TEXT,
    outcome TEXT NOT NULL CHECK (outcome IN ('success', 'failure', 'blocked')),
    details TEXT, -- JSON with additional event details
    severity TEXT DEFAULT 'info' CHECK (severity IN ('info', 'warning', 'error', 'critical')),
    session_id TEXT,
    request_id TEXT,
    country TEXT,
    city TEXT,
    device_type TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE SET NULL,
    FOREIGN KEY (related_email_id) REFERENCES emails(email_id) ON DELETE SET NULL
);

-- Create audit_log_retention table for configurable retention policies
CREATE TABLE IF NOT EXISTS audit_log_retention (
    retention_id TEXT PRIMARY KEY,
    event_type TEXT NOT NULL,
    retention_days INTEGER DEFAULT 90,
    auto_purge BOOLEAN DEFAULT TRUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create audit_log_exports table for tracking export requests
CREATE TABLE IF NOT EXISTS audit_log_exports (
    export_id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    export_type TEXT NOT NULL CHECK (export_type IN ('csv', 'json')),
    date_from DATETIME,
    date_to DATETIME,
    event_types TEXT, -- Comma-separated list of event types
    filters TEXT, -- JSON with additional filters
    file_path TEXT,
    file_size INTEGER,
    status TEXT DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
    error_message TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME,
    expires_at DATETIME, -- When the export file should be deleted
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

-- Create audit_log_filters table for saved user filters
CREATE TABLE IF NOT EXISTS audit_log_filters (
    filter_id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    filter_name TEXT NOT NULL,
    filter_config TEXT NOT NULL, -- JSON with filter configuration
    is_default BOOLEAN DEFAULT FALSE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

-- Add indexes for performance and querying
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

CREATE INDEX IF NOT EXISTS idx_audit_log_filters_user_id ON audit_log_filters(user_id);

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

-- Add comments to document the audit log system
-- audit_log: Main table for storing all system events with comprehensive metadata
-- audit_log_retention: Configurable retention policies for different event types
-- audit_log_exports: Tracking of export requests and generated files
-- audit_log_filters: Saved user filters for audit log queries

-- Event types that will be logged:
-- - email_creation: When emails are created
-- - email_access: Email access attempts (success/fail)
-- - email_deletion: Email deletions (manual or automatic)
-- - login_attempt: User login attempts (success/fail)
-- - api_key_use: API key usage events
-- - read_receipt: Read receipt events
-- - expiration_alert: Expiration alert events
-- - system_event: System-level events
-- - user_registration: User registration events
-- - password_change: Password change events
-- - mfa_setup: MFA setup/change events
-- - geolocation_verification: Geolocation verification events





















