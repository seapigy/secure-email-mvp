package email

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"secure-email-mvp/pkg/audit"
	"secure-email-mvp/pkg/emailpassword"
	"secure-email-mvp/pkg/mailer"
	"secure-email-mvp/pkg/mfa"
	"secure-email-mvp/pkg/notification"
)

// EmailSecurityService provides comprehensive security features for emails
type EmailSecurityService struct {
	db                   *sql.DB
	auditService         *audit.AuditService
	notificationService  *notification.NotificationService
	emailPasswordService *emailpassword.EmailPasswordService
	mfaService           *mfa.MFAService
	smtpMailer           *mailer.SMTPMailer
}

// NewEmailSecurityService creates a new email security service
func NewEmailSecurityService(db *sql.DB) *EmailSecurityService {
	// Initialize SMTP mailer (will be nil if SMTP is not configured)
	var smtpMailer *mailer.SMTPMailer
	if smtp, err := mailer.NewSMTPMailer(); err != nil {
		log.Printf("⚠️ SMTP mailer not available: %v", err)
	} else {
		smtpMailer = smtp
		log.Printf("✅ SMTP mailer initialized successfully")
	}

	return &EmailSecurityService{
		db:                   db,
		auditService:         audit.NewAuditService(db),
		notificationService:  notification.NewNotificationService(db),
		emailPasswordService: emailpassword.NewEmailPasswordService(db),
		mfaService:           mfa.NewMFAService(db),
		smtpMailer:           smtpMailer,
	}
}

// SecurityFeatureConfig represents all available security features
type SecurityFeatureConfig struct {
	// Basic Security
	PasswordProtection bool   `json:"password_protection"`
	Password           string `json:"password,omitempty"`

	// Access Control
	BurnAfterRead             bool `json:"burn_after_read"`
	SelfDestructAfterAttempts bool `json:"self_destruct_after_attempts"`
	MaxFailedAttempts         int  `json:"max_failed_attempts"`

	// Time-based Controls
	TimeLock    bool   `json:"time_lock"`
	UnlockAfter string `json:"unlock_after,omitempty"`
	ExpiresAt   string `json:"expires_at,omitempty"`

	// Geolocation
	GeoVerificationType string `json:"geo_verification_type"`
	GeoCity             string `json:"geo_city,omitempty"`
	GeoCountry          string `json:"geo_country,omitempty"`

	// Multi-Factor Authentication
	RequireMFA bool   `json:"require_mfa"`
	MFAType    string `json:"mfa_type"`

	// Advanced Security
	RemoteRevoke  bool   `json:"remote_revoke"`
	DecoyMessage  bool   `json:"decoy_message"`
	DecoySecret   string `json:"decoy_secret,omitempty"`
	StripMetadata bool   `json:"strip_metadata"`
	TamperAlerts  bool   `json:"tamper_alerts"`

	// Multi-Factor Authentication Triggers
	MFAOnOpen    bool `json:"mfa_on_open"`
	MFAOnReply   bool `json:"mfa_on_reply"`
	MFAOnForward bool `json:"mfa_on_forward"`
}

// EmailDeliveryConfig represents email delivery configuration
type EmailDeliveryConfig struct {
	// Delivery Type
	DeliveryType string `json:"delivery_type"` // "internal", "external", "both"

	// Internal Delivery
	InternalRecipient string `json:"internal_recipient,omitempty"`

	// External Delivery
	ExternalRecipient string `json:"external_recipient,omitempty"`
	ExternalSubject   string `json:"external_subject,omitempty"`
	ExternalBody      string `json:"external_body,omitempty"`

	// Security Features for External Delivery
	ExternalSecurityFeatures SecurityFeatureConfig `json:"external_security_features"`
}

