-- Migration: Add TOTP fields to users table
-- Micro-Iteration 4.5: Multi-Factor Authentication for User Login

-- Add TOTP secret field for user authentication
ALTER TABLE users ADD COLUMN totp_secret TEXT;

-- Add password_hash field (rename from password to match auth package expectations)
ALTER TABLE users ADD COLUMN password_hash TEXT;

-- Add index for TOTP-related queries
CREATE INDEX IF NOT EXISTS idx_users_totp ON users(email, totp_secret);

-- Add comment to document the new fields
-- totp_secret: Base32-encoded TOTP secret for 2FA authentication
-- password_hash: Argon2 hash of password (replaces the old password field)

-- Note: Existing users will need to have their passwords rehashed and TOTP secrets generated
-- This migration should be run after user data migration





















