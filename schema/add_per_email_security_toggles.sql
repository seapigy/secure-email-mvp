-- Migration: Add per-email security toggles (Micro-Iteration 4.7)
-- This migration adds new security toggle fields to the emails table
-- for individualized security rules per email

-- Add new security toggle columns to emails table
-- These fields enable per-email security controls as specified in Micro-Iteration 4.7

-- Time Lock: Unix timestamp for when email becomes accessible
ALTER TABLE emails ADD COLUMN not_before INTEGER;

-- Expiration: Unix timestamp for when email expires
ALTER TABLE emails ADD COLUMN expires_at INTEGER;

-- Read Once: Burn after first successful access
ALTER TABLE emails ADD COLUMN read_once BOOLEAN DEFAULT FALSE;

-- MFA on Open: Require secondary TOTP for access
ALTER TABLE emails ADD COLUMN mfa_on_open BOOLEAN DEFAULT FALSE;

-- Decoy Secret: Argon2id hash of decoy password or TOTP
ALTER TABLE emails ADD COLUMN decoy_secret TEXT;

-- Remote Revoke: Sender can revoke access anytime
ALTER TABLE emails ADD COLUMN remote_revoke BOOLEAN DEFAULT FALSE;

-- Strip Metadata: Remove EXIF/headers from attachments
ALTER TABLE emails ADD COLUMN strip_metadata BOOLEAN DEFAULT FALSE;

-- Self Destruct Threshold: Max failed attempts before destruction
ALTER TABLE emails ADD COLUMN self_destruct_threshold INTEGER DEFAULT 3;

-- Geo Rules Reference: JSON reference to geofencing rules
ALTER TABLE emails ADD COLUMN geo_rules_ref TEXT;

-- Create indexes for performance on new security fields
CREATE INDEX IF NOT EXISTS idx_emails_not_before ON emails(not_before);
CREATE INDEX IF NOT EXISTS idx_emails_expires_at ON emails(expires_at);
CREATE INDEX IF NOT EXISTS idx_emails_read_once ON emails(read_once);
CREATE INDEX IF NOT EXISTS idx_emails_mfa_on_open ON emails(mfa_on_open);
CREATE INDEX IF NOT EXISTS idx_emails_remote_revoke ON emails(remote_revoke);
CREATE INDEX IF NOT EXISTS idx_emails_self_destruct_threshold ON emails(self_destruct_threshold);

-- Add comments to document the new fields
-- Note: SQLite doesn't support column comments, but we document them here
/*
New Security Toggle Fields (Micro-Iteration 4.7):

not_before (INTEGER, nullable):
  - Unix timestamp for time lock functionality
  - Email access denied before this timestamp
  - Allows scheduling future access to emails

expires_at (INTEGER, nullable):
  - Unix timestamp for expiration
  - Email access denied after this timestamp
  - Provides absolute TTL for email access

read_once (BOOLEAN, default false):
  - Burn after first successful access
  - Email is marked as consumed after first read
  - Prevents multiple access to sensitive content

mfa_on_open (BOOLEAN, default false):
  - Require secondary TOTP authentication on email open
  - Separate from account MFA
  - Provides additional layer of protection per email

decoy_secret (TEXT, nullable):
  - Argon2id hash of decoy password or TOTP
  - Enables plausible deniability features
  - Shows alternate content if decoy credentials used

remote_revoke (BOOLEAN, default false):
  - Sender can revoke access at any time
  - Immediate access denial when enabled
  - Provides sender control over email access

strip_metadata (BOOLEAN, default false):
  - Remove EXIF data and headers from attachments
  - Privacy protection for file metadata
  - Applied during upload process

self_destruct_threshold (INTEGER, default 3):
  - Maximum failed access attempts before email destruction
  - Configurable per email (minimum 1)
  - Provides brute force protection

geo_rules_ref (TEXT, nullable):
  - JSON reference to geofencing rules
  - Stored as JSON string for flexibility
  - Not parsed or enforced in this iteration
*/
