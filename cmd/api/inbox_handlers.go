package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// InboxEmailItem represents an email in the user's inbox
type InboxEmailItem struct {
	EmailID   string    `json:"email_id"`
	SenderID  string    `json:"sender_id"` // UUID reference only
	Subject   string    `json:"subject"`
	CreatedAt time.Time `json:"created_at"`
	Status    string    `json:"status"`
	IsRead    bool      `json:"is_read"`
}

// ListInboxResponse represents the response for listing inbox emails
type ListInboxResponse struct {
	Emails []InboxEmailItem `json:"emails"`
	Status string           `json:"status"`
	Error  string           `json:"error,omitempty"`
}

// GetInboxEmailResponse represents the response for getting a single inbox email
type GetInboxEmailResponse struct {
	Email  InboxEmailItem `json:"email"`
	Status string         `json:"status"`
	Error  string         `json:"error,omitempty"`
}

// DeleteInboxEmailResponse represents the response for deleting an inbox email
type DeleteInboxEmailResponse struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// listInboxHandler handles GET /api/inbox/list. It returns a list of emails sent TO the authenticated user.
func (srv *Server) listInboxHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Get authenticated user_id from context
	userID, ok := GetUserIDFromContext(r)
	if !ok {
		LogSecurityEvent(r, "unauthorized_access", "Authentication required for inbox access", LogLevelWarning, "", nil)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"Authentication required"}`))
		return
	}

	// Check if database is available
	if srv.db == nil {
		LogError("inbox_api", "Database connection unavailable for inbox operation", nil, nil)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Service temporarily unavailable"}`))
		return
	}

	// Query inbox_messages table for user's inbox with email details
	rows, err := srv.db.Query(`
		SELECT e.email_id, e.sender_id, e.subject, im.created_at,
		       CASE 
		           WHEN e.self_destructed = 1 THEN 'deleted'
		           WHEN e.expires_at IS NOT NULL AND e.expires_at < CURRENT_TIMESTAMP THEN 'expired'
		           WHEN im.is_read = 1 THEN 'read'
		           ELSE 'delivered'
		       END as status,
		       im.is_read
		FROM inbox_messages im
		INNER JOIN emails e ON im.email_id = e.email_id
		WHERE im.user_id = ? AND im.is_deleted = 0
		ORDER BY im.created_at DESC`,
		userID,
	)
	if err != nil {
		LogDatabaseOperation("SELECT", "inbox_messages", time.Since(startTime), err.Error())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to retrieve inbox"}`))
		return
	}
	defer rows.Close()

	// Build response
	var emails []InboxEmailItem
	for rows.Next() {
		var email InboxEmailItem
		err := rows.Scan(&email.EmailID, &email.SenderID, &email.Subject, &email.CreatedAt, &email.Status, &email.IsRead)
		if err != nil {
			LogWarning("inbox_api", "Failed to scan inbox email row", map[string]interface{}{"error": err.Error()})
			continue // Skip this row and continue with others
		}
		emails = append(emails, email)
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		LogDatabaseOperation("SCAN", "inbox_messages", time.Since(startTime), err.Error())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to process inbox"}`))
		return
	}

	// Return response
	response := ListInboxResponse{
		Emails: emails,
		Status: "success",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	// Log successful operation with Zero Visibility compliance
	LogInboxList(r, http.StatusOK, time.Since(startTime), userID, len(emails), "")
}

