-- Migration: Add AI DLP Support
-- Version: 7
-- Description: Adds AI-powered DLP scanning with content classification, severity scoring, and override capabilities

-- AI DLP Scan Results table
CREATE TABLE IF NOT EXISTS ai_dlp_scan_results (
    scan_id TEXT PRIMARY KEY,
    link_id TEXT,
    reply_id TEXT,
    attachment_id TEXT,
    content_type TEXT NOT NULL,
    content_hash TEXT NOT NULL,
    classification TEXT, -- JSON field for AIContentClassification
    severity_score REAL NOT NULL,
    risk_level TEXT NOT NULL,
    action_recommended TEXT NOT NULL,
    action_taken TEXT NOT NULL,
    override_reason TEXT,
    override_by TEXT,
    model_version TEXT NOT NULL,
    processing_time REAL NOT NULL,
    scan_timestamp DATETIME NOT NULL,
    created_by TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (link_id) REFERENCES secure_links(link_id) ON DELETE CASCADE,
    FOREIGN KEY (reply_id) REFERENCES secure_replies(reply_id) ON DELETE CASCADE,
    FOREIGN KEY (attachment_id) REFERENCES secure_attachments(attachment_id) ON DELETE CASCADE
);

-- AI DLP Policies table
CREATE TABLE IF NOT EXISTS ai_dlp_policies (
    policy_id TEXT PRIMARY KEY,
    policy_name TEXT NOT NULL,
    description TEXT,
    is_active BOOLEAN DEFAULT 1,
    categories TEXT NOT NULL, -- JSON array of categories to scan
    severity_thresholds TEXT NOT NULL, -- JSON map of severity thresholds
    actions TEXT NOT NULL, -- JSON map of severity -> action
    confidence_threshold REAL DEFAULT 0.5,
    risk_threshold REAL DEFAULT 0.7,
    allow_override BOOLEAN DEFAULT 0,
    override_roles TEXT, -- JSON array of roles that can override
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    created_by TEXT
);

-- AI DLP Overrides table
CREATE TABLE IF NOT EXISTS ai_dlp_overrides (
    override_id TEXT PRIMARY KEY,
    scan_id TEXT NOT NULL,
    original_action TEXT NOT NULL,
    override_reason TEXT NOT NULL,
    justification TEXT NOT NULL,
    user_id TEXT NOT NULL,
    user_role TEXT NOT NULL,
    override_timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (scan_id) REFERENCES ai_dlp_scan_results(scan_id) ON DELETE CASCADE
);

-- AI DLP Metrics table
CREATE TABLE IF NOT EXISTS ai_dlp_metrics (
    metric_id TEXT PRIMARY KEY,
    total_scans INTEGER DEFAULT 0,
    average_time REAL DEFAULT 0.0,
    accuracy REAL DEFAULT 0.0,
    false_positives INTEGER DEFAULT 0,
    false_negatives INTEGER DEFAULT 0,
    overrides INTEGER DEFAULT 0,
    blocked_content INTEGER DEFAULT 0,
    warned_content INTEGER DEFAULT 0,
    allowed_content INTEGER DEFAULT 0,
    model_version TEXT NOT NULL,
    last_updated DATETIME DEFAULT CURRENT_TIMESTAMP,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Update compliance_audit_log to support AI DLP events
ALTER TABLE compliance_audit_log ADD COLUMN ai_dlp_scan_id TEXT;
ALTER TABLE compliance_audit_log ADD COLUMN ai_dlp_override_id TEXT;
ALTER TABLE compliance_audit_log ADD COLUMN severity_score REAL;
ALTER TABLE compliance_audit_log ADD COLUMN risk_level TEXT;
ALTER TABLE compliance_audit_log ADD COLUMN model_version TEXT;

-- Add indexes for performance
CREATE INDEX IF NOT EXISTS idx_ai_dlp_scan_results_link_id ON ai_dlp_scan_results(link_id);
CREATE INDEX IF NOT EXISTS idx_ai_dlp_scan_results_content_hash ON ai_dlp_scan_results(content_hash);
CREATE INDEX IF NOT EXISTS idx_ai_dlp_scan_results_risk_level ON ai_dlp_scan_results(risk_level);
CREATE INDEX IF NOT EXISTS idx_ai_dlp_scan_results_action_taken ON ai_dlp_scan_results(action_taken);
CREATE INDEX IF NOT EXISTS idx_ai_dlp_scan_results_scan_timestamp ON ai_dlp_scan_results(scan_timestamp);

CREATE INDEX IF NOT EXISTS idx_ai_dlp_policies_is_active ON ai_dlp_policies(is_active);
CREATE INDEX IF NOT EXISTS idx_ai_dlp_policies_created_at ON ai_dlp_policies(created_at);

CREATE INDEX IF NOT EXISTS idx_ai_dlp_overrides_scan_id ON ai_dlp_overrides(scan_id);
CREATE INDEX IF NOT EXISTS idx_ai_dlp_overrides_user_id ON ai_dlp_overrides(user_id);
CREATE INDEX IF NOT EXISTS idx_ai_dlp_overrides_override_timestamp ON ai_dlp_overrides(override_timestamp);

CREATE INDEX IF NOT EXISTS idx_compliance_audit_ai_dlp_scan_id ON compliance_audit_log(ai_dlp_scan_id);
CREATE INDEX IF NOT EXISTS idx_compliance_audit_ai_dlp_override_id ON compliance_audit_log(ai_dlp_override_id);
CREATE INDEX IF NOT EXISTS idx_compliance_audit_severity_score ON compliance_audit_log(severity_score);
CREATE INDEX IF NOT EXISTS idx_compliance_audit_risk_level ON compliance_audit_log(risk_level);

-- Insert default AI DLP policy
INSERT OR IGNORE INTO ai_dlp_policies (
    policy_id,
    policy_name,
    description,
    is_active,
    categories,
    severity_thresholds,
    actions,
    confidence_threshold,
    risk_threshold,
    allow_override,
    override_roles,
    created_by
) VALUES (
    'default_ai_dlp_policy',
    'Default AI DLP Policy',
    'Default policy for AI-powered data loss prevention scanning',
    1,
    '["pii", "financial", "healthcare", "legal", "confidential"]',
    '{"critical": 0.9, "high": 0.7, "medium": 0.5, "low": 0.3}',
    '{"critical": "block", "high": "warn", "medium": "warn", "low": "allow"}',
    0.5,
    0.7,
    1,
    '["admin", "security_officer", "compliance_manager"]',
    'system'
);

-- Insert initial metrics record
INSERT OR IGNORE INTO ai_dlp_metrics (
    metric_id,
    model_version,
    last_updated
) VALUES (
    'default_metrics',
    'nlp-v1.0.0',
    CURRENT_TIMESTAMP
);

-- Update database version
PRAGMA user_version = 7;
