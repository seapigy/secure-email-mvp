package main

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"secure-email-mvp/pkg/auth"
)

// RefreshRequest represents the refresh token request structure
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// RefreshResponse represents the refresh token response
type RefreshResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

// refreshHandler handles POST /api/auth/refresh
func refreshHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse request
		var req RefreshRequest
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

		// Validate refresh token and get user ID
		userID, err := sessionManager.ValidateRefreshToken(req.RefreshToken, db)
		if err != nil {
			http.Error(w, `{"error":"Invalid or expired refresh token"}`, http.StatusUnauthorized)
			return
		}

		// Get user email for the new access token
		var email string
		err = db.QueryRow("SELECT email FROM users WHERE id = ?", userID).Scan(&email)
		if err != nil {
			http.Error(w, `{"error":"User not found"}`, http.StatusInternalServerError)
			return
		}

		// Generate new access token
		accessToken, err := sessionManager.GenerateAccessToken(userID, email)
		if err != nil {
			http.Error(w, `{"error":"Failed to generate access token"}`, http.StatusInternalServerError)
			return
		}

		// Return response
		response := RefreshResponse{
			AccessToken: accessToken,
			TokenType:   "Bearer",
			ExpiresIn:   int64(sessionManager.GetAccessTokenExpiry().Seconds()),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
