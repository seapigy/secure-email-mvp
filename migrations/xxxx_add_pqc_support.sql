-- =============================================================================
-- PQC (Post-Quantum Cryptography) Support Migration
-- =============================================================================
-- This migration adds support for hybrid PQC + symmetric encryption
-- to the Secure Email MVP system.
-- =============================================================================

-- Add PQC-related columns to the emails table
ALTER TABLE emails ADD COLUMN pqc_enabled BOOLEAN DEFAULT FALSE;
ALTER TABLE emails ADD COLUMN pqc_encrypted_data TEXT; -- JSON serialized hybrid encrypted data
ALTER TABLE emails ADD COLUMN encryption_version TEXT DEFAULT 'AES-256-GCM'; -- Track encryption method
ALTER TABLE emails ADD COLUMN pqc_key_id TEXT; -- HSM key identifier
ALTER TABLE emails ADD COLUMN pqc_encryption_time TIMESTAMP; -- When PQC encryption was applied

-- Create PQC key management table
CREATE TABLE IF NOT EXISTS pqc_keys (
    id TEXT PRIMARY KEY,
    key_id TEXT UNIQUE NOT NULL, -- HSM key identifier
    kyber_level INTEGER NOT NULL, -- 512, 768, or 1024
    public_key TEXT NOT NULL, -- Base64 encoded public key
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    hsm_key_id TEXT, -- HSM-specific key identifier
    rotation_count INTEGER DEFAULT 0
);

-- Create PQC audit log table
CREATE TABLE IF NOT EXISTS pqc_audit_log (
    id TEXT PRIMARY KEY,
    event_type TEXT NOT NULL,
    description TEXT NOT NULL,
    severity TEXT NOT NULL, -- INFO, WARN, ERROR, CRITICAL
    user_id TEXT,
    ip_address TEXT,
    session_id TEXT,
    details TEXT, -- JSON serialized event details
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    key_id TEXT, -- Associated PQC key ID
    email_id TEXT, -- Associated email ID
    FOREIGN KEY (key_id) REFERENCES pqc_keys(id),
    FOREIGN KEY (email_id) REFERENCES emails(id)
);

