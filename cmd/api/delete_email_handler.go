package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"secure-email-mvp/pkg/storage"

	"github.com/gorilla/mux"
)

type DeleteEmailResponse struct {
	Status  string `json:"status"`
	EmailID string `json:"email_id"`
	Error   string `json:"error,omitempty"`
}

// deleteEmailHandler handles DELETE /api/email/{id}. It securely deletes an email and its encrypted content.
func (srv *Server) deleteEmailHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("deleteEmailHandler started")

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

	// Check if database is available
	if srv.db == nil {
		log.Printf("Database connection is nil")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Database connection unavailable"}`))
		return
	}

	// 1. Retrieve email metadata from database to verify ownership
	var (
		blobID, senderID string
		createdAt        time.Time
	)

	err := srv.db.QueryRow(`
		SELECT encrypted_blob_url, sender_id, created_at
		FROM emails WHERE email_id = ?`,
		emailID,
	).Scan(&blobID, &senderID, &createdAt)

	if err != nil {
		log.Printf("Database query failed: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"Email not found"}`))
		return
	}

	// 2. Check if the authenticated user is the sender of this email
	if senderID != userID {
		log.Printf("Unauthorized deletion attempt: user %s trying to delete email from sender %s", userID, senderID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error":"Access denied"}`))
		return
	}

	// 3. Delete encrypted blob from R2
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = storage.DeleteBlob(ctx, blobID)
	if err != nil {
		log.Printf("Failed to delete blob from R2: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to delete encrypted content"}`))
		return
	}

	// 4. Remove email record from SQLite
	_, err = srv.db.Exec(`
		DELETE FROM emails WHERE email_id = ?`,
		emailID,
	)
	if err != nil {
		log.Printf("Failed to delete email record from database: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to delete email record"}`))
		return
	}

	// 5. Return success response
	response := DeleteEmailResponse{
		Status:  "deleted",
		EmailID: emailID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	log.Printf("Successfully deleted email %s for user %s", emailID, userID)
} 