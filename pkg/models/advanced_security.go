package models

import (
	"encoding/json"
	"time"
)

// DLPRule represents a data loss prevention rule
type DLPRule struct {
	RuleID      string    `json:"rule_id" db:"rule_id"`
	RuleName    string    `json:"rule_name" db:"rule_name"`
	RuleType    string    `json:"rule_type" db:"rule_type"` // 'regex', 'keyword', 'ai_pattern'
	Pattern     string    `json:"pattern" db:"pattern"`
	Description *string   `json:"description,omitempty" db:"description"`
	Severity    string    `json:"severity" db:"severity"` // 'low', 'medium', 'high', 'critical'
	Action      string    `json:"action" db:"action"`     // 'allow', 'warn', 'block'
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	CreatedBy   *string   `json:"created_by,omitempty" db:"created_by"`
	Priority    int       `json:"priority" db:"priority"`
}

// SecurityPolicy represents security controls for a message
type SecurityPolicy struct {
	PolicyID              string     `json:"policy_id" db:"policy_id"`
	LinkID                string     `json:"link_id" db:"link_id"`
	ReplyID               *string    `json:"reply_id,omitempty" db:"reply_id"`
	EmailID               *string    `json:"email_id,omitempty" db:"email_id"`
	DLPEnabled            bool       `json:"dlp_enabled" db:"dlp_enabled"`
	WatermarkEnabled      bool       `json:"watermark_enabled" db:"watermark_enabled"`
	DownloadDisabled      bool       `json:"download_disabled" db:"download_disabled"`
	ForwardingDisabled    bool       `json:"forwarding_disabled" db:"forwarding_disabled"`
	AutoRevokeAfterReply  bool       `json:"auto_revoke_after_reply" db:"auto_revoke_after_reply"`
	MaxViews              *int       `json:"max_views,omitempty" db:"max_views"`
	ExpiresAt             *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	ExpiresAfterViews     *int       `json:"expires_after_views,omitempty" db:"expires_after_views"`
	NotifyOnExpiry        bool       `json:"notify_on_expiry" db:"notify_on_expiry"`
	NotifyOnRevoke        bool       `json:"notify_on_revoke" db:"notify_on_revoke"`
	CreatedAt             time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at" db:"updated_at"`
	CreatedBy             *string    `json:"created_by,omitempty" db:"created_by"`
}

// DLPScanResult represents the result of a DLP scan
type DLPScanResult struct {
	ScanID         string     `json:"scan_id" db:"scan_id"`
	LinkID         *string    `json:"link_id,omitempty" db:"link_id"`
	ReplyID        *string    `json:"reply_id,omitempty" db:"reply_id"`
	AttachmentID   *string    `json:"attachment_id,omitempty" db:"attachment_id"`
	RuleID         string     `json:"rule_id" db:"rule_id"`
	ContentType    string     `json:"content_type" db:"content_type"` // 'email_body', 'reply_body', 'attachment'
	MatchedContent *string    `json:"matched_content,omitempty" db:"matched_content"`
	ConfidenceScore float64   `json:"confidence_score" db:"confidence_score"`
	ActionTaken    string     `json:"action_taken" db:"action_taken"` // 'allowed', 'warned', 'blocked'
	ScanTimestamp  time.Time  `json:"scan_timestamp" db:"scan_timestamp"`
	CreatedBy      *string    `json:"created_by,omitempty" db:"created_by"`
}

// WatermarkConfig represents watermarking configuration for attachments
type WatermarkConfig struct {
	ConfigID         string     `json:"config_id" db:"config_id"`
	AttachmentID     string     `json:"attachment_id" db:"attachment_id"`
	WatermarkText    string     `json:"watermark_text" db:"watermark_text"`
	WatermarkPosition string    `json:"watermark_position" db:"watermark_position"` // 'top-left', 'top-right', 'bottom-left', 'bottom-right', 'center'
	WatermarkOpacity float64    `json:"watermark_opacity" db:"watermark_opacity"`
	WatermarkFontSize int       `json:"watermark_font_size" db:"watermark_font_size"`
	WatermarkColor   string     `json:"watermark_color" db:"watermark_color"`
	WatermarkRotation int       `json:"watermark_rotation" db:"watermark_rotation"`
	AppliedAt        time.Time  `json:"applied_at" db:"applied_at"`
	WatermarkHash    *string    `json:"watermark_hash,omitempty" db:"watermark_hash"`
	CreatedBy        *string    `json:"created_by,omitempty" db:"created_by"`
	// Advanced watermarking fields (Iteration 8)
	RecipientEmail   *string    `json:"recipient_email,omitempty" db:"recipient_email"`
	RecipientID      *string    `json:"recipient_id,omitempty" db:"recipient_id"`
	WatermarkType    string     `json:"watermark_type" db:"watermark_type"` // 'text', 'image', 'audio', 'video', 'inline'
	ContentType      string     `json:"content_type" db:"content_type"` // 'pdf', 'image', 'document', 'audio', 'video', 'email_content'
	WatermarkData    *string    `json:"watermark_data,omitempty" db:"watermark_data"` // JSON data for complex watermarks
	IsRecipientSpecific bool    `json:"is_recipient_specific" db:"is_recipient_specific"`
}

