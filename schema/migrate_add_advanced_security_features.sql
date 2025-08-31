-- =============================================================================
-- SECURE EMAIL MVP - ADVANCED SECURITY FEATURES MIGRATION
-- =============================================================================
-- This migration adds comprehensive security features to the emails table
-- including time-based controls, advanced access control, and security metadata
-- =============================================================================

-- Add time-based access control fields
ALTER TABLE emails ADD COLUMN time_lock BOOLEAN DEFAULT FALSE;
ALTER TABLE emails ADD COLUMN unlock_after TEXT;
ALTER TABLE emails ADD COLUMN expires_at TEXT;

-- Add advanced access control fields
ALTER TABLE emails ADD COLUMN burn_after_read BOOLEAN DEFAULT FALSE;
ALTER TABLE emails ADD COLUMN self_destruct_after_attempts BOOLEAN DEFAULT FALSE;
ALTER TABLE emails ADD COLUMN max_attempts INTEGER DEFAULT 3;

-- Add multi-factor authentication fields
ALTER TABLE emails ADD COLUMN require_mfa BOOLEAN DEFAULT FALSE;
ALTER TABLE emails ADD COLUMN mfa_type TEXT DEFAULT 'TOTP';
ALTER TABLE emails ADD COLUMN mfa_on_open BOOLEAN DEFAULT FALSE;
ALTER TABLE emails ADD COLUMN mfa_on_reply BOOLEAN DEFAULT FALSE;
ALTER TABLE emails ADD COLUMN mfa_on_forward BOOLEAN DEFAULT FALSE;

-- Add advanced security features
ALTER TABLE emails ADD COLUMN remote_revoke BOOLEAN DEFAULT FALSE;
ALTER TABLE emails ADD COLUMN decoy_message BOOLEAN DEFAULT FALSE;
ALTER TABLE emails ADD COLUMN decoy_secret TEXT;
ALTER TABLE emails ADD COLUMN strip_metadata BOOLEAN DEFAULT FALSE;
ALTER TABLE emails ADD COLUMN tamper_alerts BOOLEAN DEFAULT FALSE;

-- Add geolocation restrictions (JSON field for flexibility)
ALTER TABLE emails ADD COLUMN geolocation_json TEXT;

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_emails_time_lock ON emails(time_lock);
CREATE INDEX IF NOT EXISTS idx_emails_burn_after_read ON emails(burn_after_read);
CREATE INDEX IF NOT EXISTS idx_emails_require_mfa ON emails(require_mfa);
CREATE INDEX IF NOT EXISTS idx_emails_remote_revoke ON emails(remote_revoke);
CREATE INDEX IF NOT EXISTS idx_emails_decoy_message ON emails(decoy_message);
CREATE INDEX IF NOT EXISTS idx_emails_expires_at ON emails(expires_at);

-- Create composite indexes for common queries
CREATE INDEX IF NOT EXISTS idx_emails_security_features ON emails(
    time_lock, 
    burn_after_read, 
    require_mfa, 
    remote_revoke
);

-- Add sample data for testing
INSERT INTO emails (
    email_id, 
    sender_id, 
    recipient, 
    subject, 
    body, 
    time_lock, 
    unlock_after, 
    burn_after_read, 
    require_mfa, 
    mfa_type,
    remote_revoke,
    decoy_message,
    strip_metadata,
    tamper_alerts,
    geolocation_json
) VALUES (
    'test_advanced_security_1',
    'test@securesystem.email',
    'recipient@example.com',
    'Test Email with Advanced Security',
    'This is a test email with advanced security features enabled.',
    TRUE,
    '2024-12-31T23:59:59Z',
    TRUE,
    TRUE,
    'TOTP',
    TRUE,
    TRUE,
    TRUE,
    TRUE,
    '{"verification_type": "city_country", "city": "New York", "country": "US"}'
);

INSERT INTO emails (
    email_id, 
    sender_id, 
    recipient, 
    subject, 
    body, 
    time_lock, 
    unlock_after, 
    self_destruct_after_attempts,
    max_attempts,
    require_mfa, 
    mfa_type,
    mfa_on_open,
    mfa_on_reply,
    mfa_on_forward,
    remote_revoke,
    strip_metadata,
    tamper_alerts,
    geolocation_json
) VALUES (
    'test_advanced_security_2',
    'test@securesystem.email',
    'recipient2@example.com',
    'Test Email with MFA Triggers',
    'This is a test email with MFA triggers enabled.',
    FALSE,
    NULL,
    TRUE,
    5,
    TRUE,
    'EMAIL_CODE',
    TRUE,
    TRUE,
    TRUE,
    FALSE,
    TRUE,
    TRUE,
    '{"verification_type": "country", "country": "CA"}'
);

-- Log the migration
INSERT INTO audit_log (
    log_id, 
    timestamp, 
    event_type, 
    user_id, 
    outcome, 
    details, 
    severity
) VALUES (
    'migration_advanced_security_' || strftime('%s', 'now'),
    datetime('now'),
    'system_event',
    'system',
    'success',
    '{"migration": "add_advanced_security_features", "version": "1.0", "features_added": ["time_lock", "burn_after_read", "mfa_triggers", "remote_revoke", "decoy_message", "strip_metadata", "tamper_alerts", "geolocation_json"]}',
    'info'
);

-- Verify the migration
SELECT 
    'Migration completed successfully' as status,
    COUNT(*) as emails_count,
    SUM(CASE WHEN time_lock = 1 THEN 1 ELSE 0 END) as time_locked_emails,
    SUM(CASE WHEN burn_after_read = 1 THEN 1 ELSE 0 END) as burn_after_read_emails,
    SUM(CASE WHEN require_mfa = 1 THEN 1 ELSE 0 END) as mfa_required_emails,
    SUM(CASE WHEN remote_revoke = 1 THEN 1 ELSE 0 END) as remote_revoke_emails,
    SUM(CASE WHEN decoy_message = 1 THEN 1 ELSE 0 END) as decoy_message_emails
FROM emails;






