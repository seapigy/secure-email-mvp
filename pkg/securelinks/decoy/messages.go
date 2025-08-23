package decoy

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"strings"
	"time"
)

// =============================================================================
// DECOY MESSAGE SYSTEM
// =============================================================================

// DecoyMessageService handles decoy message generation and display
type DecoyMessageService struct {
	db *sql.DB
}

// DecoyMessage represents a decoy message
type DecoyMessage struct {
	ID          string    `json:"id" db:"id"`
	LinkID      string    `json:"link_id" db:"link_id"`
	TriggerType string    `json:"trigger_type" db:"trigger_type"` // "wrong_password", "revoked", "expired", "blocked"
	Subject     string    `json:"subject" db:"subject"`
	Body        string    `json:"body" db:"body"`
	SenderName  string    `json:"sender_name" db:"sender_name"`
	SenderEmail string    `json:"sender_email" db:"sender_email"`
	SentAt      time.Time `json:"sent_at" db:"sent_at"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	IsActive    bool      `json:"is_active" db:"is_active"`
}

// DecoyTrigger represents a decoy trigger condition
type DecoyTrigger struct {
	ID          string    `json:"id" db:"id"`
	LinkID      string    `json:"link_id" db:"link_id"`
	TriggerType string    `json:"trigger_type" db:"trigger_type"`
	Condition   string    `json:"condition" db:"condition"` // JSON condition
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// DecoyRequest represents a request to get a decoy message
type DecoyRequest struct {
	LinkID      string `json:"link_id" validate:"required"`
	TriggerType string `json:"trigger_type" validate:"required"`
	IPAddress   string `json:"ip_address"`
	UserAgent   string `json:"user_agent"`
}

// DecoyResponse represents a decoy message response
type DecoyResponse struct {
	Success     bool          `json:"success"`
	Message     *DecoyMessage `json:"message,omitempty"`
	Error       string        `json:"error,omitempty"`
	TriggerType string        `json:"trigger_type,omitempty"`
}

// NewDecoyMessageService creates a new decoy message service
func NewDecoyMessageService(db *sql.DB) *DecoyMessageService {
	return &DecoyMessageService{
		db: db,
	}
}

// GetDecoyMessage retrieves a decoy message based on trigger conditions
func (d *DecoyMessageService) GetDecoyMessage(ctx context.Context, req DecoyRequest) (*DecoyResponse, error) {
	// Validate trigger type
	if !d.isValidTriggerType(req.TriggerType) {
		return &DecoyResponse{
			Success: false,
			Error:   "Invalid trigger type",
		}, fmt.Errorf("invalid trigger type: %s", req.TriggerType)
	}

	// Check if decoy message exists for this link and trigger
	decoyMessage, err := d.getDecoyMessageForLink(ctx, req.LinkID, req.TriggerType)
	if err != nil {
		// If no specific decoy message exists, generate a default one
		decoyMessage, err = d.generateDefaultDecoyMessage(ctx, req.LinkID, req.TriggerType)
		if err != nil {
			return &DecoyResponse{
				Success: false,
				Error:   "Failed to generate decoy message",
			}, fmt.Errorf("failed to generate decoy message: %w", err)
		}
	}

	// Log decoy message access for security monitoring
	d.logDecoyAccess(ctx, req.LinkID, req.TriggerType, req.IPAddress, req.UserAgent)

	return &DecoyResponse{
		Success:     true,
		Message:     decoyMessage,
		TriggerType: req.TriggerType,
	}, nil
}

// CreateDecoyMessage creates a custom decoy message
func (d *DecoyMessageService) CreateDecoyMessage(ctx context.Context, decoyMessage *DecoyMessage) error {
	// Generate ID if not provided
	if decoyMessage.ID == "" {
		decoyMessage.ID = d.generateDecoyID()
	}

	// Set timestamps
	if decoyMessage.CreatedAt.IsZero() {
		decoyMessage.CreatedAt = time.Now()
	}
	if decoyMessage.SentAt.IsZero() {
		decoyMessage.SentAt = time.Now()
	}

	// Store in database
	query := `
		INSERT INTO link_decoy_messages (
			id, link_id, trigger_type, subject, body, sender_name, sender_email,
			sent_at, created_at, is_active
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := d.db.ExecContext(ctx, query,
		decoyMessage.ID, decoyMessage.LinkID, decoyMessage.TriggerType,
		decoyMessage.Subject, decoyMessage.Body, decoyMessage.SenderName,
		decoyMessage.SenderEmail, decoyMessage.SentAt, decoyMessage.CreatedAt,
		decoyMessage.IsActive,
	)

	return err
}

// getDecoyMessageForLink retrieves a decoy message for a specific link and trigger
func (d *DecoyMessageService) getDecoyMessageForLink(ctx context.Context, linkID, triggerType string) (*DecoyMessage, error) {
	query := `
		SELECT id, link_id, trigger_type, subject, body, sender_name, sender_email,
		       sent_at, created_at, is_active
		FROM link_decoy_messages
		WHERE link_id = ? AND trigger_type = ? AND is_active = 1
		ORDER BY created_at DESC
		LIMIT 1
	`

	var decoyMessage DecoyMessage
	err := d.db.QueryRowContext(ctx, query, linkID, triggerType).Scan(
		&decoyMessage.ID, &decoyMessage.LinkID, &decoyMessage.TriggerType,
		&decoyMessage.Subject, &decoyMessage.Body, &decoyMessage.SenderName,
		&decoyMessage.SenderEmail, &decoyMessage.SentAt, &decoyMessage.CreatedAt,
		&decoyMessage.IsActive,
	)

	if err != nil {
		return nil, err
	}

	return &decoyMessage, nil
}

// generateDefaultDecoyMessage generates a default decoy message based on trigger type
func (d *DecoyMessageService) generateDefaultDecoyMessage(ctx context.Context, linkID, triggerType string) (*DecoyMessage, error) {
	decoyMessage := &DecoyMessage{
		ID:          d.generateDecoyID(),
		LinkID:      linkID,
		TriggerType: triggerType,
		CreatedAt:   time.Now(),
		SentAt:      time.Now(),
		IsActive:    true,
	}

	// Generate appropriate content based on trigger type
	switch triggerType {
	case "wrong_password":
		decoyMessage.Subject = "Meeting Reminder"
		decoyMessage.Body = "Hi there,\n\nJust a friendly reminder about our upcoming meeting tomorrow at 2 PM.\n\nPlease let me know if you need to reschedule.\n\nBest regards,\nJohn"
		decoyMessage.SenderName = "John Smith"
		decoyMessage.SenderEmail = "john.smith@company.com"

	case "revoked":
		decoyMessage.Subject = "Project Update"
		decoyMessage.Body = "Hello,\n\nI wanted to share the latest updates on our project progress.\n\nWe're on track to meet our deadlines and the team is doing great work.\n\nRegards,\nSarah"
		decoyMessage.SenderName = "Sarah Johnson"
		decoyMessage.SenderEmail = "sarah.johnson@company.com"

	case "expired":
		decoyMessage.Subject = "Weekly Newsletter"
		decoyMessage.Body = "Hi,\n\nHere's this week's newsletter with the latest company updates and announcements.\n\nHave a great week!\n\nBest,\nMarketing Team"
		decoyMessage.SenderName = "Marketing Team"
		decoyMessage.SenderEmail = "marketing@company.com"

	case "blocked":
		decoyMessage.Subject = "Invoice #INV-2024-001"
		decoyMessage.Body = "Dear Customer,\n\nPlease find attached the invoice for services rendered.\n\nPayment is due within 30 days.\n\nThank you,\nAccounting Department"
		decoyMessage.SenderName = "Accounting Department"
		decoyMessage.SenderEmail = "accounting@company.com"

	default:
		decoyMessage.Subject = "General Communication"
		decoyMessage.Body = "Hello,\n\nThis is a general communication message.\n\nBest regards,\nCompany Team"
		decoyMessage.SenderName = "Company Team"
		decoyMessage.SenderEmail = "team@company.com"
	}

	// Store the generated decoy message
	if err := d.CreateDecoyMessage(ctx, decoyMessage); err != nil {
		return nil, fmt.Errorf("failed to store decoy message: %w", err)
	}

	return decoyMessage, nil
}

// logDecoyAccess logs decoy message access for security monitoring
func (d *DecoyMessageService) logDecoyAccess(ctx context.Context, linkID, triggerType, ipAddress, userAgent string) {
	query := `
		INSERT INTO link_tamper_alerts (
			id, link_id, alert_type, severity, ip_address, user_agent, details, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	alertID := d.generateAlertID()
	details := fmt.Sprintf(`{"trigger_type": "%s", "decoy_accessed": true}`, triggerType)

	_, err := d.db.ExecContext(ctx, query,
		alertID, linkID, "decoy_access", "medium", ipAddress, userAgent, details, time.Now(),
	)

	if err != nil {
		log.Printf("Warning: Failed to log decoy access: %v", err)
	}
}

// generateDecoyID generates a unique decoy message ID
func (d *DecoyMessageService) generateDecoyID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}

// generateAlertID generates a unique alert ID
func (d *DecoyMessageService) generateAlertID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}

// isValidTriggerType checks if the trigger type is valid
func (d *DecoyMessageService) isValidTriggerType(triggerType string) bool {
	validTypes := []string{"wrong_password", "revoked", "expired", "blocked", "suspicious_activity"}
	for _, validType := range validTypes {
		if triggerType == validType {
			return true
		}
	}
	return false
}

// GetDecoyMessageTemplates returns predefined decoy message templates
func (d *DecoyMessageService) GetDecoyMessageTemplates() map[string]DecoyMessage {
	return map[string]DecoyMessage{
		"meeting_reminder": {
			Subject:     "Meeting Reminder",
			Body:        "Hi there,\n\nJust a friendly reminder about our upcoming meeting tomorrow at 2 PM.\n\nPlease let me know if you need to reschedule.\n\nBest regards,\nJohn",
			SenderName:  "John Smith",
			SenderEmail: "john.smith@company.com",
		},
		"project_update": {
			Subject:     "Project Update",
			Body:        "Hello,\n\nI wanted to share the latest updates on our project progress.\n\nWe're on track to meet our deadlines and the team is doing great work.\n\nRegards,\nSarah",
			SenderName:  "Sarah Johnson",
			SenderEmail: "sarah.johnson@company.com",
		},
		"newsletter": {
			Subject:     "Weekly Newsletter",
			Body:        "Hi,\n\nHere's this week's newsletter with the latest company updates and announcements.\n\nHave a great week!\n\nBest,\nMarketing Team",
			SenderName:  "Marketing Team",
			SenderEmail: "marketing@company.com",
		},
		"invoice": {
			Subject:     "Invoice #INV-2024-001",
			Body:        "Dear Customer,\n\nPlease find attached the invoice for services rendered.\n\nPayment is due within 30 days.\n\nThank you,\nAccounting Department",
			SenderName:  "Accounting Department",
			SenderEmail: "accounting@company.com",
		},
		"general": {
			Subject:     "General Communication",
			Body:        "Hello,\n\nThis is a general communication message.\n\nBest regards,\nCompany Team",
			SenderName:  "Company Team",
			SenderEmail: "team@company.com",
		},
	}
}

// SanitizeDecoyContent sanitizes decoy message content to remove any sensitive information
func (d *DecoyMessageService) SanitizeDecoyContent(content string) string {
	// Remove any potential sensitive patterns
	sanitized := content

	// Remove email addresses (replace with generic ones)
	sanitized = strings.ReplaceAll(sanitized, "@", "[at]")

	// Remove phone numbers (replace with generic ones)
	// This is a simple pattern - in production, use regex
	sanitized = strings.ReplaceAll(sanitized, "+1-", "[phone]")
	sanitized = strings.ReplaceAll(sanitized, "555-", "[phone]")

	// Remove specific dates (replace with generic ones)
	// This is a simple pattern - in production, use regex
	sanitized = strings.ReplaceAll(sanitized, "2024-", "[year]-")
	sanitized = strings.ReplaceAll(sanitized, "2025-", "[year]-")

	return sanitized
}
