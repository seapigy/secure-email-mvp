-- =============================================================================
-- Migration: Add E2E PQC System
-- =============================================================================
-- This migration adds the database schema for the end-to-end PQC encryption system
-- including E2E messages, Key Transparency, and Threshold HSM tables.
-- =============================================================================

-- Begin transaction
BEGIN TRANSACTION;

-- =============================================================================
-- E2E Messages Table
-- =============================================================================
-- Stores end-to-end encrypted messages with minimal server-side metadata
CREATE TABLE IF NOT EXISTS e2e_messages (
    id TEXT PRIMARY KEY,
    thread_id TEXT NOT NULL,
    sequence_number INTEGER NOT NULL,
    sender_uuid TEXT NOT NULL,
    recipient_uuid TEXT NOT NULL,
    envelope_hash TEXT NOT NULL,
    envelope_version TEXT NOT NULL DEFAULT '1.0',
    key_rotation_id TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,
    
    -- Routing metadata (minimal)
    routing_token TEXT,
    delivery_status TEXT DEFAULT 'pending',
    
    -- Encryption metadata
    kem_algorithm TEXT NOT NULL DEFAULT 'kyber768',
    dem_algorithm TEXT NOT NULL DEFAULT 'aes256gcm',
    signature_algorithm TEXT NOT NULL DEFAULT 'dilithium3',
    
    -- Audit fields
    correlation_id TEXT,
    trace_id TEXT,
    
    -- Feature flag tracking
    e2e_enabled BOOLEAN DEFAULT FALSE,
    migration_status TEXT DEFAULT 'legacy',
    
    UNIQUE(thread_id, sequence_number)
);

-- =============================================================================
-- Key Transparency Tables
-- =============================================================================
-- Stores public key bindings with Merkle tree proofs
CREATE TABLE IF NOT EXISTS kt_public_keys (
    id TEXT PRIMARY KEY,
    user_uuid TEXT NOT NULL,
    public_key TEXT NOT NULL,
    key_type TEXT NOT NULL DEFAULT 'kyber768',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,
    merkle_index INTEGER,
    merkle_proof TEXT,
    signature TEXT NOT NULL,
    
    -- Key metadata
    key_version TEXT DEFAULT '1.0',
    key_algorithm TEXT DEFAULT 'kyber768',
    key_purpose TEXT DEFAULT 'encryption',
    
    UNIQUE(user_uuid, key_type)
);

-- Key Transparency log entries (append-only)
CREATE TABLE IF NOT EXISTS kt_log_entries (
    id TEXT PRIMARY KEY,
    entry_hash TEXT NOT NULL,
    merkle_index INTEGER NOT NULL,
    merkle_root TEXT NOT NULL,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    signature TEXT NOT NULL,
    
    -- Entry metadata
    entry_type TEXT DEFAULT 'key_publish',
    user_uuid TEXT,
    key_id TEXT,
    
    UNIQUE(merkle_index)
);

-- =============================================================================
-- Threshold HSM Tables
-- =============================================================================
-- Manages threshold operations for key management
CREATE TABLE IF NOT EXISTS hsm_key_operations (
    id TEXT PRIMARY KEY,
    operation_type TEXT NOT NULL, -- 'wrap', 'unwrap', 'rotate', 'backup'
    key_id TEXT NOT NULL,
    threshold_m INTEGER NOT NULL DEFAULT 3,
    threshold_n INTEGER NOT NULL DEFAULT 5,
    operator_signatures TEXT, -- JSON array of signatures
    status TEXT DEFAULT 'pending', -- 'pending', 'in_progress', 'completed', 'failed'
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,
    
    -- Operation metadata
    operation_data TEXT, -- Encrypted operation data
    error_message TEXT,
    correlation_id TEXT,
    
    -- Audit fields
    requester_uuid TEXT,
    approver_uuids TEXT -- JSON array of approver UUIDs
);

