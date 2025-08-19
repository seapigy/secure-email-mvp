package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"secure-email-mvp/pkg/auth"
	"secure-email-mvp/pkg/password"

	"golang.org/x/crypto/bcrypt"
)

// PasswordResetRequest represents the JSON request body for password reset
type PasswordResetRequest struct {
	Email       string `json:"email"`
	NewPassword string `json:"new_password"`
	ResetToken  string `json:"reset_token"`
}

// PasswordResetResponse represents the JSON response for password reset
type PasswordResetResponse struct {
	Message string `json:"message"`
}

// passwordResetHandler handles password reset requests with comprehensive validation
func passwordResetHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only allow POST method
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Set content type to JSON
		w.Header().Set("Content-Type", "application/json")

		// Parse JSON request body
		var req PasswordResetRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"Invalid JSON format"}`, http.StatusBadRequest)
			return
		}

		// Validate required fields
		if req.Email == "" || req.NewPassword == "" || req.ResetToken == "" {
			http.Error(w, `{"error":"Email, new password, and reset token are required"}`, http.StatusBadRequest)
			return
		}

		// Validate email format
		if !isValidEmail(req.Email) {
			http.Error(w, `{"error":"Invalid email format"}`, http.StatusBadRequest)
			return
		}

		// Validate new password using comprehensive password service
		passwordService := password.NewPasswordService()
		ctx := context.Background()

		passwordResult, err := passwordService.ValidatePassword(ctx, req.NewPassword)
		if err != nil {
			log.Printf("Password validation failed during reset for email %s: %v", req.Email, err)
			http.Error(w, `{"error":"Password validation error"}`, http.StatusInternalServerError)
			return
		}

		if !passwordResult.IsValid {
			// Log validation failures for audit
			log.Printf("Password reset validation failed for email %s: %v", req.Email, passwordResult.Errors)

			// Return generic error message without revealing specific validation details
			http.Error(w, `{"error":"New password does not meet security requirements"}`, http.StatusBadRequest)
			return
		}

		// Log successful password validation
		log.Printf("Password reset validation passed for email %s (score: %d, breach count: %d)",
			req.Email, passwordResult.Score, passwordResult.BreachCount)

		// Verify reset token and update password
		err = verifyAndUpdatePassword(db, req.Email, req.NewPassword, req.ResetToken)
		if err != nil {
			if err.Error() == "invalid token" {
				http.Error(w, `{"error":"Invalid or expired reset token"}`, http.StatusBadRequest)
				return
			}
			if err.Error() == "user not found" {
				http.Error(w, `{"error":"User not found"}`, http.StatusNotFound)
				return
			}
			log.Printf("Password reset failed for email %s: %v", req.Email, err)
			http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
			return
		}

		// Return success response
		response := PasswordResetResponse{
			Message: "Password updated successfully",
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

// verifyAndUpdatePassword verifies the reset token and updates the user's password
func verifyAndUpdatePassword(db *sql.DB, email, newPassword, resetToken string) error {
	// First, check if user exists
	var userID int
	err := db.QueryRow("SELECT id FROM users WHERE email = ?", email).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user not found")
		}
		return err
	}

	// Check if reset token is valid and not expired
	var tokenExpiration time.Time
	err = db.QueryRow("SELECT reset_token_expiration FROM users WHERE email = ? AND reset_token = ?",
		email, resetToken).Scan(&tokenExpiration)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("invalid token")
		}
		return err
	}

	// Check if token is expired
	if time.Now().After(tokenExpiration) {
		return fmt.Errorf("invalid token")
	}

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Update password and clear reset token
	_, err = db.Exec("UPDATE users SET password = ?, reset_token = NULL, reset_token_expiration = NULL, updated_at = ? WHERE email = ?",
		string(hashedPassword), time.Now(), email)
	if err != nil {
		return err
	}

	log.Printf("Password successfully updated for user %s", email)
	return nil
}

// initiatePasswordResetHandler handles the initiation of password reset
func initiatePasswordResetHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only allow POST method
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Set content type to JSON
		w.Header().Set("Content-Type", "application/json")

		// Parse JSON request body
		var req struct {
			Email string `json:"email"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"Invalid JSON format"}`, http.StatusBadRequest)
			return
		}

		// Validate email
		if req.Email == "" {
			http.Error(w, `{"error":"Email is required"}`, http.StatusBadRequest)
			return
		}

		if !isValidEmail(req.Email) {
			http.Error(w, `{"error":"Invalid email format"}`, http.StatusBadRequest)
			return
		}

		// Generate reset token and send email
		err := generateAndSendResetToken(db, req.Email)
		if err != nil {
			if err.Error() == "user not found" {
				// Don't reveal if user exists or not
				log.Printf("Password reset initiated for non-existent email: %s", req.Email)
			} else {
				log.Printf("Password reset initiation failed for email %s: %v", req.Email, err)
			}
			// Always return success to prevent email enumeration
			response := map[string]string{
				"message": "If the email exists, a password reset link has been sent",
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
			return
		}

		// Return success response
		response := map[string]string{
			"message": "If the email exists, a password reset link has been sent",
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

// generateAndSendResetToken generates a reset token and sends it via email
func generateAndSendResetToken(db *sql.DB, email string) error {
	// Check if user exists
	var userID int
	err := db.QueryRow("SELECT id FROM users WHERE email = ?", email).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user not found")
		}
		return err
	}

	// Generate reset token and expiration
	resetToken := auth.GenerateResetToken()
	resetExpiration := time.Now().Add(1 * time.Hour) // Token expires in 1 hour

	// Update user with reset token
	_, err = db.Exec("UPDATE users SET reset_token = ?, reset_token_expiration = ? WHERE email = ?",
		resetToken, resetExpiration, email)
	if err != nil {
		return err
	}

	// Send reset email (stub implementation)
	if err := auth.SendPasswordResetEmail(email, resetToken); err != nil {
		log.Printf("Failed to send password reset email to %s: %v", email, err)
		// Don't fail the request if email sending fails
	}

	log.Printf("Password reset token generated for user %s", email)
	return nil
}