// ComplianceAuditLog represents immutable compliance audit records
type ComplianceAuditLog struct {
	AuditID           string     `json:"audit_id" db:"audit_id"`
	EventType         string     `json:"event_type" db:"event_type"` // 'dlp_scan', 'watermark_applied', 'policy_enforced', 'expiration_triggered', 'revocation_triggered'
	LinkID            *string    `json:"link_id,omitempty" db:"link_id"`
	ReplyID           *string    `json:"reply_id,omitempty" db:"reply_id"`
	AttachmentID      *string    `json:"attachment_id,omitempty" db:"attachment_id"`
	PolicyID          *string    `json:"policy_id,omitempty" db:"policy_id"`
	RuleID            *string    `json:"rule_id,omitempty" db:"rule_id"`
	UserID            *string    `json:"user_id,omitempty" db:"user_id"`
	IPAddress         *string    `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent         *string    `json:"user_agent,omitempty" db:"user_agent"`
	EventDetails      *string    `json:"event_details,omitempty" db:"event_details"` // JSON details
	Severity          string     `json:"severity" db:"severity"`                     // 'info', 'warning', 'error', 'critical'
	ComplianceCategory *string   `json:"compliance_category,omitempty" db:"compliance_category"` // 'dlp', 'watermarking', 'expiration', 'revocation', 'access_control'
	RetentionRequired  bool      `json:"retention_required" db:"retention_required"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	CreatedBy         *string    `json:"created_by,omitempty" db:"created_by"`
}

// SecurityPolicyTemplate represents reusable security policy templates
type SecurityPolicyTemplate struct {
	TemplateID          string     `json:"template_id" db:"template_id"`
	TemplateName        string     `json:"template_name" db:"template_name"`
	TemplateDescription *string    `json:"template_description,omitempty" db:"template_description"`
	DLPEnabled          bool       `json:"dlp_enabled" db:"dlp_enabled"`
	WatermarkEnabled    bool       `json:"watermark_enabled" db:"watermark_enabled"`
	DownloadDisabled    bool       `json:"download_disabled" db:"download_disabled"`
	ForwardingDisabled  bool       `json:"forwarding_disabled" db:"forwarding_disabled"`
	AutoRevokeAfterReply bool      `json:"auto_revoke_after_reply" db:"auto_revoke_after_reply"`
	MaxViews            *int       `json:"max_views,omitempty" db:"max_views"`
	DefaultExpiryHours  *int       `json:"default_expiry_hours,omitempty" db:"default_expiry_hours"`
	NotifyOnExpiry      bool       `json:"notify_on_expiry" db:"notify_on_expiry"`
	NotifyOnRevoke      bool       `json:"notify_on_revoke" db:"notify_on_revoke"`
	IsDefault           bool       `json:"is_default" db:"is_default"`
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at" db:"updated_at"`
	CreatedBy           *string    `json:"created_by,omitempty" db:"created_by"`
}

// Request/Response structs for API endpoints

// CreateSecurityPolicyRequest represents a request to create a security policy
type CreateSecurityPolicyRequest struct {
	LinkID                string     `json:"link_id"`
	ReplyID               *string    `json:"reply_id,omitempty"`
	EmailID               *string    `json:"email_id,omitempty"`
	DLPEnabled            bool       `json:"dlp_enabled"`
	WatermarkEnabled      bool       `json:"watermark_enabled"`
	DownloadDisabled      bool       `json:"download_disabled"`
	ForwardingDisabled    bool       `json:"forwarding_disabled"`
	AutoRevokeAfterReply  bool       `json:"auto_revoke_after_reply"`
	MaxViews              *int       `json:"max_views,omitempty"`
	ExpiresAt             *time.Time `json:"expires_at,omitempty"`
	ExpiresAfterViews     *int       `json:"expires_after_views,omitempty"`
	NotifyOnExpiry        bool       `json:"notify_on_expiry"`
	NotifyOnRevoke        bool       `json:"notify_on_revoke"`
	TemplateID            *string    `json:"template_id,omitempty"` // Optional template to apply
}

