package security

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"secure-email-mvp/pkg/models"
	"secure-email-mvp/pkg/securelinks"
)

// SecurityRepository defines the interface for security data operations
type SecurityRepository interface {
	// Security policy operations
	CreateSecurityPolicy(policy *models.SecurityPolicy) error
	GetSecurityPolicy(linkID string) (*models.SecurityPolicy, error)
	UpdateSecurityPolicy(policy *models.SecurityPolicy) error
	
	// Security policy template operations
	GetSecurityPolicyTemplate(templateID string) (*models.SecurityPolicyTemplate, error)
	GetSecurityPolicyTemplates() ([]models.SecurityPolicyTemplate, error)
	
	// Compliance audit operations
	CreateComplianceAuditLog(audit *models.ComplianceAuditLog) error
	
	// Secure link operations
	GetSecureLink(linkID string) (*securelinks.SecureLink, error)
	UpdateSecureLink(link *securelinks.SecureLink) error
}

// SQLiteSecurityRepository implements SecurityRepository using SQLite
type SQLiteSecurityRepository struct {
	db *sql.DB
}

// NewSQLiteSecurityRepository creates a new SQLite-based security repository
func NewSQLiteSecurityRepository(db *sql.DB) *SQLiteSecurityRepository {
	return &SQLiteSecurityRepository{db: db}
}

