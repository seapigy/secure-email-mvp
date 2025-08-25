-- Migration: Add predictive retention forecasting and anomaly detection (Micro-Iteration 4.29)
-- This migration adds support for predictive analytics and anomaly detection for retention operations

-- Retention forecasts table for storing predictive analytics
CREATE TABLE IF NOT EXISTS retention_forecasts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    forecast_type TEXT NOT NULL, -- "user", "domain", "global"
    forecast_key TEXT NOT NULL, -- user_id, domain, or "global"
    generated_at TIMESTAMP NOT NULL,
    target_period_end TIMESTAMP NOT NULL,
    
    -- Predicted metrics
    predicted_usage_bytes INTEGER DEFAULT 0,
    predicted_archival_count INTEGER DEFAULT 0,
    predicted_deletion_count INTEGER DEFAULT 0,
    predicted_policy_impact REAL DEFAULT 0.0,
    predicted_storage_savings_bytes INTEGER DEFAULT 0,
    predicted_cost_savings_usd REAL DEFAULT 0.0,
    
    -- Forecast confidence and accuracy
    confidence_score REAL DEFAULT 0.0, -- 0.0 to 1.0
    accuracy_score REAL DEFAULT 0.0, -- Historical accuracy of this forecast type
    forecast_model_version TEXT DEFAULT "v1.0",
    
    -- Input data summary
    historical_data_points INTEGER DEFAULT 0,
    data_freshness_hours INTEGER DEFAULT 0,
    
    -- Metadata
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(forecast_type, forecast_key, target_period_end)
);

-- Retention anomalies table for storing detected anomalies
CREATE TABLE IF NOT EXISTS retention_anomalies (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    anomaly_type TEXT NOT NULL, -- "spike_deletion", "drop_policy_matches", "forecast_deviation", "unusual_archival"
    severity TEXT NOT NULL, -- "low", "medium", "high", "critical"
    
    -- Anomaly details
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    detected_at TIMESTAMP NOT NULL,
    
    -- Scope and context
    scope_type TEXT NOT NULL, -- "user", "domain", "global"
    scope_key TEXT, -- user_id, domain, or NULL for global
    
    -- Anomaly metrics
    baseline_value REAL DEFAULT 0.0,
    current_value REAL DEFAULT 0.0,
    deviation_percentage REAL DEFAULT 0.0,
    threshold_percentage REAL DEFAULT 0.0,
    
    -- Related data
    affected_policies TEXT, -- JSON array of affected policy IDs
    affected_emails_count INTEGER DEFAULT 0,
    time_window_hours INTEGER DEFAULT 24,
    
    -- Status and resolution
    status TEXT DEFAULT "active", -- "active", "acknowledged", "resolved", "false_positive"
    acknowledged_at TIMESTAMP,
    acknowledged_by TEXT,
    resolution_notes TEXT,
    resolved_at TIMESTAMP,
    resolved_by TEXT,
    
    -- Recommended actions
    recommended_action TEXT, -- JSON describing recommended action
    auto_resolution_enabled BOOLEAN DEFAULT FALSE,
    
    -- Metadata
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Forecast accuracy tracking table
CREATE TABLE IF NOT EXISTS forecast_accuracy_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    forecast_id INTEGER NOT NULL,
    actual_usage_bytes INTEGER DEFAULT 0,
    actual_archival_count INTEGER DEFAULT 0,
    actual_deletion_count INTEGER DEFAULT 0,
    actual_policy_impact REAL DEFAULT 0.0,
    actual_storage_savings_bytes INTEGER DEFAULT 0,
    actual_cost_savings_usd REAL DEFAULT 0.0,
    
    -- Accuracy metrics
    usage_accuracy_percentage REAL DEFAULT 0.0,
    archival_accuracy_percentage REAL DEFAULT 0.0,
    deletion_accuracy_percentage REAL DEFAULT 0.0,
    policy_impact_accuracy_percentage REAL DEFAULT 0.0,
    overall_accuracy_score REAL DEFAULT 0.0,
    
    -- Evaluation metadata
    evaluated_at TIMESTAMP NOT NULL,
    evaluation_window_hours INTEGER DEFAULT 24,
    
    -- Metadata
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (forecast_id) REFERENCES retention_forecasts(id)
);

