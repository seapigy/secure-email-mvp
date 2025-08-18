-- ZKID Layer: Encrypted email mappings and recovery codes

-- Email mappings table: stores encrypted external emails mapped to internal user UUIDs
CREATE TABLE IF NOT EXISTS zkid_email_mappings (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL UNIQUE,
    email_hash TEXT NOT NULL UNIQUE, -- SHA-256 of normalized email with pepper
    email_ciphertext BLOB NOT NULL,
    email_nonce BLOB NOT NULL,
    email_tag BLOB NOT NULL,
    wrapped_key BLOB NOT NULL,           -- per-record data key wrapped with master key
    wrapped_key_nonce BLOB NOT NULL,
    wrapped_key_tag BLOB NOT NULL,
    fallback_email_ciphertext BLOB,
    fallback_email_nonce BLOB,
    fallback_email_tag BLOB,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_zkid_email_hash ON zkid_email_mappings(email_hash);
CREATE INDEX IF NOT EXISTS idx_zkid_user_id ON zkid_email_mappings(user_id);

-- Recovery codes table: Bitwarden-style one-time recovery codes
CREATE TABLE IF NOT EXISTS zkid_recovery_codes (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    salt BLOB NOT NULL,
    hash BLOB NOT NULL, -- Argon2id hash of code+pepper with salt
    used BOOLEAN NOT NULL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    used_at DATETIME,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_zkid_recovery_user ON zkid_recovery_codes(user_id);


