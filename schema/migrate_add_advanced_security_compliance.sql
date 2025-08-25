-- Iteration 6: Advanced Security & Compliance Controls
-- Migration to add support for DLP, watermarking, advanced expiration, and compliance logging

-- DLP Rules table for configurable data loss prevention
CREATE TABLE IF NOT EXISTS dlp_rules (
    rule_id TEXT PRIMARY KEY,
    rule_name TEXT NOT NULL,
    rule_type TEXT NOT NULL, -- 'regex', 'keyword', 'ai_pattern'
    pattern TEXT NOT NULL, -- regex pattern or keyword
    description TEXT,
    severity TEXT DEFAULT 'medium', -- 'low', 'medium', 'high', 'critical'
    action TEXT DEFAULT 'warn', -- 'allow', 'warn', 'block'
    is_active BOOLEAN DEFAULT TRUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    created_by TEXT,
    priority INTEGER DEFAULT 0 -- higher number = higher priority
);

-- Security Policies table for per-message security controls
CREATE TABLE IF NOT EXISTS security_policies (
    policy_id TEXT PRIMARY KEY,
    link_id TEXT NOT NULL,
    reply_id TEXT,
    email_id TEXT,
    dlp_enabled BOOLEAN DEFAULT TRUE,
    watermark_enabled BOOLEAN DEFAULT TRUE,
    download_disabled BOOLEAN DEFAULT FALSE,
    forwarding_disabled BOOLEAN DEFAULT FALSE,
    auto_revoke_after_reply BOOLEAN DEFAULT FALSE,
    max_views INTEGER DEFAULT NULL,
    expires_at DATETIME DEFAULT NULL,
    expires_after_views INTEGER DEFAULT NULL,
    notify_on_expiry BOOLEAN DEFAULT TRUE,
    notify_on_revoke BOOLEAN DEFAULT TRUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    created_by TEXT,
    FOREIGN KEY (link_id) REFERENCES secure_links(link_id) ON DELETE CASCADE,
    FOREIGN KEY (reply_id) REFERENCES secure_replies(reply_id) ON DELETE CASCADE,
    FOREIGN KEY (email_id) REFERENCES emails(email_id) ON DELETE CASCADE
);

-- DLP Scan Results table
CREATE TABLE IF NOT EXISTS dlp_scan_results (
    scan_id TEXT PRIMARY KEY,
    link_id TEXT,
    reply_id TEXT,
    attachment_id TEXT,
    rule_id TEXT NOT NULL,
    content_type TEXT NOT NULL, -- 'email_body', 'reply_body', 'attachment'
    matched_content TEXT, -- the content that triggered the rule
    confidence_score REAL DEFAULT 0.0,
    action_taken TEXT NOT NULL, -- 'allowed', 'warned', 'blocked'
    scan_timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    created_by TEXT,
    FOREIGN KEY (link_id) REFERENCES secure_links(link_id) ON DELETE CASCADE,
    FOREIGN KEY (reply_id) REFERENCES secure_replies(reply_id) ON DELETE CASCADE,
    FOREIGN KEY (attachment_id) REFERENCES secure_attachments(attachment_id) ON DELETE CASCADE,
    FOREIGN KEY (rule_id) REFERENCES dlp_rules(rule_id) ON DELETE CASCADE
);

-- Watermarking Configuration table
CREATE TABLE IF NOT EXISTS watermark_configs (
    config_id TEXT PRIMARY KEY,
    attachment_id TEXT NOT NULL,
    watermark_text TEXT NOT NULL, -- e.g., "Confidential - {recipient_email} - {timestamp}"
    watermark_position TEXT DEFAULT 'bottom-right', -- 'top-left', 'top-right', 'bottom-left', 'bottom-right', 'center'
    watermark_opacity REAL DEFAULT 0.7,
    watermark_font_size INTEGER DEFAULT 12,
    watermark_color TEXT DEFAULT '#FF0000',
    watermark_rotation INTEGER DEFAULT -45,
    applied_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    watermark_hash TEXT, -- hash of the watermarked file
    created_by TEXT,
    FOREIGN KEY (attachment_id) REFERENCES secure_attachments(attachment_id) ON DELETE CASCADE
);

