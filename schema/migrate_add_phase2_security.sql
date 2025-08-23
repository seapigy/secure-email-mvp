-- =============================================================================
-- MIGRATION: Add Phase 2 Security Enforcement Features
-- =============================================================================
-- This migration adds the database schema required for Phase 2 security
-- enforcement features for the Secure Link External Email Flow.
-- 
-- New Tables & Fields:
-- - Enhanced secure_links table with security settings
-- - link_password_attempts table for password protection
-- - link_geolocation_logs table for enhanced location tracking
-- - link_mfa_sessions table for external user MFA
-- - link_decoy_messages table for decoy content
-- - link_tamper_alerts table for suspicious activity detection
-- =============================================================================

-- =============================================================================
-- ENHANCED SECURE LINKS TABLE
-- =============================================================================
-- Add new security-related fields to the existing secure_links table

-- Add password protection fields
ALTER TABLE secure_links ADD COLUMN password_hash TEXT;
ALTER TABLE secure_links ADD COLUMN password_required BOOLEAN DEFAULT FALSE;
ALTER TABLE secure_links ADD COLUMN password_max_attempts INTEGER DEFAULT 3;
ALTER TABLE secure_links ADD COLUMN password_lockout_until TIMESTAMP;

-- Add time lock fields
ALTER TABLE secure_links ADD COLUMN time_lock_enabled BOOLEAN DEFAULT FALSE;
ALTER TABLE secure_links ADD COLUMN time_lock_until TIMESTAMP;
ALTER TABLE secure_links ADD COLUMN time_lock_timezone TEXT DEFAULT 'UTC';

-- Add auto-destruct and read-once fields
ALTER TABLE secure_links ADD COLUMN auto_destruct_enabled BOOLEAN DEFAULT FALSE;
ALTER TABLE secure_links ADD COLUMN read_once_enabled BOOLEAN DEFAULT FALSE;
ALTER TABLE secure_links ADD COLUMN read_once_consumed BOOLEAN DEFAULT FALSE;
ALTER TABLE secure_links ADD COLUMN auto_destruct_after_attempts INTEGER DEFAULT 5;

-- Add enhanced geolocation fields
ALTER TABLE secure_links ADD COLUMN geo_restriction_enabled BOOLEAN DEFAULT FALSE;
ALTER TABLE secure_links ADD COLUMN geo_allowed_countries TEXT; -- JSON array
ALTER TABLE secure_links ADD COLUMN geo_allowed_cities TEXT; -- JSON array
ALTER TABLE secure_links ADD COLUMN geo_blocked_countries TEXT; -- JSON array
ALTER TABLE secure_links ADD COLUMN geo_blocked_cities TEXT; -- JSON array

-- Add MFA fields
ALTER TABLE secure_links ADD COLUMN mfa_required BOOLEAN DEFAULT FALSE;
ALTER TABLE secure_links ADD COLUMN mfa_type TEXT CHECK (mfa_type IN ('totp', 'email', 'sms', 'none')) DEFAULT 'none';
ALTER TABLE secure_links ADD COLUMN mfa_secret TEXT; -- For TOTP
ALTER TABLE secure_links ADD COLUMN mfa_email TEXT; -- For email OTP
ALTER TABLE secure_links ADD COLUMN mfa_phone TEXT; -- For SMS OTP

-- Add decoy message fields
ALTER TABLE secure_links ADD COLUMN decoy_enabled BOOLEAN DEFAULT FALSE;
ALTER TABLE secure_links ADD COLUMN decoy_message TEXT;
ALTER TABLE secure_links ADD COLUMN decoy_trigger_conditions TEXT; -- JSON object

-- Add metadata stripping fields
ALTER TABLE secure_links ADD COLUMN strip_metadata_enabled BOOLEAN DEFAULT TRUE;
ALTER TABLE secure_links ADD COLUMN strip_headers BOOLEAN DEFAULT TRUE;
ALTER TABLE secure_links ADD COLUMN strip_attachments BOOLEAN DEFAULT FALSE;

-- Add tamper alert fields
ALTER TABLE secure_links ADD COLUMN tamper_alerts_enabled BOOLEAN DEFAULT TRUE;
ALTER TABLE secure_links ADD COLUMN tamper_alert_threshold INTEGER DEFAULT 3;

