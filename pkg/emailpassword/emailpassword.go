package emailpassword

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Argon2Config holds configuration for Argon2id password hashing
type Argon2Config struct {
	Memory      uint32 // Memory cost in KiB
	Time        uint32 // Time cost (number of iterations)
	Parallelism uint8  // Parallelism degree
	KeyLength   uint32 // Length of the derived key
	SaltLength  uint32 // Length of the salt
}

// DefaultArgon2Config returns the default Argon2id configuration
func DefaultArgon2Config() *Argon2Config {
	return &Argon2Config{
		Memory:      64 * 1024, // 64 MiB
		Time:        3,          // 3 iterations
		Parallelism: 2,          // 2 threads
		KeyLength:   32,         // 32 bytes
		SaltLength:  16,         // 16 bytes
	}
}

// EmailPasswordService provides methods for managing email password protection
type EmailPasswordService struct {
	db     *sql.DB
	config *Argon2Config
}

// NewEmailPasswordService creates a new email password service
func NewEmailPasswordService(db *sql.DB) *EmailPasswordService {
	return &EmailPasswordService{
		db:     db,
		config: DefaultArgon2Config(),
	}
}

// NewEmailPasswordServiceWithConfig creates a new email password service with custom configuration
func NewEmailPasswordServiceWithConfig(db *sql.DB, config *Argon2Config) *EmailPasswordService {
	return &EmailPasswordService{
		db:     db,
		config: config,
	}
}

// SetEmailPassword sets a password for an email, hashing it with Argon2id
func (s *EmailPasswordService) SetEmailPassword(emailID string, rawPassword string) error {
	// Validate input
	if emailID == "" {
		return fmt.Errorf("email ID cannot be empty")
	}
	if rawPassword == "" {
		return fmt.Errorf("password cannot be empty")
	}

	// Generate random salt
	salt := make([]byte, s.config.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return fmt.Errorf("failed to generate salt: %w", err)
	}

	// Hash password with Argon2id
	hash := argon2.IDKey(
		[]byte(rawPassword),
		salt,
		s.config.Time,
		s.config.Memory,
		s.config.Parallelism,
		s.config.KeyLength,
	)

	// Encode hash and salt to base64
	hashB64 := base64.StdEncoding.EncodeToString(hash)
	saltB64 := base64.StdEncoding.EncodeToString(salt)

	// Store in database
	query := `
		UPDATE emails 
		SET is_password_protected = TRUE,
		    password_hash = ?,
		    password_salt = ?
		WHERE email_id = ?
	`

	result, err := s.db.Exec(query, hashB64, saltB64, emailID)
	if err != nil {
		return fmt.Errorf("failed to update email password: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("email not found: %s", emailID)
	}

	return nil
}

// CheckEmailPassword verifies a password against the stored hash for an email
func (s *EmailPasswordService) CheckEmailPassword(emailID string, providedPassword string) (bool, error) {
	// Validate input
	if emailID == "" {
		return false, fmt.Errorf("email ID cannot be empty")
	}
	if providedPassword == "" {
		return false, fmt.Errorf("password cannot be empty")
	}

	// Get stored hash and salt
	var hashB64, saltB64 *string
	var isPasswordProtected bool

	query := `
		SELECT is_password_protected, password_hash, password_salt
		FROM emails 
		WHERE email_id = ?
	`

	err := s.db.QueryRow(query, emailID).Scan(&isPasswordProtected, &hashB64, &saltB64)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, fmt.Errorf("email not found: %s", emailID)
		}
		return false, fmt.Errorf("failed to get email password data: %w", err)
	}

	// If email is not password-protected, return true
	if !isPasswordProtected {
		return true, nil
	}

	// If password-protected but no hash/salt stored, return false
	if hashB64 == nil || saltB64 == nil || *hashB64 == "" || *saltB64 == "" {
		return false, fmt.Errorf("email is password-protected but no password hash found")
	}

	// Decode stored hash and salt
	storedHash, err := base64.StdEncoding.DecodeString(*hashB64)
	if err != nil {
		return false, fmt.Errorf("failed to decode stored hash: %w", err)
	}

	storedSalt, err := base64.StdEncoding.DecodeString(*saltB64)
	if err != nil {
		return false, fmt.Errorf("failed to decode stored salt: %w", err)
	}

	// Hash the provided password with the stored salt
	providedHash := argon2.IDKey(
		[]byte(providedPassword),
		storedSalt,
		s.config.Time,
		s.config.Memory,
		s.config.Parallelism,
		s.config.KeyLength,
	)

	// Compare hashes
	if len(storedHash) != len(providedHash) {
		return false, nil
	}

	// Use constant-time comparison to prevent timing attacks
	for i := 0; i < len(storedHash); i++ {
		if storedHash[i] != providedHash[i] {
			return false, nil
		}
	}

	return true, nil
}

