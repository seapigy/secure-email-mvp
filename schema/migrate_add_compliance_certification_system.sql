-- Migration: Add automated compliance & retention certification system (Micro-Iteration 4.30)
-- This migration adds support for enterprise compliance frameworks, certifications, and audit trails

-- Compliance frameworks table
CREATE TABLE IF NOT EXISTS compliance_frameworks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    framework_name TEXT NOT NULL UNIQUE, -- "GDPR", "HIPAA", "SOX", "Custom"
    framework_version TEXT NOT NULL DEFAULT "1.0",
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Compliance rules table
CREATE TABLE IF NOT EXISTS compliance_rules (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    framework_id INTEGER NOT NULL,
    rule_code TEXT NOT NULL, -- e.g., "GDPR_5_1_E", "HIPAA_164_312", "SOX_302"
    rule_name TEXT NOT NULL,
    rule_description TEXT NOT NULL,
    retention_period_days INTEGER,
    archival_required BOOLEAN DEFAULT FALSE,
    encryption_required BOOLEAN DEFAULT FALSE,
    audit_logging_required BOOLEAN DEFAULT TRUE,
    auto_enforcement_enabled BOOLEAN DEFAULT TRUE,
    severity_level TEXT DEFAULT "medium", -- "low", "medium", "high", "critical"
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (framework_id) REFERENCES compliance_frameworks(id),
    UNIQUE(framework_id, rule_code)
);

-- Policy-to-compliance mapping table
CREATE TABLE IF NOT EXISTS policy_compliance_mapping (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    retention_policy_id INTEGER NOT NULL,
    compliance_rule_id INTEGER NOT NULL,
    mapping_type TEXT NOT NULL, -- "direct", "partial", "exemption"
    mapping_notes TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (retention_policy_id) REFERENCES retention_policies(id),
    FOREIGN KEY (compliance_rule_id) REFERENCES compliance_rules(id),
    UNIQUE(retention_policy_id, compliance_rule_id)
);

-- Compliance exemptions table
CREATE TABLE IF NOT EXISTS compliance_exemptions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    framework_id INTEGER NOT NULL,
    exemption_type TEXT NOT NULL, -- "user", "domain", "email_type", "legal_hold"
    exemption_key TEXT NOT NULL, -- user_id, domain, or specific identifier
    exemption_reason TEXT NOT NULL,
    exemption_duration_days INTEGER, -- NULL for permanent
    approved_by TEXT NOT NULL,
    approved_at TIMESTAMP NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (framework_id) REFERENCES compliance_frameworks(id)
);

-- Compliance certifications table
CREATE TABLE IF NOT EXISTS compliance_certifications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    certification_id TEXT NOT NULL UNIQUE, -- UUID for external reference
    framework_id INTEGER NOT NULL,
    certification_type TEXT NOT NULL, -- "monthly", "quarterly", "annual", "ad_hoc"
    certification_period_start TIMESTAMP NOT NULL,
    certification_period_end TIMESTAMP NOT NULL,
    generated_at TIMESTAMP NOT NULL,
    generated_by TEXT NOT NULL,
    
    -- Certification status
    status TEXT DEFAULT "draft", -- "draft", "pending_review", "approved", "rejected"
    approved_at TIMESTAMP,
    approved_by TEXT,
    approval_notes TEXT,
    
    -- Compliance metrics
    total_emails_analyzed INTEGER DEFAULT 0,
    compliant_emails_count INTEGER DEFAULT 0,
    non_compliant_emails_count INTEGER DEFAULT 0,
    violations_count INTEGER DEFAULT 0,
    exemptions_count INTEGER DEFAULT 0,
    compliance_score REAL DEFAULT 0.0, -- 0.0 to 1.0
    
    -- Evidence and audit
    evidence_summary TEXT, -- JSON summary of evidence
    audit_trail_hash TEXT, -- Cryptographic hash of audit trail
    digital_signature TEXT, -- Digital signature for audit validity
    
    -- Export and storage
    report_file_path TEXT, -- Path to generated report file
    report_file_hash TEXT, -- Hash of report file for integrity
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (framework_id) REFERENCES compliance_frameworks(id)
);

