package viewer

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"secure-email-mvp/pkg/securelinks"
)

// =============================================================================
// SECURE VIEWER SERVICE
// =============================================================================

// ViewerService handles secure viewing of emails for external users
type ViewerService struct {
	db *sql.DB
}

// ViewSession represents a secure viewing session
type ViewSession struct {
	ID           string     `json:"id" db:"id"`
	LinkID       string     `json:"link_id" db:"link_id"`
	IPAddress    string     `json:"ip_address" db:"ip_address"`
	UserAgent    string     `json:"user_agent" db:"user_agent"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	ExpiresAt    time.Time  `json:"expires_at" db:"expires_at"`
	IsActive     bool       `json:"is_active" db:"is_active"`
	EmailViewed  bool       `json:"email_viewed" db:"email_viewed"`
	ViewedAt     *time.Time `json:"viewed_at,omitempty" db:"viewed_at"`
	SessionToken string     `json:"session_token" db:"session_token"`
}

// EmailView represents the sanitized email content for external viewing
type EmailView struct {
	Subject      string       `json:"subject"`
	Body         string       `json:"body"`
	SenderName   string       `json:"sender_name"`
	SenderEmail  string       `json:"sender_email"`
	SentAt       time.Time    `json:"sent_at"`
	Attachments  []Attachment `json:"attachments"`
	SecurityInfo SecurityInfo `json:"security_info"`
}

// Attachment represents a secure attachment
type Attachment struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Size        int64  `json:"size"`
	ContentType string `json:"content_type"`
	SecureURL   string `json:"secure_url"`
}

// SecurityInfo provides security context to external users
type SecurityInfo struct {
	IsSecure       bool       `json:"is_secure"`
	EncryptionType string     `json:"encryption_type"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	ReadOnce       bool       `json:"read_once"`
	AutoDestruct   bool       `json:"auto_destruct"`
}

// CreateViewSessionRequest represents a request to create a view session
type CreateViewSessionRequest struct {
	LinkID    string `json:"link_id" validate:"required"`
	IPAddress string `json:"ip_address" validate:"required"`
	UserAgent string `json:"user_agent"`
}

// CreateViewSessionResponse represents the response to creating a view session
type CreateViewSessionResponse struct {
	Success      bool         `json:"success"`
	SessionToken string       `json:"session_token,omitempty"`
	ExpiresAt    time.Time    `json:"expires_at,omitempty"`
	Error        string       `json:"error,omitempty"`
	SecurityInfo SecurityInfo `json:"security_info,omitempty"`
}

// GetEmailViewRequest represents a request to get email content
type GetEmailViewRequest struct {
	SessionToken string `json:"session_token" validate:"required"`
}

// GetEmailViewResponse represents the response to getting email content
type GetEmailViewResponse struct {
	Success   bool      `json:"success"`
	EmailView *EmailView `json:"email_view,omitempty"`
	Error     string    `json:"error,omitempty"`
}

// NewViewerService creates a new secure viewer service
func NewViewerService(db *sql.DB) *ViewerService {
	return &ViewerService{
		db: db,
	}
}

// CreateViewSession creates a new secure viewing session
func (v *ViewerService) CreateViewSession(ctx context.Context, req CreateViewSessionRequest) (*CreateViewSessionResponse, error) {
	// Validate that the link exists and is active
	link, err := v.getSecureLink(ctx, req.LinkID)
	if err != nil {
		return &CreateViewSessionResponse{
			Success: false,
			Error:   "Invalid or expired secure link",
		}, fmt.Errorf("failed to get secure link: %w", err)
	}

	// Check if link is expired
	if link.ExpiresAt != nil && time.Now().After(*link.ExpiresAt) {
		return &CreateViewSessionResponse{
			Success: false,
			Error:   "Secure link has expired",
		}, fmt.Errorf("secure link expired")
	}

	// Check if link is destroyed
	if link.Status == "destroyed" {
		return &CreateViewSessionResponse{
			Success: false,
			Error:   "Secure link has been destroyed",
		}, fmt.Errorf("secure link destroyed")
	}

	// Check if link is read-once and already consumed
	if link.SecuritySettings.ReadOnce && link.Status == "destroyed" {
		return &CreateViewSessionResponse{
			Success: false,
			Error:   "Secure link has already been viewed",
		}, fmt.Errorf("secure link already consumed")
	}

	// Generate session token
	sessionToken := v.generateSessionToken()

	// Create view session
	session := &ViewSession{
		ID:           v.generateSessionID(),
		LinkID:       req.LinkID,
		IPAddress:    req.IPAddress,
		UserAgent:    req.UserAgent,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(30 * time.Minute), // Sessions expire in 30 minutes
		IsActive:     true,
		EmailViewed:  false,
		SessionToken: sessionToken,
	}

	// Store session in database
	if err := v.storeViewSession(ctx, session); err != nil {
		return &CreateViewSessionResponse{
			Success: false,
			Error:   "Failed to create viewing session",
		}, fmt.Errorf("failed to store view session: %w", err)
	}

	// Create security info for response
	securityInfo := SecurityInfo{
		IsSecure:       true,
		EncryptionType: "AES-256-GCM",
		ReadOnce:       link.SecuritySettings.ReadOnce,
		AutoDestruct:   link.SecuritySettings.AutoDestruct,
	}

	if link.ExpiresAt != nil {
		securityInfo.ExpiresAt = link.ExpiresAt
	}

	return &CreateViewSessionResponse{
		Success:      true,
		SessionToken: sessionToken,
		ExpiresAt:    session.ExpiresAt,
		SecurityInfo: securityInfo,
	}, nil
}