-- HSM operator management
CREATE TABLE IF NOT EXISTS hsm_operators (
    id TEXT PRIMARY KEY,
    operator_uuid TEXT NOT NULL UNIQUE,
    operator_name TEXT NOT NULL,
    public_key TEXT NOT NULL,
    status TEXT DEFAULT 'active', -- 'active', 'inactive', 'revoked'
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_operation_at TIMESTAMP,
    
    -- Operator metadata
    role TEXT DEFAULT 'operator',
    permissions TEXT -- JSON array of permissions
);

-- =============================================================================
-- E2E Thread Management
-- =============================================================================
-- Manages conversation threads for E2E encryption
CREATE TABLE IF NOT EXISTS e2e_threads (
    id TEXT PRIMARY KEY,
    thread_key_id TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_activity_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Thread metadata
    thread_type TEXT DEFAULT 'conversation',
    participant_count INTEGER DEFAULT 2,
    
    -- Key rotation
    current_key_rotation_id TEXT NOT NULL,
    key_rotation_interval_days INTEGER DEFAULT 30,
    next_rotation_at TIMESTAMP
);

-- Thread participants
CREATE TABLE IF NOT EXISTS e2e_thread_participants (
    thread_id TEXT NOT NULL,
    user_uuid TEXT NOT NULL,
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    left_at TIMESTAMP,
    role TEXT DEFAULT 'participant', -- 'participant', 'admin', 'readonly'
    
    PRIMARY KEY (thread_id, user_uuid),
    FOREIGN KEY (thread_id) REFERENCES e2e_threads(id)
);

-- =============================================================================
-- E2E Performance Metrics
-- =============================================================================
-- Tracks performance metrics for E2E operations
CREATE TABLE IF NOT EXISTS e2e_performance_metrics (
    id TEXT PRIMARY KEY,
    operation TEXT NOT NULL, -- 'encrypt', 'decrypt', 'kt_append', 'kt_verify', 'hsm_wrap', 'hsm_unwrap'
    duration_ms INTEGER NOT NULL,
    data_size INTEGER,
    kyber_level INTEGER,
    hsm_enabled BOOLEAN DEFAULT FALSE,
    success BOOLEAN NOT NULL,
    error_message TEXT,
    context TEXT,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Performance metadata
    correlation_id TEXT,
    trace_id TEXT,
    user_uuid TEXT,
    thread_id TEXT
);

-- =============================================================================
-- E2E Audit Logs
-- =============================================================================
-- Comprehensive audit logging for E2E operations
CREATE TABLE IF NOT EXISTS e2e_audit_logs (
    id TEXT PRIMARY KEY,
    user_uuid TEXT,
    action TEXT NOT NULL,
    resource_type TEXT, -- 'message', 'key', 'thread', 'kt_entry', 'hsm_operation'
    resource_id TEXT,
    ip_address TEXT,
    user_agent TEXT,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    details TEXT, -- JSON details (no plaintext)
    
    -- E2E specific fields
    correlation_id TEXT,
    trace_id TEXT,
    e2e_enabled BOOLEAN DEFAULT FALSE,
    encryption_method TEXT,
    thread_id TEXT,
    sequence_number INTEGER,
    
    FOREIGN KEY (user_uuid) REFERENCES users(id)
);

-- =============================================================================
-- Feature Flag Management
-- =============================================================================
-- Manages feature flags for E2E system
CREATE TABLE IF NOT EXISTS e2e_feature_flags (
    id TEXT PRIMARY KEY,
    flag_name TEXT NOT NULL UNIQUE,
    flag_value BOOLEAN NOT NULL DEFAULT FALSE,
    scope TEXT NOT NULL DEFAULT 'global', -- 'global', 'org', 'user'
    scope_id TEXT, -- org_id or user_id for scoped flags
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Flag metadata
    description TEXT,
    created_by TEXT,
    updated_by TEXT
);

