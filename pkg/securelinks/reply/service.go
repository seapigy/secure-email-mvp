package reply

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"secure-email-mvp/pkg/securelinks"
)

// =============================================================================
// REPLY HANDLING SERVICE
// =============================================================================

// ReplyService handles secure replies from external users
type ReplyService struct {
	db *sql.DB
}

// ReplyRequest represents a reply from an external user
type ReplyRequest struct {
	LinkID      string `json:"link_id" validate:"required"`
	Subject     string `json:"subject" validate:"required"`
	Body        string `json:"body" validate:"required"`
	SenderEmail string `json:"sender_email" validate:"required"`
	IPAddress   string `json:"ip_address"`
	UserAgent   string `json:"user_agent"`
}

// ReplyResponse represents the response to a reply request
type ReplyResponse struct {
	Success bool   `json:"success"`
	ReplyID string `json:"reply_id,omitempty"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
	ChainID string `json:"chain_id,omitempty"`
}

// SecureReply represents a secure reply in the database
type SecureReply struct {
	ID              string     `json:"id" db:"id"`
	LinkID          string     `json:"link_id" db:"link_id"`
	ChainID         string     `json:"chain_id" db:"chain_id"`
	Subject         string     `json:"subject" db:"subject"`
	Body            string     `json:"body" db:"body"`
	SenderEmail     string     `json:"sender_email" db:"sender_email"`
	RecipientEmail  string     `json:"recipient_email" db:"recipient_email"`
	IPAddress       string     `json:"ip_address" db:"ip_address"`
	UserAgent       string     `json:"user_agent" db:"user_agent"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	ProcessedAt     *time.Time `json:"processed_at,omitempty" db:"processed_at"`
	Status          string     `json:"status" db:"status"` // "pending", "processed", "failed"
	InternalEmailID *string    `json:"internal_email_id,omitempty" db:"internal_email_id"`
}

// NewReplyService creates a new reply service
func NewReplyService(db *sql.DB) *ReplyService {
	return &ReplyService{
		db: db,
	}
}

// ProcessReply processes a secure reply from an external user
func (r *ReplyService) ProcessReply(ctx context.Context, req ReplyRequest) (*ReplyResponse, error) {
	// Validate that the link exists and is active
	link, err := r.getSecureLink(ctx, req.LinkID)
	if err != nil {
		return &ReplyResponse{
			Success: false,
			Error:   "Invalid or expired secure link",
		}, fmt.Errorf("failed to get secure link: %w", err)
	}

	// Check if link is expired
	if link.ExpiresAt != nil && time.Now().After(*link.ExpiresAt) {
		return &ReplyResponse{
			Success: false,
			Error:   "Secure link has expired",
		}, fmt.Errorf("secure link expired")
	}

	// Check if link is destroyed
	if link.Status == "destroyed" {
		return &ReplyResponse{
			Success: false,
			Error:   "Secure link has been destroyed",
		}, fmt.Errorf("secure link destroyed")
	}

	// Validate and sanitize reply content
	if err := r.validateReplyContent(req); err != nil {
		return &ReplyResponse{
			Success: false,
			Error:   "Invalid reply content",
		}, fmt.Errorf("reply content validation failed: %w", err)
	}

	// Get or create email chain
	chainID, err := r.getOrCreateEmailChain(ctx, req.LinkID, link.EmailID, link.RecipientEmail)
	if err != nil {
		return &ReplyResponse{
			Success: false,
			Error:   "Failed to process email chain",
		}, fmt.Errorf("failed to get or create email chain: %w", err)
	}

	// Create secure reply record
	reply := &SecureReply{
		ID:             r.generateReplyID(),
		LinkID:         req.LinkID,
		ChainID:        chainID,
		Subject:        req.Subject,
		Body:           req.Body,
		SenderEmail:    req.SenderEmail,
		RecipientEmail: link.RecipientEmail,
		IPAddress:      req.IPAddress,
		UserAgent:      req.UserAgent,
		CreatedAt:      time.Now(),
		Status:         "pending",
	}

	// Store reply in database
	if err := r.storeSecureReply(ctx, reply); err != nil {
		return &ReplyResponse{
			Success: false,
			Error:   "Failed to store reply",
		}, fmt.Errorf("failed to store secure reply: %w", err)
	}

	// Add message to chain
	if err := r.addMessageToChain(ctx, chainID, reply); err != nil {
		log.Printf("Warning: Failed to add message to chain: %v", err)
		// Don't fail the request, just log the warning
	}

	// Forward reply to internal system (async)
	go func() {
		if err := r.ForwardReplyToInternal(context.Background(), reply); err != nil {
			log.Printf("Error forwarding reply to internal system: %v", err)
		}
	}()

	return &ReplyResponse{
		Success: true,
		ReplyID: reply.ID,
		Message: "Reply submitted successfully",
		ChainID: chainID,
	}, nil
}

