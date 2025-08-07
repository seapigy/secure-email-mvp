-- =============================================================================
-- SECURE EMAIL MVP - DATABASE SCHEMA
-- =============================================================================
-- This schema defines the database structure for the Secure Email MVP application.
-- It includes tables for user management, email storage, access tracking, and
-- folder organization.
-- =============================================================================

-- Users table - Stores user account information and authentication data
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    totp_secret TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Emails table - Stores encrypted email content and metadata
CREATE TABLE IF NOT EXISTS emails (
    id TEXT PRIMARY KEY,
    sender_id TEXT NOT NULL,
    recipient_email TEXT NOT NULL,
    subject TEXT,
    encrypted_content TEXT NOT NULL,
    access_password_hash TEXT,
    geolocation_circles TEXT,
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (sender_id) REFERENCES users(id)
);

-- Access attempts table - Tracks failed access attempts for security monitoring
CREATE TABLE IF NOT EXISTS access_attempts (
    id TEXT PRIMARY KEY,
    email_id TEXT NOT NULL,
    ip_address TEXT NOT NULL,
    attempt_count INTEGER DEFAULT 0,
    last_attempt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (email_id) REFERENCES emails(id)
);

-- Folders table - User-defined email organization folders
CREATE TABLE IF NOT EXISTS folders (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    name TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- Email folders mapping - Many-to-many relationship between emails and folders
CREATE TABLE IF NOT EXISTS email_folders (
    email_id TEXT NOT NULL,
    folder_id TEXT NOT NULL,
    PRIMARY KEY (email_id, folder_id),
    FOREIGN KEY (email_id) REFERENCES emails(id),
    FOREIGN KEY (folder_id) REFERENCES folders(id)
);

-- =============================================================================
-- DATABASE INDEXES
-- =============================================================================
-- Indexes for optimizing query performance
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_emails_sender ON emails(sender_id);
CREATE INDEX IF NOT EXISTS idx_emails_recipient ON emails(recipient_email);
CREATE INDEX IF NOT EXISTS idx_emails_expires ON emails(expires_at);
CREATE INDEX IF NOT EXISTS idx_access_attempts_email ON access_attempts(email_id);
CREATE INDEX IF NOT EXISTS idx_folders_user ON folders(user_id); 