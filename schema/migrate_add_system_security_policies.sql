-- Migration to add system security policies table
-- This table stores system-wide security policies that apply to the entire application

-- System Security Policies table
CREATE TABLE IF NOT EXISTS system_security_policies (
    policy_id TEXT PRIMARY KEY,
    policy_name TEXT NOT NULL,
    policy_description TEXT,
    policy_type TEXT NOT NULL, -- 'password', 'mfa', 'session', 'retention', 'dlp', 'general'
    is_active BOOLEAN DEFAULT TRUE,
    policy_value TEXT NOT NULL, -- JSON string for complex policies
    policy_category TEXT NOT NULL, -- 'authentication', 'authorization', 'data_protection', 'compliance'
    severity TEXT DEFAULT 'medium', -- 'low', 'medium', 'high', 'critical'
    enforcement_level TEXT DEFAULT 'recommended', -- 'recommended', 'required', 'mandatory'
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    created_by TEXT,
    last_modified_by TEXT
);

-- Insert default system security policies
INSERT OR IGNORE INTO system_security_policies (policy_id, policy_name, policy_description, policy_type, is_active, policy_value, policy_category, severity, enforcement_level, created_by) VALUES
-- Password Policies
('sys_pwd_001', 'Password Complexity', 'Minimum password complexity requirements', 'password', TRUE, '{"min_length": 12, "require_uppercase": true, "require_lowercase": true, "require_numbers": true, "require_special": true, "max_age_days": 90}', 'authentication', 'high', 'required', 'system'),
('sys_pwd_002', 'Password History', 'Prevent password reuse', 'password', TRUE, '{"history_count": 5}', 'authentication', 'medium', 'recommended', 'system'),
('sys_pwd_003', 'Password Lockout', 'Account lockout after failed attempts', 'password', TRUE, '{"max_attempts": 5, "lockout_duration_minutes": 30}', 'authentication', 'high', 'required', 'system'),

-- MFA Policies
('sys_mfa_001', 'MFA Requirement', 'Multi-factor authentication requirements', 'mfa', TRUE, '{"enabled": true, "methods": ["totp", "email"], "grace_period_days": 7}', 'authentication', 'critical', 'mandatory', 'system'),
('sys_mfa_002', 'MFA Backup Codes', 'Backup codes for MFA recovery', 'mfa', TRUE, '{"enabled": true, "code_count": 10, "code_length": 8}', 'authentication', 'medium', 'recommended', 'system'),

-- Session Policies
('sys_sess_001', 'Session Timeout', 'Automatic session timeout settings', 'session', TRUE, '{"timeout_minutes": 30, "extend_on_activity": true, "max_session_hours": 24}', 'authorization', 'high', 'required', 'system'),
('sys_sess_002', 'Concurrent Sessions', 'Limit concurrent user sessions', 'session', TRUE, '{"max_concurrent": 3, "terminate_oldest": true}', 'authorization', 'medium', 'recommended', 'system'),
('sys_sess_003', 'Session Security', 'Session security requirements', 'session', TRUE, '{"require_https": true, "secure_cookies": true, "http_only": true}', 'authorization', 'high', 'required', 'system'),

-- Retention Policies
('sys_ret_001', 'Email Retention', 'Email retention and deletion policies', 'retention', TRUE, '{"default_retention_days": 365, "max_retention_days": 2555, "auto_delete": true}', 'data_protection', 'high', 'required', 'system'),
('sys_ret_002', 'Audit Log Retention', 'Audit log retention requirements', 'retention', TRUE, '{"retention_days": 2555, "archive_after_days": 90}', 'compliance', 'critical', 'mandatory', 'system'),
('sys_ret_003', 'DLP Scan Retention', 'DLP scan results retention', 'retention', TRUE, '{"retention_days": 2555, "anonymize_after_days": 90}', 'data_protection', 'high', 'required', 'system'),

-- DLP Policies
('sys_dlp_001', 'DLP Scanning', 'Data Loss Prevention scanning settings', 'dlp', TRUE, '{"enabled": true, "scan_email_body": true, "scan_attachments": true, "scan_replies": true, "auto_block": false}', 'data_protection', 'high', 'required', 'system'),
('sys_dlp_002', 'DLP Patterns', 'DLP pattern detection settings', 'dlp', TRUE, '{"credit_card": true, "ssn": true, "email": true, "phone": true, "custom_patterns": []}', 'data_protection', 'high', 'required', 'system'),

-- General Security Policies
('sys_gen_001', 'Rate Limiting', 'API rate limiting settings', 'general', TRUE, '{"requests_per_minute": 60, "burst_limit": 10, "apply_to_all": true}', 'authorization', 'medium', 'recommended', 'system'),
('sys_gen_002', 'IP Restrictions', 'IP address restrictions', 'general', TRUE, '{"enabled": false, "allowed_networks": [], "blocked_networks": []}', 'authorization', 'medium', 'recommended', 'system'),
('sys_gen_003', 'File Upload Security', 'File upload security settings', 'general', TRUE, '{"max_file_size_mb": 25, "allowed_types": ["pdf", "doc", "docx", "txt", "jpg", "png"], "scan_uploads": true}', 'data_protection', 'high', 'required', 'system'),
('sys_gen_004', 'Encryption Standards', 'Data encryption requirements', 'general', TRUE, '{"storage_encryption": true, "transit_encryption": true, "algorithm": "AES-256-GCM", "key_rotation_days": 365}', 'data_protection', 'critical', 'mandatory', 'system');

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_system_security_policies_type ON system_security_policies(policy_type);
CREATE INDEX IF NOT EXISTS idx_system_security_policies_category ON system_security_policies(policy_category);
CREATE INDEX IF NOT EXISTS idx_system_security_policies_active ON system_security_policies(is_active);
CREATE INDEX IF NOT EXISTS idx_system_security_policies_updated ON system_security_policies(updated_at);

-- Update schema version
PRAGMA user_version = 7;







