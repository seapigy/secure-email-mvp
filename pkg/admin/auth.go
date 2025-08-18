package admin

import (
	"crypto/rand"
	"crypto/subtle"
	"database/sql"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"log"
	mathrand "math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/argon2"
)

// AdminUser represents an admin user in the system
type AdminUser struct {
	ID                  string     `json:"id"`
	Email               string     `json:"email"`
	Role                string     `json:"role"`
	TOTPEnabled         bool       `json:"totp_enabled"`
	IsActive            bool       `json:"is_active"`
	LastLogin           *time.Time `json:"last_login,omitempty"`
	FailedLoginAttempts int        `json:"failed_login_attempts"`
	LockedUntil         *time.Time `json:"locked_until,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
	CreatedBy           *string    `json:"created_by,omitempty"`
}

// AdminInvitation represents an admin invitation
type AdminInvitation struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	Role        string    `json:"role"`
	ExpiresAt   time.Time `json:"expires_at"`
	MaxUses     int       `json:"max_uses"`
	CurrentUses int       `json:"current_uses"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
}

// AdminSession represents an admin session
type AdminSession struct {
	ID           string    `json:"id"`
	AdminID      string    `json:"admin_id"`
	SessionToken string    `json:"session_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	IPAddress    string    `json:"ip_address"`
	UserAgent    string    `json:"user_agent"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
}

// AdminAuditLog represents an admin audit log entry
type AdminAuditLog struct {
	ID           string          `json:"id"`
	AdminID      *string         `json:"admin_id,omitempty"`
	Action       string          `json:"action"`
	ResourceType *string         `json:"resource_type,omitempty"`
	ResourceID   *string         `json:"resource_id,omitempty"`
	Details      json.RawMessage `json:"details,omitempty"`
	IPAddress    string          `json:"ip_address"`
	UserAgent    string          `json:"user_agent"`
	Success      bool            `json:"success"`
	CreatedAt    time.Time       `json:"created_at"`
}

// AdminAuthService handles admin authentication and management
type AdminAuthService struct {
	db *sql.DB
}

// NewAdminAuthService creates a new admin authentication service
func NewAdminAuthService(db *sql.DB) *AdminAuthService {
	return &AdminAuthService{db: db}
}

// ValidateAdminEmail checks if email is valid for admin accounts
func ValidateAdminEmail(email string) bool {
	// For now, allow any email format for admin accounts
	// In production, you might want to restrict to specific domains
	emailRegex := regexp.MustCompile(`^[^@]+@[^@]+\.[^@]+$`)
	return emailRegex.MatchString(email)
}

// ValidateAdminPassword checks if password meets admin security requirements
func ValidateAdminPassword(password string) error {
	if len(password) < 16 {
		return fmt.Errorf("password must be at least 16 characters long")
	}

	// Check for complexity requirements
	var (
		hasUpper   bool
		hasLower   bool
		hasDigit   bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasDigit = true
		case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;:,.<>?", char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if !hasDigit {
		return fmt.Errorf("password must contain at least one digit")
	}
	if !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}

// HashAdminPassword creates an Argon2id hash for admin passwords
func HashAdminPassword(password string) (string, error) {
	// Generate a random salt
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Hash password with Argon2id
	hash := argon2.IDKey([]byte(password), salt, 3, 64*1024, 4, 32)

	// Encode as base64
	encodedHash := base32.StdEncoding.EncodeToString(hash)
	encodedSalt := base32.StdEncoding.EncodeToString(salt)

	return fmt.Sprintf("$argon2id$%s$%s", encodedSalt, encodedHash), nil
}

// VerifyAdminPassword verifies an admin password against its hash
func VerifyAdminPassword(password, hash string) (bool, error) {
	// Parse the hash format
	parts := strings.Split(hash, "$")
	if len(parts) != 4 || parts[1] != "argon2id" {
		return false, fmt.Errorf("invalid hash format")
	}

	// Decode salt and hash
	salt, err := base32.StdEncoding.DecodeString(parts[2])
	if err != nil {
		return false, fmt.Errorf("invalid salt encoding: %w", err)
	}

	storedHash, err := base32.StdEncoding.DecodeString(parts[3])
	if err != nil {
		return false, fmt.Errorf("invalid hash encoding: %w", err)
	}

	// Hash the provided password with the same parameters
	computedHash := argon2.IDKey([]byte(password), salt, 3, 64*1024, 4, 32)

	// Compare hashes
	return subtle.ConstantTimeCompare(computedHash, storedHash) == 1, nil
}

