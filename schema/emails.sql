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

    -- Encryption nonce for AES-256-GCM
    encryption_nonce TEXT NOT NULL,

    -- Encryption auth tag for AES-256-GCM
    encryption_auth_tag TEXT NOT NULL,

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

    -- Self-destruct flag
    self_destructed INTEGER DEFAULT 0,

    -- 🔒 Enhanced Security Features (Micro-Iterations) --

    -- Simple geolocation restrictions (Micro-Iteration 4.10)
    allowed_city TEXT,
    allowed_country TEXT,

    -- Enhanced geolocation verification (Micro-Iteration 4.15)
    geo_verification_type TEXT CHECK (geo_verification_type IN ('none', 'country', 'city', 'city_country')) DEFAULT 'none',
    geo_city TEXT,
    geo_country TEXT,

    -- MFA settings (Micro-Iteration 4.12)
    require_mfa INTEGER DEFAULT 0,
    mfa_type TEXT CHECK (mfa_type IN ('TOTP', 'EMAIL_CODE')),
    encrypted_totp_secret TEXT,
    mfa_failed_attempts INTEGER DEFAULT 0,
    mfa_locked_until DATETIME,

    -- Brute-force protection (Micro-Iteration 4.12)
    brute_force_failed_attempts INTEGER DEFAULT 0,
    brute_force_last_failed_attempt DATETIME,
    brute_force_lockout_until DATETIME,
    brute_force_max_attempts INTEGER DEFAULT 3,
    brute_force_lockout_duration_minutes INTEGER DEFAULT 15,

    -- Password protection (Micro-Iteration 4.14)
    is_password_protected BOOLEAN DEFAULT FALSE,
    password_salt TEXT,

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

-- Enhanced security indexes
CREATE INDEX IF NOT EXISTS idx_emails_simple_geolocation ON emails(allowed_city, allowed_country);
CREATE INDEX IF NOT EXISTS idx_emails_geo_verification ON emails(geo_verification_type, geo_city, geo_country);
CREATE INDEX IF NOT EXISTS idx_emails_mfa ON emails(require_mfa, mfa_type);
CREATE INDEX IF NOT EXISTS idx_emails_brute_force ON emails(brute_force_failed_attempts, brute_force_lockout_until);
CREATE INDEX IF NOT EXISTS idx_emails_password_protection ON emails(is_password_protected);

-- Trigger to update updated_at timestamp
CREATE TRIGGER IF NOT EXISTS update_emails_updated_at 
    AFTER UPDATE ON emails
    FOR EACH ROW
BEGIN
    UPDATE emails SET updated_at = CURRENT_TIMESTAMP WHERE email_id = NEW.email_id;
END; 