-- Add remote revocation fields
ALTER TABLE secure_links ADD COLUMN revoked_by TEXT;
ALTER TABLE secure_links ADD COLUMN revoked_at TIMESTAMP;
ALTER TABLE secure_links ADD COLUMN revocation_reason TEXT;

-- =============================================================================
-- LINK PASSWORD ATTEMPTS TABLE
-- =============================================================================
-- Tracks password attempts for secure links with rate limiting
CREATE TABLE IF NOT EXISTS link_password_attempts (
    id TEXT PRIMARY KEY,
    link_id TEXT NOT NULL,
    ip_address TEXT NOT NULL,
    user_agent TEXT,
    attempt_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    success BOOLEAN DEFAULT FALSE,
    attempt_number INTEGER NOT NULL,
    geolocation_data JSON,
    FOREIGN KEY (link_id) REFERENCES secure_links(link_id) ON DELETE CASCADE
);

-- =============================================================================
-- LINK GEOLOCATION LOGS TABLE
-- =============================================================================
-- Enhanced geolocation tracking for security enforcement
CREATE TABLE IF NOT EXISTS link_geolocation_logs (
    id TEXT PRIMARY KEY,
    link_id TEXT NOT NULL,
    ip_address TEXT NOT NULL,
    country TEXT,
    city TEXT,
    region TEXT,
    latitude REAL,
    longitude REAL,
    timezone TEXT,
    isp TEXT,
    access_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    access_allowed BOOLEAN DEFAULT TRUE,
    restriction_reason TEXT,
    geolocation_data JSON,
    FOREIGN KEY (link_id) REFERENCES secure_links(link_id) ON DELETE CASCADE
);

-- =============================================================================
-- LINK MFA SESSIONS TABLE
-- =============================================================================
-- Manages MFA sessions for external secure link users
CREATE TABLE IF NOT EXISTS link_mfa_sessions (
    id TEXT PRIMARY KEY,
    link_id TEXT NOT NULL,
    session_token TEXT NOT NULL,
    mfa_type TEXT NOT NULL CHECK (mfa_type IN ('totp', 'email', 'sms')),
    mfa_secret TEXT, -- For TOTP
    mfa_email TEXT, -- For email OTP
    mfa_phone TEXT, -- For SMS OTP
    otp_code TEXT,
    otp_expires_at TIMESTAMP,
    verified BOOLEAN DEFAULT FALSE,
    verified_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ip_address TEXT,
    user_agent TEXT,
    FOREIGN KEY (link_id) REFERENCES secure_links(link_id) ON DELETE CASCADE
);

-- =============================================================================
-- LINK DECOY MESSAGES TABLE
-- =============================================================================
-- Stores decoy messages and trigger conditions
CREATE TABLE IF NOT EXISTS link_decoy_messages (
    id TEXT PRIMARY KEY,
    link_id TEXT NOT NULL,
    decoy_type TEXT NOT NULL CHECK (decoy_type IN ('wrong_password', 'revoked', 'expired', 'blocked', 'generic')),
    decoy_title TEXT NOT NULL,
    decoy_content TEXT NOT NULL,
    trigger_condition TEXT NOT NULL, -- JSON object describing when to show
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (link_id) REFERENCES secure_links(link_id) ON DELETE CASCADE
);

-- =============================================================================
-- LINK TAMPER ALERTS TABLE
-- =============================================================================
-- Tracks suspicious activity and tamper attempts
CREATE TABLE IF NOT EXISTS link_tamper_alerts (
    id TEXT PRIMARY KEY,
    link_id TEXT NOT NULL,
    alert_type TEXT NOT NULL CHECK (alert_type IN (
        'multiple_failed_attempts', 'suspicious_location', 'unusual_timing',
        'multiple_ips', 'user_agent_mismatch', 'rapid_access_attempts',
        'geolocation_violation', 'password_brute_force', 'session_hijacking'
    )),
    severity TEXT NOT NULL CHECK (severity IN ('low', 'medium', 'high', 'critical')) DEFAULT 'medium',
    ip_address TEXT,
    user_agent TEXT,
    geolocation_data JSON,
    alert_details TEXT,
    triggered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    acknowledged BOOLEAN DEFAULT FALSE,
    acknowledged_by TEXT,
    acknowledged_at TIMESTAMP,
    resolved BOOLEAN DEFAULT FALSE,
    resolved_at TIMESTAMP,
    resolution_notes TEXT,
    FOREIGN KEY (link_id) REFERENCES secure_links(link_id) ON DELETE CASCADE
);

