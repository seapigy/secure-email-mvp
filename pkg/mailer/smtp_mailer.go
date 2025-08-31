package mailer

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strings"
	"time"
)

// SMTPConfig holds SMTP configuration
type SMTPConfig struct {
	Host     string
	Port     string
	Username string
	Password string
}

// EmailMessage represents an email to be sent
type EmailMessage struct {
	To      string
	Subject string
	Body    string
	From    string // Required - must be a valid @securesystem.email address
}

// SMTPMailer provides SMTP email sending functionality
type SMTPMailer struct {
	config *SMTPConfig
	client *smtp.Client
}

// NewSMTPMailer creates a new SMTP mailer instance
func NewSMTPMailer() (*SMTPMailer, error) {
	config := &SMTPConfig{
		Host:     os.Getenv("SES_SMTP_HOST"),
		Port:     os.Getenv("SES_SMTP_PORT"),
		Username: os.Getenv("SES_SMTP_USERNAME"),
		Password: os.Getenv("SES_SMTP_PASSWORD"),
	}

	// Validate required configuration
	if config.Host == "" {
		return nil, fmt.Errorf("SES_SMTP_HOST environment variable is required")
	}
	if config.Port == "" {
		return nil, fmt.Errorf("SES_SMTP_PORT environment variable is required")
	}
	if config.Username == "" {
		return nil, fmt.Errorf("SES_SMTP_USERNAME environment variable is required")
	}
	if config.Password == "" {
		return nil, fmt.Errorf("SES_SMTP_PASSWORD environment variable is required")
	}

	log.Printf("📧 SMTP Mailer initialized with host: %s:%s", config.Host, config.Port)

	return &SMTPMailer{
		config: config,
	}, nil
}

// ValidateFromAddress validates that the From address is a valid @securesystem.email address
func (m *SMTPMailer) ValidateFromAddress(from string) error {
	if from == "" {
		return fmt.Errorf("from address is required")
	}

	if !strings.HasSuffix(from, "@securesystem.email") {
		return fmt.Errorf("from address must end with @securesystem.email, got: %s", from)
	}

	// Basic email format validation
	parts := strings.Split(from, "@")
	if len(parts) != 2 {
		return fmt.Errorf("invalid email format: %s", from)
	}

	localPart := parts[0]
	if localPart == "" {
		return fmt.Errorf("local part of email address cannot be empty")
	}

	// Check for valid characters in local part
	for _, char := range localPart {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '.' || char == '_' || char == '-') {
			return fmt.Errorf("invalid character in local part of email address: %c", char)
		}
	}

	return nil
}

// SendEmail sends an email using SMTP
func (m *SMTPMailer) SendEmail(msg EmailMessage) error {
	// Validate From address
	if err := m.ValidateFromAddress(msg.From); err != nil {
		return fmt.Errorf("invalid from address: %w", err)
	}

	// Validate message
	if msg.To == "" {
		return fmt.Errorf("recipient email address is required")
	}
	if msg.Subject == "" {
		return fmt.Errorf("email subject is required")
	}
	if msg.Body == "" {
		return fmt.Errorf("email body is required")
	}

	log.Printf("📧 Attempting to send email from %s to: %s", msg.From, msg.To)

	// Create SMTP connection
	client, err := m.createSMTPConnection()
	if err != nil {
		log.Printf("❌ Failed to create SMTP connection: %v", err)
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer client.Close()

	// Send email
	if err := m.sendEmailViaClient(client, msg); err != nil {
		log.Printf("❌ Failed to send email from %s to %s: %v", msg.From, msg.To, err)
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("✅ Email sent successfully from %s to: %s", msg.From, msg.To)
	return nil
}

// createSMTPConnection creates a secure SMTP connection
func (m *SMTPMailer) createSMTPConnection() (*smtp.Client, error) {
	// Create auth
	auth := smtp.PlainAuth("", m.config.Username, m.config.Password, m.config.Host)

	// Connect to server
	addr := fmt.Sprintf("%s:%s", m.config.Host, m.config.Port)
	client, err := smtp.Dial(addr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial SMTP server: %w", err)
	}

	// Start TLS
	if err = client.StartTLS(&tls.Config{
		ServerName: m.config.Host,
	}); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to start TLS: %w", err)
	}

	// Authenticate
	if err = client.Auth(auth); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to authenticate: %w", err)
	}

	return client, nil
}

// sendEmailViaClient sends an email using an existing SMTP client
func (m *SMTPMailer) sendEmailViaClient(client *smtp.Client, msg EmailMessage) error {
	// Set sender
	if err := client.Mail(msg.From); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Set recipient
	if err := client.Rcpt(msg.To); err != nil {
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	// Send data
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}
	defer writer.Close()

	// Build email content
	emailContent := m.buildEmailContent(msg)
	_, err = writer.Write([]byte(emailContent))
	if err != nil {
		return fmt.Errorf("failed to write email content: %w", err)
	}

	return nil
}

// buildEmailContent builds the complete email content with headers
func (m *SMTPMailer) buildEmailContent(msg EmailMessage) string {
	headers := []string{
		fmt.Sprintf("From: %s", msg.From),
		fmt.Sprintf("To: %s", msg.To),
		fmt.Sprintf("Subject: %s", msg.Subject),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		fmt.Sprintf("Date: %s", time.Now().Format(time.RFC1123Z)),
		"",
		msg.Body,
	}

	return strings.Join(headers, "\r\n")
}

// TestConnection tests SMTP connectivity by attempting to send a test email
func (m *SMTPMailer) TestConnection(testEmail string) error {
	if testEmail == "" {
		return fmt.Errorf("test email address is required")
	}

	log.Printf("🧪 Testing SMTP connection by sending test email to: %s", testEmail)

	testMsg := EmailMessage{
		To:      testEmail,
		Subject: "Secure Email MVP - SMTP Test",
		Body: fmt.Sprintf(`Hello,

This is a test email from the Secure Email MVP system.

Time: %s
SMTP Host: %s:%s

If you receive this email, the SMTP configuration is working correctly.

Best regards,
Secure Email MVP System`, time.Now().Format("2006-01-02 15:04:05 UTC"), m.config.Host, m.config.Port),
		From: "test@securesystem.email", // Use a valid test address
	}

	return m.SendEmail(testMsg)
}

// GetConfig returns a copy of the SMTP configuration (without sensitive data)
func (m *SMTPMailer) GetConfig() map[string]string {
	return map[string]string{
		"host": m.config.Host,
		"port": m.config.Port,
		// Deliberately not including username/password for security
	}
}
