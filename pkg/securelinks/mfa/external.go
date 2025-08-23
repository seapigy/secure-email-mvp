package mfa

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"time"

	"github.com/pquerna/otp/totp"
)

// =============================================================================
// EXTERNAL MFA SERVICE
// =============================================================================

// ExternalMFAService handles MFA for external secure link users
type ExternalMFAService struct {
	db *sql.DB
}

// MFASession represents an MFA session for external users
type MFASession struct {
	ID          string     `json:"id" db:"id"`
	LinkID      string     `json:"link_id" db:"link_id"`
	MFAType     string     `json:"mfa_type" db:"mfa_type"` // "totp", "email", "sms"
	Secret      string     `json:"secret,omitempty" db:"secret"`
	Code        string     `json:"code,omitempty" db:"code"`
	Email       string     `json:"email,omitempty" db:"email"`
	Phone       string     `json:"phone,omitempty" db:"phone"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	ExpiresAt   time.Time  `json:"expires_at" db:"expires_at"`
	Verified    bool       `json:"verified" db:"verified"`
	Attempts    int        `json:"attempts" db:"attempts"`
	MaxAttempts int        `json:"max_attempts" db:"max_attempts"`
	LockedUntil *time.Time `json:"locked_until,omitempty" db:"locked_until"`
}

// MFARequest represents an MFA request
type MFARequest struct {
	LinkID  string `json:"link_id" validate:"required"`
	MFAType string `json:"mfa_type" validate:"required"` // "totp", "email", "sms"
	Email   string `json:"email,omitempty"`
	Phone   string `json:"phone,omitempty"`
}

// MFAVerificationRequest represents an MFA verification request
type MFAVerificationRequest struct {
	SessionID string `json:"session_id" validate:"required"`
	Code      string `json:"code" validate:"required"`
}

// MFAResponse represents an MFA response
type MFAResponse struct {
	Success      bool       `json:"success"`
	SessionID    string     `json:"session_id,omitempty"`
	Message      string     `json:"message,omitempty"`
	Error        string     `json:"error,omitempty"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	AttemptsLeft int        `json:"attempts_left,omitempty"`
}

// NewExternalMFAService creates a new external MFA service
func NewExternalMFAService(db *sql.DB) *ExternalMFAService {
	return &ExternalMFAService{
		db: db,
	}
}

// InitiateMFA initiates an MFA session for external users
func (m *ExternalMFAService) InitiateMFA(ctx context.Context, req MFARequest) (*MFAResponse, error) {
	// Validate MFA type
	if !m.isValidMFAType(req.MFAType) {
		return &MFAResponse{
			Success: false,
			Error:   "Invalid MFA type",
		}, fmt.Errorf("invalid MFA type: %s", req.MFAType)
	}

	// Generate session ID
	sessionID := m.generateSessionID()

	// Create MFA session based on type
	var session *MFASession
	var err error

	switch req.MFAType {
	case "totp":
		session, err = m.createTOTPSession(ctx, sessionID, req.LinkID)
	case "email":
		session, err = m.createEmailSession(ctx, sessionID, req.LinkID, req.Email)
	case "sms":
		session, err = m.createSMSSession(ctx, sessionID, req.LinkID, req.Phone)
	default:
		return &MFAResponse{
			Success: false,
			Error:   "Unsupported MFA type",
		}, fmt.Errorf("unsupported MFA type: %s", req.MFAType)
	}

	if err != nil {
		return &MFAResponse{
			Success: false,
			Error:   "Failed to create MFA session",
		}, fmt.Errorf("failed to create MFA session: %w", err)
	}

	// Store session in database
	if err := m.storeMFASession(ctx, session); err != nil {
		return &MFAResponse{
			Success: false,
			Error:   "Failed to store MFA session",
		}, fmt.Errorf("failed to store MFA session: %w", err)
	}

	// Send MFA code if needed
	if req.MFAType == "email" || req.MFAType == "sms" {
		if err := m.sendMFACode(ctx, session); err != nil {
			log.Printf("Warning: Failed to send MFA code: %v", err)
			// Don't fail the request, just log the warning
		}
	}

	return &MFAResponse{
		Success:   true,
		SessionID: sessionID,
		Message:   m.getMFAMessage(req.MFAType),
		ExpiresAt: &session.ExpiresAt,
	}, nil
}

