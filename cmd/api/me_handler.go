package main

import (
	"encoding/json"
	"net/http"
)

// MeResponse represents the response from the /api/auth/me endpoint
type MeResponse struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}

// meHandler handles GET /api/auth/me
// This endpoint requires authentication via EnhancedJWTMiddleware
func meHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get user information from context (set by EnhancedJWTMiddleware)
		userID, email, ok := GetUserFromContext(r)
		if !ok {
			http.Error(w, `{"error":"User information not found in context"}`, http.StatusInternalServerError)
			return
		}

		// Return user information
		response := MeResponse{
			UserID: userID,
			Email:  email,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
