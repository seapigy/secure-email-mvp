-- Emails table for secure email system
-- Stores encrypted email metadata and access control information

CREATE TABLE IF NOT EXISTS emails (
    -- Unique identifier for the email
    email_id TEXT PRIMARY KEY,

    -- Who is sending the message (references users table)
    sender_id TEXT NOT NULL,

    -- Email address of the recipient (not necessarily a user)
    recipient TEXT NOT NULL,

    -- Subject line of the message (stored in plaintext)
    subject TEXT,

    -- URL to the encrypted blob (stored in Cloudflare R2)
    encrypted_blob_url TEXT NOT NULL,

    -- Encrypted data key used to decrypt the blob
    encrypted_key TEXT NOT NULL,

    -- Compression algorithm used before encryption
    compression_algo TEXT DEFAULT 'gzip',

    -- SHA-256 hash of the original (plaintext) content
    sha256_hash TEXT NOT NULL,

    -- 🔐 Security Features --

    -- Is password required to access this email?
    requires_password INTEGER DEFAULT 0,

    -- Argon2-hashed password (only if requires_password = 1)
    password_hash TEXT,

    -- Optional geofence restriction (JSON: lat/lng radius)
    geolocation_json TEXT,

    -- When the email auto-expires (e.g., 30 days)
    expires_at DATETIME,

    -- Burn after read: If true, email is deleted after 1 successful access
    burn_after_read INTEGER DEFAULT 0,

    -- Track how many failed attempts were made to access this email
    failed_attempts INTEGER DEFAULT 0,

    -- Max allowed failed attempts before the email self-deletes
    max_attempts INTEGER DEFAULT 3,

    -- Should the email self-destruct after max failed attempts?
    self_destruct_after_attempts INTEGER DEFAULT 0,

    -- 🔁 Interactivity Controls --

    -- Can recipient reply to sender?
    reply_enabled INTEGER DEFAULT 0,

    -- If reply is enabled, is a password required to reply?
    reply_requires_password INTEGER DEFAULT 1,

    -- Can this email be forwarded by the recipient?
    allow_forwarding INTEGER DEFAULT 0,

    -- Should the sender's metadata (IP, device) be visible to recipient?
    show_sender_metadata INTEGER DEFAULT 0,

    -- Has the sender explicitly requested metadata stripping?
    metadata_stripped INTEGER DEFAULT 1,

    -- Is this email a decoy for security monitoring (honeytoken)?
    is_honeytoken INTEGER DEFAULT 0,

    -- 🌍 Secure Link Features --

    -- Unique secure link ID (used for anonymous access)
    secure_link_id TEXT UNIQUE,

    -- When the secure link was generated
    link_created_at DATETIME,

    -- Timestamp of last successful access (if any)
    last_access_at DATETIME,

    -- Total number of access attempts (successful or failed)
    access_count INTEGER DEFAULT 0,

    -- 🕓 System Tracking --

    -- When this record was created
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    -- Optional timestamp of last update to this record
    updated_at DATETIME,

    -- Enforce foreign key relationship with users table
    FOREIGN KEY (sender_id) REFERENCES users(id)
);

-- Indexes for performance and query optimization
CREATE INDEX IF NOT EXISTS idx_emails_sender_id ON emails(sender_id);
CREATE INDEX IF NOT EXISTS idx_emails_recipient ON emails(recipient);
CREATE INDEX IF NOT EXISTS idx_emails_secure_link_id ON emails(secure_link_id);
CREATE INDEX IF NOT EXISTS idx_emails_created_at ON emails(created_at);
CREATE INDEX IF NOT EXISTS idx_emails_expires_at ON emails(expires_at);
CREATE INDEX IF NOT EXISTS idx_emails_burn_after_read ON emails(burn_after_read);
CREATE INDEX IF NOT EXISTS idx_emails_self_destruct ON emails(self_destruct_after_attempts);

-- Trigger to update updated_at timestamp
CREATE TRIGGER IF NOT EXISTS update_emails_updated_at 
    AFTER UPDATE ON emails
    FOR EACH ROW
BEGIN
    UPDATE emails SET updated_at = CURRENT_TIMESTAMP WHERE email_id = NEW.email_id;
END; 