// VerifyMFA verifies an MFA code
func (m *ExternalMFAService) VerifyMFA(ctx context.Context, req MFAVerificationRequest) (*MFAResponse, error) {
	// Get MFA session
	session, err := m.getMFASession(ctx, req.SessionID)
	if err != nil {
		return &MFAResponse{
			Success: false,
			Error:   "Invalid session",
		}, fmt.Errorf("failed to get MFA session: %w", err)
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		return &MFAResponse{
			Success: false,
			Error:   "MFA session expired",
		}, fmt.Errorf("MFA session expired")
	}

	// Check if session is locked
	if session.LockedUntil != nil && time.Now().Before(*session.LockedUntil) {
		return &MFAResponse{
			Success: false,
			Error:   "MFA session temporarily locked",
		}, fmt.Errorf("MFA session temporarily locked")
	}

	// Check if max attempts exceeded
	if session.Attempts >= session.MaxAttempts {
		// Lock session for 15 minutes
		lockUntil := time.Now().Add(15 * time.Minute)
		session.LockedUntil = &lockUntil
		session.Attempts = 0

		if err := m.updateMFASession(ctx, session); err != nil {
			log.Printf("Warning: Failed to update MFA session: %v", err)
		}

		return &MFAResponse{
			Success: false,
			Error:   "Too many failed attempts. Session locked for 15 minutes.",
		}, fmt.Errorf("too many failed attempts")
	}

	// Verify code based on MFA type
	var isValid bool
	switch session.MFAType {
	case "totp":
		isValid = m.verifyTOTP(session.Secret, req.Code)
	case "email", "sms":
		isValid = m.verifyCode(session.Code, req.Code)
	default:
		return &MFAResponse{
			Success: false,
			Error:   "Invalid MFA type",
		}, fmt.Errorf("invalid MFA type: %s", session.MFAType)
	}

	// Increment attempts
	session.Attempts++

	if isValid {
		// Mark session as verified
		session.Verified = true
		session.Attempts = 0
		session.LockedUntil = nil

		if err := m.updateMFASession(ctx, session); err != nil {
			log.Printf("Warning: Failed to update MFA session: %v", err)
		}

		return &MFAResponse{
			Success: true,
			Message: "MFA verification successful",
		}, nil
	} else {
		// Update session with failed attempt
		if err := m.updateMFASession(ctx, session); err != nil {
			log.Printf("Warning: Failed to update MFA session: %v", err)
		}

		attemptsLeft := session.MaxAttempts - session.Attempts
		return &MFAResponse{
			Success:      false,
			Error:        "Invalid MFA code",
			AttemptsLeft: attemptsLeft,
		}, fmt.Errorf("invalid MFA code")
	}
}

// createTOTPSession creates a TOTP-based MFA session
func (m *ExternalMFAService) createTOTPSession(ctx context.Context, sessionID, linkID string) (*MFASession, error) {
	// Generate TOTP secret
	secret, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "SecureEmail",
		AccountName: linkID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate TOTP secret: %w", err)
	}

	return &MFASession{
		ID:          sessionID,
		LinkID:      linkID,
		MFAType:     "totp",
		Secret:      secret.Secret(),
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(10 * time.Minute), // TOTP sessions expire in 10 minutes
		Verified:    false,
		Attempts:    0,
		MaxAttempts: 5,
	}, nil
}

// createEmailSession creates an email-based MFA session
func (m *ExternalMFAService) createEmailSession(ctx context.Context, sessionID, linkID, email string) (*MFASession, error) {
	// Generate 6-digit code
	code := m.generateCode(6)

	return &MFASession{
		ID:          sessionID,
		LinkID:      linkID,
		MFAType:     "email",
		Code:        code,
		Email:       email,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(5 * time.Minute), // Email codes expire in 5 minutes
		Verified:    false,
		Attempts:    0,
		MaxAttempts: 3,
	}, nil
}

// createSMSSession creates an SMS-based MFA session
func (m *ExternalMFAService) createSMSSession(ctx context.Context, sessionID, linkID, phone string) (*MFASession, error) {
	// Generate 6-digit code
	code := m.generateCode(6)

	return &MFASession{
		ID:          sessionID,
		LinkID:      linkID,
		MFAType:     "sms",
		Code:        code,
		Phone:       phone,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(5 * time.Minute), // SMS codes expire in 5 minutes
		Verified:    false,
		Attempts:    0,
		MaxAttempts: 3,
	}, nil
}

