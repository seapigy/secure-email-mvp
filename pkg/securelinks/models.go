package securelinks

import (
	"encoding/json"
	"time"
)

// =============================================================================
// SECURE LINK MODELS
// =============================================================================

// SecureLink represents a secure link for external email access
type SecureLink struct {
	LinkID            string           `json:"link_id" db:"link_id"`
	EmailID           string           `json:"email_id" db:"email_id"`
	RecipientEmail    string           `json:"recipient_email" db:"recipient_email"`
	SenderID          string           `json:"sender_id" db:"sender_id"`
	SecuritySettings  SecuritySettings `json:"security_settings" db:"security_settings"`
	CreatedAt         time.Time        `json:"created_at" db:"created_at"`
	ExpiresAt         *time.Time       `json:"expires_at,omitempty" db:"expires_at"`
	AccessCount       int              `json:"access_count" db:"access_count"`
	LastAccessed      *time.Time       `json:"last_accessed,omitempty" db:"last_accessed"`
	Status            LinkStatus       `json:"status" db:"status"`
	FailedAttempts    int              `json:"failed_attempts" db:"failed_attempts"`
	LastFailedAttempt *time.Time       `json:"last_failed_attempt,omitempty" db:"last_failed_attempt"`
	LockoutUntil      *time.Time       `json:"lockout_until,omitempty" db:"lockout_until"`
}

// LinkStatus represents the current status of a secure link
type LinkStatus string

const (
	LinkStatusActive    LinkStatus = "active"
	LinkStatusRevoked   LinkStatus = "revoked"
	LinkStatusExpired   LinkStatus = "expired"
	LinkStatusDestroyed LinkStatus = "destroyed"
)

// SecuritySettings contains all security configuration for a secure link
type SecuritySettings struct {
	// Inherit from email security toggles
	NotBefore             *int64  `json:"not_before,omitempty"`
	ExpiresAt             *int64  `json:"expires_at,omitempty"`
	ReadOnce              bool    `json:"read_once"`
	MFAOnOpen             bool    `json:"mfa_on_open"`
	DecoySecret           *string `json:"decoy_secret,omitempty"`
	RemoteRevoke          bool    `json:"remote_revoke"`
	StripMetadata         bool    `json:"strip_metadata"`
	SelfDestructThreshold *int    `json:"self_destruct_threshold,omitempty"`
	GeoRulesRef           *string `json:"geo_rules_ref,omitempty"`

	// Link-specific settings
	RequirePassword   bool    `json:"require_password"`
	PasswordHash      *string `json:"password_hash,omitempty"`
	MaxAccessAttempts int     `json:"max_access_attempts"`
	RequireMFA        bool    `json:"require_mfa"`
	MFAType           string  `json:"mfa_type"` // "totp", "email", "sms"

	// Additional link security
	AutoDestruct           bool     `json:"auto_destruct"`
	TimeLock               bool     `json:"time_lock"`
	TimeLockUntil          *int64   `json:"time_lock_until,omitempty"`
	GeolocationRestriction bool     `json:"geolocation_restriction"`
	AllowedCountries       []string `json:"allowed_countries,omitempty"`
	AllowedCities          []string `json:"allowed_cities,omitempty"`
}

// =============================================================================
// AUDIT LOG MODELS
// =============================================================================

// LinkAuditEvent represents an audit event for a secure link
type LinkAuditEvent struct {
	ID              string           `json:"id" db:"id"`
	LinkID          string           `json:"link_id" db:"link_id"`
	EventType       AuditEventType   `json:"event_type" db:"event_type"`
	IPAddress       string           `json:"ip_address" db:"ip_address"`
	UserAgent       string           `json:"user_agent" db:"user_agent"`
	GeolocationData *GeolocationData `json:"geolocation_data,omitempty" db:"geolocation_data"`
	Timestamp       time.Time        `json:"timestamp" db:"timestamp"`
	Details         string           `json:"details" db:"details"`
	Severity        AuditSeverity    `json:"severity" db:"severity"`
}

// AuditEventType represents the type of audit event
type AuditEventType string

