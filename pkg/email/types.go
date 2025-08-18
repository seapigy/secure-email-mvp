package email

import (
	"encoding/json"
	"fmt"
	"time"
)

// EmailSecurityToggles represents the per-email security settings
//
// This struct defines all the security features that can be configured
// on a per-email basis by the sender. Each email can have its own
// individualized security rules, providing fine-grained control over
// access, timing, and protection mechanisms.
//
// SECURITY FEATURES:
// - Time Lock: Control when emails become accessible
// - Expiration: Set automatic expiration dates
// - Read-Once: Burn-after-read functionality
// - MFA-on-Open: Require additional authentication for access
// - Decoy Messages: Plausible deniability for invalid access
// - Remote Revoke: Allow senders to revoke access anytime
// - Metadata Stripping: Remove EXIF data from attachments
// - Self-Destruct: Automatic deletion after failed attempts
// - Geofencing: Location-based access restrictions
// - Self-Destruct on Read: Delete after first successful access
//
// IMPLEMENTATION NOTES:
// - All fields are optional to support partial updates
// - Timestamps are stored as Unix timestamps (int64) for consistency
// - Boolean fields default to false for security
// - Threshold values have reasonable limits (1-100 for self-destruct)
// - Decoy secrets are stored as Argon2id hashes for security
// - Geo rules are stored as JSON strings for flexibility
//
// VALIDATION:
// - Time windows must be logical (not_before < expires_at)
// - Self-destruct threshold must be between 1 and 100
// - Geo rules must be valid JSON if provided
// - Decoy secrets cannot be empty if provided
type EmailSecurityToggles struct {
	// Time Lock: Unix timestamp for when email becomes accessible
	// Email access denied before this timestamp
	NotBefore *int64 `json:"not_before,omitempty"`

	// Expiration: Unix timestamp for when email expires
	// Email access denied after this timestamp
	ExpiresAt *int64 `json:"expires_at,omitempty"`

	// Read Once: Burn after first successful access
	// Email is marked as consumed after first read
	ReadOnce bool `json:"read_once,omitempty"`

	// MFA on Open: Require secondary TOTP for access
	// Separate from account MFA, provides additional layer of protection
	MFAOnOpen bool `json:"mfa_on_open,omitempty"`

	// Decoy Secret: Argon2id hash of decoy password or TOTP
	// Enables plausible deniability features
	DecoySecret *string `json:"decoy_secret,omitempty"`

	// Remote Revoke: Sender can revoke access anytime
	// Immediate access denial when enabled
	RemoteRevoke bool `json:"remote_revoke,omitempty"`

	// Strip Metadata: Remove EXIF data and headers from attachments
	// Privacy protection for file metadata
	StripMetadata bool `json:"strip_metadata,omitempty"`

	// Self Destruct Threshold: Max failed attempts before destruction
	// Configurable per email (minimum 1), provides brute force protection
	SelfDestructThreshold *int `json:"self_destruct_threshold,omitempty"`

	// Geo Rules Reference: JSON reference to geofencing rules
	// Stored as JSON string for flexibility, not parsed in this iteration
	GeoRulesRef *string `json:"geo_rules_ref,omitempty"`

	// Self Destruct on Read Once: If true, delete email immediately after first read
	// Provides additional security by removing all traces after consumption
	SelfDestructOnReadOnce bool `json:"self_destruct_on_read_once,omitempty"`
}

// EmailSecurityRequest represents the request structure for updating email security settings
type EmailSecurityRequest struct {
	// Email ID to update security settings for
	EmailID string `json:"email_id"`

	// Security toggles to apply
	Toggles EmailSecurityToggles `json:"toggles"`
}

// EmailSecurityResponse represents the response structure for security operations
type EmailSecurityResponse struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// EmailSecurityInfo represents the current security settings for an email
type EmailSecurityInfo struct {
	EmailID string               `json:"email_id"`
	Toggles EmailSecurityToggles `json:"toggles"`
}

// EmailAccessResult represents the result of an email access attempt
// Used for tracking failed attempts and enforcing self-destruct logic
type EmailAccessResult struct {
	Success bool
	Reason  string // For logging purposes only, not exposed to client
}

// SelfDestructError is returned when an email has been destroyed due to too many failed attempts
type SelfDestructError struct {
	EmailID string
}

func (e SelfDestructError) Error() string {
	return fmt.Sprintf("email %s has been destroyed due to too many failed attempts", e.EmailID)
}

// ReadOnceConsumedError is returned when trying to mark an already consumed read-once email
type ReadOnceConsumedError struct {
	EmailID string
}

func (e ReadOnceConsumedError) Error() string {
	return fmt.Sprintf("email %s has already been consumed", e.EmailID)
}

// ReadOnceInfo contains information about read-once consumption status
type ReadOnceInfo struct {
	IsConsumed         bool      `json:"is_consumed"`
	ConsumedAt         time.Time `json:"consumed_at,omitempty"`
	ConsumerDevice     string    `json:"consumer_device,omitempty"`
	SelfDestructOnRead bool      `json:"self_destruct_on_read"`
}

