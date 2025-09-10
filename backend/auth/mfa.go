package auth

// DO NOT EDIT EXISTING CODE - new file added
// MFA/TOTP handlers: setup, validation, and backup codes

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// MFA encryption key (in production, this should come from environment)
var mfaEncryptionKey = []byte("your-32-byte-encryption-key-here") // 32 bytes for AES-256

// type setupMFARequest struct {
//	UserID string `json:"user_id"` // Optional, can be extracted from session
// }

type setupMFAResponse struct {
	Secret      string   `json:"secret"`       // TOTP secret for QR code generation
	QRCodeURL   string   `json:"qr_code_url"`  // URL for QR code
	BackupCodes []string `json:"backup_codes"` // One-time backup codes
}

type validateMFARequest struct {
	Code string `json:"code"` // TOTP code or backup code
}

type validateMFAResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// encryptTOTPSecret encrypts TOTP secret using AES-256-GCM
func encryptTOTPSecret(secret string) (string, error) {
	block, err := aes.NewCipher(mfaEncryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(secret), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decryptTOTPSecret decrypts TOTP secret using AES-256-GCM
func decryptTOTPSecret(encryptedSecret string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encryptedSecret)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(mfaEncryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// generateBackupCodes creates 10 one-time backup codes
func generateBackupCodes() []string {
	codes := make([]string, 10)
	for i := 0; i < 10; i++ {
		code, _ := GenerateRandomToken(8) // 8-character backup codes
		codes[i] = code
	}
	return codes
}

// SetupMFAForUser sets up MFA for a user during signup (internal function)
func SetupMFAForUser(userID, username string) (*setupMFAResponse, error) {
	// Generate TOTP secret
	secret, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "SecureEmail",
		AccountName: username,
		Period:      30,
		Digits:      otp.DigitsSix,
		Algorithm:   otp.AlgorithmSHA1,
	})
	if err != nil {
		return nil, fmt.Errorf("generating TOTP secret: %v", err)
	}

	// Encrypt the secret
	encryptedSecret, err := encryptTOTPSecret(secret.Secret())
	if err != nil {
		return nil, fmt.Errorf("encrypting TOTP secret: %v", err)
	}

	// Generate backup codes
	backupCodes := generateBackupCodes()

	// Store encrypted secret and backup codes in database
	tx, err := DB.Begin()
	if err != nil {
		return nil, fmt.Errorf("beginning transaction: %v", err)
	}
	defer tx.Rollback()

	// Update user with MFA enabled
	_, err = tx.Exec(`
		UPDATE users 
		SET mfa_enabled = true, mfa_secret = ?
		WHERE id = ?
	`, encryptedSecret, userID)
	if err != nil {
		return nil, fmt.Errorf("updating user MFA: %v", err)
	}

	// Store backup codes
	for _, code := range backupCodes {
		hashedCode := HashToken(code)
		_, err = tx.Exec(`
			INSERT INTO recovery_codes (user_id, code_hash, used, created_at)
			VALUES (?, ?, false, ?)
		`, userID, hashedCode, time.Now().UTC())
		if err != nil {
			return nil, fmt.Errorf("storing backup code: %v", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("committing transaction: %v", err)
	}

	return &setupMFAResponse{
		Secret:      secret.Secret(),
		QRCodeURL:   secret.URL(),
		BackupCodes: backupCodes,
	}, nil
}

// POST /api/auth/setup-mfa
func SetupMFAHandler(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from session context (set by TokenAuthMiddleware)
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if MFA is already enabled
	var mfaEnabled bool
	err := DB.QueryRow("SELECT mfa_enabled FROM users WHERE id = ?", userID).Scan(&mfaEnabled)
	if err != nil {
		log.Printf("ERROR MFA setup lookup: %v", err)
		http.Error(w, "database error", http.StatusServiceUnavailable)
		return
	}

	if mfaEnabled {
		http.Error(w, "MFA already enabled", http.StatusConflict)
		return
	}

	// Generate TOTP secret
	secret, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "SecureEmail",
		AccountName: userID, // In production, use email or username
		Period:      30,
		Digits:      otp.DigitsSix,
		Algorithm:   otp.AlgorithmSHA1,
	})
	if err != nil {
		log.Printf("ERROR generating TOTP secret: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Encrypt the secret
	encryptedSecret, err := encryptTOTPSecret(secret.Secret())
	if err != nil {
		log.Printf("ERROR encrypting TOTP secret: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Generate backup codes
	backupCodes := generateBackupCodes()
	hashedBackupCodes := make([]string, len(backupCodes))
	for i, code := range backupCodes {
		hashedBackupCodes[i] = HashToken(code)
	}

	// Store encrypted secret and hashed backup codes
	backupCodesJSON, err := json.Marshal(hashedBackupCodes)
	if err != nil {
		log.Printf("ERROR marshaling backup codes: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	_, err = DB.Exec(`
		UPDATE users 
		SET totp_secret = ?, 
		    backup_codes_hashed = ?,
		    mfa_enabled = TRUE,
		    updated_at = ?
		WHERE id = ?
	`, encryptedSecret, string(backupCodesJSON), time.Now().UTC(), userID)

	if err != nil {
		log.Printf("ERROR storing MFA setup: %v", err)
		http.Error(w, "database error", http.StatusServiceUnavailable)
		return
	}

	// Log MFA setup (non-sensitive)
	log.Printf("INFO mfa_setup user_id=%s", userID)

	resp := setupMFAResponse{
		Secret:      secret.Secret(),
		QRCodeURL:   secret.URL(),
		BackupCodes: backupCodes, // Return raw codes to user (they won't be shown again)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// POST /api/auth/validate-mfa
func ValidateMFAHandler(w http.ResponseWriter, r *http.Request) {
	var req validateMFARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Code == "" {
		http.Error(w, "missing MFA code", http.StatusBadRequest)
		return
	}

	// Extract user ID from session context
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user's encrypted TOTP secret and backup codes
	var encryptedSecret string
	var backupCodesJSON string
	err := DB.QueryRow(`
		SELECT totp_secret, backup_codes_hashed 
		FROM users 
		WHERE id = ? AND mfa_enabled = TRUE
	`, userID).Scan(&encryptedSecret, &backupCodesJSON)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "MFA not enabled", http.StatusBadRequest)
			return
		}
		log.Printf("ERROR MFA validation lookup: %v", err)
		http.Error(w, "database error", http.StatusServiceUnavailable)
		return
	}

	// Decrypt TOTP secret
	secret, err := decryptTOTPSecret(encryptedSecret)
	if err != nil {
		log.Printf("ERROR decrypting TOTP secret: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Try TOTP validation first
	valid := totp.Validate(req.Code, secret)
	if valid {
		// Log successful MFA validation (non-sensitive)
		log.Printf("INFO mfa_validation_success user_id=%s method=totp", userID)

		resp := validateMFAResponse{
			Success: true,
			Message: "MFA validation successful",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Try backup code validation
	var backupCodes []string
	if err := json.Unmarshal([]byte(backupCodesJSON), &backupCodes); err != nil {
		log.Printf("ERROR unmarshaling backup codes: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	hashedInputCode := HashToken(req.Code)
	for i, hashedCode := range backupCodes {
		if hashedCode == hashedInputCode {
			// Remove used backup code
			backupCodes = append(backupCodes[:i], backupCodes[i+1:]...)
			updatedBackupCodesJSON, _ := json.Marshal(backupCodes)

			_, err := DB.Exec(`
				UPDATE users 
				SET backup_codes_hashed = ?,
				    updated_at = ?
				WHERE id = ?
			`, string(updatedBackupCodesJSON), time.Now().UTC(), userID)

			if err != nil {
				log.Printf("ERROR updating backup codes: %v", err)
			}

			// Log successful backup code usage (non-sensitive)
			log.Printf("INFO mfa_validation_success user_id=%s method=backup_code", userID)

			resp := validateMFAResponse{
				Success: true,
				Message: "MFA validation successful",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
	}

	// Log failed MFA validation (non-sensitive)
	log.Printf("WARN mfa_validation_failed user_id=%s", userID)

	http.Error(w, "invalid MFA code", http.StatusUnauthorized)
}
