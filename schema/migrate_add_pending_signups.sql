-- Migration: Add pending signups table for fallback email verification
-- This allows us to verify fallback emails before creating actual user accounts

-- Pending signups table for storing unconfirmed signup requests
CREATE TABLE IF NOT EXISTS pending_signups (
    id TEXT PRIMARY KEY,                    -- UUID for the pending signup
    email TEXT NOT NULL UNIQUE,             -- Email address (user@securesystem.email)
    password_hash TEXT NOT NULL,            -- Argon2 hash of password
    totp_secret TEXT NOT NULL,              -- Base32 TOTP secret for 2FA
    fallback_email TEXT NOT NULL,           -- Fallback email address
    fallback_token TEXT NOT NULL,           -- HMAC token for fallback verification
    fallback_token_expiration DATETIME NOT NULL, -- When the token expires
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Index for email lookups
CREATE INDEX IF NOT EXISTS idx_pending_signups_email ON pending_signups(email);

-- Index for fallback token lookups
CREATE INDEX IF NOT EXISTS idx_pending_signups_fallback_token ON pending_signups(fallback_token);

-- Index for expiration cleanup
CREATE INDEX IF NOT EXISTS idx_pending_signups_expiration ON pending_signups(fallback_token_expiration);

-- Trigger to update updated_at timestamp
CREATE TRIGGER IF NOT EXISTS update_pending_signups_updated_at 
    AFTER UPDATE ON pending_signups
    FOR EACH ROW
BEGIN
    UPDATE pending_signups SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- Add cleanup job for expired pending signups (older than 24 hours)
-- This can be run periodically to clean up expired signup attempts
-- DELETE FROM pending_signups WHERE fallback_token_expiration < datetime('now', '-24 hours');
