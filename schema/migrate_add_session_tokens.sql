-- =============================================================================
-- MICRO-ITERATION 4.15: PER-SESSION ACCESS TOKENS & ONE-TIME LINKS
-- =============================================================================
-- Migration to add per-session access tokens and one-time links functionality
--
-- This migration adds support for:
-- - email_sessions table: Stores session tokens for email access
-- - one_time_link_only: Boolean flag to require one-time use tokens
-- - Session token validation: Short-lived tokens for secure email access
-- - One-time links: Tokens that become invalid after first use
--
-- Per-session tokens reduce replay risk and make stolen links or tokens useless
-- by requiring a fresh token for each email access session.
-- =============================================================================

-- Add one_time_link_only column to emails table
ALTER TABLE emails ADD COLUMN one_time_link_only BOOLEAN DEFAULT FALSE;

-- Create email_sessions table
CREATE TABLE IF NOT EXISTS email_sessions (
    session_id TEXT PRIMARY KEY,  -- UUID for the session record
    email_id TEXT NOT NULL,  -- Foreign key to emails table
    token_hash TEXT NOT NULL,  -- Argon2id hash of the session token
    expires_at DATETIME NOT NULL,  -- When the session token expires
    used BOOLEAN DEFAULT FALSE,  -- Whether the token has been used (for one-time mode)
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,  -- When the session was created
    user_agent TEXT,  -- User-Agent string for audit purposes
    ip_address TEXT,  -- IP address for audit purposes
    
    -- Foreign key constraint
    FOREIGN KEY (email_id) REFERENCES emails(email_id) ON DELETE CASCADE
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_email_sessions_email_id ON email_sessions(email_id);
CREATE INDEX IF NOT EXISTS idx_email_sessions_token_hash ON email_sessions(token_hash);
CREATE INDEX IF NOT EXISTS idx_email_sessions_email_hash ON email_sessions(email_id, token_hash);
CREATE INDEX IF NOT EXISTS idx_email_sessions_expires_at ON email_sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_email_sessions_used ON email_sessions(used);

-- Add index for one_time_link_only queries
CREATE INDEX IF NOT EXISTS idx_emails_one_time_link_only ON emails(one_time_link_only);

-- Add comments for documentation
PRAGMA table_info(emails);
PRAGMA table_info(email_sessions);

-- Verify the migration was applied successfully
SELECT
    name,
    type,
    "notnull",
    dflt_value,
    pk
FROM pragma_table_info('emails')
WHERE name = 'one_time_link_only'
ORDER BY name;

-- Verify email_sessions table structure
SELECT
    name,
    type,
    "notnull",
    dflt_value,
    pk
FROM pragma_table_info('email_sessions')
ORDER BY name;
