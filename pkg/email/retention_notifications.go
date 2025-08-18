package email

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"secure-email-mvp/pkg/notification"
)

// RetentionNotificationConfig holds configuration for retention notifications
type RetentionNotificationConfig struct {
	EnableExpirationNotifications bool   // Whether to send expiration notifications
	EnableCleanupNotifications    bool   // Whether to send cleanup notifications
	ExpirationAdvanceNoticeHours  int    // Hours before expiration to send notification
	NotificationEmailTemplate     string // Email template for notifications
}

// RetentionNotificationService provides retention notification functionality
type RetentionNotificationService struct {
	db           *sql.DB
	notification *notification.NotificationService
	config       RetentionNotificationConfig
}

// ExpirationNotification represents an expiration notification event
type ExpirationNotification struct {
	EmailID            string     `json:"email_id"`
	SenderID           string     `json:"sender_id"`
	Recipient          string     `json:"recipient"`
	Subject            string     `json:"subject"`
	CreatedAt          time.Time  `json:"created_at"`
	ExpiresAt          time.Time  `json:"expires_at"`
	HoursUntil         int        `json:"hours_until_expiry"`
	NotificationSentAt *time.Time `json:"notification_sent_at,omitempty"`
}

// CleanupNotification represents a cleanup notification event
type CleanupNotification struct {
	EmailID            string     `json:"email_id"`
	SenderID           string     `json:"sender_id"`
	Recipient          string     `json:"recipient"`
	Subject            string     `json:"subject"`
	CleanupReason      string     `json:"cleanup_reason"` // "expired", "burned", "self_destructed"
	CleanupTime        time.Time  `json:"cleanup_time"`
	Initiator          string     `json:"initiator"` // "worker", "manual"
	NotificationSentAt *time.Time `json:"notification_sent_at,omitempty"`
}

// NewRetentionNotificationService creates a new retention notification service
func NewRetentionNotificationService(db *sql.DB) *RetentionNotificationService {
	config := RetentionNotificationConfig{
		EnableExpirationNotifications: getEnableExpirationNotifications(),
		EnableCleanupNotifications:    getEnableCleanupNotifications(),
		ExpirationAdvanceNoticeHours:  getExpirationAdvanceNoticeHours(),
		NotificationEmailTemplate:     getNotificationEmailTemplate(),
	}

	return &RetentionNotificationService{
		db:           db,
		notification: notification.NewNotificationService(db),
		config:       config,
	}
}

// getEnableExpirationNotifications gets whether to enable expiration notifications from environment
func getEnableExpirationNotifications() bool {
	enableStr := os.Getenv("ENABLE_EXPIRATION_NOTIFICATIONS")
	if enableStr == "" {
		return true // Default to enabling expiration notifications
	}

	enable, err := strconv.ParseBool(enableStr)
	if err != nil {
		return true // Default fallback
	}

	return enable
}

// getEnableCleanupNotifications gets whether to enable cleanup notifications from environment
func getEnableCleanupNotifications() bool {
	enableStr := os.Getenv("ENABLE_CLEANUP_NOTIFICATIONS")
	if enableStr == "" {
		return false // Default to disabling cleanup notifications
	}

	enable, err := strconv.ParseBool(enableStr)
	if err != nil {
		return false // Default fallback
	}

	return enable
}

// getExpirationAdvanceNoticeHours gets the advance notice hours from environment
func getExpirationAdvanceNoticeHours() int {
	hoursStr := os.Getenv("EXPIRATION_ADVANCE_NOTICE_HOURS")
	if hoursStr == "" {
		return 24 // Default to 24 hours advance notice
	}

	hours, err := strconv.Atoi(hoursStr)
	if err != nil || hours <= 0 {
		return 24 // Default fallback
	}

	return hours
}

