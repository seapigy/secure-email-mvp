-- Migration 0002: Inbox messages table
-- Creates the inbox_messages table for better user isolation and inbox functionality

-- Create inbox_messages table
CREATE TABLE IF NOT EXISTS inbox_messages (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    email_id TEXT NOT NULL,
    is_read BOOLEAN DEFAULT FALSE,
    is_deleted BOOLEAN DEFAULT FALSE,
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

-- Add additional email indexes for inbox functionality
CREATE INDEX IF NOT EXISTS idx_emails_recipient_created ON emails(recipient, created_at);
