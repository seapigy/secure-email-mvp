package security

import (
	"encoding/json"
	"time"
)

// =============================================================================
// PHASE 2 SECURITY MODELS
// =============================================================================
// Comprehensive data models for all Phase 2 security enforcement features
// =============================================================================

// SecuritySettings represents the complete security configuration for a secure link
type SecuritySettings struct {
	// Password Protection
	PasswordRequired     bool       `json:"password_required" db:"password_required"`
	PasswordHash         string     `json:"password_hash,omitempty" db:"password_hash"`
	PasswordMaxAttempts  int        `json:"password_max_attempts" db:"password_max_attempts"`
	PasswordLockoutUntil *time.Time `json:"password_lockout_until,omitempty" db:"password_lockout_until"`

	// Time Lock
	TimeLockEnabled  bool       `json:"time_lock_enabled" db:"time_lock_enabled"`
	TimeLockUntil    *time.Time `json:"time_lock_until,omitempty" db:"time_lock_until"`
	TimeLockTimezone string     `json:"time_lock_timezone" db:"time_lock_timezone"`

	// Auto-Destruct & Read-Once
	AutoDestructEnabled       bool `json:"auto_destruct_enabled" db:"auto_destruct_enabled"`
	ReadOnceEnabled           bool `json:"read_once_enabled" db:"read_once_enabled"`
	ReadOnceConsumed          bool `json:"read_once_consumed" db:"read_once_consumed"`
	AutoDestructAfterAttempts int  `json:"auto_destruct_after_attempts" db:"auto_destruct_after_attempts"`

	// Enhanced Geolocation
	GeoRestrictionEnabled bool     `json:"geo_restriction_enabled" db:"geo_restriction_enabled"`
	GeoAllowedCountries   []string `json:"geo_allowed_countries" db:"geo_allowed_countries"`
	GeoAllowedCities      []string `json:"geo_allowed_cities" db:"geo_allowed_cities"`
	GeoBlockedCountries   []string `json:"geo_blocked_countries" db:"geo_blocked_countries"`
	GeoBlockedCities      []string `json:"geo_blocked_cities" db:"geo_blocked_cities"`

	// Multi-Factor Authentication
	MFARequired bool   `json:"mfa_required" db:"mfa_required"`
	MFAType     string `json:"mfa_type" db:"mfa_type"` // totp, email, sms, none
	MFASecret   string `json:"mfa_secret,omitempty" db:"mfa_secret"`
	MFAEmail    string `json:"mfa_email,omitempty" db:"mfa_email"`
	MFAPhone    string `json:"mfa_phone,omitempty" db:"mfa_phone"`

	// Decoy Messages
	DecoyEnabled           bool   `json:"decoy_enabled" db:"decoy_enabled"`
	DecoyMessage           string `json:"decoy_message,omitempty" db:"decoy_message"`
	DecoyTriggerConditions string `json:"decoy_trigger_conditions,omitempty" db:"decoy_trigger_conditions"`

	// Metadata Stripping
	StripMetadataEnabled bool `json:"strip_metadata_enabled" db:"strip_metadata_enabled"`
	StripHeaders         bool `json:"strip_headers" db:"strip_headers"`
	StripAttachments     bool `json:"strip_attachments" db:"strip_attachments"`

	// Tamper Alerts
	TamperAlertsEnabled  bool `json:"tamper_alerts_enabled" db:"tamper_alerts_enabled"`
	TamperAlertThreshold int  `json:"tamper_alert_threshold" db:"tamper_alert_threshold"`

	// Remote Revocation
	RevokedBy        string     `json:"revoked_by,omitempty" db:"revoked_by"`
	RevokedAt        *time.Time `json:"revoked_at,omitempty" db:"revoked_at"`
	RevocationReason string     `json:"revocation_reason,omitempty" db:"revocation_reason"`
}

// PasswordAttempt represents a password attempt for a secure link
type PasswordAttempt struct {
	ID              string          `json:"id" db:"id"`
	LinkID          string          `json:"link_id" db:"link_id"`
	IPAddress       string          `json:"ip_address" db:"ip_address"`
	UserAgent       string          `json:"user_agent,omitempty" db:"user_agent"`
	AttemptTime     time.Time       `json:"attempt_time" db:"attempt_time"`
	Success         bool            `json:"success" db:"success"`
	AttemptNumber   int             `json:"attempt_number" db:"attempt_number"`
	GeolocationData json.RawMessage `json:"geolocation_data,omitempty" db:"geolocation_data"`
}