// getInboxEmailHandler handles GET /api/inbox/{id}. It returns a single email from the user's inbox.
func (srv *Server) getInboxEmailHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Get authenticated user_id from context
	userID, ok := GetUserIDFromContext(r)
	if !ok {
		LogSecurityEvent(r, "unauthorized_access", "Authentication required for inbox email access", LogLevelWarning, "", nil)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"Authentication required"}`))
		return
	}

	// Get email ID from URL parameters
	vars := mux.Vars(r)
	emailID := vars["id"]
	if emailID == "" {
		LogSecurityEvent(r, "invalid_request", "Missing email ID in inbox request", LogLevelWarning, userID, nil)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Missing email ID"}`))
		return
	}

	// Check if database is available
	if srv.db == nil {
		LogError("inbox_api", "Database connection unavailable for inbox email operation", nil, nil)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Service temporarily unavailable"}`))
		return
	}

	// Query the specific email from inbox_messages, ensuring it belongs to the authenticated user
	var email InboxEmailItem
	err := srv.db.QueryRow(`
		SELECT e.email_id, e.sender_id, e.subject, im.created_at,
		       CASE 
		           WHEN e.self_destructed = 1 THEN 'deleted'
		           WHEN e.expires_at IS NOT NULL AND e.expires_at < CURRENT_TIMESTAMP THEN 'expired'
		           WHEN im.is_read = 1 THEN 'read'
		           ELSE 'delivered'
		       END as status,
		       im.is_read
		FROM inbox_messages im
		INNER JOIN emails e ON im.email_id = e.email_id
		WHERE im.user_id = ? AND im.email_id = ? AND im.is_deleted = 0`,
		userID, emailID,
	).Scan(&email.EmailID, &email.SenderID, &email.Subject, &email.CreatedAt, &email.Status, &email.IsRead)

	if err != nil {
		LogSecurityEvent(r, "access_denied", "Email not found or access denied", LogLevelWarning, userID, map[string]interface{}{"email_id": emailID})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"Email not found"}`))
		return
	}

	// Return response
	response := GetInboxEmailResponse{
		Email:  email,
		Status: "success",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	// Log successful operation with Zero Visibility compliance
	LogInboxGet(r, http.StatusOK, time.Since(startTime), userID, emailID, "")
}

// deleteInboxEmailHandler handles DELETE /api/inbox/{id}. It soft deletes an email from the user's inbox.
func (srv *Server) deleteInboxEmailHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Get authenticated user_id from context
	userID, ok := GetUserIDFromContext(r)
	if !ok {
		LogSecurityEvent(r, "unauthorized_access", "Authentication required for inbox email deletion", LogLevelWarning, "", nil)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"Authentication required"}`))
		return
	}

	// Get email ID from URL parameters
	vars := mux.Vars(r)
	emailID := vars["id"]
	if emailID == "" {
		LogSecurityEvent(r, "invalid_request", "Missing email ID in inbox deletion request", LogLevelWarning, userID, nil)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Missing email ID"}`))
		return
	}

	// Check if database is available
	if srv.db == nil {
		LogError("inbox_api", "Database connection unavailable for inbox deletion operation", nil, nil)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Service temporarily unavailable"}`))
		return
	}

	// Soft delete the email from inbox_messages table
	result, err := srv.db.Exec(`
		UPDATE inbox_messages 
		SET is_deleted = 1, updated_at = CURRENT_TIMESTAMP
		WHERE user_id = ? AND email_id = ? AND is_deleted = 0`,
		userID, emailID,
	)
	if err != nil {
		LogDatabaseOperation("UPDATE", "inbox_messages", time.Since(startTime), err.Error())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to delete email"}`))
		return
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		LogDatabaseOperation("ROWS_AFFECTED", "inbox_messages", time.Since(startTime), err.Error())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to delete email"}`))
		return
	}

	if rowsAffected == 0 {
		LogSecurityEvent(r, "access_denied", "Email not found or access denied for deletion", LogLevelWarning, userID, map[string]interface{}{"email_id": emailID})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"Email not found"}`))
		return
	}

	// Return success response
	response := DeleteInboxEmailResponse{
		Status: "success",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	// Log successful operation with Zero Visibility compliance
	LogInboxDelete(r, http.StatusOK, time.Since(startTime), userID, emailID, "")
}

// createInboxMessageForEmail creates an inbox message entry when a new email is sent
// This function should be called whenever a new email is created
func (srv *Server) createInboxMessageForEmail(emailID, recipientEmail string) error {
	// Check if the recipient is a registered user
	var userID string
	err := srv.db.QueryRow("SELECT id FROM users WHERE email = ?", recipientEmail).Scan(&userID)
	if err != nil {
		// Recipient is not a registered user, no inbox message needed
		LogInfo("inbox_api", "Recipient not registered, skipping inbox message creation", map[string]interface{}{
			"email_id":  emailID,
			"recipient": "anonymous", // Zero Visibility: don't log actual email
		})
		return nil
	}

	// Create inbox message for the recipient
	_, err = srv.db.Exec(`
		INSERT OR IGNORE INTO inbox_messages (id, user_id, email_id, is_read, is_deleted, created_at, updated_at)
		VALUES (lower(hex(randomblob(16))), ?, ?, 0, 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		userID, emailID,
	)
	if err != nil {
		LogError("inbox_api", "Failed to create inbox message for email", err, map[string]interface{}{
			"email_id": emailID,
			"user_id":  userID,
		})
		return err
	}

	LogInfo("inbox_api", "Inbox message created successfully", map[string]interface{}{
		"email_id": emailID,
		"user_id":  userID,
	})

	return nil
}
