package email

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"os"
	"strings"
	"time"

	"secure-email-mvp/pkg/email"
)

// =============================================================================
// SECURE LINK EMAIL SERVICE
// =============================================================================

// SESHandlerInterface defines the interface for SES operations
type SESHandlerInterface interface {
	SendEmailViaSES(ctx context.Context, emailID, senderID, recipient, subject, body string) (*email.SESTransaction, error)
}

// Service provides email delivery functionality for secure links
type Service struct {
	db         *sql.DB
	sesHandler SESHandlerInterface
	baseURL    string
}

// SecureLinkEmailRequest represents a request to send a secure link email
type SecureLinkEmailRequest struct {
	LinkID          string          `json:"link_id" validate:"required"`
	RecipientEmail  string          `json:"recipient_email" validate:"required,email"`
	SenderName      string          `json:"sender_name"`
	SenderEmail     string          `json:"sender_email"`
	SecurityContext SecurityContext `json:"security_context"`
	CustomMessage   *string         `json:"custom_message,omitempty"`
	LinkExpiresAt   *time.Time      `json:"link_expires_at,omitempty"`
}

// SecurityContext provides security information for the email template
type SecurityContext struct {
	RequirePassword        bool       `json:"require_password"`
	RequireMFA             bool       `json:"require_mfa"`
	MFAType                string     `json:"mfa_type,omitempty"`
	GeolocationRestriction bool       `json:"geolocation_restriction"`
	AllowedCountries       []string   `json:"allowed_countries,omitempty"`
	AllowedCities          []string   `json:"allowed_cities,omitempty"`
	TimeLock               bool       `json:"time_lock"`
	TimeLockUntil          *time.Time `json:"time_lock_until,omitempty"`
	ReadOnce               bool       `json:"read_once"`
	AutoDestruct           bool       `json:"auto_destruct"`
	ExpiresAt              *time.Time `json:"expires_at,omitempty"`
}

// SecureLinkEmailResponse represents the response from sending a secure link email
type SecureLinkEmailResponse struct {
	Success       bool   `json:"success"`
	TransactionID string `json:"transaction_id,omitempty"`
	MessageID     string `json:"message_id,omitempty"`
	Error         string `json:"error,omitempty"`
}

// NewService creates a new secure link email service
func NewService(db *sql.DB, sesHandler SESHandlerInterface, baseURL string) *Service {
	return &Service{
		db:         db,
		sesHandler: sesHandler,
		baseURL:    baseURL,
	}
}

// SendSecureLinkEmail sends a secure link notification email to an external recipient
func (s *Service) SendSecureLinkEmail(ctx context.Context, req SecureLinkEmailRequest) (*SecureLinkEmailResponse, error) {
	// Validate request
	if err := s.validateRequest(req); err != nil {
		return &SecureLinkEmailResponse{
			Success: false,
			Error:   fmt.Sprintf("Invalid request: %v", err),
		}, err
	}

	// Generate email subject and body
	subject, body, err := s.generateEmailContent(req)
	if err != nil {
		return &SecureLinkEmailResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to generate email content: %v", err),
		}, err
	}

	// Send email via SES
	transaction, err := s.sesHandler.SendEmailViaSES(ctx, req.LinkID, req.SenderEmail, req.RecipientEmail, subject, body)
	if err != nil {
		return &SecureLinkEmailResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to send email: %v", err),
		}, err
	}

	// Log the secure link email event
	if err := s.logSecureLinkEmail(ctx, req, transaction); err != nil {
		log.Printf("⚠️ Failed to log secure link email event: %v", err)
		// Don't fail the operation for logging errors
	}

	return &SecureLinkEmailResponse{
		Success:       true,
		TransactionID: transaction.TransactionID,
		MessageID:     transaction.MessageID,
	}, nil
}

// validateRequest validates the secure link email request
func (s *Service) validateRequest(req SecureLinkEmailRequest) error {
	if req.LinkID == "" {
		return fmt.Errorf("link_id is required")
	}
	if req.RecipientEmail == "" {
		return fmt.Errorf("recipient_email is required")
	}
	if !strings.Contains(req.RecipientEmail, "@") {
		return fmt.Errorf("invalid recipient_email format")
	}
	if req.SenderEmail == "" {
		return fmt.Errorf("sender_email is required")
	}
	return nil
}

