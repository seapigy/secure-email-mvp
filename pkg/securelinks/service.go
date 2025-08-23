package securelinks

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"secure-email-mvp/pkg/securelinks/geolocation"
	"secure-email-mvp/pkg/securelinks/mfa"
	"secure-email-mvp/pkg/storage"
)

// =============================================================================
// SECURE LINKS SERVICE
// =============================================================================

// Service provides secure link functionality for external email access
type Service struct {
	db       *sql.DB
	r2Client *storage.R2Client
	baseURL  string
	auditSvc AuditService
}

// AuditService interface for audit logging
type AuditService interface {
	LogEvent(ctx context.Context, eventType string, details map[string]interface{}) error
}

// NewService creates a new secure links service
func NewService(db *sql.DB, r2Client *storage.R2Client, baseURL string, auditSvc AuditService) *Service {
	return &Service{
		db:       db,
		r2Client: r2Client,
		baseURL:  baseURL,
		auditSvc: auditSvc,
	}
}

// =============================================================================
// LINK GENERATION & CREATION
// =============================================================================

// CreateSecureLink creates a new secure link for external email access
func (s *Service) CreateSecureLink(ctx context.Context, req CreateSecureLinkRequest, senderID string) (*CreateSecureLinkResponse, error) {
	// Validate request
	if err := s.validateCreateRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Verify email exists and sender has access
	if err := s.verifyEmailAccess(ctx, req.EmailID, senderID); err != nil {
		return nil, fmt.Errorf("email access verification failed: %w", err)
	}

	// Generate secure link ID
	linkID, err := s.generateSecureLinkID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate link ID: %w", err)
	}

	// Create secure link record
	link := &SecureLink{
		LinkID:           linkID,
		EmailID:          req.EmailID,
		RecipientEmail:   req.RecipientEmail,
		SenderID:         senderID,
		SecuritySettings: req.SecuritySettings,
		CreatedAt:        time.Now(),
		Status:           LinkStatusActive,
	}

	// Set expiration if specified
	if req.SecuritySettings.ExpiresAt != nil {
		expiresAt := time.Unix(*req.SecuritySettings.ExpiresAt, 0)
		link.ExpiresAt = &expiresAt
	}

	// Store link in database
	if err := s.storeSecureLink(ctx, link); err != nil {
		return nil, fmt.Errorf("failed to store secure link: %w", err)
	}

	// Log audit event
	s.logAuditEvent(ctx, link.LinkID, AuditEventCreated, map[string]interface{}{
		"email_id":          link.EmailID,
		"recipient_email":   link.RecipientEmail,
		"sender_id":         link.SenderID,
		"security_settings": link.SecuritySettings,
	})

	// Build response
	response := &CreateSecureLinkResponse{
		LinkID:       link.LinkID,
		SecureURL:    s.buildSecureURL(link.LinkID),
		ExpiresAt:    link.ExpiresAt,
		CreatedAt:    link.CreatedAt,
		SecurityInfo: s.buildSecurityInfo(link.SecuritySettings),
	}

	return response, nil
}

// =============================================================================
// LINK ACCESS & VALIDATION
// =============================================================================

