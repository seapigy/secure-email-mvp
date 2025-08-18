// =============================================================================
// SECURE EMAIL MVP - EXPIRATION ALERTS WORKER
// =============================================================================
// Worker for handling expiration alerts and reminders.
// Micro-Iteration 4.19: Email Read Receipt & Expiration Alerts
// =============================================================================

package readreceipts

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ExpirationWorker handles expiration alerts
type ExpirationWorker struct {
	db *sql.DB
}

// NewExpirationWorker creates a new expiration worker
func NewExpirationWorker(db *sql.DB) *ExpirationWorker {
	return &ExpirationWorker{
		db: db,
	}
}

// ProcessExpirationAlerts processes emails that need expiration alerts
func (ew *ExpirationWorker) ProcessExpirationAlerts(ctx context.Context) error {
	log.Println("Processing expiration alerts...")

	// Find emails that need reminder alerts (X hours before expiration)
	reminderEmails, err := ew.getEmailsNeedingReminder(ctx)
	if err != nil {
		return fmt.Errorf("failed to get emails needing reminder: %v", err)
	}

	for _, email := range reminderEmails {
		if err := ew.sendExpirationReminder(ctx, email); err != nil {
			log.Printf("Failed to send expiration reminder for email %s: %v", email.EmailID, err)
			continue
		}
	}

	// Find emails that need final alerts (expired)
	finalEmails, err := ew.getEmailsNeedingFinalAlert(ctx)
	if err != nil {
		return fmt.Errorf("failed to get emails needing final alert: %v", err)
	}

	for _, email := range finalEmails {
		if err := ew.sendFinalExpirationAlert(ctx, email); err != nil {
			log.Printf("Failed to send final expiration alert for email %s: %v", email.EmailID, err)
			continue
		}
	}

	log.Printf("Processed %d reminder alerts and %d final alerts", len(reminderEmails), len(finalEmails))
	return nil
}

// EmailExpirationData represents email data for expiration processing
type EmailExpirationData struct {
	EmailID                string
	SenderID               string
	Recipient              string
	Subject                string
	ExpiresAt              time.Time
	ExpirationAlertHours   int
	EnableExpirationAlerts bool
}

// getEmailsNeedingReminder finds emails that need reminder alerts
func (ew *ExpirationWorker) getEmailsNeedingReminder(ctx context.Context) ([]*EmailExpirationData, error) {
	query := `
		SELECT e.email_id, e.sender_id, e.recipient, e.subject, e.expires_at,
			   e.expiration_alert_hours, e.enable_expiration_alerts
		FROM emails e
		WHERE e.expires_at IS NOT NULL
		  AND e.expires_at > CURRENT_TIMESTAMP
		  AND e.expiration_alert_sent = FALSE
		  AND e.enable_expiration_alerts = TRUE
		  AND e.expires_at <= datetime(CURRENT_TIMESTAMP, '+' || e.expiration_alert_hours || ' hours')
		ORDER BY e.expires_at ASC
	`

	rows, err := ew.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query emails needing reminder: %v", err)
	}
	defer rows.Close()

	var emails []*EmailExpirationData
	for rows.Next() {
		var email EmailExpirationData
		err := rows.Scan(
			&email.EmailID, &email.SenderID, &email.Recipient, &email.Subject,
			&email.ExpiresAt, &email.ExpirationAlertHours, &email.EnableExpirationAlerts,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan email data: %v", err)
		}
		emails = append(emails, &email)
	}

	return emails, nil
}

// getEmailsNeedingFinalAlert finds emails that need final alerts
func (ew *ExpirationWorker) getEmailsNeedingFinalAlert(ctx context.Context) ([]*EmailExpirationData, error) {
	query := `
		SELECT e.email_id, e.sender_id, e.recipient, e.subject, e.expires_at,
			   e.expiration_alert_hours, e.enable_expiration_alerts
		FROM emails e
		WHERE e.expires_at IS NOT NULL
		  AND e.expires_at <= CURRENT_TIMESTAMP
		  AND e.final_expiration_alert_sent = FALSE
		  AND e.enable_expiration_alerts = TRUE
		ORDER BY e.expires_at ASC
	`

	rows, err := ew.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query emails needing final alert: %v", err)
	}
	defer rows.Close()

	var emails []*EmailExpirationData
	for rows.Next() {
		var email EmailExpirationData
		err := rows.Scan(
			&email.EmailID, &email.SenderID, &email.Recipient, &email.Subject,
			&email.ExpiresAt, &email.ExpirationAlertHours, &email.EnableExpirationAlerts,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan email data: %v", err)
		}
		emails = append(emails, &email)
	}

	return emails, nil
}