// GetEmailView retrieves sanitized email content for external viewing
func (v *ViewerService) GetEmailView(ctx context.Context, req GetEmailViewRequest) (*GetEmailViewResponse, error) {
	// Get view session
	session, err := v.getViewSession(ctx, req.SessionToken)
	if err != nil {
		return &GetEmailViewResponse{
			Success: false,
			Error:   "Invalid or expired session",
		}, fmt.Errorf("failed to get view session: %w", err)
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		return &GetEmailViewResponse{
			Success: false,
			Error:   "Viewing session has expired",
		}, fmt.Errorf("view session expired")
	}

	// Check if session is active
	if !session.IsActive {
		return &GetEmailViewResponse{
			Success: false,
			Error:   "Viewing session is no longer active",
		}, fmt.Errorf("view session inactive")
	}

	// Get secure link
	link, err := v.getSecureLink(ctx, session.LinkID)
	if err != nil {
		return &GetEmailViewResponse{
			Success: false,
			Error:   "Invalid secure link",
		}, fmt.Errorf("failed to get secure link: %w", err)
	}

	// Get email content
	emailContent, err := v.getEmailContent(ctx, link.EmailID)
	if err != nil {
		return &GetEmailViewResponse{
			Success: false,
			Error:   "Failed to retrieve email content",
		}, fmt.Errorf("failed to get email content: %w", err)
	}

	// Sanitize email content
	sanitizedContent := v.sanitizeEmailContent(emailContent, link.SecuritySettings.StripMetadata)

	// Create email view
	emailView := &EmailView{
		Subject:     sanitizedContent.Subject,
		Body:        sanitizedContent.Body,
		SenderName:  sanitizedContent.SenderName,
		SenderEmail: sanitizedContent.SenderEmail,
		SentAt:      sanitizedContent.SentAt,
		Attachments: []Attachment{}, // TODO: Implement attachment retrieval
		SecurityInfo: SecurityInfo{
			IsSecure:       true,
			EncryptionType: "AES-256-GCM",
			ReadOnce:       link.SecuritySettings.ReadOnce,
			AutoDestruct:   link.SecuritySettings.AutoDestruct,
		},
	}

	if link.ExpiresAt != nil {
		emailView.SecurityInfo.ExpiresAt = link.ExpiresAt
	}

	// Record that email has been viewed
	if err := v.RecordEmailView(ctx, req.SessionToken); err != nil {
		log.Printf("Warning: Failed to record email view: %v", err)
		// Don't fail the request, just log the warning
	}

	return &GetEmailViewResponse{
		Success:   true,
		EmailView: emailView,
	}, nil
}

// RecordEmailView records that an email has been viewed
func (v *ViewerService) RecordEmailView(ctx context.Context, sessionToken string) error {
	// Update view session to mark email as viewed
	query := `
		UPDATE link_view_sessions 
		SET email_viewed = 1, viewed_at = CURRENT_TIMESTAMP 
		WHERE session_token = ?
	`

	_, err := v.db.ExecContext(ctx, query, sessionToken)
	if err != nil {
		return fmt.Errorf("failed to update view session: %w", err)
	}

	// Get the session to find the link ID
	session, err := v.getViewSession(ctx, sessionToken)
	if err != nil {
		return fmt.Errorf("failed to get view session: %w", err)
	}

	// Get the secure link
	link, err := v.getSecureLink(ctx, session.LinkID)
	if err != nil {
		return fmt.Errorf("failed to get secure link: %w", err)
	}

	// Handle read-once functionality
	if link.SecuritySettings.ReadOnce {
		query := `UPDATE secure_links SET read_once_consumed = 1 WHERE link_id = ?`
		if _, err := v.db.ExecContext(ctx, query, session.LinkID); err != nil {
			return fmt.Errorf("failed to mark link as consumed: %w", err)
		}
	}

	// Handle auto-destruct functionality
	if link.SecuritySettings.AutoDestruct {
		query := `UPDATE secure_links SET status = 'destroyed' WHERE link_id = ?`
		if _, err := v.db.ExecContext(ctx, query, session.LinkID); err != nil {
			return fmt.Errorf("failed to destroy link: %w", err)
		}
	}

	return nil
}

