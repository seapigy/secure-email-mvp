package email

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/smtp"
	"os"
	"strconv"
	"time"
)

// EmailConfig holds SMTP configuration
type EmailConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

// GetEmailConfig returns email configuration from environment variables
func GetEmailConfig() *EmailConfig {
	port, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if port == 0 {
		port = 587 // Default SMTP port
	}
	
	return &EmailConfig{
		Host:     os.Getenv("SMTP_HOST"),
		Port:     port,
		Username: os.Getenv("SMTP_USER"),
		Password: os.Getenv("SMTP_PASS"),
		From:     os.Getenv("EMAIL_FROM"),
	}
}

// VerificationEmailData holds data for verification email templates
type VerificationEmailData struct {
	Name             string
	VerificationLink string
	Expiry           string
}

// SendVerificationEmail sends a verification email to the user
func SendVerificationEmail(toEmail, subject, plaintextToken, verificationLink string) error {
	config := GetEmailConfig()
	
	// Validate configuration
	if config.Host == "" || config.Username == "" || config.Password == "" || config.From == "" {
		return fmt.Errorf("email configuration incomplete: missing SMTP settings")
	}
	
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	// Load email templates
	htmlTemplate, err := loadHTMLTemplate()
	if err != nil {
		return fmt.Errorf("failed to load HTML template: %w", err)
	}
	
	textTemplate, err := loadTextTemplate()
	if err != nil {
		return fmt.Errorf("failed to load text template: %w", err)
	}
	
	// Prepare email data
	data := VerificationEmailData{
		Name:             toEmail, // Use email as name for now
		VerificationLink: verificationLink,
		Expiry:           "24 hours",
	}
	
	// Render HTML template
	var htmlBody bytes.Buffer
	if err := htmlTemplate.Execute(&htmlBody, data); err != nil {
		return fmt.Errorf("failed to render HTML template: %w", err)
	}
	
	// Render text template
	var textBody bytes.Buffer
	if err := textTemplate.Execute(&textBody, data); err != nil {
		return fmt.Errorf("failed to render text template: %w", err)
	}
	
	// Create multipart email
	boundary := "boundary123456789"
	message := fmt.Sprintf(`From: %s
To: %s
Subject: %s
MIME-Version: 1.0
Content-Type: multipart/alternative; boundary="%s"

--%s
Content-Type: text/plain; charset=UTF-8

%s

--%s
Content-Type: text/html; charset=UTF-8

%s

--%s--
`, config.From, toEmail, subject, boundary, boundary, textBody.String(), boundary, htmlBody.String(), boundary)
	
	// Send email with retry logic
	return sendWithRetry(ctx, config, toEmail, []byte(message))
}

// sendWithRetry sends email with retry logic
func sendWithRetry(ctx context.Context, config *EmailConfig, to string, message []byte) error {
	maxRetries := 3
	baseDelay := time.Second
	
	for attempt := 0; attempt < maxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		
		err := sendEmail(config, to, message)
		if err == nil {
			return nil // Success
		}
		
		if attempt < maxRetries-1 {
			delay := baseDelay * time.Duration(1<<attempt) // Exponential backoff
			time.Sleep(delay)
		}
	}
	
	return fmt.Errorf("failed to send email after %d attempts", maxRetries)
}

// sendEmail sends a single email
func sendEmail(config *EmailConfig, to string, message []byte) error {
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	
	auth := smtp.PlainAuth("", config.Username, config.Password, config.Host)
	
	return smtp.SendMail(addr, auth, config.From, []string{to}, message)
}

// loadHTMLTemplate loads the HTML email template
func loadHTMLTemplate() (*template.Template, error) {
	templatePath := "email_templates/verification.html"
	
	// Check if template file exists
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		// Return a default template if file doesn't exist
		return template.New("verification").Parse(defaultHTMLTemplate)
	}
	
	return template.ParseFiles(templatePath)
}

// loadTextTemplate loads the text email template
func loadTextTemplate() (*template.Template, error) {
	templatePath := "email_templates/verification.txt"
	
	// Check if template file exists
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		// Return a default template if file doesn't exist
		return template.New("verification").Parse(defaultTextTemplate)
	}
	
	return template.ParseFiles(templatePath)
}

// Default templates (fallback if files don't exist)
const defaultHTMLTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Email Verification</title>
</head>
<body>
    <h2>Verify Your Email Address</h2>
    <p>Hello {{.Name}},</p>
    <p>Please click the link below to verify your email address:</p>
    <p><a href="{{.VerificationLink}}">Verify Email Address</a></p>
    <p>This link will expire in {{.Expiry}}.</p>
    <p>If you didn't create an account, please ignore this email.</p>
    <hr>
    <p><small>Secure Email System</small></p>
</body>
</html>
`

const defaultTextTemplate = `
Verify Your Email Address

Hello {{.Name}},

Please click the link below to verify your email address:

{{.VerificationLink}}

This link will expire in {{.Expiry}}.

If you didn't create an account, please ignore this email.

---
Secure Email System
`
