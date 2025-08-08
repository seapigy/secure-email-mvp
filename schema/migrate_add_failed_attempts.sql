-- Migration to add self-destruct tracking columns to emails table
-- This migration is idempotent and adds columns only if they don't exist

-- Add self_destructed column if it doesn't exist
ALTER TABLE emails ADD COLUMN self_destructed INTEGER DEFAULT 0;

-- Add deleted_at column if it doesn't exist  
ALTER TABLE emails ADD COLUMN deleted_at DATETIME;

-- Add failed_access_attempts column if it doesn't exist
ALTER TABLE emails ADD COLUMN failed_access_attempts INTEGER DEFAULT 0;

-- Add indexes for performance
CREATE INDEX IF NOT EXISTS idx_emails_self_destructed ON emails(self_destructed);
CREATE INDEX IF NOT EXISTS idx_emails_deleted_at ON emails(deleted_at);
CREATE INDEX IF NOT EXISTS idx_emails_failed_access_attempts ON emails(failed_access_attempts);