// AccessSecureLink handles secure link access with all security checks
func (s *Service) AccessSecureLink(ctx context.Context, req AccessSecureLinkRequest) (*AccessSecureLinkResponse, error) {
	// Retrieve secure link
	link, err := s.getSecureLink(ctx, req.LinkID)
	if err != nil {
		return nil, fmt.Errorf("link not found or access denied: %w", err)
	}

	// Log access attempt
	s.logAuditEvent(ctx, link.LinkID, AuditEventAccessed, map[string]interface{}{
		"ip_address": req.IPAddress,
		"user_agent": req.UserAgent,
	})

	// Check if link is active
	if !link.Status.IsActive() {
		return s.handleInactiveLink(link, req)
	}

	// Check expiration
	if link.ExpiresAt != nil && time.Now().After(*link.ExpiresAt) {
		return s.handleExpiredLink(link, req)
	}

	// Check time lock
	if link.SecuritySettings.TimeLock && link.SecuritySettings.TimeLockUntil != nil {
		if time.Now().Unix() < *link.SecuritySettings.TimeLockUntil {
			return s.handleTimeLockedLink(link, req)
		}
	}

	// Check if link is locked due to failed attempts
	if link.LockoutUntil != nil && time.Now().Before(*link.LockoutUntil) {
		return s.handleLockedLink(link, req)
	}

	// Validate password if required
	if link.SecuritySettings.RequirePassword {
		if req.Password == nil {
			return s.handlePasswordRequired(link, req)
		}

		if err := s.validatePassword(link, *req.Password); err != nil {
			return s.handlePasswordFailure(link, req, err)
		}
	}

	// Validate MFA if required
	if link.SecuritySettings.RequireMFA {
		if req.MFACode == nil {
			return s.handleMFARequired(link, req)
		}

		if err := s.validateMFA(link, *req.MFACode); err != nil {
			return s.handleMFAFailure(link, req, err)
		}
	}

	// Check geolocation restrictions
	if link.SecuritySettings.GeolocationRestriction {
		if err := s.validateGeolocation(link, req.IPAddress); err != nil {
			return s.handleGeolocationBlocked(link, req, err)
		}
	}

	// All security checks passed - retrieve email content
	emailContent, err := s.retrieveEmailContent(ctx, link)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve email content: %w", err)
	}

	// Update access count and last accessed time
	if err := s.updateAccessStats(ctx, link.LinkID); err != nil {
		log.Printf("Warning: Failed to update access stats for link %s: %v", link.LinkID, err)
	}

	// Handle read-once and auto-destruct
	if link.SecuritySettings.ReadOnce || link.SecuritySettings.AutoDestruct {
		if err := s.handleReadOnceOrAutoDestruct(ctx, link); err != nil {
			log.Printf("Warning: Failed to handle read-once/auto-destruct for link %s: %v", link.LinkID, err)
		}
	}

	// Log successful access
	s.logAuditEvent(ctx, link.LinkID, AuditEventAccessed, map[string]interface{}{
		"ip_address": req.IPAddress,
		"user_agent": req.UserAgent,
		"success":    true,
	})

	return &AccessSecureLinkResponse{
		Success:      true,
		EmailContent: emailContent,
	}, nil
}

// =============================================================================
// LINK MANAGEMENT
// =============================================================================

// RevokeSecureLink revokes a secure link (sender only)
func (s *Service) RevokeSecureLink(ctx context.Context, linkID, senderID string) error {
	// Retrieve link and verify sender ownership
	link, err := s.getSecureLink(ctx, linkID)
	if err != nil {
		return fmt.Errorf("link not found: %w", err)
	}

	if link.SenderID != senderID {
		return errors.New("unauthorized: only the sender can revoke this link")
	}

	// Update link status to revoked
	if err := s.updateLinkStatus(ctx, linkID, LinkStatusRevoked); err != nil {
		return fmt.Errorf("failed to revoke link: %w", err)
	}

	// Log revocation event
	s.logAuditEvent(ctx, linkID, AuditEventRevoked, map[string]interface{}{
		"sender_id": senderID,
		"reason":    "sender_revoked",
	})

	return nil
}

// GetSecureLinkInfo returns information about a secure link (sender only)
func (s *Service) GetSecureLinkInfo(ctx context.Context, linkID, senderID string) (*SecureLink, error) {
	link, err := s.getSecureLink(ctx, linkID)
	if err != nil {
		return nil, fmt.Errorf("link not found: %w", err)
	}

	if link.SenderID != senderID {
		return nil, errors.New("unauthorized: only the sender can view this link info")
	}

	return link, nil
}