// GeolocationLog represents enhanced geolocation tracking for security enforcement
type GeolocationLog struct {
	ID                string          `json:"id" db:"id"`
	LinkID            string          `json:"link_id" db:"link_id"`
	IPAddress         string          `json:"ip_address" db:"ip_address"`
	Country           string          `json:"country,omitempty" db:"country"`
	City              string          `json:"city,omitempty" db:"city"`
	Region            string          `json:"region,omitempty" db:"region"`
	Latitude          float64         `json:"latitude,omitempty" db:"latitude"`
	Longitude         float64         `json:"longitude,omitempty" db:"longitude"`
	Timezone          string          `json:"timezone,omitempty" db:"timezone"`
	ISP               string          `json:"isp,omitempty" db:"isp"`
	AccessTime        time.Time       `json:"access_time" db:"access_time"`
	AccessAllowed     bool            `json:"access_allowed" db:"access_allowed"`
	RestrictionReason string          `json:"restriction_reason,omitempty" db:"restriction_reason"`
	GeolocationData   json.RawMessage `json:"geolocation_data,omitempty" db:"geolocation_data"`
}

// MFASession represents an MFA session for external secure link users
type MFASession struct {
	ID           string     `json:"id" db:"id"`
	LinkID       string     `json:"link_id" db:"link_id"`
	SessionToken string     `json:"session_token" db:"session_token"`
	MFAType      string     `json:"mfa_type" db:"mfa_type"`
	MFASecret    string     `json:"mfa_secret,omitempty" db:"mfa_secret"`
	MFAEmail     string     `json:"mfa_email,omitempty" db:"mfa_email"`
	MFAPhone     string     `json:"mfa_phone,omitempty" db:"mfa_phone"`
	OTPCode      string     `json:"otp_code,omitempty" db:"otp_code"`
	OTPExpiresAt *time.Time `json:"otp_expires_at,omitempty" db:"otp_expires_at"`
	Verified     bool       `json:"verified" db:"verified"`
	VerifiedAt   *time.Time `json:"verified_at,omitempty" db:"verified_at"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	IPAddress    string     `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent    string     `json:"user_agent,omitempty" db:"user_agent"`
}

// DecoyMessage represents a decoy message for secure links
type DecoyMessage struct {
	ID               string    `json:"id" db:"id"`
	LinkID           string    `json:"link_id" db:"link_id"`
	DecoyType        string    `json:"decoy_type" db:"decoy_type"`
	DecoyTitle       string    `json:"decoy_title" db:"decoy_title"`
	DecoyContent     string    `json:"decoy_content" db:"decoy_content"`
	TriggerCondition string    `json:"trigger_condition" db:"trigger_condition"`
	IsActive         bool      `json:"is_active" db:"is_active"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

// TamperAlert represents a tamper alert for suspicious activity
type TamperAlert struct {
	ID              string          `json:"id" db:"id"`
	LinkID          string          `json:"link_id" db:"link_id"`
	AlertType       string          `json:"alert_type" db:"alert_type"`
	Severity        string          `json:"severity" db:"severity"`
	IPAddress       string          `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent       string          `json:"user_agent,omitempty" db:"user_agent"`
	GeolocationData json.RawMessage `json:"geolocation_data,omitempty" db:"geolocation_data"`
	AlertDetails    string          `json:"alert_details,omitempty" db:"alert_details"`
	TriggeredAt     time.Time       `json:"triggered_at" db:"triggered_at"`
	Acknowledged    bool            `json:"acknowledged" db:"acknowledged"`
	AcknowledgedBy  string          `json:"acknowledged_by,omitempty" db:"acknowledged_by"`
	AcknowledgedAt  *time.Time      `json:"acknowledged_at,omitempty" db:"acknowledged_at"`
	Resolved        bool            `json:"resolved" db:"resolved"`
	ResolvedAt      *time.Time      `json:"resolved_at,omitempty" db:"resolved_at"`
	ResolutionNotes string          `json:"resolution_notes,omitempty" db:"resolution_notes"`
}

// AccessSession represents an individual access session for security monitoring
type AccessSession struct {
	ID                     string          `json:"id" db:"id"`
	LinkID                 string          `json:"link_id" db:"link_id"`
	SessionToken           string          `json:"session_token" db:"session_token"`
	IPAddress              string          `json:"ip_address" db:"ip_address"`
	UserAgent              string          `json:"user_agent,omitempty" db:"user_agent"`
	GeolocationData        json.RawMessage `json:"geolocation_data,omitempty" db:"geolocation_data"`
	SessionStart           time.Time       `json:"session_start" db:"session_start"`
	SessionEnd             *time.Time      `json:"session_end,omitempty" db:"session_end"`
	SessionDurationSeconds *int            `json:"session_duration_seconds,omitempty" db:"session_duration_seconds"`
	AccessGranted          bool            `json:"access_granted" db:"access_granted"`
	AccessReason           string          `json:"access_reason,omitempty" db:"access_reason"`
	SecurityChecksPassed   json.RawMessage `json:"security_checks_passed,omitempty" db:"security_checks_passed"`
	SecurityChecksFailed   json.RawMessage `json:"security_checks_failed,omitempty" db:"security_checks_failed"`
	MFASessionID           string          `json:"mfa_session_id,omitempty" db:"mfa_session_id"`
	PasswordAttemptID      string          `json:"password_attempt_id,omitempty" db:"password_attempt_id"`
}

