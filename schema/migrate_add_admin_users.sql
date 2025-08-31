-- Migration: Add admin_users table
-- This migration creates the admin_users table for admin authentication

CREATE TABLE IF NOT EXISTS admin_users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL, -- hashed password
    role TEXT NOT NULL DEFAULT 'admin',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create index on email for fast lookups
CREATE INDEX IF NOT EXISTS idx_admin_users_email ON admin_users(email);

-- Create index on role for role-based queries
CREATE INDEX IF NOT EXISTS idx_admin_users_role ON admin_users(role);

-- Insert a default admin user (password: admin123456)
-- In production, this should be changed immediately
INSERT OR IGNORE INTO admin_users (email, password, role, created_at) VALUES (
    'admin@securesystem.email',
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', -- bcrypt hash of 'admin123456'
    'admin',
    CURRENT_TIMESTAMP
);
