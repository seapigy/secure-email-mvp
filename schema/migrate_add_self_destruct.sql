-- Migration: Add self_destruct_after_attempts field to emails table
-- This migration adds support for the self-destruct after failed attempts feature

-- Add the new column to the emails table
ALTER TABLE emails ADD COLUMN self_destruct_after_attempts INTEGER DEFAULT 0;

-- Update existing records to have the default value
UPDATE emails SET self_destruct_after_attempts = 0 WHERE self_destruct_after_attempts IS NULL;

-- Create an index for better query performance on self-destruct emails
CREATE INDEX IF NOT EXISTS idx_emails_self_destruct ON emails(self_destruct_after_attempts);

-- Add a comment to document the new field
PRAGMA table_info(emails);
