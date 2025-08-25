// =============================================================================
// SECURE EMAIL MVP - READ RECEIPT HANDLERS
// =============================================================================
// HTTP handlers for read receipts and expiration alerts.
// Micro-Iteration 4.19: Email Read Receipt & Expiration Alerts
// =============================================================================

package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"secure-email-mvp/pkg/readreceipts"

	"github.com/gorilla/mux"
)

// ReadReceiptPreferencesRequest represents a request to update read receipt preferences
type ReadReceiptPreferencesRequest struct {
	EnableReadReceipts     bool   `json:"enable_read_receipts"`
	EnableExpirationAlerts bool   `json:"enable_expiration_alerts"`
	ExpirationAlertHours   int    `json:"expiration_alert_hours"`
	DeliveryMethods        string `json:"delivery_methods"`
}

// EmailReadReceiptSettingsRequest represents a request to update email-specific settings
type EmailReadReceiptSettingsRequest struct {
	EnableReadReceipts     bool `json:"enable_read_receipts"`
	EnableExpirationAlerts bool `json:"enable_expiration_alerts"`
	ExpirationAlertHours   int  `json:"expiration_alert_hours"`
}

// getReadReceiptPreferencesHandler handles GET /api/read-receipts/preferences
func (srv *Server) getReadReceiptPreferencesHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("getReadReceiptPreferencesHandler started")

	// Get authenticated user_id from context
	userID, ok := GetUserIDFromContext(r)
	if !ok {
		log.Printf("User ID not found in context")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"Authentication required"}`))
		return
	}

	// Get user preferences
	prefs, err := srv.readReceiptService.GetReadReceiptPreferences(r.Context(), userID)
	if err != nil {
		log.Printf("Failed to get read receipt preferences: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to retrieve preferences"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prefs)

	log.Printf("Retrieved read receipt preferences for user %s", userID)
}

// updateReadReceiptPreferencesHandler handles PUT /api/read-receipts/preferences
func (srv *Server) updateReadReceiptPreferencesHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("updateReadReceiptPreferencesHandler started")

	// Get authenticated user_id from context
	userID, ok := GetUserIDFromContext(r)
	if !ok {
		log.Printf("User ID not found in context")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"Authentication required"}`))
		return
	}

	// Parse request body
	var req ReadReceiptPreferencesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Failed to decode request body: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Invalid request body"}`))
		return
	}

	// Validate request
	if req.ExpirationAlertHours < 1 || req.ExpirationAlertHours > 168 { // Max 1 week
		log.Printf("Invalid expiration alert hours: %d", req.ExpirationAlertHours)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Expiration alert hours must be between 1 and 168"}`))
		return
	}

	// Get current preferences to preserve created_at
	currentPrefs, err := srv.readReceiptService.GetReadReceiptPreferences(r.Context(), userID)
	if err != nil {
		log.Printf("Failed to get current preferences: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to retrieve current preferences"}`))
		return
	}

	// Update preferences
	updatedPrefs := &readreceipts.ReadReceiptPreferences{
		UserID:                 userID,
		EnableReadReceipts:     req.EnableReadReceipts,
		EnableExpirationAlerts: req.EnableExpirationAlerts,
		ExpirationAlertHours:   req.ExpirationAlertHours,
		DeliveryMethods:        req.DeliveryMethods,
		CreatedAt:              currentPrefs.CreatedAt,
		UpdatedAt:              time.Now(),
	}

	if err := srv.readReceiptService.UpdateReadReceiptPreferences(r.Context(), updatedPrefs); err != nil {
		log.Printf("Failed to update read receipt preferences: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to update preferences"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedPrefs)

	log.Printf("Updated read receipt preferences for user %s", userID)
}

// getEmailReadReceiptInfoHandler handles GET /api/emails/{id}/read-receipts
func (srv *Server) getEmailReadReceiptInfoHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("getEmailReadReceiptInfoHandler started")

	// Get authenticated user_id from context
	userID, ok := GetUserIDFromContext(r)
	if !ok {
		log.Printf("User ID not found in context")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"Authentication required"}`))
		return
	}

	// Extract email_id from URL path
	vars := mux.Vars(r)
	emailID := vars["id"]
	if emailID == "" {
		log.Printf("Missing email_id in URL path")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Missing email_id"}`))
		return
	}

	// Verify user owns this email
	var senderID string
	err := srv.db.QueryRow("SELECT sender_id FROM emails WHERE email_id = ?", emailID).Scan(&senderID)
	if err != nil {
		log.Printf("Failed to get email sender: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"Email not found"}`))
		return
	}

	if senderID != userID {
		log.Printf("Unauthorized access attempt: user %s trying to access email %s owned by %s", userID, emailID, senderID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error":"Access denied"}`))
		return
	}

	// Get read receipt info
	info, err := srv.readReceiptService.GetEmailReadReceiptInfo(r.Context(), emailID)
	if err != nil {
		log.Printf("Failed to get read receipt info: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to retrieve read receipt info"}`))
		return
	}

	// Get expiration info
	expirationInfo, err := srv.readReceiptService.GetEmailExpirationInfo(r.Context(), emailID)
	if err != nil {
		log.Printf("Failed to get expiration info: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to retrieve expiration info"}`))
		return
	}

	// Combine info into response
	response := map[string]interface{}{
		"read_receipts": info,
		"expiration":    expirationInfo,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	log.Printf("Retrieved read receipt info for email %s by user %s", emailID, userID)
}

