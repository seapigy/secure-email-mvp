// =============================================================================
// SECURE EMAIL MVP - NOTIFICATION SYSTEM
// =============================================================================
// Package notification handles access notifications for secure emails.
// Provides email and SMS notifications with metadata about access attempts.
// =============================================================================

package notification

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeEmail NotificationType = "email"
	NotificationTypeSMS   NotificationType = "sms"
)

// AccessEventType represents the type of access event
type AccessEventType string

const (
	AccessEventTypeSuccess AccessEventType = "success"
	AccessEventTypeFailure AccessEventType = "failure"
	AccessEventTypeBlocked AccessEventType = "blocked"
)

// NotificationPreferences represents user notification preferences
type NotificationPreferences struct {
	UserID             string    `json:"user_id"`
	EmailNotifications bool      `json:"email_notifications"`
	SMSNotifications   bool      `json:"sms_notifications"`
	NotifyOnSuccess    bool      `json:"notify_on_success"`
	NotifyOnFailure    bool      `json:"notify_on_failure"`
	NotifyOnBlocked    bool      `json:"notify_on_blocked"`
	IncludeGeolocation bool      `json:"include_geolocation"`
	IncludeDeviceInfo  bool      `json:"include_device_info"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// AccessEvent represents an email access event
type AccessEvent struct {
	EventID       string          `json:"event_id"`
	EmailID       string          `json:"email_id"`
	UserID        string          `json:"user_id"`
	EventType     AccessEventType `json:"event_type"`
	IPAddress     string          `json:"ip_address"`
	UserAgent     string          `json:"user_agent"`
	Country       string          `json:"country,omitempty"`
	City          string          `json:"city,omitempty"`
	DeviceType    string          `json:"device_type,omitempty"`
	FailureReason string          `json:"failure_reason,omitempty"`
	Timestamp     time.Time       `json:"timestamp"`
}

// NotificationService handles notification operations
type NotificationService struct {
	db *sql.DB
}

// NewNotificationService creates a new notification service
func NewNotificationService(db *sql.DB) *NotificationService {
	return &NotificationService{
		db: db,
	}
}

// RecordAccessEvent records an access event in the database
func (ns *NotificationService) RecordAccessEvent(ctx context.Context, event *AccessEvent) error {
	query := `
		INSERT INTO access_events (
			event_id, email_id, user_id, event_type, ip_address, user_agent,
			country, city, device_type, failure_reason, timestamp
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := ns.db.ExecContext(ctx, query,
		event.EventID,
		event.EmailID,
		event.UserID,
		event.EventType,
		event.IPAddress,
		event.UserAgent,
		event.Country,
		event.City,
		event.DeviceType,
		event.FailureReason,
		event.Timestamp,
	)

	if err != nil {
		log.Printf("Failed to record access event: %v", err)
		return err
	}

	return nil
}

// GetNotificationPreferences retrieves notification preferences for a user
func (ns *NotificationService) GetNotificationPreferences(ctx context.Context, userID string) (*NotificationPreferences, error) {
	query := `
		SELECT user_id, email_notifications, sms_notifications, notify_on_success,
			   notify_on_failure, notify_on_blocked, include_geolocation, include_device_info,
			   created_at, updated_at
		FROM notification_preferences
		WHERE user_id = ?
	`

	var prefs NotificationPreferences
	err := ns.db.QueryRowContext(ctx, query, userID).Scan(
		&prefs.UserID,
		&prefs.EmailNotifications,
		&prefs.SMSNotifications,
		&prefs.NotifyOnSuccess,
		&prefs.NotifyOnFailure,
		&prefs.NotifyOnBlocked,
		&prefs.IncludeGeolocation,
		&prefs.IncludeDeviceInfo,
		&prefs.CreatedAt,
		&prefs.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		// Return default preferences if none exist
		return &NotificationPreferences{
			UserID:             userID,
			EmailNotifications: true,
			SMSNotifications:   false,
			NotifyOnSuccess:    true,
			NotifyOnFailure:    true,
			NotifyOnBlocked:    true,
			IncludeGeolocation: true,
			IncludeDeviceInfo:  true,
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		}, nil
	}

	if err != nil {
		log.Printf("Failed to get notification preferences: %v", err)
		return nil, err
	}

	return &prefs, nil
}

// UpdateNotificationPreferences updates notification preferences for a user
func (ns *NotificationService) UpdateNotificationPreferences(ctx context.Context, prefs *NotificationPreferences) error {
	query := `
		INSERT OR REPLACE INTO notification_preferences (
			user_id, email_notifications, sms_notifications, notify_on_success,
			notify_on_failure, notify_on_blocked, include_geolocation, include_device_info,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	if prefs.CreatedAt.IsZero() {
		prefs.CreatedAt = now
	}
	prefs.UpdatedAt = now

	_, err := ns.db.ExecContext(ctx, query,
		prefs.UserID,
		prefs.EmailNotifications,
		prefs.SMSNotifications,
		prefs.NotifyOnSuccess,
		prefs.NotifyOnFailure,
		prefs.NotifyOnBlocked,
		prefs.IncludeGeolocation,
		prefs.IncludeDeviceInfo,
		prefs.CreatedAt,
		prefs.UpdatedAt,
	)

	if err != nil {
		log.Printf("Failed to update notification preferences: %v", err)
		return err
	}

	return nil
}

// ShouldSendNotification determines if a notification should be sent based on preferences
func (ns *NotificationService) ShouldSendNotification(prefs *NotificationPreferences, eventType AccessEventType) bool {
	switch eventType {
	case AccessEventTypeSuccess:
		return prefs.NotifyOnSuccess
	case AccessEventTypeFailure:
		return prefs.NotifyOnFailure
	case AccessEventTypeBlocked:
		return prefs.NotifyOnBlocked
	default:
		return false
	}
}

// SendNotification sends a notification based on the event and preferences
func (ns *NotificationService) SendNotification(ctx context.Context, event *AccessEvent, prefs *NotificationPreferences) error {
	if !ns.ShouldSendNotification(prefs, event.EventType) {
		log.Printf("Notification skipped for event type %s based on preferences", event.EventType)
		return nil
	}

	// Send email notification if enabled
	if prefs.EmailNotifications {
		if err := ns.sendEmailNotification(ctx, event, prefs); err != nil {
			log.Printf("Failed to send email notification: %v", err)
			// Continue with SMS notification even if email fails
		}
	}

	// Send SMS notification if enabled
	if prefs.SMSNotifications {
		if err := ns.sendSMSNotification(ctx, event, prefs); err != nil {
			log.Printf("Failed to send SMS notification: %v", err)
			return err
		}
	}

	return nil
}

// sendEmailNotification sends an email notification
func (ns *NotificationService) sendEmailNotification(ctx context.Context, event *AccessEvent, prefs *NotificationPreferences) error {
	// Get sender email from database
	var senderEmail string
	query := `SELECT email FROM users WHERE user_id = ?`
	err := ns.db.QueryRowContext(ctx, query, event.UserID).Scan(&senderEmail)
	if err != nil {
		log.Printf("Failed to get sender email: %v", err)
		return err
	}

	// Build notification content
	subject := ns.buildEmailSubject(event)
	body := ns.buildEmailBody(event, prefs)

	// For now, log the email notification
	// In production, this would integrate with an email service like SendGrid, AWS SES, etc.
	log.Printf("EMAIL NOTIFICATION - To: %s, Subject: %s, Body: %s", senderEmail, subject, body)

	// TODO: Integrate with actual email service
	// Example with SendGrid:
	// message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	// response, err := client.Send(message)

	return nil
}

// sendSMSNotification sends an SMS notification
func (ns *NotificationService) sendSMSNotification(ctx context.Context, event *AccessEvent, prefs *NotificationPreferences) error {
	// Get sender phone number from database
	var phoneNumber string
	query := `SELECT phone_number FROM users WHERE user_id = ?`
	err := ns.db.QueryRowContext(ctx, query, event.UserID).Scan(&phoneNumber)
	if err != nil {
		log.Printf("Failed to get sender phone number: %v", err)
		return err
	}

	// Build SMS content
	message := ns.buildSMSMessage(event, prefs)

	// For now, log the SMS notification
	// In production, this would integrate with Twilio or similar service
	log.Printf("SMS NOTIFICATION - To: %s, Message: %s", phoneNumber, message)

	// TODO: Integrate with actual SMS service
	// Example with Twilio:
	// message := client.CreateMessage(&twilio.CreateMessageParams{
	//     To:   &phoneNumber,
	//     From: &fromNumber,
	//     Body: &message,
	// })

	return nil
}

// buildEmailSubject builds the email subject line
func (ns *NotificationService) buildEmailSubject(event *AccessEvent) string {
	switch event.EventType {
	case AccessEventTypeSuccess:
		return "Secure Email Accessed Successfully"
	case AccessEventTypeFailure:
		return "Secure Email Access Attempt Failed"
	case AccessEventTypeBlocked:
		return "Secure Email Access Blocked"
	default:
		return "Secure Email Access Notification"
	}
}

// buildEmailBody builds the email body content
func (ns *NotificationService) buildEmailBody(event *AccessEvent, prefs *NotificationPreferences) string {
	var body strings.Builder

	body.WriteString("Hello,\n\n")
	body.WriteString("Your secure email has been accessed.\n\n")

	// Event details
	body.WriteString("Event Details:\n")
	body.WriteString(fmt.Sprintf("- Event Type: %s\n", event.EventType))
	body.WriteString(fmt.Sprintf("- Timestamp: %s\n", event.Timestamp.Format("2006-01-02 15:04:05 UTC")))

	// Include geolocation if enabled
	if prefs.IncludeGeolocation && event.Country != "" {
		body.WriteString(fmt.Sprintf("- Location: %s", event.Country))
		if event.City != "" {
			body.WriteString(fmt.Sprintf(", %s", event.City))
		}
		body.WriteString("\n")
	}

	// Include device info if enabled
	if prefs.IncludeDeviceInfo && event.DeviceType != "" {
		body.WriteString(fmt.Sprintf("- Device: %s\n", event.DeviceType))
	}

	// Include failure reason if applicable
	if event.FailureReason != "" {
		body.WriteString(fmt.Sprintf("- Reason: %s\n", event.FailureReason))
	}

	body.WriteString("\n")
	body.WriteString("If you did not expect this access, please review your security settings.\n\n")
	body.WriteString("Best regards,\nSecure Email MVP Team")

	return body.String()
}

// buildSMSMessage builds the SMS message content
func (ns *NotificationService) buildSMSMessage(event *AccessEvent, prefs *NotificationPreferences) string {
	var message strings.Builder

	switch event.EventType {
	case AccessEventTypeSuccess:
		message.WriteString("Secure email accessed successfully")
	case AccessEventTypeFailure:
		message.WriteString("Secure email access failed")
	case AccessEventTypeBlocked:
		message.WriteString("Secure email access blocked")
	}

	message.WriteString(fmt.Sprintf(" at %s", event.Timestamp.Format("15:04 UTC")))

	// Include location if enabled and available
	if prefs.IncludeGeolocation && event.Country != "" {
		message.WriteString(fmt.Sprintf(" from %s", event.Country))
	}

	return message.String()
}

// GetAccessEventHistory retrieves access event history for a user
func (ns *NotificationService) GetAccessEventHistory(ctx context.Context, userID string, limit int) ([]*AccessEvent, error) {
	query := `
		SELECT event_id, email_id, user_id, event_type, ip_address, user_agent,
			   country, city, device_type, failure_reason, timestamp
		FROM access_events
		WHERE user_id = ?
		ORDER BY timestamp DESC
		LIMIT ?
	`

	rows, err := ns.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		log.Printf("Failed to get access event history: %v", err)
		return nil, err
	}
	defer rows.Close()

	var events []*AccessEvent
	for rows.Next() {
		var event AccessEvent
		err := rows.Scan(
			&event.EventID,
			&event.EmailID,
			&event.UserID,
			&event.EventType,
			&event.IPAddress,
			&event.UserAgent,
			&event.Country,
			&event.City,
			&event.DeviceType,
			&event.FailureReason,
			&event.Timestamp,
		)
		if err != nil {
			log.Printf("Failed to scan access event: %v", err)
			continue
		}
		events = append(events, &event)
	}

	return events, nil
}

// DetectDeviceType attempts to detect device type from user agent
func DetectDeviceType(userAgent string) string {
	ua := strings.ToLower(userAgent)

	if strings.Contains(ua, "mobile") || strings.Contains(ua, "android") || strings.Contains(ua, "iphone") {
		return "Mobile"
	}
	if strings.Contains(ua, "tablet") || strings.Contains(ua, "ipad") {
		return "Tablet"
	}
	if strings.Contains(ua, "desktop") || strings.Contains(ua, "windows") || strings.Contains(ua, "macintosh") || strings.Contains(ua, "linux") {
		return "Desktop"
	}

	return "Unknown"
}
