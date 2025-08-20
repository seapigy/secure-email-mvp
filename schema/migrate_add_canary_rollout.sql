-- =============================================================================
-- Migration: Add Canary Rollout + Monitoring + Runbooks (Sprint 6)
-- =============================================================================

BEGIN TRANSACTION;

-- Canary rollout configuration and state
CREATE TABLE IF NOT EXISTS canary_config (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    enabled BOOLEAN DEFAULT FALSE,
    traffic_percentage REAL DEFAULT 0.0,
    user_segments TEXT, -- JSON array
    rollback_threshold REAL DEFAULT 5.0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- A/B test results and metrics
CREATE TABLE IF NOT EXISTS ab_test_results (
    id TEXT PRIMARY KEY,
    test_name TEXT NOT NULL,
    variant TEXT NOT NULL, -- 'legacy' or 'e2e'
    metric_name TEXT NOT NULL,
    metric_value REAL NOT NULL,
    sample_size INTEGER NOT NULL,
    confidence_interval_lower REAL,
    confidence_interval_upper REAL,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Rollback events and triggers
CREATE TABLE IF NOT EXISTS rollback_events (
    id TEXT PRIMARY KEY,
    trigger_type TEXT NOT NULL, -- 'manual', 'automatic', 'threshold'
    trigger_condition TEXT NOT NULL,
    rollback_reason TEXT,
    affected_users INTEGER,
    duration_seconds INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Monitoring alerts and notifications
CREATE TABLE IF NOT EXISTS monitoring_alerts (
    id TEXT PRIMARY KEY,
    alert_name TEXT NOT NULL,
    severity TEXT NOT NULL, -- 'critical', 'warning', 'info'
    message TEXT NOT NULL,
    metric_name TEXT,
    metric_value REAL,
    threshold_value REAL,
    status TEXT DEFAULT 'active', -- 'active', 'acknowledged', 'resolved'
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMP
);

-- Runbook execution tracking
CREATE TABLE IF NOT EXISTS runbook_executions (
    id TEXT PRIMARY KEY,
    runbook_name TEXT NOT NULL,
    procedure_name TEXT NOT NULL,
    status TEXT NOT NULL, -- 'running', 'completed', 'failed', 'rolled_back'
    started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,
    error_message TEXT,
    execution_log TEXT, -- JSON array of step results
    operator_id TEXT,
    correlation_id TEXT
);

-- Performance baselines for comparison
CREATE TABLE IF NOT EXISTS performance_baselines (
    id TEXT PRIMARY KEY,
    baseline_name TEXT NOT NULL,
    metric_name TEXT NOT NULL,
    baseline_value REAL NOT NULL,
    acceptable_range_min REAL,
    acceptable_range_max REAL,
    sample_size INTEGER NOT NULL,
    confidence_level REAL DEFAULT 0.95,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Incident tracking and response
CREATE TABLE IF NOT EXISTS incidents (
    id TEXT PRIMARY KEY,
    incident_id TEXT NOT NULL UNIQUE,
    severity TEXT NOT NULL, -- 'critical', 'high', 'medium', 'low'
    status TEXT NOT NULL, -- 'open', 'investigating', 'resolved', 'closed'
    title TEXT NOT NULL,
    description TEXT,
    affected_services TEXT, -- JSON array
    impact_assessment TEXT,
    root_cause TEXT,
    resolution TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMP,
    closed_at TIMESTAMP,
    assigned_to TEXT,
    correlation_id TEXT
);

-- Metrics collection and aggregation
CREATE TABLE IF NOT EXISTS metrics_collection (
    id TEXT PRIMARY KEY,
    metric_name TEXT NOT NULL,
    metric_value REAL NOT NULL,
    metric_type TEXT NOT NULL, -- 'counter', 'gauge', 'histogram'
    labels TEXT, -- JSON object
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    source TEXT NOT NULL, -- 'application', 'infrastructure', 'business'
    correlation_id TEXT
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_canary_config_enabled ON canary_config(enabled);
CREATE INDEX IF NOT EXISTS idx_ab_test_results_test_variant ON ab_test_results(test_name, variant);
CREATE INDEX IF NOT EXISTS idx_ab_test_results_timestamp ON ab_test_results(timestamp);
CREATE INDEX IF NOT EXISTS idx_rollback_events_trigger_type ON rollback_events(trigger_type);
CREATE INDEX IF NOT EXISTS idx_rollback_events_created_at ON rollback_events(created_at);
CREATE INDEX IF NOT EXISTS idx_monitoring_alerts_severity ON monitoring_alerts(severity);
CREATE INDEX IF NOT EXISTS idx_monitoring_alerts_status ON monitoring_alerts(status);
CREATE INDEX IF NOT EXISTS idx_monitoring_alerts_created_at ON monitoring_alerts(created_at);
CREATE INDEX IF NOT EXISTS idx_runbook_executions_status ON runbook_executions(status);
CREATE INDEX IF NOT EXISTS idx_runbook_executions_started_at ON runbook_executions(started_at);
CREATE INDEX IF NOT EXISTS idx_performance_baselines_name ON performance_baselines(baseline_name);
CREATE INDEX IF NOT EXISTS idx_incidents_severity ON incidents(severity);
CREATE INDEX IF NOT EXISTS idx_incidents_status ON incidents(status);
CREATE INDEX IF NOT EXISTS idx_incidents_created_at ON incidents(created_at);
CREATE INDEX IF NOT EXISTS idx_metrics_collection_name ON metrics_collection(metric_name);
CREATE INDEX IF NOT EXISTS idx_metrics_collection_timestamp ON metrics_collection(timestamp);
CREATE INDEX IF NOT EXISTS idx_metrics_collection_source ON metrics_collection(source);

-- Create views for common queries
CREATE VIEW IF NOT EXISTS v_active_alerts AS
SELECT 
    alert_name,
    severity,
    message,
    metric_name,
    metric_value,
    threshold_value,
    created_at
FROM monitoring_alerts 
WHERE status = 'active'
ORDER BY severity DESC, created_at DESC;

CREATE VIEW IF NOT EXISTS v_recent_rollbacks AS
SELECT 
    trigger_type,
    trigger_condition,
    rollback_reason,
    affected_users,
    duration_seconds,
    created_at
FROM rollback_events 
WHERE created_at >= datetime('now', '-24 hours')
ORDER BY created_at DESC;

CREATE VIEW IF NOT EXISTS v_performance_comparison AS
SELECT 
    ab.test_name,
    ab.variant,
    ab.metric_name,
    ab.metric_value,
    pb.baseline_value,
    (ab.metric_value - pb.baseline_value) / pb.baseline_value * 100 as percent_change
FROM ab_test_results ab
LEFT JOIN performance_baselines pb ON ab.metric_name = pb.metric_name
WHERE ab.timestamp >= datetime('now', '-1 hour')
ORDER BY ab.test_name, ab.metric_name, ab.variant;

CREATE VIEW IF NOT EXISTS v_incident_summary AS
SELECT 
    severity,
    status,
    COUNT(*) as count,
    AVG(JULIANDAY(resolved_at) - JULIANDAY(created_at)) * 24 * 60 as avg_resolution_minutes
FROM incidents 
WHERE created_at >= datetime('now', '-7 days')
GROUP BY severity, status
ORDER BY severity DESC, status;

-- Insert initial canary configuration
INSERT OR IGNORE INTO canary_config (id, name, enabled, traffic_percentage, user_segments, rollback_threshold) VALUES
('canary_main', 'Main E2E Canary', FALSE, 0.0, '["internal_testers", "beta_users"]', 5.0),
('canary_performance', 'Performance Test Canary', FALSE, 0.0, '["performance_testers"]', 3.0),
('canary_security', 'Security Test Canary', FALSE, 0.0, '["security_team"]', 2.0);

-- Insert sample performance baselines
INSERT OR IGNORE INTO performance_baselines (id, baseline_name, metric_name, baseline_value, acceptable_range_min, acceptable_range_max, sample_size) VALUES
('baseline_legacy_latency', 'Legacy System Latency', 'response_time_ms', 150.0, 100.0, 300.0, 10000),
('baseline_legacy_throughput', 'Legacy System Throughput', 'requests_per_second', 1000.0, 800.0, 1200.0, 10000),
('baseline_legacy_error_rate', 'Legacy System Error Rate', 'error_rate_percent', 0.1, 0.0, 1.0, 10000),
('baseline_e2e_latency', 'E2E System Latency', 'response_time_ms', 200.0, 150.0, 400.0, 10000),
('baseline_e2e_throughput', 'E2E System Throughput', 'requests_per_second', 800.0, 600.0, 1000.0, 10000),
('baseline_e2e_error_rate', 'E2E System Error Rate', 'error_rate_percent', 0.2, 0.0, 2.0, 10000);

-- Create triggers for audit logging
CREATE TRIGGER IF NOT EXISTS tr_canary_config_audit
AFTER UPDATE ON canary_config
BEGIN
    INSERT INTO audit_logs (action, table_name, record_id, old_values, new_values, user_id, timestamp)
    VALUES (
        'UPDATE',
        'canary_config',
        NEW.id,
        json_object('enabled', OLD.enabled, 'traffic_percentage', OLD.traffic_percentage),
        json_object('enabled', NEW.enabled, 'traffic_percentage', NEW.traffic_percentage),
        'system',
        CURRENT_TIMESTAMP
    );
END;

CREATE TRIGGER IF NOT EXISTS tr_rollback_events_audit
AFTER INSERT ON rollback_events
BEGIN
    INSERT INTO audit_logs (action, table_name, record_id, new_values, user_id, timestamp)
    VALUES (
        'INSERT',
        'rollback_events',
        NEW.id,
        json_object('trigger_type', NEW.trigger_type, 'trigger_condition', NEW.trigger_condition, 'affected_users', NEW.affected_users),
        'system',
        CURRENT_TIMESTAMP
    );
END;

CREATE TRIGGER IF NOT EXISTS tr_monitoring_alerts_audit
AFTER INSERT ON monitoring_alerts
BEGIN
    INSERT INTO audit_logs (action, table_name, record_id, new_values, user_id, timestamp)
    VALUES (
        'INSERT',
        'monitoring_alerts',
        NEW.id,
        json_object('alert_name', NEW.alert_name, 'severity', NEW.severity, 'metric_name', NEW.metric_name),
        'system',
        CURRENT_TIMESTAMP
    );
END;

-- Create function to update canary config timestamps
CREATE TRIGGER IF NOT EXISTS tr_canary_config_timestamp
AFTER UPDATE ON canary_config
BEGIN
    UPDATE canary_config SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- Create function to update performance baselines timestamps
CREATE TRIGGER IF NOT EXISTS tr_performance_baselines_timestamp
AFTER UPDATE ON performance_baselines
BEGIN
    UPDATE performance_baselines SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

COMMIT;