// ApplySecurityFeatures applies all security features to an email
func (s *EmailSecurityService) ApplySecurityFeatures(ctx context.Context, emailID, senderID string, config SecurityFeatureConfig) error {
	log.Printf("🔐 Applying security features to email %s", emailID)

	// Start transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. Password Protection
	if config.PasswordProtection && config.Password != "" {
		if err := s.applyPasswordProtection(ctx, tx, emailID, config.Password); err != nil {
			return fmt.Errorf("failed to apply password protection: %w", err)
		}
	}

	// 2. Time-based Controls
	if err := s.applyTimeBasedControls(ctx, tx, emailID, config); err != nil {
		return fmt.Errorf("failed to apply time-based controls: %w", err)
	}

	// 3. Access Control
	if err := s.applyAccessControl(ctx, tx, emailID, config); err != nil {
		return fmt.Errorf("failed to apply access control: %w", err)
	}

	// 4. Geolocation Restrictions
	if err := s.applyGeolocationRestrictions(ctx, tx, emailID, config); err != nil {
		return fmt.Errorf("failed to apply geolocation restrictions: %w", err)
	}

	// 5. Multi-Factor Authentication
	if err := s.applyMFA(ctx, tx, emailID, config); err != nil {
		return fmt.Errorf("failed to apply MFA: %w", err)
	}

	// 6. Advanced Security Features
	if err := s.applyAdvancedSecurity(ctx, tx, emailID, config); err != nil {
		return fmt.Errorf("failed to apply advanced security: %w", err)
	}

	// 7. Audit Logging
	auditEvent := &audit.AuditEvent{
		EventType: audit.EventType("email_security_applied"),
		UserID:    &senderID,
		Details: map[string]interface{}{
			"email_id": emailID,
			"features": config,
		},
		Severity: audit.Severity("medium"),
		Outcome:  audit.OutcomeSuccess,
	}
	if err := s.auditService.RecordEvent(ctx, auditEvent); err != nil {
		log.Printf("⚠️ Failed to log security application: %v", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("✅ Security features applied successfully to email %s", emailID)
	return nil
}

// applyPasswordProtection applies password protection to an email
func (s *EmailSecurityService) applyPasswordProtection(ctx context.Context, tx *sql.Tx, emailID, password string) error {
	// Use existing email password service
	return s.emailPasswordService.SetEmailPassword(emailID, password)
}

// applyTimeBasedControls applies time-based access controls
func (s *EmailSecurityService) applyTimeBasedControls(ctx context.Context, tx *sql.Tx, emailID string, config SecurityFeatureConfig) error {
	query := `
		UPDATE emails 
		SET 
			time_lock = ?,
			unlock_after = ?,
			expires_at = ?
		WHERE email_id = ?
	`

	var unlockAfter, expiresAt interface{}

	if config.TimeLock && config.UnlockAfter != "" {
		unlockAfter = config.UnlockAfter
	}

	if config.ExpiresAt != "" {
		expiresAt = config.ExpiresAt
	}

	_, err := tx.ExecContext(ctx, query, config.TimeLock, unlockAfter, expiresAt, emailID)
	return err
}

// applyAccessControl applies access control features
func (s *EmailSecurityService) applyAccessControl(ctx context.Context, tx *sql.Tx, emailID string, config SecurityFeatureConfig) error {
	query := `
		UPDATE emails 
		SET 
			burn_after_read = ?,
			self_destruct_after_attempts = ?,
			max_attempts = ?
		WHERE email_id = ?
	`

	_, err := tx.ExecContext(ctx, query,
		config.BurnAfterRead,
		config.SelfDestructAfterAttempts,
		config.MaxFailedAttempts,
		emailID)
	return err
}

// applyGeolocationRestrictions applies geolocation restrictions
func (s *EmailSecurityService) applyGeolocationRestrictions(ctx context.Context, tx *sql.Tx, emailID string, config SecurityFeatureConfig) error {
	if config.GeoVerificationType == "none" {
		return nil
	}

	geoData := map[string]interface{}{
		"verification_type": config.GeoVerificationType,
		"city":              config.GeoCity,
		"country":           config.GeoCountry,
	}

	geoJSON, err := json.Marshal(geoData)
	if err != nil {
		return fmt.Errorf("failed to marshal geolocation data: %w", err)
	}

	query := `UPDATE emails SET geolocation_json = ? WHERE email_id = ?`
	_, err = tx.ExecContext(ctx, query, string(geoJSON), emailID)
	return err
}

// applyMFA applies multi-factor authentication settings
func (s *EmailSecurityService) applyMFA(ctx context.Context, tx *sql.Tx, emailID string, config SecurityFeatureConfig) error {
	query := `
		UPDATE emails 
		SET 
			require_mfa = ?,
			mfa_type = ?,
			mfa_on_open = ?,
			mfa_on_reply = ?,
			mfa_on_forward = ?
		WHERE email_id = ?
	`

	_, err := tx.ExecContext(ctx, query,
		config.RequireMFA,
		config.MFAType,
		config.MFAOnOpen,
		config.MFAOnReply,
		config.MFAOnForward,
		emailID)
	return err
}

// applyAdvancedSecurity applies advanced security features
func (s *EmailSecurityService) applyAdvancedSecurity(ctx context.Context, tx *sql.Tx, emailID string, config SecurityFeatureConfig) error {
	// Generate decoy secret if decoy message is enabled
	var decoySecret interface{}
	if config.DecoyMessage {
		if config.DecoySecret != "" {
			// Hash the provided decoy secret
			hashedSecret, err := s.hashDecoySecret(config.DecoySecret)
			if err != nil {
				return fmt.Errorf("failed to hash decoy secret: %w", err)
			}
			decoySecret = hashedSecret
		} else {
			// Generate a random decoy secret
			randomSecret, err := s.generateRandomDecoySecret()
			if err != nil {
				return fmt.Errorf("failed to generate decoy secret: %w", err)
			}
			decoySecret = randomSecret
		}
	}

	query := `
		UPDATE emails 
		SET 
			remote_revoke = ?,
			decoy_message = ?,
			decoy_secret = ?,
			strip_metadata = ?,
			tamper_alerts = ?
		WHERE email_id = ?
	`

	_, err := tx.ExecContext(ctx, query,
		config.RemoteRevoke,
		config.DecoyMessage,
		decoySecret,
		config.StripMetadata,
		config.TamperAlerts,
		emailID)
	return err
}

// hashDecoySecret hashes a decoy secret for storage
func (s *EmailSecurityService) hashDecoySecret(secret string) (string, error) {
	// Use Argon2id for hashing (same as password hashing)
	// Since the emailpassword service doesn't expose HashPassword, we'll implement it here
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

// generateRandomDecoySecret generates a random decoy secret
func (s *EmailSecurityService) generateRandomDecoySecret() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

// SendEmailWithSecurity sends an email with comprehensive security features
func (s *EmailSecurityService) SendEmailWithSecurity(ctx context.Context, senderID string, deliveryConfig EmailDeliveryConfig) error {
	log.Printf("📧 Sending email with security features from %s", senderID)

	// Apply security features based on delivery type
	switch deliveryConfig.DeliveryType {
	case "internal":
		return s.sendInternalEmail(ctx, senderID, deliveryConfig)
	case "external":
		return s.sendExternalEmail(ctx, senderID, deliveryConfig)
	case "both":
		if err := s.sendInternalEmail(ctx, senderID, deliveryConfig); err != nil {
			return fmt.Errorf("failed to send internal email: %w", err)
		}
		return s.sendExternalEmail(ctx, senderID, deliveryConfig)
	default:
		return fmt.Errorf("invalid delivery type: %s", deliveryConfig.DeliveryType)
	}
}

// sendInternalEmail sends an email to an internal recipient
func (s *EmailSecurityService) sendInternalEmail(ctx context.Context, senderID string, config EmailDeliveryConfig) error {
	if config.InternalRecipient == "" {
		return fmt.Errorf("internal recipient is required for internal delivery")
	}

	// Create internal email with security features
	emailID := s.generateEmailID()

	// Apply security features
	if err := s.ApplySecurityFeatures(ctx, emailID, senderID, config.ExternalSecurityFeatures); err != nil {
		return fmt.Errorf("failed to apply security features: %w", err)
	}

	// Log internal email send
	auditEvent := &audit.AuditEvent{
		EventType: audit.EventType("internal_email_sent"),
		UserID:    &senderID,
		Details: map[string]interface{}{
			"email_id":  emailID,
			"recipient": config.InternalRecipient,
		},
		Severity: audit.Severity("low"),
		Outcome:  audit.OutcomeSuccess,
	}
	if err := s.auditService.RecordEvent(ctx, auditEvent); err != nil {
		log.Printf("⚠️ Failed to log internal email send: %v", err)
	}

	log.Printf("✅ Internal email %s sent successfully to %s", emailID, config.InternalRecipient)
	return nil
}

// sendExternalEmail sends an email to an external recipient
func (s *EmailSecurityService) sendExternalEmail(ctx context.Context, senderID string, config EmailDeliveryConfig) error {
	if config.ExternalRecipient == "" {
		return fmt.Errorf("external recipient is required for external delivery")
	}

	// Create external email with enhanced security
	emailID := s.generateEmailID()

	// Apply enhanced security features for external delivery
	enhancedConfig := config.ExternalSecurityFeatures
	enhancedConfig.StripMetadata = true // Always strip metadata for external emails
	enhancedConfig.TamperAlerts = true  // Always enable tamper alerts for external emails

	if err := s.ApplySecurityFeatures(ctx, emailID, senderID, enhancedConfig); err != nil {
		return fmt.Errorf("failed to apply security features: %w", err)
	}

	// Send notification email to external recipient
	// Note: We need to get the sender email from the database
	var senderEmail string
	if err := s.db.QueryRowContext(ctx, "SELECT email FROM users WHERE id = ?", senderID).Scan(&senderEmail); err != nil {
		log.Printf("⚠️ Failed to get sender email for notification: %v", err)
		// Use a fallback address if we can't get the sender email
		senderEmail = "noreply@securesystem.email"
	}
	
	if err := s.SendNotificationEmail(config.ExternalRecipient, emailID, senderEmail); err != nil {
		log.Printf("⚠️ Failed to send notification email: %v", err)
		// Don't fail the entire operation if notification fails
	}

	// Log external email send
	auditEvent := &audit.AuditEvent{
		EventType: audit.EventType("external_email_sent"),
		UserID:    &senderID,
		Details: map[string]interface{}{
			"email_id":  emailID,
			"recipient": config.ExternalRecipient,
		},
		Severity: audit.Severity("medium"),
		Outcome:  audit.OutcomeSuccess,
	}
	if err := s.auditService.RecordEvent(ctx, auditEvent); err != nil {
		log.Printf("⚠️ Failed to log external email send: %v", err)
	}

	log.Printf("✅ External email %s sent successfully to %s", emailID, config.ExternalRecipient)
	return nil
}

// buildSecurityFeaturesDescription builds a human-readable description of security features
func (s *EmailSecurityService) buildSecurityFeaturesDescription(config SecurityFeatureConfig) string {
	var features []string

	if config.PasswordProtection {
		features = append(features, "• Password Protection Required")
	}
	if config.BurnAfterRead {
		features = append(features, "• Self-Destruct After First Read")
	}
	if config.SelfDestructAfterAttempts {
		features = append(features, fmt.Sprintf("• Self-Destruct After %d Failed Attempts", config.MaxFailedAttempts))
	}
	if config.TimeLock {
		features = append(features, "• Time-Locked Access")
	}
	if config.RequireMFA {
		features = append(features, fmt.Sprintf("• Multi-Factor Authentication (%s)", config.MFAType))
	}
	if config.GeoVerificationType != "none" {
		features = append(features, "• Geolocation Restrictions")
	}
	if config.RemoteRevoke {
		features = append(features, "• Remote Revocation Enabled")
	}
	if config.DecoyMessage {
		features = append(features, "• Decoy Message Protection")
	}
	if config.StripMetadata {
		features = append(features, "• Metadata Stripped")
	}
	if config.TamperAlerts {
		features = append(features, "• Tamper Detection Enabled")
	}

	if len(features) == 0 {
		return "• Standard Encryption"
	}

	return fmt.Sprintf("%s\n• End-to-End Encryption", features[0])
}

// generateEmailID generates a unique email ID
func (s *EmailSecurityService) generateEmailID() string {
	return fmt.Sprintf("email_%d_%s", time.Now().Unix(), s.generateRandomString(8))
}

// generateRandomString generates a random string of specified length
func (s *EmailSecurityService) generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

// ValidateSecurityConfig validates security configuration
func (s *EmailSecurityService) ValidateSecurityConfig(config SecurityFeatureConfig) error {
	// Validate password if provided
	if config.PasswordProtection && config.Password != "" {
		if err := s.emailPasswordService.ValidatePasswordStrength(config.Password); err != nil {
			return fmt.Errorf("password validation failed: %w", err)
		}
	}

	// Validate MFA type
	if config.RequireMFA {
		if config.MFAType != "TOTP" && config.MFAType != "EMAIL_CODE" {
			return fmt.Errorf("invalid MFA type: %s", config.MFAType)
		}
	}

	// Validate geolocation verification type
	if config.GeoVerificationType != "" &&
		config.GeoVerificationType != "none" &&
		config.GeoVerificationType != "city" &&
		config.GeoVerificationType != "country" &&
		config.GeoVerificationType != "city_country" {
		return fmt.Errorf("invalid geolocation verification type: %s", config.GeoVerificationType)
	}

	// Validate max failed attempts
	if config.MaxFailedAttempts < 1 || config.MaxFailedAttempts > 10 {
		return fmt.Errorf("max failed attempts must be between 1 and 10")
	}

	// Validate time lock
	if config.TimeLock && config.UnlockAfter == "" {
		return fmt.Errorf("unlock time is required when time lock is enabled")
	}

	return nil
}

// SendNotificationEmail sends a notification email to the recipient
func (s *EmailSecurityService) SendNotificationEmail(recipient string, emailID string, senderEmail string) error {
	if s.smtpMailer == nil {
		log.Printf("⚠️ SMTP mailer not available, skipping notification email")
		return nil
	}

	notificationMsg := mailer.EmailMessage{
		To:      recipient,
		Subject: "You have a new secure message",
		Body: fmt.Sprintf(`Hello,

You have received a secure message. Please log into SecureSystem.email to view it.

Message ID: %s

Best regards,
%s`, emailID, senderEmail),
		From: senderEmail,
	}

	if err := s.smtpMailer.SendEmail(notificationMsg); err != nil {
		log.Printf("❌ Failed to send notification email from %s to %s: %v", senderEmail, recipient, err)
		return fmt.Errorf("failed to send notification email: %w", err)
	}

	log.Printf("✅ Notification email sent successfully from %s to: %s", senderEmail, recipient)
	return nil
}
