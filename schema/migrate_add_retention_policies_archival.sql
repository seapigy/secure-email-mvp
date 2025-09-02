-- Migration: Add retention policies and archival tables (Micro-Iteration 4.26)
-- This migration adds support for smart retention policies and automated email archival

-- Retention policies table
CREATE TABLE IF NOT EXISTS retention_policies (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT,
    priority INTEGER DEFAULT 0, -- Higher number = higher priority
    active BOOLEAN DEFAULT 1,
    
    -- Rule conditions
    user_id TEXT, -- Specific user
    sender_domain TEXT, -- Sender email domain
    recipient_domain TEXT, -- Recipient email domain
    email_status TEXT, -- read, unread, expired
    custom_tags TEXT, -- JSON array of tags
    min_age_hours INTEGER, -- Minimum email age
    max_age_hours INTEGER, -- Maximum email age
    
    -- Actions
    retention_days INTEGER NOT NULL DEFAULT 30, -- How long to keep emails
    archive_instead BOOLEAN DEFAULT 0, -- Archive instead of delete
    archive_retention_days INTEGER DEFAULT 365, -- How long to keep archived
    
    -- Metadata
    created_by TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Archived emails table
CREATE TABLE IF NOT EXISTS archived_emails (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    original_email_id TEXT NOT NULL,
    sender_id TEXT NOT NULL,
    recipient TEXT NOT NULL,
    subject TEXT,
    archived_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    archive_reason TEXT NOT NULL, -- "expired", "policy", "manual"
    retention_days INTEGER NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    
    -- Storage information
    archive_blob_url TEXT NOT NULL,
    encryption_key TEXT NOT NULL, -- Encrypted key
    compressed_size INTEGER DEFAULT 0,
    original_size INTEGER DEFAULT 0,
    
    -- Metadata
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Policy evaluation logs table
CREATE TABLE IF NOT EXISTS policy_evaluation_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email_id TEXT NOT NULL,
    policy_id INTEGER NOT NULL,
    evaluation_result TEXT NOT NULL, -- "matched", "not_matched", "applied"
    match_score INTEGER DEFAULT 0,
    match_reasons TEXT, -- JSON array of reasons
    evaluated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (policy_id) REFERENCES retention_policies(id)
);

-- Archive operation logs table
CREATE TABLE IF NOT EXISTS archive_operation_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    operation_type TEXT NOT NULL, -- "archive", "restore", "cleanup"
    email_id TEXT,
    archive_id INTEGER,
    operation_result TEXT NOT NULL, -- "success", "failed", "skipped"
    error_message TEXT,
    operation_duration_ms INTEGER,
    processed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (archive_id) REFERENCES archived_emails(id)
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_retention_policies_active ON retention_policies(active);
CREATE INDEX IF NOT EXISTS idx_retention_policies_priority ON retention_policies(priority);
CREATE INDEX IF NOT EXISTS idx_retention_policies_user_id ON retention_policies(user_id);
CREATE INDEX IF NOT EXISTS idx_retention_policies_sender_domain ON retention_policies(sender_domain);
CREATE INDEX IF NOT EXISTS idx_retention_policies_recipient_domain ON retention_policies(recipient_domain);

CREATE INDEX IF NOT EXISTS idx_archived_emails_original_id ON archived_emails(original_email_id);
CREATE INDEX IF NOT EXISTS idx_archived_emails_sender_id ON archived_emails(sender_id);
CREATE INDEX IF NOT EXISTS idx_archived_emails_archived_at ON archived_emails(archived_at);
CREATE INDEX IF NOT EXISTS idx_archived_emails_expires_at ON archived_emails(expires_at);
CREATE INDEX IF NOT EXISTS idx_archived_emails_reason ON archived_emails(archive_reason);

CREATE INDEX IF NOT EXISTS idx_policy_evaluation_email_id ON policy_evaluation_logs(email_id);
CREATE INDEX IF NOT EXISTS idx_policy_evaluation_policy_id ON policy_evaluation_logs(policy_id);
CREATE INDEX IF NOT EXISTS idx_policy_evaluation_at ON policy_evaluation_logs(evaluated_at);

CREATE INDEX IF NOT EXISTS idx_archive_operations_type ON archive_operation_logs(operation_type);
CREATE INDEX IF NOT EXISTS idx_archive_operations_email_id ON archive_operation_logs(email_id);
CREATE INDEX IF NOT EXISTS idx_archive_operations_processed_at ON archive_operation_logs(processed_at);

-- Create views for analytics
CREATE VIEW IF NOT EXISTS retention_policy_summary AS
SELECT 
    COUNT(*) as total_policies,
    SUM(CASE WHEN active = 1 THEN 1 ELSE 0 END) as active_policies,
    SUM(CASE WHEN archive_instead = 1 THEN 1 ELSE 0 END) as archive_policies,
    AVG(retention_days) as avg_retention_days,
    AVG(archive_retention_days) as avg_archive_retention_days
FROM retention_policies;

CREATE VIEW IF NOT EXISTS archived_emails_summary AS
SELECT 
    COUNT(*) as total_archived,
    SUM(CASE WHEN expires_at <= datetime('now') THEN 1 ELSE 0 END) as expired_archives,
    SUM(compressed_size) as total_compressed_size,
    SUM(original_size) as total_original_size,
    AVG(compressed_size) as avg_compressed_size,
    archive_reason,
    COUNT(*) as count_by_reason
FROM archived_emails
GROUP BY archive_reason;

CREATE VIEW IF NOT EXISTS policy_evaluation_summary AS
SELECT 
    policy_id,
    COUNT(*) as total_evaluations,
    SUM(CASE WHEN evaluation_result = 'matched' THEN 1 ELSE 0 END) as matches,
    SUM(CASE WHEN evaluation_result = 'applied' THEN 1 ELSE 0 END) as applied,
    AVG(match_score) as avg_match_score
FROM policy_evaluation_logs
GROUP BY policy_id;

-- Create triggers for automatic timestamp updates
CREATE TRIGGER IF NOT EXISTS update_retention_policies_updated_at
    AFTER UPDATE ON retention_policies
    FOR EACH ROW
BEGIN
    UPDATE retention_policies SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS update_archived_emails_updated_at
    AFTER UPDATE ON archived_emails
    FOR EACH ROW
BEGIN
    UPDATE archived_emails SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- Insert default retention policy
INSERT OR IGNORE INTO retention_policies (
    name, description, priority, active, retention_days, 
    archive_instead, archive_retention_days, created_by, created_at, updated_at
) VALUES (
    'Default Policy',
    'Default retention policy when no specific policy matches',
    0,
    1,
    30,
    0,
    365,
    'system',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
);

-- Create cleanup trigger for expired policy evaluation logs (keep last 30 days)
CREATE TRIGGER IF NOT EXISTS cleanup_old_policy_evaluations
    AFTER INSERT ON policy_evaluation_logs
    FOR EACH ROW
BEGIN
    DELETE FROM policy_evaluation_logs 
    WHERE evaluated_at < datetime('now', '-30 days');
END;

-- Create cleanup trigger for expired archive operation logs (keep last 90 days)
CREATE TRIGGER IF NOT EXISTS cleanup_old_archive_operations
    AFTER INSERT ON archive_operation_logs
    FOR EACH ROW
BEGIN
    DELETE FROM archive_operation_logs 
    WHERE processed_at < datetime('now', '-90 days');
END;



















