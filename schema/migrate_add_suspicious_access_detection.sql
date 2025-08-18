-- =============================================================================
-- SECURE EMAIL MVP - SUSPICIOUS ACCESS PATTERN DETECTION MIGRATION
-- =============================================================================
-- Migration to add suspicious access pattern detection features.
-- This enhances the existing access_events table and adds new tables for
-- suspicious activity tracking and detection rules.
-- =============================================================================

-- Add suspicious_flag to emails table
ALTER TABLE emails ADD COLUMN suspicious_flag BOOLEAN DEFAULT FALSE;
ALTER TABLE emails ADD COLUMN suspicious_flag_set_at DATETIME;
ALTER TABLE emails ADD COLUMN suspicious_flag_cleared_at DATETIME;
ALTER TABLE emails ADD COLUMN suspicious_flag_cleared_by TEXT;

-- Add suspicious_access_events table for detailed detection logging
CREATE TABLE IF NOT EXISTS suspicious_access_events (
    detection_id TEXT PRIMARY KEY,
    email_id TEXT NOT NULL,
    detection_type TEXT NOT NULL CHECK (detection_type IN ('multiple_failed_attempts', 'unusual_geolocation', 'rapid_multiple_ips', 'impossible_travel')),
    detection_rule TEXT NOT NULL,
    severity TEXT NOT NULL CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    triggered_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    resolved_at DATETIME,
    resolved_by TEXT,
    resolution_notes TEXT,
    detection_metadata TEXT, -- JSON field for storing detection-specific data
    FOREIGN KEY (email_id) REFERENCES emails(email_id) ON DELETE CASCADE,
    FOREIGN KEY (resolved_by) REFERENCES users(user_id) ON DELETE SET NULL
);

-- Add suspicious_access_patterns table for storing detected patterns
CREATE TABLE IF NOT EXISTS suspicious_access_patterns (
    pattern_id TEXT PRIMARY KEY,
    email_id TEXT NOT NULL,
    pattern_type TEXT NOT NULL,
    pattern_data TEXT NOT NULL, -- JSON field for storing pattern details
    first_detected_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE,
    confidence_score REAL DEFAULT 0.0, -- 0.0 to 1.0 confidence in the detection
    FOREIGN KEY (email_id) REFERENCES emails(email_id) ON DELETE CASCADE
);

-- Add detection_rules table for configurable detection rules
CREATE TABLE IF NOT EXISTS detection_rules (
    rule_id TEXT PRIMARY KEY,
    rule_name TEXT NOT NULL UNIQUE,
    rule_type TEXT NOT NULL CHECK (rule_type IN ('multiple_failed_attempts', 'unusual_geolocation', 'rapid_multiple_ips', 'impossible_travel')),
    is_enabled BOOLEAN DEFAULT TRUE,
    threshold_value INTEGER NOT NULL, -- e.g., number of attempts, time window in minutes
    time_window_minutes INTEGER NOT NULL, -- time window for pattern detection
    severity TEXT NOT NULL CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Add user_suspicious_activity_preferences table
CREATE TABLE IF NOT EXISTS user_suspicious_activity_preferences (
    user_id TEXT PRIMARY KEY,
    enable_suspicious_detection BOOLEAN DEFAULT TRUE,
    notify_on_suspicious_activity BOOLEAN DEFAULT TRUE,
    auto_flag_suspicious_emails BOOLEAN DEFAULT TRUE,
    minimum_severity_for_notification TEXT CHECK (minimum_severity_for_notification IN ('low', 'medium', 'high', 'critical')) DEFAULT 'medium',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

-- Add indexes for performance
CREATE INDEX IF NOT EXISTS idx_emails_suspicious_flag ON emails(suspicious_flag, suspicious_flag_set_at);
CREATE INDEX IF NOT EXISTS idx_suspicious_access_events_email_id ON suspicious_access_events(email_id);
CREATE INDEX IF NOT EXISTS idx_suspicious_access_events_detection_type ON suspicious_access_events(detection_type);
CREATE INDEX IF NOT EXISTS idx_suspicious_access_events_triggered_at ON suspicious_access_events(triggered_at);
CREATE INDEX IF NOT EXISTS idx_suspicious_access_events_severity ON suspicious_access_events(severity);
CREATE INDEX IF NOT EXISTS idx_suspicious_access_patterns_email_id ON suspicious_access_patterns(email_id);
CREATE INDEX IF NOT EXISTS idx_suspicious_access_patterns_pattern_type ON suspicious_access_patterns(pattern_type);
CREATE INDEX IF NOT EXISTS idx_suspicious_access_patterns_is_active ON suspicious_access_patterns(is_active);
CREATE INDEX IF NOT EXISTS idx_detection_rules_rule_type ON detection_rules(rule_type);
CREATE INDEX IF NOT EXISTS idx_detection_rules_is_enabled ON detection_rules(is_enabled);

-- Insert default detection rules
INSERT INTO detection_rules (rule_id, rule_name, rule_type, threshold_value, time_window_minutes, severity, description) VALUES
    ('rule_001', 'Multiple Failed Attempts', 'multiple_failed_attempts', 3, 5, 'high', 'Flag email if 3 or more failed access attempts within 5 minutes'),
    ('rule_002', 'Unusual Geolocation', 'unusual_geolocation', 1, 60, 'medium', 'Flag email if access from geolocation not seen in previous successful accesses'),
    ('rule_003', 'Rapid Multiple IPs', 'rapid_multiple_ips', 2, 10, 'high', 'Flag email if accessed from 2 or more different IPs within 10 minutes'),
    ('rule_004', 'Impossible Travel', 'impossible_travel', 1, 5, 'critical', 'Flag email if access from geographically impossible locations within 5 minutes');

-- Add trigger to update suspicious_flag_set_at when suspicious_flag is set
CREATE TRIGGER IF NOT EXISTS update_suspicious_flag_set_at
    AFTER UPDATE ON emails
    FOR EACH ROW
    WHEN NEW.suspicious_flag = 1 AND (OLD.suspicious_flag = 0 OR OLD.suspicious_flag IS NULL)
BEGIN
    UPDATE emails SET suspicious_flag_set_at = CURRENT_TIMESTAMP WHERE email_id = NEW.email_id;
END;

-- Add trigger to update suspicious_flag_cleared_at when suspicious_flag is cleared
CREATE TRIGGER IF NOT EXISTS update_suspicious_flag_cleared_at
    AFTER UPDATE ON emails
    FOR EACH ROW
    WHEN NEW.suspicious_flag = 0 AND OLD.suspicious_flag = 1
BEGIN
    UPDATE emails SET suspicious_flag_cleared_at = CURRENT_TIMESTAMP WHERE email_id = NEW.email_id;
END;

-- Add trigger to update updated_at timestamp for detection_rules
CREATE TRIGGER IF NOT EXISTS update_detection_rules_updated_at
    AFTER UPDATE ON detection_rules
    FOR EACH ROW
BEGIN
    UPDATE detection_rules SET updated_at = CURRENT_TIMESTAMP WHERE rule_id = NEW.rule_id;
END;

-- Add trigger to update updated_at timestamp for user_suspicious_activity_preferences
CREATE TRIGGER IF NOT EXISTS update_user_suspicious_activity_preferences_updated_at
    AFTER UPDATE ON user_suspicious_activity_preferences
    FOR EACH ROW
BEGIN
    UPDATE user_suspicious_activity_preferences SET updated_at = CURRENT_TIMESTAMP WHERE user_id = NEW.user_id;
END;







