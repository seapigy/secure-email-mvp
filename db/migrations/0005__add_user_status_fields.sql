-- Migration 0005: Add user status fields
-- Adds is_active, email_verified, and updated_at fields to users table

-- Add is_active column to users table (default to true for existing users)
ALTER TABLE users ADD COLUMN is_active BOOLEAN DEFAULT TRUE;

-- Add email_verified column to users table (default to false for existing users)
ALTER TABLE users ADD COLUMN email_verified BOOLEAN DEFAULT FALSE;

-- Add updated_at column to users table
ALTER TABLE users ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

-- Update existing users to have verified email (since they exist)
UPDATE users SET email_verified = TRUE WHERE email_verified IS NULL;

-- Create index for active users
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);
