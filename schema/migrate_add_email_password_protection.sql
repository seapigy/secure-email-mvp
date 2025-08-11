-- Migration: Add email password protection fields to emails table
-- Micro-Iteration 4.14: Password Protection for Email Access

-- Add password protection fields
ALTER TABLE emails ADD COLUMN is_password_protected BOOLEAN DEFAULT FALSE;
ALTER TABLE emails ADD COLUMN password_hash TEXT;
ALTER TABLE emails ADD COLUMN password_salt TEXT;

-- Add index for password protection queries
CREATE INDEX IF NOT EXISTS idx_emails_password_protection ON emails(is_password_protected);

-- Add comment to document the new fields
-- is_password_protected: Boolean flag indicating if email requires password for access
-- password_hash: Argon2id hash of the email password (stored as base64)
-- password_salt: Random salt used for password hashing (stored as base64)
--
-- Password protection works alongside existing security features:
-- - MFA (Multi-Factor Authentication)
-- - Geolocation restrictions
-- - Per-email brute-force protection (Micro-Iteration 4.12)
-- - IP-based tracking and lockout (Micro-Iteration 4.13)
--
-- Security flow order:
-- 1. Authentication Check
-- 2. IP-Based Lockout Check (Micro-Iteration 4.13)
-- 3. Geolocation Check (if restrictions set)
-- 4. Per-Email Brute-Force Lockout Check (Micro-Iteration 4.12)
-- 5. Password Check (if password-protected) - NEW
-- 6. MFA Check (if enabled)
-- 7. Email Decryption
