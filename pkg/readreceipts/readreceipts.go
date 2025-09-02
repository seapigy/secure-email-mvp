// =============================================================================
// SECURE EMAIL MVP - READ RECEIPTS & EXPIRATION ALERTS
// =============================================================================
// Package readreceipts handles read receipts and expiration alerts for secure emails.
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

// ReadReceiptService handles read receipts and expiration alerts
type ReadReceiptService struct {
	db *sql.DB
}

// ReadEvent represents a read event
type ReadEvent struct {
	EventID     string    `json:"event_id"`
	EmailID     string    `json:"email_id"`
	UserID      string    `json:"user_id"`
	ReadAt      time.Time `json:"read_at"`
	IPAddress   string    `json:"ip_address"`
	UserAgent   string    `json:"user_agent"`
	Country     string    `json:"country,omitempty"`
	City        string    `json:"city,omitempty"`
	DeviceType  string    `json:"device_type,omitempty"`
	IsFirstRead bool      `json:"is_first_read"`
}

// ReadReceipt represents a read receipt notification
type ReadReceipt struct {
	ReceiptID      string    `json:"receipt_id"`
	EmailID        string    `json:"email_id"`
	SenderID       string    `json:"sender_id"`
	RecipientID    string    `json:"recipient_id"`
	SentAt         time.Time `json:"sent_at"`
	DeliveryMethod string    `json:"delivery_method"`
	DeliveryStatus string    `json:"delivery_status"`
	ErrorMessage   string    `json:"error_message,omitempty"`
	Metadata       string    `json:"metadata"`
}

// ExpirationAlert represents an expiration alert notification
type ExpirationAlert struct {
	AlertID        string    `json:"alert_id"`
	EmailID        string    `json:"email_id"`
	SenderID       string    `json:"sender_id"`
	AlertType      string    `json:"alert_type"` // 'reminder' or 'final'
	SentAt         time.Time `json:"sent_at"`
	DeliveryMethod string    `json:"delivery_method"`
	DeliveryStatus string    `json:"delivery_status"`
	ErrorMessage   string    `json:"error_message,omitempty"`
	Metadata       string    `json:"metadata"`
}

// ReadReceiptPreferences represents user preferences for read receipts
type ReadReceiptPreferences struct {
	UserID                 string    `json:"user_id"`
	EnableReadReceipts     bool      `json:"enable_read_receipts"`
	EnableExpirationAlerts bool      `json:"enable_expiration_alerts"`
	ExpirationAlertHours   int       `json:"expiration_alert_hours"`
	DeliveryMethods        string    `json:"delivery_methods"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

// EmailReadReceiptInfo represents read receipt information for an email
type EmailReadReceiptInfo struct {
	EmailID            string     `json:"email_id"`
	FirstReadAt        *time.Time `json:"first_read_at,omitempty"`
	ReadCount          int        `json:"read_count"`
	ReadReceiptSent    bool       `json:"read_receipt_sent"`
	EnableReadReceipts bool       `json:"enable_read_receipts"`
}

// EmailExpirationInfo represents expiration information for an email
type EmailExpirationInfo struct {
	EmailID                  string     `json:"email_id"`
	ExpiresAt                *time.Time `json:"expires_at,omitempty"`
	ExpirationAlertSent      bool       `json:"expiration_alert_sent"`
	FinalExpirationAlertSent bool       `json:"final_expiration_alert_sent"`
	EnableExpirationAlerts   bool       `json:"enable_expiration_alerts"`
	ExpirationAlertHours     int        `json:"expiration_alert_hours"`
}

// NewReadReceiptService creates a new read receipt service
func NewReadReceiptService(db *sql.DB) *ReadReceiptService {
	return &ReadReceiptService{
		db: db,
	}
}

// RecordReadEvent records a read event and handles first read logic
func (rrs *ReadReceiptService) RecordReadEvent(ctx context.Context, event *ReadEvent) error {
	// Check if this is the first read
	var isFirstRead bool
	err := rrs.db.QueryRowContext(ctx,
		"SELECT first_read_at IS NULL FROM emails WHERE email_id = ?",
		event.EmailID).Scan(&isFirstRead)
	if err != nil {
		return fmt.Errorf("failed to check first read status: %v", err)
	}

	event.IsFirstRead = isFirstRead
	event.EventID = uuid.New().String()

	// Insert read event
	_, err = rrs.db.ExecContext(ctx, `
		INSERT INTO read_events (
			event_id, email_id, user_id, read_at, ip_address, user_agent,
			country, city, device_type, is_first_read
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		event.EventID, event.EmailID, event.UserID, event.ReadAt,
		event.IPAddress, event.UserAgent, event.Country, event.City,
		event.DeviceType, event.IsFirstRead,
	)
	if err != nil {
		return fmt.Errorf("failed to insert read event: %v", err)
	}

	// Update email read count and first read timestamp
	if isFirstRead {
		_, err = rrs.db.ExecContext(ctx, `
			UPDATE emails SET 
				first_read_at = ?,
				read_count = read_count + 1
			WHERE email_id = ?`,
			event.ReadAt, event.EmailID,
		)
	} else {
		_, err = rrs.db.ExecContext(ctx, `
			UPDATE emails SET 
				read_count = read_count + 1
			WHERE email_id = ?`,
			event.EmailID,
		)
	}
	if err != nil {
		return fmt.Errorf("failed to update email read count: %v", err)
	}

	log.Printf("Recorded read event for email %s by user %s (first read: %v)",
		event.EmailID, event.UserID, isFirstRead)

	return nil
}

