package auth

// DO NOT EDIT EXISTING CODE - new file added
// Account recovery handler (Go). Assumes a global DB variable is set by main application.
//
// Required env:
// - DATABASE_URL used by main to open DB and set auth.DB

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type recoveryRequest struct {
	FallbackEmail string `json:"fallback_email"`
	RecoveryKey   string `json:"recovery_key"`
	NewPassword   string `json:"new_password,omitempty"`
	NewEmail      string `json:"new_email,omitempty"`
	Action        string `json:"action"` // "reset_password" or "reset_email"
}

type recoveryResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	UserID  string `json:"user_id,omitempty"`
}

// POST /api/account/recover
func RecoveryHandler(w http.ResponseWriter, r *http.Request) {
	// Basic JSON decode
	var req recoveryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.FallbackEmail == "" || req.RecoveryKey == "" || req.Action == "" {
		http.Error(w, "missing required fields", http.StatusBadRequest)
		return
	}

	// Validate action
	if req.Action != "reset_password" && req.Action != "reset_email" {
		http.Error(w, "invalid action", http.StatusBadRequest)
		return
	}

	// For password reset, new password is required
	if req.Action == "reset_password" && req.NewPassword == "" {
		http.Error(w, "new password required for password reset", http.StatusBadRequest)
		return
	}

	// For email reset, new email is required
	if req.Action == "reset_email" && req.NewEmail == "" {
		http.Error(w, "new email required for email reset", http.StatusBadRequest)
		return
	}

	// Hash the provided recovery key
	hashedRecoveryKey := HashToken(req.RecoveryKey)

	// Find user by fallback email and verify recovery key
	var userID string
	var fallbackEmailVerified bool
	var currentEmail string
	var currentUsername string

	err := DB.QueryRow(`
		SELECT id, fallback_email_verified, email, username 
		FROM users 
		WHERE fallback_email = ? AND recovery_private_key_hashed = ?
	`, req.FallbackEmail, hashedRecoveryKey).Scan(&userID, &fallbackEmailVerified, &currentEmail, &currentUsername)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "invalid fallback email or recovery key", http.StatusUnauthorized)
			return
		}
		log.Printf("ERROR querying user for recovery: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Verify fallback email is verified
	if !fallbackEmailVerified {
		http.Error(w, "fallback email must be verified before recovery", http.StatusForbidden)
		return
	}

	// Start transaction for recovery operations
	tx, err := DB.Begin()
	if err != nil {
		log.Printf("ERROR begin tx for recovery: %v", err)
		http.Error(w, "database error", http.StatusServiceUnavailable)
		return
	}
	defer tx.Rollback()

	now := time.Now().UTC()

	switch req.Action {
	case "reset_password":
		// Hash new password
		hashedPassword, err := HashPassword(req.NewPassword)
		if err != nil {
			log.Printf("ERROR hashing new password: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		// Update password
		_, err = tx.Exec(`
			UPDATE users 
			SET hashed_password = ?, updated_at = ?
			WHERE id = ?
		`, hashedPassword, now, userID)

		if err != nil {
			log.Printf("ERROR updating password: %v", err)
			http.Error(w, "database error", http.StatusServiceUnavailable)
			return
		}

		// Invalidate all existing sessions
		_, err = tx.Exec(`
			DELETE FROM sessions 
			WHERE user_id = ?
		`, userID)

		if err != nil {
			log.Printf("ERROR invalidating sessions: %v", err)
			http.Error(w, "database error", http.StatusServiceUnavailable)
			return
		}

	case "reset_email":
		// Check if new email is already in use
		var exists int
		err = tx.QueryRow("SELECT 1 FROM users WHERE email = ? AND id != ? LIMIT 1", req.NewEmail, userID).Scan(&exists)
		if err != nil && err != sql.ErrNoRows {
			log.Printf("ERROR checking email uniqueness: %v", err)
			http.Error(w, "database error", http.StatusServiceUnavailable)
			return
		}
		if err == nil {
			http.Error(w, "email already in use", http.StatusConflict)
			return
		}

		// Generate verification code for new email
		verificationCode, err := GenerateRandomToken(6)
		if err != nil {
			log.Printf("ERROR generating verification code: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		hashedVerificationCode := HashToken(verificationCode)
		verificationExpiresAt := now.Add(30 * time.Minute)

		// Update email and set to pending verification
		_, err = tx.Exec(`
			UPDATE users 
			SET email = ?, email_verified = FALSE, verification_code = ?, verification_code_expires_at = ?, updated_at = ?
			WHERE id = ?
		`, req.NewEmail, hashedVerificationCode, verificationExpiresAt, now, userID)

		if err != nil {
			log.Printf("ERROR updating email: %v", err)
			http.Error(w, "database error", http.StatusServiceUnavailable)
			return
		}

		// Invalidate all existing sessions
		_, err = tx.Exec(`
			DELETE FROM sessions 
			WHERE user_id = ?
		`, userID)

		if err != nil {
			log.Printf("ERROR invalidating sessions: %v", err)
			http.Error(w, "database error", http.StatusServiceUnavailable)
			return
		}

		// TODO: Send verification email to new email address
		log.Printf("INFO email_reset_verification_code_generated user_id=%s new_email=%s code=%s", userID, req.NewEmail, verificationCode)
	default:
		http.Error(w, "invalid action", http.StatusBadRequest)
		return
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		log.Printf("ERROR commit recovery tx: %v", err)
		http.Error(w, "database error", http.StatusServiceUnavailable)
		return
	}

	// Log recovery event
	log.Printf("INFO account_recovery_successful user_id=%s action=%s", userID, req.Action)

	// Log analytics event
	LogAnalyticsEvent(userID, EventAccountRecovery, map[string]interface{}{
		"action":  req.Action,
		"success": true,
	})

	// Success response
	resp := recoveryResponse{
		Success: true,
		Message: "Recovery successful",
		UserID:  userID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