-- Create PQC performance metrics table
CREATE TABLE IF NOT EXISTS pqc_performance_metrics (
    id TEXT PRIMARY KEY,
    operation TEXT NOT NULL, -- ENCRYPT, DECRYPT, KEY_GENERATION, etc.
    duration_ms INTEGER NOT NULL, -- Duration in milliseconds
    data_size INTEGER, -- Size of data processed in bytes
    kyber_level INTEGER, -- Kyber security level used
    hsm_enabled BOOLEAN, -- Whether HSM was used
    success BOOLEAN NOT NULL, -- Whether operation succeeded
    error_message TEXT, -- Error message if failed
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    context TEXT -- Operation context (email_id, user_id, etc.)
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_emails_pqc_enabled ON emails(pqc_enabled);
CREATE INDEX IF NOT EXISTS idx_emails_encryption_version ON emails(encryption_version);
CREATE INDEX IF NOT EXISTS idx_emails_pqc_key_id ON emails(pqc_key_id);
CREATE INDEX IF NOT EXISTS idx_pqc_keys_active ON pqc_keys(is_active);
CREATE INDEX IF NOT EXISTS idx_pqc_keys_expires ON pqc_keys(expires_at);
CREATE INDEX IF NOT EXISTS idx_pqc_audit_log_event_type ON pqc_audit_log(event_type);
CREATE INDEX IF NOT EXISTS idx_pqc_audit_log_timestamp ON pqc_audit_log(timestamp);
CREATE INDEX IF NOT EXISTS idx_pqc_audit_log_severity ON pqc_audit_log(severity);
CREATE INDEX IF NOT EXISTS idx_pqc_performance_operation ON pqc_performance_metrics(operation);
CREATE INDEX IF NOT EXISTS idx_pqc_performance_timestamp ON pqc_performance_metrics(timestamp);

-- Create view for PQC statistics
CREATE VIEW IF NOT EXISTS pqc_statistics AS
SELECT 
    COUNT(*) as total_emails,
    COUNT(CASE WHEN pqc_enabled = TRUE THEN 1 END) as pqc_encrypted_emails,
    COUNT(CASE WHEN pqc_enabled = FALSE THEN 1 END) as aes_only_emails,
    COUNT(CASE WHEN encryption_version = 'AES-256-GCM' THEN 1 END) as aes_encrypted_emails,
    COUNT(CASE WHEN encryption_version = 'PQC-HYBRID' THEN 1 END) as pqc_hybrid_emails,
    ROUND(CAST(COUNT(CASE WHEN pqc_enabled = TRUE THEN 1 END) AS FLOAT) / COUNT(*) * 100, 2) as pqc_adoption_percentage
FROM emails;

-- Create view for PQC key statistics
CREATE VIEW IF NOT EXISTS pqc_key_statistics AS
SELECT 
    COUNT(*) as total_keys,
    COUNT(CASE WHEN is_active = TRUE THEN 1 END) as active_keys,
    COUNT(CASE WHEN is_active = FALSE THEN 1 END) as inactive_keys,
    COUNT(CASE WHEN expires_at > CURRENT_TIMESTAMP THEN 1 END) as valid_keys,
    COUNT(CASE WHEN expires_at <= CURRENT_TIMESTAMP THEN 1 END) as expired_keys,
    AVG(kyber_level) as average_kyber_level,
    MAX(rotation_count) as max_rotation_count
FROM pqc_keys;

-- Create view for PQC performance statistics
CREATE VIEW IF NOT EXISTS pqc_performance_statistics AS
SELECT 
    operation,
    COUNT(*) as total_operations,
    COUNT(CASE WHEN success = TRUE THEN 1 END) as successful_operations,
    COUNT(CASE WHEN success = FALSE THEN 1 END) as failed_operations,
    AVG(duration_ms) as avg_duration_ms,
    MAX(duration_ms) as max_duration_ms,
    MIN(duration_ms) as min_duration_ms,
    AVG(data_size) as avg_data_size,
    ROUND(CAST(COUNT(CASE WHEN success = TRUE THEN 1 END) AS FLOAT) / COUNT(*) * 100, 2) as success_rate
FROM pqc_performance_metrics
GROUP BY operation;

-- Create view for PQC audit summary
CREATE VIEW IF NOT EXISTS pqc_audit_summary AS
SELECT 
    event_type,
    severity,
    COUNT(*) as event_count,
    MIN(timestamp) as first_occurrence,
    MAX(timestamp) as last_occurrence
FROM pqc_audit_log
GROUP BY event_type, severity
ORDER BY event_count DESC;

-- Insert initial PQC configuration
INSERT OR IGNORE INTO pqc_keys (
    id, 
    key_id, 
    kyber_level, 
    public_key, 
    created_at, 
    expires_at, 
    is_active, 
    rotation_count
) VALUES (
    'initial-pqc-key',
    'initial-key-001',
    768,
    'dummy-public-key-base64', -- Will be replaced with actual key
    CURRENT_TIMESTAMP,
    datetime('now', '+30 days'),
    TRUE,
    0
);

-- Create trigger to log PQC encryption events
CREATE TRIGGER IF NOT EXISTS trigger_pqc_encryption_log
AFTER UPDATE OF pqc_enabled ON emails
FOR EACH ROW
WHEN NEW.pqc_enabled = TRUE AND OLD.pqc_enabled = FALSE
BEGIN
    INSERT INTO pqc_audit_log (
        id,
        event_type,
        description,
        severity,
        email_id,
        details,
        timestamp
    ) VALUES (
        hex(randomblob(16)),
        'EMAIL_PQC_ENABLED',
        'Email migrated to PQC encryption',
        'INFO',
        NEW.id,
        json_object(
            'old_encryption_version', OLD.encryption_version,
            'new_encryption_version', NEW.encryption_version,
            'pqc_key_id', NEW.pqc_key_id
        ),
        CURRENT_TIMESTAMP
    );
END;

-- Create trigger to log PQC key rotation events
CREATE TRIGGER IF NOT EXISTS trigger_pqc_key_rotation_log
AFTER UPDATE OF is_active ON pqc_keys
FOR EACH ROW
WHEN NEW.is_active = FALSE AND OLD.is_active = TRUE
BEGIN
    INSERT INTO pqc_audit_log (
        id,
        event_type,
        description,
        severity,
        key_id,
        details,
        timestamp
    ) VALUES (
        hex(randomblob(16)),
        'KEY_ROTATION',
        'PQC key rotated and deactivated',
        'INFO',
        OLD.id,
        json_object(
            'old_key_id', OLD.key_id,
            'rotation_count', OLD.rotation_count,
            'kyber_level', OLD.kyber_level
        ),
        CURRENT_TIMESTAMP
    );
END;

-- Create trigger to log PQC performance metrics
CREATE TRIGGER IF NOT EXISTS trigger_pqc_performance_log
AFTER INSERT ON pqc_performance_metrics
FOR EACH ROW
WHEN NEW.duration_ms > 1000 -- Log slow operations
BEGIN
    INSERT INTO pqc_audit_log (
        id,
        event_type,
        description,
        severity,
        details,
        timestamp
    ) VALUES (
        hex(randomblob(16)),
        'PERFORMANCE_WARNING',
        'Slow PQC operation detected',
        'WARN',
        json_object(
            'operation', NEW.operation,
            'duration_ms', NEW.duration_ms,
            'data_size', NEW.data_size,
            'kyber_level', NEW.kyber_level
        ),
        CURRENT_TIMESTAMP
    );
END;

-- =============================================================================
-- Migration completed successfully
-- =============================================================================
