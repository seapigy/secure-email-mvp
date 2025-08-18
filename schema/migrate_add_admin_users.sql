-- Admin Users Migration: Secure Admin Authentication System
-- This migration adds admin user management with strong authentication, MFA, and audit logging

-- Admin users table: stores admin authentication and role information
CREATE TABLE IF NOT EXISTS admin_users (
    id TEXT PRIMARY KEY,                    -- UUID for admin user
    email TEXT NOT NULL UNIQUE,             -- Admin email address (unique constraint)
    password_hash TEXT NOT NULL,            -- Argon2id hash of password
    totp_secret TEXT,                       -- TOTP secret for 2FA
    totp_enabled BOOLEAN DEFAULT FALSE,     -- Whether 2FA is enabled
    role TEXT NOT NULL DEFAULT 'root_admin', -- Role: 'root_admin', 'full_admin', 'read_only_admin'
    is_active BOOLEAN DEFAULT TRUE,         -- Whether admin account is active
    last_login DATETIME,                    -- Last successful login timestamp
    failed_login_attempts INTEGER DEFAULT 0, -- Number of consecutive failed login attempts
    locked_until DATETIME,                  -- Account lockout until timestamp
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    created_by TEXT,                        -- UUID of admin who created this account
    FOREIGN KEY (created_by) REFERENCES admin_users(id) ON DELETE SET NULL
);

-- Admin invitation keys table: for secure admin account creation
CREATE TABLE IF NOT EXISTS admin_invitation_keys (
    id TEXT PRIMARY KEY,                    -- UUID for invitation
    email TEXT NOT NULL,                    -- Email address for the invitation
    invitation_token TEXT NOT NULL UNIQUE,  -- Secure invitation token
    role TEXT NOT NULL DEFAULT 'full_admin', -- Role to assign to new admin
    expires_at DATETIME NOT NULL,           -- When invitation expires
    max_uses INTEGER DEFAULT 1,             -- Maximum number of uses (default 1)
    current_uses INTEGER DEFAULT 0,         -- Current number of uses
    created_by TEXT NOT NULL,               -- UUID of admin who created invitation
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (created_by) REFERENCES admin_users(id) ON DELETE CASCADE
);

-- Admin audit logs table: comprehensive logging of all admin actions
CREATE TABLE IF NOT EXISTS admin_audit_logs (
    id TEXT PRIMARY KEY,                    -- UUID for audit log entry
    admin_id TEXT,                          -- UUID of admin who performed action
    action TEXT NOT NULL,                   -- Action performed (login, logout, invite, revoke, etc.)
    resource_type TEXT,                     -- Type of resource affected (user, email, system, etc.)
    resource_id TEXT,                       -- UUID of resource affected
    details TEXT,                           -- JSON details of the action
    ip_address TEXT,                        -- IP address of admin
    user_agent TEXT,                        -- User agent string
    success BOOLEAN NOT NULL,               -- Whether action was successful
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (admin_id) REFERENCES admin_users(id) ON DELETE SET NULL
);

-- Admin sessions table: for secure session management
CREATE TABLE IF NOT EXISTS admin_sessions (
    id TEXT PRIMARY KEY,                    -- UUID for session
    admin_id TEXT NOT NULL,                 -- UUID of admin user
    session_token TEXT NOT NULL UNIQUE,     -- JWT session token
    refresh_token TEXT NOT NULL UNIQUE,     -- Refresh token for session renewal
    expires_at DATETIME NOT NULL,           -- When session expires
    ip_address TEXT,                        -- IP address where session was created
    user_agent TEXT,                        -- User agent string
    is_active BOOLEAN DEFAULT TRUE,         -- Whether session is active
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (admin_id) REFERENCES admin_users(id) ON DELETE CASCADE
);

-- Indexes for performance and security
CREATE INDEX IF NOT EXISTS idx_admin_users_email ON admin_users(email);
CREATE INDEX IF NOT EXISTS idx_admin_users_role ON admin_users(role);
CREATE INDEX IF NOT EXISTS idx_admin_users_active ON admin_users(is_active);
CREATE INDEX IF NOT EXISTS idx_admin_invitation_token ON admin_invitation_keys(invitation_token);
CREATE INDEX IF NOT EXISTS idx_admin_invitation_email ON admin_invitation_keys(email);
CREATE INDEX IF NOT EXISTS idx_admin_invitation_expires ON admin_invitation_keys(expires_at);
CREATE INDEX IF NOT EXISTS idx_admin_audit_admin_id ON admin_audit_logs(admin_id);
CREATE INDEX IF NOT EXISTS idx_admin_audit_action ON admin_audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_admin_audit_created_at ON admin_audit_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_admin_sessions_token ON admin_sessions(session_token);
CREATE INDEX IF NOT EXISTS idx_admin_sessions_refresh ON admin_sessions(refresh_token);
CREATE INDEX IF NOT EXISTS idx_admin_sessions_admin_id ON admin_sessions(admin_id);
CREATE INDEX IF NOT EXISTS idx_admin_sessions_expires ON admin_sessions(expires_at);

-- Trigger to update updated_at timestamp
CREATE TRIGGER IF NOT EXISTS update_admin_users_updated_at 
    AFTER UPDATE ON admin_users
    FOR EACH ROW
    BEGIN
        UPDATE admin_users SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;