-- Compliance violations table
CREATE TABLE IF NOT EXISTS compliance_violations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    violation_id TEXT NOT NULL UNIQUE, -- UUID for external reference
    framework_id INTEGER NOT NULL,
    compliance_rule_id INTEGER NOT NULL,
    email_id TEXT,
    user_id TEXT,
    domain TEXT,
    
    -- Violation details
    violation_type TEXT NOT NULL, -- "retention_exceeded", "archival_missing", "encryption_missing", "audit_missing"
    violation_severity TEXT NOT NULL, -- "low", "medium", "high", "critical"
    violation_description TEXT NOT NULL,
    detected_at TIMESTAMP NOT NULL,
    
    -- Current state
    status TEXT DEFAULT "open", -- "open", "acknowledged", "resolved", "exempted"
    acknowledged_at TIMESTAMP,
    acknowledged_by TEXT,
    resolved_at TIMESTAMP,
    resolved_by TEXT,
    resolution_notes TEXT,
    
    -- Auto-resolution
    auto_resolved BOOLEAN DEFAULT FALSE,
    auto_resolution_action TEXT, -- "archived", "deleted", "exempted"
    
    -- Related data
    retention_policy_id INTEGER,
    affected_emails_count INTEGER DEFAULT 1,
    days_over_limit INTEGER DEFAULT 0,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (framework_id) REFERENCES compliance_frameworks(id),
    FOREIGN KEY (compliance_rule_id) REFERENCES compliance_rules(id),
    FOREIGN KEY (retention_policy_id) REFERENCES retention_policies(id)
);

-- Compliance audit logs table
CREATE TABLE IF NOT EXISTS compliance_audit_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    audit_id TEXT NOT NULL UNIQUE, -- UUID for external reference
    framework_id INTEGER NOT NULL,
    compliance_rule_id INTEGER,
    certification_id INTEGER,
    violation_id INTEGER,
    
    -- Audit event details
    event_type TEXT NOT NULL, -- "policy_evaluation", "archival_operation", "violation_detected", "certification_generated"
    event_timestamp TIMESTAMP NOT NULL,
    event_source TEXT NOT NULL, -- "retention_engine", "compliance_checker", "admin_action"
    
    -- Event data
    event_data TEXT, -- JSON data about the event
    affected_emails TEXT, -- JSON array of affected email IDs
    affected_users TEXT, -- JSON array of affected user IDs
    
    -- Evidence and integrity
    evidence_hash TEXT, -- Hash of evidence data
    previous_state_hash TEXT, -- Hash of previous state
    new_state_hash TEXT, -- Hash of new state
    
    -- Metadata
    user_agent TEXT,
    ip_address TEXT,
    session_id TEXT,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (framework_id) REFERENCES compliance_frameworks(id),
    FOREIGN KEY (compliance_rule_id) REFERENCES compliance_rules(id),
    FOREIGN KEY (certification_id) REFERENCES compliance_certifications(id),
    FOREIGN KEY (violation_id) REFERENCES compliance_violations(id)
);

-- Enterprise organizations table
CREATE TABLE IF NOT EXISTS enterprise_organizations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    org_id TEXT NOT NULL UNIQUE, -- UUID for external reference
    org_name TEXT NOT NULL,
    org_domain TEXT NOT NULL UNIQUE,
    org_type TEXT NOT NULL, -- "healthcare", "financial", "legal", "general"
    
    -- Enterprise features
    enterprise_enabled BOOLEAN DEFAULT FALSE,
    compliance_enabled BOOLEAN DEFAULT FALSE,
    auto_enforcement_enabled BOOLEAN DEFAULT TRUE,
    
    -- Compliance configuration
    primary_framework_id INTEGER,
    secondary_frameworks TEXT, -- JSON array of framework IDs
    compliance_contact_email TEXT,
    compliance_contact_name TEXT,
    
    -- Billing and licensing (future)
    subscription_tier TEXT DEFAULT "basic",
    license_expires_at TIMESTAMP,
    
    -- Status
    status TEXT DEFAULT "active", -- "active", "suspended", "cancelled"
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (primary_framework_id) REFERENCES compliance_frameworks(id)
);