// GenerateTOTPSecret generates a new TOTP secret for admin 2FA
func GenerateTOTPSecret() (string, error) {
	// Generate a random secret
	secret := make([]byte, 20)
	if _, err := rand.Read(secret); err != nil {
		return "", fmt.Errorf("failed to generate TOTP secret: %w", err)
	}

	return base32.StdEncoding.EncodeToString(secret), nil
}

// ValidateTOTP validates a TOTP code against a secret
func ValidateTOTP(secret, code string) bool {
	return totp.Validate(code, secret)
}

// CheckAdminExists checks if any admin user exists in the system
func (s *AdminAuthService) CheckAdminExists() (bool, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM admin_users WHERE is_active = 1").Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check admin existence: %w", err)
	}
	return count > 0, nil
}

// CreateRootAdmin creates the first root admin account
func (s *AdminAuthService) CreateRootAdmin(email, password string) (*AdminUser, error) {
	// Validate inputs
	if !ValidateAdminEmail(email) {
		return nil, fmt.Errorf("invalid admin email format")
	}

	if err := ValidateAdminPassword(password); err != nil {
		return nil, fmt.Errorf("password validation failed: %w", err)
	}

	// Check if admin already exists
	exists, err := s.CheckAdminExists()
	if err != nil {
		return nil, fmt.Errorf("failed to check admin existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("root admin already exists")
	}

	// Hash password
	passwordHash, err := HashAdminPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Generate TOTP secret
	totpSecret, err := GenerateTOTPSecret()
	if err != nil {
		return nil, fmt.Errorf("failed to generate TOTP secret: %w", err)
	}

	// Create admin user
	adminID := uuid.New().String()
	now := time.Now()

	_, err = s.db.Exec(`
		INSERT INTO admin_users (
			id, email, password_hash, totp_secret, role, is_active, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, adminID, email, passwordHash, totpSecret, "root_admin", true, now, now)

	if err != nil {
		return nil, fmt.Errorf("failed to create root admin: %w", err)
	}

	// Log the creation
	s.logAdminAction(adminID, "admin_created", "admin_users", adminID, map[string]interface{}{
		"email": email,
		"role":  "root_admin",
	}, "127.0.0.1", "system", true)

	return &AdminUser{
		ID:          adminID,
		Email:       email,
		Role:        "root_admin",
		TOTPEnabled: false,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// AuthenticateAdmin authenticates an admin user
func (s *AdminAuthService) AuthenticateAdmin(email, password, totpCode string, ipAddress, userAgent string) (*AdminUser, *AdminSession, error) {
	// Get admin user
	var admin AdminUser
	var passwordHash, totpSecret string

	err := s.db.QueryRow(`
		SELECT id, email, password_hash, totp_secret, role, totp_enabled, is_active, 
		       failed_login_attempts, locked_until, created_at, updated_at
		FROM admin_users 
		WHERE email = ? AND is_active = 1
	`, email).Scan(
		&admin.ID, &admin.Email, &passwordHash, &totpSecret, &admin.Role,
		&admin.TOTPEnabled, &admin.IsActive, &admin.FailedLoginAttempts,
		&admin.LockedUntil, &admin.CreatedAt, &admin.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			s.logAdminAction("", "admin_login_failed", "admin_users", "", map[string]interface{}{
				"email":  email,
				"reason": "user_not_found",
			}, ipAddress, userAgent, false)
			return nil, nil, fmt.Errorf("invalid credentials")
		}
		return nil, nil, fmt.Errorf("failed to query admin user: %w", err)
	}

	// Check if account is locked
	if admin.LockedUntil != nil && admin.LockedUntil.After(time.Now()) {
		s.logAdminAction(admin.ID, "admin_login_failed", "admin_users", admin.ID, map[string]interface{}{
			"email":        email,
			"reason":       "account_locked",
			"locked_until": admin.LockedUntil,
		}, ipAddress, userAgent, false)
		return nil, nil, fmt.Errorf("account is locked until %s", admin.LockedUntil.Format(time.RFC3339))
	}

	// Verify password
	passwordValid, err := VerifyAdminPassword(password, passwordHash)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to verify password: %w", err)
	}

	if !passwordValid {
		// Increment failed login attempts
		s.incrementFailedLoginAttempts(admin.ID, admin.FailedLoginAttempts)

		s.logAdminAction(admin.ID, "admin_login_failed", "admin_users", admin.ID, map[string]interface{}{
			"email":  email,
			"reason": "invalid_password",
		}, ipAddress, userAgent, false)

		return nil, nil, fmt.Errorf("invalid credentials")
	}

	// Verify TOTP if enabled
	if admin.TOTPEnabled {
		if !ValidateTOTP(totpSecret, totpCode) {
			// Increment failed login attempts
			s.incrementFailedLoginAttempts(admin.ID, admin.FailedLoginAttempts)

			s.logAdminAction(admin.ID, "admin_login_failed", "admin_users", admin.ID, map[string]interface{}{
				"email":  email,
				"reason": "invalid_totp",
			}, ipAddress, userAgent, false)

			return nil, nil, fmt.Errorf("invalid TOTP code")
		}
	}

	// Reset failed login attempts on successful login
	if admin.FailedLoginAttempts > 0 {
		_, err = s.db.Exec("UPDATE admin_users SET failed_login_attempts = 0, locked_until = NULL WHERE id = ?", admin.ID)
		if err != nil {
			log.Printf("Failed to reset failed login attempts: %v", err)
		}
	}

	// Update last login
	now := time.Now()
	_, err = s.db.Exec("UPDATE admin_users SET last_login = ? WHERE id = ?", now, admin.ID)
	if err != nil {
		log.Printf("Failed to update last login: %v", err)
	}

	// Create session
	session, err := s.CreateAdminSession(admin.ID, ipAddress, userAgent)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Log successful login
	s.logAdminAction(admin.ID, "admin_login_success", "admin_users", admin.ID, map[string]interface{}{
		"email":      email,
		"session_id": session.ID,
	}, ipAddress, userAgent, true)

	admin.LastLogin = &now
	admin.FailedLoginAttempts = 0
	admin.LockedUntil = nil

	return &admin, session, nil
}

// incrementFailedLoginAttempts increments failed login attempts and locks account if necessary
func (s *AdminAuthService) incrementFailedLoginAttempts(adminID string, currentAttempts int) {
	newAttempts := currentAttempts + 1
	var lockedUntil *time.Time

	// Lock account after 5 failed attempts for 30 minutes
	if newAttempts >= 5 {
		lockTime := time.Now().Add(30 * time.Minute)
		lockedUntil = &lockTime
	}

	_, err := s.db.Exec("UPDATE admin_users SET failed_login_attempts = ?, locked_until = ? WHERE id = ?",
		newAttempts, lockedUntil, adminID)
	if err != nil {
		log.Printf("Failed to update failed login attempts: %v", err)
	}
}

// CreateAdminSession creates a new admin session
func (s *AdminAuthService) CreateAdminSession(adminID, ipAddress, userAgent string) (*AdminSession, error) {
	sessionID := uuid.New().String()
	sessionToken := uuid.New().String()
	refreshToken := uuid.New().String()
	expiresAt := time.Now().Add(30 * time.Minute) // 30 minute session

	_, err := s.db.Exec(`
		INSERT INTO admin_sessions (
			id, admin_id, session_token, refresh_token, expires_at, ip_address, user_agent, is_active, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, sessionID, adminID, sessionToken, refreshToken, expiresAt, ipAddress, userAgent, true, time.Now())

	if err != nil {
		return nil, fmt.Errorf("failed to create admin session: %w", err)
	}

	return &AdminSession{
		ID:           sessionID,
		AdminID:      adminID,
		SessionToken: sessionToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		IsActive:     true,
		CreatedAt:    time.Now(),
	}, nil
}

// ValidateAdminSession validates an admin session token
func (s *AdminAuthService) ValidateAdminSession(sessionToken string) (*AdminUser, error) {
	var admin AdminUser
	var sessionID string

	err := s.db.QueryRow(`
		SELECT s.id, a.id, a.email, a.role, a.totp_enabled, a.is_active, 
		       a.last_login, a.failed_login_attempts, a.locked_until, 
		       a.created_at, a.updated_at
		FROM admin_sessions s
		JOIN admin_users a ON s.admin_id = a.id
		WHERE s.session_token = ? AND s.is_active = 1 AND s.expires_at > ?
	`, sessionToken, time.Now()).Scan(
		&sessionID, &admin.ID, &admin.Email, &admin.Role, &admin.TOTPEnabled,
		&admin.IsActive, &admin.LastLogin, &admin.FailedLoginAttempts,
		&admin.LockedUntil, &admin.CreatedAt, &admin.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invalid or expired session")
		}
		return nil, fmt.Errorf("failed to validate session: %w", err)
	}

	return &admin, nil
}

// LogoutAdmin logs out an admin user
func (s *AdminAuthService) LogoutAdmin(sessionToken, ipAddress, userAgent string) error {
	// Get admin ID from session
	var adminID string
	err := s.db.QueryRow("SELECT admin_id FROM admin_sessions WHERE session_token = ? AND is_active = 1",
		sessionToken).Scan(&adminID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("invalid session")
		}
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Deactivate session
	_, err = s.db.Exec("UPDATE admin_sessions SET is_active = 0 WHERE session_token = ?", sessionToken)
	if err != nil {
		return fmt.Errorf("failed to deactivate session: %w", err)
	}

	// Log logout
	s.logAdminAction(adminID, "admin_logout", "admin_sessions", sessionToken, map[string]interface{}{
		"session_token": sessionToken,
	}, ipAddress, userAgent, true)

	return nil
}

// logAdminAction logs an admin action for audit purposes
func (s *AdminAuthService) logAdminAction(adminID, action, resourceType, resourceID string,
	details map[string]interface{}, ipAddress, userAgent string, success bool) {

	detailsJSON, err := json.Marshal(details)
	if err != nil {
		log.Printf("Failed to marshal audit details: %v", err)
		detailsJSON = []byte("{}")
	}

	var adminIDPtr *string
	if adminID != "" {
		adminIDPtr = &adminID
	}

	var resourceTypePtr *string
	if resourceType != "" {
		resourceTypePtr = &resourceType
	}

	var resourceIDPtr *string
	if resourceID != "" {
		resourceIDPtr = &resourceID
	}

	_, err = s.db.Exec(`
		INSERT INTO admin_audit_logs (
			id, admin_id, action, resource_type, resource_id, details, ip_address, user_agent, success, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, uuid.New().String(), adminIDPtr, action, resourceTypePtr, resourceIDPtr,
		string(detailsJSON), ipAddress, userAgent, success, time.Now())

	if err != nil {
		log.Printf("Failed to log admin action: %v", err)
	}
}

// GetAdminAuditLogs retrieves admin audit logs
func (s *AdminAuthService) GetAdminAuditLogs(limit int) ([]AdminAuditLog, error) {
	rows, err := s.db.Query(`
		SELECT id, admin_id, action, resource_type, resource_id, details, 
		       ip_address, user_agent, success, created_at
		FROM admin_audit_logs
		ORDER BY created_at DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()

	var logs []AdminAuditLog
	for rows.Next() {
		var auditLog AdminAuditLog
		var detailsStr string

		err := rows.Scan(
			&auditLog.ID, &auditLog.AdminID, &auditLog.Action, &auditLog.ResourceType, &auditLog.ResourceID,
			&detailsStr, &auditLog.IPAddress, &auditLog.UserAgent, &auditLog.Success, &auditLog.CreatedAt,
		)
		if err != nil {
			log.Printf("Failed to scan audit log: %v", err)
			continue
		}

		if detailsStr != "" {
			auditLog.Details = json.RawMessage(detailsStr)
		}

		logs = append(logs, auditLog)
	}

	return logs, nil
}

// ============================================================================
// INVITATION KEY SYSTEM - ITERATION 4
// ============================================================================

// CreateInvitationKey creates a new invitation key for admin account creation
func (s *AdminAuthService) CreateInvitationKey(createdByAdminID, email, role string, maxUses int, ipAddress, userAgent string) (*AdminInvitation, error) {
	// Validate inputs
	if !ValidateAdminEmail(email) {
		return nil, fmt.Errorf("invalid email format")
	}

	if !ValidateAdminRole(role) {
		return nil, fmt.Errorf("invalid role: %s", role)
	}

	if maxUses < 1 || maxUses > 10 {
		return nil, fmt.Errorf("max_uses must be between 1 and 10")
	}

	// Check if creator has permission to create invitations
	creator, err := s.GetAdminByID(createdByAdminID)
	if err != nil {
		return nil, fmt.Errorf("failed to get creator admin: %w", err)
	}

	if !CanCreateInvitations(creator.Role, role) {
		return nil, fmt.Errorf("insufficient permissions to create invitation for role: %s", role)
	}

	// Check if email already has an active invitation
	exists, err := s.CheckInvitationExists(email)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing invitation: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("invitation already exists for email: %s", email)
	}

	// Generate secure invitation token
	invitationToken, err := GenerateSecureToken(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate invitation token: %w", err)
	}

	// Set expiration (24 hours from now)
	expiresAt := time.Now().Add(24 * time.Hour)

	// Create invitation
	invitationID := uuid.New().String()
	now := time.Now()

	_, err = s.db.Exec(`
		INSERT INTO admin_invitation_keys (
			id, email, invitation_token, role, expires_at, max_uses, current_uses, created_by, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, invitationID, email, invitationToken, role, expiresAt, maxUses, 0, createdByAdminID, now)

	if err != nil {
		return nil, fmt.Errorf("failed to create invitation: %w", err)
	}

	// Log the invitation creation
	s.logAdminAction(createdByAdminID, "invitation_created", "admin_invitation_keys", invitationID, map[string]interface{}{
		"email":      email,
		"role":       role,
		"max_uses":   maxUses,
		"expires_at": expiresAt,
	}, ipAddress, userAgent, true)

	return &AdminInvitation{
		ID:          invitationID,
		Email:       email,
		Role:        role,
		ExpiresAt:   expiresAt,
		MaxUses:     maxUses,
		CurrentUses: 0,
		CreatedBy:   createdByAdminID,
		CreatedAt:   now,
	}, nil
}

// ValidateInvitationKey validates an invitation key and returns the invitation details
func (s *AdminAuthService) ValidateInvitationKey(invitationToken string) (*AdminInvitation, error) {
	var invitation AdminInvitation
	var createdBy string

	err := s.db.QueryRow(`
		SELECT id, email, role, expires_at, max_uses, current_uses, created_by, created_at
		FROM admin_invitation_keys
		WHERE invitation_token = ?
	`, invitationToken).Scan(
		&invitation.ID, &invitation.Email, &invitation.Role, &invitation.ExpiresAt,
		&invitation.MaxUses, &invitation.CurrentUses, &createdBy, &invitation.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invalid invitation token")
		}
		return nil, fmt.Errorf("failed to query invitation: %w", err)
	}

	// Check if invitation has expired
	if invitation.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("invitation has expired")
	}

	// Check if invitation has reached max uses
	if invitation.CurrentUses >= invitation.MaxUses {
		return nil, fmt.Errorf("invitation has reached maximum uses")
	}

	invitation.CreatedBy = createdBy
	return &invitation, nil
}

// UseInvitationKey marks an invitation as used and creates the admin account
func (s *AdminAuthService) UseInvitationKey(invitationToken, password string, ipAddress, userAgent string) (*AdminUser, error) {
	// Validate invitation
	invitation, err := s.ValidateInvitationKey(invitationToken)
	if err != nil {
		return nil, fmt.Errorf("invalid invitation: %w", err)
	}

	// Validate password
	if err := ValidateAdminPassword(password); err != nil {
		return nil, fmt.Errorf("password validation failed: %w", err)
	}

	// Check if admin with this email already exists
	exists, err := s.CheckAdminEmailExists(invitation.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check admin existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("admin account already exists for email: %s", invitation.Email)
	}

	// Begin transaction
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Increment invitation usage
	_, err = tx.Exec("UPDATE admin_invitation_keys SET current_uses = current_uses + 1 WHERE id = ?", invitation.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to update invitation usage: %w", err)
	}

	// Hash password
	passwordHash, err := HashAdminPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Generate TOTP secret
	totpSecret, err := GenerateTOTPSecret()
	if err != nil {
		return nil, fmt.Errorf("failed to generate TOTP secret: %w", err)
	}

	// Create admin user
	adminID := uuid.New().String()
	now := time.Now()

	_, err = tx.Exec(`
		INSERT INTO admin_users (
			id, email, password_hash, totp_secret, role, is_active, created_at, updated_at, created_by
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, adminID, invitation.Email, passwordHash, totpSecret, invitation.Role, true, now, now, invitation.CreatedBy)

	if err != nil {
		return nil, fmt.Errorf("failed to create admin user: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Log the admin creation
	s.logAdminAction(invitation.CreatedBy, "admin_created_via_invitation", "admin_users", adminID, map[string]interface{}{
		"email":            invitation.Email,
		"role":             invitation.Role,
		"invitation_id":    invitation.ID,
		"invitation_token": invitationToken,
	}, ipAddress, userAgent, true)

	return &AdminUser{
		ID:          adminID,
		Email:       invitation.Email,
		Role:        invitation.Role,
		TOTPEnabled: false,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
		CreatedBy:   &invitation.CreatedBy,
	}, nil
}

// ListInvitationKeys retrieves all invitation keys (admin only)
func (s *AdminAuthService) ListInvitationKeys(adminID string) ([]AdminInvitation, error) {
	// Check if admin has permission
	admin, err := s.GetAdminByID(adminID)
	if err != nil {
		return nil, fmt.Errorf("failed to get admin: %w", err)
	}

	if !CanViewInvitations(admin.Role) {
		return nil, fmt.Errorf("insufficient permissions to view invitations")
	}

	rows, err := s.db.Query(`
		SELECT id, email, role, expires_at, max_uses, current_uses, created_by, created_at
		FROM admin_invitation_keys
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query invitations: %w", err)
	}
	defer rows.Close()

	var invitations []AdminInvitation
	for rows.Next() {
		var invitation AdminInvitation
		var createdBy string

		err := rows.Scan(
			&invitation.ID, &invitation.Email, &invitation.Role, &invitation.ExpiresAt,
			&invitation.MaxUses, &invitation.CurrentUses, &createdBy, &invitation.CreatedAt,
		)
		if err != nil {
			log.Printf("Failed to scan invitation: %v", err)
			continue
		}

		invitation.CreatedBy = createdBy
		invitations = append(invitations, invitation)
	}

	return invitations, nil
}

// RevokeInvitationKey revokes an invitation key (admin only)
func (s *AdminAuthService) RevokeInvitationKey(adminID, invitationID, ipAddress, userAgent string) error {
	// Check if admin has permission
	admin, err := s.GetAdminByID(adminID)
	if err != nil {
		return fmt.Errorf("failed to get admin: %w", err)
	}

	if !CanRevokeInvitations(admin.Role) {
		return fmt.Errorf("insufficient permissions to revoke invitations")
	}

	// Get invitation details
	var email, role string
	err = s.db.QueryRow("SELECT email, role FROM admin_invitation_keys WHERE id = ?", invitationID).Scan(&email, &role)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("invitation not found")
		}
		return fmt.Errorf("failed to get invitation: %w", err)
	}

	// Check if admin can revoke this specific invitation
	if !CanRevokeSpecificInvitation(admin.Role, role) {
		return fmt.Errorf("insufficient permissions to revoke invitation for role: %s", role)
	}

	// Delete invitation
	_, err = s.db.Exec("DELETE FROM admin_invitation_keys WHERE id = ?", invitationID)
	if err != nil {
		return fmt.Errorf("failed to revoke invitation: %w", err)
	}

	// Log the revocation
	s.logAdminAction(adminID, "invitation_revoked", "admin_invitation_keys", invitationID, map[string]interface{}{
		"email": email,
		"role":  role,
	}, ipAddress, userAgent, true)

	return nil
}

// GetAdminByID retrieves an admin user by ID
func (s *AdminAuthService) GetAdminByID(adminID string) (*AdminUser, error) {
	var admin AdminUser
	var passwordHash, totpSecret string

	err := s.db.QueryRow(`
		SELECT id, email, password_hash, totp_secret, role, totp_enabled, is_active, 
		       last_login, failed_login_attempts, locked_until, created_at, updated_at, created_by
		FROM admin_users 
		WHERE id = ? AND is_active = 1
	`, adminID).Scan(
		&admin.ID, &admin.Email, &passwordHash, &totpSecret, &admin.Role,
		&admin.TOTPEnabled, &admin.IsActive, &admin.LastLogin, &admin.FailedLoginAttempts,
		&admin.LockedUntil, &admin.CreatedAt, &admin.UpdatedAt, &admin.CreatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("admin not found")
		}
		return nil, fmt.Errorf("failed to query admin: %w", err)
	}

	return &admin, nil
}

// CheckInvitationExists checks if an invitation exists for the given email
func (s *AdminAuthService) CheckInvitationExists(email string) (bool, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM admin_invitation_keys WHERE email = ? AND expires_at > ?", email, time.Now()).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check invitation existence: %w", err)
	}
	return count > 0, nil
}

// CheckAdminEmailExists checks if an admin with the given email exists
func (s *AdminAuthService) CheckAdminEmailExists(email string) (bool, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM admin_users WHERE email = ? AND is_active = 1", email).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check admin email existence: %w", err)
	}
	return count > 0, nil
}

// ============================================================================
// RBAC PERMISSION FUNCTIONS
// ============================================================================

// ValidateAdminRole validates if a role is valid
func ValidateAdminRole(role string) bool {
	validRoles := []string{"root_admin", "full_admin", "read_only_admin"}
	for _, validRole := range validRoles {
		if role == validRole {
			return true
		}
	}
	return false
}

// CanCreateInvitations checks if an admin can create invitations for a specific role
func CanCreateInvitations(adminRole, targetRole string) bool {
	switch adminRole {
	case "root_admin":
		return true // Root admin can create any role
	case "full_admin":
		return targetRole == "full_admin" || targetRole == "read_only_admin"
	case "read_only_admin":
		return false // Read-only admins cannot create invitations
	default:
		return false
	}
}

// CanViewInvitations checks if an admin can view invitations
func CanViewInvitations(adminRole string) bool {
	return adminRole == "root_admin" || adminRole == "full_admin"
}

// CanRevokeInvitations checks if an admin can revoke invitations
func CanRevokeInvitations(adminRole string) bool {
	return adminRole == "root_admin" || adminRole == "full_admin"
}

// CanRevokeSpecificInvitation checks if an admin can revoke a specific invitation
func CanRevokeSpecificInvitation(adminRole, targetRole string) bool {
	switch adminRole {
	case "root_admin":
		return true // Root admin can revoke any invitation
	case "full_admin":
		return targetRole == "full_admin" || targetRole == "read_only_admin"
	default:
		return false
	}
}

// CanManageAdmins checks if an admin can manage other admins
func CanManageAdmins(adminRole string) bool {
	return adminRole == "root_admin" || adminRole == "full_admin"
}

// CanViewSensitiveData checks if an admin can view sensitive data
func CanViewSensitiveData(adminRole string) bool {
	return adminRole == "root_admin" || adminRole == "full_admin"
}

// ============================================================================
// UTILITY FUNCTIONS
// ============================================================================

// GenerateSecureToken generates a secure random token
func GenerateSecureToken(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[mathrand.Intn(len(charset))]
	}
	return string(b), nil
}
