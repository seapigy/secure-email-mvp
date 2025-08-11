package mfa

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"secure-email-mvp/pkg/auth"

	"github.com/pquerna/otp/totp"
)

// MFAType represents the type of MFA required
type MFAType string

const (
	MFATypeTOTP      MFAType = "TOTP"
	MFATypeEmailCode MFAType = "EMAIL_CODE"
)

// MFAConfig represents the MFA configuration for an email
type MFAConfig struct {
	RequireMFA          bool       `json:"require_mfa"`
	MFAType             MFAType    `json:"mfa_type"`
	EncryptedTOTPSecret string     `json:"encrypted_totp_secret,omitempty"`
	FailedAttempts      int        `json:"failed_attempts"`
	LockedUntil         *time.Time `json:"locked_until,omitempty"`
}

// MFARequest represents a request to validate MFA
type MFARequest struct {
	EmailID string `json:"email_id"`
	MFACode string `json:"mfa_code"`
}

// MFAResponse represents the response from MFA validation
type MFAResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// EmailCodeRequest represents a request to send an email-based MFA code
type EmailCodeRequest struct {
	EmailID string `json:"email_id"`
}

// EmailCodeResponse represents the response from email code generation
type EmailCodeResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// TOTPConfig represents TOTP configuration for QR code generation
type TOTPConfig struct {
	Secret    string `json:"secret"`
	QRCodeURL string `json:"qr_code_url"`
	Issuer    string `json:"issuer"`
	Account   string `json:"account"`
}

// MFAService provides MFA functionality
type MFAService struct {
	db *sql.DB
}

// NewMFAService creates a new MFA service
func NewMFAService(db *sql.DB) *MFAService {
	return &MFAService{db: db}
}

// GenerateTOTPSecret generates a new TOTP secret and encrypts it
func (m *MFAService) GenerateTOTPSecret(emailID string) (*TOTPConfig, error) {
	// Generate a new TOTP secret
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Secure Email MVP",
		AccountName: emailID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate TOTP secret: %w", err)
	}

	// Encrypt the TOTP secret using AES-256-GCM
	encryptedData, err := auth.EncryptAES256GCM([]byte(key.Secret()))
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt TOTP secret: %w", err)
	}

	// Convert encrypted data to base64 for storage
	encryptedSecret := base64.StdEncoding.EncodeToString(encryptedData.Ciphertext)
	encryptedKey := base64.StdEncoding.EncodeToString(encryptedData.Key)
	encryptedNonce := base64.StdEncoding.EncodeToString(encryptedData.Nonce)
	encryptedAuthTag := base64.StdEncoding.EncodeToString(encryptedData.AuthTag)

	// Store encrypted components as JSON
	encryptedComponents := map[string]string{
		"ciphertext": encryptedSecret,
		"key":        encryptedKey,
		"nonce":      encryptedNonce,
		"auth_tag":   encryptedAuthTag,
	}

	_, err = json.Marshal(encryptedComponents)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal encrypted components: %w", err)
	}

	return &TOTPConfig{
		Secret:    key.Secret(),
		QRCodeURL: key.URL(),
		Issuer:    key.Issuer(),
		Account:   key.AccountName(),
	}, nil
}

// ValidateTOTP validates a TOTP code for an email
func (m *MFAService) ValidateTOTP(emailID, code string) (bool, error) {
	// Get the encrypted TOTP secret from the database
	var encryptedTOTPSecret string
	err := m.db.QueryRow("SELECT encrypted_totp_secret FROM emails WHERE email_id = ?", emailID).Scan(&encryptedTOTPSecret)
	if err != nil {
		return false, fmt.Errorf("failed to get TOTP secret: %w", err)
	}

	// Parse the encrypted components
	var encryptedComponents map[string]string
	if err := json.Unmarshal([]byte(encryptedTOTPSecret), &encryptedComponents); err != nil {
		return false, fmt.Errorf("failed to parse encrypted components: %w", err)
	}

	// Decode the encrypted components
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedComponents["ciphertext"])
	if err != nil {
		return false, fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	key, err := base64.StdEncoding.DecodeString(encryptedComponents["key"])
	if err != nil {
		return false, fmt.Errorf("failed to decode key: %w", err)
	}

	nonce, err := base64.StdEncoding.DecodeString(encryptedComponents["nonce"])
	if err != nil {
		return false, fmt.Errorf("failed to decode nonce: %w", err)
	}

	authTag, err := base64.StdEncoding.DecodeString(encryptedComponents["auth_tag"])
	if err != nil {
		return false, fmt.Errorf("failed to decode auth tag: %w", err)
	}

	// Create EncryptedData struct
	encryptedData := &auth.EncryptedData{
		Ciphertext: ciphertext,
		Key:        key,
		Nonce:      nonce,
		AuthTag:    authTag,
	}

	// Decrypt the TOTP secret
	secretBytes, err := auth.DecryptAES256GCM(encryptedData)
	if err != nil {
		return false, fmt.Errorf("failed to decrypt TOTP secret: %w", err)
	}

	secret := string(secretBytes)

	// Validate the TOTP code
	valid := totp.Validate(code, secret)
	return valid, nil
}

