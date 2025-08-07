-- Refresh tokens table for secure session management
-- Stores refresh tokens with expiration and user association

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id TEXT PRIMARY KEY,                    -- UUID of the refresh token
    user_id TEXT NOT NULL,                 -- Associated user ID
    token_hash TEXT NOT NULL,              -- Hashed refresh token for security
    expires_at TIMESTAMP NOT NULL,         -- When the refresh token expires
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_used_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_revoked BOOLEAN DEFAULT FALSE,      -- Whether token has been revoked
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Indexes for efficient lookups
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_hash ON refresh_tokens(token_hash);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires ON refresh_tokens(expires_at);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_revoked ON refresh_tokens(is_revoked);

-- Clean up expired tokens periodically (optional)
-- This can be run by a scheduled job
-- DELETE FROM refresh_tokens WHERE expires_at < CURRENT_TIMESTAMP; 