const (
	// Link lifecycle events
	AuditEventCreated   AuditEventType = "created"
	AuditEventAccessed  AuditEventType = "accessed"
	AuditEventRevoked   AuditEventType = "revoked"
	AuditEventExpired   AuditEventType = "expired"
	AuditEventDestroyed AuditEventType = "destroyed"

	// Security check events
	AuditEventPasswordRequired  AuditEventType = "password_required"
	AuditEventPasswordValidated AuditEventType = "password_validated"
	AuditEventPasswordFailed    AuditEventType = "password_failed"
	AuditEventFailedAttempt     AuditEventType = "failed_attempt"

	// Geolocation events
	AuditEventGeolocationCheck   AuditEventType = "geolocation_check"
	AuditEventGeolocationBlocked AuditEventType = "geolocation_blocked"

	// MFA events
	AuditEventMFARequired  AuditEventType = "mfa_required"
	AuditEventMFAValidated AuditEventType = "mfa_validated"
	AuditEventMFAFailed    AuditEventType = "mfa_failed"

	// Time and access events
	AuditEventTimeLockActive        AuditEventType = "time_lock_active"
	AuditEventReadOnceConsumed      AuditEventType = "read_once_consumed"
	AuditEventAutoDestructTriggered AuditEventType = "auto_destruct_triggered"
)

// AuditSeverity represents the severity level of an audit event
type AuditSeverity string

const (
	AuditSeverityInfo     AuditSeverity = "info"
	AuditSeverityWarning  AuditSeverity = "warning"
	AuditSeverityError    AuditSeverity = "error"
	AuditSeverityCritical AuditSeverity = "critical"
)

// GeolocationData contains geolocation information for audit events
type GeolocationData struct {
	Country   string  `json:"country,omitempty"`
	City      string  `json:"city,omitempty"`
	Region    string  `json:"region,omitempty"`
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
	Timezone  string  `json:"timezone,omitempty"`
	ISP       string  `json:"isp,omitempty"`
	ASN       string  `json:"asn,omitempty"`
}

// =============================================================================
// EMAIL CHAIN MODELS
// =============================================================================

// EmailChain represents an email conversation chain
type EmailChain struct {
	ChainID         string      `json:"chain_id" db:"chain_id"`
	OriginalEmailID string      `json:"original_email_id" db:"original_email_id"`
	CurrentLinkID   string      `json:"current_link_id" db:"current_link_id"`
	ChainDepth      int         `json:"chain_depth" db:"chain_depth"`
	CreatedAt       time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at" db:"updated_at"`
	Status          ChainStatus `json:"status" db:"status"`
}

// ChainStatus represents the status of an email chain
type ChainStatus string

const (
	ChainStatusActive   ChainStatus = "active"
	ChainStatusClosed   ChainStatus = "closed"
	ChainStatusArchived ChainStatus = "archived"
)

// =============================================================================
// TEMPLATE MODELS
// =============================================================================

// SecureLinkTemplate represents a message template for external recipients
type SecureLinkTemplate struct {
	ID              string    `json:"id" db:"id"`
	TemplateName    string    `json:"template_name" db:"template_name"`
	TemplateContent string    `json:"template_content" db:"template_content"`
	IsDefault       bool      `json:"is_default" db:"is_default"`
	CreatedBy       *string   `json:"created_by,omitempty" db:"created_by"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
	IsActive        bool      `json:"is_active" db:"is_active"`
}

// =============================================================================
// REQUEST/RESPONSE MODELS
// =============================================================================

// CreateSecureLinkRequest represents a request to create a secure link
type CreateSecureLinkRequest struct {
	EmailID          string           `json:"email_id" validate:"required"`
	RecipientEmail   string           `json:"recipient_email" validate:"required,email"`
	SecuritySettings SecuritySettings `json:"security_settings"`
	CustomMessage    *string          `json:"custom_message,omitempty"`
}

// CreateSecureLinkResponse represents the response from creating a secure link
type CreateSecureLinkResponse struct {
	LinkID       string       `json:"link_id"`
	SecureURL    string       `json:"secure_url"`
	ExpiresAt    *time.Time   `json:"expires_at,omitempty"`
	CreatedAt    time.Time    `json:"created_at"`
	SecurityInfo SecurityInfo `json:"security_info"`
}

// SecurityInfo provides information about the security features enabled
type SecurityInfo struct {
	RequirePassword        bool       `json:"require_password"`
	RequireMFA             bool       `json:"require_mfa"`
	MFAType                string     `json:"mfa_type,omitempty"`
	TimeLock               bool       `json:"time_lock"`
	TimeLockUntil          *time.Time `json:"time_lock_until,omitempty"`
	GeolocationRestriction bool       `json:"geolocation_restriction"`
	AllowedCountries       []string   `json:"allowed_countries,omitempty"`
	ReadOnce               bool       `json:"read_once"`
	AutoDestruct           bool       `json:"auto_destruct"`
	MaxAccessAttempts      int        `json:"max_access_attempts"`
}

// AccessSecureLinkRequest represents a request to access a secure link
type AccessSecureLinkRequest struct {
	LinkID    string  `json:"link_id" validate:"required"`
	Password  *string `json:"password,omitempty"`
	MFACode   *string `json:"mfa_code,omitempty"`
	IPAddress string  `json:"ip_address"`
	UserAgent string  `json:"user_agent"`
}

// AccessSecureLinkResponse represents the response from accessing a secure link
type AccessSecureLinkResponse struct {
	Success       bool           `json:"success"`
	EmailContent  *EmailContent  `json:"email_content,omitempty"`
	SecurityCheck *SecurityCheck `json:"security_check,omitempty"`
	Error         *AccessError   `json:"error,omitempty"`
}

// EmailContent contains the decrypted email content
type EmailContent struct {
	Subject     string       `json:"subject"`
	Body        string       `json:"body"`
	SenderName  string       `json:"sender_name"`
	SenderEmail string       `json:"sender_email"`
	SentAt      time.Time    `json:"sent_at"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

// Attachment represents an email attachment
type Attachment struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Size     int64  `json:"size"`
	MimeType string `json:"mime_type"`
	URL      string `json:"url,omitempty"`
}

