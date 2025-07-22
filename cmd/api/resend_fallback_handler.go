package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"secure-email-mvp/pkg/auth"
)

// ResendFallbackRequest is the JSON body for resend fallback
type ResendFallbackRequest struct {
	Email string `json:"email"`
}

// ResendFallbackResponse is the JSON response
type ResendFallbackResponse struct {
	Message string `json:"message"`
}

// resendFallbackHandlerFactory returns an HTTP handler for resending fallback confirmation emails. It generates a new token and sends a new link.
func resendFallbackHandlerFactory(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")

		var req ResendFallbackRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"Invalid JSON format"}`, http.StatusBadRequest)
			return
		}
		email := strings.TrimSpace(req.Email)
		if !isValidEmail(email) {
			http.Error(w, `{"error":"Invalid email format"}`, http.StatusBadRequest)
			return
		}

		// Query user
		var fallbackConfirmed bool
		var fallbackEmail string
		query := `SELECT fallback_confirmed, fallback_email FROM users WHERE email = ?`
		err := db.QueryRow(query, email).Scan(&fallbackConfirmed, &fallbackEmail)
		if err == sql.ErrNoRows {
			http.Error(w, `{"error":"User not found"}`, http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
			return
		}
		if fallbackConfirmed {
			http.Error(w, `{"error":"Fallback already confirmed"}`, http.StatusBadRequest)
			return
		}

		// Generate new token and expiration
		newToken := auth.GenerateFallbackToken(fallbackEmail)
		newExpiration := auth.GenerateFallbackExpiration()
		_, err = db.Exec(`UPDATE users SET fallback_token = ?, fallback_token_expiration = ? WHERE email = ?`, newToken, newExpiration, email)
		if err != nil {
			http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
			return
		}

		// Send confirmation email (stub)
		err = auth.SendFallbackConfirmationEmail(fallbackEmail, newToken)
		if err != nil {
			http.Error(w, `{"error":"Failed to send email"}`, http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(ResendFallbackResponse{Message: "Fallback confirmation email sent"})
	}
}