// GenerateEmailCode generates a random 6-digit code for email-based MFA
func (m *MFAService) GenerateEmailCode() (string, error) {
	// Generate a random 6-digit code
	code := make([]byte, 3)
	if _, err := rand.Read(code); err != nil {
		return "", fmt.Errorf("failed to generate random code: %w", err)
	}

	// Convert to 6-digit number (0-999999)
	codeNum := int(code[0])<<16 | int(code[1])<<8 | int(code[2])
	codeNum = codeNum % 1000000 // Ensure it's 6 digits

	// Format as 6-digit string with leading zeros
	return fmt.Sprintf("%06d", codeNum), nil
}

// StoreEmailCode stores an email code temporarily (in production, this would be in Redis or similar)
func (m *MFAService) StoreEmailCode(emailID, code string) error {
	// For this implementation, we'll store the code in the database with an expiration
	// In production, this should be in a fast cache like Redis
	expiresAt := time.Now().Add(10 * time.Minute) // 10-minute expiration

	_, err := m.db.Exec(`
		UPDATE emails SET 
			encrypted_totp_secret = ?,
			mfa_locked_until = ?
		WHERE email_id = ?`,
		code, expiresAt, emailID,
	)
	return err
}

// ValidateEmailCode validates an email-based MFA code
func (m *MFAService) ValidateEmailCode(emailID, code string) (bool, error) {
	// Get the stored code and expiration
	var storedCode string
	var expiresAt time.Time
	err := m.db.QueryRow(`
		SELECT encrypted_totp_secret, mfa_locked_until 
		FROM emails 
		WHERE email_id = ?`, emailID).Scan(&storedCode, &expiresAt)
	if err != nil {
		return false, fmt.Errorf("failed to get stored code: %w", err)
	}

	// Check if code has expired
	if time.Now().After(expiresAt) {
		return false, fmt.Errorf("email code has expired")
	}

	// Compare codes (case-insensitive)
	return strings.EqualFold(code, storedCode), nil
}

// CheckMFALockout checks if MFA is locked due to too many failed attempts
func (m *MFAService) CheckMFALockout(emailID string) (bool, *time.Time, error) {
	var lockedUntil *time.Time
	err := m.db.QueryRow("SELECT mfa_locked_until FROM emails WHERE email_id = ?", emailID).Scan(&lockedUntil)
	if err != nil {
		return false, nil, fmt.Errorf("failed to check MFA lockout: %w", err)
	}

	if lockedUntil != nil && time.Now().Before(*lockedUntil) {
		return true, lockedUntil, nil
	}

	return false, nil, nil
}

// IncrementFailedAttempts increments the failed MFA attempts counter
func (m *MFAService) IncrementFailedAttempts(emailID string) error {
	// Get current failed attempts
	var failedAttempts int
	err := m.db.QueryRow("SELECT mfa_failed_attempts FROM emails WHERE email_id = ?", emailID).Scan(&failedAttempts)
	if err != nil {
		return fmt.Errorf("failed to get failed attempts: %w", err)
	}

	failedAttempts++

	// Check if we should lock the account
	maxAttempts := 5 // Configurable
	var lockedUntil *time.Time
	if failedAttempts >= maxAttempts {
		lockoutDuration := 30 * time.Minute // 30-minute lockout
		lockoutTime := time.Now().Add(lockoutDuration)
		lockedUntil = &lockoutTime
	}

	// Update the database
	_, err = m.db.Exec(`
		UPDATE emails SET 
			mfa_failed_attempts = ?,
			mfa_locked_until = ?
		WHERE email_id = ?`,
		failedAttempts, lockedUntil, emailID,
	)
	return err
}

// ResetFailedAttempts resets the failed MFA attempts counter
func (m *MFAService) ResetFailedAttempts(emailID string) error {
	_, err := m.db.Exec(`
		UPDATE emails SET 
			mfa_failed_attempts = 0,
			mfa_locked_until = NULL
		WHERE email_id = ?`,
		emailID,
	)
	return err
}

// GetMFAConfig retrieves the MFA configuration for an email
func (m *MFAService) GetMFAConfig(emailID string) (*MFAConfig, error) {
	var requireMFA int
	var mfaType sql.NullString
	var encryptedTOTPSecret sql.NullString
	var failedAttempts int
	var lockedUntil sql.NullTime

	err := m.db.QueryRow(`
		SELECT require_mfa, mfa_type, encrypted_totp_secret, mfa_failed_attempts, mfa_locked_until
		FROM emails WHERE email_id = ?`, emailID).Scan(
		&requireMFA, &mfaType, &encryptedTOTPSecret, &failedAttempts, &lockedUntil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get MFA config: %w", err)
	}

	config := &MFAConfig{
		RequireMFA:     requireMFA == 1,
		FailedAttempts: failedAttempts,
	}

	if mfaType.Valid {
		config.MFAType = MFAType(mfaType.String)
	}

	if encryptedTOTPSecret.Valid {
		config.EncryptedTOTPSecret = encryptedTOTPSecret.String
	}

	if lockedUntil.Valid {
		config.LockedUntil = &lockedUntil.Time
	}

	return config, nil
}
