-- Migration: Add read-once consumed tracking (Micro-Iteration 4.10)
-- This migration adds fields to track when read-once emails have been consumed
-- and optionally trigger immediate deletion after first read

-- Add read-once consumed timestamp (UTC unix epoch seconds when consumed)
ALTER TABLE emails ADD COLUMN read_once_consumed_at INTEGER;

-- Add optional device fingerprint for auditing (do not store PII)
ALTER TABLE emails ADD COLUMN read_once_consumer_device TEXT;

-- Add optional flag to trigger immediate deletion after first read
ALTER TABLE emails ADD COLUMN self_destruct_on_read_once BOOLEAN DEFAULT FALSE;

-- Create indexes for performance on new read-once fields
CREATE INDEX IF NOT EXISTS idx_emails_read_once_consumed_at ON emails(read_once_consumed_at);
CREATE INDEX IF NOT EXISTS idx_emails_self_destruct_on_read_once ON emails(self_destruct_on_read_once);

-- Create unique index for optimistic locking on read-once consumption
-- This ensures only one successful read can mark an email as consumed
CREATE UNIQUE INDEX IF NOT EXISTS idx_emails_read_once_atomic 
ON emails(id, read_once_consumed_at) 
WHERE read_once = TRUE AND read_once_consumed_at IS NULL;
