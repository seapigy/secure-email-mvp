-- =============================================================================
-- SECURE EMAIL MVP - ENHANCED AUDIT LOG MIGRATION
-- =============================================================================
-- This migration adds enhanced audit logging support for Iteration 4:
-- Hardening & UX Polish with structured JSON logging and suspicious activity detection
-- =============================================================================

-- Create enhanced_audit_log table for comprehensive structured audit logging
CREATE TABLE IF NOT EXISTS enhanced_audit_log (
    event_id TEXT PRIMARY KEY,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    event_type TEXT NOT NULL,
    severity TEXT NOT NULL CHECK (severity IN ('info', 'warning', 'error', 'critical')),
    category TEXT NOT NULL,
    user_id TEXT,
    ip_address TEXT NOT NULL,
    user_agent TEXT NOT NULL,
    link_id TEXT,
    email_id TEXT,
    session_id TEXT,
    request_id TEXT,
    geolocation_data JSON,
    device_info JSON,
    security_flags JSON,
    outcome TEXT NOT NULL,
    details JSON,
    correlation_id TEXT,
    parent_event_id TEXT,
    tags JSON,
    risk_score REAL DEFAULT 0.0,
    is_suspicious BOOLEAN DEFAULT FALSE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (link_id) REFERENCES secure_links(link_id) ON DELETE SET NULL,
    FOREIGN KEY (email_id) REFERENCES emails(email_id) ON DELETE SET NULL,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE SET NULL
);

