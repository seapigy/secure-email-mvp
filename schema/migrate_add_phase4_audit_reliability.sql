-- ============================================================================
-- PHASE 4: AUDIT & RELIABILITY DATABASE SCHEMA
-- ============================================================================

-- Audit Events Table
CREATE TABLE IF NOT EXISTS audit_events (
    id TEXT PRIMARY KEY,
    event_type TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id TEXT NOT NULL,
    user_id TEXT,
    ip_address TEXT NOT NULL,
    user_agent TEXT NOT NULL,
    action TEXT NOT NULL,
    details TEXT NOT NULL, -- JSON
    timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    success BOOLEAN NOT NULL DEFAULT 1,
    error_code TEXT,
    error_message TEXT,
    session_id TEXT,
    request_id TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Audit Events Indexes
CREATE INDEX IF NOT EXISTS idx_audit_events_entity ON audit_events(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_audit_events_user ON audit_events(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_events_timestamp ON audit_events(timestamp);
CREATE INDEX IF NOT EXISTS idx_audit_events_type_action ON audit_events(event_type, action);
CREATE INDEX IF NOT EXISTS idx_audit_events_success ON audit_events(success);
CREATE INDEX IF NOT EXISTS idx_audit_events_ip ON audit_events(ip_address);

-- Retry Tasks Table
CREATE TABLE IF NOT EXISTS retry_tasks (
    id TEXT PRIMARY KEY,
    task_type TEXT NOT NULL,
    entity_id TEXT NOT NULL,
    payload TEXT NOT NULL, -- JSON
    max_attempts INTEGER NOT NULL DEFAULT 3,
    current_attempt INTEGER NOT NULL DEFAULT 0,
    next_retry_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status TEXT NOT NULL DEFAULT 'pending', -- pending, running, completed, failed, cancelled
    last_error TEXT,
    completed_at DATETIME
);

-- Retry Tasks Indexes
CREATE INDEX IF NOT EXISTS idx_retry_tasks_status ON retry_tasks(status);
CREATE INDEX IF NOT EXISTS idx_retry_tasks_next_retry ON retry_tasks(next_retry_at);
CREATE INDEX IF NOT EXISTS idx_retry_tasks_type ON retry_tasks(task_type);
CREATE INDEX IF NOT EXISTS idx_retry_tasks_entity ON retry_tasks(entity_id);
CREATE INDEX IF NOT EXISTS idx_retry_tasks_created ON retry_tasks(created_at);

-- Retry Attempts Table
CREATE TABLE IF NOT EXISTS retry_attempts (
    id TEXT PRIMARY KEY,
    task_id TEXT NOT NULL,
    attempt_no INTEGER NOT NULL,
    started_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ended_at DATETIME,
    success BOOLEAN NOT NULL DEFAULT 0,
    error TEXT,
    duration_ms INTEGER, -- Duration in milliseconds
    FOREIGN KEY (task_id) REFERENCES retry_tasks(id) ON DELETE CASCADE
);

-- Retry Attempts Indexes
CREATE INDEX IF NOT EXISTS idx_retry_attempts_task ON retry_attempts(task_id);
CREATE INDEX IF NOT EXISTS idx_retry_attempts_success ON retry_attempts(success);
CREATE INDEX IF NOT EXISTS idx_retry_attempts_started ON retry_attempts(started_at);

-- Quotas Table
CREATE TABLE IF NOT EXISTS quotas (
    id TEXT PRIMARY KEY,
    entity_type TEXT NOT NULL, -- user, domain, global
    entity_id TEXT NOT NULL,
    quota_type TEXT NOT NULL, -- email_send, link_create, storage, bandwidth
    limit_value INTEGER NOT NULL,
    period TEXT NOT NULL, -- minute, hour, day, month
    reset_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN NOT NULL DEFAULT 1
);

-- Quotas Indexes
CREATE INDEX IF NOT EXISTS idx_quotas_entity ON quotas(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_quotas_type ON quotas(quota_type);
CREATE INDEX IF NOT EXISTS idx_quotas_reset ON quotas(reset_at);
CREATE INDEX IF NOT EXISTS idx_quotas_active ON quotas(is_active);
CREATE UNIQUE INDEX IF NOT EXISTS idx_quotas_unique ON quotas(entity_type, entity_id, quota_type) WHERE is_active = 1;

-- Quota Usage Table
CREATE TABLE IF NOT EXISTS quota_usage (
    id TEXT PRIMARY KEY,
    quota_id TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id TEXT NOT NULL,
    quota_type TEXT NOT NULL,
    usage INTEGER NOT NULL DEFAULT 0,
    period TEXT NOT NULL,
    period_start DATETIME NOT NULL,
    period_end DATETIME NOT NULL,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (quota_id) REFERENCES quotas(id) ON DELETE CASCADE
);

-- Quota Usage Indexes
CREATE INDEX IF NOT EXISTS idx_quota_usage_quota ON quota_usage(quota_id);
CREATE INDEX IF NOT EXISTS idx_quota_usage_entity ON quota_usage(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_quota_usage_period ON quota_usage(period_start, period_end);
CREATE INDEX IF NOT EXISTS idx_quota_usage_type ON quota_usage(quota_type);
CREATE UNIQUE INDEX IF NOT EXISTS idx_quota_usage_unique ON quota_usage(quota_id, entity_type, entity_id, period_start);

-- Performance Monitoring Table
CREATE TABLE IF NOT EXISTS performance_metrics (
    id TEXT PRIMARY KEY,
    metric_name TEXT NOT NULL,
    metric_type TEXT NOT NULL, -- counter, gauge, histogram
    value REAL NOT NULL,
    labels TEXT, -- JSON for additional labels
    timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    component TEXT NOT NULL,
    instance_id TEXT NOT NULL
);

-- Performance Metrics Indexes
CREATE INDEX IF NOT EXISTS idx_perf_metrics_name ON performance_metrics(metric_name);
CREATE INDEX IF NOT EXISTS idx_perf_metrics_timestamp ON performance_metrics(timestamp);
CREATE INDEX IF NOT EXISTS idx_perf_metrics_component ON performance_metrics(component);
CREATE INDEX IF NOT EXISTS idx_perf_metrics_type ON performance_metrics(metric_type);

-- System Health Table
CREATE TABLE IF NOT EXISTS system_health (
    id TEXT PRIMARY KEY,
    component TEXT NOT NULL,
    status TEXT NOT NULL, -- healthy, degraded, unhealthy
    last_check DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    details TEXT, -- JSON
    uptime_seconds INTEGER,
    memory_usage_mb REAL,
    cpu_usage_percent REAL,
    disk_usage_percent REAL,
    active_connections INTEGER,
    error_rate REAL
);

-- System Health Indexes
CREATE INDEX IF NOT EXISTS idx_system_health_component ON system_health(component);
CREATE INDEX IF NOT EXISTS idx_system_health_status ON system_health(status);
CREATE INDEX IF NOT EXISTS idx_system_health_check ON system_health(last_check);

-- Compliance Reports Table
CREATE TABLE IF NOT EXISTS compliance_reports (
    id TEXT PRIMARY KEY,
    report_type TEXT NOT NULL, -- gdpr, hipaa, sox, custom
    generated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    period_start DATETIME NOT NULL,
    period_end DATETIME NOT NULL,
    status TEXT NOT NULL DEFAULT 'generating', -- generating, completed, failed
    file_path TEXT,
    file_size INTEGER,
    summary TEXT, -- JSON
    generated_by TEXT,
    error_message TEXT
);

-- Compliance Reports Indexes
CREATE INDEX IF NOT EXISTS idx_compliance_reports_type ON compliance_reports(report_type);
CREATE INDEX IF NOT EXISTS idx_compliance_reports_generated ON compliance_reports(generated_at);
CREATE INDEX IF NOT EXISTS idx_compliance_reports_status ON compliance_reports(status);
CREATE INDEX IF NOT EXISTS idx_compliance_reports_period ON compliance_reports(period_start, period_end);

-- Data Retention Policies Table
CREATE TABLE IF NOT EXISTS data_retention_policies (
    id TEXT PRIMARY KEY,
    table_name TEXT NOT NULL,
    retention_period_days INTEGER NOT NULL,
    last_cleanup DATETIME,
    is_active BOOLEAN NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Data Retention Indexes
CREATE INDEX IF NOT EXISTS idx_retention_policies_table ON data_retention_policies(table_name);
CREATE INDEX IF NOT EXISTS idx_retention_policies_active ON data_retention_policies(is_active);
CREATE INDEX IF NOT EXISTS idx_retention_policies_cleanup ON data_retention_policies(last_cleanup);

-- Alert Rules Table
CREATE TABLE IF NOT EXISTS alert_rules (
    id TEXT PRIMARY KEY,
    rule_name TEXT NOT NULL,
    metric_name TEXT NOT NULL,
    condition_type TEXT NOT NULL, -- greater_than, less_than, equals, not_equals
    threshold_value REAL NOT NULL,
    duration_minutes INTEGER NOT NULL DEFAULT 5,
    severity TEXT NOT NULL DEFAULT 'warning', -- info, warning, critical
    is_active BOOLEAN NOT NULL DEFAULT 1,
    notification_channels TEXT, -- JSON array
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Alert Rules Indexes
CREATE INDEX IF NOT EXISTS idx_alert_rules_metric ON alert_rules(metric_name);
CREATE INDEX IF NOT EXISTS idx_alert_rules_active ON alert_rules(is_active);
CREATE INDEX IF NOT EXISTS idx_alert_rules_severity ON alert_rules(severity);

-- Alert Incidents Table
CREATE TABLE IF NOT EXISTS alert_incidents (
    id TEXT PRIMARY KEY,
    rule_id TEXT NOT NULL,
    started_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    resolved_at DATETIME,
    status TEXT NOT NULL DEFAULT 'firing', -- firing, resolved
    current_value REAL NOT NULL,
    threshold_value REAL NOT NULL,
    description TEXT,
    acknowledged_by TEXT,
    acknowledged_at DATETIME,
    FOREIGN KEY (rule_id) REFERENCES alert_rules(id) ON DELETE CASCADE
);

-- Alert Incidents Indexes
CREATE INDEX IF NOT EXISTS idx_alert_incidents_rule ON alert_incidents(rule_id);
CREATE INDEX IF NOT EXISTS idx_alert_incidents_status ON alert_incidents(status);
CREATE INDEX IF NOT EXISTS idx_alert_incidents_started ON alert_incidents(started_at);

-- Insert default data retention policies
INSERT OR IGNORE INTO data_retention_policies (id, table_name, retention_period_days, is_active) VALUES
('retention_audit_events', 'audit_events', 365, 1),
('retention_retry_tasks', 'retry_tasks', 30, 1),
('retention_retry_attempts', 'retry_attempts', 30, 1),
('retention_quota_usage', 'quota_usage', 90, 1),
('retention_performance_metrics', 'performance_metrics', 30, 1),
('retention_system_health', 'system_health', 7, 1),
('retention_compliance_reports', 'compliance_reports', 2555, 1), -- 7 years
('retention_alert_incidents', 'alert_incidents', 90, 1);

-- Insert default alert rules
INSERT OR IGNORE INTO alert_rules (id, rule_name, metric_name, condition_type, threshold_value, duration_minutes, severity, is_active) VALUES
('alert_high_error_rate', 'High Error Rate', 'error_rate', 'greater_than', 0.05, 5, 'warning', 1),
('alert_high_response_time', 'High Response Time', 'response_time_ms', 'greater_than', 5000, 5, 'warning', 1),
('alert_high_memory_usage', 'High Memory Usage', 'memory_usage_percent', 'greater_than', 85, 10, 'warning', 1),
('alert_high_cpu_usage', 'High CPU Usage', 'cpu_usage_percent', 'greater_than', 90, 10, 'critical', 1),
('alert_disk_space', 'Low Disk Space', 'disk_usage_percent', 'greater_than', 90, 5, 'critical', 1),
('alert_failed_logins', 'High Failed Login Rate', 'failed_login_rate', 'greater_than', 10, 5, 'warning', 1);

-- Insert default global quotas
INSERT OR IGNORE INTO quotas (id, entity_type, entity_id, quota_type, limit_value, period, reset_at, is_active) VALUES
('quota_global_email_send', 'global', 'default', 'email_send', 10000, 'hour', datetime('now', '+1 hour'), 1),
('quota_global_link_create', 'global', 'default', 'link_create', 50000, 'hour', datetime('now', '+1 hour'), 1),
('quota_global_storage', 'global', 'default', 'storage', 107374182400, 'month', datetime('now', '+1 month'), 1), -- 100GB
('quota_user_email_send', 'user', 'default', 'email_send', 1000, 'hour', datetime('now', '+1 hour'), 1),
('quota_user_link_create', 'user', 'default', 'link_create', 5000, 'hour', datetime('now', '+1 hour'), 1),
('quota_user_storage', 'user', 'default', 'storage', 1073741824, 'month', datetime('now', '+1 month'), 1); -- 1GB

-- Create views for easier querying

-- Audit summary view
CREATE VIEW IF NOT EXISTS audit_summary AS
SELECT 
    DATE(timestamp) as audit_date,
    event_type,
    action,
    COUNT(*) as event_count,
    SUM(CASE WHEN success = 1 THEN 1 ELSE 0 END) as successful_events,
    SUM(CASE WHEN success = 0 THEN 1 ELSE 0 END) as failed_events
FROM audit_events
GROUP BY DATE(timestamp), event_type, action;

-- Active retry tasks view
CREATE VIEW IF NOT EXISTS active_retry_tasks AS
SELECT 
    rt.*,
    COALESCE(attempts.attempt_count, 0) as total_attempts,
    COALESCE(attempts.last_attempt, rt.created_at) as last_attempt_time
FROM retry_tasks rt
LEFT JOIN (
    SELECT 
        task_id,
        COUNT(*) as attempt_count,
        MAX(started_at) as last_attempt
    FROM retry_attempts
    GROUP BY task_id
) attempts ON rt.id = attempts.task_id
WHERE rt.status IN ('pending', 'running');

-- Quota utilization view
CREATE VIEW IF NOT EXISTS quota_utilization AS
SELECT 
    q.id as quota_id,
    q.entity_type,
    q.entity_id,
    q.quota_type,
    q.limit_value,
    COALESCE(qu.usage, 0) as current_usage,
    ROUND((COALESCE(qu.usage, 0) * 100.0 / q.limit_value), 2) as utilization_percent,
    q.period,
    q.reset_at
FROM quotas q
LEFT JOIN quota_usage qu ON q.id = qu.quota_id
WHERE q.is_active = 1;

-- System health summary view
CREATE VIEW IF NOT EXISTS system_health_summary AS
SELECT 
    component,
    status,
    last_check,
    ROUND(AVG(memory_usage_mb), 2) as avg_memory_mb,
    ROUND(AVG(cpu_usage_percent), 2) as avg_cpu_percent,
    ROUND(AVG(disk_usage_percent), 2) as avg_disk_percent,
    AVG(active_connections) as avg_connections,
    ROUND(AVG(error_rate), 4) as avg_error_rate
FROM system_health
WHERE last_check > datetime('now', '-1 hour')
GROUP BY component;
