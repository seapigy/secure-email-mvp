package main

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"secure-email-mvp/pkg/auth"
)

// LogoutRequest represents the logout request structure
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// LogoutResponse represents the logout response
type LogoutResponse struct {
	Message string `json:"message"`
}

// logoutHandler handles POST /api/auth/logout
func logoutHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse request
		var req LogoutRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"Invalid request format"}`, http.StatusBadRequest)
			return
		}

		// Validate required fields
		if req.RefreshToken == "" {
			http.Error(w, `{"error":"Refresh token is required"}`, http.StatusBadRequest)
			return
		}

		// Create session manager
		sessionManager, err := auth.NewSessionManager()
		if err != nil {
			http.Error(w, `{"error":"Session configuration error"}`, http.StatusInternalServerError)
			return
		}

		// Revoke the refresh token
		err = sessionManager.RevokeRefreshToken(req.RefreshToken, db)
		if err != nil {
			// Don't return an error if the token is invalid - just log it
			// This prevents information leakage about token validity
			http.Error(w, `{"error":"Invalid refresh token"}`, http.StatusBadRequest)
			return
		}

		// Return success response
		response := LogoutResponse{
			Message: "Successfully logged out",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
