-- Migration: Add Organizations and Enterprise Multi-Tenancy Support
-- This migration adds enterprise multi-tenancy with role-based access control

-- Create organizations table
CREATE TABLE IF NOT EXISTS organizations (
    id TEXT PRIMARY KEY,                    -- UUID for organization identification
    name TEXT NOT NULL UNIQUE,              -- Organization name (unique)
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Add organization and role fields to users table
-- Note: We'll handle this carefully to maintain backward compatibility
ALTER TABLE users ADD COLUMN organization_id TEXT;
ALTER TABLE users ADD COLUMN role TEXT CHECK (role IN ('system_admin', 'enterprise_admin', 'enterprise_user')) DEFAULT 'enterprise_user';

-- Create foreign key constraint for organization_id
CREATE INDEX IF NOT EXISTS idx_users_organization_id ON users(organization_id);

-- Create trigger to update updated_at timestamp for organizations
CREATE TRIGGER IF NOT EXISTS update_organizations_updated_at 
    AFTER UPDATE ON organizations
    FOR EACH ROW
BEGIN
    UPDATE organizations SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_organizations_name ON organizations(name);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_org_role ON users(organization_id, role);

-- Insert default system organization for existing users
INSERT OR IGNORE INTO organizations (id, name) VALUES ('system-default', 'System Default Organization');

-- Update existing users to have default role and organization
-- This ensures backward compatibility
UPDATE users SET 
    role = COALESCE(role, 'enterprise_user'),
    organization_id = COALESCE(organization_id, 'system-default')
WHERE role IS NULL OR organization_id IS NULL;

-- Create view for organization membership
CREATE VIEW IF NOT EXISTS organization_members AS
SELECT 
    u.id as user_id,
    u.email,
    u.role,
    o.id as organization_id,
    o.name as organization_name,
    u.created_at as user_created_at,
    o.created_at as organization_created_at
FROM users u
LEFT JOIN organizations o ON u.organization_id = o.id;

-- Create view for admin permissions
CREATE VIEW IF NOT EXISTS admin_permissions AS
SELECT 
    u.id as user_id,
    u.email,
    u.role,
    u.organization_id,
    o.name as organization_name,
    CASE 
        WHEN u.role = 'system_admin' THEN 1
        WHEN u.role = 'enterprise_admin' THEN 1
        ELSE 0
    END as can_manage_organizations,
    CASE 
        WHEN u.role = 'system_admin' THEN 1
        ELSE 0
    END as can_manage_all_organizations,
    CASE 
        WHEN u.role = 'system_admin' THEN 1
        WHEN u.role = 'enterprise_admin' THEN 1
        ELSE 0
    END as can_manage_users
FROM users u
LEFT JOIN organizations o ON u.organization_id = o.id;