// CreateSecurityPolicy creates a new security policy
func (r *SQLiteSecurityRepository) CreateSecurityPolicy(policy *models.SecurityPolicy) error {
	query := `
		INSERT INTO security_policies (
			policy_id, link_id, reply_id, email_id, dlp_enabled, watermark_enabled,
			download_disabled, forwarding_disabled, auto_revoke_after_reply,
			max_views, expires_at, expires_after_views, notify_on_expiry,
			notify_on_revoke, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	_, err := r.db.Exec(query,
		policy.PolicyID, policy.LinkID, policy.ReplyID, policy.EmailID,
		policy.DLPEnabled, policy.WatermarkEnabled, policy.DownloadDisabled,
		policy.ForwardingDisabled, policy.AutoRevokeAfterReply, policy.MaxViews,
		policy.ExpiresAt, policy.ExpiresAfterViews, policy.NotifyOnExpiry,
		policy.NotifyOnRevoke, policy.CreatedAt, policy.UpdatedAt,
	)
	return err
}

// GetSecurityPolicy retrieves a security policy by link ID
func (r *SQLiteSecurityRepository) GetSecurityPolicy(linkID string) (*models.SecurityPolicy, error) {
	query := `
		SELECT policy_id, link_id, reply_id, email_id, dlp_enabled, watermark_enabled,
		       download_disabled, forwarding_disabled, auto_revoke_after_reply,
		       max_views, expires_at, expires_after_views, notify_on_expiry,
		       notify_on_revoke, created_at, updated_at
		FROM security_policies WHERE link_id = ?
	`
	
	var policy models.SecurityPolicy
	err := r.db.QueryRow(query, linkID).Scan(
		&policy.PolicyID, &policy.LinkID, &policy.ReplyID, &policy.EmailID,
		&policy.DLPEnabled, &policy.WatermarkEnabled, &policy.DownloadDisabled,
		&policy.ForwardingDisabled, &policy.AutoRevokeAfterReply, &policy.MaxViews,
		&policy.ExpiresAt, &policy.ExpiresAfterViews, &policy.NotifyOnExpiry,
		&policy.NotifyOnRevoke, &policy.CreatedAt, &policy.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &policy, nil
}

// UpdateSecurityPolicy updates an existing security policy
func (r *SQLiteSecurityRepository) UpdateSecurityPolicy(policy *models.SecurityPolicy) error {
	query := `
		UPDATE security_policies SET
			reply_id = ?, email_id = ?, dlp_enabled = ?, watermark_enabled = ?,
			download_disabled = ?, forwarding_disabled = ?, auto_revoke_after_reply = ?,
			max_views = ?, expires_at = ?, expires_after_views = ?, notify_on_expiry = ?,
			notify_on_revoke = ?, updated_at = ?
		WHERE policy_id = ?
	`
	
	_, err := r.db.Exec(query,
		policy.ReplyID, policy.EmailID, policy.DLPEnabled, policy.WatermarkEnabled,
		policy.DownloadDisabled, policy.ForwardingDisabled, policy.AutoRevokeAfterReply,
		policy.MaxViews, policy.ExpiresAt, policy.ExpiresAfterViews, policy.NotifyOnExpiry,
		policy.NotifyOnRevoke, time.Now(), policy.PolicyID,
	)
	return err
}

// GetSecurityPolicyTemplate retrieves a security policy template by ID
func (r *SQLiteSecurityRepository) GetSecurityPolicyTemplate(templateID string) (*models.SecurityPolicyTemplate, error) {
	query := `
		SELECT template_id, template_name, template_description, dlp_enabled, watermark_enabled,
		       download_disabled, forwarding_disabled, auto_revoke_after_reply,
		       max_views, default_expiry_hours, notify_on_expiry, notify_on_revoke,
		       created_at, updated_at
		FROM security_policy_templates WHERE template_id = ?
	`
	
	var template models.SecurityPolicyTemplate
	err := r.db.QueryRow(query, templateID).Scan(
		&template.TemplateID, &template.TemplateName, &template.TemplateDescription,
		&template.DLPEnabled, &template.WatermarkEnabled, &template.DownloadDisabled,
		&template.ForwardingDisabled, &template.AutoRevokeAfterReply, &template.MaxViews,
		&template.DefaultExpiryHours, &template.NotifyOnExpiry, &template.NotifyOnRevoke,
		&template.CreatedAt, &template.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &template, nil
}

// GetSecurityPolicyTemplates retrieves all security policy templates
func (r *SQLiteSecurityRepository) GetSecurityPolicyTemplates() ([]models.SecurityPolicyTemplate, error) {
	query := `
		SELECT template_id, template_name, template_description, dlp_enabled, watermark_enabled,
		       download_disabled, forwarding_disabled, auto_revoke_after_reply,
		       max_views, default_expiry_hours, notify_on_expiry, notify_on_revoke,
		       created_at, updated_at
		FROM security_policy_templates ORDER BY template_name
	`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var templates []models.SecurityPolicyTemplate
	for rows.Next() {
		var template models.SecurityPolicyTemplate
		err := rows.Scan(
			&template.TemplateID, &template.TemplateName, &template.TemplateDescription,
			&template.DLPEnabled, &template.WatermarkEnabled, &template.DownloadDisabled,
			&template.ForwardingDisabled, &template.AutoRevokeAfterReply, &template.MaxViews,
			&template.DefaultExpiryHours, &template.NotifyOnExpiry, &template.NotifyOnRevoke,
			&template.CreatedAt, &template.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		templates = append(templates, template)
	}
	
	return templates, nil
}

// CreateComplianceAuditLog creates a new compliance audit log entry
func (r *SQLiteSecurityRepository) CreateComplianceAuditLog(audit *models.ComplianceAuditLog) error {
	query := `
		INSERT INTO compliance_audit_log (
			audit_id, event_type, link_id, reply_id, attachment_id, policy_id, rule_id,
			user_id, ip_address, user_agent, severity, compliance_category,
			retention_required, event_details, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	eventDetails, err := audit.GetEventDetails()
	if err != nil {
		return fmt.Errorf("failed to serialize event details: %w", err)
	}
	
	_, err = r.db.Exec(query,
		audit.AuditID, audit.EventType, audit.LinkID, audit.ReplyID, audit.AttachmentID,
		audit.PolicyID, audit.RuleID, audit.UserID, audit.IPAddress, audit.UserAgent,
		audit.Severity, audit.ComplianceCategory, audit.RetentionRequired, eventDetails,
		audit.CreatedAt,
	)
	return err
}

// GetSecureLink retrieves a secure link by ID
func (r *SQLiteSecurityRepository) GetSecureLink(linkID string) (*securelinks.SecureLink, error) {
	query := `
		SELECT link_id, email_id, recipient_email, sender_id, security_settings,
		       created_at, expires_at, access_count, last_accessed, status,
		       failed_attempts, last_failed_attempt, lockout_until
		FROM secure_links WHERE link_id = ?
	`
	
	var link securelinks.SecureLink
	var securitySettingsJSON string
	err := r.db.QueryRow(query, linkID).Scan(
		&link.LinkID, &link.EmailID, &link.RecipientEmail, &link.SenderID, &securitySettingsJSON,
		&link.CreatedAt, &link.ExpiresAt, &link.AccessCount, &link.LastAccessed, &link.Status,
		&link.FailedAttempts, &link.LastFailedAttempt, &link.LockoutUntil,
	)
	if err != nil {
		return nil, err
	}
	return &link, nil
}

// UpdateSecureLink updates an existing secure link
func (r *SQLiteSecurityRepository) UpdateSecureLink(link *securelinks.SecureLink) error {
	query := `
		UPDATE secure_links SET
			email_id = ?, recipient_email = ?, sender_id = ?, security_settings = ?,
			expires_at = ?, access_count = ?, last_accessed = ?, status = ?,
			failed_attempts = ?, last_failed_attempt = ?, lockout_until = ?
		WHERE link_id = ?
	`
	
	securitySettingsJSON, err := json.Marshal(link.SecuritySettings)
	if err != nil {
		return fmt.Errorf("failed to marshal security settings: %w", err)
	}
	
	_, err = r.db.Exec(query,
		link.EmailID, link.RecipientEmail, link.SenderID, securitySettingsJSON,
		link.ExpiresAt, link.AccessCount, link.LastAccessed, link.Status,
		link.FailedAttempts, link.LastFailedAttempt, link.LockoutUntil, link.LinkID,
	)
	return err
}