// ListSecureLinks returns all secure links for a sender
func (s *Service) ListSecureLinks(ctx context.Context, senderID string, limit, offset int) ([]*SecureLink, error) {
	query := `
		SELECT link_id, email_id, recipient_email, sender_id, security_settings,
		       created_at, expires_at, access_count, last_accessed, status,
		       failed_attempts, last_failed_attempt, lockout_until
		FROM secure_links 
		WHERE sender_id = ? 
		ORDER BY created_at DESC 
		LIMIT ? OFFSET ?
	`

	rows, err := s.db.QueryContext(ctx, query, senderID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query secure links: %w", err)
	}
	defer rows.Close()

	var links []*SecureLink
	for rows.Next() {
		link, err := s.scanSecureLink(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan secure link: %w", err)
		}
		links = append(links, link)
	}

	return links, nil
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

// generateSecureLinkID generates a cryptographically secure link ID
func (s *Service) generateSecureLinkID() (string, error) {
	// Generate 32 random bytes
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Encode as base64 URL-safe string
	linkID := base64.URLEncoding.EncodeToString(randomBytes)

	// Remove padding characters
	linkID = linkID[:len(linkID)-2] // Remove "==" padding

	return linkID, nil
}

// buildSecureURL constructs the secure URL for a link
func (s *Service) buildSecureURL(linkID string) string {
	return fmt.Sprintf("%s/v/%s", s.baseURL, linkID)
}

// buildSecurityInfo creates security information for the response
func (s *Service) buildSecurityInfo(settings SecuritySettings) SecurityInfo {
	var timeLockUntil *time.Time
	if settings.TimeLockUntil != nil {
		t := time.Unix(*settings.TimeLockUntil, 0)
		timeLockUntil = &t
	}

	return SecurityInfo{
		RequirePassword:        settings.RequirePassword,
		RequireMFA:             settings.RequireMFA,
		MFAType:                settings.MFAType,
		TimeLock:               settings.TimeLock,
		TimeLockUntil:          timeLockUntil,
		GeolocationRestriction: settings.GeolocationRestriction,
		AllowedCountries:       settings.AllowedCountries,
		ReadOnce:               settings.ReadOnce,
		AutoDestruct:           settings.AutoDestruct,
		MaxAccessAttempts:      settings.MaxAccessAttempts,
	}
}

// validateCreateRequest validates the create secure link request
func (s *Service) validateCreateRequest(req CreateSecureLinkRequest) error {
	if req.EmailID == "" {
		return errors.New("email_id is required")
	}
	if req.RecipientEmail == "" {
		return errors.New("recipient_email is required")
	}
	// Add more validation as needed
	return nil
}

// verifyEmailAccess verifies that the sender has access to the email
func (s *Service) verifyEmailAccess(ctx context.Context, emailID, senderID string) error {
	query := `SELECT id FROM emails WHERE id = ? AND sender_id = ?`
	var id string
	err := s.db.QueryRowContext(ctx, query, emailID, senderID).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("email not found or access denied")
		}
		return fmt.Errorf("database error: %w", err)
	}
	return nil
}