// SecurityCheck represents a security check that needs to be completed
type SecurityCheck struct {
	Type         string `json:"type"` // "password", "mfa", "geolocation"
	Required     bool   `json:"required"`
	Message      string `json:"message"`
	AttemptsLeft int    `json:"attempts_left,omitempty"`
}

// AccessError represents an error that occurred during link access
type AccessError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

// IsValid checks if a link status is valid
func (s LinkStatus) IsValid() bool {
	switch s {
	case LinkStatusActive, LinkStatusRevoked, LinkStatusExpired, LinkStatusDestroyed:
		return true
	default:
		return false
	}
}

// IsActive checks if a link is currently active
func (s LinkStatus) IsActive() bool {
	return s == LinkStatusActive
}

// IsExpired checks if a link has expired
func (s LinkStatus) IsExpired() bool {
	return s == LinkStatusExpired
}

// IsRevoked checks if a link has been revoked
func (s LinkStatus) IsRevoked() bool {
	return s == LinkStatusRevoked
}

// IsDestroyed checks if a link has been destroyed
func (s LinkStatus) IsDestroyed() bool {
	return s == LinkStatusDestroyed
}

// IsValid checks if an audit event type is valid
func (t AuditEventType) IsValid() bool {
	switch t {
	case AuditEventCreated, AuditEventAccessed, AuditEventRevoked, AuditEventExpired,
		AuditEventDestroyed, AuditEventPasswordRequired, AuditEventPasswordValidated,
		AuditEventPasswordFailed, AuditEventFailedAttempt, AuditEventGeolocationCheck,
		AuditEventGeolocationBlocked, AuditEventMFARequired, AuditEventMFAValidated,
		AuditEventMFAFailed, AuditEventTimeLockActive, AuditEventReadOnceConsumed,
		AuditEventAutoDestructTriggered:
		return true
	default:
		return false
	}
}

// IsValid checks if an audit severity is valid
func (s AuditSeverity) IsValid() bool {
	switch s {
	case AuditSeverityInfo, AuditSeverityWarning, AuditSeverityError, AuditSeverityCritical:
		return true
	default:
		return false
	}
}

// IsValid checks if a chain status is valid
func (s ChainStatus) IsValid() bool {
	switch s {
	case ChainStatusActive, ChainStatusClosed, ChainStatusArchived:
		return true
	default:
		return false
	}
}

// =============================================================================
// JSON MARSHALING HELPERS
// =============================================================================

// MarshalJSON implements custom JSON marshaling for SecuritySettings
func (s SecuritySettings) MarshalJSON() ([]byte, error) {
	type Alias SecuritySettings
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(&s),
	})
}

// UnmarshalJSON implements custom JSON unmarshaling for SecuritySettings
func (s *SecuritySettings) UnmarshalJSON(data []byte) error {
	type Alias SecuritySettings
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(s),
	}
	return json.Unmarshal(data, &aux)
}

// MarshalJSON implements custom JSON marshaling for GeolocationData
func (g GeolocationData) MarshalJSON() ([]byte, error) {
	type Alias GeolocationData
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(&g),
	})
}

// UnmarshalJSON implements custom JSON unmarshaling for GeolocationData
func (g *GeolocationData) UnmarshalJSON(data []byte) error {
	type Alias GeolocationData
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(g),
	}
	return json.Unmarshal(data, &aux)
}
