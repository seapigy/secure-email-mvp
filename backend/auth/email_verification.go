package auth

// DO NOT EDIT EXISTING CODE - new file added
// Email verification handlers: generate, store, and validate verification codes

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type verifyEmailRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type verifyEmailResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type resendVerificationRequest struct {
	Email string `json:"email"`
}

type resendVerificationResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// POST /api/auth/verify-email
func VerifyEmailHandler(w http.ResponseWriter, r *http.Request) {
	var req verifyEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Code == "" {
		http.Error(w, "missing email or code", http.StatusBadRequest)
		return
	}

	// Hash the provided code for comparison
	hashedCode := HashToken(req.Code)

	// Look up user and verification code
	var userID string
	var storedHashedCode string
	var expiresAt time.Time
	var emailVerified bool

	err := DB.QueryRow(`
		SELECT id, verification_code, verification_code_expires_at, email_verified 
		FROM users 
		WHERE email = ? AND verification_code = ?
	`, req.Email, hashedCode).Scan(&userID, &storedHashedCode, &expiresAt, &emailVerified)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "invalid verification code", http.StatusUnauthorized)
			return
		}
		log.Printf("ERROR verification lookup: %v", err)
		http.Error(w, "database error", http.StatusServiceUnavailable)
		return
	}

	// Check if already verified
	if emailVerified {
		http.Error(w, "email already verified", http.StatusConflict)
		return
	}

	// Check if code is expired
	if time.Now().After(expiresAt) {
		http.Error(w, "verification code expired", http.StatusUnauthorized)
		return
	}

	// Mark email as verified and clear verification code
	_, err = DB.Exec(`
		UPDATE users 
		SET email_verified = TRUE, 
		    verification_code = NULL, 
		    verification_code_expires_at = NULL,
		    updated_at = ?
		WHERE id = ?
	`, time.Now().UTC(), userID)

	if err != nil {
		log.Printf("ERROR updating email verification: %v", err)
		http.Error(w, "database error", http.StatusServiceUnavailable)
		return
	}

	// Log successful verification (non-sensitive)
	log.Printf("INFO email_verified user_id=%s", userID)

	resp := verifyEmailResponse{
		Success: true,
		Message: "Email verified successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// POST /api/auth/resend-verification
func ResendVerificationHandler(w http.ResponseWriter, r *http.Request) {
	var req resendVerificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		http.Error(w, "missing email", http.StatusBadRequest)
		return
	}

	// Check if user exists and is not already verified
	var userID string
	var emailVerified bool
	err := DB.QueryRow(`
		SELECT id, email_verified 
		FROM users 
		WHERE email = ?
	`, req.Email).Scan(&userID, &emailVerified)

	if err != nil {
		if err == sql.ErrNoRows {
			// Don't reveal if email exists or not
			resp := resendVerificationResponse{
				Success: true,
				Message: "If the email exists and is unverified, a new verification code has been sent",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
		log.Printf("ERROR resend verification lookup: %v", err)
		http.Error(w, "database error", http.StatusServiceUnavailable)
		return
	}

	// If already verified, don't reveal this
	if emailVerified {
		resp := resendVerificationResponse{
			Success: true,
			Message: "If the email exists and is unverified, a new verification code has been sent",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Generate new verification code
	verificationCode, err := GenerateRandomToken(6) // 6-character code
	if err != nil {
		log.Printf("ERROR generating verification code: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	hashedCode := HashToken(verificationCode)
	expiresAt := time.Now().Add(30 * time.Minute).UTC() // 30-minute expiry

	// Update user with new verification code
	_, err = DB.Exec(`
		UPDATE users 
		SET verification_code = ?, 
		    verification_code_expires_at = ?,
		    updated_at = ?
		WHERE id = ?
	`, hashedCode, expiresAt, time.Now().UTC(), userID)

	if err != nil {
		log.Printf("ERROR updating verification code: %v", err)
		http.Error(w, "database error", http.StatusServiceUnavailable)
		return
	}

	// TODO: Send email with verification code
	// For now, log the code (in production, this would be sent via email service)
	log.Printf("INFO verification_code_generated user_id=%s code=%s", userID, verificationCode)

	// Log successful resend (non-sensitive)
	log.Printf("INFO verification_resent user_id=%s", userID)

	resp := resendVerificationResponse{
		Success: true,
		Message: "If the email exists and is unverified, a new verification code has been sent",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