-- =============================================================================
-- LINK ACCESS SESSIONS TABLE
-- =============================================================================
-- Tracks individual access sessions for security monitoring
CREATE TABLE IF NOT EXISTS link_access_sessions (
    id TEXT PRIMARY KEY,
    link_id TEXT NOT NULL,
    session_token TEXT NOT NULL,
    ip_address TEXT NOT NULL,
    user_agent TEXT,
    geolocation_data JSON,
    session_start TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    session_end TIMESTAMP,
    session_duration_seconds INTEGER,
    access_granted BOOLEAN DEFAULT FALSE,
    access_reason TEXT, -- Why access was granted/denied
    security_checks_passed JSON, -- JSON object of passed security checks
    security_checks_failed JSON, -- JSON object of failed security checks
    mfa_session_id TEXT,
    password_attempt_id TEXT,
    FOREIGN KEY (link_id) REFERENCES secure_links(link_id) ON DELETE CASCADE,
    FOREIGN KEY (mfa_session_id) REFERENCES link_mfa_sessions(id) ON DELETE SET NULL
);

-- =============================================================================
-- LINK SECURITY SETTINGS TEMPLATES TABLE
-- =============================================================================
-- Predefined security setting templates for quick configuration
CREATE TABLE IF NOT EXISTS link_security_templates (
    id TEXT PRIMARY KEY,
    template_name TEXT NOT NULL,
    template_description TEXT,
    security_settings JSON NOT NULL, -- Complete security configuration
    is_default BOOLEAN DEFAULT FALSE,
    created_by TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE,
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL
);

-- =============================================================================
-- DATABASE INDEXES FOR PERFORMANCE
-- =============================================================================

-- Link Password Attempts Indexes
CREATE INDEX IF NOT EXISTS idx_link_password_attempts_link_id ON link_password_attempts(link_id);
CREATE INDEX IF NOT EXISTS idx_link_password_attempts_ip_address ON link_password_attempts(ip_address);
CREATE INDEX IF NOT EXISTS idx_link_password_attempts_attempt_time ON link_password_attempts(attempt_time);
CREATE INDEX IF NOT EXISTS idx_link_password_attempts_success ON link_password_attempts(success);

-- Link Geolocation Logs Indexes
CREATE INDEX IF NOT EXISTS idx_link_geolocation_logs_link_id ON link_geolocation_logs(link_id);
CREATE INDEX IF NOT EXISTS idx_link_geolocation_logs_ip_address ON link_geolocation_logs(ip_address);
CREATE INDEX IF NOT EXISTS idx_link_geolocation_logs_country ON link_geolocation_logs(country);
CREATE INDEX IF NOT EXISTS idx_link_geolocation_logs_city ON link_geolocation_logs(city);
CREATE INDEX IF NOT EXISTS idx_link_geolocation_logs_access_time ON link_geolocation_logs(access_time);
CREATE INDEX IF NOT EXISTS idx_link_geolocation_logs_access_allowed ON link_geolocation_logs(access_allowed);

-- Link MFA Sessions Indexes
CREATE INDEX IF NOT EXISTS idx_link_mfa_sessions_link_id ON link_mfa_sessions(link_id);
CREATE INDEX IF NOT EXISTS idx_link_mfa_sessions_session_token ON link_mfa_sessions(session_token);
CREATE INDEX IF NOT EXISTS idx_link_mfa_sessions_mfa_type ON link_mfa_sessions(mfa_type);
CREATE INDEX IF NOT EXISTS idx_link_mfa_sessions_verified ON link_mfa_sessions(verified);
CREATE INDEX IF NOT EXISTS idx_link_mfa_sessions_created_at ON link_mfa_sessions(created_at);

-- Link Decoy Messages Indexes
CREATE INDEX IF NOT EXISTS idx_link_decoy_messages_link_id ON link_decoy_messages(link_id);
CREATE INDEX IF NOT EXISTS idx_link_decoy_messages_decoy_type ON link_decoy_messages(decoy_type);
CREATE INDEX IF NOT EXISTS idx_link_decoy_messages_is_active ON link_decoy_messages(is_active);

