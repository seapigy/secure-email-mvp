-- =============================================================================
-- SECURE EMAIL MVP - RETENTION NOTIFICATIONS & ANALYTICS MIGRATION
-- =============================================================================
-- Migration to add retention notifications and analytics tables.
-- Micro-Iteration 4.25: Advanced Notification & Retention Analytics Enhancements
-- =============================================================================

-- Add retention_notifications table for tracking retention notifications
CREATE TABLE IF NOT EXISTS retention_notifications (
    notification_id TEXT PRIMARY KEY,
    email_id TEXT NOT NULL,
    sender_id TEXT NOT NULL,
    notification_type TEXT NOT NULL CHECK (notification_type IN ('expiration', 'cleanup')),
    notification_data TEXT, -- JSON or text representation of notification data
    sent_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (email_id) REFERENCES emails(email_id) ON DELETE CASCADE,
    FOREIGN KEY (sender_id) REFERENCES users(user_id) ON DELETE CASCADE
);

-- Add indexes for retention_notifications table
CREATE INDEX IF NOT EXISTS idx_retention_notifications_email_id ON retention_notifications(email_id);
CREATE INDEX IF NOT EXISTS idx_retention_notifications_sender_id ON retention_notifications(sender_id);
CREATE INDEX IF NOT EXISTS idx_retention_notifications_type ON retention_notifications(notification_type);
CREATE INDEX IF NOT EXISTS idx_retention_notifications_created_at ON retention_notifications(created_at);

-- Add cleanup_logs table for detailed cleanup logging
CREATE TABLE IF NOT EXISTS cleanup_logs (
    log_id TEXT PRIMARY KEY,
    email_id TEXT,
    sender_id TEXT,
    cleanup_reason TEXT NOT NULL CHECK (cleanup_reason IN ('expired', 'burned', 'self_destructed')),
    cleanup_time DATETIME NOT NULL,
    initiator TEXT NOT NULL CHECK (initiator IN ('worker', 'manual')),
    emails_processed INTEGER DEFAULT 0,
    emails_deleted INTEGER DEFAULT 0,
    emails_skipped INTEGER DEFAULT 0,
    audit_logs_deleted INTEGER DEFAULT 0,
    duration TEXT, -- Duration of cleanup operation
    metadata TEXT, -- JSON metadata about the cleanup operation
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (email_id) REFERENCES emails(email_id) ON DELETE SET NULL,
    FOREIGN KEY (sender_id) REFERENCES users(user_id) ON DELETE SET NULL
);

-- Add indexes for cleanup_logs table
CREATE INDEX IF NOT EXISTS idx_cleanup_logs_email_id ON cleanup_logs(email_id);
CREATE INDEX IF NOT EXISTS idx_cleanup_logs_sender_id ON cleanup_logs(sender_id);
CREATE INDEX IF NOT EXISTS idx_cleanup_logs_cleanup_time ON cleanup_logs(cleanup_time);
CREATE INDEX IF NOT EXISTS idx_cleanup_logs_reason ON cleanup_logs(cleanup_reason);
CREATE INDEX IF NOT EXISTS idx_cleanup_logs_initiator ON cleanup_logs(initiator);

