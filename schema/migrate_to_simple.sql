-- Migration script to add fallback columns to existing users table
-- This script adds the missing columns that exist in users_simple.sql but not in users.sql

-- Add fallback_email column if it doesn't exist
ALTER TABLE users ADD COLUMN fallback_email TEXT;

-- Add fallback_token column if it doesn't exist
ALTER TABLE users ADD COLUMN fallback_token TEXT;

-- Add fallback_confirmed column if it doesn't exist
ALTER TABLE users ADD COLUMN fallback_confirmed BOOLEAN DEFAULT FALSE;

-- Add fallback_token_expiration column if it doesn't exist
ALTER TABLE users ADD COLUMN fallback_token_expiration TIMESTAMP;

-- Create index for fallback token lookups if it doesn't exist
CREATE INDEX IF NOT EXISTS idx_users_fallback_token ON users(fallback_token); 