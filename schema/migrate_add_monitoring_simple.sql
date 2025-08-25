-- Iteration 9: Real-Time Monitoring & Dashboards Migration
-- Simplified version without system_config dependencies

-- Create monitoring_events table
CREATE TABLE IF NOT EXISTS monitoring_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    event_type VARCHAR(100) NOT NULL,
    event_subtype VARCHAR(100),
    timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    metadata TEXT, -- JSON metadata
    severity VARCHAR(20) NOT NULL DEFAULT 'info',
    source VARCHAR(100) NOT NULL,
    user_id VARCHAR(100),
    session_id VARCHAR(100),
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for monitoring_events
CREATE INDEX IF NOT EXISTS idx_monitoring_events_type ON monitoring_events(event_type);
CREATE INDEX IF NOT EXISTS idx_monitoring_events_timestamp ON monitoring_events(timestamp);
CREATE INDEX IF NOT EXISTS idx_monitoring_events_severity ON monitoring_events(severity);
CREATE INDEX IF NOT EXISTS idx_monitoring_events_source ON monitoring_events(source);
CREATE INDEX IF NOT EXISTS idx_monitoring_events_user_id ON monitoring_events(user_id);
CREATE INDEX IF NOT EXISTS idx_monitoring_events_session_id ON monitoring_events(session_id);

-- Create composite index for common queries
CREATE INDEX IF NOT EXISTS idx_monitoring_events_type_timestamp ON monitoring_events(event_type, timestamp);
CREATE INDEX IF NOT EXISTS idx_monitoring_events_source_timestamp ON monitoring_events(source, timestamp);

-- Create retention policy trigger (delete events older than 30 days)
CREATE TRIGGER IF NOT EXISTS trigger_monitoring_events_retention
    AFTER INSERT ON monitoring_events
    BEGIN
        DELETE FROM monitoring_events 
        WHERE timestamp < datetime('now', '-30 days');
    END;

-- Create monitoring_metrics_summary table for aggregated metrics
CREATE TABLE IF NOT EXISTS monitoring_metrics_summary (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    metric_name VARCHAR(100) NOT NULL,
    metric_value REAL NOT NULL,
    metric_unit VARCHAR(20), -- count, percentage, milliseconds, etc.
    time_bucket VARCHAR(20) NOT NULL, -- minute, hour, day
    bucket_start DATETIME NOT NULL,
    bucket_end DATETIME NOT NULL,
    source VARCHAR(100),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(metric_name, time_bucket, bucket_start, source)
);

-- Create indexes for metrics summary
CREATE INDEX IF NOT EXISTS idx_metrics_summary_name ON monitoring_metrics_summary(metric_name);
CREATE INDEX IF NOT EXISTS idx_metrics_summary_bucket ON monitoring_metrics_summary(time_bucket, bucket_start);
CREATE INDEX IF NOT EXISTS idx_metrics_summary_source ON monitoring_metrics_summary(source);

-- Insert sample monitoring events for testing
INSERT OR IGNORE INTO monitoring_events (event_type, event_subtype, metadata, severity, source) VALUES
('system.startup', 'api_server', '{"version": "1.0.0", "port": 8080}', 'info', 'api'),
('dlp.scan', 'content_analysis', '{"content_type": "email", "scan_result": "clean", "processing_time_ms": 45}', 'info', 'dlp'),
('watermarking.apply', 'text_watermark', '{"watermark_type": "text", "content_type": "pdf", "processing_time_ms": 120}', 'info', 'watermarking'),
('api.request', 'endpoint_call', '{"endpoint": "/api/watermark/templates", "method": "GET", "status_code": 200, "latency_ms": 15}', 'info', 'api'),
('security.alert', 'failed_login', '{"ip_address": "192.168.1.100", "attempts": 3}', 'warning', 'security');

-- Create view for real-time metrics
CREATE VIEW IF NOT EXISTS v_monitoring_realtime_metrics AS
SELECT 
    'request_count' as metric_name,
    COUNT(*) as metric_value,
    'count' as metric_unit,
    'minute' as time_bucket,
    datetime('now', '-1 minute') as bucket_start,
    datetime('now') as bucket_end,
    source
FROM monitoring_events 
WHERE event_type = 'api.request' 
    AND timestamp >= datetime('now', '-1 minute')
GROUP BY source

UNION ALL

SELECT 
    'error_rate' as metric_name,
    (COUNT(CASE WHEN severity IN ('error', 'critical') THEN 1 END) * 100.0 / COUNT(*)) as metric_value,
    'percentage' as metric_unit,
    'minute' as time_bucket,
    datetime('now', '-1 minute') as bucket_start,
    datetime('now') as bucket_end,
    source
FROM monitoring_events 
WHERE timestamp >= datetime('now', '-1 minute')
GROUP BY source

UNION ALL

SELECT 
    'avg_latency' as metric_name,
    AVG(CAST(json_extract(metadata, '$.latency_ms') AS REAL)) as metric_value,
    'milliseconds' as metric_unit,
    'minute' as time_bucket,
    datetime('now', '-1 minute') as bucket_start,
    datetime('now') as bucket_end,
    source
FROM monitoring_events 
WHERE event_type = 'api.request' 
    AND timestamp >= datetime('now', '-1 minute')
    AND json_extract(metadata, '$.latency_ms') IS NOT NULL
GROUP BY source;

-- Log migration completion
INSERT INTO monitoring_events (event_type, event_subtype, metadata, severity, source) VALUES
('migration.complete', 'iteration9_monitoring', '{"migration": "add_monitoring", "version": "1.0.0"}', 'info', 'system');
