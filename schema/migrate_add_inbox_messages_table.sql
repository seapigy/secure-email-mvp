-- Migration: Add inbox_messages table for better user isolation and inbox functionality
-- This migration creates a dedicated table for inbox messages with proper user isolation

-- Create inbox_messages table
CREATE TABLE IF NOT EXISTS inbox_messages (
    id TEXT PRIMARY KEY,                    -- UUID for inbox message identification
    user_id TEXT NOT NULL,                  -- UUID of the user who owns this inbox message
    email_id TEXT NOT NULL,                 -- UUID reference to the emails table
    is_read BOOLEAN DEFAULT FALSE,          -- Whether the user has read this message
    is_deleted BOOLEAN DEFAULT FALSE,       -- Soft delete flag for inbox messages
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    -- Foreign key constraints
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (email_id) REFERENCES emails(email_id) ON DELETE CASCADE,
    
    -- Unique constraint to prevent duplicate inbox entries
    UNIQUE(user_id, email_id)
);

-- Indexes for performance optimization
CREATE INDEX IF NOT EXISTS idx_inbox_messages_user_id ON inbox_messages(user_id);
CREATE INDEX IF NOT EXISTS idx_inbox_messages_user_read ON inbox_messages(user_id, is_read);
CREATE INDEX IF NOT EXISTS idx_inbox_messages_user_deleted ON inbox_messages(user_id, is_deleted);
CREATE INDEX IF NOT EXISTS idx_inbox_messages_user_created ON inbox_messages(user_id, created_at);
CREATE INDEX IF NOT EXISTS idx_inbox_messages_email_id ON inbox_messages(email_id);

-- Trigger to update updated_at timestamp
CREATE TRIGGER IF NOT EXISTS update_inbox_messages_updated_at 
    AFTER UPDATE ON inbox_messages
    FOR EACH ROW
BEGIN
    UPDATE inbox_messages SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- Add is_deleted column to emails table if it doesn't exist
-- This provides a proper soft delete mechanism for emails
-- SQLite doesn't support IF NOT EXISTS for ADD COLUMN, so we'll handle this in the application
-- ALTER TABLE emails ADD COLUMN is_deleted BOOLEAN DEFAULT FALSE;

-- Index for the new is_deleted column (commented out since column doesn't exist yet)
-- CREATE INDEX IF NOT EXISTS idx_emails_is_deleted ON emails(is_deleted);
-- CREATE INDEX IF NOT EXISTS idx_emails_recipient_not_deleted ON emails(recipient, is_deleted);

-- Migration script to populate inbox_messages from existing emails
-- This will create inbox entries for all existing emails where recipient is a registered user
INSERT OR IGNORE INTO inbox_messages (id, user_id, email_id, is_read, is_deleted, created_at, updated_at)
SELECT 
    lower(hex(randomblob(16))) as id,  -- Generate UUID for inbox message
    u.id as user_id,                   -- User ID from users table
    e.email_id as email_id,            -- Email ID from emails table
    CASE WHEN e.access_count > 0 THEN 1 ELSE 0 END as is_read,  -- Mark as read if accessed
    CASE WHEN e.self_destructed = 1 THEN 1 ELSE 0 END as is_deleted,  -- Mark as deleted if self-destructed
    e.created_at as created_at,        -- Use email creation time
    e.created_at as updated_at         -- Use email creation time for initial update
FROM emails e
INNER JOIN users u ON e.recipient = u.email
WHERE e.self_destructed = 0;  -- Only include non-self-destructed emails

-- Verify the migration
-- This query will show the count of inbox messages created
-- SELECT COUNT(*) as inbox_messages_created FROM inbox_messages;
