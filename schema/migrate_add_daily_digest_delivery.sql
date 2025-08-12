-- =============================================================================
-- SECURE EMAIL MVP - DAILY DIGEST DELIVERY MIGRATION
-- =============================================================================
-- This migration adds daily digest delivery functionality to the notification system.
-- Extends notification preferences with digest-specific settings and creates
-- digest delivery tracking tables.
-- =============================================================================

-- Add daily digest delivery settings to notification_preferences table
ALTER TABLE notification_preferences ADD COLUMN digest_delivery_time TEXT DEFAULT '08:00' CHECK (digest_delivery_time LIKE '__:__');
ALTER TABLE notification_preferences ADD COLUMN digest_email_enabled BOOLEAN DEFAULT TRUE;
ALTER TABLE notification_preferences ADD COLUMN digest_sms_enabled BOOLEAN DEFAULT FALSE;

-- Add daily digest delivery settings to email_notification_preferences table
ALTER TABLE email_notification_preferences ADD COLUMN digest_delivery_time TEXT DEFAULT '08:00' CHECK (digest_delivery_time LIKE '__:__');
ALTER TABLE email_notification_preferences ADD COLUMN digest_email_enabled BOOLEAN DEFAULT TRUE;
ALTER TABLE email_notification_preferences ADD COLUMN digest_sms_enabled BOOLEAN DEFAULT FALSE;

-- Create daily digest delivery tracking table
CREATE TABLE IF NOT EXISTS daily_digest_deliveries (
    delivery_id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    digest_date DATE NOT NULL,
    email_sent BOOLEAN DEFAULT FALSE,
    sms_sent BOOLEAN DEFAULT FALSE,
    email_sent_at DATETIME,
    sms_sent_at DATETIME,
    event_count INTEGER DEFAULT 0,
    email_count INTEGER DEFAULT 0,
    success_count INTEGER DEFAULT 0,
    failure_count INTEGER DEFAULT 0,
    blocked_count INTEGER DEFAULT 0,
    suppression_count INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    UNIQUE(user_id, digest_date)
);

-- Create daily digest content table for audit purposes
CREATE TABLE IF NOT EXISTS daily_digest_content (
    content_id TEXT PRIMARY KEY,
    delivery_id TEXT NOT NULL,
    email_id TEXT NOT NULL,
    email_subject TEXT,
    recipient TEXT,
    success_count INTEGER DEFAULT 0,
    failure_count INTEGER DEFAULT 0,
    blocked_count INTEGER DEFAULT 0,
    last_access_at DATETIME,
    last_ip_address TEXT,
    last_device_type TEXT,
    last_country TEXT,
    last_city TEXT,
    suppression_count INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (delivery_id) REFERENCES daily_digest_deliveries(delivery_id) ON DELETE CASCADE,
    FOREIGN KEY (email_id) REFERENCES emails(email_id) ON DELETE CASCADE
);

-- Add indexes for performance
CREATE INDEX IF NOT EXISTS idx_daily_digest_deliveries_user_id ON daily_digest_deliveries(user_id);
CREATE INDEX IF NOT EXISTS idx_daily_digest_deliveries_digest_date ON daily_digest_deliveries(digest_date);
CREATE INDEX IF NOT EXISTS idx_daily_digest_deliveries_email_sent ON daily_digest_deliveries(email_sent);
CREATE INDEX IF NOT EXISTS idx_daily_digest_deliveries_sms_sent ON daily_digest_deliveries(sms_sent);
CREATE INDEX IF NOT EXISTS idx_daily_digest_content_delivery_id ON daily_digest_content(delivery_id);
CREATE INDEX IF NOT EXISTS idx_daily_digest_content_email_id ON daily_digest_content(email_id);

-- Add trigger to update updated_at timestamp for email_notification_preferences
-- (This trigger already exists, but we're ensuring it's here for completeness)
CREATE TRIGGER IF NOT EXISTS update_email_notification_preferences_updated_at 
    AFTER UPDATE ON email_notification_preferences
    FOR EACH ROW
BEGIN
    UPDATE email_notification_preferences SET updated_at = CURRENT_TIMESTAMP WHERE email_id = NEW.email_id;
END;

-- Add comments to document the new columns and tables
-- digest_delivery_time: Time of day to send daily digest (format: HH:MM in UTC)
-- digest_email_enabled: Whether to send digest via email
-- digest_sms_enabled: Whether to send digest via SMS
-- daily_digest_deliveries: Tracks daily digest delivery status and summary statistics
-- daily_digest_content: Stores digest content for audit purposes (without sensitive data)