// generateEmailContent generates the subject and body for the secure link email
func (s *Service) generateEmailContent(req SecureLinkEmailRequest) (string, string, error) {
	// Generate subject
	subject := s.generateSubject(req)

	// Generate body using template
	body, err := s.generateBody(req)
	if err != nil {
		return "", "", err
	}

	return subject, body, nil
}

// generateSubject generates the email subject
func (s *Service) generateSubject(req SecureLinkEmailRequest) string {
	if req.SenderName != "" {
		return fmt.Sprintf("Secure Message from %s", req.SenderName)
	}
	return "Secure Message Received"
}

// generateBody generates the email body using the template
func (s *Service) generateBody(req SecureLinkEmailRequest) (string, error) {
	// Load template
	tmpl, err := s.loadTemplate()
	if err != nil {
		return "", err
	}

	// Prepare template data
	data := s.prepareTemplateData(req)

	// Execute template
	var body strings.Builder
	if err := tmpl.Execute(&body, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return body.String(), nil
}

// loadTemplate loads the secure link email template
func (s *Service) loadTemplate() (*template.Template, error) {
	// Try to load from templates directory
	templatePath := "templates/secure_link_notification.html"

	// Check if template file exists
	if _, err := os.Stat(templatePath); err == nil {
		return template.ParseFiles(templatePath)
	}

	// Fallback to embedded template
	return template.New("secure_link_notification").Parse(defaultTemplate)
}

// prepareTemplateData prepares the data for the email template
func (s *Service) prepareTemplateData(req SecureLinkEmailRequest) map[string]interface{} {
	// Build secure link URL
	secureURL := fmt.Sprintf("%s/v/%s", s.baseURL, req.LinkID)

	// Prepare security features description
	securityFeatures := s.buildSecurityFeaturesDescription(req.SecurityContext)

	// Determine expiration info
	var expirationInfo string
	if req.LinkExpiresAt != nil {
		expirationInfo = fmt.Sprintf("This secure link expires on %s.", req.LinkExpiresAt.Format("January 2, 2006 at 3:04 PM"))
	} else if req.SecurityContext.ExpiresAt != nil {
		expirationInfo = fmt.Sprintf("This secure link expires on %s.", req.SecurityContext.ExpiresAt.Format("January 2, 2006 at 3:04 PM"))
	} else {
		expirationInfo = "This secure link will expire automatically for security."
	}

	return map[string]interface{}{
		"SenderName":             req.SenderName,
		"SenderEmail":            req.SenderEmail,
		"SecureURL":              secureURL,
		"LinkID":                 req.LinkID,
		"CustomMessage":          req.CustomMessage,
		"SecurityFeatures":       securityFeatures,
		"ExpirationInfo":         expirationInfo,
		"RequirePassword":        req.SecurityContext.RequirePassword,
		"RequireMFA":             req.SecurityContext.RequireMFA,
		"MFAType":                req.SecurityContext.MFAType,
		"GeolocationRestriction": req.SecurityContext.GeolocationRestriction,
		"AllowedCountries":       req.SecurityContext.AllowedCountries,
		"AllowedCities":          req.SecurityContext.AllowedCities,
		"TimeLock":               req.SecurityContext.TimeLock,
		"TimeLockUntil":          req.SecurityContext.TimeLockUntil,
		"ReadOnce":               req.SecurityContext.ReadOnce,
		"AutoDestruct":           req.SecurityContext.AutoDestruct,
		"CurrentTime":            time.Now().Format("January 2, 2006 at 3:04 PM"),
	}
}

// buildSecurityFeaturesDescription builds a human-readable description of security features
func (s *Service) buildSecurityFeaturesDescription(ctx SecurityContext) string {
	var features []string

	if ctx.RequirePassword {
		features = append(features, "password protection")
	}
	if ctx.RequireMFA {
		features = append(features, fmt.Sprintf("%s verification", strings.ToLower(ctx.MFAType)))
	}
	if ctx.GeolocationRestriction {
		if len(ctx.AllowedCountries) > 0 {
			features = append(features, "location restrictions")
		}
	}
	if ctx.TimeLock {
		features = append(features, "time-based access control")
	}
	if ctx.ReadOnce {
		features = append(features, "one-time viewing")
	}
	if ctx.AutoDestruct {
		features = append(features, "auto-destruct protection")
	}

	if len(features) == 0 {
		return "standard security measures"
	}

	if len(features) == 1 {
		return features[0]
	}

	// Join features with commas and "and"
	if len(features) == 2 {
		return fmt.Sprintf("%s and %s", features[0], features[1])
	}

	last := features[len(features)-1]
	others := features[:len(features)-1]
	return fmt.Sprintf("%s, and %s", strings.Join(others, ", "), last)
}

// logSecureLinkEmail logs the secure link email event for audit purposes
func (s *Service) logSecureLinkEmail(ctx context.Context, req SecureLinkEmailRequest, transaction *email.SESTransaction) error {
	query := `
		INSERT INTO link_audit_log (
			id, link_id, event_type, timestamp, ip_address, user_agent,
			details, ses_transaction_id, recipient_email
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	auditID := fmt.Sprintf("email_%s_%d", req.LinkID, time.Now().Unix())
	details := fmt.Sprintf(`{
		"sender_name": "%s",
		"sender_email": "%s",
		"security_context": %s,
		"custom_message": %s
	}`,
		req.SenderName,
		req.SenderEmail,
		s.securityContextToJSON(req.SecurityContext),
		s.stringToJSON(req.CustomMessage),
	)

	_, err := s.db.ExecContext(ctx, query,
		auditID,
		req.LinkID,
		"secure_link_email_sent",
		time.Now(),
		"system", // IP address for system-generated emails
		"secure-link-email-service",
		details,
		transaction.TransactionID,
		req.RecipientEmail,
	)

	return err
}

// securityContextToJSON converts security context to JSON string
func (s *Service) securityContextToJSON(ctx SecurityContext) string {
	// This is a simplified JSON conversion - in production, use proper JSON marshaling
	return fmt.Sprintf(`{
		"require_password": %t,
		"require_mfa": %t,
		"mfa_type": "%s",
		"geolocation_restriction": %t,
		"time_lock": %t,
		"read_once": %t,
		"auto_destruct": %t
	}`,
		ctx.RequirePassword,
		ctx.RequireMFA,
		ctx.MFAType,
		ctx.GeolocationRestriction,
		ctx.TimeLock,
		ctx.ReadOnce,
		ctx.AutoDestruct,
	)
}

// stringToJSON converts a string pointer to JSON string
func (s *Service) stringToJSON(sptr *string) string {
	if sptr == nil {
		return "null"
	}
	return fmt.Sprintf(`"%s"`, *sptr)
}

// defaultTemplate is the fallback template if no template file is found
const defaultTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Secure Message</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #f8f9fa; padding: 20px; border-radius: 8px; margin-bottom: 20px; }
        .content { background: #ffffff; padding: 20px; border: 1px solid #e9ecef; border-radius: 8px; }
        .button { display: inline-block; background: #007bff; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; margin: 20px 0; }
        .security-info { background: #e7f3ff; padding: 15px; border-radius: 6px; margin: 20px 0; border-left: 4px solid #007bff; }
        .footer { margin-top: 30px; padding-top: 20px; border-top: 1px solid #e9ecef; font-size: 14px; color: #6c757d; }
        .warning { background: #fff3cd; padding: 15px; border-radius: 6px; margin: 20px 0; border-left: 4px solid #ffc107; }
    </style>
</head>
<body>
    <div class="header">
        <h2>🔒 Secure Message Received</h2>
        {{if .SenderName}}
        <p>You have received a secure message from <strong>{{.SenderName}}</strong> ({{.SenderEmail}}).</p>
        {{else}}
        <p>You have received a secure message from <strong>{{.SenderEmail}}</strong>.</p>
        {{end}}
    </div>

    <div class="content">
        {{if .CustomMessage}}
        <div class="message">
            <p><strong>Message from sender:</strong></p>
            <p>{{.CustomMessage}}</p>
        </div>
        {{end}}

        <div class="security-info">
            <h3>🔐 Security Features</h3>
            <p>This message is protected with <strong>{{.SecurityFeatures}}</strong>.</p>
            <p>{{.ExpirationInfo}}</p>
        </div>

        <div class="warning">
            <p><strong>⚠️ Important:</strong> For security, you may be asked to verify your identity before viewing this message.</p>
        </div>

        <a href="{{.SecureURL}}" class="button">View Secure Message</a>

        <p><small>If the button doesn't work, copy and paste this link into your browser:</small></p>
        <p style="word-break: break-all; font-family: monospace; background: #f8f9fa; padding: 10px; border-radius: 4px;">{{.SecureURL}}</p>
    </div>

    <div class="footer">
        <p>This secure message was sent using SecureMail's encrypted email system.</p>
        <p>Generated on {{.CurrentTime}}</p>
    </div>
</body>
</html>`
