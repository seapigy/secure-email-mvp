-- Consolidated Schema for SecureMail Backend
-- This file is read-only for review purposes
-- All changes should be made via migrations in /migrations/ directory

-- Core Authentication Tables

-- Users table (core)
CREATE TABLE users (
    id TEXT PRIMARY KEY,                       -- UUID
    username TEXT NOT NULL,
    email TEXT NOT NULL,
    hashed_password TEXT NOT NULL,             -- Argon2id output
    totp_secret_encrypted BLOB,                -- AES-256-GCM encrypted, nullable
    totp_configured BOOLEAN DEFAULT FALSE,
    recovery_codes_hashed JSON,                -- JSON array of hashed one-time codes
    public_pqc_key TEXT NULL,                  -- public key (PQC) for hybrid KEM
    public_sign_key TEXT NULL,                 -- public signing key (optional)
    encrypted_profile_blob BLOB NULL,          -- optional client-side encrypted profile
    account_type TEXT NOT NULL DEFAULT 'free', -- enum: free|premium|enterprise|admin
    account_status TEXT NOT NULL DEFAULT 'pending_verification', -- pending_verification|active|suspended|deleted
    domain TEXT NULL,                          -- user's mailbox domain (nullable until bound)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Recovery codes (one-time codes table)
CREATE TABLE recovery_codes (
    id TEXT PRIMARY KEY,           -- UUID
    user_id TEXT NOT NULL,
    code_hashed TEXT NOT NULL,     -- hashed one-time code (Argon2id or SHA256+pepper)
    used BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_recovery_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Sessions / refresh tokens (store hashed tokens)
CREATE TABLE sessions (
    id TEXT PRIMARY KEY,           -- UUID
    user_id TEXT NOT NULL,
    token_hash TEXT NOT NULL,      -- store hash of opaque token (never store raw token)
    device_info TEXT NULL,
    ip_address TEXT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_session_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Domains (for paid / enterprise)
CREATE TABLE domains (
    id TEXT PRIMARY KEY,
    domain_name TEXT NOT NULL UNIQUE,
    owner_user_id TEXT NULL,
    org_id TEXT NULL,
    verified BOOLEAN DEFAULT FALSE,
    dns_verification_token TEXT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_domain_owner FOREIGN KEY (owner_user_id) REFERENCES users(id) ON DELETE SET NULL
);

-- Mailbox identities (username@domain aliasing)
CREATE TABLE mailbox_identities (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    domain_id TEXT NOT NULL,
    email_identity TEXT NOT NULL,   -- username@domain
    is_primary BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_mailbox_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_mailbox_domain FOREIGN KEY (domain_id) REFERENCES domains(id) ON DELETE CASCADE
);

-- Audit logs
CREATE TABLE audit_logs (
    id TEXT PRIMARY KEY,
    user_id TEXT NULL,
    action TEXT NOT NULL,
    resource_type TEXT,
    resource_id TEXT,
    ip_address TEXT,
    user_agent TEXT,
    metadata JSON,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE UNIQUE INDEX idx_users_email ON users(email);
CREATE UNIQUE INDEX idx_users_username_domain ON users(username, domain);
CREATE INDEX idx_recovery_user ON recovery_codes(user_id);
CREATE INDEX idx_sessions_user ON sessions(user_id);
CREATE UNIQUE INDEX idx_mailbox_identity ON mailbox_identities(email_identity);
CREATE INDEX idx_audit_user ON audit_logs(user_id);
CREATE INDEX idx_audit_timestamp ON audit_logs(timestamp);
