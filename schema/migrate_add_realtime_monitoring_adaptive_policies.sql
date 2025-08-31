-- Migration: Add real-time retention monitoring and adaptive policy enforcement (Micro-Iteration 4.28)
-- This migration adds support for real-time monitoring and adaptive policy adjustments

-- Real-time retention metrics cache table
CREATE TABLE IF NOT EXISTS realtime_retention_metrics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    metric_type TEXT NOT NULL, -- "user", "domain", "policy", "global"
    metric_key TEXT NOT NULL, -- user_id, domain, policy_id, or "global"
    
    -- Live metrics
    active_emails_count INTEGER DEFAULT 0,
    archived_emails_count INTEGER DEFAULT 0,
    deleted_emails_count INTEGER DEFAULT 0,
    total_storage_bytes INTEGER DEFAULT 0,
    compressed_storage_bytes INTEGER DEFAULT 0,
    
    -- Policy performance metrics
    policy_evaluations_count INTEGER DEFAULT 0,
    policy_matches_count INTEGER DEFAULT 0,
    policy_applications_count INTEGER DEFAULT 0,
    avg_match_score REAL DEFAULT 0.0,
    avg_impact_score REAL DEFAULT 0.0,
    
    -- Archival load metrics
    archival_operations_count INTEGER DEFAULT 0,
    avg_archival_duration_ms INTEGER DEFAULT 0,
    archival_success_rate REAL DEFAULT 0.0,
    
    -- Timestamps
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(metric_type, metric_key)
);

-- Adaptive policy changes tracking table
CREATE TABLE IF NOT EXISTS adaptive_policy_changes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    policy_id INTEGER NOT NULL,
    
    -- Change details
    change_type TEXT NOT NULL, -- "retention_days", "archive_instead", "archive_retention_days"
    old_value TEXT NOT NULL, -- Previous value
    new_value TEXT NOT NULL, -- New value
    change_reason TEXT NOT NULL, -- Reason for the change
    change_percentage REAL DEFAULT 0.0, -- Percentage change
    
    -- Impact analysis
    expected_storage_savings_bytes INTEGER DEFAULT 0,
    expected_archival_load_impact REAL DEFAULT 0.0,
    risk_assessment TEXT DEFAULT "low", -- "low", "medium", "high"
    
    -- Safety controls
    cooldown_until TIMESTAMP, -- When this policy can be changed again
    requires_admin_approval BOOLEAN DEFAULT 0, -- Whether admin approval is required
    
    -- Status tracking
    status TEXT DEFAULT "pending", -- "pending", "approved", "applied", "rejected"
    applied_at TIMESTAMP,
    applied_by TEXT, -- "system" or admin user
    applied_result TEXT, -- "success", "partial", "failed"
    
    -- Metadata
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (policy_id) REFERENCES retention_policies(id)
);

-- Real-time event stream table for WebSocket/SSE
CREATE TABLE IF NOT EXISTS retention_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    event_type TEXT NOT NULL, -- "policy_evaluation", "email_deletion", "email_archival", "policy_change"
    event_data TEXT NOT NULL, -- JSON event data
    user_id TEXT, -- Target user for the event
    policy_id INTEGER, -- Related policy if applicable
    
    -- Event metadata
    event_timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    processed BOOLEAN DEFAULT 0, -- Whether event has been sent to clients
    
    FOREIGN KEY (policy_id) REFERENCES retention_policies(id)
);