// SendReadReceipt sends a read receipt notification to the sender
func (rrs *ReadReceiptService) SendReadReceipt(ctx context.Context, emailID, senderID, recipientID string, readEvent *ReadEvent) error {
	// Check if read receipts are enabled for this email
	var enableReadReceipts bool
	err := rrs.db.QueryRowContext(ctx,
		"SELECT enable_read_receipts FROM emails WHERE email_id = ?",
		emailID).Scan(&enableReadReceipts)
	if err != nil {
		return fmt.Errorf("failed to check read receipt settings: %v", err)
	}

	if !enableReadReceipts {
		log.Printf("Read receipts disabled for email %s", emailID)
		return nil
	}

	// Check if read receipt already sent
	var readReceiptSent bool
	err = rrs.db.QueryRowContext(ctx,
		"SELECT read_receipt_sent FROM emails WHERE email_id = ?",
		emailID).Scan(&readReceiptSent)
	if err != nil {
		return fmt.Errorf("failed to check read receipt status: %v", err)
	}

	if readReceiptSent {
		log.Printf("Read receipt already sent for email %s", emailID)
		return nil
	}

	// Get sender preferences
	prefs, err := rrs.GetReadReceiptPreferences(ctx, senderID)
	if err != nil {
		log.Printf("Failed to get sender preferences, using defaults: %v", err)
		prefs = &ReadReceiptPreferences{
			EnableReadReceipts: true,
			DeliveryMethods:    "email,sms",
		}
	}

	if !prefs.EnableReadReceipts {
		log.Printf("Read receipts disabled in sender preferences for user %s", senderID)
		return nil
	}

	// Create metadata
	metadata := map[string]interface{}{
		"email_id":      emailID,
		"recipient_id":  recipientID,
		"read_at":       readEvent.ReadAt,
		"ip_address":    readEvent.IPAddress,
		"country":       readEvent.Country,
		"city":          readEvent.City,
		"device_type":   readEvent.DeviceType,
		"is_first_read": readEvent.IsFirstRead,
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

		receipt := &ReadReceipt{
			ReceiptID:      uuid.New().String(),
			EmailID:        emailID,
			SenderID:       senderID,
			RecipientID:    recipientID,
			SentAt:         time.Now(),
			DeliveryMethod: method,
			DeliveryStatus: "pending",
			Metadata:       string(metadataJSON),
		}

		// Store receipt record
		_, err = rrs.db.ExecContext(ctx, `
			INSERT INTO read_receipts (
				receipt_id, email_id, sender_id, recipient_id, sent_at,
				delivery_method, delivery_status, metadata
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			receipt.ReceiptID, receipt.EmailID, receipt.SenderID, receipt.RecipientID,
			receipt.SentAt, receipt.DeliveryMethod, receipt.DeliveryStatus, receipt.Metadata,
		)
		if err != nil {
			log.Printf("Failed to store read receipt record: %v", err)
			continue
		}

		// TODO: Implement actual delivery logic (email/SMS)
		// For now, mark as sent
		_, err = rrs.db.ExecContext(ctx, `
			UPDATE read_receipts SET 
				delivery_status = 'sent'
			WHERE receipt_id = ?`,
			receipt.ReceiptID,
		)
		if err != nil {
			log.Printf("Failed to update read receipt status: %v", err)
		} else {
			successCount++
		}
	}

	// Mark read receipt as sent if at least one delivery succeeded
	if successCount > 0 {
		_, err = rrs.db.ExecContext(ctx, `
			UPDATE emails SET 
				read_receipt_sent = TRUE
			WHERE email_id = ?`,
			emailID,
		)
		if err != nil {
			log.Printf("Failed to mark read receipt as sent: %v", err)
		}
	}

	log.Printf("Sent read receipt for email %s via %d methods", emailID, successCount)
	return nil
}

// GetReadReceiptPreferences retrieves read receipt preferences for a user
func (rrs *ReadReceiptService) GetReadReceiptPreferences(ctx context.Context, userID string) (*ReadReceiptPreferences, error) {
	query := `
		SELECT user_id, enable_read_receipts, enable_expiration_alerts,
			   expiration_alert_hours, delivery_methods, created_at, updated_at
		FROM read_receipt_preferences
		WHERE user_id = ?
	`

	var prefs ReadReceiptPreferences
	err := rrs.db.QueryRowContext(ctx, query, userID).Scan(
		&prefs.UserID, &prefs.EnableReadReceipts, &prefs.EnableExpirationAlerts,
		&prefs.ExpirationAlertHours, &prefs.DeliveryMethods, &prefs.CreatedAt, &prefs.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		// Create default preferences
		prefs = ReadReceiptPreferences{
			UserID:                 userID,
			EnableReadReceipts:     true,
			EnableExpirationAlerts: true,
			ExpirationAlertHours:   24,
			DeliveryMethods:        "email,sms",
			CreatedAt:              time.Now(),
			UpdatedAt:              time.Now(),
		}

		_, err = rrs.db.ExecContext(ctx, `
			INSERT INTO read_receipt_preferences (
				user_id, enable_read_receipts, enable_expiration_alerts,
				expiration_alert_hours, delivery_methods, created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			prefs.UserID, prefs.EnableReadReceipts, prefs.EnableExpirationAlerts,
			prefs.ExpirationAlertHours, prefs.DeliveryMethods, prefs.CreatedAt, prefs.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create default preferences: %v", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("failed to get preferences: %v", err)
	}

	return &prefs, nil
}

// UpdateReadReceiptPreferences updates read receipt preferences for a user
func (rrs *ReadReceiptService) UpdateReadReceiptPreferences(ctx context.Context, prefs *ReadReceiptPreferences) error {
	prefs.UpdatedAt = time.Now()

	_, err := rrs.db.ExecContext(ctx, `
		INSERT OR REPLACE INTO read_receipt_preferences (
			user_id, enable_read_receipts, enable_expiration_alerts,
			expiration_alert_hours, delivery_methods, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		prefs.UserID, prefs.EnableReadReceipts, prefs.EnableExpirationAlerts,
		prefs.ExpirationAlertHours, prefs.DeliveryMethods, prefs.CreatedAt, prefs.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update preferences: %v", err)
	}

	return nil
}

// GetEmailReadReceiptInfo retrieves read receipt information for an email
func (rrs *ReadReceiptService) GetEmailReadReceiptInfo(ctx context.Context, emailID string) (*EmailReadReceiptInfo, error) {
	query := `
		SELECT email_id, first_read_at, read_count, read_receipt_sent, enable_read_receipts
		FROM emails
		WHERE email_id = ?
	`

	var info EmailReadReceiptInfo
	err := rrs.db.QueryRowContext(ctx, query, emailID).Scan(
		&info.EmailID, &info.FirstReadAt, &info.ReadCount, &info.ReadReceiptSent, &info.EnableReadReceipts,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get read receipt info: %v", err)
	}

	return &info, nil
}

// GetEmailExpirationInfo retrieves expiration information for an email
func (rrs *ReadReceiptService) GetEmailExpirationInfo(ctx context.Context, emailID string) (*EmailExpirationInfo, error) {
	query := `
		SELECT email_id, expires_at, expiration_alert_sent, final_expiration_alert_sent,
			   enable_expiration_alerts, expiration_alert_hours
		FROM emails
		WHERE email_id = ?
	`

	var info EmailExpirationInfo
	err := rrs.db.QueryRowContext(ctx, query, emailID).Scan(
		&info.EmailID, &info.ExpiresAt, &info.ExpirationAlertSent, &info.FinalExpirationAlertSent,
		&info.EnableExpirationAlerts, &info.ExpirationAlertHours,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get expiration info: %v", err)
	}

	return &info, nil
}

// GetReadEvents retrieves read events for an email
func (rrs *ReadReceiptService) GetReadEvents(ctx context.Context, emailID string, limit int) ([]*ReadEvent, error) {
	query := `
		SELECT event_id, email_id, user_id, read_at, ip_address, user_agent,
			   country, city, device_type, is_first_read
		FROM read_events
		WHERE email_id = ?
		ORDER BY read_at DESC
		LIMIT ?
	`

	rows, err := rrs.db.QueryContext(ctx, query, emailID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query read events: %v", err)
	}
	defer rows.Close()

	var events []*ReadEvent
	for rows.Next() {
		var event ReadEvent
		err := rows.Scan(
			&event.EventID, &event.EmailID, &event.UserID, &event.ReadAt,
			&event.IPAddress, &event.UserAgent, &event.Country, &event.City,
			&event.DeviceType, &event.IsFirstRead,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan read event: %v", err)
		}
		events = append(events, &event)
	}

	return events, nil
}






