-- Compliance Audit Log table for immutable security records
CREATE TABLE IF NOT EXISTS compliance_audit_log (
    audit_id TEXT PRIMARY KEY,
    event_type TEXT NOT NULL, -- 'dlp_scan', 'watermark_applied', 'policy_enforced', 'expiration_triggered', 'revocation_triggered'
    link_id TEXT,
    reply_id TEXT,
    attachment_id TEXT,
    policy_id TEXT,
    rule_id TEXT,
    user_id TEXT,
    ip_address TEXT,
    user_agent TEXT,
    event_details TEXT, -- JSON details of the event
    severity TEXT DEFAULT 'info', -- 'info', 'warning', 'error', 'critical'
    compliance_category TEXT, -- 'dlp', 'watermarking', 'expiration', 'revocation', 'access_control'
    retention_required BOOLEAN DEFAULT TRUE, -- whether this record must be retained for compliance
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    created_by TEXT,
    FOREIGN KEY (link_id) REFERENCES secure_links(link_id) ON DELETE CASCADE,
    FOREIGN KEY (reply_id) REFERENCES secure_replies(reply_id) ON DELETE CASCADE,
    FOREIGN KEY (attachment_id) REFERENCES secure_attachments(attachment_id) ON DELETE CASCADE,
    FOREIGN KEY (policy_id) REFERENCES security_policies(policy_id) ON DELETE CASCADE,
    FOREIGN KEY (rule_id) REFERENCES dlp_rules(rule_id) ON DELETE CASCADE
);

-- Security Policy Templates for quick application
CREATE TABLE IF NOT EXISTS security_policy_templates (
    template_id TEXT PRIMARY KEY,
    template_name TEXT NOT NULL,
    template_description TEXT,
    dlp_enabled BOOLEAN DEFAULT TRUE,
    watermark_enabled BOOLEAN DEFAULT TRUE,
    download_disabled BOOLEAN DEFAULT FALSE,
    forwarding_disabled BOOLEAN DEFAULT FALSE,
    auto_revoke_after_reply BOOLEAN DEFAULT FALSE,
    max_views INTEGER DEFAULT NULL,
    default_expiry_hours INTEGER DEFAULT 24,
    notify_on_expiry BOOLEAN DEFAULT TRUE,
    notify_on_revoke BOOLEAN DEFAULT TRUE,
    is_default BOOLEAN DEFAULT FALSE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    created_by TEXT
);

-- Insert default DLP rules
INSERT OR IGNORE INTO dlp_rules (rule_id, rule_name, rule_type, pattern, description, severity, action) VALUES
('dlp_001', 'Credit Card Numbers', 'regex', r'\b\d{4}[-\s]?\d{4}[-\s]?\d{4}[-\s]?\d{4}\b', 'Detects credit card numbers in various formats', 'high', 'warn'),
('dlp_002', 'Social Security Numbers', 'regex', r'\b\d{3}[-\s]?\d{2}[-\s]?\d{4}\b', 'Detects US Social Security Numbers', 'high', 'warn'),
('dlp_003', 'Email Addresses', 'regex', r'\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b', 'Detects email addresses', 'medium', 'allow'),
('dlp_004', 'Phone Numbers', 'regex', r'\b\d{3}[-\s]?\d{3}[-\s]?\d{4}\b', 'Detects US phone numbers', 'medium', 'allow'),
('dlp_005', 'Confidential Keywords', 'keyword', 'confidential|secret|private|internal|classified', 'Detects confidential keywords', 'medium', 'warn'),
('dlp_006', 'Financial Keywords', 'keyword', 'password|account|login|credit|debit|bank|routing', 'Detects financial-related keywords', 'high', 'warn');

-- Insert default security policy templates
INSERT OR IGNORE INTO security_policy_templates (template_id, template_name, template_description, dlp_enabled, watermark_enabled, download_disabled, forwarding_disabled, auto_revoke_after_reply, max_views, default_expiry_hours, is_default) VALUES
('template_001', 'Standard', 'Standard security policy with basic protections', TRUE, TRUE, FALSE, FALSE, FALSE, NULL, 24, TRUE),
('template_002', 'High Security', 'Enhanced security with download restrictions and auto-revoke', TRUE, TRUE, TRUE, TRUE, TRUE, 3, 12, FALSE),
('template_003', 'View Only', 'View-only policy with no downloads or forwarding', TRUE, TRUE, TRUE, TRUE, FALSE, 1, 6, FALSE),
('template_004', 'Confidential', 'Maximum security for confidential information', TRUE, TRUE, TRUE, TRUE, TRUE, 1, 4, FALSE);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_security_policies_link_id ON security_policies(link_id);
CREATE INDEX IF NOT EXISTS idx_dlp_scan_results_link_id ON dlp_scan_results(link_id);
CREATE INDEX IF NOT EXISTS idx_compliance_audit_log_link_id ON compliance_audit_log(link_id);
CREATE INDEX IF NOT EXISTS idx_compliance_audit_log_event_type ON compliance_audit_log(event_type);
CREATE INDEX IF NOT EXISTS idx_compliance_audit_log_created_at ON compliance_audit_log(created_at);
CREATE INDEX IF NOT EXISTS idx_watermark_configs_attachment_id ON watermark_configs(attachment_id);

-- Update schema version
PRAGMA user_version = 6;
