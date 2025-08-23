package security

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

	"secure-email-mvp/pkg/securelinks"

	"golang.org/x/crypto/argon2"
)

// =============================================================================
// SECURITY ENFORCEMENT SERVICE
// =============================================================================

// SecurityEnforcementService handles all Phase 2 security features
type SecurityEnforcementService struct {
	db *sql.DB
}

// NewSecurityEnforcementService creates a new security enforcement service
func NewSecurityEnforcementService(db *sql.DB) *SecurityEnforcementService {
	return &SecurityEnforcementService{
		db: db,
	}
}

// =============================================================================
// PASSWORD PROTECTION SYSTEM
// =============================================================================

// PasswordProtectionService handles password protection for secure links
type PasswordProtectionService struct {
	db *sql.DB
}

// NewPasswordProtectionService creates a new password protection service
func NewPasswordProtectionService(db *sql.DB) *PasswordProtectionService {
	return &PasswordProtectionService{
		db: db,
	}
}

// ValidatePassword validates a password for a secure link
func (p *PasswordProtectionService) ValidatePassword(ctx context.Context, req PasswordValidationRequest) (*PasswordValidationResponse, error) {
	// Get current password attempt count
	attemptCount, err := p.getCurrentAttemptCount(ctx, req.LinkID, req.IPAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get attempt count: %w", err)
	}

	// Check if link is locked out
	lockoutUntil, err := p.getLockoutUntil(ctx, req.LinkID)
	if err != nil {
		return nil, fmt.Errorf("failed to check lockout status: %w", err)
	}

	if lockoutUntil != nil && time.Now().Before(*lockoutUntil) {
		return &PasswordValidationResponse{
			Valid:         false,
			AttemptNumber: attemptCount,
			MaxAttempts:   3, // Default, will be updated from settings
			LockoutUntil:  lockoutUntil,
			LockedOut:     true,
			NextStep:      "locked",
		}, nil
	}

	// Get secure link and security settings
	link, err := p.getSecureLink(ctx, req.LinkID)
	if err != nil {
		return nil, fmt.Errorf("failed to get secure link: %w", err)
	}

	// Validate password
	if link.SecuritySettings.PasswordHash == nil {
		return nil, errors.New("no password hash configured for this link")
	}
	valid, err := p.verifyPassword(*link.SecuritySettings.PasswordHash, req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to verify password: %w", err)
	}

	// Record password attempt
	attempt := &PasswordAttempt{
		ID:            p.generateAttemptID(),
		LinkID:        req.LinkID,
		IPAddress:     req.IPAddress,
		UserAgent:     req.UserAgent,
		AttemptTime:   time.Now(),
		Success:       valid,
		AttemptNumber: attemptCount + 1,
	}

	if err := p.recordPasswordAttempt(ctx, attempt); err != nil {
		log.Printf("Warning: Failed to record password attempt: %v", err)
	}

	// Handle failed attempt
	if !valid {
		return p.handleFailedPasswordAttempt(ctx, link, attempt)
	}

	// Password is valid - clear failed attempts and generate session token
	if err := p.ClearFailedAttempts(ctx, req.LinkID); err != nil {
		log.Printf("Warning: Failed to clear failed attempts: %v", err)
	}

	sessionToken := p.generateSessionToken()

	// Determine next step based on security settings
	nextStep := "access"
	if link.SecuritySettings.RequireMFA {
		nextStep = "mfa"
	}

	return &PasswordValidationResponse{
		Valid:         true,
		AttemptNumber: attemptCount + 1,
		MaxAttempts:   link.SecuritySettings.MaxAccessAttempts,
		SessionToken:  sessionToken,
		NextStep:      nextStep,
	}, nil
}

// HashPassword creates a secure hash of a password using Argon2
func (p *PasswordProtectionService) HashPassword(password string) (string, error) {
	// Generate a random salt
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Hash the password using Argon2id
	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)

	// Combine salt and hash
	combined := make([]byte, 0, len(salt)+len(hash))
	combined = append(combined, salt...)
	combined = append(combined, hash...)

	// Encode as base64
	encoded := base64.StdEncoding.EncodeToString(combined)
	return encoded, nil
}