// verifyTOTP verifies a TOTP code
func (m *ExternalMFAService) verifyTOTP(secret, code string) bool {
	return totp.Validate(code, secret)
}

// verifyCode verifies a simple code
func (m *ExternalMFAService) verifyCode(expected, provided string) bool {
	return expected == provided
}

// sendMFACode sends MFA code via email or SMS
func (m *ExternalMFAService) sendMFACode(ctx context.Context, session *MFASession) error {
	switch session.MFAType {
	case "email":
		return m.sendEmailCode(ctx, session)
	case "sms":
		return m.sendSMSCode(ctx, session)
	default:
		return fmt.Errorf("unsupported MFA type for sending: %s", session.MFAType)
	}
}

// sendEmailCode sends MFA code via email
func (m *ExternalMFAService) sendEmailCode(ctx context.Context, session *MFASession) error {
	// TODO: Implement actual email sending
	// For now, just log the code
	log.Printf("MFA Email Code for %s: %s", session.Email, session.Code)
	return nil
}

// sendSMSCode sends MFA code via SMS
func (m *ExternalMFAService) sendSMSCode(ctx context.Context, session *MFASession) error {
	// TODO: Implement actual SMS sending
	// For now, just log the code
	log.Printf("MFA SMS Code for %s: %s", session.Phone, session.Code)
	return nil
}

// storeMFASession stores an MFA session in the database
func (m *ExternalMFAService) storeMFASession(ctx context.Context, session *MFASession) error {
	query := `
		INSERT INTO link_mfa_sessions (
			id, link_id, mfa_type, secret, code, email, phone, 
			created_at, expires_at, verified, attempts, max_attempts, locked_until
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := m.db.ExecContext(ctx, query,
		session.ID, session.LinkID, session.MFAType, session.Secret, session.Code,
		session.Email, session.Phone, session.CreatedAt, session.ExpiresAt,
		session.Verified, session.Attempts, session.MaxAttempts, session.LockedUntil,
	)

	return err
}

// getMFASession retrieves an MFA session from the database
func (m *ExternalMFAService) getMFASession(ctx context.Context, sessionID string) (*MFASession, error) {
	query := `
		SELECT id, link_id, mfa_type, secret, code, email, phone,
		       created_at, expires_at, verified, attempts, max_attempts, locked_until
		FROM link_mfa_sessions
		WHERE id = ?
	`

	var session MFASession
	err := m.db.QueryRowContext(ctx, query, sessionID).Scan(
		&session.ID, &session.LinkID, &session.MFAType, &session.Secret, &session.Code,
		&session.Email, &session.Phone, &session.CreatedAt, &session.ExpiresAt,
		&session.Verified, &session.Attempts, &session.MaxAttempts, &session.LockedUntil,
	)

	if err != nil {
		return nil, err
	}

	return &session, nil
}

// updateMFASession updates an MFA session in the database
func (m *ExternalMFAService) updateMFASession(ctx context.Context, session *MFASession) error {
	query := `
		UPDATE link_mfa_sessions
		SET verified = ?, attempts = ?, locked_until = ?
		WHERE id = ?
	`

	_, err := m.db.ExecContext(ctx, query,
		session.Verified, session.Attempts, session.LockedUntil, session.ID,
	)

	return err
}

// generateSessionID generates a unique session ID
func (m *ExternalMFAService) generateSessionID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}

// generateCode generates a numeric code of specified length
func (m *ExternalMFAService) generateCode(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)

	code := ""
	for _, b := range bytes {
		code += fmt.Sprintf("%d", int(b)%10)
	}

	return code
}

// isValidMFAType checks if the MFA type is valid
func (m *ExternalMFAService) isValidMFAType(mfaType string) bool {
	validTypes := []string{"totp", "email", "sms"}
	for _, validType := range validTypes {
		if mfaType == validType {
			return true
		}
	}
	return false
}

// getMFAMessage returns the appropriate message for MFA type
func (m *ExternalMFAService) getMFAMessage(mfaType string) string {
	switch mfaType {
	case "totp":
		return "Please enter the code from your authenticator app"
	case "email":
		return "Please check your email for the verification code"
	case "sms":
		return "Please check your phone for the SMS verification code"
	default:
		return "Please enter the verification code"
	}
}
