-- Simplified users table for signup handler with fallback email verification
-- Stores basic user authentication data and fallback email verification

CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY,                 -- Auto-incrementing integer ID
    email TEXT NOT NULL UNIQUE,             -- Email address (unique constraint)
    password TEXT NOT NULL,                 -- Bcrypt hash of password
    fallback_email TEXT,                    -- Fallback email for recovery
    fallback_token TEXT,                    -- Secure token for fallback verification
    fallback_confirmed BOOLEAN DEFAULT FALSE, -- Whether fallback email is confirmed
    fallback_token_expiration TIMESTAMP,    -- When the fallback token expires
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Index for email lookups
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Index for fallback token lookups
CREATE INDEX IF NOT EXISTS idx_users_fallback_token ON users(fallback_token); 