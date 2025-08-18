package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// EmailSummary represents a single email in the list response
type EmailSummary struct {
	ID              string    `json:"id"`              // UUID string
	SenderID        string    `json:"sender_id"`       // Sender's user ID
	SenderEmail     string    `json:"sender_email"`    // Sender's email address
	Recipient       string    `json:"recipient"`       // Recipient email address
	Subject         string    `json:"subject"`         // Email subject
	CreatedAt       time.Time `json:"created_at"`      // ISO8601 UTC timestamp
	IsRead          bool      `json:"is_read"`         // Whether email has been read
	HasAttachments  bool      `json:"has_attachments"` // Whether email has attachments
}

// EmailsResponse represents the response structure for GET /api/emails
type EmailsResponse struct {
	Emails []EmailSummary `json:"emails"`
	Status string         `json:"status"`
	Error  string         `json:"error,omitempty"`
}

// emailsHandler handles GET /api/emails. It returns a list of all emails for the authenticated user
// (as sender or recipient), sorted by created_at descending. This is a metadata-only endpoint.
func (srv *Server) emailsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("emailsHandler started")

	// Get authenticated user_id and email from context
	userID, userEmail, ok := GetUserFromContext(r)
	if !ok {
		log.Printf("User information not found in context")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Authentication required"})
		return
	}

	log.Printf("Processing emails for user: %s (%s)", userID, userEmail)

	// Check if database is available
	if srv.db == nil {
		log.Printf("Database connection is nil")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Database connection unavailable"})
		return
	}

	// Query database for emails where user is either sender or recipient
	// Use case-insensitive match for recipient email
	rows, err := srv.db.Query(`
		SELECT 
			e.email_id,
			e.sender_id,
			u.email as sender_email,
			e.recipient,
			e.subject,
			e.created_at,
			COALESCE(e.access_count > 0, FALSE) as is_read,
			COALESCE(e.has_attachments, FALSE) as has_attachments
		FROM emails e
		LEFT JOIN users u ON e.sender_id = u.id
		WHERE e.sender_id = ? OR LOWER(e.recipient) = LOWER(?)
		ORDER BY e.created_at DESC`,
		userID, userEmail,
	)
	if err != nil {
		log.Printf("Database query failed: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to retrieve emails"})
		return
	}
	defer rows.Close()

	// Build response
	var emails []EmailSummary
	for rows.Next() {
		var email EmailSummary
		var isReadInt, hasAttachmentsInt int
		
		err := rows.Scan(
			&email.ID,
			&email.SenderID,
			&email.SenderEmail,
			&email.Recipient,
			&email.Subject,
			&email.CreatedAt,
			&isReadInt,
			&hasAttachmentsInt,
		)
		if err != nil {
			log.Printf("Failed to scan email row: %v", err)
			continue // Skip this row and continue with others
		}
		
		// Convert integer flags to boolean
		email.IsRead = isReadInt == 1
		email.HasAttachments = hasAttachmentsInt == 1
		
		emails = append(emails, email)
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		log.Printf("Error iterating over rows: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to process emails"})
		return
	}

	// Return response
	response := EmailsResponse{
		Emails: emails,
		Status: "success",
	}

	log.Printf("Successfully retrieved %d emails for user %s", len(emails), userID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
