-- Migration: Add IP-based tracking and lockout functionality
-- Micro-Iteration 4.13: IP-Based Tracking & Lockout for Email Access Attempts

-- Create IP access attempts table
CREATE TABLE IF NOT EXISTS ip_access_attempts (
    ip_address TEXT PRIMARY KEY,
    failed_attempts INTEGER DEFAULT 0,
    last_attempt_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    lockout_until TIMESTAMP NULL
);

-- Create index for cleanup queries
CREATE INDEX IF NOT EXISTS idx_ip_access_attempts_last_attempt ON ip_access_attempts(last_attempt_at);

-- Create index for lockout queries
CREATE INDEX IF NOT EXISTS idx_ip_access_attempts_lockout ON ip_access_attempts(lockout_until);

-- Add comment to document the table
-- ip_access_attempts: Tracks failed access attempts by IP address
-- ip_address: Client IP address (primary key)
-- failed_attempts: Count of consecutive failed attempts from this IP
-- last_attempt_at: Timestamp of the last attempt (for cleanup)
-- lockout_until: Timestamp until which this IP is locked out (NULL if not locked)
--
-- Default configuration:
-- - 5 failed attempts within 15 minutes → 30-minute lockout
-- - Automatic cleanup of records older than 24 hours
-- - Works alongside per-email brute-force protection
