-- Migration: Add V2 signup columns to users table
-- This migration is idempotent and safe to run multiple times

-- Add plan column (TEXT, required, default 'free')
ALTER TABLE users ADD COLUMN plan TEXT DEFAULT 'free' NOT NULL;

-- Add company_code column (TEXT, nullable)
ALTER TABLE users ADD COLUMN company_code TEXT;

-- Add updated_at column (TIMESTAMP, auto-updated)
ALTER TABLE users ADD COLUMN updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

-- Add fallback_email column (TEXT, required, default '')
ALTER TABLE users ADD COLUMN fallback_email TEXT DEFAULT '' NOT NULL;

-- Add fallback_confirmed column (BOOLEAN, default 0)
ALTER TABLE users ADD COLUMN fallback_confirmed BOOLEAN DEFAULT 0;

-- Create index on plan for efficient queries
CREATE INDEX IF NOT EXISTS idx_users_plan ON users(plan);

-- Create index on company_code for efficient queries
CREATE INDEX IF NOT EXISTS idx_users_company_code ON users(company_code);

-- Create index on updated_at for efficient queries
CREATE INDEX IF NOT EXISTS idx_users_updated_at ON users(updated_at);

-- Create index on fallback_email for efficient queries
CREATE INDEX IF NOT EXISTS idx_users_fallback_email ON users(fallback_email);

-- Create index on fallback_confirmed for efficient queries
CREATE INDEX IF NOT EXISTS idx_users_fallback_confirmed ON users(fallback_confirmed);