// ForwardReplyToInternal forwards the reply to the internal email system
func (r *ReplyService) ForwardReplyToInternal(ctx context.Context, reply *SecureReply) error {
	// Get the original email to find the internal recipient
	originalEmail, err := r.getOriginalEmail(ctx, reply.LinkID)
	if err != nil {
		return fmt.Errorf("failed to get original email: %w", err)
	}

	// Create internal email from the reply
	internalEmailID, err := r.createInternalEmail(ctx, reply, originalEmail)
	if err != nil {
		return fmt.Errorf("failed to create internal email: %w", err)
	}

	// Update reply status
	query := `
		UPDATE secure_replies 
		SET status = 'processed', processed_at = CURRENT_TIMESTAMP, internal_email_id = ?
		WHERE id = ?
	`

	_, err = r.db.ExecContext(ctx, query, internalEmailID, reply.ID)
	if err != nil {
		return fmt.Errorf("failed to update reply status: %w", err)
	}

	log.Printf("Reply forwarded to internal system: %s -> %s", reply.ID, internalEmailID)
	return nil
}

// GetReplyHistory gets reply history for a secure link chain
func (r *ReplyService) GetReplyHistory(ctx context.Context, chainID string) ([]*SecureReply, error) {
	query := `
		SELECT id, link_id, chain_id, subject, body, sender_email, recipient_email,
		       ip_address, user_agent, created_at, processed_at, status, internal_email_id
		FROM secure_replies
		WHERE chain_id = ?
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to query reply history: %w", err)
	}
	defer rows.Close()

	var replies []*SecureReply
	for rows.Next() {
		var reply SecureReply
		err := rows.Scan(
			&reply.ID, &reply.LinkID, &reply.ChainID, &reply.Subject, &reply.Body,
			&reply.SenderEmail, &reply.RecipientEmail, &reply.IPAddress, &reply.UserAgent,
			&reply.CreatedAt, &reply.ProcessedAt, &reply.Status, &reply.InternalEmailID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reply: %w", err)
		}
		replies = append(replies, &reply)
	}

	return replies, nil
}

// Helper methods

// getSecureLink retrieves a secure link from the database
func (r *ReplyService) getSecureLink(ctx context.Context, linkID string) (*securelinks.SecureLink, error) {
	query := `
		SELECT link_id, email_id, recipient_email, created_at, expires_at, status,
		       security_settings
		FROM secure_links
		WHERE link_id = ?
	`

	var link securelinks.SecureLink
	var securitySettingsJSON string

	err := r.db.QueryRowContext(ctx, query, linkID).Scan(
		&link.LinkID, &link.EmailID, &link.RecipientEmail, &link.CreatedAt,
		&link.ExpiresAt, &link.Status, &securitySettingsJSON,
	)
	if err != nil {
		return nil, err
	}

	// Parse security settings using JSON unmarshaling
	if err := json.Unmarshal([]byte(securitySettingsJSON), &link.SecuritySettings); err != nil {
		return nil, fmt.Errorf("failed to parse security settings: %w", err)
	}

	return &link, nil
}

// validateReplyContent validates and sanitizes reply content
func (r *ReplyService) validateReplyContent(req ReplyRequest) error {
	// Check content length
	if len(req.Subject) > 500 {
		return fmt.Errorf("subject too long (max 500 characters)")
	}
	if len(req.Body) > 10000 {
		return fmt.Errorf("body too long (max 10000 characters)")
	}

	// Basic email validation for sender
	if !r.isValidEmail(req.SenderEmail) {
		return fmt.Errorf("invalid sender email format")
	}

	// TODO: Add more content validation (XSS prevention, spam detection, etc.)
	return nil
}

// isValidEmail performs basic email validation
func (r *ReplyService) isValidEmail(email string) bool {
	// Simple email validation - in production, use a proper email validation library
	return len(email) > 3 && len(email) < 254 && r.contains(email, "@") && r.contains(email, ".")
}

// contains checks if a string contains a substring
func (r *ReplyService) contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || r.indexOf(s, substr) >= 0)
}

// indexOf finds the index of a substring in a string
func (r *ReplyService) indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// getOrCreateEmailChain gets an existing email chain or creates a new one
func (r *ReplyService) getOrCreateEmailChain(ctx context.Context, linkID, emailID, externalEmail string) (string, error) {
	// First, try to find an existing chain
	query := `
		SELECT id FROM email_chains 
		WHERE initial_link_id = ? OR external_email = ?
		ORDER BY created_at ASC LIMIT 1
	`

	var chainID string
	err := r.db.QueryRowContext(ctx, query, linkID, externalEmail).Scan(&chainID)
	if err == nil {
		// Chain exists, update last activity
		updateQuery := `UPDATE email_chains SET last_activity = CURRENT_TIMESTAMP WHERE id = ?`
		r.db.ExecContext(ctx, updateQuery, chainID)
		return chainID, nil
	}

	// Create new chain
	chainID = r.generateChainID()
	insertQuery := `
		INSERT INTO email_chains (
			id, initial_link_id, internal_user_id, external_email, subject, 
			status, created_at, last_activity, message_count
		) VALUES (?, ?, ?, ?, ?, 'active', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 1)
	`

	// Get the subject from the original email
	subject, err := r.getEmailSubject(ctx, emailID)
	if err != nil {
		subject = "Secure Email Conversation"
	}

	// Get internal user ID from the link
	internalUserID, err := r.getInternalUserID(ctx, linkID)
	if err != nil {
		return "", fmt.Errorf("failed to get internal user ID: %w", err)
	}

	_, err = r.db.ExecContext(ctx, insertQuery, chainID, linkID, internalUserID, externalEmail, subject)
	if err != nil {
		return "", fmt.Errorf("failed to create email chain: %w", err)
	}

	return chainID, nil
}

// getEmailSubject gets the subject of an email
func (r *ReplyService) getEmailSubject(ctx context.Context, emailID string) (string, error) {
	query := `SELECT subject FROM emails WHERE id = ?`
	var subject string
	err := r.db.QueryRowContext(ctx, query, emailID).Scan(&subject)
	return subject, err
}

// getInternalUserID gets the internal user ID from a secure link
func (r *ReplyService) getInternalUserID(ctx context.Context, linkID string) (string, error) {
	query := `SELECT sender_id FROM secure_links WHERE link_id = ?`
	var senderID string
	err := r.db.QueryRowContext(ctx, query, linkID).Scan(&senderID)
	return senderID, err
}

// storeSecureReply stores a secure reply in the database
func (r *ReplyService) storeSecureReply(ctx context.Context, reply *SecureReply) error {
	query := `
		INSERT INTO secure_replies (
			id, link_id, chain_id, subject, body, sender_email, recipient_email,
			ip_address, user_agent, created_at, processed_at, status, internal_email_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		reply.ID, reply.LinkID, reply.ChainID, reply.Subject, reply.Body,
		reply.SenderEmail, reply.RecipientEmail, reply.IPAddress, reply.UserAgent,
		reply.CreatedAt, reply.ProcessedAt, reply.Status, reply.InternalEmailID,
	)

	return err
}