// verifyPassword verifies a password against its hash
func (p *PasswordProtectionService) verifyPassword(hash, password string) (bool, error) {
	if hash == "" {
		return false, errors.New("no password hash provided")
	}

	// Decode the hash
	decoded, err := base64.StdEncoding.DecodeString(hash)
	if err != nil {
		return false, fmt.Errorf("invalid hash format: %w", err)
	}

	if len(decoded) < 16 {
		return false, errors.New("hash too short")
	}

	// Extract salt and hash
	salt := decoded[:16]
	storedHash := decoded[16:]

	// Hash the provided password with the same salt
	computedHash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)

	// Compare hashes
	if len(computedHash) != len(storedHash) {
		return false, nil
	}

	for i := range computedHash {
		if computedHash[i] != storedHash[i] {
			return false, nil
		}
	}

	return true, nil
}

// handleFailedPasswordAttempt handles a failed password attempt
func (p *PasswordProtectionService) handleFailedPasswordAttempt(ctx context.Context, link *securelinks.SecureLink, attempt *PasswordAttempt) (*PasswordValidationResponse, error) {
	// Update failed attempts count
	if err := p.incrementFailedAttempts(ctx, attempt.LinkID); err != nil {
		log.Printf("Warning: Failed to increment failed attempts: %v", err)
	}

	// Check if we should lock out the link
	if attempt.AttemptNumber >= link.SecuritySettings.MaxAccessAttempts {
		lockoutDuration := time.Hour // Default 1 hour lockout
		lockoutUntil := time.Now().Add(lockoutDuration)

		if err := p.setLockoutUntil(ctx, attempt.LinkID, lockoutUntil); err != nil {
			log.Printf("Warning: Failed to set lockout: %v", err)
		}

		// Check if we should auto-destruct after failed attempts
		if link.SecuritySettings.SelfDestructThreshold != nil &&
			attempt.AttemptNumber >= *link.SecuritySettings.SelfDestructThreshold {
			if err := p.autoDestructLink(ctx, attempt.LinkID); err != nil {
				log.Printf("Warning: Failed to auto-destruct link: %v", err)
			}
		}

		return &PasswordValidationResponse{
			Valid:         false,
			AttemptNumber: attempt.AttemptNumber,
			MaxAttempts:   link.SecuritySettings.MaxAccessAttempts,
			LockoutUntil:  &lockoutUntil,
			LockedOut:     true,
			NextStep:      "locked",
		}, nil
	}

	return &PasswordValidationResponse{
		Valid:         false,
		AttemptNumber: attempt.AttemptNumber,
		MaxAttempts:   link.SecuritySettings.MaxAccessAttempts,
		NextStep:      "password",
	}, nil
}

// =============================================================================
// DATABASE OPERATIONS
// =============================================================================

