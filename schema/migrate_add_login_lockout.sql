-- Migration: Add login lockout fields to users table for rate limiting/account lockout
ALTER TABLE users ADD COLUMN failed_login_attempts INTEGER DEFAULT 0;
ALTER TABLE users ADD COLUMN last_failed_login TIMESTAMP;
ALTER TABLE users ADD COLUMN account_locked_until TIMESTAMP; 