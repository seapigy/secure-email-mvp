-- =============================================================================
-- SECURE EMAIL MVP - NOTIFICATION DELIVERY CONTROLS MIGRATION
-- =============================================================================
-- Migration to add notification delivery frequency controls and rate limiting.
-- Micro-Iteration 4.18: Notification Delivery Controls & Rate Limiting
-- =============================================================================

-- Add notification delivery frequency controls to notification_preferences table
ALTER TABLE notification_preferences ADD COLUMN delivery_frequency TEXT DEFAULT 'immediate' CHECK (delivery_frequency IN ('immediate', 'daily_digest', 'first_attempt_only', 'threshold_trigger'));
ALTER TABLE notification_preferences ADD COLUMN threshold_attempts INTEGER DEFAULT 3;
ALTER TABLE notification_preferences ADD COLUMN rate_limit_window_minutes INTEGER DEFAULT 15;
ALTER TABLE notification_preferences ADD COLUMN rate_limit_max_notifications INTEGER DEFAULT 5;

-- Add per-email notification preferences table
CREATE TABLE IF NOT EXISTS email_notification_preferences (
    email_id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    delivery_frequency TEXT DEFAULT 'immediate' CHECK (delivery_frequency IN ('immediate', 'daily_digest', 'first_attempt_only', 'threshold_trigger')),
    threshold_attempts INTEGER DEFAULT 3,
    rate_limit_window_minutes INTEGER DEFAULT 15,
    rate_limit_max_notifications INTEGER DEFAULT 5,
    inherit_global_settings BOOLEAN DEFAULT TRUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (email_id) REFERENCES emails(email_id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

-- Add notification suppression tracking table
CREATE TABLE IF NOT EXISTS notification_suppressions (
    suppression_id TEXT PRIMARY KEY,
    email_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    event_id TEXT NOT NULL,
    suppression_reason TEXT NOT NULL CHECK (suppression_reason IN ('rate_limited', 'frequency_controlled', 'threshold_not_met', 'first_attempt_only')),
    suppressed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    original_event_type TEXT NOT NULL,
    ip_address TEXT,
    user_agent TEXT,
    country TEXT,
    city TEXT,
    device_type TEXT,
    failure_reason TEXT,
    FOREIGN KEY (email_id) REFERENCES emails(email_id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (event_id) REFERENCES access_events(event_id) ON DELETE CASCADE
);

-- Add notification rate limiting tracking table
CREATE TABLE IF NOT EXISTS notification_rate_limits (
    rate_limit_id TEXT PRIMARY KEY,
    email_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    ip_address TEXT NOT NULL,
    notification_count INTEGER DEFAULT 1,
    window_start DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_notification_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (email_id) REFERENCES emails(email_id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

-- Add daily digest tracking table
CREATE TABLE IF NOT EXISTS daily_digest_tracking (
    digest_id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    email_id TEXT NOT NULL,
    digest_date DATE NOT NULL,
    event_count INTEGER DEFAULT 0,
    last_event_at DATETIME,
    digest_sent BOOLEAN DEFAULT FALSE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (email_id) REFERENCES emails(email_id) ON DELETE CASCADE,
    UNIQUE(user_id, email_id, digest_date)
);

-- Add indexes for performance
CREATE INDEX IF NOT EXISTS idx_email_notification_preferences_user_id ON email_notification_preferences(user_id);
CREATE INDEX IF NOT EXISTS idx_email_notification_preferences_email_id ON email_notification_preferences(email_id);
CREATE INDEX IF NOT EXISTS idx_notification_suppressions_user_id ON notification_suppressions(user_id);
CREATE INDEX IF NOT EXISTS idx_notification_suppressions_email_id ON notification_suppressions(email_id);
CREATE INDEX IF NOT EXISTS idx_notification_suppressions_suppressed_at ON notification_suppressions(suppressed_at);
CREATE INDEX IF NOT EXISTS idx_notification_rate_limits_email_id ON notification_rate_limits(email_id);
CREATE INDEX IF NOT EXISTS idx_notification_rate_limits_ip_address ON notification_rate_limits(ip_address);
CREATE INDEX IF NOT EXISTS idx_notification_rate_limits_window_start ON notification_rate_limits(window_start);
CREATE INDEX IF NOT EXISTS idx_daily_digest_tracking_user_id ON daily_digest_tracking(user_id);
CREATE INDEX IF NOT EXISTS idx_daily_digest_tracking_digest_date ON daily_digest_tracking(digest_date);
CREATE INDEX IF NOT EXISTS idx_daily_digest_tracking_digest_sent ON daily_digest_tracking(digest_sent);

-- Add trigger to update updated_at timestamp for email_notification_preferences
CREATE TRIGGER IF NOT EXISTS update_email_notification_preferences_updated_at 
    AFTER UPDATE ON email_notification_preferences
    FOR EACH ROW
BEGIN
    UPDATE email_notification_preferences SET updated_at = CURRENT_TIMESTAMP WHERE email_id = NEW.email_id;
END;

-- Add comments to document the new tables and columns
-- delivery_frequency: Controls how often notifications are sent ('immediate', 'daily_digest', 'first_attempt_only', 'threshold_trigger')
-- threshold_attempts: Number of failed attempts before sending notification (for threshold_trigger mode)
-- rate_limit_window_minutes: Time window for rate limiting notifications
-- rate_limit_max_notifications: Maximum notifications allowed within the rate limit window
-- inherit_global_settings: Whether to use global user preferences or email-specific preferences
-- notification_suppressions: Tracks suppressed notifications for audit transparency
-- notification_rate_limits: Tracks rate limiting state for each email/IP combination
-- daily_digest_tracking: Tracks daily digest state for users with daily_digest frequency
