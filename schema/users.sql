-- Users table for secure email system
-- Stores user authentication data with Argon2 password hashes and TOTP secrets

CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,                    -- UUID for user identification
    email TEXT NOT NULL UNIQUE,             -- Email address (user@securesystem.email)
    password_hash TEXT NOT NULL,            -- Argon2 hash of password
    totp_secret TEXT NOT NULL,              -- Base32 TOTP secret for 2FA
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Index for email lookups
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Trigger to update updated_at timestamp
CREATE TRIGGER IF NOT EXISTS update_users_updated_at 
    AFTER UPDATE ON users
    FOR EACH ROW
BEGIN
    UPDATE users SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END; 