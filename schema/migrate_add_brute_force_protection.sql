-- Migration: Add brute-force protection fields to emails table
-- Micro-Iteration 4.12: Rate Limiting & Brute-Force Protection for Email Access Attempts

-- Add brute-force protection fields
ALTER TABLE emails ADD COLUMN brute_force_failed_attempts INTEGER DEFAULT 0;
ALTER TABLE emails ADD COLUMN brute_force_last_failed_attempt DATETIME;
ALTER TABLE emails ADD COLUMN brute_force_lockout_until DATETIME;
ALTER TABLE emails ADD COLUMN brute_force_max_attempts INTEGER DEFAULT 3;
ALTER TABLE emails ADD COLUMN brute_force_lockout_duration_minutes INTEGER DEFAULT 15;

-- Add index for brute-force protection queries
CREATE INDEX IF NOT EXISTS idx_emails_brute_force ON emails(brute_force_failed_attempts, brute_force_lockout_until);

-- Add comment to document the new fields
-- brute_force_failed_attempts: Count of consecutive failed access attempts
-- brute_force_last_failed_attempt: Timestamp of the last failed attempt
-- brute_force_lockout_until: Timestamp until which access is locked out (NULL if not locked)
-- brute_force_max_attempts: Maximum failed attempts before lockout (default: 3)
-- brute_force_lockout_duration_minutes: Lockout duration in minutes (default: 15)
-- 
-- Brute-force protection works independently of existing failed_attempts tracking
-- and applies to all types of security failures (MFA, password, geolocation, etc.)
