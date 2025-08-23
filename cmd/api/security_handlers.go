package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"secure-email-mvp/pkg/securelinks"
	"secure-email-mvp/pkg/securelinks/security"
)

// =============================================================================
// PHASE 2 SECURITY HANDLERS
// =============================================================================

// PasswordValidationHandler handles password validation for secure links
func (srv *Server) PasswordValidationHandler(w http.ResponseWriter, r *http.Request) {
	// Parse request
	var req security.PasswordValidationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get client IP
	req.IPAddress = getClientIPForSecureLinks(r)
	req.UserAgent = r.UserAgent()

	// Validate request
	if req.LinkID == "" {
		http.Error(w, "link_id is required", http.StatusBadRequest)
		return
	}
	if req.Password == "" {
		http.Error(w, "password is required", http.StatusBadRequest)
		return
	}

	// Create password protection service
	passwordService := security.NewPasswordProtectionService(srv.db)

	// Validate password
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	response, err := passwordService.ValidatePassword(ctx, req)
	if err != nil {
		log.Printf("Password validation error for link %s: %v", req.LinkID, err)
		http.Error(w, "Password validation failed", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CreatePasswordProtectedLinkHandler creates a secure link with password protection
func (srv *Server) CreatePasswordProtectedLinkHandler(w http.ResponseWriter, r *http.Request) {
	// Parse request
	var req struct {
		EmailID        string  `json:"email_id" validate:"required"`
		RecipientEmail string  `json:"recipient_email" validate:"required,email"`
		Password       string  `json:"password" validate:"required,min=6"`
		MaxAttempts    int     `json:"max_attempts,omitempty"`
		CustomMessage  *string `json:"custom_message,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get sender ID from JWT context
	senderID := r.Context().Value("user_id").(string)

	// Create password protection service
	passwordService := security.NewPasswordProtectionService(srv.db)

	// Hash password
	passwordHash, err := passwordService.HashPassword(req.Password)
	if err != nil {
		log.Printf("Failed to hash password: %v", err)
		http.Error(w, "Failed to process password", http.StatusInternalServerError)
		return
	}

	// Set default max attempts if not provided
	if req.MaxAttempts <= 0 {
		req.MaxAttempts = 3
	}

	// Create secure link with password protection
	secureLinkReq := securelinks.CreateSecureLinkRequest{
		EmailID:        req.EmailID,
		RecipientEmail: req.RecipientEmail,
		SecuritySettings: securelinks.SecuritySettings{
			RequirePassword:   true,
			PasswordHash:      &passwordHash,
			MaxAccessAttempts: req.MaxAttempts,
		},
		CustomMessage: req.CustomMessage,
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	response, err := srv.secureLinksService.CreateSecureLink(ctx, secureLinkReq, senderID)
	if err != nil {
		log.Printf("Failed to create password-protected link: %v", err)
		http.Error(w, "Failed to create secure link", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetPasswordAttemptsHandler returns password attempt history for a link (sender only)
func (srv *Server) GetPasswordAttemptsHandler(w http.ResponseWriter, r *http.Request) {
	// Get link ID from URL
	linkID := r.URL.Query().Get("link_id")
	if linkID == "" {
		http.Error(w, "link_id is required", http.StatusBadRequest)
		return
	}

	// Get sender ID from JWT context
	senderID := r.Context().Value("user_id").(string)

	// Verify sender owns the link
	_, err := srv.secureLinksService.GetSecureLinkInfo(r.Context(), linkID, senderID)
	if err != nil {
		http.Error(w, "Link not found or access denied", http.StatusNotFound)
		return
	}

	// Get password attempts
	query := `
		SELECT id, link_id, ip_address, user_agent, attempt_time, success, 
		       attempt_number, geolocation_data
		FROM link_password_attempts 
		WHERE link_id = ? 
		ORDER BY attempt_time DESC 
		LIMIT 100
	`

	rows, err := srv.db.QueryContext(r.Context(), query, linkID)
	if err != nil {
		log.Printf("Failed to query password attempts: %v", err)
		http.Error(w, "Failed to retrieve password attempts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var attempts []security.PasswordAttempt
	for rows.Next() {
		var attempt security.PasswordAttempt
		err := rows.Scan(
			&attempt.ID,
			&attempt.LinkID,
			&attempt.IPAddress,
			&attempt.UserAgent,
			&attempt.AttemptTime,
			&attempt.Success,
			&attempt.AttemptNumber,
			&attempt.GeolocationData,
		)
		if err != nil {
			log.Printf("Failed to scan password attempt: %v", err)
			continue
		}
		attempts = append(attempts, attempt)
	}

	// Return response
	response := struct {
		LinkID   string                     `json:"link_id"`
		Attempts []security.PasswordAttempt `json:"attempts"`
		Total    int                        `json:"total"`
	}{
		LinkID:   linkID,
		Attempts: attempts,
		Total:    len(attempts),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ClearPasswordAttemptsHandler clears password attempts for a link (sender only)
func (srv *Server) ClearPasswordAttemptsHandler(w http.ResponseWriter, r *http.Request) {
	// Parse request
	var req struct {
		LinkID string `json:"link_id" validate:"required"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get sender ID from JWT context
	senderID := r.Context().Value("user_id").(string)

	// Verify sender owns the link
	_, err := srv.secureLinksService.GetSecureLinkInfo(r.Context(), req.LinkID, senderID)
	if err != nil {
		http.Error(w, "Link not found or access denied", http.StatusNotFound)
		return
	}

	// Create password protection service
	passwordService := security.NewPasswordProtectionService(srv.db)

	// Clear failed attempts
	if err := passwordService.ClearFailedAttempts(r.Context(), req.LinkID); err != nil {
		log.Printf("Failed to clear password attempts: %v", err)
		http.Error(w, "Failed to clear password attempts", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Password attempts cleared successfully",
	})
}