-- Adaptive policy configuration table
CREATE TABLE IF NOT EXISTS adaptive_policy_config (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    policy_id INTEGER NOT NULL,
    
    -- Configuration settings
    adaptive_enabled BOOLEAN DEFAULT 0, -- Whether adaptive adjustments are enabled
    max_change_percentage REAL DEFAULT 20.0, -- Maximum percentage change allowed
    cooldown_days INTEGER DEFAULT 7, -- Days to wait between changes
    requires_admin_approval BOOLEAN DEFAULT 1, -- Whether admin approval is required
    
    -- Thresholds
    min_retention_days INTEGER DEFAULT 1, -- Minimum retention days
    max_retention_days INTEGER DEFAULT 365, -- Maximum retention days
    min_archive_retention_days INTEGER DEFAULT 30, -- Minimum archive retention days
    max_archive_retention_days INTEGER DEFAULT 2555, -- Maximum archive retention days (7 years)
    
    -- Safety limits
    max_storage_impact_bytes INTEGER DEFAULT 1073741824, -- 1GB max storage impact
    max_archival_load_impact REAL DEFAULT 0.5, -- Max 50% archival load impact
    
    -- Metadata
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (policy_id) REFERENCES retention_policies(id),
    UNIQUE(policy_id)
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_realtime_metrics_type_key ON realtime_retention_metrics(metric_type, metric_key);
CREATE INDEX IF NOT EXISTS idx_realtime_metrics_last_updated ON realtime_retention_metrics(last_updated);

CREATE INDEX IF NOT EXISTS idx_adaptive_changes_policy_id ON adaptive_policy_changes(policy_id);
CREATE INDEX IF NOT EXISTS idx_adaptive_changes_status ON adaptive_policy_changes(status);
CREATE INDEX IF NOT EXISTS idx_adaptive_changes_created_at ON adaptive_policy_changes(created_at);
CREATE INDEX IF NOT EXISTS idx_adaptive_changes_cooldown ON adaptive_policy_changes(cooldown_until);

CREATE INDEX IF NOT EXISTS idx_retention_events_type ON retention_events(event_type);
CREATE INDEX IF NOT EXISTS idx_retention_events_user_id ON retention_events(user_id);
CREATE INDEX IF NOT EXISTS idx_retention_events_timestamp ON retention_events(event_timestamp);
CREATE INDEX IF NOT EXISTS idx_retention_events_processed ON retention_events(processed);

CREATE INDEX IF NOT EXISTS idx_adaptive_config_policy_id ON adaptive_policy_config(policy_id);
CREATE INDEX IF NOT EXISTS idx_adaptive_config_enabled ON adaptive_policy_config(adaptive_enabled);

-- Create views for real-time analytics
CREATE VIEW IF NOT EXISTS realtime_retention_summary AS
SELECT 
    metric_type,
    COUNT(*) as total_metrics,
    SUM(active_emails_count) as total_active_emails,
    SUM(archived_emails_count) as total_archived_emails,
    SUM(deleted_emails_count) as total_deleted_emails,
    SUM(total_storage_bytes) as total_storage_bytes,
    SUM(compressed_storage_bytes) as total_compressed_storage_bytes,
    AVG(avg_impact_score) as avg_impact_score,
    MAX(last_updated) as last_updated
FROM realtime_retention_metrics
GROUP BY metric_type;

CREATE VIEW IF NOT EXISTS adaptive_policy_changes_summary AS
SELECT 
    policy_id,
    COUNT(*) as total_changes,
    SUM(CASE WHEN status = 'applied' THEN 1 ELSE 0 END) as applied_changes,
    SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) as pending_changes,
    AVG(change_percentage) as avg_change_percentage,
    MAX(created_at) as last_change_at
FROM adaptive_policy_changes
GROUP BY policy_id;

CREATE VIEW IF NOT EXISTS retention_events_summary AS
SELECT 
    event_type,
    COUNT(*) as total_events,
    SUM(CASE WHEN processed = 1 THEN 1 ELSE 0 END) as processed_events,
    MAX(event_timestamp) as last_event_at
FROM retention_events
GROUP BY event_type;

-- Create triggers for automatic timestamp updates
CREATE TRIGGER IF NOT EXISTS update_realtime_metrics_last_updated
    AFTER UPDATE ON realtime_retention_metrics
    FOR EACH ROW
BEGIN
    UPDATE realtime_retention_metrics SET last_updated = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS update_adaptive_changes_updated_at
    AFTER UPDATE ON adaptive_policy_changes
    FOR EACH ROW
BEGIN
    UPDATE adaptive_policy_changes SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS update_adaptive_config_updated_at
    AFTER UPDATE ON adaptive_policy_config
    FOR EACH ROW
BEGIN
    UPDATE adaptive_policy_config SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- Create cleanup triggers for old data
CREATE TRIGGER IF NOT EXISTS cleanup_old_retention_events
    AFTER INSERT ON retention_events
    FOR EACH ROW
BEGIN
    DELETE FROM retention_events 
    WHERE event_timestamp < datetime('now', '-7 days') AND processed = 1;
END;

CREATE TRIGGER IF NOT EXISTS cleanup_old_adaptive_changes
    AFTER INSERT ON adaptive_policy_changes
    FOR EACH ROW
BEGIN
    DELETE FROM adaptive_policy_changes 
    WHERE created_at < datetime('now', '-90 days') AND status IN ('applied', 'rejected');
END;

-- Insert default adaptive policy configuration for existing policies
INSERT OR IGNORE INTO adaptive_policy_config (
    policy_id, adaptive_enabled, max_change_percentage, cooldown_days, 
    requires_admin_approval, min_retention_days, max_retention_days,
    min_archive_retention_days, max_archive_retention_days,
    max_storage_impact_bytes, max_archival_load_impact
)
SELECT 
    id as policy_id,
    0 as adaptive_enabled, -- Disabled by default for safety
    20.0 as max_change_percentage,
    7 as cooldown_days,
    1 as requires_admin_approval,
    1 as min_retention_days,
    365 as max_retention_days,
    30 as min_archive_retention_days,
    2555 as max_archive_retention_days,
    1073741824 as max_storage_impact_bytes, -- 1GB
    0.5 as max_archival_load_impact
FROM retention_policies
WHERE active = 1;

-- Insert initial real-time metrics for global scope
INSERT OR IGNORE INTO realtime_retention_metrics (
    metric_type, metric_key, active_emails_count, archived_emails_count,
    deleted_emails_count, total_storage_bytes, compressed_storage_bytes,
    policy_evaluations_count, policy_matches_count, policy_applications_count,
    avg_match_score, avg_impact_score, archival_operations_count,
    avg_archival_duration_ms, archival_success_rate, last_updated, created_at
) VALUES (
    'global', 'global', 0, 0, 0, 0, 0, 0, 0, 0, 0.0, 0.0, 0, 0, 0.0,
    CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
);

















