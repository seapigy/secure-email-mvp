package main

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"secure-email-mvp/pkg/auth"
)

// LoginRequest represents the login request structure
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	TOTPCode string `json:"totp_code"`
}

// LoginResponse represents the login response with token pair
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	UserID       string `json:"user_id"`
	Email        string `json:"email"`
}

// loginHandler handles POST /api/auth/login with enhanced session management
func loginHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse request
		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"Invalid request format"}`, http.StatusBadRequest)
			return
		}

		// Validate required fields
		if req.Email == "" || req.Password == "" || req.TOTPCode == "" {
			http.Error(w, `{"error":"Email, password, and TOTP code are required"}`, http.StatusBadRequest)
			return
		}

		// Authenticate user (existing logic)
		userID, email, err := authenticateUser(db, req.Email, req.Password, req.TOTPCode)
		if err != nil {
			http.Error(w, `{"error":"Invalid credentials"}`, http.StatusUnauthorized)
			return
		}

		// Create session manager
		sessionManager, err := auth.NewSessionManager()
		if err != nil {
			http.Error(w, `{"error":"Session configuration error"}`, http.StatusInternalServerError)
			return
		}

		// Generate token pair
		tokenPair, err := sessionManager.GenerateTokenPair(userID, email, db)
		if err != nil {
			http.Error(w, `{"error":"Failed to generate tokens"}`, http.StatusInternalServerError)
			return
		}

		// Return response
		response := LoginResponse{
			AccessToken:  tokenPair.AccessToken,
			RefreshToken: tokenPair.RefreshToken,
			TokenType:    tokenPair.TokenType,
			ExpiresIn:    tokenPair.ExpiresIn,
			UserID:       userID,
			Email:        email,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// authenticateUser handles the authentication logic using existing auth package
func authenticateUser(db *sql.DB, email, password, totpCode string) (string, string, error) {
	// Use the existing Authenticate function from auth package
	_, userID, err := auth.Authenticate(db, email, password, totpCode)
	if err != nil {
		return "", "", err
	}

	// For now, we'll use the existing authentication but return userID and email
	// The token will be replaced by our new session management
	return userID, email, nil
}