-- User enterprise mapping table
CREATE TABLE IF NOT EXISTS user_enterprise_mapping (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id TEXT NOT NULL,
    org_id TEXT NOT NULL,
    role TEXT NOT NULL, -- "admin", "user", "compliance_officer", "auditor"
    permissions TEXT, -- JSON array of permissions
    is_active BOOLEAN DEFAULT TRUE,
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (org_id) REFERENCES enterprise_organizations(org_id),
    UNIQUE(user_id, org_id)
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_compliance_rules_framework ON compliance_rules(framework_id);
CREATE INDEX IF NOT EXISTS idx_compliance_rules_active ON compliance_rules(is_active);
CREATE INDEX IF NOT EXISTS idx_policy_compliance_mapping_policy ON policy_compliance_mapping(retention_policy_id);
CREATE INDEX IF NOT EXISTS idx_policy_compliance_mapping_rule ON policy_compliance_mapping(compliance_rule_id);
CREATE INDEX IF NOT EXISTS idx_compliance_exemptions_framework ON compliance_exemptions(framework_id);
CREATE INDEX IF NOT EXISTS idx_compliance_exemptions_type_key ON compliance_exemptions(exemption_type, exemption_key);
CREATE INDEX IF NOT EXISTS idx_compliance_certifications_framework ON compliance_certifications(framework_id);
CREATE INDEX IF NOT EXISTS idx_compliance_certifications_period ON compliance_certifications(certification_period_start, certification_period_end);
CREATE INDEX IF NOT EXISTS idx_compliance_certifications_status ON compliance_certifications(status);
CREATE INDEX IF NOT EXISTS idx_compliance_violations_framework ON compliance_violations(framework_id);
CREATE INDEX IF NOT EXISTS idx_compliance_violations_status ON compliance_violations(status);
CREATE INDEX IF NOT EXISTS idx_compliance_violations_detected_at ON compliance_violations(detected_at);
CREATE INDEX IF NOT EXISTS idx_compliance_audit_logs_framework ON compliance_audit_logs(framework_id);
CREATE INDEX IF NOT EXISTS idx_compliance_audit_logs_event_timestamp ON compliance_audit_logs(event_timestamp);
CREATE INDEX IF NOT EXISTS idx_compliance_audit_logs_event_type ON compliance_audit_logs(event_type);
CREATE INDEX IF NOT EXISTS idx_enterprise_organizations_domain ON enterprise_organizations(org_domain);
CREATE INDEX IF NOT EXISTS idx_enterprise_organizations_enabled ON enterprise_organizations(enterprise_enabled, compliance_enabled);
CREATE INDEX IF NOT EXISTS idx_user_enterprise_mapping_user ON user_enterprise_mapping(user_id);
CREATE INDEX IF NOT EXISTS idx_user_enterprise_mapping_org ON user_enterprise_mapping(org_id);

-- Create views for analytics
CREATE VIEW IF NOT EXISTS compliance_status_summary AS
SELECT
    cf.framework_name,
    cf.framework_version,
    COUNT(cr.id) as total_rules,
    COUNT(CASE WHEN cr.is_active = 1 THEN 1 END) as active_rules,
    COUNT(cv.id) as total_violations,
    COUNT(CASE WHEN cv.status = 'open' THEN 1 END) as open_violations,
    COUNT(cc.id) as total_certifications,
    COUNT(CASE WHEN cc.status = 'approved' THEN 1 END) as approved_certifications,
    AVG(cc.compliance_score) as avg_compliance_score
FROM compliance_frameworks cf
LEFT JOIN compliance_rules cr ON cf.id = cr.framework_id
LEFT JOIN compliance_violations cv ON cf.id = cv.framework_id
LEFT JOIN compliance_certifications cc ON cf.id = cc.framework_id
WHERE cf.is_active = 1
GROUP BY cf.id, cf.framework_name, cf.framework_version;

CREATE VIEW IF NOT EXISTS enterprise_compliance_overview AS
SELECT
    eo.org_name,
    eo.org_domain,
    cf.framework_name,
    COUNT(cv.id) as violations_last_30_days,
    COUNT(cc.id) as certifications_last_90_days,
    AVG(cc.compliance_score) as avg_compliance_score,
    eo.compliance_contact_email,
    eo.status
FROM enterprise_organizations eo
LEFT JOIN compliance_frameworks cf ON eo.primary_framework_id = cf.id
LEFT JOIN compliance_violations cv ON cf.id = cv.framework_id 
    AND cv.detected_at >= datetime('now', '-30 days')
LEFT JOIN compliance_certifications cc ON cf.id = cc.framework_id 
    AND cc.generated_at >= datetime('now', '-90 days')
WHERE eo.enterprise_enabled = 1 AND eo.compliance_enabled = 1
GROUP BY eo.id, eo.org_name, eo.org_domain, cf.framework_name, eo.compliance_contact_email, eo.status;

CREATE VIEW IF NOT EXISTS compliance_violation_trends AS
SELECT
    DATE(cv.detected_at) as violation_date,
    cf.framework_name,
    cv.violation_type,
    cv.violation_severity,
    COUNT(*) as violation_count,
    COUNT(CASE WHEN cv.status = 'resolved' THEN 1 END) as resolved_count
FROM compliance_violations cv
JOIN compliance_frameworks cf ON cv.framework_id = cf.id
WHERE cv.detected_at >= datetime('now', '-90 days')
GROUP BY DATE(cv.detected_at), cf.framework_name, cv.violation_type, cv.violation_severity
ORDER BY violation_date DESC;

-- Create triggers for automatic timestamp updates
CREATE TRIGGER IF NOT EXISTS update_compliance_frameworks_updated_at
    AFTER UPDATE ON compliance_frameworks
    FOR EACH ROW
BEGIN
    UPDATE compliance_frameworks SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS update_compliance_rules_updated_at
    AFTER UPDATE ON compliance_rules
    FOR EACH ROW
BEGIN
    UPDATE compliance_rules SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS update_policy_compliance_mapping_updated_at
    AFTER UPDATE ON policy_compliance_mapping
    FOR EACH ROW
BEGIN
    UPDATE policy_compliance_mapping SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS update_compliance_exemptions_updated_at
    AFTER UPDATE ON compliance_exemptions
    FOR EACH ROW
BEGIN
    UPDATE compliance_exemptions SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS update_compliance_certifications_updated_at
    AFTER UPDATE ON compliance_certifications
    FOR EACH ROW
BEGIN
    UPDATE compliance_certifications SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS update_compliance_violations_updated_at
    AFTER UPDATE ON compliance_violations
    FOR EACH ROW
BEGIN
    UPDATE compliance_violations SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS update_enterprise_organizations_updated_at
    AFTER UPDATE ON enterprise_organizations
    FOR EACH ROW
BEGIN
    UPDATE enterprise_organizations SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS update_user_enterprise_mapping_updated_at
    AFTER UPDATE ON user_enterprise_mapping
    FOR EACH ROW
BEGIN
    UPDATE user_enterprise_mapping SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- Insert default compliance frameworks
INSERT OR IGNORE INTO compliance_frameworks (framework_name, framework_version, description) VALUES
('GDPR', '1.0', 'General Data Protection Regulation (EU) 2016/679'),
('HIPAA', '1.0', 'Health Insurance Portability and Accountability Act'),
('SOX', '1.0', 'Sarbanes-Oxley Act of 2002'),
('CCPA', '1.0', 'California Consumer Privacy Act'),
('LGPD', '1.0', 'Lei Geral de Proteção de Dados (Brazil)'),
('Custom', '1.0', 'Custom compliance framework');

-- Insert default GDPR rules
INSERT OR IGNORE INTO compliance_rules (framework_id, rule_code, rule_name, rule_description, retention_period_days, archival_required, encryption_required, severity_level) VALUES
(1, 'GDPR_5_1_E', 'Data Minimization', 'Personal data shall be kept in a form which permits identification of data subjects for no longer than is necessary', 2555, TRUE, TRUE, 'high'),
(1, 'GDPR_17_1', 'Right to Erasure', 'Data subject has the right to obtain erasure of personal data without undue delay', 30, FALSE, TRUE, 'critical'),
(1, 'GDPR_25_1', 'Data Protection by Design', 'Appropriate technical and organizational measures shall be implemented', NULL, TRUE, TRUE, 'high'),
(1, 'GDPR_30_1', 'Records of Processing Activities', 'Each controller shall maintain a record of processing activities', 2555, TRUE, FALSE, 'medium'),
(1, 'GDPR_32_1', 'Security of Processing', 'Appropriate security measures including encryption and pseudonymization', NULL, TRUE, TRUE, 'critical'),
(1, 'GDPR_33_1', 'Breach Notification', 'Notification of personal data breach to supervisory authority within 72 hours', 30, TRUE, TRUE, 'critical');

-- Insert default HIPAA rules
INSERT OR IGNORE INTO compliance_rules (framework_id, rule_code, rule_name, rule_description, retention_period_days, archival_required, encryption_required, severity_level) VALUES
(2, 'HIPAA_164_312', 'Access Control', 'Implement technical policies and procedures for electronic information systems', 2555, TRUE, TRUE, 'critical'),
(2, 'HIPAA_164_314', 'Information Access Management', 'Implement policies and procedures for authorizing access to electronic protected health information', 2555, TRUE, TRUE, 'high'),
(2, 'HIPAA_164_316', 'Audit Controls', 'Implement hardware, software, and/or procedural mechanisms to record and examine access', 2555, TRUE, FALSE, 'high'),
(2, 'HIPAA_164_320', 'Data Backup and Recovery', 'Establish and implement procedures to create and maintain retrievable exact copies of electronic protected health information', 2555, TRUE, TRUE, 'critical'),
(2, 'HIPAA_164_322', 'Person or Entity Authentication', 'Implement procedures to verify that a person or entity seeking access to electronic protected health information is the one claimed', NULL, TRUE, TRUE, 'critical');

-- Insert default SOX rules
INSERT OR IGNORE INTO compliance_rules (framework_id, rule_code, rule_name, rule_description, retention_period_days, archival_required, encryption_required, severity_level) VALUES
(3, 'SOX_302', 'Corporate Responsibility for Financial Reports', 'CEO and CFO must certify financial reports and internal controls', 2555, TRUE, TRUE, 'critical'),
(3, 'SOX_404', 'Management Assessment of Internal Controls', 'Annual evaluation of internal control structure and procedures', 2555, TRUE, TRUE, 'high'),
(3, 'SOX_409', 'Real-Time Disclosure', 'Disclosure of material changes in financial condition or operations', 30, TRUE, TRUE, 'critical'),
(3, 'SOX_802', 'Criminal Penalties for Altering Documents', 'Criminal penalties for knowingly altering, destroying, or falsifying records', 2555, TRUE, TRUE, 'critical');

-- Create cleanup triggers for old data
CREATE TRIGGER IF NOT EXISTS cleanup_old_compliance_audit_logs
    AFTER INSERT ON compliance_audit_logs
    FOR EACH ROW
BEGIN
    DELETE FROM compliance_audit_logs
    WHERE event_timestamp < datetime('now', '-2 years');
END;

CREATE TRIGGER IF NOT EXISTS cleanup_old_compliance_violations
    AFTER INSERT ON compliance_violations
    FOR EACH ROW
BEGIN
    DELETE FROM compliance_violations
    WHERE detected_at < datetime('now', '-1 year') AND status IN ('resolved', 'exempted');
END;

CREATE TRIGGER IF NOT EXISTS cleanup_old_compliance_certifications
    AFTER INSERT ON compliance_certifications
    FOR EACH ROW
BEGIN
    DELETE FROM compliance_certifications
    WHERE generated_at < datetime('now', '-3 years');
END;






