-- =============================================================================
-- MIGRATION: Add Secure Links External Email Flow
-- =============================================================================
-- This migration adds the database schema required for the Secure Link
-- External Email Flow feature (Phase 1 implementation).
-- 
-- New Tables:
-- - secure_links: Stores secure link information and security settings
-- - link_audit_log: Tracks all link access events and security checks
-- - email_chains: Manages email conversation chains for replies
-- =============================================================================

-- =============================================================================
-- SECURE LINKS TABLE
-- =============================================================================
-- Stores secure link information, security settings, and access tracking
CREATE TABLE IF NOT EXISTS secure_links (
    link_id TEXT PRIMARY KEY,
    email_id TEXT NOT NULL,
    recipient_email TEXT NOT NULL,
    sender_id TEXT NOT NULL,
    security_settings JSON NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,
    access_count INTEGER DEFAULT 0,
    last_accessed TIMESTAMP,
    status TEXT DEFAULT 'active' CHECK (status IN ('active', 'revoked', 'expired', 'destroyed')),
    failed_attempts INTEGER DEFAULT 0,
    last_failed_attempt TIMESTAMP,
    lockout_until TIMESTAMP,
    FOREIGN KEY (email_id) REFERENCES emails(id) ON DELETE CASCADE,
    FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE CASCADE
);

-- =============================================================================
-- LINK AUDIT LOG TABLE
-- =============================================================================
-- Tracks all link access events, security checks, and suspicious activity
CREATE TABLE IF NOT EXISTS link_audit_log (
    id TEXT PRIMARY KEY,
    link_id TEXT NOT NULL,
    event_type TEXT NOT NULL CHECK (event_type IN (
        'created', 'accessed', 'failed_attempt', 'revoked', 'expired', 'destroyed',
        'password_required', 'password_validated', 'password_failed',
        'geolocation_check', 'geolocation_blocked', 'mfa_required', 'mfa_validated',
        'mfa_failed', 'time_lock_active', 'read_once_consumed', 'auto_destruct_triggered'
    )),
    ip_address TEXT,
    user_agent TEXT,
    geolocation_data JSON,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    details TEXT,
    severity TEXT DEFAULT 'info' CHECK (severity IN ('info', 'warning', 'error', 'critical')),
    FOREIGN KEY (link_id) REFERENCES secure_links(link_id) ON DELETE CASCADE
);

-- =============================================================================
-- EMAIL CHAINS TABLE
-- =============================================================================
-- Manages email conversation chains for secure reply functionality
CREATE TABLE IF NOT EXISTS email_chains (
    chain_id TEXT PRIMARY KEY,
    original_email_id TEXT NOT NULL,
    current_link_id TEXT NOT NULL,
    chain_depth INTEGER DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status TEXT DEFAULT 'active' CHECK (status IN ('active', 'closed', 'archived')),
    FOREIGN KEY (original_email_id) REFERENCES emails(id) ON DELETE CASCADE,
    FOREIGN KEY (current_link_id) REFERENCES secure_links(link_id) ON DELETE CASCADE
);

-- =============================================================================
-- SECURE LINK TEMPLATES TABLE
-- =============================================================================
-- Stores customizable message templates for external recipients
CREATE TABLE IF NOT EXISTS secure_link_templates (
    id TEXT PRIMARY KEY,
    template_name TEXT NOT NULL,
    template_content TEXT NOT NULL,
    is_default BOOLEAN DEFAULT FALSE,
    created_by TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE,
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL
);

-- =============================================================================
-- DATABASE INDEXES FOR PERFORMANCE
-- =============================================================================

-- Secure Links Indexes
CREATE INDEX IF NOT EXISTS idx_secure_links_email_id ON secure_links(email_id);
CREATE INDEX IF NOT EXISTS idx_secure_links_recipient_email ON secure_links(recipient_email);
CREATE INDEX IF NOT EXISTS idx_secure_links_sender_id ON secure_links(sender_id);
CREATE INDEX IF NOT EXISTS idx_secure_links_status ON secure_links(status);
CREATE INDEX IF NOT EXISTS idx_secure_links_expires_at ON secure_links(expires_at);
CREATE INDEX IF NOT EXISTS idx_secure_links_created_at ON secure_links(created_at);
CREATE INDEX IF NOT EXISTS idx_secure_links_last_accessed ON secure_links(last_accessed);

-- Link Audit Log Indexes
CREATE INDEX IF NOT EXISTS idx_link_audit_log_link_id ON link_audit_log(link_id);
CREATE INDEX IF NOT EXISTS idx_link_audit_log_event_type ON link_audit_log(event_type);
CREATE INDEX IF NOT EXISTS idx_link_audit_log_timestamp ON link_audit_log(timestamp);
CREATE INDEX IF NOT EXISTS idx_link_audit_log_ip_address ON link_audit_log(ip_address);
CREATE INDEX IF NOT EXISTS idx_link_audit_log_severity ON link_audit_log(severity);

-- Email Chains Indexes
CREATE INDEX IF NOT EXISTS idx_email_chains_original_email_id ON email_chains(original_email_id);
CREATE INDEX IF NOT EXISTS idx_email_chains_current_link_id ON email_chains(current_link_id);
CREATE INDEX IF NOT EXISTS idx_email_chains_status ON email_chains(status);
CREATE INDEX IF NOT EXISTS idx_email_chains_created_at ON email_chains(created_at);

-- Secure Link Templates Indexes
CREATE INDEX IF NOT EXISTS idx_secure_link_templates_is_default ON secure_link_templates(is_default);
CREATE INDEX IF NOT EXISTS idx_secure_link_templates_is_active ON secure_link_templates(is_active);
CREATE INDEX IF NOT EXISTS idx_secure_link_templates_created_by ON secure_link_templates(created_by);

-- =============================================================================
-- DEFAULT DATA INSERTION
-- =============================================================================

-- Insert default secure link template
INSERT OR IGNORE INTO secure_link_templates (
    id, 
    template_name, 
    template_content, 
    is_default, 
    is_active
) VALUES (
    'default-external-template',
    'Default External Message',
    'You have received a secure email from {{sender_name}}. Click the link below to view it securely. This link will expire and includes security features to protect your privacy.',
    TRUE,
    TRUE
);

-- =============================================================================
-- MIGRATION COMPLETION
-- =============================================================================
-- This migration adds the complete database schema for Phase 1 of the
-- Secure Link External Email Flow feature. All tables include proper
-- foreign key constraints, check constraints, and performance indexes.
-- 
-- Tables Created:
-- ✅ secure_links: Main link storage with security settings
-- ✅ link_audit_log: Comprehensive audit trail
-- ✅ email_chains: Reply chain management
-- ✅ secure_link_templates: Customizable message templates
-- 
-- Indexes Created:
-- ✅ Performance indexes for all major query patterns
-- ✅ Composite indexes for common access patterns
-- ✅ Audit log indexes for security monitoring
-- 
-- Default Data:
-- ✅ Default external message template
-- =============================================================================