// SecurityPolicyResponse represents a response with security policy information
type SecurityPolicyResponse struct {
	Success       bool             `json:"success"`
	PolicyID      string           `json:"policy_id"`
	Policy        *SecurityPolicy  `json:"policy,omitempty"`
	Message       string           `json:"message,omitempty"`
	Error         string           `json:"error,omitempty"`
	ErrorCode     string           `json:"error_code,omitempty"`
}

// DLPScanRequest represents a request to scan content for DLP violations
type DLPScanRequest struct {
	Content     string `json:"content"`
	ContentType string `json:"content_type"` // 'email_body', 'reply_body', 'attachment'
	LinkID      string `json:"link_id"`
	ReplyID     *string `json:"reply_id,omitempty"`
	AttachmentID *string `json:"attachment_id,omitempty"`
}

// DLPScanResponse represents the result of a DLP scan
type DLPScanResponse struct {
	Success      bool            `json:"success"`
	Violations   []DLPScanResult `json:"violations,omitempty"`
	ActionTaken  string          `json:"action_taken"` // 'allowed', 'warned', 'blocked'
	Message      string          `json:"message,omitempty"`
	Error        string          `json:"error,omitempty"`
	ErrorCode    string          `json:"error_code,omitempty"`
}

// WatermarkRequest represents a request to apply watermarking
type WatermarkRequest struct {
	AttachmentID      string  `json:"attachment_id"`
	WatermarkText     string  `json:"watermark_text"`
	WatermarkPosition string  `json:"watermark_position,omitempty"`
	WatermarkOpacity  float64 `json:"watermark_opacity,omitempty"`
	WatermarkFontSize int     `json:"watermark_font_size,omitempty"`
	WatermarkColor    string  `json:"watermark_color,omitempty"`
	WatermarkRotation int     `json:"watermark_rotation,omitempty"`
	// Advanced watermarking fields (Iteration 8)
	RecipientEmail    *string `json:"recipient_email,omitempty"`
	RecipientID       *string `json:"recipient_id,omitempty"`
	WatermarkType     string  `json:"watermark_type,omitempty"` // 'text', 'image', 'audio', 'video', 'inline'
	ContentType       string  `json:"content_type,omitempty"` // 'pdf', 'image', 'document', 'audio', 'video', 'email_content'
	WatermarkData     *string `json:"watermark_data,omitempty"` // JSON data for complex watermarks
	IsRecipientSpecific bool  `json:"is_recipient_specific,omitempty"`
}

// WatermarkResponse represents the result of watermarking
type WatermarkResponse struct {
	Success       bool            `json:"success"`
	ConfigID      string          `json:"config_id"`
	WatermarkedURL string         `json:"watermarked_url,omitempty"`
	Message       string          `json:"message,omitempty"`
	Error         string          `json:"error,omitempty"`
	ErrorCode     string          `json:"error_code,omitempty"`
}

// AdvancedWatermarkRequest represents a request for advanced watermarking features
type AdvancedWatermarkRequest struct {
	LinkID            string                 `json:"link_id"`
	AttachmentID      *string                `json:"attachment_id,omitempty"`
	ContentID         *string                `json:"content_id,omitempty"` // For inline content
	WatermarkType     string                 `json:"watermark_type"` // 'text', 'image', 'audio', 'video', 'inline'
	ContentType       string                 `json:"content_type"` // 'pdf', 'image', 'document', 'audio', 'video', 'email_content'
	RecipientEmail    string                 `json:"recipient_email"`
	RecipientID       *string                `json:"recipient_id,omitempty"`
	WatermarkConfig   map[string]interface{} `json:"watermark_config"`
	IsRecipientSpecific bool                 `json:"is_recipient_specific"`
	ApplyToAllContent bool                   `json:"apply_to_all_content,omitempty"` // For applying to all content in a link
}

// AdvancedWatermarkResponse represents the result of advanced watermarking
type AdvancedWatermarkResponse struct {
	Success           bool                   `json:"success"`
	ConfigID          string                 `json:"config_id"`
	WatermarkedURL    string                 `json:"watermarked_url,omitempty"`
	WatermarkedContent *string               `json:"watermarked_content,omitempty"` // For inline content
	Message           string                 `json:"message,omitempty"`
	Error             string                 `json:"error,omitempty"`
	ErrorCode         string                 `json:"error_code,omitempty"`
	AppliedTo         []string               `json:"applied_to,omitempty"` // List of content IDs that were watermarked
	RecipientInfo     map[string]interface{} `json:"recipient_info,omitempty"`
}