-- =============================================================================
-- Migration Tracking
-- =============================================================================
-- Tracks E2E migration progress
CREATE TABLE IF NOT EXISTS e2e_migration_status (
    id TEXT PRIMARY KEY,
    migration_phase TEXT NOT NULL, -- 'dual_mode', 'key_publication', 'gradual_rollout', 'background_migration'
    status TEXT NOT NULL DEFAULT 'pending', -- 'pending', 'in_progress', 'completed', 'failed', 'rolled_back'
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    
    -- Migration metadata
    total_emails INTEGER DEFAULT 0,
    migrated_emails INTEGER DEFAULT 0,
    failed_emails INTEGER DEFAULT 0,
    error_message TEXT,
    
    -- Progress tracking
    current_batch_size INTEGER DEFAULT 100,
    current_batch_start INTEGER DEFAULT 0,
    last_processed_id TEXT
);

-- =============================================================================
-- INDEXES FOR PERFORMANCE
-- =============================================================================

-- E2E Messages indexes
CREATE INDEX IF NOT EXISTS idx_e2e_messages_thread ON e2e_messages(thread_id);
CREATE INDEX IF NOT EXISTS idx_e2e_messages_recipient ON e2e_messages(recipient_uuid);
CREATE INDEX IF NOT EXISTS idx_e2e_messages_created ON e2e_messages(created_at);
CREATE INDEX IF NOT EXISTS idx_e2e_messages_status ON e2e_messages(delivery_status);
CREATE INDEX IF NOT EXISTS idx_e2e_messages_migration ON e2e_messages(migration_status);

-- Key Transparency indexes
CREATE INDEX IF NOT EXISTS idx_kt_public_keys_user ON kt_public_keys(user_uuid);
CREATE INDEX IF NOT EXISTS idx_kt_public_keys_type ON kt_public_keys(key_type);
CREATE INDEX IF NOT EXISTS idx_kt_public_keys_expires ON kt_public_keys(expires_at);
CREATE INDEX IF NOT EXISTS idx_kt_log_entries_index ON kt_log_entries(merkle_index);
CREATE INDEX IF NOT EXISTS idx_kt_log_entries_timestamp ON kt_log_entries(timestamp);

-- HSM Operations indexes
CREATE INDEX IF NOT EXISTS idx_hsm_operations_type ON hsm_key_operations(operation_type);
CREATE INDEX IF NOT EXISTS idx_hsm_operations_status ON hsm_key_operations(status);
CREATE INDEX IF NOT EXISTS idx_hsm_operations_created ON hsm_key_operations(created_at);
CREATE INDEX IF NOT EXISTS idx_hsm_operators_status ON hsm_operators(status);

-- E2E Threads indexes
CREATE INDEX IF NOT EXISTS idx_e2e_threads_activity ON e2e_threads(last_activity_at);
CREATE INDEX IF NOT EXISTS idx_e2e_threads_rotation ON e2e_threads(next_rotation_at);
CREATE INDEX IF NOT EXISTS idx_e2e_thread_participants_user ON e2e_thread_participants(user_uuid);

-- Performance metrics indexes
CREATE INDEX IF NOT EXISTS idx_e2e_performance_operation ON e2e_performance_metrics(operation);
CREATE INDEX IF NOT EXISTS idx_e2e_performance_timestamp ON e2e_performance_metrics(timestamp);
CREATE INDEX IF NOT EXISTS idx_e2e_performance_success ON e2e_performance_metrics(success);

-- Audit logs indexes
CREATE INDEX IF NOT EXISTS idx_e2e_audit_user ON e2e_audit_logs(user_uuid);
CREATE INDEX IF NOT EXISTS idx_e2e_audit_action ON e2e_audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_e2e_audit_timestamp ON e2e_audit_logs(timestamp);
CREATE INDEX IF NOT EXISTS idx_e2e_audit_correlation ON e2e_audit_logs(correlation_id);