-- Add retention_analytics_cache table for caching analytics results
CREATE TABLE IF NOT EXISTS retention_analytics_cache (
    cache_key TEXT PRIMARY KEY,
    analytics_data TEXT NOT NULL, -- JSON representation of analytics data
    filters_hash TEXT NOT NULL, -- Hash of the filters used for this cache entry
    generated_at DATETIME NOT NULL,
    expires_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Add indexes for retention_analytics_cache table
CREATE INDEX IF NOT EXISTS idx_retention_analytics_cache_expires_at ON retention_analytics_cache(expires_at);
CREATE INDEX IF NOT EXISTS idx_retention_analytics_cache_filters_hash ON retention_analytics_cache(filters_hash);

-- Add retention_notification_preferences table for user notification preferences
CREATE TABLE IF NOT EXISTS retention_notification_preferences (
    user_id TEXT PRIMARY KEY,
    enable_expiration_notifications BOOLEAN DEFAULT TRUE,
    enable_cleanup_notifications BOOLEAN DEFAULT FALSE,
    expiration_advance_notice_hours INTEGER DEFAULT 24,
    notification_email_template TEXT DEFAULT 'default',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

-- Add indexes for retention_notification_preferences table
CREATE INDEX IF NOT EXISTS idx_retention_notification_preferences_enable_expiration ON retention_notification_preferences(enable_expiration_notifications);
CREATE INDEX IF NOT EXISTS idx_retention_notification_preferences_enable_cleanup ON retention_notification_preferences(enable_cleanup_notifications);

-- Add retention_analytics_schedules table for scheduled analytics generation
CREATE TABLE IF NOT EXISTS retention_analytics_schedules (
    schedule_id TEXT PRIMARY KEY,
    schedule_name TEXT NOT NULL,
    schedule_type TEXT NOT NULL CHECK (schedule_type IN ('daily', 'weekly', 'monthly')),
    filters TEXT, -- JSON representation of filters to apply
    last_run DATETIME,
    next_run DATETIME,
    is_active BOOLEAN DEFAULT TRUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Add indexes for retention_analytics_schedules table
CREATE INDEX IF NOT EXISTS idx_retention_analytics_schedules_next_run ON retention_analytics_schedules(next_run);
CREATE INDEX IF NOT EXISTS idx_retention_analytics_schedules_is_active ON retention_analytics_schedules(is_active);

-- Insert default retention notification preferences for existing users
INSERT OR IGNORE INTO retention_notification_preferences (user_id, enable_expiration_notifications, enable_cleanup_notifications, expiration_advance_notice_hours)
SELECT user_id, TRUE, FALSE, 24 FROM users;

-- Create a view for retention statistics summary
CREATE VIEW IF NOT EXISTS retention_stats_summary AS
SELECT 
    COUNT(*) as total_emails,
    COUNT(CASE WHEN expires_at IS NOT NULL AND expires_at <= datetime('now') THEN 1 END) as expired_emails,
    COUNT(CASE WHEN self_destructed = 1 THEN 1 END) as self_destructed_emails,
    COUNT(CASE WHEN encrypted_blob_url IS NULL THEN 1 END) as deleted_emails,
    COUNT(CASE WHEN encrypted_blob_url IS NOT NULL AND (expires_at IS NULL OR expires_at > datetime('now')) AND self_destructed = 0 THEN 1 END) as active_emails,
    AVG(
        CASE 
            WHEN expires_at IS NOT NULL THEN 
                CAST((julianday(expires_at) - julianday(created_at)) AS REAL)
            ELSE 
                CAST((julianday('now') - julianday(created_at)) AS REAL)
        END
    ) as average_retention_days
FROM emails;

-- Create a view for user retention statistics
CREATE VIEW IF NOT EXISTS user_retention_stats AS
SELECT 
    sender_id,
    COUNT(*) as emails_sent,
    COUNT(CASE WHEN expires_at IS NOT NULL AND expires_at <= datetime('now') THEN 1 END) as emails_expired,
    COUNT(CASE WHEN self_destructed = 1 THEN 1 END) as emails_self_destructed,
    COUNT(CASE WHEN encrypted_blob_url IS NULL THEN 1 END) as emails_deleted,
    COUNT(CASE WHEN encrypted_blob_url IS NOT NULL AND (expires_at IS NULL OR expires_at > datetime('now')) AND self_destructed = 0 THEN 1 END) as emails_active,
    AVG(
        CASE 
            WHEN expires_at IS NOT NULL THEN 
                CAST((julianday(expires_at) - julianday(created_at)) AS REAL)
            ELSE 
                CAST((julianday('now') - julianday(created_at)) AS REAL)
        END
    ) as average_retention_days
FROM emails
GROUP BY sender_id;

-- Create a view for daily retention trends
CREATE VIEW IF NOT EXISTS daily_retention_trends AS
SELECT 
    DATE(created_at) as date,
    COUNT(*) as emails_sent,
    COUNT(CASE WHEN expires_at IS NOT NULL AND expires_at <= datetime('now') THEN 1 END) as emails_expired,
    COUNT(CASE WHEN encrypted_blob_url IS NULL THEN 1 END) as emails_deleted,
    AVG(
        CASE 
            WHEN expires_at IS NOT NULL THEN 
                CAST((julianday(expires_at) - julianday(created_at)) AS REAL)
            ELSE 
                CAST((julianday('now') - julianday(created_at)) AS REAL)
        END
    ) as average_retention_days
FROM emails
WHERE created_at >= datetime('now', '-30 days')
GROUP BY DATE(created_at)
ORDER BY date DESC;

-- Create a view for cleanup operations summary
CREATE VIEW IF NOT EXISTS cleanup_operations_summary AS
SELECT 
    DATE(cleanup_time) as cleanup_date,
    initiator,
    COUNT(*) as cleanup_operations,
    SUM(emails_processed) as total_emails_processed,
    SUM(emails_deleted) as total_emails_deleted,
    SUM(emails_skipped) as total_emails_skipped,
    SUM(audit_logs_deleted) as total_audit_logs_deleted,
    AVG(CAST(REPLACE(duration, 'ms', '') AS REAL)) as average_duration_ms
FROM cleanup_logs
WHERE cleanup_time >= datetime('now', '-30 days')
GROUP BY DATE(cleanup_time), initiator
ORDER BY cleanup_date DESC;

-- Add trigger to automatically update updated_at timestamp for retention_notification_preferences
CREATE TRIGGER IF NOT EXISTS update_retention_notification_preferences_updated_at
    AFTER UPDATE ON retention_notification_preferences
    FOR EACH ROW
BEGIN
    UPDATE retention_notification_preferences 
    SET updated_at = CURRENT_TIMESTAMP 
    WHERE user_id = NEW.user_id;
END;

-- Add trigger to automatically update updated_at timestamp for retention_analytics_schedules
CREATE TRIGGER IF NOT EXISTS update_retention_analytics_schedules_updated_at
    AFTER UPDATE ON retention_analytics_schedules
    FOR EACH ROW
BEGIN
    UPDATE retention_analytics_schedules 
    SET updated_at = CURRENT_TIMESTAMP 
    WHERE schedule_id = NEW.schedule_id;
END;

-- Add trigger to clean up expired analytics cache entries
CREATE TRIGGER IF NOT EXISTS cleanup_expired_analytics_cache
    AFTER INSERT ON retention_analytics_cache
    FOR EACH ROW
BEGIN
    DELETE FROM retention_analytics_cache WHERE expires_at < datetime('now');
END;



















