-- =============================================================================
-- SES TRANSACTIONS TABLE
-- =============================================================================
-- This table stores Amazon SES transaction information for audit logging
-- and compliance purposes. Each successful email send via SES creates
-- a transaction record with detailed metadata.
-- =============================================================================

CREATE TABLE IF NOT EXISTS ses_transactions (
    -- Primary identifier
    transaction_id TEXT PRIMARY KEY NOT NULL,
    
    -- Email metadata
    message_id TEXT NOT NULL,
    sender_id TEXT NOT NULL,
    recipient TEXT NOT NULL,
    blob_id TEXT,
    
    -- Transaction details
    timestamp DATETIME NOT NULL,
    status TEXT NOT NULL DEFAULT 'sent',
    retry_count INTEGER NOT NULL DEFAULT 0,
    
    -- Audit fields
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Indexes for efficient querying
    INDEX idx_ses_transactions_message_id (message_id),
    INDEX idx_ses_transactions_sender_id (sender_id),
    INDEX idx_ses_transactions_recipient (recipient),
    INDEX idx_ses_transactions_timestamp (timestamp),
    INDEX idx_ses_transactions_status (status),
    INDEX idx_ses_transactions_date (DATE(timestamp))
);

-- =============================================================================
-- SES QUOTA TRACKING TABLE
-- =============================================================================
-- This table tracks SES quota usage for rate limiting and monitoring
-- =============================================================================

CREATE TABLE IF NOT EXISTS ses_quota_usage (
    -- Primary identifier
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    
    -- Date tracking
    usage_date DATE NOT NULL,
    
    -- Quota metrics
    daily_quota INTEGER NOT NULL,
    sent_count INTEGER NOT NULL DEFAULT 0,
    failed_count INTEGER NOT NULL DEFAULT 0,
    rate_limit INTEGER NOT NULL,
    
    -- Timestamps
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Ensure one record per date
    UNIQUE(usage_date)
);

-- =============================================================================
-- SES VALIDATION LOGS TABLE
-- =============================================================================
-- This table logs PQC + KT validation results for compliance and debugging
-- =============================================================================

CREATE TABLE IF NOT EXISTS ses_validation_logs (
    -- Primary identifier
    validation_id TEXT PRIMARY KEY NOT NULL,
    
    -- Email reference
    email_id TEXT NOT NULL,
    sender_id TEXT NOT NULL,
    recipient TEXT,
    
    -- Validation results
    pqc_valid BOOLEAN NOT NULL,
    kt_valid BOOLEAN NOT NULL,
    overall_valid BOOLEAN NOT NULL,
    
    -- Error details
    error_code TEXT,
    error_message TEXT,
    
    -- Timestamps
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Indexes
    INDEX idx_ses_validation_email_id (email_id),
    INDEX idx_ses_validation_sender_id (sender_id),
    INDEX idx_ses_validation_created_at (created_at),
    INDEX idx_ses_validation_overall_valid (overall_valid)
);

-- =============================================================================
-- TRIGGERS FOR UPDATED_AT TIMESTAMPS
-- =============================================================================

-- Trigger to update updated_at timestamp for ses_transactions
CREATE TRIGGER IF NOT EXISTS update_ses_transactions_updated_at
    AFTER UPDATE ON ses_transactions
    FOR EACH ROW
BEGIN
    UPDATE ses_transactions SET updated_at = CURRENT_TIMESTAMP WHERE transaction_id = NEW.transaction_id;
END;

-- Trigger to update updated_at timestamp for ses_quota_usage
CREATE TRIGGER IF NOT EXISTS update_ses_quota_usage_updated_at
    AFTER UPDATE ON ses_quota_usage
    FOR EACH ROW
BEGIN
    UPDATE ses_quota_usage SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- =============================================================================
-- VIEWS FOR EASY QUERYING
-- =============================================================================

-- View for daily SES statistics
CREATE VIEW IF NOT EXISTS ses_daily_stats AS
SELECT 
    DATE(timestamp) as date,
    COUNT(*) as total_sent,
    COUNT(CASE WHEN status = 'sent' THEN 1 END) as successful_sends,
    COUNT(CASE WHEN status != 'sent' THEN 1 END) as failed_sends,
    AVG(retry_count) as avg_retry_count,
    MAX(retry_count) as max_retry_count
FROM ses_transactions
GROUP BY DATE(timestamp)
ORDER BY date DESC;

-- View for sender statistics
CREATE VIEW IF NOT EXISTS ses_sender_stats AS
SELECT 
    sender_id,
    COUNT(*) as total_sent,
    COUNT(CASE WHEN status = 'sent' THEN 1 END) as successful_sends,
    COUNT(CASE WHEN status != 'sent' THEN 1 END) as failed_sends,
    AVG(retry_count) as avg_retry_count,
    MAX(timestamp) as last_sent
FROM ses_transactions
GROUP BY sender_id
ORDER BY total_sent DESC;

-- View for validation statistics
CREATE VIEW IF NOT EXISTS ses_validation_stats AS
SELECT 
    DATE(created_at) as date,
    COUNT(*) as total_validations,
    COUNT(CASE WHEN overall_valid = 1 THEN 1 END) as successful_validations,
    COUNT(CASE WHEN overall_valid = 0 THEN 1 END) as failed_validations,
    COUNT(CASE WHEN pqc_valid = 1 THEN 1 END) as pqc_successful,
    COUNT(CASE WHEN kt_valid = 1 THEN 1 END) as kt_successful
FROM ses_validation_logs
GROUP BY DATE(created_at)
ORDER BY date DESC;