// WatermarkTemplate represents a reusable watermark template
type WatermarkTemplate struct {
	TemplateID        string                 `json:"template_id" db:"template_id"`
	TemplateName      string                 `json:"template_name" db:"template_name"`
	TemplateDescription string               `json:"template_description" db:"template_description"`
	WatermarkType     string                 `json:"watermark_type" db:"watermark_type"`
	ContentTypes      string                 `json:"content_types" db:"content_types"` // JSON array of supported content types
	DefaultConfig     string                 `json:"default_config" db:"default_config"` // JSON configuration
	IsRecipientSpecific bool                 `json:"is_recipient_specific" db:"is_recipient_specific"`
	IsActive          bool                   `json:"is_active" db:"is_active"`
	CreatedAt         time.Time              `json:"created_at" db:"created_at"`
	CreatedBy         *string                `json:"created_by,omitempty" db:"created_by"`
}

// ComplianceAuditResponse represents a response for compliance audit logging
type ComplianceAuditResponse struct {
	Success   bool   `json:"success"`
	AuditID   string `json:"audit_id"`
	Message   string `json:"message,omitempty"`
	Error     string `json:"error,omitempty"`
	ErrorCode string `json:"error_code,omitempty"`
}

// SecurityPolicyTemplateResponse represents a response with security policy templates
type SecurityPolicyTemplateResponse struct {
	Success   bool                      `json:"success"`
	Templates []SecurityPolicyTemplate  `json:"templates,omitempty"`
	Message   string                    `json:"message,omitempty"`
	Error     string                    `json:"error,omitempty"`
	ErrorCode string                    `json:"error_code,omitempty"`
}

// Helper methods

// ToJSON converts event details to JSON string
func (cal *ComplianceAuditLog) SetEventDetails(details map[string]interface{}) error {
	jsonData, err := json.Marshal(details)
	if err != nil {
		return err
	}
	detailsStr := string(jsonData)
	cal.EventDetails = &detailsStr
	return nil
}

// GetEventDetails converts JSON string back to map
func (cal *ComplianceAuditLog) GetEventDetails() (map[string]interface{}, error) {
	if cal.EventDetails == nil {
		return nil, nil
	}
	var details map[string]interface{}
	err := json.Unmarshal([]byte(*cal.EventDetails), &details)
	return details, err
}

// IsExpired checks if a security policy has expired
func (sp *SecurityPolicy) IsExpired() bool {
	if sp.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*sp.ExpiresAt)
}

// ShouldExpireAfterViews checks if the policy should expire based on view count
func (sp *SecurityPolicy) ShouldExpireAfterViews(currentViews int) bool {
	if sp.ExpiresAfterViews == nil {
		return false
	}
	return currentViews >= *sp.ExpiresAfterViews
}

// GetExpiryTime returns the effective expiry time
func (sp *SecurityPolicy) GetExpiryTime() *time.Time {
	if sp.ExpiresAt != nil {
		return sp.ExpiresAt
	}
	return nil
}

// WatermarkAuditLog represents audit records for advanced watermarking operations
type WatermarkAuditLog struct {
	AuditID         string     `json:"audit_id" db:"audit_id"`
	LinkID          *string    `json:"link_id,omitempty" db:"link_id"`
	AttachmentID    *string    `json:"attachment_id,omitempty" db:"attachment_id"`
	WatermarkType   string     `json:"watermark_type" db:"watermark_type"`
	ContentType     string     `json:"content_type" db:"content_type"`
	RecipientEmail  *string    `json:"recipient_email,omitempty" db:"recipient_email"`
	RecipientID     *string    `json:"recipient_id,omitempty" db:"recipient_id"`
	WatermarkData   *string    `json:"watermark_data,omitempty" db:"watermark_data"` // JSON data
	ProcessingTime  float64    `json:"processing_time" db:"processing_time"`
	Success         bool       `json:"success" db:"success"`
	ErrorMessage    *string    `json:"error_message,omitempty" db:"error_message"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	CreatedBy       *string    `json:"created_by,omitempty" db:"created_by"`
}

// ComplianceAuditRequest represents a request to log a compliance event
type ComplianceAuditRequest struct {
	EventType         string                 `json:"event_type"`
	LinkID            *string                `json:"link_id,omitempty"`
	ReplyID           *string                `json:"reply_id,omitempty"`
	AttachmentID      *string                `json:"attachment_id,omitempty"`
	PolicyID          *string                `json:"policy_id,omitempty"`
	RuleID            *string                `json:"rule_id,omitempty"`
	UserID            *string                `json:"user_id,omitempty"`
	IPAddress         *string                `json:"ip_address,omitempty"`
	UserAgent         *string                `json:"user_agent,omitempty"`
	EventDetails      map[string]interface{} `json:"event_details,omitempty"`
	Severity          string                 `json:"severity,omitempty"`
	ComplianceCategory *string               `json:"compliance_category,omitempty"`
	RetentionRequired  bool                  `json:"retention_required,omitempty"`
}
