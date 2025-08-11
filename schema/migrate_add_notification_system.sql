-- =============================================================================
-- SECURE EMAIL MVP - NOTIFICATION SYSTEM MIGRATION
-- =============================================================================
-- Migration to add notification system tables for access events and preferences.
-- =============================================================================

-- Add notification_preferences table
CREATE TABLE IF NOT EXISTS notification_preferences (
    user_id TEXT PRIMARY KEY,
    email_notifications BOOLEAN DEFAULT TRUE,
    sms_notifications BOOLEAN DEFAULT FALSE,
    notify_on_success BOOLEAN DEFAULT TRUE,
    notify_on_failure BOOLEAN DEFAULT TRUE,
    notify_on_blocked BOOLEAN DEFAULT TRUE,
    include_geolocation BOOLEAN DEFAULT TRUE,
    include_device_info BOOLEAN DEFAULT TRUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

-- Add access_events table
CREATE TABLE IF NOT EXISTS access_events (
    event_id TEXT PRIMARY KEY,
    email_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    event_type TEXT NOT NULL CHECK (event_type IN ('success', 'failure', 'blocked')),
    ip_address TEXT NOT NULL,
    user_agent TEXT,
    country TEXT,
    city TEXT,
    device_type TEXT,
    failure_reason TEXT,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (email_id) REFERENCES emails(email_id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

-- Add indexes for performance
CREATE INDEX IF NOT EXISTS idx_access_events_user_id ON access_events(user_id);
CREATE INDEX IF NOT EXISTS idx_access_events_email_id ON access_events(email_id);
CREATE INDEX IF NOT EXISTS idx_access_events_timestamp ON access_events(timestamp);
CREATE INDEX IF NOT EXISTS idx_access_events_event_type ON access_events(event_type);

-- Add phone_number column to users table if it doesn't exist
-- This is needed for SMS notifications
ALTER TABLE users ADD COLUMN phone_number TEXT;

-- Add comments to document the new tables
-- notification_preferences: Stores user preferences for access notifications
-- access_events: Logs all email access attempts with metadata
-- phone_number: User's phone number for SMS notifications