-- Link Tamper Alerts Indexes
CREATE INDEX IF NOT EXISTS idx_link_tamper_alerts_link_id ON link_tamper_alerts(link_id);
CREATE INDEX IF NOT EXISTS idx_link_tamper_alerts_alert_type ON link_tamper_alerts(alert_type);
CREATE INDEX IF NOT EXISTS idx_link_tamper_alerts_severity ON link_tamper_alerts(severity);
CREATE INDEX IF NOT EXISTS idx_link_tamper_alerts_triggered_at ON link_tamper_alerts(triggered_at);
CREATE INDEX IF NOT EXISTS idx_link_tamper_alerts_acknowledged ON link_tamper_alerts(acknowledged);
CREATE INDEX IF NOT EXISTS idx_link_tamper_alerts_resolved ON link_tamper_alerts(resolved);

-- Link Access Sessions Indexes
CREATE INDEX IF NOT EXISTS idx_link_access_sessions_link_id ON link_access_sessions(link_id);
CREATE INDEX IF NOT EXISTS idx_link_access_sessions_session_token ON link_access_sessions(session_token);
CREATE INDEX IF NOT EXISTS idx_link_access_sessions_ip_address ON link_access_sessions(ip_address);
CREATE INDEX IF NOT EXISTS idx_link_access_sessions_session_start ON link_access_sessions(session_start);
CREATE INDEX IF NOT EXISTS idx_link_access_sessions_access_granted ON link_access_sessions(access_granted);

-- Link Security Templates Indexes
CREATE INDEX IF NOT EXISTS idx_link_security_templates_template_name ON link_security_templates(template_name);
CREATE INDEX IF NOT EXISTS idx_link_security_templates_is_default ON link_security_templates(is_default);
CREATE INDEX IF NOT EXISTS idx_link_security_templates_is_active ON link_security_templates(is_active);

-- Enhanced Secure Links Indexes (for new fields)
CREATE INDEX IF NOT EXISTS idx_secure_links_password_required ON secure_links(password_required);
CREATE INDEX IF NOT EXISTS idx_secure_links_time_lock_enabled ON secure_links(time_lock_enabled);
CREATE INDEX IF NOT EXISTS idx_secure_links_time_lock_until ON secure_links(time_lock_until);
CREATE INDEX IF NOT EXISTS idx_secure_links_auto_destruct_enabled ON secure_links(auto_destruct_enabled);
CREATE INDEX IF NOT EXISTS idx_secure_links_read_once_enabled ON secure_links(read_once_enabled);
CREATE INDEX IF NOT EXISTS idx_secure_links_geo_restriction_enabled ON secure_links(geo_restriction_enabled);
CREATE INDEX IF NOT EXISTS idx_secure_links_mfa_required ON secure_links(mfa_required);
CREATE INDEX IF NOT EXISTS idx_secure_links_decoy_enabled ON secure_links(decoy_enabled);
CREATE INDEX IF NOT EXISTS idx_secure_links_tamper_alerts_enabled ON secure_links(tamper_alerts_enabled);
CREATE INDEX IF NOT EXISTS idx_secure_links_revoked_at ON secure_links(revoked_at);

-- =============================================================================
-- DEFAULT SECURITY TEMPLATES
-- =============================================================================

-- High Security Template
INSERT OR IGNORE INTO link_security_templates (
    id, template_name, template_description, security_settings, is_default, created_at
) VALUES (
    'high-security',
    'High Security',
    'Maximum security with all features enabled',
    '{
        "password_required": true,
        "password_max_attempts": 3,
        "time_lock_enabled": false,
        "auto_destruct_enabled": true,
        "read_once_enabled": true,
        "geo_restriction_enabled": true,
        "geo_allowed_countries": ["US", "CA", "GB"],
        "mfa_required": true,
        "mfa_type": "totp",
        "decoy_enabled": true,
        "strip_metadata_enabled": true,
        "tamper_alerts_enabled": true,
        "tamper_alert_threshold": 2,
        "auto_destruct_after_attempts": 3
    }',
    false,
    CURRENT_TIMESTAMP
);