// getCurrentAttemptCount gets the current password attempt count for an IP
func (p *PasswordProtectionService) getCurrentAttemptCount(ctx context.Context, linkID, ipAddress string) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM link_password_attempts 
		WHERE link_id = ? AND ip_address = ? AND success = false
		AND attempt_time > datetime('now', '-1 hour')
	`

	var count int
	err := p.db.QueryRowContext(ctx, query, linkID, ipAddress).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get attempt count: %w", err)
	}

	return count, nil
}

// getLockoutUntil gets the lockout until time for a link
func (p *PasswordProtectionService) getLockoutUntil(ctx context.Context, linkID string) (*time.Time, error) {
	query := `SELECT password_lockout_until FROM secure_links WHERE link_id = ?`

	var lockoutUntil *time.Time
	err := p.db.QueryRowContext(ctx, query, linkID).Scan(&lockoutUntil)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get lockout until: %w", err)
	}

	return lockoutUntil, nil
}

// getSecureLink gets a secure link with security settings
func (p *PasswordProtectionService) getSecureLink(ctx context.Context, linkID string) (*securelinks.SecureLink, error) {
	query := `
		SELECT link_id, email_id, recipient_email, sender_id, security_settings,
		       created_at, expires_at, access_count, last_accessed, status,
		       failed_attempts, last_failed_attempt, lockout_until
		FROM secure_links 
		WHERE link_id = ?
	`

	row := p.db.QueryRowContext(ctx, query, linkID)

	var link securelinks.SecureLink
	var settingsJSON []byte

	err := row.Scan(
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
		if err == sql.ErrNoRows {
			return nil, errors.New("secure link not found")
		}
		return nil, fmt.Errorf("failed to scan secure link: %w", err)
	}

	// Unmarshal security settings
	if err := json.Unmarshal(settingsJSON, &link.SecuritySettings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal security settings: %w", err)
	}

	return &link, nil
}

// recordPasswordAttempt records a password attempt
func (p *PasswordProtectionService) recordPasswordAttempt(ctx context.Context, attempt *PasswordAttempt) error {
	query := `
		INSERT INTO link_password_attempts (
			id, link_id, ip_address, user_agent, attempt_time, success, 
			attempt_number, geolocation_data
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := p.db.ExecContext(ctx, query,
		attempt.ID,
		attempt.LinkID,
		attempt.IPAddress,
		attempt.UserAgent,
		attempt.AttemptTime,
		attempt.Success,
		attempt.AttemptNumber,
		attempt.GeolocationData,
	)

	if err != nil {
		return fmt.Errorf("failed to insert password attempt: %w", err)
	}

	return nil
}

// incrementFailedAttempts increments the failed attempts count
func (p *PasswordProtectionService) incrementFailedAttempts(ctx context.Context, linkID string) error {
	query := `
		UPDATE secure_links 
		SET failed_attempts = failed_attempts + 1, last_failed_attempt = ?
		WHERE link_id = ?
	`

	_, err := p.db.ExecContext(ctx, query, time.Now(), linkID)
	if err != nil {
		return fmt.Errorf("failed to increment failed attempts: %w", err)
	}

	return nil
}

// ClearFailedAttempts clears the failed attempts for a link
func (p *PasswordProtectionService) ClearFailedAttempts(ctx context.Context, linkID string) error {
	query := `
		UPDATE secure_links 
		SET failed_attempts = 0, last_failed_attempt = NULL, password_lockout_until = NULL
		WHERE link_id = ?
	`

	_, err := p.db.ExecContext(ctx, query, linkID)
	if err != nil {
		return fmt.Errorf("failed to clear failed attempts: %w", err)
	}

	return nil
}

// setLockoutUntil sets the lockout until time for a link
func (p *PasswordProtectionService) setLockoutUntil(ctx context.Context, linkID string, lockoutUntil time.Time) error {
	query := `UPDATE secure_links SET password_lockout_until = ? WHERE link_id = ?`

	_, err := p.db.ExecContext(ctx, query, lockoutUntil, linkID)
	if err != nil {
		return fmt.Errorf("failed to set lockout until: %w", err)
	}

	return nil
}

// autoDestructLink marks a link as destroyed
func (p *PasswordProtectionService) autoDestructLink(ctx context.Context, linkID string) error {
	query := `UPDATE secure_links SET status = 'destroyed' WHERE link_id = ?`

	_, err := p.db.ExecContext(ctx, query, linkID)
	if err != nil {
		return fmt.Errorf("failed to auto-destruct link: %w", err)
	}

	return nil
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

// generateAttemptID generates a unique attempt ID
func (p *PasswordProtectionService) generateAttemptID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}

// generateSessionToken generates a session token for successful authentication
func (p *PasswordProtectionService) generateSessionToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}

// =============================================================================
// PLACEHOLDER FUNCTIONS FOR OTHER SECURITY FEATURES
// =============================================================================

// TODO: Implement enhanced geolocation verification
// TODO: Implement time lock functionality
// TODO: Implement auto-destruct and read-once features
// TODO: Implement decoy message system
// TODO: Implement tamper alerts
// TODO: Implement MFA for external users