-- Feature flags indexes
CREATE INDEX IF NOT EXISTS idx_e2e_feature_flags_scope ON e2e_feature_flags(scope, scope_id);
CREATE INDEX IF NOT EXISTS idx_e2e_feature_flags_value ON e2e_feature_flags(flag_value);

-- Migration status indexes
CREATE INDEX IF NOT EXISTS idx_e2e_migration_phase ON e2e_migration_status(migration_phase);
CREATE INDEX IF NOT EXISTS idx_e2e_migration_status ON e2e_migration_status(status);

-- =============================================================================
-- INITIAL DATA
-- =============================================================================

-- Insert default feature flags
INSERT OR IGNORE INTO e2e_feature_flags (id, flag_name, flag_value, scope, description) VALUES
('e2e_global_enabled', 'E2E_ENABLED', FALSE, 'global', 'Global E2E encryption enable/disable'),
('e2e_kt_enabled', 'KT_ENABLED', FALSE, 'global', 'Key Transparency service enable/disable'),
('e2e_hsm_enabled', 'HSM_ENABLED', FALSE, 'global', 'HSM integration enable/disable'),
('e2e_debug_enabled', 'E2E_DEBUG', FALSE, 'global', 'E2E debug logging enable/disable'),
('e2e_observability_enabled', 'E2E_OBSERVABILITY', TRUE, 'global', 'E2E observability features enable/disable');

-- Insert initial migration status
INSERT OR IGNORE INTO e2e_migration_status (id, migration_phase, status, started_at) VALUES
('initial_migration', 'dual_mode', 'completed', CURRENT_TIMESTAMP);

-- =============================================================================
-- TRIGGERS FOR AUDIT LOGGING
-- =============================================================================

-- Trigger to log E2E message creation
CREATE TRIGGER IF NOT EXISTS trigger_e2e_messages_audit_insert
AFTER INSERT ON e2e_messages
BEGIN
    INSERT INTO e2e_audit_logs (
        id, user_uuid, action, resource_type, resource_id, 
        correlation_id, trace_id, e2e_enabled, encryption_method,
        thread_id, sequence_number, details
    ) VALUES (
        'audit_' || NEW.id, NEW.sender_uuid, 'e2e_message_created', 
        'message', NEW.id, NEW.correlation_id, NEW.trace_id, 
        NEW.e2e_enabled, NEW.dem_algorithm, NEW.thread_id, 
        NEW.sequence_number, 
        json_object(
            'envelope_version', NEW.envelope_version,
            'kem_algorithm', NEW.kem_algorithm,
            'signature_algorithm', NEW.signature_algorithm,
            'delivery_status', NEW.delivery_status
        )
    );
END;

-- Trigger to log KT key publication
CREATE TRIGGER IF NOT EXISTS trigger_kt_keys_audit_insert
AFTER INSERT ON kt_public_keys
BEGIN
    INSERT INTO e2e_audit_logs (
        id, user_uuid, action, resource_type, resource_id, 
        details
    ) VALUES (
        'audit_kt_' || NEW.id, NEW.user_uuid, 'kt_key_published', 
        'key', NEW.id, 
        json_object(
            'key_type', NEW.key_type,
            'key_version', NEW.key_version,
            'key_algorithm', NEW.key_algorithm,
            'merkle_index', NEW.merkle_index
        )
    );
END;

-- Trigger to log HSM operations
CREATE TRIGGER IF NOT EXISTS trigger_hsm_operations_audit_insert
AFTER INSERT ON hsm_key_operations
BEGIN
    INSERT INTO e2e_audit_logs (
        id, user_uuid, action, resource_type, resource_id, 
        correlation_id, details
    ) VALUES (
        'audit_hsm_' || NEW.id, NEW.requester_uuid, 'hsm_operation_created', 
        'hsm_operation', NEW.id, NEW.correlation_id,
        json_object(
            'operation_type', NEW.operation_type,
            'threshold_m', NEW.threshold_m,
            'threshold_n', NEW.threshold_n,
            'status', NEW.status
        )
    );