// sendExpirationReminder sends a reminder alert for an email
func (ew *ExpirationWorker) sendExpirationReminder(ctx context.Context, email *EmailExpirationData) error {
	// Get sender preferences
	prefs, err := ew.getReadReceiptPreferences(ctx, email.SenderID)
	if err != nil {
		log.Printf("Failed to get sender preferences for email %s, using defaults: %v", email.EmailID, err)
		prefs = &ReadReceiptPreferences{
			EnableExpirationAlerts: true,
			DeliveryMethods:        "email,sms",
		}
	}

	if !prefs.EnableExpirationAlerts {
		log.Printf("Expiration alerts disabled in sender preferences for user %s", email.SenderID)
		return nil
	}

	// Calculate time until expiration
	timeUntilExpiration := time.Until(email.ExpiresAt)
	hoursUntilExpiration := int(timeUntilExpiration.Hours())

	// Create metadata
	metadata := map[string]interface{}{
		"email_id":               email.EmailID,
		"recipient":              email.Recipient,
		"subject":                email.Subject,
		"expires_at":             email.ExpiresAt,
		"hours_until_expiration": hoursUntilExpiration,
		"alert_type":             "reminder",
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %v", err)
	}

	// Send notifications via different methods
	deliveryMethods := strings.Split(prefs.DeliveryMethods, ",")
	successCount := 0

	for _, method := range deliveryMethods {
		method = strings.TrimSpace(method)
		if method == "" {
			continue
		}

		alert := &ExpirationAlert{
			AlertID:        uuid.New().String(),
			EmailID:        email.EmailID,
			SenderID:       email.SenderID,
			AlertType:      "reminder",
			SentAt:         time.Now(),
			DeliveryMethod: method,
			DeliveryStatus: "pending",
			Metadata:       string(metadataJSON),
		}

		// Store alert record
		_, err = ew.db.ExecContext(ctx, `
			INSERT INTO expiration_alerts (
				alert_id, email_id, sender_id, alert_type, sent_at,
				delivery_method, delivery_status, metadata
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			alert.AlertID, alert.EmailID, alert.SenderID, alert.AlertType,
			alert.SentAt, alert.DeliveryMethod, alert.DeliveryStatus, alert.Metadata,
		)
		if err != nil {
			log.Printf("Failed to store expiration alert record: %v", err)
			continue
		}

		// TODO: Implement actual delivery logic (email/SMS)
		// For now, mark as sent
		_, err = ew.db.ExecContext(ctx, `
			UPDATE expiration_alerts SET 
				delivery_status = 'sent'
			WHERE alert_id = ?`,
			alert.AlertID,
		)
		if err != nil {
			log.Printf("Failed to update expiration alert status: %v", err)
		} else {
			successCount++
		}
	}

	// Mark reminder as sent if at least one delivery succeeded
	if successCount > 0 {
		_, err = ew.db.ExecContext(ctx, `
			UPDATE emails SET 
				expiration_alert_sent = TRUE
			WHERE email_id = ?`,
			email.EmailID,
		)
		if err != nil {
			log.Printf("Failed to mark expiration alert as sent: %v", err)
		}
	}

	log.Printf("Sent expiration reminder for email %s via %d methods", email.EmailID, successCount)
	return nil
}

// sendFinalExpirationAlert sends a final alert for an expired email
func (ew *ExpirationWorker) sendFinalExpirationAlert(ctx context.Context, email *EmailExpirationData) error {
	// Get sender preferences
	prefs, err := ew.getReadReceiptPreferences(ctx, email.SenderID)
	if err != nil {
		log.Printf("Failed to get sender preferences for email %s, using defaults: %v", email.EmailID, err)
		prefs = &ReadReceiptPreferences{
			EnableExpirationAlerts: true,
			DeliveryMethods:        "email,sms",
		}
	}

	if !prefs.EnableExpirationAlerts {
		log.Printf("Expiration alerts disabled in sender preferences for user %s", email.SenderID)
		return nil
	}

	// Create metadata
	metadata := map[string]interface{}{
		"email_id":   email.EmailID,
		"recipient":  email.Recipient,
		"subject":    email.Subject,
		"expires_at": email.ExpiresAt,
		"alert_type": "final",
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %v", err)
	}

	// Send notifications via different methods
	deliveryMethods := strings.Split(prefs.DeliveryMethods, ",")
	successCount := 0

	for _, method := range deliveryMethods {
		method = strings.TrimSpace(method)
		if method == "" {
			continue
		}

		alert := &ExpirationAlert{
			AlertID:        uuid.New().String(),
			EmailID:        email.EmailID,
			SenderID:       email.SenderID,
			AlertType:      "final",
			SentAt:         time.Now(),
			DeliveryMethod: method,
			DeliveryStatus: "pending",
			Metadata:       string(metadataJSON),
		}

		// Store alert record
		_, err = ew.db.ExecContext(ctx, `
			INSERT INTO expiration_alerts (
				alert_id, email_id, sender_id, alert_type, sent_at,
				delivery_method, delivery_status, metadata
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			alert.AlertID, alert.EmailID, alert.SenderID, alert.AlertType,
			alert.SentAt, alert.DeliveryMethod, alert.DeliveryStatus, alert.Metadata,
		)
		if err != nil {
			log.Printf("Failed to store expiration alert record: %v", err)
			continue
		}

		// TODO: Implement actual delivery logic (email/SMS)
		// For now, mark as sent
		_, err = ew.db.ExecContext(ctx, `
			UPDATE expiration_alerts SET 
				delivery_status = 'sent'
			WHERE alert_id = ?`,
			alert.AlertID,
		)
		if err != nil {
			log.Printf("Failed to update expiration alert status: %v", err)
		} else {
			successCount++
		}
	}

	// Mark final alert as sent if at least one delivery succeeded
	if successCount > 0 {
		_, err = ew.db.ExecContext(ctx, `
			UPDATE emails SET 
				final_expiration_alert_sent = TRUE
			WHERE email_id = ?`,
			email.EmailID,
		)
		if err != nil {
			log.Printf("Failed to mark final expiration alert as sent: %v", err)
		}
	}

	log.Printf("Sent final expiration alert for email %s via %d methods", email.EmailID, successCount)
	return nil
}

// getReadReceiptPreferences retrieves read receipt preferences for a user
func (ew *ExpirationWorker) getReadReceiptPreferences(ctx context.Context, userID string) (*ReadReceiptPreferences, error) {
	query := `
		SELECT user_id, enable_read_receipts, enable_expiration_alerts,
			   expiration_alert_hours, delivery_methods, created_at, updated_at
		FROM read_receipt_preferences
		WHERE user_id = ?
	`

	var prefs ReadReceiptPreferences
	err := ew.db.QueryRowContext(ctx, query, userID).Scan(
		&prefs.UserID, &prefs.EnableReadReceipts, &prefs.EnableExpirationAlerts,
		&prefs.ExpirationAlertHours, &prefs.DeliveryMethods, &prefs.CreatedAt, &prefs.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		// Return default preferences
		prefs = ReadReceiptPreferences{
			UserID:                 userID,
			EnableReadReceipts:     true,
			EnableExpirationAlerts: true,
			ExpirationAlertHours:   24,
			DeliveryMethods:        "email,sms",
			CreatedAt:              time.Now(),
			UpdatedAt:              time.Now(),
		}
	} else if err != nil {
		return nil, fmt.Errorf("failed to get preferences: %v", err)
	}

	return &prefs, nil
}