// addMessageToChain adds a message to an email chain
func (r *ReplyService) addMessageToChain(ctx context.Context, chainID string, reply *SecureReply) error {
	query := `
		INSERT INTO chain_messages (
			id, chain_id, message_type, subject, body, sender_email, sender_type,
			created_at, link_id, email_id
		) VALUES (?, ?, 'reply', ?, ?, ?, 'external', CURRENT_TIMESTAMP, ?, NULL)
	`

	messageID := r.generateMessageID()
	_, err := r.db.ExecContext(ctx, query,
		messageID, chainID, reply.Subject, reply.Body, reply.SenderEmail,
		reply.LinkID,
	)

	if err != nil {
		return err
	}

	// Update message count in chain
	updateQuery := `UPDATE email_chains SET message_count = message_count + 1 WHERE id = ?`
	_, err = r.db.ExecContext(ctx, updateQuery, chainID)

	return err
}

// getOriginalEmail gets the original email information
func (r *ReplyService) getOriginalEmail(ctx context.Context, linkID string) (*securelinks.EmailContent, error) {
	query := `
		SELECT e.subject, e.body, e.sender_name, e.sender_email, e.created_at
		FROM emails e
		JOIN secure_links sl ON e.id = sl.email_id
		WHERE sl.link_id = ?
	`

	var content securelinks.EmailContent
	err := r.db.QueryRowContext(ctx, query, linkID).Scan(
		&content.Subject, &content.Body, &content.SenderName,
		&content.SenderEmail, &content.SentAt,
	)
	if err != nil {
		return nil, err
	}

	return &content, nil
}

// createInternalEmail creates an internal email from the reply
func (r *ReplyService) createInternalEmail(ctx context.Context, reply *SecureReply, originalEmail *securelinks.EmailContent) (string, error) {
	// Generate internal email ID
	internalEmailID := r.generateEmailID()

	// Create the internal email
	query := `
		INSERT INTO emails (
			id, sender_email, sender_name, recipient_email, subject, body,
			created_at, is_external_reply, external_reply_id
		) VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, 1, ?)
	`

	// Use the external sender's email as the sender
	senderName := reply.SenderEmail
	if reply.SenderEmail == "" {
		senderName = "External User"
	}

	_, err := r.db.ExecContext(ctx, query,
		internalEmailID, reply.SenderEmail, senderName, reply.RecipientEmail,
		reply.Subject, reply.Body, reply.ID,
	)

	if err != nil {
		return "", fmt.Errorf("failed to create internal email: %w", err)
	}

	return internalEmailID, nil
}

// generateReplyID generates a unique reply ID
func (r *ReplyService) generateReplyID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}

// generateChainID generates a unique chain ID
func (r *ReplyService) generateChainID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}

// generateMessageID generates a unique message ID
func (r *ReplyService) generateMessageID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}

// generateEmailID generates a unique email ID
func (r *ReplyService) generateEmailID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}