// getEmailReadEventsHandler handles GET /api/emails/{id}/read-events
func (srv *Server) getEmailReadEventsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("getEmailReadEventsHandler started")

	// Get authenticated user_id from context
	userID, ok := GetUserIDFromContext(r)
	if !ok {
		log.Printf("User ID not found in context")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"Authentication required"}`))
		return
	}

	// Extract email_id from URL path
	vars := mux.Vars(r)
	emailID := vars["id"]
	if emailID == "" {
		log.Printf("Missing email_id in URL path")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Missing email_id"}`))
		return
	}

	// Verify user owns this email
	var senderID string
	err := srv.db.QueryRow("SELECT sender_id FROM emails WHERE email_id = ?", emailID).Scan(&senderID)
	if err != nil {
		log.Printf("Failed to get email sender: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"Email not found"}`))
		return
	}

	if senderID != userID {
		log.Printf("Unauthorized access attempt: user %s trying to access email %s owned by %s", userID, emailID, senderID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error":"Access denied"}`))
		return
	}

	// Get limit from query parameter
	limitStr := r.URL.Query().Get("limit")
	limit := 50 // Default limit
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	// Get read events
	events, err := srv.readReceiptService.GetReadEvents(r.Context(), emailID, limit)
	if err != nil {
		log.Printf("Failed to get read events: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to retrieve read events"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)

	log.Printf("Retrieved %d read events for email %s by user %s", len(events), emailID, userID)
}

// updateEmailReadReceiptSettingsHandler handles PUT /api/emails/{id}/read-receipt-settings
func (srv *Server) updateEmailReadReceiptSettingsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("updateEmailReadReceiptSettingsHandler started")

	// Get authenticated user_id from context
	userID, ok := GetUserIDFromContext(r)
	if !ok {
		log.Printf("User ID not found in context")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"Authentication required"}`))
		return
	}

	// Extract email_id from URL path
	vars := mux.Vars(r)
	emailID := vars["id"]
	if emailID == "" {
		log.Printf("Missing email_id in URL path")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Missing email_id"}`))
		return
	}

	// Verify user owns this email
	var senderID string
	err := srv.db.QueryRow("SELECT sender_id FROM emails WHERE email_id = ?", emailID).Scan(&senderID)
	if err != nil {
		log.Printf("Failed to get email sender: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"Email not found"}`))
		return
	}

	if senderID != userID {
		log.Printf("Unauthorized access attempt: user %s trying to access email %s owned by %s", userID, emailID, senderID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error":"Access denied"}`))
		return
	}

	// Parse request body
	var req EmailReadReceiptSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Failed to decode request body: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Invalid request body"}`))
		return
	}

	// Validate request
	if req.ExpirationAlertHours < 1 || req.ExpirationAlertHours > 168 { // Max 1 week
		log.Printf("Invalid expiration alert hours: %d", req.ExpirationAlertHours)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Expiration alert hours must be between 1 and 168"}`))
		return
	}

	// Update email settings
	_, err = srv.db.Exec(`
		UPDATE emails SET 
			enable_read_receipts = ?,
			enable_expiration_alerts = ?,
			expiration_alert_hours = ?
		WHERE email_id = ?`,
		req.EnableReadReceipts, req.EnableExpirationAlerts, req.ExpirationAlertHours, emailID,
	)
	if err != nil {
		log.Printf("Failed to update email read receipt settings: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to update settings"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"Settings updated successfully"}`))

	log.Printf("Updated read receipt settings for email %s by user %s", emailID, userID)
}












