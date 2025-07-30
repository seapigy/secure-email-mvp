package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type EmailListItem struct {
	EmailID   string    `json:"email_id"`
	Recipient string    `json:"recipient"`
	Subject   string    `json:"subject"`
	CreatedAt time.Time `json:"created_at"`
}

type ListEmailResponse struct {
	Emails []EmailListItem `json:"emails"`
	Status string          `json:"status"`
	Error  string          `json:"error,omitempty"`
}

// listEmailHandler handles GET /api/email/list. It returns a list of emails sent by the authenticated user.
func (srv *Server) listEmailHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("listEmailHandler started")

	// Get authenticated user_id from context
	userID, ok := GetUserIDFromContext(r)
	if !ok {
		log.Printf("User ID not found in context")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"Authentication required"}`))
		return
	}

	// Check if database is available
	if srv.db == nil {
		log.Printf("Database connection is nil")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Database connection unavailable"}`))
		return
	}

	// Query database for emails sent by this user
	rows, err := srv.db.Query(`
		SELECT email_id, recipient, subject, created_at
		FROM emails 
		WHERE sender_id = ?
		ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		log.Printf("Database query failed: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to retrieve emails"}`))
		return
	}
	defer rows.Close()

	// Build response
	var emails []EmailListItem
	for rows.Next() {
		var email EmailListItem
		err := rows.Scan(&email.EmailID, &email.Recipient, &email.Subject, &email.CreatedAt)
		if err != nil {
			log.Printf("Failed to scan email row: %v", err)
			continue // Skip this row and continue with others
		}
		emails = append(emails, email)
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		log.Printf("Error iterating over rows: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to process emails"}`))
		return
	}

	// Return response
	response := ListEmailResponse{
		Emails: emails,
		Status: "success",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
} 