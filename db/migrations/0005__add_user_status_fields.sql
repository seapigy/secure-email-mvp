-- Migration 0005: Add user status fields
-- Adds is_active, email_verified, and updated_at fields to users table
-- SQLite-compatible version: no non-constant defaults in ALTER TABLE

-- 1) Add columns if they do not exist (without defaults)
ALTER TABLE users ADD COLUMN is_active INTEGER;
ALTER TABLE users ADD COLUMN email_verified INTEGER;
ALTER TABLE users ADD COLUMN updated_at DATETIME;

-- 2) Backfill sensible defaults for existing rows
UPDATE users SET is_active = 1 WHERE is_active IS NULL;
UPDATE users SET email_verified = 0 WHERE email_verified IS NULL;
UPDATE users SET updated_at = CURRENT_TIMESTAMP WHERE updated_at IS NULL;

-- 3) Create index for active users
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);