// ValidateEmailSecurityToggles validates the security toggle settings
// Returns an error if any validation rules are violated
func ValidateEmailSecurityToggles(toggles EmailSecurityToggles) error {
	// Validate time window constraints
	if toggles.NotBefore != nil && toggles.ExpiresAt != nil {
		if *toggles.NotBefore >= *toggles.ExpiresAt {
			return fmt.Errorf("not_before must be before expires_at")
		}
	}

	// Validate self-destruct threshold
	if toggles.SelfDestructThreshold != nil {
		if *toggles.SelfDestructThreshold < 1 {
			return fmt.Errorf("self_destruct_threshold must be at least 1")
		}
		if *toggles.SelfDestructThreshold > 100 {
			return fmt.Errorf("self_destruct_threshold cannot exceed 100")
		}
	}

	// Validate geo rules reference is valid JSON (if provided)
	if toggles.GeoRulesRef != nil && *toggles.GeoRulesRef != "" {
		var testJSON interface{}
		if err := json.Unmarshal([]byte(*toggles.GeoRulesRef), &testJSON); err != nil {
			return fmt.Errorf("geo_rules_ref must be valid JSON: %w", err)
		}
	}

	// Validate decoy secret is not empty if provided
	if toggles.DecoySecret != nil && *toggles.DecoySecret == "" {
		return fmt.Errorf("decoy_secret cannot be empty if provided")
	}

	return nil
}

// IsTimeLocked checks if the email is currently time-locked
// Returns true if the current time is before the not_before timestamp
func (toggles EmailSecurityToggles) IsTimeLocked() bool {
	if toggles.NotBefore == nil {
		return false
	}
	return time.Now().Unix() < *toggles.NotBefore
}

// IsExpired checks if the email has expired
// Returns true if the current time is after the expires_at timestamp
func (toggles EmailSecurityToggles) IsExpired() bool {
	if toggles.ExpiresAt == nil {
		return false
	}
	return time.Now().Unix() > *toggles.ExpiresAt
}

// IsRevoked checks if the email has been remotely revoked by the sender
func (toggles EmailSecurityToggles) IsRevoked() bool {
	return toggles.RemoteRevoke
}

// GetSelfDestructThreshold returns the configured self-destruct threshold
// Returns the default value of 3 if not configured
func (toggles EmailSecurityToggles) GetSelfDestructThreshold() int {
	if toggles.SelfDestructThreshold == nil {
		return 3 // Default value
	}
	return *toggles.SelfDestructThreshold
}

// RequiresMFA checks if the email requires MFA for access
func (toggles EmailSecurityToggles) RequiresMFA() bool {
	return toggles.MFAOnOpen
}

// IsReadOnce checks if the email should be burned after first access
func (toggles EmailSecurityToggles) IsReadOnce() bool {
	return toggles.ReadOnce
}

// ShouldStripMetadata checks if metadata should be stripped from attachments
func (toggles EmailSecurityToggles) ShouldStripMetadata() bool {
	return toggles.StripMetadata
}

// HasDecoySecret checks if a decoy secret is configured
func (toggles EmailSecurityToggles) HasDecoySecret() bool {
	return toggles.DecoySecret != nil && *toggles.DecoySecret != ""
}

// HasGeoRules checks if geofencing rules are configured
func (toggles EmailSecurityToggles) HasGeoRules() bool {
	return toggles.GeoRulesRef != nil && *toggles.GeoRulesRef != ""
}

// ShouldSelfDestructOnReadOnce checks if the email should be deleted after first read
func (toggles EmailSecurityToggles) ShouldSelfDestructOnReadOnce() bool {
	return toggles.SelfDestructOnReadOnce
}

// GetTimeWindowStatus returns a human-readable status of the time window
func (toggles EmailSecurityToggles) GetTimeWindowStatus() string {
	now := time.Now().Unix()

	if toggles.NotBefore != nil && now < *toggles.NotBefore {
		timeUntil := time.Until(time.Unix(*toggles.NotBefore, 0))
		return fmt.Sprintf("Time-locked until %s (%s from now)",
			time.Unix(*toggles.NotBefore, 0).Format(time.RFC3339),
			timeUntil.Round(time.Minute))
	}

	if toggles.ExpiresAt != nil && now > *toggles.ExpiresAt {
		return fmt.Sprintf("Expired at %s",
			time.Unix(*toggles.ExpiresAt, 0).Format(time.RFC3339))
	}

	if toggles.NotBefore != nil && toggles.ExpiresAt != nil {
		return fmt.Sprintf("Available from %s to %s",
			time.Unix(*toggles.NotBefore, 0).Format(time.RFC3339),
			time.Unix(*toggles.ExpiresAt, 0).Format(time.RFC3339))
	}

	if toggles.NotBefore != nil {
		return fmt.Sprintf("Available from %s",
			time.Unix(*toggles.NotBefore, 0).Format(time.RFC3339))
	}

	if toggles.ExpiresAt != nil {
		return fmt.Sprintf("Expires at %s",
			time.Unix(*toggles.ExpiresAt, 0).Format(time.RFC3339))
	}

	return "No time restrictions"
}