-- Standard Security Template
INSERT OR IGNORE INTO link_security_templates (
    id, template_name, template_description, security_settings, is_default, created_at
) VALUES (
    'standard-security',
    'Standard Security',
    'Balanced security with essential features',
    '{
        "password_required": true,
        "password_max_attempts": 5,
        "time_lock_enabled": false,
        "auto_destruct_enabled": false,
        "read_once_enabled": false,
        "geo_restriction_enabled": false,
        "mfa_required": false,
        "mfa_type": "none",
        "decoy_enabled": false,
        "strip_metadata_enabled": true,
        "tamper_alerts_enabled": true,
        "tamper_alert_threshold": 5,
        "auto_destruct_after_attempts": 10
    }',
    true,
    CURRENT_TIMESTAMP
);

-- Basic Security Template
INSERT OR IGNORE INTO link_security_templates (
    id, template_name, template_description, security_settings, is_default, created_at
) VALUES (
    'basic-security',
    'Basic Security',
    'Minimal security with basic protection',
    '{
        "password_required": false,
        "password_max_attempts": 0,
        "time_lock_enabled": false,
        "auto_destruct_enabled": false,
        "read_once_enabled": false,
        "geo_restriction_enabled": false,
        "mfa_required": false,
        "mfa_type": "none",
        "decoy_enabled": false,
        "strip_metadata_enabled": true,
        "tamper_alerts_enabled": true,
        "tamper_alert_threshold": 10,
        "auto_destruct_after_attempts": 0
    }',
    false,
    CURRENT_TIMESTAMP
);

-- =============================================================================
-- DEFAULT DECOY MESSAGES
-- =============================================================================

-- Generic decoy message template
INSERT OR IGNORE INTO link_decoy_messages (
    id, link_id, decoy_type, decoy_title, decoy_content, trigger_condition, is_active
) VALUES (
    'generic-decoy-template',
    'template',
    'generic',
    'Secure Message',
    'This secure message has been accessed. Thank you for using our secure messaging system.',
    '{"condition": "generic", "enabled": true}',
    true
);

-- Wrong password decoy message template
INSERT OR IGNORE INTO link_decoy_messages (
    id, link_id, decoy_type, decoy_title, decoy_content, trigger_condition, is_active
) VALUES (
    'wrong-password-decoy-template',
    'template',
    'wrong_password',
    'Invalid Access',
    'The password you entered is incorrect. Please try again or contact the sender for assistance.',
    '{"condition": "wrong_password", "enabled": true}',
    true
);

-- Revoked link decoy message template
INSERT OR IGNORE INTO link_decoy_messages (
    id, link_id, decoy_type, decoy_title, decoy_content, trigger_condition, is_active
) VALUES (
    'revoked-decoy-template',
    'template',
    'revoked',
    'Message Unavailable',
    'This secure message is no longer available. It may have been revoked by the sender or expired.',
    '{"condition": "revoked", "enabled": true}',
    true
);

-- =============================================================================
-- MIGRATION COMPLETION
-- =============================================================================

-- Update migration tracking
INSERT OR IGNORE INTO schema_migrations (version, name, applied_at) 
VALUES ('2024-08-21-002', 'Add Phase 2 Security Enforcement Features', CURRENT_TIMESTAMP);

-- Log migration completion
PRAGMA user_version = 2;

-- =============================================================================
-- MIGRATION SUMMARY
-- =============================================================================
-- 
-- This migration adds comprehensive security enforcement features for Phase 2:
-- 
-- 1. Enhanced secure_links table with 20+ new security fields
-- 2. link_password_attempts table for password protection tracking
-- 3. link_geolocation_logs table for enhanced location monitoring
-- 4. link_mfa_sessions table for external user MFA management
-- 5. link_decoy_messages table for decoy content system
-- 6. link_tamper_alerts table for suspicious activity detection
-- 7. link_access_sessions table for detailed session tracking
-- 8. link_security_templates table for predefined configurations
-- 9. Comprehensive indexing for optimal performance
-- 10. Default security templates and decoy messages
-- 
-- Total new tables: 7
-- Total new fields: 20+
-- Total new indexes: 25+
-- 
-- Ready for Phase 2 security feature implementation!
-- =============================================================================
