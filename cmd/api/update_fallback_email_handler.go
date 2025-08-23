package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"secure-email-mvp/pkg/auth"
)

// UpdateFallbackEmailRequest represents the JSON request body for updating fallback email
type UpdateFallbackEmailRequest struct {
	Email         string `json:"email"`
	FallbackEmail string `json:"fallback_email"`
}

// UpdateFallbackEmailResponse represents the JSON response for updating fallback email
type UpdateFallbackEmailResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

// updateFallbackEmailHandlerFactory returns an HTTP handler for updating fallback email on pending signups
func updateFallbackEmailHandlerFactory(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only allow POST method
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Set content type to JSON
		w.Header().Set("Content-Type", "application/json")

		// Parse JSON request body
		var req UpdateFallbackEmailRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"Invalid JSON format"}`, http.StatusBadRequest)
			return
		}

		// Validate email format
		if !isValidEmailFormatForUpdate(req.Email) {
			http.Error(w, `{"error":"Invalid email format"}`, http.StatusBadRequest)
			return
		}

		// Validate fallback email
		if req.FallbackEmail == "" {
			http.Error(w, `{"error":"Fallback email is required"}`, http.StatusBadRequest)
			return
		}
		if !isValidEmailFormatForUpdate(req.FallbackEmail) {
			http.Error(w, `{"error":"Invalid fallback email format"}`, http.StatusBadRequest)
			return
		}

		// Check if user is trying to use their own email as fallback
		if req.Email == req.FallbackEmail {
			http.Error(w, `{"error":"Fallback email cannot be the same as your main email"}`, http.StatusBadRequest)
			return
		}

		// Check if fallback email already exists as someone's main email
		var fallbackAsMainEmail int
		err := db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", req.FallbackEmail).Scan(&fallbackAsMainEmail)
		if err != nil {
			log.Printf("Database error checking fallback email as main email: %v", err)
			http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
			return
		}
		if fallbackAsMainEmail > 0 {
			http.Error(w, `{"error":"This email is already registered as a main account"}`, http.StatusBadRequest)
			return
		}

		// Check if fallback email is already used by another user as fallback
		var fallbackEmailInUse int
		err = db.QueryRow("SELECT COUNT(*) FROM users WHERE fallback_email = ?", req.FallbackEmail).Scan(&fallbackEmailInUse)
		if err != nil {
			log.Printf("Database error checking fallback email in use: %v", err)
			http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
			return
		}
		if fallbackEmailInUse > 0 {
			http.Error(w, `{"error":"This fallback email is already in use by another account"}`, http.StatusBadRequest)
			return
		}

		// Check if fallback email is used in other pending signups (excluding current one)
		var fallbackInPending int
		err = db.QueryRow("SELECT COUNT(*) FROM pending_signups WHERE fallback_email = ? AND email != ?", req.FallbackEmail, req.Email).Scan(&fallbackInPending)
		if err != nil {
			log.Printf("Database error checking fallback email in pending signups: %v", err)
			http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
			return
		}
		if fallbackInPending > 0 {
			http.Error(w, `{"error":"This fallback email is already in use by another pending signup"}`, http.StatusBadRequest)
			return
		}

		// Check if pending signup exists
		pendingSignupService := auth.NewPendingSignupService(db)
		pendingSignup, err := pendingSignupService.GetPendingSignupByEmail(req.Email)
		if err != nil {
			log.Printf("Failed to get pending signup for email %s: %v", req.Email, err)
			http.Error(w, `{"error":"No pending signup found for this email"}`, http.StatusNotFound)
			return
		}

		// Generate new fallback token and expiration
		newFallbackToken := auth.GenerateFallbackToken(req.FallbackEmail)
		newFallbackExpiration := auth.GenerateFallbackExpiration()

		// Update the pending signup with new fallback email and token
		err = updatePendingSignupFallbackEmail(db, pendingSignup.ID, req.FallbackEmail, newFallbackToken, newFallbackExpiration)
		if err != nil {
			log.Printf("Failed to update pending signup fallback email: %v", err)
			http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
			return
		}

		// Send new fallback confirmation email
		if err := auth.SendFallbackConfirmationEmail(req.FallbackEmail, newFallbackToken); err != nil {
			log.Printf("Failed to send updated fallback confirmation email: %v", err)
			http.Error(w, `{"error":"Failed to send confirmation email. Please try again."}`, http.StatusInternalServerError)
			return
		}

		// Return success response
		response := UpdateFallbackEmailResponse{
			Message: "Fallback email updated successfully. Please check your new fallback email for confirmation.",
			Status:  "fallback_updated",
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)

		log.Printf("Fallback email updated for pending signup %s (ID: %s)", req.Email, pendingSignup.ID)
	}
}

// updatePendingSignupFallbackEmail updates the fallback email and token for a pending signup
func updatePendingSignupFallbackEmail(db *sql.DB, pendingSignupID, newFallbackEmail, newFallbackToken string, newFallbackExpiration time.Time) error {
	query := `
		UPDATE pending_signups 
		SET fallback_email = ?, fallback_token = ?, fallback_token_expiration = ?, updated_at = ?
		WHERE id = ?
	`

	_, err := db.Exec(query, newFallbackEmail, newFallbackToken, newFallbackExpiration, time.Now(), pendingSignupID)
	return err
}

// isValidEmailFormatForUpdate checks if the email format is valid using regex.
func isValidEmailFormatForUpdate(email string) bool {
	// Basic email regex pattern
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(strings.TrimSpace(email))
}