// getNotificationEmailTemplate gets the notification email template from environment
func getNotificationEmailTemplate() string {
	template := os.Getenv("RETENTION_NOTIFICATION_EMAIL_TEMPLATE")
	if template == "" {
		return "default" // Default template
	}

	return template
}

// CheckAndSendExpirationNotifications checks for emails about to expire and sends notifications
func (rns *RetentionNotificationService) CheckAndSendExpirationNotifications(ctx context.Context) error {
	if !rns.config.EnableExpirationNotifications {
		log.Printf("Expiration notifications are disabled")
		return nil
	}

	log.Printf("Checking for emails about to expire (advance notice: %d hours)", rns.config.ExpirationAdvanceNoticeHours)

	// Query for emails that will expire within the advance notice period
	query := `
		SELECT email_id, sender_id, recipient, subject, created_at, expires_at
		FROM emails 
		WHERE expires_at IS NOT NULL 
		AND expires_at > datetime('now') 
		AND expires_at <= datetime('now', '+' || ? || ' hours')
		AND encrypted_blob_url IS NOT NULL
		AND email_id NOT IN (
			SELECT email_id FROM retention_notifications 
			WHERE notification_type = 'expiration' 
			AND created_at >= datetime('now', '-1 day')
		)
	`

	rows, err := rns.db.QueryContext(ctx, query, rns.config.ExpirationAdvanceNoticeHours)
	if err != nil {
		return fmt.Errorf("failed to query expiring emails: %w", err)
	}
	defer rows.Close()

	var notificationsSent int
	now := time.Now()

	for rows.Next() {
		var email ExpirationNotification
		var expiresAtStr string

		err := rows.Scan(
			&email.EmailID, &email.SenderID, &email.Recipient, &email.Subject,
			&email.CreatedAt, &expiresAtStr,
		)
		if err != nil {
			log.Printf("Error scanning expiring email row: %v", err)
			continue
		}

		// Parse expires_at
		if expiresAt, err := time.Parse("2006-01-02 15:04:05", expiresAtStr); err == nil {
			email.ExpiresAt = expiresAt
			email.HoursUntil = int(expiresAt.Sub(now).Hours())
		} else {
			log.Printf("Error parsing expiration time for email %s: %v", email.EmailID, err)
			continue
		}

		// Send expiration notification
		if err := rns.sendExpirationNotification(ctx, &email); err != nil {
			log.Printf("Failed to send expiration notification for email %s: %v", email.EmailID, err)
			continue
		}

		notificationsSent++
	}

	log.Printf("Sent %d expiration notifications", notificationsSent)
	return nil
}

// sendExpirationNotification sends an expiration notification for a specific email
func (rns *RetentionNotificationService) sendExpirationNotification(ctx context.Context, email *ExpirationNotification) error {
	// Get sender email from database
	var senderEmail string
	query := `SELECT email FROM users WHERE user_id = ?`
	err := rns.db.QueryRowContext(ctx, query, email.SenderID).Scan(&senderEmail)
	if err != nil {
		return fmt.Errorf("failed to get sender email: %w", err)
	}

	// Build notification content
	subject := fmt.Sprintf("Email Expiring Soon: %s", email.Subject)
	body := rns.buildExpirationNotificationBody(email)

	// For now, log the notification
	// In production, this would integrate with an email service
	log.Printf("EXPIRATION NOTIFICATION - To: %s, Subject: %s, EmailID: %s, HoursUntil: %d, Body: %s",
		senderEmail, subject, email.EmailID, email.HoursUntil, body)

	// Record the notification
	if err := rns.recordRetentionNotification(ctx, email.EmailID, email.SenderID, "expiration", email); err != nil {
		log.Printf("Failed to record retention notification: %v", err)
	}

	// TODO: Integrate with actual email service
	// Example with SendGrid:
	// message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	// response, err := client.Send(message)

	return nil
}

