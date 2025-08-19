-- Migration: Add retention insights and policy recommendations (Micro-Iteration 4.27)
-- This migration adds support for intelligent retention insights and proactive policy recommendations

-- Retention insights table for storing daily rollups and analytics
CREATE TABLE IF NOT EXISTS retention_insights (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    insight_date DATE NOT NULL, -- Date for which insights are calculated
    insight_type TEXT NOT NULL, -- "daily_rollup", "trend_analysis", "policy_effectiveness"
    
    -- Policy trigger insights
    most_common_policy_triggers TEXT, -- JSON array of trigger types and counts
    volume_trends_archived_vs_deleted TEXT, -- JSON with archived/deleted volume trends
    avg_storage_savings_compression REAL DEFAULT 0.0, -- Average compression ratio
    policy_effectiveness_score REAL DEFAULT 0.0, -- 0.0 to 1.0 effectiveness score
    
    -- Storage and cost insights
    total_storage_savings_bytes INTEGER DEFAULT 0,
    estimated_cost_savings_usd REAL DEFAULT 0.0,
    compression_ratio_avg REAL DEFAULT 0.0,
    
    -- Policy performance metrics
    policies_most_effective TEXT, -- JSON array of most effective policies
    policies_least_effective TEXT, -- JSON array of least effective policies
    override_frequency INTEGER DEFAULT 0, -- Number of policy overrides
    
    -- Metadata
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(insight_date, insight_type)
);

-- Policy recommendations table
CREATE TABLE IF NOT EXISTS retention_recommendations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    recommendation_type TEXT NOT NULL, -- "policy_optimization", "storage_optimization", "risk_mitigation"
    priority TEXT NOT NULL, -- "low", "medium", "high", "critical"
    
    -- Recommendation details
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    current_state TEXT, -- JSON describing current state
    recommended_action TEXT NOT NULL, -- JSON describing recommended action
    expected_impact TEXT, -- JSON describing expected impact
    
    -- Impact scoring
    impact_score REAL DEFAULT 0.0, -- 0.0 to 1.0 impact score
    confidence_score REAL DEFAULT 0.0, -- 0.0 to 1.0 confidence in recommendation
    risk_level TEXT DEFAULT "low", -- "low", "medium", "high"
    
    -- Applicable scope
    user_id TEXT, -- Specific user if applicable
    domain TEXT, -- Specific domain if applicable
    policy_id INTEGER, -- Specific policy if applicable
    
    -- Status tracking
    status TEXT DEFAULT "pending", -- "pending", "approved", "applied", "rejected"
    applied_at TIMESTAMP,
    applied_by TEXT,
    applied_result TEXT, -- "success", "partial", "failed"
    
    -- Metadata
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (policy_id) REFERENCES retention_policies(id)
);

-- Enhanced policy evaluation logs with impact scoring
ALTER TABLE policy_evaluation_logs ADD COLUMN impact_score REAL DEFAULT 0.0;
ALTER TABLE policy_evaluation_logs ADD COLUMN storage_savings_bytes INTEGER DEFAULT 0;
ALTER TABLE policy_evaluation_logs ADD COLUMN archival_load_impact REAL DEFAULT 0.0;

-- Retention recommendation application logs
CREATE TABLE IF NOT EXISTS recommendation_application_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    recommendation_id INTEGER NOT NULL,
    applied_by TEXT NOT NULL,
    application_type TEXT NOT NULL, -- "preview", "apply", "reject"
    changes_applied TEXT, -- JSON describing what was applied
    result_summary TEXT, -- Summary of application results
    affected_policies INTEGER DEFAULT 0, -- Number of policies affected
    affected_emails INTEGER DEFAULT 0, -- Number of emails affected
    estimated_savings_bytes INTEGER DEFAULT 0, -- Estimated storage savings
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (recommendation_id) REFERENCES retention_recommendations(id)
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_retention_insights_date ON retention_insights(insight_date);
CREATE INDEX IF NOT EXISTS idx_retention_insights_type ON retention_insights(insight_type);
CREATE INDEX IF NOT EXISTS idx_retention_insights_date_type ON retention_insights(insight_date, insight_type);