// storeSecureLink stores a secure link in the database
func (s *Service) storeSecureLink(ctx context.Context, link *SecureLink) error {
	settingsJSON, err := json.Marshal(link.SecuritySettings)
	if err != nil {
		return fmt.Errorf("failed to marshal security settings: %w", err)
	}

	query := `
		INSERT INTO secure_links (
			link_id, email_id, recipient_email, sender_id, security_settings,
			created_at, expires_at, status
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = s.db.ExecContext(ctx, query,
		link.LinkID,
		link.EmailID,
		link.RecipientEmail,
		link.SenderID,
		settingsJSON,
		link.CreatedAt,
		link.ExpiresAt,
		link.Status,
	)

	if err != nil {
		return fmt.Errorf("failed to insert secure link: %w", err)
	}

	return nil
}

// getSecureLink retrieves a secure link from the database
func (s *Service) getSecureLink(ctx context.Context, linkID string) (*SecureLink, error) {
	query := `
		SELECT link_id, email_id, recipient_email, sender_id, security_settings,
		       created_at, expires_at, access_count, last_accessed, status,
		       failed_attempts, last_failed_attempt, lockout_until
		FROM secure_links 
		WHERE link_id = ?
	`

	row := s.db.QueryRowContext(ctx, query, linkID)
	return s.scanSecureLink(row)
}

// scanSecureLink scans a database row into a SecureLink struct
func (s *Service) scanSecureLink(row interface{}) (*SecureLink, error) {
	var link SecureLink
	var settingsJSON []byte

	var scanFunc func(...interface{}) error
	switch r := row.(type) {
	case *sql.Row:
		scanFunc = r.Scan
	case *sql.Rows:
		scanFunc = r.Scan
	default:
		return nil, fmt.Errorf("invalid row type")
	}

	err := scanFunc(
		&link.LinkID,
		&link.EmailID,
		&link.RecipientEmail,
		&link.SenderID,
		&settingsJSON,
		&link.CreatedAt,
		&link.ExpiresAt,
		&link.AccessCount,
		&link.LastAccessed,
		&link.Status,
		&link.FailedAttempts,
		&link.LastFailedAttempt,
		&link.LockoutUntil,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to scan secure link: %w", err)
	}

	// Unmarshal security settings
	if err := json.Unmarshal(settingsJSON, &link.SecuritySettings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal security settings: %w", err)
	}

	return &link, nil
}

// updateLinkStatus updates the status of a secure link
func (s *Service) updateLinkStatus(ctx context.Context, linkID string, status LinkStatus) error {
	query := `UPDATE secure_links SET status = ? WHERE link_id = ?`
	_, err := s.db.ExecContext(ctx, query, status, linkID)
	if err != nil {
		return fmt.Errorf("failed to update link status: %w", err)
	}
	return nil
}

// updateAccessStats updates access count and last accessed time
func (s *Service) updateAccessStats(ctx context.Context, linkID string) error {
	query := `
		UPDATE secure_links 
		SET access_count = access_count + 1, last_accessed = ? 
		WHERE link_id = ?
	`
	_, err := s.db.ExecContext(ctx, query, time.Now(), linkID)
	if err != nil {
		return fmt.Errorf("failed to update access stats: %w", err)
	}
	return nil
}

// logAuditEvent logs an audit event
func (s *Service) logAuditEvent(ctx context.Context, linkID string, eventType AuditEventType, details map[string]interface{}) {
	if s.auditSvc != nil {
		details["link_id"] = linkID
		details["event_type"] = string(eventType)
		details["timestamp"] = time.Now()

		if err := s.auditSvc.LogEvent(ctx, "secure_link_"+string(eventType), details); err != nil {
			log.Printf("Warning: Failed to log audit event for link %s: %v", linkID, err)
		}
	}
}

// =============================================================================
// PLACEHOLDER FUNCTIONS FOR PHASE 2 IMPLEMENTATION
// =============================================================================

// These functions will be implemented in Phase 2

func (s *Service) handleInactiveLink(link *SecureLink, req AccessSecureLinkRequest) (*AccessSecureLinkResponse, error) {
	// TODO: Implement in Phase 2
	return &AccessSecureLinkResponse{
		Success: false,
		Error: &AccessError{
			Code:    "LINK_INACTIVE",
			Message: "This secure link is no longer active",
		},
	}, nil
}

func (s *Service) handleExpiredLink(link *SecureLink, req AccessSecureLinkRequest) (*AccessSecureLinkResponse, error) {
	// TODO: Implement in Phase 2
	return &AccessSecureLinkResponse{
		Success: false,
		Error: &AccessError{
			Code:    "LINK_EXPIRED",
			Message: "This secure link has expired",
		},
	}, nil
}

func (s *Service) handleTimeLockedLink(link *SecureLink, req AccessSecureLinkRequest) (*AccessSecureLinkResponse, error) {
	// Check if time lock is enabled and not yet unlocked
	if link.SecuritySettings.TimeLock && link.SecuritySettings.TimeLockUntil != nil {
		now := time.Now().Unix()
		if now < *link.SecuritySettings.TimeLockUntil {
			// Calculate remaining time
			remaining := *link.SecuritySettings.TimeLockUntil - now
			remainingTime := time.Duration(remaining) * time.Second

			return &AccessSecureLinkResponse{
				Success: false,
				Error: &AccessError{
					Code:    "TIME_LOCKED",
					Message: fmt.Sprintf("This secure link is time-locked. Unlocks in %s", remainingTime.String()),
				},
			}, nil
		}
	}

	// If we reach here, the time lock has expired or is not enabled
	return &AccessSecureLinkResponse{
		Success: false,
		Error: &AccessError{
			Code:    "TIME_LOCK_EXPIRED",
			Message: "Time lock has expired",
		},
	}, nil
}

func (s *Service) handleLockedLink(link *SecureLink, req AccessSecureLinkRequest) (*AccessSecureLinkResponse, error) {
	// TODO: Implement in Phase 2
	return &AccessSecureLinkResponse{
		Success: false,
		Error: &AccessError{
			Code:    "LINK_LOCKED",
			Message: "This secure link is temporarily locked",
		},
	}, nil
}

func (s *Service) handlePasswordRequired(link *SecureLink, req AccessSecureLinkRequest) (*AccessSecureLinkResponse, error) {
	// TODO: Implement in Phase 2
	return &AccessSecureLinkResponse{
		Success: false,
		SecurityCheck: &SecurityCheck{
			Type:     "password",
			Required: true,
			Message:  "Password required to access this secure link",
		},
	}, nil
}

func (s *Service) validatePassword(link *SecureLink, password string) error {
	// TODO: Implement in Phase 2
	return nil
}

func (s *Service) handlePasswordFailure(link *SecureLink, req AccessSecureLinkRequest, err error) (*AccessSecureLinkResponse, error) {
	// TODO: Implement in Phase 2
	return &AccessSecureLinkResponse{
		Success: false,
		Error: &AccessError{
			Code:    "PASSWORD_FAILED",
			Message: "Invalid password",
		},
	}, nil
}

func (s *Service) handleMFARequired(link *SecureLink, req AccessSecureLinkRequest) (*AccessSecureLinkResponse, error) {
	// TODO: Implement in Phase 2
	return &AccessSecureLinkResponse{
		Success: false,
		SecurityCheck: &SecurityCheck{
			Type:     "mfa",
			Required: true,
			Message:  "Multi-factor authentication required",
		},
	}, nil
}

func (s *Service) validateMFA(link *SecureLink, mfaCode string) error {
	// If MFA is not required, allow access
	if !link.SecuritySettings.RequireMFA {
		return nil
	}

	// Create MFA service
	mfaService := mfa.NewExternalMFAService(s.db)

	// Create verification request
	req := mfa.MFAVerificationRequest{
		SessionID: link.LinkID, // Use link ID as session ID for simplicity
		Code:      mfaCode,
	}

	// Verify MFA code
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := mfaService.VerifyMFA(ctx, req)
	if err != nil {
		return fmt.Errorf("MFA verification failed: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("invalid MFA code: %s", result.Error)
	}

	return nil
}

func (s *Service) handleMFAFailure(link *SecureLink, req AccessSecureLinkRequest, err error) (*AccessSecureLinkResponse, error) {
	// TODO: Implement in Phase 2
	return &AccessSecureLinkResponse{
		Success: false,
		Error: &AccessError{
			Code:    "MFA_FAILED",
			Message: "Invalid MFA code",
		},
	}, nil
}

func (s *Service) validateGeolocation(link *SecureLink, ipAddress string) error {
	// If geolocation restrictions are not enabled, allow access
	if !link.SecuritySettings.GeolocationRestriction {
		return nil
	}

	// Create geolocation verification service
	geoService := geolocation.NewGeolocationVerificationService()

	// Create geolocation restriction from link settings
	restriction := geolocation.GeolocationRestriction{
		Enabled:          link.SecuritySettings.GeolocationRestriction,
		AllowedCountries: link.SecuritySettings.AllowedCountries,
		AllowedCities:    link.SecuritySettings.AllowedCities,
		BlockedCountries: []string{}, // TODO: Add blocked countries to security settings
		BlockedCities:    []string{}, // TODO: Add blocked cities to security settings
	}

	// Verify location
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := geoService.VerifyLocation(ctx, ipAddress, restriction)
	if err != nil {
		return fmt.Errorf("geolocation verification failed: %w", err)
	}

	if !result.Allowed {
		return fmt.Errorf("access denied from your location: %s", result.Reason)
	}

	return nil
}

func (s *Service) handleGeolocationBlocked(link *SecureLink, req AccessSecureLinkRequest, err error) (*AccessSecureLinkResponse, error) {
	// TODO: Implement in Phase 2
	return &AccessSecureLinkResponse{
		Success: false,
		Error: &AccessError{
			Code:    "GEO_BLOCKED",
			Message: "Access denied from your location",
		},
	}, nil
}

func (s *Service) retrieveEmailContent(ctx context.Context, link *SecureLink) (*EmailContent, error) {
	// Query the emails table to get the email content
	query := `
		SELECT e.subject, e.body, e.sender_name, e.sender_email, e.created_at, e.blob_id
		FROM emails e
		WHERE e.id = ?
	`

	var subject, body, senderName, senderEmail, blobID string
	var createdAt time.Time

	err := s.db.QueryRowContext(ctx, query, link.EmailID).Scan(
		&subject, &body, &senderName, &senderEmail, &createdAt, &blobID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve email content: %w", err)
	}

	// If there's a blob ID, try to retrieve encrypted content from R2
	if blobID != "" && s.r2Client != nil {
		// TODO: Implement R2 content retrieval and decryption
		// For now, use the database content
		log.Printf("Note: R2 content retrieval not yet implemented for blob %s", blobID)
	}

	// Strip metadata if required
	if link.SecuritySettings.StripMetadata {
		subject = s.stripMetadata(subject)
		body = s.stripMetadata(body)
	}

	return &EmailContent{
		Subject:     subject,
		Body:        body,
		SenderName:  senderName,
		SenderEmail: senderEmail,
		SentAt:      createdAt,
		Attachments: []Attachment{}, // TODO: Implement attachment retrieval
	}, nil
}

// stripMetadata removes sensitive metadata from content
func (s *Service) stripMetadata(content string) string {
	// TODO: Implement comprehensive metadata stripping
	// For now, just return the content as-is
	return content
}

func (s *Service) handleReadOnceOrAutoDestruct(ctx context.Context, link *SecureLink) error {
	// Handle read-once functionality
	if link.SecuritySettings.ReadOnce {
		// Mark the link as consumed
		query := `UPDATE secure_links SET read_once_consumed = 1 WHERE link_id = ?`
		if _, err := s.db.ExecContext(ctx, query, link.LinkID); err != nil {
			return fmt.Errorf("failed to mark link as consumed: %w", err)
		}

		// Log the read-once consumption
		s.logAuditEvent(ctx, link.LinkID, AuditEventReadOnceConsumed, map[string]interface{}{
			"email_id": link.EmailID,
			"consumed": true,
		})
	}

	// Handle auto-destruct functionality
	if link.SecuritySettings.AutoDestruct {
		// Mark the link as destroyed
		query := `UPDATE secure_links SET status = 'destroyed' WHERE link_id = ?`
		if _, err := s.db.ExecContext(ctx, query, link.LinkID); err != nil {
			return fmt.Errorf("failed to destroy link: %w", err)
		}

		// Log the auto-destruct event
		s.logAuditEvent(ctx, link.LinkID, AuditEventAutoDestructTriggered, map[string]interface{}{
			"email_id":  link.EmailID,
			"destroyed": true,
		})
	}

	return nil
}