// buildExpirationNotificationBody builds the expiration notification email body
func (rns *RetentionNotificationService) buildExpirationNotificationBody(email *ExpirationNotification) string {
	body := fmt.Sprintf(`
Hello,

Your secure email "%s" (ID: %s) will expire in %d hours.

Email Details:
- Subject: %s
- Recipient: %s
- Created: %s
- Expires: %s

This email will be automatically deleted after expiration. If you need to extend the expiration time, please contact support.

Best regards,
Secure Email System
`, email.Subject, email.EmailID, email.HoursUntil, email.Subject, email.Recipient,
		email.CreatedAt.Format("January 2, 2006 at 3:04 PM"),
		email.ExpiresAt.Format("January 2, 2006 at 3:04 PM"))

	return body
}

// SendCleanupNotification sends a cleanup notification for a deleted email
func (rns *RetentionNotificationService) SendCleanupNotification(ctx context.Context, email *CleanupNotification) error {
	if !rns.config.EnableCleanupNotifications {
		return nil // Cleanup notifications are disabled
	}

	// Get sender email from database
	var senderEmail string
	query := `SELECT email FROM users WHERE user_id = ?`
	err := rns.db.QueryRowContext(ctx, query, email.SenderID).Scan(&senderEmail)
	if err != nil {
		return fmt.Errorf("failed to get sender email: %w", err)
	}

	// Build notification content
	subject := fmt.Sprintf("Email Deleted: %s", email.Subject)
	body := rns.buildCleanupNotificationBody(email)

	// For now, log the notification
	// In production, this would integrate with an email service
	log.Printf("CLEANUP NOTIFICATION - To: %s, Subject: %s, EmailID: %s, Reason: %s, Initiator: %s, Body: %s",
		senderEmail, subject, email.EmailID, email.CleanupReason, email.Initiator, body)

	// Record the notification
	if err := rns.recordRetentionNotification(ctx, email.EmailID, email.SenderID, "cleanup", email); err != nil {
		log.Printf("Failed to record retention notification: %v", err)
	}

	// TODO: Integrate with actual email service
	// Example with SendGrid:
	// message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	// response, err := client.Send(message)

	return nil
}

// buildCleanupNotificationBody builds the cleanup notification email body
func (rns *RetentionNotificationService) buildCleanupNotificationBody(email *CleanupNotification) string {
	reasonText := email.CleanupReason
	switch email.CleanupReason {
	case "expired":
		reasonText = "expired"
	case "burned":
		reasonText = "was accessed (burn-after-read)"
	case "self_destructed":
		reasonText = "was self-destructed due to failed attempts"
	}

	body := fmt.Sprintf(`
Hello,

Your secure email "%s" (ID: %s) has been automatically deleted.

Email Details:
- Subject: %s
- Recipient: %s
- Deletion Reason: %s
- Deletion Time: %s
- Deletion Method: %s

The email was deleted as part of our automated retention policy. This action cannot be undone.

Best regards,
Secure Email System
`, email.Subject, email.EmailID, email.Subject, email.Recipient, reasonText,
		email.CleanupTime.Format("January 2, 2006 at 3:04 PM"), email.Initiator)

	return body
}

// recordRetentionNotification records a retention notification in the database
func (rns *RetentionNotificationService) recordRetentionNotification(ctx context.Context, emailID, senderID, notificationType string, data interface{}) error {
	query := `
		INSERT INTO retention_notifications (
			email_id, sender_id, notification_type, notification_data, created_at
		) VALUES (?, ?, ?, ?, ?)
	`

	// For now, store a simple JSON representation
	notificationData := fmt.Sprintf("%+v", data)

	_, err := rns.db.ExecContext(ctx, query, emailID, senderID, notificationType, notificationData, time.Now())
	if err != nil {
		return fmt.Errorf("failed to record retention notification: %w", err)
	}

	return nil
}

// GetNotificationConfig returns the current notification configuration
func (rns *RetentionNotificationService) GetNotificationConfig() RetentionNotificationConfig {
	return rns.config
}
