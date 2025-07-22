package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"secure-email-mvp/pkg/auth"
	"time"
)

// FallbackConfirmResponse represents the JSON response for fallback confirmation
type FallbackConfirmResponse struct {
	Message string `json:"message"`
}

// confirmFallbackHandlerFactory returns an HTTP handler for confirming fallback email. It validates the token and updates user status.
func confirmFallbackHandlerFactory(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only allow GET method
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Set content type to JSON
		w.Header().Set("Content-Type", "application/json")

		// Get token from query parameter
		token := r.URL.Query().Get("token")
		if token == "" {
			response := map[string]string{"error": "Invalid or expired confirmation token"}
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		// Validate token format
		if !auth.ValidateFallbackToken(token) {
			response := map[string]string{"error": "Invalid or expired confirmation token"}
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		// Look up user by fallback token
		userID, fallbackExpiration, err := getUserByFallbackToken(db, token)
		if err != nil {
			if err == sql.ErrNoRows {
				response := map[string]string{"error": "Invalid or expired confirmation token"}
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(response)
				return
			}
			http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
			return
		}

		// Check if token has expired
		if auth.IsTokenExpired(fallbackExpiration) {
			response := map[string]string{"error": "Invalid or expired confirmation token"}
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		// Update fallback_confirmed to true
		err = confirmFallbackEmail(db, userID)
		if err != nil {
			http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
			return
		}

		// Return success response
		response := map[string]string{
			"message": "Fallback email confirmed successfully. You may now log in.",
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

// getUserByFallbackToken looks up a user by their fallback token for confirmation.
func getUserByFallbackToken(db *sql.DB, token string) (int, time.Time, error) {
	var userID int
	var fallbackExpiration time.Time
	query := `SELECT id, fallback_token_expiration FROM users WHERE fallback_token = ?`
	err := db.QueryRow(query, token).Scan(&userID, &fallbackExpiration)
	return userID, fallbackExpiration, err
}

// confirmFallbackEmail sets fallback_confirmed to true for the given user, enabling login.
func confirmFallbackEmail(db *sql.DB, userID int) error {
	query := `UPDATE users SET fallback_confirmed = true WHERE id = ?`
	_, err := db.Exec(query, userID)
	return err
}
