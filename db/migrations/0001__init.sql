-- Migration 0001: Initial schema
-- Creates the core tables for the Secure Email system

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    fallback_email TEXT,
    fallback_token TEXT,
    fallback_confirmed BOOLEAN DEFAULT FALSE,
    fallback_token_expiration TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    failed_login_attempts INTEGER DEFAULT 0,
    last_failed_login TIMESTAMP,
    account_locked_until TIMESTAMP,
    phone_number TEXT,
    totp_secret TEXT,
    password_hash TEXT
);

-- Emails table
CREATE TABLE IF NOT EXISTS emails (
    email_id TEXT PRIMARY KEY,
    sender_id INTEGER NOT NULL,
    recipient TEXT NOT NULL,
    subject TEXT,
    encrypted_blob_url TEXT NOT NULL,
    encrypted_key TEXT NOT NULL,
    encryption_nonce TEXT NOT NULL,
    encryption_auth_tag TEXT NOT NULL,
    compression_algo TEXT DEFAULT 'gzip',
    sha256_hash TEXT NOT NULL,
    requires_password INTEGER DEFAULT 0,
    password_hash TEXT,
    geolocation_json TEXT,
    expires_at DATETIME,
    burn_after_read INTEGER DEFAULT 0,
    failed_attempts INTEGER DEFAULT 0,
    max_attempts INTEGER DEFAULT 3,
    self_destruct_after_attempts INTEGER DEFAULT 0,
    reply_enabled INTEGER DEFAULT 0,
    reply_requires_password INTEGER DEFAULT 1,
    allow_forwarding INTEGER DEFAULT 0,
    show_sender_metadata INTEGER DEFAULT 0,
    metadata_stripped INTEGER DEFAULT 1,
    is_honeytoken INTEGER DEFAULT 0,
    secure_link_id TEXT,
    link_created_at DATETIME,
    last_access_at DATETIME,
    access_count INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME,
    self_destructed INTEGER DEFAULT 0,
    allowed_city TEXT,
    allowed_country TEXT,
    geo_verification_type TEXT DEFAULT 'none',
    geo_city TEXT,
    geo_country TEXT,
    require_mfa INTEGER DEFAULT 0,
    mfa_type TEXT,
    encrypted_totp_secret TEXT,
    mfa_failed_attempts INTEGER DEFAULT 0,
    mfa_locked_until DATETIME,
    brute_force_failed_attempts INTEGER DEFAULT 0,
    brute_force_last_failed_attempt DATETIME,
    brute_force_lockout_until DATETIME,
    brute_force_max_attempts INTEGER DEFAULT 3,
    brute_force_lockout_duration_minutes INTEGER DEFAULT 15,
    is_password_protected BOOLEAN DEFAULT FALSE,
    password_salt TEXT,
    fail_count INTEGER DEFAULT 0,
    allowed_countries TEXT,
    allowed_cities TEXT,
    trusted_devices_only BOOLEAN DEFAULT FALSE,
    one_time_link_only BOOLEAN DEFAULT FALSE,
    recipient_id TEXT,
    first_read_at DATETIME,
    read_count INTEGER DEFAULT 0,
    read_receipt_sent BOOLEAN DEFAULT FALSE,
    expiration_alert_sent BOOLEAN DEFAULT FALSE,
    final_expiration_alert_sent BOOLEAN DEFAULT FALSE,
    enable_read_receipts BOOLEAN DEFAULT TRUE,
    enable_expiration_alerts BOOLEAN DEFAULT TRUE,
    expiration_alert_hours INTEGER DEFAULT 24,
    suspicious_flag BOOLEAN DEFAULT FALSE,
    suspicious_flag_set_at DATETIME,
    suspicious_flag_cleared_at DATETIME,
    suspicious_flag_cleared_by TEXT,
    geo_restriction_rules TEXT,
    geo_restriction_config TEXT,
    geo_restriction_enabled INTEGER DEFAULT 1,
    geo_restriction_violations INTEGER DEFAULT 0,
    geo_restriction_last_violation DATETIME
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_totp ON users(totp_secret);
CREATE INDEX IF NOT EXISTS idx_emails_sender_id ON emails(sender_id);
CREATE INDEX IF NOT EXISTS idx_emails_recipient ON emails(recipient);
CREATE INDEX IF NOT EXISTS idx_emails_secure_link_id ON emails(secure_link_id);
CREATE INDEX IF NOT EXISTS idx_emails_created_at ON emails(created_at);
CREATE INDEX IF NOT EXISTS idx_emails_expires_at ON emails(expires_at);
CREATE INDEX IF NOT EXISTS idx_emails_burn_after_read ON emails(burn_after_read);
CREATE INDEX IF NOT EXISTS idx_emails_self_destruct ON emails(self_destructed);
CREATE INDEX IF NOT EXISTS idx_emails_simple_geolocation ON emails(allowed_city, allowed_country);
CREATE INDEX IF NOT EXISTS idx_emails_geo_verification ON emails(geo_verification_type);
CREATE INDEX IF NOT EXISTS idx_emails_mfa ON emails(require_mfa);
CREATE INDEX IF NOT EXISTS idx_emails_brute_force ON emails(brute_force_lockout_until);
CREATE INDEX IF NOT EXISTS idx_emails_password_protection ON emails(is_password_protected);
CREATE INDEX IF NOT EXISTS idx_emails_fail_count ON emails(fail_count);
CREATE INDEX IF NOT EXISTS idx_emails_geolocation ON emails(allowed_countries, allowed_cities);
CREATE INDEX IF NOT EXISTS idx_emails_trusted_devices_only ON emails(trusted_devices_only);
CREATE INDEX IF NOT EXISTS idx_emails_one_time_link_only ON emails(one_time_link_only);
CREATE INDEX IF NOT EXISTS idx_emails_recipient_id ON emails(recipient_id);
CREATE INDEX IF NOT EXISTS idx_emails_read_receipts ON emails(enable_read_receipts);
CREATE INDEX IF NOT EXISTS idx_emails_expiration_alerts ON emails(enable_expiration_alerts);
CREATE INDEX IF NOT EXISTS idx_emails_suspicious_flag ON emails(suspicious_flag);
CREATE INDEX IF NOT EXISTS idx_emails_geo_restriction ON emails(geo_restriction_enabled);
