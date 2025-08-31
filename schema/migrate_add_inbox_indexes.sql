-- Migration: Add inbox indexes for improved performance and user isolation
-- This migration adds indexes to support inbox functionality where users can view emails sent TO them

-- Add index on recipient field for inbox queries
-- This allows fast lookups of emails sent to a specific user
CREATE INDEX IF NOT EXISTS idx_emails_recipient ON emails(recipient);

-- Add composite index on recipient and created_at for inbox queries with sorting
-- This optimizes queries that list inbox emails ordered by creation date
CREATE INDEX IF NOT EXISTS idx_emails_recipient_created ON emails(recipient, created_at);

-- Add index on recipient and email_id for specific email lookups in inbox
-- This optimizes queries that fetch a specific email from a user's inbox
CREATE INDEX IF NOT EXISTS idx_emails_recipient_email_id ON emails(recipient, email_id);

-- Add index on recipient and self_destructed for soft delete operations
-- This optimizes queries that filter out deleted emails from inbox
CREATE INDEX IF NOT EXISTS idx_emails_recipient_deleted ON emails(recipient, self_destructed);

-- Add index on recipient and access_count for read status queries
-- This optimizes queries that determine if an email has been read
CREATE INDEX IF NOT EXISTS idx_emails_recipient_access_count ON emails(recipient, access_count);

-- Add index on recipient and expires_at for expiration status queries
-- This optimizes queries that determine if an email has expired
CREATE INDEX IF NOT EXISTS idx_emails_recipient_expires ON emails(recipient, expires_at);

-- Verify indexes were created successfully
-- This query will show all indexes on the emails table
-- SELECT name, sql FROM sqlite_master WHERE type='index' AND tbl_name='emails';