-- Create indexes for performance and querying
CREATE INDEX IF NOT EXISTS idx_enhanced_audit_log_timestamp ON enhanced_audit_log(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_enhanced_audit_log_event_type ON enhanced_audit_log(event_type);
CREATE INDEX IF NOT EXISTS idx_enhanced_audit_log_severity ON enhanced_audit_log(severity);
CREATE INDEX IF NOT EXISTS idx_enhanced_audit_log_link_id ON enhanced_audit_log(link_id);
CREATE INDEX IF NOT EXISTS idx_enhanced_audit_log_ip_address ON enhanced_audit_log(ip_address);
CREATE INDEX IF NOT EXISTS idx_enhanced_audit_log_is_suspicious ON enhanced_audit_log(is_suspicious);
CREATE INDEX IF NOT EXISTS idx_enhanced_audit_log_risk_score ON enhanced_audit_log(risk_score DESC);
CREATE INDEX IF NOT EXISTS idx_enhanced_audit_log_category ON enhanced_audit_log(category);

-- Create composite indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_enhanced_audit_log_link_severity ON enhanced_audit_log(link_id, severity);
CREATE INDEX IF NOT EXISTS idx_enhanced_audit_log_ip_timestamp ON enhanced_audit_log(ip_address, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_enhanced_audit_log_suspicious_timestamp ON enhanced_audit_log(is_suspicious, timestamp DESC);

-- Create audit_log_retention_policies table for configurable retention
CREATE TABLE IF NOT EXISTS audit_log_retention_policies (
    policy_id TEXT PRIMARY KEY,
    event_type TEXT,
    category TEXT,
    severity TEXT,
    retention_days INTEGER DEFAULT 90,
    auto_purge BOOLEAN DEFAULT TRUE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Insert default retention policies
INSERT OR IGNORE INTO audit_log_retention_policies (policy_id, event_type, category, severity, retention_days, auto_purge) VALUES
('default_info', NULL, NULL, 'info', 30, TRUE),
('default_warning', NULL, NULL, 'warning', 60, TRUE),
('default_error', NULL, NULL, 'error', 90, TRUE),
('default_critical', NULL, NULL, 'critical', 365, TRUE),
('suspicious_activity', 'suspicious_activity', 'security_monitoring', NULL, 180, TRUE),
('secure_link_access', 'secure_link_access', 'access_control', NULL, 90, TRUE),
('security_validation', 'security_validation', 'authentication', NULL, 90, TRUE),
('secure_link_reply', 'secure_link_reply', 'communication', NULL, 90, TRUE);

-- Create audit_log_alerts table for alerting on suspicious activity
CREATE TABLE IF NOT EXISTS audit_log_alerts (
    alert_id TEXT PRIMARY KEY,
    event_id TEXT NOT NULL,
    alert_type TEXT NOT NULL CHECK (alert_type IN ('suspicious_activity', 'high_risk_score', 'multiple_failures', 'unusual_pattern')),
    severity TEXT NOT NULL CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    alert_message TEXT NOT NULL,
    alert_details JSON,
    is_acknowledged BOOLEAN DEFAULT FALSE,
    acknowledged_by TEXT,
    acknowledged_at DATETIME,
    is_resolved BOOLEAN DEFAULT FALSE,
    resolved_at DATETIME,
    resolution_notes TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (event_id) REFERENCES enhanced_audit_log(event_id) ON DELETE CASCADE,
    FOREIGN KEY (acknowledged_by) REFERENCES users(user_id) ON DELETE SET NULL
);

-- Create indexes for alerts
CREATE INDEX IF NOT EXISTS idx_audit_log_alerts_event_id ON audit_log_alerts(event_id);
CREATE INDEX IF NOT EXISTS idx_audit_log_alerts_severity ON audit_log_alerts(severity);
CREATE INDEX IF NOT EXISTS idx_audit_log_alerts_is_acknowledged ON audit_log_alerts(is_acknowledged);
CREATE INDEX IF NOT EXISTS idx_audit_log_alerts_created_at ON audit_log_alerts(created_at DESC);

-- Create audit_log_metrics table for performance monitoring
CREATE TABLE IF NOT EXISTS audit_log_metrics (
    metric_id TEXT PRIMARY KEY,
    metric_name TEXT NOT NULL,
    metric_value REAL NOT NULL,
    metric_unit TEXT,
    metric_timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    metric_metadata JSON,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for metrics
CREATE INDEX IF NOT EXISTS idx_audit_log_metrics_name ON audit_log_metrics(metric_name);
CREATE INDEX IF NOT EXISTS idx_audit_log_metrics_timestamp ON audit_log_metrics(metric_timestamp DESC);

-- Create triggers for automatic cleanup and alerting

-- Trigger to automatically create alerts for suspicious activity
CREATE TRIGGER IF NOT EXISTS trigger_suspicious_activity_alert
AFTER INSERT ON enhanced_audit_log
WHEN NEW.is_suspicious = TRUE
BEGIN
    INSERT INTO audit_log_alerts (
        alert_id,
        event_id,
        alert_type,
        severity,
        alert_message,
        alert_details
    ) VALUES (
        'alert_' || NEW.event_id,
        NEW.event_id,
        'suspicious_activity',
        CASE 
            WHEN NEW.risk_score >= 80 THEN 'critical'
            WHEN NEW.risk_score >= 60 THEN 'high'
            WHEN NEW.risk_score >= 40 THEN 'medium'
            ELSE 'low'
        END,
        'Suspicious activity detected: ' || NEW.event_type,
        json_object(
            'risk_score', NEW.risk_score,
            'ip_address', NEW.ip_address,
            'user_agent', NEW.user_agent,
            'security_flags', NEW.security_flags
        )
    );
END;

-- Trigger to create alerts for high risk scores
CREATE TRIGGER IF NOT EXISTS trigger_high_risk_alert
AFTER INSERT ON enhanced_audit_log
WHEN NEW.risk_score >= 75
BEGIN
    INSERT INTO audit_log_alerts (
        alert_id,
        event_id,
        alert_type,
        severity,
        alert_message,
        alert_details
    ) VALUES (
        'risk_alert_' || NEW.event_id,
        NEW.event_id,
        'high_risk_score',
        CASE 
            WHEN NEW.risk_score >= 90 THEN 'critical'
            WHEN NEW.risk_score >= 80 THEN 'high'
            ELSE 'medium'
        END,
        'High risk activity detected: Risk score ' || NEW.risk_score,
        json_object(
            'risk_score', NEW.risk_score,
            'event_type', NEW.event_type,
            'ip_address', NEW.ip_address
        )
    );
END;

-- Trigger to automatically purge old audit logs based on retention policies
CREATE TRIGGER IF NOT EXISTS trigger_audit_log_cleanup
AFTER INSERT ON enhanced_audit_log
BEGIN
    -- Delete old info level logs (30 days)
    DELETE FROM enhanced_audit_log 
    WHERE severity = 'info' 
    AND timestamp < datetime('now', '-30 days');
    
    -- Delete old warning level logs (60 days)
    DELETE FROM enhanced_audit_log 
    WHERE severity = 'warning' 
    AND timestamp < datetime('now', '-60 days');
    
    -- Delete old error level logs (90 days)
    DELETE FROM enhanced_audit_log 
    WHERE severity = 'error' 
    AND timestamp < datetime('now', '-90 days');
    
    -- Keep critical logs for 1 year (no automatic deletion)
END;

-- Create view for easy querying of recent suspicious activity
CREATE VIEW IF NOT EXISTS v_recent_suspicious_activity AS
SELECT 
    event_id,
    timestamp,
    event_type,
    severity,
    ip_address,
    user_agent,
    risk_score,
    outcome,
    details
FROM enhanced_audit_log 
WHERE is_suspicious = TRUE 
AND timestamp > datetime('now', '-24 hours')
ORDER BY timestamp DESC;

-- Create view for audit log summary statistics
CREATE VIEW IF NOT EXISTS v_audit_log_summary AS
SELECT 
    DATE(timestamp) as date,
    event_type,
    severity,
    COUNT(*) as event_count,
    AVG(risk_score) as avg_risk_score,
    SUM(CASE WHEN is_suspicious THEN 1 ELSE 0 END) as suspicious_count
FROM enhanced_audit_log 
WHERE timestamp > datetime('now', '-30 days')
GROUP BY DATE(timestamp), event_type, severity
ORDER BY date DESC, event_count DESC;

-- Insert initial metrics
INSERT OR IGNORE INTO audit_log_metrics (metric_id, metric_name, metric_value, metric_unit, metric_metadata) VALUES
('initial_setup', 'audit_log_initialized', 1, 'boolean', json_object('version', '1.0', 'iteration', '4'));

-- Update schema version
PRAGMA user_version = 4;