// ClearEmailPassword removes password protection from an email
func (s *EmailPasswordService) ClearEmailPassword(emailID string) error {
	// Validate input
	if emailID == "" {
		return fmt.Errorf("email ID cannot be empty")
	}

	// Clear password protection
	query := `
		UPDATE emails 
		SET is_password_protected = FALSE,
		    password_hash = NULL,
		    password_salt = NULL
		WHERE email_id = ?
	`

	result, err := s.db.Exec(query, emailID)
	if err != nil {
		return fmt.Errorf("failed to clear email password: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("email not found: %s", emailID)
	}

	return nil
}

// IsEmailPasswordProtected checks if an email requires password protection
func (s *EmailPasswordService) IsEmailPasswordProtected(emailID string) (bool, error) {
	// Validate input
	if emailID == "" {
		return false, fmt.Errorf("email ID cannot be empty")
	}

	// Check if email is password-protected
	var isPasswordProtected bool

	query := `
		SELECT is_password_protected
		FROM emails 
		WHERE email_id = ?
	`

	err := s.db.QueryRow(query, emailID).Scan(&isPasswordProtected)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, fmt.Errorf("email not found: %s", emailID)
		}
		return false, fmt.Errorf("failed to check email password protection: %w", err)
	}

	return isPasswordProtected, nil
}

// ValidatePasswordStrength validates password strength requirements
func (s *EmailPasswordService) ValidatePasswordStrength(password string) error {
	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	if len(password) > 128 {
		return fmt.Errorf("password must be no more than 128 characters long")
	}

	// Check for common weak passwords
	weakPasswords := []string{
		"password", "123456", "12345678", "qwerty", "abc123",
		"password123", "admin", "letmein", "welcome", "monkey",
	}

	lowerPassword := strings.ToLower(password)
	for _, weak := range weakPasswords {
		if lowerPassword == weak {
			return fmt.Errorf("password is too common, please choose a stronger password")
		}
	}

	return nil
}

// GetPasswordProtectionStatus returns the password protection status for an email
type PasswordProtectionStatus struct {
	EmailID            string `json:"email_id"`
	IsPasswordProtected bool   `json:"is_password_protected"`
	HasPasswordSet     bool   `json:"has_password_set"`
}

func (s *EmailPasswordService) GetPasswordProtectionStatus(emailID string) (*PasswordProtectionStatus, error) {
	// Validate input
	if emailID == "" {
		return nil, fmt.Errorf("email ID cannot be empty")
	}

	// Get password protection status
	var isPasswordProtected bool
	var passwordHash *string

	query := `
		SELECT is_password_protected, password_hash
		FROM emails 
		WHERE email_id = ?
	`

	err := s.db.QueryRow(query, emailID).Scan(&isPasswordProtected, &passwordHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("email not found: %s", emailID)
		}
		return nil, fmt.Errorf("failed to get password protection status: %w", err)
	}

	hasPasswordSet := passwordHash != nil && *passwordHash != ""

	return &PasswordProtectionStatus{
		EmailID:            emailID,
		IsPasswordProtected: isPasswordProtected,
		HasPasswordSet:     hasPasswordSet,
	}, nil
}