// SecurityTemplate represents a predefined security configuration template
type SecurityTemplate struct {
	ID                  string          `json:"id" db:"id"`
	TemplateName        string          `json:"template_name" db:"template_name"`
	TemplateDescription string          `json:"template_description" db:"template_description"`
	SecuritySettings    json.RawMessage `json:"security_settings" db:"security_settings"`
	IsDefault           bool            `json:"is_default" db:"is_default"`
	CreatedBy           string          `json:"created_by,omitempty" db:"created_by"`
	CreatedAt           time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at" db:"updated_at"`
	IsActive            bool            `json:"is_active" db:"is_active"`
}

// =============================================================================
// ENUMS AND CONSTANTS
// =============================================================================

// MFAType represents the type of MFA
type MFAType string

const (
	MFATypeTOTP  MFAType = "totp"
	MFATypeEmail MFAType = "email"
	MFATypeSMS   MFAType = "sms"
	MFATypeNone  MFAType = "none"
)

// DecoyType represents the type of decoy message
type DecoyType string

const (
	DecoyTypeWrongPassword DecoyType = "wrong_password"
	DecoyTypeRevoked       DecoyType = "revoked"
	DecoyTypeExpired       DecoyType = "expired"
	DecoyTypeBlocked       DecoyType = "blocked"
	DecoyTypeGeneric       DecoyType = "generic"
)

// AlertType represents the type of tamper alert
type AlertType string

const (
	AlertTypeMultipleFailedAttempts AlertType = "multiple_failed_attempts"
	AlertTypeSuspiciousLocation     AlertType = "suspicious_location"
	AlertTypeUnusualTiming          AlertType = "unusual_timing"
	AlertTypeMultipleIPs            AlertType = "multiple_ips"
	AlertTypeUserAgentMismatch      AlertType = "user_agent_mismatch"
	AlertTypeRapidAccessAttempts    AlertType = "rapid_access_attempts"
	AlertTypeGeolocationViolation   AlertType = "geolocation_violation"
	AlertTypePasswordBruteForce     AlertType = "password_brute_force"
	AlertTypeSessionHijacking       AlertType = "session_hijacking"
)

// AlertSeverity represents the severity level of a tamper alert
type AlertSeverity string

const (
	AlertSeverityLow      AlertSeverity = "low"
	AlertSeverityMedium   AlertSeverity = "medium"
	AlertSeverityHigh     AlertSeverity = "high"
	AlertSeverityCritical AlertSeverity = "critical"
)

// =============================================================================
// REQUEST/RESPONSE MODELS
// =============================================================================

// PasswordValidationRequest represents a password validation request
type PasswordValidationRequest struct {
	LinkID    string `json:"link_id" validate:"required"`
	Password  string `json:"password" validate:"required"`
	IPAddress string `json:"ip_address"`
	UserAgent string `json:"user_agent,omitempty"`
}

// PasswordValidationResponse represents a password validation response
type PasswordValidationResponse struct {
	Valid         bool       `json:"valid"`
	AttemptNumber int        `json:"attempt_number"`
	MaxAttempts   int        `json:"max_attempts"`
	LockoutUntil  *time.Time `json:"lockout_until,omitempty"`
	LockedOut     bool       `json:"locked_out"`
	SessionToken  string     `json:"session_token,omitempty"`
	NextStep      string     `json:"next_step"` // password, mfa, access, locked
}

// GeolocationValidationRequest represents a geolocation validation request
type GeolocationValidationRequest struct {
	LinkID    string `json:"link_id" validate:"required"`
	IPAddress string `json:"ip_address" validate:"required"`
	UserAgent string `json:"user_agent,omitempty"`
}

// GeolocationValidationResponse represents a geolocation validation response
type GeolocationValidationResponse struct {
	Allowed           bool                   `json:"allowed"`
	Country           string                 `json:"country,omitempty"`
	City              string                 `json:"city,omitempty"`
	Region            string                 `json:"region,omitempty"`
	RestrictionReason string                 `json:"restriction_reason,omitempty"`
	GeolocationData   map[string]interface{} `json:"geolocation_data,omitempty"`
}

