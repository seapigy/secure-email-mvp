-- Migration: Add Multi-Factor Authentication (MFA) fields to emails table
-- Micro-Iteration 4.12: Multi-Factor Authentication (MFA) for Secure Email Access

-- Add require_mfa field (boolean - whether MFA is required for this email)
ALTER TABLE emails ADD COLUMN require_mfa INTEGER DEFAULT 0;

-- Add mfa_type field (enum: TOTP, EMAIL_CODE)
ALTER TABLE emails ADD COLUMN mfa_type TEXT CHECK (mfa_type IN ('TOTP', 'EMAIL_CODE'));

-- Add encrypted_totp_secret field (AES-256-GCM encrypted TOTP secret for TOTP-based MFA)
ALTER TABLE emails ADD COLUMN encrypted_totp_secret TEXT;

-- Add mfa_failed_attempts field (track failed MFA attempts for brute-force protection)
ALTER TABLE emails ADD COLUMN mfa_failed_attempts INTEGER DEFAULT 0;

-- Add mfa_locked_until field (timestamp when MFA is locked due to too many failed attempts)
ALTER TABLE emails ADD COLUMN mfa_locked_until DATETIME;

-- Add index for MFA-related queries
CREATE INDEX IF NOT EXISTS idx_emails_mfa ON emails(require_mfa, mfa_type);

-- Add comment to document the new fields
-- require_mfa: Boolean flag indicating if MFA is required for this email
-- mfa_type: Type of MFA required ('TOTP' for Google Authenticator/Authy, 'EMAIL_CODE' for email-based codes)
-- encrypted_totp_secret: AES-256-GCM encrypted TOTP secret (only for TOTP-based MFA)
-- mfa_failed_attempts: Counter for failed MFA attempts (max 5 before lockout)
-- mfa_locked_until: Timestamp when MFA is locked due to brute-force protection