CREATE INDEX IF NOT EXISTS idx_retention_recommendations_type ON retention_recommendations(recommendation_type);
CREATE INDEX IF NOT EXISTS idx_retention_recommendations_priority ON retention_recommendations(priority);
CREATE INDEX IF NOT EXISTS idx_retention_recommendations_status ON retention_recommendations(status);
CREATE INDEX IF NOT EXISTS idx_retention_recommendations_user_id ON retention_recommendations(user_id);
CREATE INDEX IF NOT EXISTS idx_retention_recommendations_domain ON retention_recommendations(domain);
CREATE INDEX IF NOT EXISTS idx_retention_recommendations_policy_id ON retention_recommendations(policy_id);
CREATE INDEX IF NOT EXISTS idx_retention_recommendations_impact_score ON retention_recommendations(impact_score);

CREATE INDEX IF NOT EXISTS idx_policy_evaluation_impact_score ON policy_evaluation_logs(impact_score);
CREATE INDEX IF NOT EXISTS idx_policy_evaluation_storage_savings ON policy_evaluation_logs(storage_savings_bytes);

CREATE INDEX IF NOT EXISTS idx_recommendation_applications_recommendation_id ON recommendation_application_logs(recommendation_id);
CREATE INDEX IF NOT EXISTS idx_recommendation_applications_applied_at ON recommendation_application_logs(applied_at);

-- Create views for insights and recommendations
CREATE VIEW IF NOT EXISTS retention_insights_summary AS
SELECT 
    insight_date,
    insight_type,
    policy_effectiveness_score,
    avg_storage_savings_compression,
    total_storage_savings_bytes,
    estimated_cost_savings_usd,
    override_frequency
FROM retention_insights
ORDER BY insight_date DESC, insight_type;

CREATE VIEW IF NOT EXISTS retention_recommendations_summary AS
SELECT 
    recommendation_type,
    priority,
    status,
    COUNT(*) as total_recommendations,
    AVG(impact_score) as avg_impact_score,
    AVG(confidence_score) as avg_confidence_score,
    SUM(CASE WHEN status = 'applied' THEN 1 ELSE 0 END) as applied_count
FROM retention_recommendations
GROUP BY recommendation_type, priority, status;

CREATE VIEW IF NOT EXISTS policy_effectiveness_ranking AS
SELECT 
    p.id as policy_id,
    p.name as policy_name,
    p.retention_days,
    p.archive_instead,
    COUNT(pel.id) as total_evaluations,
    AVG(pel.impact_score) as avg_impact_score,
    SUM(pel.storage_savings_bytes) as total_storage_savings,
    AVG(pel.archival_load_impact) as avg_archival_load_impact
FROM retention_policies p
LEFT JOIN policy_evaluation_logs pel ON p.id = pel.policy_id
WHERE p.active = 1
GROUP BY p.id, p.name, p.retention_days, p.archive_instead
ORDER BY avg_impact_score DESC;

-- Create triggers for automatic timestamp updates
CREATE TRIGGER IF NOT EXISTS update_retention_insights_updated_at
    AFTER UPDATE ON retention_insights
    FOR EACH ROW
BEGIN
    UPDATE retention_insights SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS update_retention_recommendations_updated_at
    AFTER UPDATE ON retention_recommendations
    FOR EACH ROW
BEGIN
    UPDATE retention_recommendations SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- Create cleanup trigger for old insights (keep last 90 days)
CREATE TRIGGER IF NOT EXISTS cleanup_old_retention_insights
    AFTER INSERT ON retention_insights
    FOR EACH ROW
BEGIN
    DELETE FROM retention_insights 
    WHERE insight_date < date('now', '-90 days');
END;

-- Create cleanup trigger for old recommendations (keep last 180 days)
CREATE TRIGGER IF NOT EXISTS cleanup_old_retention_recommendations
    AFTER INSERT ON retention_recommendations
    FOR EACH ROW
BEGIN
    DELETE FROM retention_recommendations 
    WHERE created_at < datetime('now', '-180 days') AND status IN ('applied', 'rejected');
END;

-- Insert default configuration for insights and recommendations
INSERT OR IGNORE INTO retention_insights (
    insight_date, insight_type, policy_effectiveness_score, 
    avg_storage_savings_compression, total_storage_savings_bytes, 
    estimated_cost_savings_usd, override_frequency, created_at, updated_at
) VALUES (
    date('now'), 'daily_rollup', 0.75, 0.65, 0, 0.0, 0, 
    CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
);