// MFAValidationRequest represents an MFA validation request
type MFAValidationRequest struct {
	LinkID       string `json:"link_id" validate:"required"`
	MFAType      string `json:"mfa_type" validate:"required"`
	OTPCode      string `json:"otp_code" validate:"required"`
	SessionToken string `json:"session_token" validate:"required"`
	IPAddress    string `json:"ip_address"`
	UserAgent    string `json:"user_agent,omitempty"`
}

// MFAValidationResponse represents an MFA validation response
type MFAValidationResponse struct {
	Valid        bool       `json:"valid"`
	Verified     bool       `json:"verified"`
	VerifiedAt   *time.Time `json:"verified_at,omitempty"`
	SessionToken string     `json:"session_token,omitempty"`
	NextStep     string     `json:"next_step"` // mfa, access, failed
}

// SecurityCheckRequest represents a comprehensive security check request
type SecurityCheckRequest struct {
	LinkID       string `json:"link_id" validate:"required"`
	IPAddress    string `json:"ip_address" validate:"required"`
	UserAgent    string `json:"user_agent,omitempty"`
	SessionToken string `json:"session_token,omitempty"`
	Password     string `json:"password,omitempty"`
	OTPCode      string `json:"otp_code,omitempty"`
}

// SecurityCheckResponse represents a comprehensive security check response
type SecurityCheckResponse struct {
	AccessGranted   bool                   `json:"access_granted"`
	AccessReason    string                 `json:"access_reason"`
	SecurityChecks  map[string]bool        `json:"security_checks"`
	FailedChecks    map[string]string      `json:"failed_checks"`
	NextStep        string                 `json:"next_step"`
	SessionToken    string                 `json:"session_token,omitempty"`
	DecoyMessage    *DecoyMessage          `json:"decoy_message,omitempty"`
	TamperAlerts    []TamperAlert          `json:"tamper_alerts,omitempty"`
	GeolocationData map[string]interface{} `json:"geolocation_data,omitempty"`
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

// IsValidMFAType checks if the MFA type is valid
func IsValidMFAType(mfaType string) bool {
	switch MFAType(mfaType) {
	case MFATypeTOTP, MFATypeEmail, MFATypeSMS, MFATypeNone:
		return true
	default:
		return false
	}
}

// IsValidDecoyType checks if the decoy type is valid
func IsValidDecoyType(decoyType string) bool {
	switch DecoyType(decoyType) {
	case DecoyTypeWrongPassword, DecoyTypeRevoked, DecoyTypeExpired, DecoyTypeBlocked, DecoyTypeGeneric:
		return true
	default:
		return false
	}
}

// IsValidAlertType checks if the alert type is valid
func IsValidAlertType(alertType string) bool {
	switch AlertType(alertType) {
	case AlertTypeMultipleFailedAttempts, AlertTypeSuspiciousLocation, AlertTypeUnusualTiming,
		AlertTypeMultipleIPs, AlertTypeUserAgentMismatch, AlertTypeRapidAccessAttempts,
		AlertTypeGeolocationViolation, AlertTypePasswordBruteForce, AlertTypeSessionHijacking:
		return true
	default:
		return false
	}
}

// IsValidAlertSeverity checks if the alert severity is valid
func IsValidAlertSeverity(severity string) bool {
	switch AlertSeverity(severity) {
	case AlertSeverityLow, AlertSeverityMedium, AlertSeverityHigh, AlertSeverityCritical:
		return true
	default:
		return false
	}
}

// GetDefaultSecuritySettings returns default security settings
func GetDefaultSecuritySettings() SecuritySettings {
	return SecuritySettings{
		PasswordRequired:          false,
		PasswordMaxAttempts:       3,
		TimeLockEnabled:           false,
		TimeLockTimezone:          "UTC",
		AutoDestructEnabled:       false,
		ReadOnceEnabled:           false,
		ReadOnceConsumed:          false,
		AutoDestructAfterAttempts: 5,
		GeoRestrictionEnabled:     false,
		GeoAllowedCountries:       []string{},
		GeoAllowedCities:          []string{},
		GeoBlockedCountries:       []string{},
		GeoBlockedCities:          []string{},
		MFARequired:               false,
		MFAType:                   string(MFATypeNone),
		DecoyEnabled:              false,
		StripMetadataEnabled:      true,
		StripHeaders:              true,
		StripAttachments:          false,
		TamperAlertsEnabled:       true,
		TamperAlertThreshold:      3,
	}
}