-- Anomaly detection configuration table
CREATE TABLE IF NOT EXISTS anomaly_detection_config (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    config_key TEXT UNIQUE NOT NULL,
    config_value TEXT NOT NULL,
    description TEXT,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_retention_forecasts_type_key ON retention_forecasts(forecast_type, forecast_key);
CREATE INDEX IF NOT EXISTS idx_retention_forecasts_period_end ON retention_forecasts(target_period_end);
CREATE INDEX IF NOT EXISTS idx_retention_forecasts_generated_at ON retention_forecasts(generated_at);
CREATE INDEX IF NOT EXISTS idx_retention_forecasts_confidence ON retention_forecasts(confidence_score);

CREATE INDEX IF NOT EXISTS idx_retention_anomalies_type ON retention_anomalies(anomaly_type);
CREATE INDEX IF NOT EXISTS idx_retention_anomalies_severity ON retention_anomalies(severity);
CREATE INDEX IF NOT EXISTS idx_retention_anomalies_status ON retention_anomalies(status);
CREATE INDEX IF NOT EXISTS idx_retention_anomalies_scope ON retention_anomalies(scope_type, scope_key);
CREATE INDEX IF NOT EXISTS idx_retention_anomalies_detected_at ON retention_anomalies(detected_at);

CREATE INDEX IF NOT EXISTS idx_forecast_accuracy_forecast_id ON forecast_accuracy_logs(forecast_id);
CREATE INDEX IF NOT EXISTS idx_forecast_accuracy_evaluated_at ON forecast_accuracy_logs(evaluated_at);
CREATE INDEX IF NOT EXISTS idx_forecast_accuracy_overall_score ON forecast_accuracy_logs(overall_accuracy_score);

-- Create views for analytics
CREATE VIEW IF NOT EXISTS retention_forecasts_summary AS
SELECT 
    forecast_type,
    COUNT(*) as total_forecasts,
    AVG(confidence_score) as avg_confidence,
    AVG(accuracy_score) as avg_accuracy,
    MAX(generated_at) as latest_forecast,
    SUM(predicted_usage_bytes) as total_predicted_usage,
    SUM(predicted_storage_savings_bytes) as total_predicted_savings
FROM retention_forecasts
GROUP BY forecast_type;

CREATE VIEW IF NOT EXISTS retention_anomalies_summary AS
SELECT 
    anomaly_type,
    severity,
    status,
    COUNT(*) as total_anomalies,
    AVG(deviation_percentage) as avg_deviation,
    MAX(detected_at) as latest_anomaly
FROM retention_anomalies
GROUP BY anomaly_type, severity, status;

CREATE VIEW IF NOT EXISTS forecast_accuracy_summary AS
SELECT 
    COUNT(*) as total_evaluations,
    AVG(overall_accuracy_score) as avg_accuracy,
    AVG(usage_accuracy_percentage) as avg_usage_accuracy,
    AVG(archival_accuracy_percentage) as avg_archival_accuracy,
    AVG(deletion_accuracy_percentage) as avg_deletion_accuracy,
    MAX(evaluated_at) as latest_evaluation
FROM forecast_accuracy_logs;

-- Create triggers for automatic timestamp updates
CREATE TRIGGER IF NOT EXISTS update_retention_forecasts_updated_at
    AFTER UPDATE ON retention_forecasts
    FOR EACH ROW
BEGIN
    UPDATE retention_forecasts SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS update_retention_anomalies_updated_at
    AFTER UPDATE ON retention_anomalies
    FOR EACH ROW
BEGIN
    UPDATE retention_anomalies SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS update_anomaly_detection_config_updated_at
    AFTER UPDATE ON anomaly_detection_config
    FOR EACH ROW
BEGIN
    UPDATE anomaly_detection_config SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- Insert default anomaly detection configuration
INSERT OR IGNORE INTO anomaly_detection_config (config_key, config_value, description) VALUES
('spike_deletion_threshold', '200.0', 'Percentage increase in deletions to trigger anomaly'),
('drop_policy_matches_threshold', '50.0', 'Percentage decrease in policy matches to trigger anomaly'),
('forecast_deviation_threshold', '25.0', 'Percentage deviation from forecast to trigger anomaly'),
('unusual_archival_threshold', '150.0', 'Percentage increase in archival operations to trigger anomaly'),
('anomaly_detection_window_hours', '24', 'Time window for anomaly detection in hours'),
('auto_resolution_enabled', 'false', 'Whether to enable automatic anomaly resolution'),
('min_confidence_threshold', '0.8', 'Minimum confidence score for forecasts'),
('forecast_periods_days', '7,30,90', 'Forecast periods in days');

-- Create cleanup triggers for old data
CREATE TRIGGER IF NOT EXISTS cleanup_old_forecasts
    AFTER INSERT ON retention_forecasts
    FOR EACH ROW
BEGIN
    DELETE FROM retention_forecasts 
    WHERE generated_at < datetime('now', '-90 days');
END;

CREATE TRIGGER IF NOT EXISTS cleanup_old_anomalies
    AFTER INSERT ON retention_anomalies
    FOR EACH ROW
BEGIN
    DELETE FROM retention_anomalies 
    WHERE detected_at < datetime('now', '-30 days') AND status IN ('resolved', 'false_positive');
END;

CREATE TRIGGER IF NOT EXISTS cleanup_old_forecast_accuracy
    AFTER INSERT ON forecast_accuracy_logs
    FOR EACH ROW
BEGIN
    DELETE FROM forecast_accuracy_logs 
    WHERE evaluated_at < datetime('now', '-60 days');
END;






