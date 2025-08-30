-- =============================================================================
-- Signup V2 Support Migration
-- =============================================================================
-- This migration adds support for the new signup endpoint that supports
-- Free, Paid, and Company plans with proper status tracking.
-- =============================================================================

-- Add new columns to the users table for signup v2 support
ALTER TABLE users ADD COLUMN plan TEXT DEFAULT 'free';
ALTER TABLE users ADD COLUMN company_code TEXT;
ALTER TABLE users ADD COLUMN status TEXT DEFAULT 'pending_verification';

-- Create indexes for the new columns
CREATE INDEX IF NOT EXISTS idx_users_plan ON users(plan);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);
CREATE INDEX IF NOT EXISTS idx_users_company_code ON users(company_code);

-- Create a view for signup statistics
CREATE VIEW IF NOT EXISTS signup_statistics AS
SELECT 
    plan,
    status,
    COUNT(*) as user_count,
    COUNT(CASE WHEN created_at >= datetime('now', '-24 hours') THEN 1 END) as new_last_24h,
    COUNT(CASE WHEN created_at >= datetime('now', '-7 days') THEN 1 END) as new_last_7d,
    COUNT(CASE WHEN created_at >= datetime('now', '-30 days') THEN 1 END) as new_last_30d
FROM users 
GROUP BY plan, status;

-- Create a view for company signup statistics
CREATE VIEW IF NOT EXISTS company_signup_statistics AS
SELECT 
    company_code,
    COUNT(*) as user_count,
    COUNT(CASE WHEN status = 'pending_verification' THEN 1 END) as pending_verification,
    COUNT(CASE WHEN status = 'verified' THEN 1 END) as verified,
    COUNT(CASE WHEN status = 'active' THEN 1 END) as active,
    MIN(created_at) as first_signup,
    MAX(created_at) as latest_signup
FROM users 
WHERE plan = 'company' AND company_code IS NOT NULL
GROUP BY company_code;

-- Add constraints to ensure data integrity
-- Ensure plan is one of the valid options
CREATE TRIGGER IF NOT EXISTS validate_plan
BEFORE INSERT ON users
FOR EACH ROW
BEGIN
    SELECT CASE 
        WHEN NEW.plan NOT IN ('free', 'paid', 'company') 
        THEN RAISE(ABORT, 'Invalid plan type')
    END;
END;

-- Ensure company code is provided for company plans
CREATE TRIGGER IF NOT EXISTS validate_company_code
BEFORE INSERT ON users
FOR EACH ROW
BEGIN
    SELECT CASE 
        WHEN NEW.plan = 'company' AND (NEW.company_code IS NULL OR NEW.company_code = '') 
        THEN RAISE(ABORT, 'Company code required for company plans')
    END;
END;

-- Ensure status is one of the valid options
CREATE TRIGGER IF NOT EXISTS validate_status
BEFORE INSERT ON users
FOR EACH ROW
BEGIN
    SELECT CASE 
        WHEN NEW.status NOT IN ('pending_verification', 'verified', 'active', 'suspended', 'deleted') 
        THEN RAISE(ABORT, 'Invalid status')
    END;
END;
