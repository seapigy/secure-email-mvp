-- Migration 0004: Add name field to users table
-- Adds a name field for user identification

-- Add name column to users table
ALTER TABLE users ADD COLUMN name TEXT;

-- Update existing users to have a default name based on email
UPDATE users SET name = SUBSTR(email, 1, INSTR(email, '@') - 1) WHERE name IS NULL;

-- Make name required for new users (but not existing ones to avoid breaking existing data)
-- Note: We'll enforce this at the application level
