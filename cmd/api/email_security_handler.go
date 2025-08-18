package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"secure-email-mvp/pkg/email"
)

// updateEmailSecurityHandler handles POST /api/email/security/{id}
// Allows the sender to update security settings for their email
func (srv *Server) updateEmailSecurityHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("updateEmailSecurityHandler started")

	// Get authenticated user information
	userID, _, ok := GetUserFromContext(r)
	if !ok {
		log.Printf("User information not found in context")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(email.EmailSecurityResponse{
			Status: "error",
			Error:  "Authentication required",
		})
		return
	}

	// Extract email ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		log.Printf("Invalid URL path: %s", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(email.EmailSecurityResponse{
			Status: "error",
			Error:  "Invalid email ID",
		})
		return
	}
	emailID := pathParts[4]

	log.Printf("Updating security settings for email %s by user %s", emailID, userID)

	// Check if database is available
	if srv.db == nil {
		log.Printf("Database connection is nil")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(email.EmailSecurityResponse{
			Status: "error",
			Error:  "Database connection unavailable",
		})
		return
	}

	// Parse request body
	var request email.EmailSecurityRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Printf("Failed to decode request body: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(email.EmailSecurityResponse{
			Status: "error",
			Error:  "Invalid request format",
		})
		return
	}

	// Validate request
	if request.EmailID == "" {
		log.Printf("Email ID is required")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(email.EmailSecurityResponse{
			Status: "error",
			Error:  "Email ID is required",
		})
		return
	}

	// Ensure email ID in path matches request body
	if request.EmailID != emailID {
		log.Printf("Email ID mismatch: path=%s, body=%s", emailID, request.EmailID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(email.EmailSecurityResponse{
			Status: "error",
			Error:  "Email ID mismatch",
		})
		return
	}

	// Validate security toggles
	if err := email.ValidateEmailSecurityToggles(request.Toggles); err != nil {
		log.Printf("Invalid security toggles: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(email.EmailSecurityResponse{
			Status: "error",
			Error:  err.Error(),
		})
		return
	}

	// Create email security DB instance
	emailSecurityDB := srv.createEmailSecurityDB()

	// Update security settings
	if err := emailSecurityDB.UpdateEmailSecurityToggles(emailID, userID, request.Toggles); err != nil {
		log.Printf("Failed to update security settings: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(email.EmailSecurityResponse{
			Status: "error",
			Error:  "Failed to update security settings",
		})
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(email.EmailSecurityResponse{
		Status: "ok",
	})

	log.Printf("Successfully updated security settings for email %s", emailID)
}

// getEmailSecurityHandler handles GET /api/email/security/{id}
// Returns the current security settings for an email (sender only)
func (srv *Server) getEmailSecurityHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("getEmailSecurityHandler started")

	// Get authenticated user information
	userID, _, ok := GetUserFromContext(r)
	if !ok {
		log.Printf("User information not found in context")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(email.EmailSecurityResponse{
			Status: "error",
			Error:  "Authentication required",
		})
		return
	}

	// Extract email ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		log.Printf("Invalid URL path: %s", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(email.EmailSecurityResponse{
			Status: "error",
			Error:  "Invalid email ID",
		})
		return
	}
	emailID := pathParts[4]

	log.Printf("Retrieving security settings for email %s by user %s", emailID, userID)

	// Check if database is available
	if srv.db == nil {
		log.Printf("Database connection is nil")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(email.EmailSecurityResponse{
			Status: "error",
			Error:  "Database connection unavailable",
		})
		return
	}

	// Create email security DB instance
	emailSecurityDB := srv.createEmailSecurityDB()

	// Get security settings
	toggles, err := emailSecurityDB.GetEmailSecurityToggles(emailID, userID)
	if err != nil {
		log.Printf("Failed to retrieve security settings: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(email.EmailSecurityResponse{
			Status: "error",
			Error:  "Failed to retrieve security settings",
		})
		return
	}

	// Return security settings
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(email.EmailSecurityInfo{
		EmailID: emailID,
		Toggles: *toggles,
	})

	log.Printf("Successfully retrieved security settings for email %s", emailID)
}
