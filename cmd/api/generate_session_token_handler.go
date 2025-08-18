package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"secure-email-mvp/pkg/audit"
	"secure-email-mvp/pkg/auth"

	"github.com/gorilla/mux"
)

// generateSessionTokenHandler handles POST /api/email/{id}/session
// This endpoint generates a session token after all security checks pass
func (srv *Server) generateSessionTokenHandler(w http.ResponseWriter, r *http.Request) {
	// Extract email ID from URL
	vars := mux.Vars(r)
	emailID := vars["id"]
	if emailID == "" {
		http.Error(w, `{"error":"Email ID is required"}`, http.StatusBadRequest)
		return
	}

	// Extract and validate JWT token
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		http.Error(w, `{"error":"Invalid authentication format"}`, http.StatusUnauthorized)
		return
	}

	// Validate JWT token and extract user ID
	userID, _, err := auth.ValidateJWT(tokenString)
	if err != nil {
		log.Printf("❌ JWT validation failed: %v", err)
		http.Error(w, `{"error":"Invalid authentication"}`, http.StatusUnauthorized)
		return
	}

	// Get client information
	clientIP := getClientIP(r)
	userAgent := r.Header.Get("User-Agent")

	// Verify email exists and user has access
	var exists int
	err = srv.db.QueryRow("SELECT 1 FROM emails WHERE email_id = ? AND user_id = ?", emailID, userID).Scan(&exists)
	if err != nil {
		log.Printf("❌ Email access denied for user %s, email %s: %v", userID, emailID, err)
		if srv.auditService != nil {
			event := &audit.AuditEvent{
				EventType:      audit.EventTypeEmailAccess,
				UserID:         &userID,
				IPAddress:      &clientIP,
				UserAgent:      &userAgent,
				RelatedEmailID: &emailID,
				Outcome:        audit.OutcomeFailure,
				Severity:       audit.SeverityWarning,
				Details: map[string]interface{}{
					"reason": "Email not found or access denied",
				},
			}
			srv.auditService.RecordEvent(r.Context(), event)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Email not found"})
		return
	}

	// Check if email requires session tokens
	var oneTimeLinkOnly bool
	err = srv.db.QueryRow("SELECT one_time_link_only FROM emails WHERE email_id = ?", emailID).Scan(&oneTimeLinkOnly)
	if err != nil {
		log.Printf("❌ Failed to check one_time_link_only setting for email %s: %v", emailID, err)
		// Continue without one-time link restriction if we can't determine the setting
	}

	// Generate session token
	sessionToken, err := srv.sessionTokenService.GenerateSessionToken(emailID, userAgent, clientIP)
	if err != nil {
		log.Printf("❌ Failed to generate session token for email %s: %v", emailID, err)
		if srv.auditService != nil {
			event := &audit.AuditEvent{
				EventType:      audit.EventTypeEmailAccess,
				UserID:         &userID,
				IPAddress:      &clientIP,
				UserAgent:      &userAgent,
				RelatedEmailID: &emailID,
				Outcome:        audit.OutcomeFailure,
				Severity:       audit.SeverityError,
				Details: map[string]interface{}{
					"reason": "Session token generation failed",
				},
			}
			srv.auditService.RecordEvent(r.Context(), event)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to generate session token"})
		return
	}

	// Log successful session token generation
	if srv.auditService != nil {
		event := &audit.AuditEvent{
			EventType:      audit.EventTypeSystemEvent,
			UserID:         &userID,
			IPAddress:      &clientIP,
			UserAgent:      &userAgent,
			RelatedEmailID: &emailID,
			Outcome:        audit.OutcomeSuccess,
			Severity:       audit.SeverityInfo,
			Details: map[string]interface{}{
				"reason": "Session token generated",
			},
		}
		srv.auditService.RecordEvent(r.Context(), event)
	}

	// Return session token
	response := map[string]interface{}{
		"session_token": sessionToken,
		"expires_in":    300, // 5 minutes in seconds
		"one_time_only": oneTimeLinkOnly,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
