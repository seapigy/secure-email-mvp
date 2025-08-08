-- Migration to add fail_count column to emails table
-- This migration is idempotent and adds the column only if it doesn't exist

-- Add fail_count column if it doesn't exist
ALTER TABLE emails ADD COLUMN fail_count INTEGER DEFAULT 0;

-- Add index for performance
CREATE INDEX IF NOT EXISTS idx_emails_fail_count ON emails(fail_count);
