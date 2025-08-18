-- =============================================================================
-- SECURE EMAIL MVP - READ RECEIPTS & EXPIRATION ALERTS MIGRATION
-- =============================================================================
-- This migration adds read receipts and expiration alerts support for Micro-Iteration 4.19:
-- Email Read Receipt & Expiration Alerts
-- =============================================================================

-- Add read receipt fields to emails table
ALTER TABLE emails ADD COLUMN first_read_at DATETIME;
ALTER TABLE emails ADD COLUMN read_count INTEGER DEFAULT 0;
ALTER TABLE emails ADD COLUMN read_receipt_sent BOOLEAN DEFAULT FALSE;
ALTER TABLE emails ADD COLUMN expiration_alert_sent BOOLEAN DEFAULT FALSE;
ALTER TABLE emails ADD COLUMN final_expiration_alert_sent BOOLEAN DEFAULT FALSE;

-- Add read receipt and expiration alert preferences to emails table
ALTER TABLE emails ADD COLUMN enable_read_receipts BOOLEAN DEFAULT TRUE;
ALTER TABLE emails ADD COLUMN enable_expiration_alerts BOOLEAN DEFAULT TRUE;
ALTER TABLE emails ADD COLUMN expiration_alert_hours INTEGER DEFAULT 24; -- Alert X hours before expiration

-- Create read_events table for detailed read tracking
CREATE TABLE IF NOT EXISTS read_events (
    event_id TEXT PRIMARY KEY,
    email_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    read_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    ip_address TEXT,
    user_agent TEXT,
    country TEXT,
    city TEXT,
    device_type TEXT,
    is_first_read BOOLEAN DEFAULT FALSE,
    FOREIGN KEY (email_id) REFERENCES emails(email_id),
    FOREIGN KEY (user_id) REFERENCES users(user_id)
);

-- Create read_receipts table for read receipt notifications
CREATE TABLE IF NOT EXISTS read_receipts (
    receipt_id TEXT PRIMARY KEY,
    email_id TEXT NOT NULL,
    sender_id TEXT NOT NULL,
    recipient_id TEXT NOT NULL,
    sent_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    delivery_method TEXT NOT NULL, -- 'email' or 'sms'
    delivery_status TEXT DEFAULT 'pending', -- 'pending', 'sent', 'failed'
    error_message TEXT,
    metadata TEXT, -- JSON with read details
    FOREIGN KEY (email_id) REFERENCES emails(email_id),
    FOREIGN KEY (sender_id) REFERENCES users(user_id),
    FOREIGN KEY (recipient_id) REFERENCES users(user_id)
);

-- Create expiration_alerts table for expiration notifications
CREATE TABLE IF NOT EXISTS expiration_alerts (
    alert_id TEXT PRIMARY KEY,
    email_id TEXT NOT NULL,
    sender_id TEXT NOT NULL,
    alert_type TEXT NOT NULL, -- 'reminder' or 'final'
    sent_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    delivery_method TEXT NOT NULL, -- 'email' or 'sms'
    delivery_status TEXT DEFAULT 'pending', -- 'pending', 'sent', 'failed'
    error_message TEXT,
    metadata TEXT, -- JSON with expiration details
    FOREIGN KEY (email_id) REFERENCES emails(email_id),
    FOREIGN KEY (sender_id) REFERENCES users(user_id)
);

-- Create global notification preferences for read receipts and expiration alerts
CREATE TABLE IF NOT EXISTS read_receipt_preferences (
    user_id TEXT PRIMARY KEY,
    enable_read_receipts BOOLEAN DEFAULT TRUE,
    enable_expiration_alerts BOOLEAN DEFAULT TRUE,
    expiration_alert_hours INTEGER DEFAULT 24,
    delivery_methods TEXT DEFAULT 'email,sms', -- Comma-separated list
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id)
);

-- Add indexes for performance
CREATE INDEX IF NOT EXISTS idx_emails_read_receipts ON emails(first_read_at, read_count, read_receipt_sent);
CREATE INDEX IF NOT EXISTS idx_emails_expiration_alerts ON emails(expires_at, expiration_alert_sent, final_expiration_alert_sent);
CREATE INDEX IF NOT EXISTS idx_read_events_email_id ON read_events(email_id);
CREATE INDEX IF NOT EXISTS idx_read_events_user_id ON read_events(user_id);
CREATE INDEX IF NOT EXISTS idx_read_events_timestamp ON read_events(read_at);
CREATE INDEX IF NOT EXISTS idx_read_receipts_email_id ON read_receipts(email_id);
CREATE INDEX IF NOT EXISTS idx_read_receipts_sender_id ON read_receipts(sender_id);
CREATE INDEX IF NOT EXISTS idx_expiration_alerts_email_id ON expiration_alerts(email_id);
CREATE INDEX IF NOT EXISTS idx_expiration_alerts_sender_id ON expiration_alerts(sender_id);
CREATE INDEX IF NOT EXISTS idx_expiration_alerts_alert_type ON expiration_alerts(alert_type);

-- Add comments to document the new fields
-- first_read_at: Timestamp of the first successful read of the email
-- read_count: Total number of successful reads
-- read_receipt_sent: Whether a read receipt notification has been sent
-- expiration_alert_sent: Whether an expiration reminder has been sent
-- final_expiration_alert_sent: Whether the final expiration alert has been sent
-- enable_read_receipts: Whether to send read receipts for this email
-- enable_expiration_alerts: Whether to send expiration alerts for this email
-- expiration_alert_hours: Hours before expiration to send reminder