// Helper methods

// getSecureLink retrieves a secure link from the database
func (v *ViewerService) getSecureLink(ctx context.Context, linkID string) (*securelinks.SecureLink, error) {
	query := `
		SELECT link_id, email_id, recipient_email, created_at, expires_at, status,
		       security_settings
		FROM secure_links
		WHERE link_id = ?
	`

	var link securelinks.SecureLink
	var securitySettingsJSON string

	err := v.db.QueryRowContext(ctx, query, linkID).Scan(
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

// getEmailContent retrieves email content from the database
func (v *ViewerService) getEmailContent(ctx context.Context, emailID string) (*securelinks.EmailContent, error) {
	query := `
		SELECT subject, body, sender_name, sender_email, created_at
		FROM emails
		WHERE id = ?
	`

	var content securelinks.EmailContent
	err := v.db.QueryRowContext(ctx, query, emailID).Scan(
		&content.Subject, &content.Body, &content.SenderName,
		&content.SenderEmail, &content.SentAt,
	)
	if err != nil {
		return nil, err
	}

	return &content, nil
}

// sanitizeEmailContent removes sensitive metadata from email content
func (v *ViewerService) sanitizeEmailContent(content *securelinks.EmailContent, stripMetadata bool) *securelinks.EmailContent {
	if !stripMetadata {
		return content
	}

	// Create a copy to avoid modifying the original
	sanitized := *content

	// Remove email addresses from body (replace with [email])
	sanitized.Body = v.removeEmailAddresses(sanitized.Body)

	// Remove phone numbers from body (replace with [phone])
	sanitized.Body = v.removePhoneNumbers(sanitized.Body)

	// Remove specific dates (replace with [date])
	sanitized.Body = v.removeSpecificDates(sanitized.Body)

	return &sanitized
}

// removeEmailAddresses removes email addresses from text
func (v *ViewerService) removeEmailAddresses(text string) string {
	// Simple regex-like replacement for email addresses
	// In production, use proper regex
	words := strings.Fields(text)
	for i, word := range words {
		if strings.Contains(word, "@") {
			words[i] = "[email]"
		}
	}
	return strings.Join(words, " ")
}

// removePhoneNumbers removes phone numbers from text
func (v *ViewerService) removePhoneNumbers(text string) string {
	// Simple replacement for common phone number patterns
	// In production, use proper regex
	text = strings.ReplaceAll(text, "+1-", "[phone]")
	text = strings.ReplaceAll(text, "555-", "[phone]")
	return text
}

// removeSpecificDates removes specific dates from text
func (v *ViewerService) removeSpecificDates(text string) string {
	// Simple replacement for year patterns
	// In production, use proper regex
	text = strings.ReplaceAll(text, "2024-", "[year]-")
	text = strings.ReplaceAll(text, "2025-", "[year]-")
	return text
}

// storeViewSession stores a view session in the database
func (v *ViewerService) storeViewSession(ctx context.Context, session *ViewSession) error {
	query := `
		INSERT INTO link_view_sessions (
			id, link_id, ip_address, user_agent, created_at, expires_at,
			is_active, email_viewed, viewed_at, session_token
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := v.db.ExecContext(ctx, query,
		session.ID, session.LinkID, session.IPAddress, session.UserAgent,
		session.CreatedAt, session.ExpiresAt, session.IsActive,
		session.EmailViewed, session.ViewedAt, session.SessionToken,
	)

	return err
}

// getViewSession retrieves a view session from the database
func (v *ViewerService) getViewSession(ctx context.Context, sessionToken string) (*ViewSession, error) {
	query := `
		SELECT id, link_id, ip_address, user_agent, created_at, expires_at,
		       is_active, email_viewed, viewed_at, session_token
		FROM link_view_sessions
		WHERE session_token = ?
	`

	var session ViewSession
	err := v.db.QueryRowContext(ctx, query, sessionToken).Scan(
		&session.ID, &session.LinkID, &session.IPAddress, &session.UserAgent,
		&session.CreatedAt, &session.ExpiresAt, &session.IsActive,
		&session.EmailViewed, &session.ViewedAt, &session.SessionToken,
	)
	if err != nil {
		return nil, err
	}

	return &session, nil
}

// generateSessionID generates a unique session ID
func (v *ViewerService) generateSessionID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}

// generateSessionToken generates a unique session token
func (v *ViewerService) generateSessionToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}