END;

-- =============================================================================
-- VIEWS FOR MONITORING
-- =============================================================================

-- View for E2E migration progress
CREATE VIEW IF NOT EXISTS v_e2e_migration_progress AS
SELECT 
    migration_phase,
    status,
    total_emails,
    migrated_emails,
    failed_emails,
    CASE 
        WHEN total_emails > 0 THEN 
            ROUND((migrated_emails * 100.0 / total_emails), 2)
        ELSE 0 
    END as progress_percentage,
    started_at,
    completed_at,
    CASE 
        WHEN completed_at IS NOT NULL THEN 
            ROUND((julianday(completed_at) - julianday(started_at)) * 24 * 60, 2)
        ELSE NULL 
    END as duration_minutes
FROM e2e_migration_status
ORDER BY started_at DESC;

-- View for E2E performance metrics
CREATE VIEW IF NOT EXISTS v_e2e_performance_summary AS
SELECT 
    operation,
    COUNT(*) as total_operations,
    COUNT(CASE WHEN success THEN 1 END) as successful_operations,
    ROUND(AVG(duration_ms), 2) as avg_duration_ms,
    ROUND(MAX(duration_ms), 2) as max_duration_ms,
    ROUND(MIN(duration_ms), 2) as min_duration_ms,
    ROUND(
        (COUNT(CASE WHEN success THEN 1 END) * 100.0 / COUNT(*)), 2
    ) as success_rate_percentage
FROM e2e_performance_metrics
WHERE timestamp >= datetime('now', '-24 hours')
GROUP BY operation
ORDER BY total_operations DESC;

-- View for E2E feature flag status
CREATE VIEW IF NOT EXISTS v_e2e_feature_flags_status AS
SELECT 
    flag_name,
    flag_value,
    scope,
    scope_id,
    description,
    updated_at
FROM e2e_feature_flags
ORDER BY scope, flag_name;

-- =============================================================================
-- COMMIT TRANSACTION
-- =============================================================================

COMMIT;

-- =============================================================================
-- MIGRATION VERIFICATION
-- =============================================================================

-- Verify all tables were created successfully
SELECT 'E2E Messages Table' as table_name, COUNT(*) as row_count FROM e2e_messages
UNION ALL
SELECT 'KT Public Keys Table', COUNT(*) FROM kt_public_keys
UNION ALL
SELECT 'KT Log Entries Table', COUNT(*) FROM kt_log_entries
UNION ALL
SELECT 'HSM Key Operations Table', COUNT(*) FROM hsm_key_operations
UNION ALL
SELECT 'HSM Operators Table', COUNT(*) FROM hsm_operators
UNION ALL
SELECT 'E2E Threads Table', COUNT(*) FROM e2e_threads
UNION ALL
SELECT 'E2E Thread Participants Table', COUNT(*) FROM e2e_thread_participants
UNION ALL
SELECT 'E2E Performance Metrics Table', COUNT(*) FROM e2e_performance_metrics
UNION ALL
SELECT 'E2E Audit Logs Table', COUNT(*) FROM e2e_audit_logs
UNION ALL
SELECT 'E2E Feature Flags Table', COUNT(*) FROM e2e_feature_flags
UNION ALL
SELECT 'E2E Migration Status Table', COUNT(*) FROM e2e_migration_status;

-- Verify indexes were created
SELECT 'Indexes created successfully' as status, COUNT(*) as index_count
FROM sqlite_master 
WHERE type = 'index' 
AND name LIKE 'idx_e2e_%';

-- =============================================================================
-- MIGRATION COMPLETE
-- =============================================================================
-- The E2E PQC system database schema has been successfully created.
-- All tables, indexes, triggers, views, and initial data are in place.
-- The system is ready for Sprint 1 implementation.
