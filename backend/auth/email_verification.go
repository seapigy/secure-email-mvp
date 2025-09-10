package auth

// DO NOT EDIT EXISTING CODE - new file added
// Email verification handler (Go). Assumes a global DB variable is set by main application.

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type verifyEmailRequest struct {
	UserID string `json:"user_id"`
	Code   string `json:"code"`
}

type verifyEmailResponse struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	RecoveryKey string `json:"recovery_key,omitempty"` // Only sent after successful verification
}

// POST /api/auth/verify-email
func VerifyEmailHandler(w http.ResponseWriter, r *http.Request) {
	// Basic JSON decode
	var req verifyEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.UserID == "" || req.Code == "" {
		http.Error(w, "missing user_id or code", http.StatusBadRequest)
		return
	}

	// Hash the provided verification code
	hashedCode := HashToken(req.Code)

	// Find user and verify code
	var userID string
	var fallbackEmail string
	var username string
	var fallbackEmailVerified bool
	var verificationCodeExpiresAt time.Time

	err := DB.QueryRow(`
		SELECT id, fallback_email, username, fallback_email_verified, verification_code_expires_at
		FROM users 
		WHERE id = ? AND verification_code = ?
	`, req.UserID, hashedCode).Scan(&userID, &fallbackEmail, &username, &fallbackEmailVerified, &verificationCodeExpiresAt)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "invalid user_id or verification code", http.StatusUnauthorized)
			return
		}
		log.Printf("ERROR querying user for email verification: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Check if already verified
	if fallbackEmailVerified {
		http.Error(w, "email already verified", http.StatusConflict)
		return
	}

	// Check if verification code has expired
	now := time.Now().UTC()
	if now.After(verificationCodeExpiresAt) {
		http.Error(w, "verification code has expired", http.StatusUnauthorized)
		return
	}

	// Start transaction for verification
	tx, err := DB.Begin()
	if err != nil {
		log.Printf("ERROR begin tx for email verification: %v", err)
		http.Error(w, "database error", http.StatusServiceUnavailable)
		return
	}
	defer tx.Rollback()

	// Mark fallback email as verified
	_, err = tx.Exec(`
		UPDATE users 
		SET fallback_email_verified = TRUE, updated_at = ?
		WHERE id = ?
	`, now, userID)

	if err != nil {
		log.Printf("ERROR updating email verification status: %v", err)
		http.Error(w, "database error", http.StatusServiceUnavailable)
		return
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		log.Printf("ERROR commit email verification tx: %v", err)
		http.Error(w, "database error", http.StatusServiceUnavailable)
		return
	}

	// Now send the recovery key to the verified email
	// Get the recovery key from the database (we need to generate a new one since we can't retrieve the original)
	recoveryKey, err := GenerateRandomToken(32) // 32 bytes = 256 bits
	if err != nil {
		log.Printf("ERROR generating recovery key for verified user: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	hashedRecoveryKey := HashToken(recoveryKey)

	// Update the recovery key in the database
	_, err = DB.Exec(`
		UPDATE users 
		SET recovery_private_key_hashed = ?, updated_at = ?
		WHERE id = ?
	`, hashedRecoveryKey, now, userID)

	if err != nil {
		log.Printf("ERROR updating recovery key: %v", err)
		http.Error(w, "database error", http.StatusServiceUnavailable)
		return
	}

	// Send recovery key via email
	if err := SendRecoveryKeyEmail(fallbackEmail, recoveryKey, username); err != nil {
		log.Printf("ERROR sending recovery key email: %v", err)
		// Don't fail verification if email sending fails, just log it
	} else {
		log.Printf("INFO recovery_key_email_sent user_id=%s fallback_email=%s", userID, fallbackEmail)
	}

	// Log verification success
	log.Printf("INFO email_verification_successful user_id=%s fallback_email=%s", userID, fallbackEmail)

	// Log analytics event
	LogAnalyticsEvent(userID, EventEmailVerified, map[string]interface{}{
		"success": true,
	})

	// Success response with recovery key
	resp := verifyEmailResponse{
		Success:     true,
		Message:     "Email verified successfully. Recovery key sent to your email.",
		RecoveryKey: recoveryKey, // Only sent after successful verification
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
