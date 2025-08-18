package email

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ComplianceFramework represents a compliance framework (GDPR, HIPAA, SOX, etc.)
type ComplianceFramework struct {
	ID               int64     `json:"id"`
	FrameworkName    string    `json:"framework_name"`
	FrameworkVersion string    `json:"framework_version"`
	Description      string    `json:"description"`
	IsActive         bool      `json:"is_active"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// ComplianceRule represents a specific rule within a compliance framework
type ComplianceRule struct {
	ID                     int64     `json:"id"`
	FrameworkID            int64     `json:"framework_id"`
	RuleCode               string    `json:"rule_code"`
	RuleName               string    `json:"rule_name"`
	RuleDescription        string    `json:"rule_description"`
	RetentionPeriodDays    *int      `json:"retention_period_days,omitempty"`
	ArchivalRequired       bool      `json:"archival_required"`
	EncryptionRequired     bool      `json:"encryption_required"`
	AuditLoggingRequired   bool      `json:"audit_logging_required"`
	AutoEnforcementEnabled bool      `json:"auto_enforcement_enabled"`
	SeverityLevel          string    `json:"severity_level"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

// PolicyComplianceMapping represents the mapping between retention policies and compliance rules
type PolicyComplianceMapping struct {
	ID                int64     `json:"id"`
	RetentionPolicyID int64     `json:"retention_policy_id"`
	ComplianceRuleID  int64     `json:"compliance_rule_id"`
	MappingType       string    `json:"mapping_type"`
	MappingNotes      string    `json:"mapping_notes"`
	IsActive          bool      `json:"is_active"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// ComplianceExemption represents an exemption from compliance rules
type ComplianceExemption struct {
	ID                    int64     `json:"id"`
	FrameworkID           int64     `json:"framework_id"`
	ExemptionType         string    `json:"exemption_type"`
	ExemptionKey          string    `json:"exemption_key"`
	ExemptionReason       string    `json:"exemption_reason"`
	ExemptionDurationDays *int      `json:"exemption_duration_days,omitempty"`
	ApprovedBy            string    `json:"approved_by"`
	ApprovedAt            time.Time `json:"approved_at"`
	IsActive              bool      `json:"is_active"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

// ComplianceCertification represents a compliance certification report
type ComplianceCertification struct {
	ID                       int64      `json:"id"`
	CertificationID          string     `json:"certification_id"`
	FrameworkID              int64      `json:"framework_id"`
	CertificationType        string     `json:"certification_type"`
	CertificationPeriodStart time.Time  `json:"certification_period_start"`
	CertificationPeriodEnd   time.Time  `json:"certification_period_end"`
	GeneratedAt              time.Time  `json:"generated_at"`
	GeneratedBy              string     `json:"generated_by"`
	Status                   string     `json:"status"`
	ApprovedAt               *time.Time `json:"approved_at,omitempty"`
	ApprovedBy               *string    `json:"approved_by,omitempty"`
	ApprovalNotes            *string    `json:"approval_notes,omitempty"`
	TotalEmailsAnalyzed      int        `json:"total_emails_analyzed"`
	CompliantEmailsCount     int        `json:"compliant_emails_count"`
	NonCompliantEmailsCount  int        `json:"non_compliant_emails_count"`
	ViolationsCount          int        `json:"violations_count"`
	ExemptionsCount          int        `json:"exemptions_count"`
	ComplianceScore          float64    `json:"compliance_score"`
	EvidenceSummary          string     `json:"evidence_summary"`
	AuditTrailHash           string     `json:"audit_trail_hash"`
	DigitalSignature         string     `json:"digital_signature"`
	ReportFilePath           *string    `json:"report_file_path,omitempty"`
	ReportFileHash           *string    `json:"report_file_hash,omitempty"`
	CreatedAt                time.Time  `json:"created_at"`
	UpdatedAt                time.Time  `json:"updated_at"`
}

// ComplianceViolation represents a compliance violation
type ComplianceViolation struct {
	ID                   int64      `json:"id"`
	ViolationID          string     `json:"violation_id"`
	FrameworkID          int64      `json:"framework_id"`
	ComplianceRuleID     int64      `json:"compliance_rule_id"`
	EmailID              *string    `json:"email_id,omitempty"`
	UserID               *string    `json:"user_id,omitempty"`
	Domain               *string    `json:"domain,omitempty"`
	ViolationType        string     `json:"violation_type"`
	ViolationSeverity    string     `json:"violation_severity"`
	ViolationDescription string     `json:"violation_description"`
	DetectedAt           time.Time  `json:"detected_at"`
	Status               string     `json:"status"`
	AcknowledgedAt       *time.Time `json:"acknowledged_at,omitempty"`
	AcknowledgedBy       *string    `json:"acknowledged_by,omitempty"`
	ResolvedAt           *time.Time `json:"resolved_at,omitempty"`
	ResolvedBy           *string    `json:"resolved_by,omitempty"`
	ResolutionNotes      *string    `json:"resolution_notes,omitempty"`
	AutoResolved         bool       `json:"auto_resolved"`
	AutoResolutionAction *string    `json:"auto_resolution_action,omitempty"`
	RetentionPolicyID    *int64     `json:"retention_policy_id,omitempty"`
	AffectedEmailsCount  int        `json:"affected_emails_count"`
	DaysOverLimit        int        `json:"days_over_limit"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

// ComplianceAuditLog represents an audit log entry for compliance events
type ComplianceAuditLog struct {
	ID                int64     `json:"id"`
	AuditID           string    `json:"audit_id"`
	FrameworkID       int64     `json:"framework_id"`
	ComplianceRuleID  *int64    `json:"compliance_rule_id,omitempty"`
	CertificationID   *int64    `json:"certification_id,omitempty"`
	ViolationID       *int64    `json:"violation_id,omitempty"`
	EventType         string    `json:"event_type"`
	EventTimestamp    time.Time `json:"event_timestamp"`
	EventSource       string    `json:"event_source"`
	EventData         string    `json:"event_data"`
	AffectedEmails    string    `json:"affected_emails"`
	AffectedUsers     string    `json:"affected_users"`
	EvidenceHash      string    `json:"evidence_hash"`
	PreviousStateHash *string   `json:"previous_state_hash,omitempty"`
	NewStateHash      *string   `json:"new_state_hash,omitempty"`
	UserAgent         *string   `json:"user_agent,omitempty"`
	IPAddress         *string   `json:"ip_address,omitempty"`
	SessionID         *string   `json:"session_id,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
}

// EnterpriseOrganization represents an enterprise organization
type EnterpriseOrganization struct {
	ID                     int64      `json:"id"`
	OrgID                  string     `json:"org_id"`
	OrgName                string     `json:"org_name"`
	OrgDomain              string     `json:"org_domain"`
	OrgType                string     `json:"org_type"`
	EnterpriseEnabled      bool       `json:"enterprise_enabled"`
	ComplianceEnabled      bool       `json:"compliance_enabled"`
	AutoEnforcementEnabled bool       `json:"auto_enforcement_enabled"`
	PrimaryFrameworkID     *int64     `json:"primary_framework_id,omitempty"`
	SecondaryFrameworks    *string    `json:"secondary_frameworks,omitempty"`
	ComplianceContactEmail *string    `json:"compliance_contact_email,omitempty"`
	ComplianceContactName  *string    `json:"compliance_contact_name,omitempty"`
	SubscriptionTier       string     `json:"subscription_tier"`
	LicenseExpiresAt       *time.Time `json:"license_expires_at,omitempty"`
	Status                 string     `json:"status"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
}

// UserEnterpriseMapping represents the mapping between users and enterprise organizations
type UserEnterpriseMapping struct {
	ID          int64     `json:"id"`
	UserID      string    `json:"user_id"`
	OrgID       string    `json:"org_id"`
	Role        string    `json:"role"`
	Permissions *string   `json:"permissions,omitempty"`
	IsActive    bool      `json:"is_active"`
	JoinedAt    time.Time `json:"joined_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ComplianceService provides comprehensive compliance management functionality
type ComplianceService struct {
	db *sql.DB
}

// NewComplianceService creates a new compliance service
func NewComplianceService(db *sql.DB) *ComplianceService {
	return &ComplianceService{
		db: db,
	}
}

// GetComplianceFrameworks retrieves all compliance frameworks
func (cs *ComplianceService) GetComplianceFrameworks(ctx context.Context) ([]ComplianceFramework, error) {
	query := `
		SELECT id, framework_name, framework_version, description, is_active, created_at, updated_at
		FROM compliance_frameworks
		WHERE is_active = 1
		ORDER BY framework_name
	`

	rows, err := cs.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query compliance frameworks: %w", err)
	}
	defer rows.Close()

	var frameworks []ComplianceFramework
	for rows.Next() {
		var f ComplianceFramework
		err := rows.Scan(
			&f.ID, &f.FrameworkName, &f.FrameworkVersion, &f.Description,
			&f.IsActive, &f.CreatedAt, &f.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan compliance framework: %w", err)
		}
		frameworks = append(frameworks, f)
	}

	return frameworks, nil
}

// GetComplianceRules retrieves compliance rules for a framework
func (cs *ComplianceService) GetComplianceRules(ctx context.Context, frameworkID int64) ([]ComplianceRule, error) {
	query := `
		SELECT id, framework_id, rule_code, rule_name, rule_description, retention_period_days,
		       archival_required, encryption_required, audit_logging_required, auto_enforcement_enabled,
		       severity_level, created_at, updated_at
		FROM compliance_rules
		WHERE framework_id = ? AND is_active = 1
		ORDER BY rule_code
	`

	rows, err := cs.db.QueryContext(ctx, query, frameworkID)
	if err != nil {
		return nil, fmt.Errorf("failed to query compliance rules: %w", err)
	}
	defer rows.Close()

	var rules []ComplianceRule
	for rows.Next() {
		var r ComplianceRule
		err := rows.Scan(
			&r.ID, &r.FrameworkID, &r.RuleCode, &r.RuleName, &r.RuleDescription,
			&r.RetentionPeriodDays, &r.ArchivalRequired, &r.EncryptionRequired,
			&r.AuditLoggingRequired, &r.AutoEnforcementEnabled, &r.SeverityLevel,
			&r.CreatedAt, &r.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan compliance rule: %w", err)
		}
		rules = append(rules, r)
	}

	return rules, nil
}

// CreateEnterpriseOrganization creates a new enterprise organization
func (cs *ComplianceService) CreateEnterpriseOrganization(ctx context.Context, org *EnterpriseOrganization) error {
	org.OrgID = uuid.New().String()

	query := `
		INSERT INTO enterprise_organizations (
			org_id, org_name, org_domain, org_type, enterprise_enabled, compliance_enabled,
			auto_enforcement_enabled, primary_framework_id, secondary_frameworks,
			compliance_contact_email, compliance_contact_name, subscription_tier, status
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := cs.db.ExecContext(ctx, query,
		org.OrgID, org.OrgName, org.OrgDomain, org.OrgType, org.EnterpriseEnabled,
		org.ComplianceEnabled, org.AutoEnforcementEnabled, org.PrimaryFrameworkID,
		org.SecondaryFrameworks, org.ComplianceContactEmail, org.ComplianceContactName,
		org.SubscriptionTier, org.Status,
	)

	if err != nil {
		return fmt.Errorf("failed to create enterprise organization: %w", err)
	}

	return nil
}

// GetEnterpriseOrganization retrieves an enterprise organization by domain
func (cs *ComplianceService) GetEnterpriseOrganization(ctx context.Context, domain string) (*EnterpriseOrganization, error) {
	query := `
		SELECT id, org_id, org_name, org_domain, org_type, enterprise_enabled, compliance_enabled,
		       auto_enforcement_enabled, primary_framework_id, secondary_frameworks,
		       compliance_contact_email, compliance_contact_name, subscription_tier,
		       license_expires_at, status, created_at, updated_at
		FROM enterprise_organizations
		WHERE org_domain = ? AND status = 'active'
	`

	var org EnterpriseOrganization
	err := cs.db.QueryRowContext(ctx, query, domain).Scan(
		&org.ID, &org.OrgID, &org.OrgName, &org.OrgDomain, &org.OrgType,
		&org.EnterpriseEnabled, &org.ComplianceEnabled, &org.AutoEnforcementEnabled,
		&org.PrimaryFrameworkID, &org.SecondaryFrameworks, &org.ComplianceContactEmail,
		&org.ComplianceContactName, &org.SubscriptionTier, &org.LicenseExpiresAt,
		&org.Status, &org.CreatedAt, &org.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get enterprise organization: %w", err)
	}

	return &org, nil
}

// AddUserToEnterprise adds a user to an enterprise organization
func (cs *ComplianceService) AddUserToEnterprise(ctx context.Context, userID, orgID, role string, permissions []string) error {
	permissionsJSON, err := json.Marshal(permissions)
	if err != nil {
		return fmt.Errorf("failed to marshal permissions: %w", err)
	}

	query := `
		INSERT OR REPLACE INTO user_enterprise_mapping (
			user_id, org_id, role, permissions, is_active
		) VALUES (?, ?, ?, ?, 1)
	`

	_, err = cs.db.ExecContext(ctx, query, userID, orgID, role, string(permissionsJSON))
	if err != nil {
		return fmt.Errorf("failed to add user to enterprise: %w", err)
	}

	return nil
}

// GetUserEnterpriseRole retrieves the user's role in an enterprise organization
func (cs *ComplianceService) GetUserEnterpriseRole(ctx context.Context, userID, orgID string) (*UserEnterpriseMapping, error) {
	query := `
		SELECT id, user_id, org_id, role, permissions, is_active, joined_at, updated_at
		FROM user_enterprise_mapping
		WHERE user_id = ? AND org_id = ? AND is_active = 1
	`

	var mapping UserEnterpriseMapping
	err := cs.db.QueryRowContext(ctx, query, userID, orgID).Scan(
		&mapping.ID, &mapping.UserID, &mapping.OrgID, &mapping.Role,
		&mapping.Permissions, &mapping.IsActive, &mapping.JoinedAt, &mapping.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user enterprise role: %w", err)
	}

	return &mapping, nil
}

// CreateComplianceViolation creates a new compliance violation
func (cs *ComplianceService) CreateComplianceViolation(ctx context.Context, violation *ComplianceViolation) error {
	violation.ViolationID = uuid.New().String()
	violation.DetectedAt = time.Now()

	query := `
		INSERT INTO compliance_violations (
			violation_id, framework_id, compliance_rule_id, email_id, user_id, domain,
			violation_type, violation_severity, violation_description, detected_at, status,
			retention_policy_id, affected_emails_count, days_over_limit
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := cs.db.ExecContext(ctx, query,
		violation.ViolationID, violation.FrameworkID, violation.ComplianceRuleID,
		violation.EmailID, violation.UserID, violation.Domain, violation.ViolationType,
		violation.ViolationSeverity, violation.ViolationDescription, violation.DetectedAt,
		violation.Status, violation.RetentionPolicyID, violation.AffectedEmailsCount,
		violation.DaysOverLimit,
	)

	if err != nil {
		return fmt.Errorf("failed to create compliance violation: %w", err)
	}

	// Log the violation in audit log
	err = cs.logComplianceEvent(ctx, "violation_detected", violation.FrameworkID, &violation.ComplianceRuleID, nil, nil, nil, nil, nil)
	if err != nil {
		log.Printf("Failed to log compliance violation event: %v", err)
	}

	return nil
}

// GetComplianceViolations retrieves compliance violations with filtering
func (cs *ComplianceService) GetComplianceViolations(ctx context.Context, frameworkID *int64, status *string, severity *string, limit, offset int) ([]ComplianceViolation, error) {
	query := `
		SELECT id, violation_id, framework_id, compliance_rule_id, email_id, user_id, domain,
		       violation_type, violation_severity, violation_description, detected_at, status,
		       acknowledged_at, acknowledged_by, resolved_at, resolved_by, resolution_notes,
		       auto_resolved, auto_resolution_action, retention_policy_id, affected_emails_count,
		       days_over_limit, created_at, updated_at
		FROM compliance_violations
		WHERE 1=1
	`

	var args []interface{}

	if frameworkID != nil {
		query += " AND framework_id = ?"
		args = append(args, *frameworkID)
	}

	if status != nil {
		query += " AND status = ?"
		args = append(args, *status)
	}

	if severity != nil {
		query += " AND violation_severity = ?"
		args = append(args, *severity)
	}

	query += " ORDER BY detected_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := cs.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query compliance violations: %w", err)
	}
	defer rows.Close()

	var violations []ComplianceViolation
	for rows.Next() {
		var v ComplianceViolation
		err := rows.Scan(
			&v.ID, &v.ViolationID, &v.FrameworkID, &v.ComplianceRuleID, &v.EmailID,
			&v.UserID, &v.Domain, &v.ViolationType, &v.ViolationSeverity, &v.ViolationDescription,
			&v.DetectedAt, &v.Status, &v.AcknowledgedAt, &v.AcknowledgedBy, &v.ResolvedAt,
			&v.ResolvedBy, &v.ResolutionNotes, &v.AutoResolved, &v.AutoResolutionAction,
			&v.RetentionPolicyID, &v.AffectedEmailsCount, &v.DaysOverLimit, &v.CreatedAt, &v.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan compliance violation: %w", err)
		}
		violations = append(violations, v)
	}

	return violations, nil
}

// AcknowledgeViolation acknowledges a compliance violation
func (cs *ComplianceService) AcknowledgeViolation(ctx context.Context, violationID string, acknowledgedBy string, notes *string) error {
	query := `
		UPDATE compliance_violations
		SET status = 'acknowledged', acknowledged_at = CURRENT_TIMESTAMP, acknowledged_by = ?, resolution_notes = ?
		WHERE violation_id = ?
	`

	_, err := cs.db.ExecContext(ctx, query, acknowledgedBy, notes, violationID)
	if err != nil {
		return fmt.Errorf("failed to acknowledge violation: %w", err)
	}

	return nil
}

// ResolveViolation resolves a compliance violation
func (cs *ComplianceService) ResolveViolation(ctx context.Context, violationID string, resolvedBy string, notes *string) error {
	query := `
		UPDATE compliance_violations
		SET status = 'resolved', resolved_at = CURRENT_TIMESTAMP, resolved_by = ?, resolution_notes = ?
		WHERE violation_id = ?
	`

	_, err := cs.db.ExecContext(ctx, query, resolvedBy, notes, violationID)
	if err != nil {
		return fmt.Errorf("failed to resolve violation: %w", err)
	}

	return nil
}

// GenerateComplianceCertification generates a compliance certification report
func (cs *ComplianceService) GenerateComplianceCertification(ctx context.Context, frameworkID int64, certType string, periodStart, periodEnd time.Time, generatedBy string) (*ComplianceCertification, error) {
	// Calculate compliance metrics
	metrics, err := cs.calculateComplianceMetrics(ctx, frameworkID, periodStart, periodEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate compliance metrics: %w", err)
	}

	// Generate evidence summary
	evidenceSummary, err := cs.generateEvidenceSummary(ctx, frameworkID, periodStart, periodEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to generate evidence summary: %w", err)
	}

	// Calculate audit trail hash
	auditTrailHash, err := cs.calculateAuditTrailHash(ctx, frameworkID, periodStart, periodEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate audit trail hash: %w", err)
	}

	// Generate digital signature (simplified for now)
	digitalSignature := cs.generateDigitalSignature(evidenceSummary, auditTrailHash)

	certification := &ComplianceCertification{
		CertificationID:          uuid.New().String(),
		FrameworkID:              frameworkID,
		CertificationType:        certType,
		CertificationPeriodStart: periodStart,
		CertificationPeriodEnd:   periodEnd,
		GeneratedAt:              time.Now(),
		GeneratedBy:              generatedBy,
		Status:                   "draft",
		TotalEmailsAnalyzed:      metrics.TotalEmails,
		CompliantEmailsCount:     metrics.CompliantEmails,
		NonCompliantEmailsCount:  metrics.NonCompliantEmails,
		ViolationsCount:          metrics.ViolationsCount,
		ExemptionsCount:          metrics.ExemptionsCount,
		ComplianceScore:          metrics.ComplianceScore,
		EvidenceSummary:          evidenceSummary,
		AuditTrailHash:           auditTrailHash,
		DigitalSignature:         digitalSignature,
	}

	query := `
		INSERT INTO compliance_certifications (
			certification_id, framework_id, certification_type, certification_period_start,
			certification_period_end, generated_at, generated_by, status, total_emails_analyzed,
			compliant_emails_count, non_compliant_emails_count, violations_count, exemptions_count,
			compliance_score, evidence_summary, audit_trail_hash, digital_signature
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = cs.db.ExecContext(ctx, query,
		certification.CertificationID, certification.FrameworkID, certification.CertificationType,
		certification.CertificationPeriodStart, certification.CertificationPeriodEnd,
		certification.GeneratedAt, certification.GeneratedBy, certification.Status,
		certification.TotalEmailsAnalyzed, certification.CompliantEmailsCount,
		certification.NonCompliantEmailsCount, certification.ViolationsCount,
		certification.ExemptionsCount, certification.ComplianceScore, certification.EvidenceSummary,
		certification.AuditTrailHash, certification.DigitalSignature,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create compliance certification: %w", err)
	}

	// Log the certification generation
	err = cs.logComplianceEvent(ctx, "certification_generated", frameworkID, nil, &certification.ID, nil, nil, nil, nil)
	if err != nil {
		log.Printf("Failed to log certification generation event: %v", err)
	}

	return certification, nil
}

// GetComplianceCertifications retrieves compliance certifications
func (cs *ComplianceService) GetComplianceCertifications(ctx context.Context, frameworkID *int64, status *string, limit, offset int) ([]ComplianceCertification, error) {
	query := `
		SELECT id, certification_id, framework_id, certification_type, certification_period_start,
		       certification_period_end, generated_at, generated_by, status, approved_at, approved_by,
		       approval_notes, total_emails_analyzed, compliant_emails_count, non_compliant_emails_count,
		       violations_count, exemptions_count, compliance_score, evidence_summary, audit_trail_hash,
		       digital_signature, report_file_path, report_file_hash, created_at, updated_at
		FROM compliance_certifications
		WHERE 1=1
	`

	var args []interface{}

	if frameworkID != nil {
		query += " AND framework_id = ?"
		args = append(args, *frameworkID)
	}

	if status != nil {
		query += " AND status = ?"
		args = append(args, *status)
	}

	query += " ORDER BY generated_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := cs.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query compliance certifications: %w", err)
	}
	defer rows.Close()

	var certifications []ComplianceCertification
	for rows.Next() {
		var c ComplianceCertification
		err := rows.Scan(
			&c.ID, &c.CertificationID, &c.FrameworkID, &c.CertificationType,
			&c.CertificationPeriodStart, &c.CertificationPeriodEnd, &c.GeneratedAt,
			&c.GeneratedBy, &c.Status, &c.ApprovedAt, &c.ApprovedBy, &c.ApprovalNotes,
			&c.TotalEmailsAnalyzed, &c.CompliantEmailsCount, &c.NonCompliantEmailsCount,
			&c.ViolationsCount, &c.ExemptionsCount, &c.ComplianceScore, &c.EvidenceSummary,
			&c.AuditTrailHash, &c.DigitalSignature, &c.ReportFilePath, &c.ReportFileHash,
			&c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan compliance certification: %w", err)
		}
		certifications = append(certifications, c)
	}

	return certifications, nil
}

// ApproveCertification approves a compliance certification
func (cs *ComplianceService) ApproveCertification(ctx context.Context, certificationID string, approvedBy string, notes *string) error {
	query := `
		UPDATE compliance_certifications
		SET status = 'approved', approved_at = CURRENT_TIMESTAMP, approved_by = ?, approval_notes = ?
		WHERE certification_id = ?
	`

	_, err := cs.db.ExecContext(ctx, query, approvedBy, notes, certificationID)
	if err != nil {
		return fmt.Errorf("failed to approve certification: %w", err)
	}

	return nil
}

// logComplianceEvent logs a compliance audit event
func (cs *ComplianceService) logComplianceEvent(ctx context.Context, eventType string, frameworkID int64, ruleID, certID, violationID *int64, eventData, affectedEmails, affectedUsers *string) error {
	auditLog := &ComplianceAuditLog{
		AuditID:          uuid.New().String(),
		FrameworkID:      frameworkID,
		ComplianceRuleID: ruleID,
		CertificationID:  certID,
		ViolationID:      violationID,
		EventType:        eventType,
		EventTimestamp:   time.Now(),
		EventSource:      "compliance_service",
		EventData:        "",
		AffectedEmails:   "[]",
		AffectedUsers:    "[]",
		EvidenceHash:     "",
	}

	if eventData != nil {
		auditLog.EventData = *eventData
	}
	if affectedEmails != nil {
		auditLog.AffectedEmails = *affectedEmails
	}
	if affectedUsers != nil {
		auditLog.AffectedUsers = *affectedUsers
	}

	// Calculate evidence hash
	evidenceData := fmt.Sprintf("%s|%d|%s|%s", eventType, frameworkID, auditLog.EventData, auditLog.EventTimestamp.Format(time.RFC3339))
	auditLog.EvidenceHash = fmt.Sprintf("%x", sha256.Sum256([]byte(evidenceData)))

	query := `
		INSERT INTO compliance_audit_logs (
			audit_id, framework_id, compliance_rule_id, certification_id, violation_id,
			event_type, event_timestamp, event_source, event_data, affected_emails,
			affected_users, evidence_hash
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := cs.db.ExecContext(ctx, query,
		auditLog.AuditID, auditLog.FrameworkID, auditLog.ComplianceRuleID,
		auditLog.CertificationID, auditLog.ViolationID, auditLog.EventType,
		auditLog.EventTimestamp, auditLog.EventSource, auditLog.EventData,
		auditLog.AffectedEmails, auditLog.AffectedUsers, auditLog.EvidenceHash,
	)

	if err != nil {
		return fmt.Errorf("failed to log compliance event: %w", err)
	}

	return nil
}

// calculateComplianceMetrics calculates compliance metrics for a given period
func (cs *ComplianceService) calculateComplianceMetrics(ctx context.Context, frameworkID int64, periodStart, periodEnd time.Time) (*ComplianceMetrics, error) {
	// This is a simplified implementation - in a real system, this would be much more complex
	metrics := &ComplianceMetrics{
		TotalEmails:        0,
		CompliantEmails:    0,
		NonCompliantEmails: 0,
		ViolationsCount:    0,
		ExemptionsCount:    0,
		ComplianceScore:    0.0,
	}

	// Count total emails in period
	err := cs.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM emails 
		WHERE created_at BETWEEN ? AND ?
	`, periodStart, periodEnd).Scan(&metrics.TotalEmails)
	if err != nil {
		return nil, fmt.Errorf("failed to count total emails: %w", err)
	}

	// Count violations in period
	err = cs.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM compliance_violations 
		WHERE framework_id = ? AND detected_at BETWEEN ? AND ?
	`, frameworkID, periodStart, periodEnd).Scan(&metrics.ViolationsCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count violations: %w", err)
	}

	// Count exemptions
	err = cs.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM compliance_exemptions 
		WHERE framework_id = ? AND is_active = 1
	`, frameworkID).Scan(&metrics.ExemptionsCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count exemptions: %w", err)
	}

	// Calculate compliance score
	if metrics.TotalEmails > 0 {
		metrics.NonCompliantEmails = metrics.ViolationsCount
		metrics.CompliantEmails = metrics.TotalEmails - metrics.NonCompliantEmails
		metrics.ComplianceScore = float64(metrics.CompliantEmails) / float64(metrics.TotalEmails)
	}

	return metrics, nil
}

// generateEvidenceSummary generates a summary of evidence for the certification period
func (cs *ComplianceService) generateEvidenceSummary(ctx context.Context, frameworkID int64, periodStart, periodEnd time.Time) (string, error) {
	// This is a simplified implementation
	summary := map[string]interface{}{
		"period_start": periodStart.Format(time.RFC3339),
		"period_end":   periodEnd.Format(time.RFC3339),
		"framework_id": frameworkID,
		"generated_at": time.Now().Format(time.RFC3339),
	}

	summaryJSON, err := json.Marshal(summary)
	if err != nil {
		return "", fmt.Errorf("failed to marshal evidence summary: %w", err)
	}

	return string(summaryJSON), nil
}

// calculateAuditTrailHash calculates a hash of the audit trail for the period
func (cs *ComplianceService) calculateAuditTrailHash(ctx context.Context, frameworkID int64, periodStart, periodEnd time.Time) (string, error) {
	// This is a simplified implementation
	auditData := fmt.Sprintf("%d|%s|%s", frameworkID, periodStart.Format(time.RFC3339), periodEnd.Format(time.RFC3339))
	hash := sha256.Sum256([]byte(auditData))
	return fmt.Sprintf("%x", hash), nil
}

// generateDigitalSignature generates a digital signature for the certification
func (cs *ComplianceService) generateDigitalSignature(evidenceSummary, auditTrailHash string) string {
	// This is a simplified implementation - in production, use proper cryptographic signing
	data := fmt.Sprintf("%s|%s|%s", evidenceSummary, auditTrailHash, time.Now().Format(time.RFC3339))
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// ComplianceMetrics represents compliance metrics for a period
type ComplianceMetrics struct {
	TotalEmails        int     `json:"total_emails"`
	CompliantEmails    int     `json:"compliant_emails"`
	NonCompliantEmails int     `json:"non_compliant_emails"`
	ViolationsCount    int     `json:"violations_count"`
	ExemptionsCount    int     `json:"exemptions_count"`
	ComplianceScore    float64 `json:"compliance_score"`
}

// UserComplianceStatus represents a user's compliance status
type UserComplianceStatus struct {
	UserID                    string                    `json:"user_id"`
	Domain                    string                    `json:"domain"`
	IsEnterpriseUser          bool                      `json:"is_enterprise_user"`
	OrganizationName          *string                   `json:"organization_name,omitempty"`
	ActiveFrameworks          []UserComplianceFramework `json:"active_frameworks"`
	ApplicablePolicies        []UserRetentionPolicy     `json:"applicable_policies"`
	RecentViolations          []UserComplianceViolation `json:"recent_violations,omitempty"`
	ComplianceScore           float64                   `json:"compliance_score"`
	LastPolicyEvaluation      *time.Time                `json:"last_policy_evaluation,omitempty"`
	NextArchivalDate          *time.Time                `json:"next_archival_date,omitempty"`
	TransparencySettings      UserTransparencySettings  `json:"transparency_settings"`
	GeneratedAt               time.Time                 `json:"generated_at"`
}

// UserComplianceFramework represents a compliance framework that applies to a user
type UserComplianceFramework struct {
	FrameworkID          int64  `json:"framework_id"`
	FrameworkName        string `json:"framework_name"`
	FrameworkVersion     string `json:"framework_version"`
	Description          string `json:"description"`
	ComplianceStatus     string `json:"compliance_status"`
	ActiveRulesCount     int    `json:"active_rules_count"`
	ViolationsCount      int    `json:"violations_count"`
	LastCertificationAt  *time.Time `json:"last_certification_at,omitempty"`
}

// UserRetentionPolicy represents a retention policy that applies to a user
type UserRetentionPolicy struct {
	PolicyID             int64     `json:"policy_id"`
	PolicyName           string    `json:"policy_name"`
	PolicyType           string    `json:"policy_type"`
	RetentionPeriodDays  int       `json:"retention_period_days"`
	ArchivalEnabled      bool      `json:"archival_enabled"`
	ArchivalLocation     *string   `json:"archival_location,omitempty"`
	ComplianceRules      []string  `json:"compliance_rules"`
	LastEvaluatedAt      *time.Time `json:"last_evaluated_at,omitempty"`
	NextEvaluationAt     *time.Time `json:"next_evaluation_at,omitempty"`
	HumanReadableSummary string    `json:"human_readable_summary"`
}

// UserComplianceViolation represents a compliance violation for a user
type UserComplianceViolation struct {
	ViolationID          string     `json:"violation_id"`
	FrameworkName        string     `json:"framework_name"`
	RuleName             string     `json:"rule_name"`
	ViolationType        string     `json:"violation_type"`
	ViolationSeverity    string     `json:"violation_severity"`
	ViolationDescription string     `json:"violation_description"`
	DetectedAt           time.Time  `json:"detected_at"`
	Status               string     `json:"status"`
	AffectedEmailsCount  int        `json:"affected_emails_count"`
	DaysOverLimit        int        `json:"days_over_limit"`
}

// UserTransparencySettings represents transparency settings for a user
type UserTransparencySettings struct {
	ShowRetentionRules    bool `json:"show_retention_rules"`
	ShowComplianceFrameworks bool `json:"show_compliance_frameworks"`
	ShowViolations        bool `json:"show_violations"`
	CacheTTLMinutes       int  `json:"cache_ttl_minutes"`
}

// GetUserComplianceStatus retrieves the compliance status for a specific user
func (cs *ComplianceService) GetUserComplianceStatus(ctx context.Context, userID string) (*UserComplianceStatus, error) {
	// Get user's domain and enterprise status
	var domain string
	var isEnterpriseUser bool
	var orgName *string
	
	err := cs.db.QueryRowContext(ctx, `
		SELECT domain FROM users WHERE user_id = ?
	`, userID).Scan(&domain)
	if err != nil {
		return nil, fmt.Errorf("failed to get user domain: %w", err)
	}

	// Check if user is part of an enterprise organization
	err = cs.db.QueryRowContext(ctx, `
		SELECT 1, eo.org_name 
		FROM user_enterprise_mapping uem
		JOIN enterprise_organizations eo ON uem.org_id = eo.org_id
		WHERE uem.user_id = ? AND uem.is_active = 1 AND eo.enterprise_enabled = 1
	`, userID).Scan(&isEnterpriseUser, &orgName)
	if err == sql.ErrNoRows {
		isEnterpriseUser = false
		orgName = nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to check enterprise status: %w", err)
	}

	// Get transparency settings
	transparencySettings := cs.getUserTransparencySettings()

	// Get active frameworks
	frameworks, err := cs.getUserComplianceFrameworks(ctx, userID, domain, isEnterpriseUser)
	if err != nil {
		return nil, fmt.Errorf("failed to get user compliance frameworks: %w", err)
	}

	// Get applicable policies
	policies, err := cs.getUserRetentionPolicies(ctx, userID, domain)
	if err != nil {
		return nil, fmt.Errorf("failed to get user retention policies: %w", err)
	}

	// Get recent violations if enabled
	var violations []UserComplianceViolation
	if transparencySettings.ShowViolations {
		violations, err = cs.getUserComplianceViolations(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to get user violations: %w", err)
		}
	}

	// Calculate compliance score
	complianceScore := cs.calculateUserComplianceScore(frameworks, violations)

	// Get last policy evaluation and next archival date
	lastPolicyEvaluation, nextArchivalDate := cs.getUserPolicyTimeline(ctx, userID)

	status := &UserComplianceStatus{
		UserID:               userID,
		Domain:               domain,
		IsEnterpriseUser:     isEnterpriseUser,
		OrganizationName:     orgName,
		ActiveFrameworks:     frameworks,
		ApplicablePolicies:   policies,
		RecentViolations:     violations,
		ComplianceScore:      complianceScore,
		LastPolicyEvaluation: lastPolicyEvaluation,
		NextArchivalDate:     nextArchivalDate,
		TransparencySettings: transparencySettings,
		GeneratedAt:          time.Now(),
	}

	return status, nil
}

// GetUserCompliancePolicies retrieves human-readable retention policies for a user
func (cs *ComplianceService) GetUserCompliancePolicies(ctx context.Context, userID string) ([]UserRetentionPolicy, error) {
	// Get user's domain
	var domain string
	err := cs.db.QueryRowContext(ctx, `
		SELECT domain FROM users WHERE user_id = ?
	`, userID).Scan(&domain)
	if err != nil {
		return nil, fmt.Errorf("failed to get user domain: %w", err)
	}

	return cs.getUserRetentionPolicies(ctx, userID, domain)
}

// getUserComplianceFrameworks retrieves compliance frameworks that apply to a user
func (cs *ComplianceService) getUserComplianceFrameworks(ctx context.Context, userID, domain string, isEnterpriseUser bool) ([]UserComplianceFramework, error) {
	var frameworks []UserComplianceFramework

	// If not enterprise user, return empty list
	if !isEnterpriseUser {
		return frameworks, nil
	}

	query := `
		SELECT DISTINCT
			cf.id,
			cf.framework_name,
			cf.framework_version,
			cf.description,
			(SELECT COUNT(*) FROM compliance_rules cr WHERE cr.framework_id = cf.id AND cr.auto_enforcement_enabled = 1) as active_rules_count,
			(SELECT COUNT(*) FROM compliance_violations cv WHERE cv.framework_id = cf.id AND cv.user_id = ? AND cv.detected_at >= datetime('now', '-90 days')) as violations_count,
			(SELECT MAX(cc.generated_at) FROM compliance_certifications cc WHERE cc.framework_id = cf.id AND cc.status = 'approved') as last_certification_at
		FROM compliance_frameworks cf
		JOIN enterprise_organizations eo ON eo.primary_framework_id = cf.id OR eo.secondary_frameworks LIKE '%' || cf.framework_name || '%'
		JOIN user_enterprise_mapping uem ON uem.org_id = eo.org_id
		WHERE cf.is_active = 1 
		AND uem.user_id = ? 
		AND uem.is_active = 1
		AND eo.enterprise_enabled = 1
	`

	rows, err := cs.db.QueryContext(ctx, query, userID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user compliance frameworks: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var framework UserComplianceFramework
		var lastCertificationAt *time.Time
		
		err := rows.Scan(
			&framework.FrameworkID,
			&framework.FrameworkName,
			&framework.FrameworkVersion,
			&framework.Description,
			&framework.ActiveRulesCount,
			&framework.ViolationsCount,
			&lastCertificationAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan compliance framework: %w", err)
		}

		framework.LastCertificationAt = lastCertificationAt
		framework.ComplianceStatus = cs.determineFrameworkComplianceStatus(framework.ViolationsCount)

		frameworks = append(frameworks, framework)
	}

	return frameworks, nil
}

// getUserRetentionPolicies retrieves retention policies that apply to a user
func (cs *ComplianceService) getUserRetentionPolicies(ctx context.Context, userID, domain string) ([]UserRetentionPolicy, error) {
	var policies []UserRetentionPolicy

	query := `
		SELECT DISTINCT
			rp.id,
			rp.policy_name,
			rp.policy_type,
			rp.retention_period_days,
			rp.archival_enabled,
			rp.archival_location,
			rp.last_evaluated_at,
			rp.next_evaluation_at
		FROM retention_policies rp
		WHERE rp.is_active = 1
		AND (
			rp.policy_scope = 'global'
			OR (rp.policy_scope = 'domain' AND rp.domain_pattern = ?)
			OR (rp.policy_scope = 'user' AND rp.user_id = ?)
		)
		ORDER BY rp.priority DESC, rp.created_at DESC
	`

	rows, err := cs.db.QueryContext(ctx, query, domain, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user retention policies: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var policy UserRetentionPolicy
		var archivalLocation *string
		var lastEvaluatedAt, nextEvaluationAt *time.Time

		err := rows.Scan(
			&policy.PolicyID,
			&policy.PolicyName,
			&policy.PolicyType,
			&policy.RetentionPeriodDays,
			&policy.ArchivalEnabled,
			&archivalLocation,
			&lastEvaluatedAt,
			&nextEvaluationAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan retention policy: %w", err)
		}

		policy.ArchivalLocation = archivalLocation
		policy.LastEvaluatedAt = lastEvaluatedAt
		policy.NextEvaluationAt = nextEvaluationAt

		// Get compliance rules for this policy
		policy.ComplianceRules, err = cs.getPolicyComplianceRules(ctx, policy.PolicyID)
		if err != nil {
			return nil, fmt.Errorf("failed to get policy compliance rules: %w", err)
		}

		// Generate human-readable summary
		policy.HumanReadableSummary = cs.generatePolicySummary(policy)

		policies = append(policies, policy)
	}

	return policies, nil
}

// getUserComplianceViolations retrieves recent compliance violations for a user
func (cs *ComplianceService) getUserComplianceViolations(ctx context.Context, userID string) ([]UserComplianceViolation, error) {
	var violations []UserComplianceViolation

	query := `
		SELECT 
			cv.violation_id,
			cf.framework_name,
			cr.rule_name,
			cv.violation_type,
			cv.violation_severity,
			cv.violation_description,
			cv.detected_at,
			cv.status,
			cv.affected_emails_count,
			cv.days_over_limit
		FROM compliance_violations cv
		JOIN compliance_frameworks cf ON cv.framework_id = cf.id
		JOIN compliance_rules cr ON cv.compliance_rule_id = cr.id
		WHERE cv.user_id = ?
		AND cv.detected_at >= datetime('now', '-90 days')
		ORDER BY cv.detected_at DESC
		LIMIT 10
	`

	rows, err := cs.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user violations: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var violation UserComplianceViolation
		
		err := rows.Scan(
			&violation.ViolationID,
			&violation.FrameworkName,
			&violation.RuleName,
			&violation.ViolationType,
			&violation.ViolationSeverity,
			&violation.ViolationDescription,
			&violation.DetectedAt,
			&violation.Status,
			&violation.AffectedEmailsCount,
			&violation.DaysOverLimit,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan violation: %w", err)
		}

		violations = append(violations, violation)
	}

	return violations, nil
}

// getPolicyComplianceRules retrieves compliance rules for a specific policy
func (cs *ComplianceService) getPolicyComplianceRules(ctx context.Context, policyID int64) ([]string, error) {
	var rules []string

	query := `
		SELECT cr.rule_name
		FROM compliance_rules cr
		JOIN policy_compliance_mapping pcm ON cr.id = pcm.compliance_rule_id
		WHERE pcm.retention_policy_id = ?
		AND pcm.is_active = 1
		AND cr.auto_enforcement_enabled = 1
	`

	rows, err := cs.db.QueryContext(ctx, query, policyID)
	if err != nil {
		return nil, fmt.Errorf("failed to query policy compliance rules: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var ruleName string
		err := rows.Scan(&ruleName)
		if err != nil {
			return nil, fmt.Errorf("failed to scan rule name: %w", err)
		}
		rules = append(rules, ruleName)
	}

	return rules, nil
}

// generatePolicySummary generates a human-readable summary of a retention policy
func (cs *ComplianceService) generatePolicySummary(policy UserRetentionPolicy) string {
	summary := fmt.Sprintf("Emails will be retained for %d days", policy.RetentionPeriodDays)
	
	if policy.ArchivalEnabled {
		summary += " and then archived"
		if policy.ArchivalLocation != nil {
			summary += fmt.Sprintf(" to %s", *policy.ArchivalLocation)
		}
	} else {
		summary += " and then deleted"
	}

	if len(policy.ComplianceRules) > 0 {
		summary += fmt.Sprintf(". This policy helps comply with: %s", strings.Join(policy.ComplianceRules, ", "))
	}

	summary += "."

	return summary
}

// calculateUserComplianceScore calculates a compliance score for a user
func (cs *ComplianceService) calculateUserComplianceScore(frameworks []UserComplianceFramework, violations []UserComplianceViolation) float64 {
	if len(frameworks) == 0 {
		return 1.0 // No frameworks means no compliance requirements
	}

	totalViolations := 0
	for _, violation := range violations {
		if violation.Status != "resolved" {
			totalViolations++
		}
	}

	if totalViolations == 0 {
		return 1.0
	}

	// Simple scoring: 1.0 - (violations / frameworks)
	score := 1.0 - (float64(totalViolations) / float64(len(frameworks)))
	if score < 0 {
		score = 0
	}

	return score
}

// determineFrameworkComplianceStatus determines the compliance status based on violation count
func (cs *ComplianceService) determineFrameworkComplianceStatus(violationsCount int) string {
	switch {
	case violationsCount == 0:
		return "compliant"
	case violationsCount <= 3:
		return "minor_issues"
	case violationsCount <= 10:
		return "moderate_issues"
	default:
		return "non_compliant"
	}
}

// getUserPolicyTimeline retrieves the last policy evaluation and next archival date for a user
func (cs *ComplianceService) getUserPolicyTimeline(ctx context.Context, userID string) (*time.Time, *time.Time) {
	var lastPolicyEvaluation, nextArchivalDate *time.Time

	// Get last policy evaluation
	err := cs.db.QueryRowContext(ctx, `
		SELECT MAX(evaluated_at) 
		FROM policy_evaluation_logs 
		WHERE user_id = ?
	`, userID).Scan(&lastPolicyEvaluation)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Failed to get last policy evaluation: %v", err)
	}

	// Get next archival date (simplified - in reality this would be more complex)
	err = cs.db.QueryRowContext(ctx, `
		SELECT MIN(created_at + (retention_period_days || ' days'))
		FROM emails e
		JOIN retention_policies rp ON (
			rp.policy_scope = 'global' 
			OR (rp.policy_scope = 'domain' AND rp.domain_pattern = (SELECT domain FROM users WHERE user_id = ?))
			OR (rp.policy_scope = 'user' AND rp.user_id = ?)
		)
		WHERE e.user_id = ? 
		AND rp.is_active = 1 
		AND rp.archival_enabled = 1
		AND e.created_at + (rp.retention_period_days || ' days') > datetime('now')
	`, userID, userID, userID).Scan(&nextArchivalDate)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Failed to get next archival date: %v", err)
	}

	return lastPolicyEvaluation, nextArchivalDate
}

// getUserTransparencySettings returns the current transparency settings
func (cs *ComplianceService) getUserTransparencySettings() UserTransparencySettings {
	// In a real implementation, these would come from environment variables or database
	return UserTransparencySettings{
		ShowRetentionRules:     true,
		ShowComplianceFrameworks: true,
		ShowViolations:         false, // Default to false for privacy
		CacheTTLMinutes:        15,
	}
}

// LogUserComplianceLookup logs when a user looks up their compliance information
func (cs *ComplianceService) LogUserComplianceLookup(ctx context.Context, userID, orgID, dataType string) error {
	eventData := map[string]interface{}{
		"user_id":   userID,
		"org_id":    orgID,
		"data_type": dataType,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	eventDataJSON, err := json.Marshal(eventData)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	// Log to compliance audit log
	_, err = cs.db.ExecContext(ctx, `
		INSERT INTO compliance_audit_logs (
			audit_id, framework_id, event_type, event_timestamp, event_source, 
			event_data, affected_users, evidence_hash, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		uuid.New().String(),
		0, // No specific framework
		"user_compliance_lookup",
		time.Now(),
		"user_api",
		string(eventDataJSON),
		userID,
		"", // No evidence hash for lookups
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to log user compliance lookup: %w", err)
	}

	return nil
}
