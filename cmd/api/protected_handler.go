package main

import (
	"encoding/json"
	"net/http"
)

// ProtectedResponse represents the JSON response for protected routes
type ProtectedResponse struct {
	Email   string `json:"email"`
	Message string `json:"message"`
}

// protectedTestHandler handles the /protected-test route
func protectedTestHandler(w http.ResponseWriter, r *http.Request) {
	// Get user email from context (set by JWT middleware)
	email, ok := GetUserEmailFromContext(r)
	if !ok {
		http.Error(w, `{"error":"User email not found in context"}`, http.StatusInternalServerError)
		return
	}

	// Set content type to JSON
	w.Header().Set("Content-Type", "application/json")

	// Return protected response
	response := ProtectedResponse{
		Email:   email,
		Message: "Access granted",
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
