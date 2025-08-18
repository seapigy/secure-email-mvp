// =============================================================================
// SECURE EMAIL MVP - RECIPIENT EMAIL HANDLER
// =============================================================================
// HTTP handler for recipient-based email access with forwarding prevention.
// Micro-Iteration 4.18: Secure Email Forwarding Prevention
// =============================================================================

package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"secure-email-mvp/pkg/auth"
	"secure-email-mvp/pkg/notification"
	"secure-email-mvp/pkg/readreceipts"
	"secure-email-mvp/pkg/storage"

	"github.com/gorilla/mux"
)

// RecipientEmailResponse represents the response for recipient email access
type RecipientEmailResponse struct {
	EmailID   string    `json:"email_id"`
	SenderID  string    `json:"sender_id"`
	Recipient string    `json:"recipient"`
	Subject   string    `json:"subject"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	Status    string    `json:"status"`
	Error     string    `json:"error,omitempty"`
}

// getRecipientEmailHandler handles GET /api/email/{id}/content for recipient access.
// This endpoint enforces recipient-based access control to prevent email forwarding.
func (srv *Server) getRecipientEmailHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("getRecipientEmailHandler started")

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

	// 1. Retrieve email metadata from database including recipient_id
	var (
		blobID, encryptedKeyB64, nonceB64, authTagB64, compressionAlgo string
		senderID, recipient, subject                                   string
		recipientID                                                    *string
		createdAt                                                      time.Time
		failCount                                                      int
	)

	err := srv.db.QueryRow(`
		SELECT encrypted_blob_url, encrypted_key, encryption_nonce, encryption_auth_tag, 
		       compression_algo, sender_id, recipient, recipient_id, subject, created_at, failed_attempts
		FROM emails WHERE email_id = ?`,
		emailID,
	).Scan(&blobID, &encryptedKeyB64, &nonceB64, &authTagB64, &compressionAlgo,
		&senderID, &recipient, &recipientID, &subject, &createdAt, &failCount)

	if err != nil {
		log.Printf("Database query failed: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"Email not found"}`))
		return
	}

	// 2. Check recipient-based access control
	// Only the intended recipient (recipient_id) can access the email
	if recipientID == nil {
		log.Printf("Email %s has no recipient_id set - access denied", emailID)
		if err := srv.handleFailedAccess(emailID, blobID, failCount, "no recipient_id set"); err != nil {
			if _, ok := err.(EmailDeletedError); ok {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusGone)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "Email deleted due to too many failed attempts",
				})
				return
			}
			log.Printf("Failed to handle failed access: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error":"Access denied"}`))
		return
	}

	if *recipientID != userID {
		log.Printf("Unauthorized access attempt: user %s trying to access email intended for recipient %s", userID, *recipientID)

		// Increment fail count and check if limit reached
		if err := srv.handleFailedAccess(emailID, blobID, failCount, "unauthorized recipient access"); err != nil {
			if _, ok := err.(EmailDeletedError); ok {
				// Email was deleted due to too many failed attempts
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusGone)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "Email deleted due to too many failed attempts",
				})
				return
			}
			log.Printf("Failed to handle failed access: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error":"Access denied"}`))
		return
	}

	// 3. Decode base64 fields
	encryptedKey, err := base64.StdEncoding.DecodeString(encryptedKeyB64)
	if err != nil {
		log.Printf("Failed to decode encrypted key: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Invalid encryption key"}`))
		return
	}

	nonce, err := base64.StdEncoding.DecodeString(nonceB64)
	if err != nil {
		log.Printf("Failed to decode nonce: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Invalid encryption nonce"}`))
		return
	}

	authTag, err := base64.StdEncoding.DecodeString(authTagB64)
	if err != nil {
		log.Printf("Failed to decode auth tag: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Invalid encryption auth tag"}`))
		return
	}

	// 4. Retrieve encrypted blob from R2
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	encryptedBlob, err := storage.GetEmailFromR2(ctx, blobID)
	if err != nil {
		log.Printf("Failed to retrieve blob from R2: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to retrieve encrypted content"}`))
		return
	}

	// 5. Separate ciphertext and auth tag
	if len(encryptedBlob) < 16 {
		log.Printf("Encrypted blob too short: %d bytes", len(encryptedBlob))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Invalid encrypted content"}`))
		return
	}

	ciphertext := encryptedBlob[:len(encryptedBlob)-16]
	blobAuthTag := encryptedBlob[len(encryptedBlob)-16:]

	// 6. Verify auth tag matches
	if !bytes.Equal(authTag, blobAuthTag) {
		log.Printf("Auth tag mismatch")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Content integrity check failed"}`))
		return
	}

	// 7. Create EncryptedData struct for decryption
	encryptedData := &auth.EncryptedData{
		Ciphertext: ciphertext,
		Key:        encryptedKey,
		Nonce:      nonce,
		AuthTag:    authTag,
	}

	// 8. Decrypt the content
	compressed, err := auth.DecryptAES256GCM(encryptedData)
	if err != nil {
		log.Printf("Decryption failed: %v", err)

		// Increment fail count and check if limit reached
		if err := srv.handleFailedAccess(emailID, blobID, failCount, "decryption failed"); err != nil {
			if _, ok := err.(EmailDeletedError); ok {
				// Email was deleted due to too many failed attempts
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusGone)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "Email deleted due to too many failed attempts",
				})
				return
			}
			log.Printf("Failed to handle failed access: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Decryption failed"}`))
		return
	}

	// 9. Decompress the content
	var plaintext []byte
	if compressionAlgo == "gzip" {
		reader := bytes.NewReader(compressed)
		gzReader, err := gzip.NewReader(reader)
		if err != nil {
			log.Printf("Failed to create gzip reader: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"Decompression failed"}`))
			return
		}
		defer gzReader.Close()

		plaintext, err = io.ReadAll(gzReader)
		if err != nil {
			log.Printf("Failed to read decompressed content: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"Decompression failed"}`))
			return
		}
	} else {
		// No compression or unknown algorithm
		plaintext = compressed
	}

	// 10. Record read event and handle read receipts
	readEvent := &readreceipts.ReadEvent{
		EmailID:   emailID,
		UserID:    userID,
		ReadAt:    time.Now(),
		IPAddress: getClientIP(r),
		UserAgent: r.UserAgent(),
		// TODO: Add geolocation data if available
	}

	if err := srv.readReceiptService.RecordReadEvent(r.Context(), readEvent); err != nil {
		log.Printf("Failed to record read event: %v", err)
		// Don't fail the request for tracking errors
	}

	// Send read receipt if this is the first read
	if readEvent.IsFirstRead {
		if err := srv.readReceiptService.SendReadReceipt(r.Context(), emailID, senderID, userID, readEvent); err != nil {
			log.Printf("Failed to send read receipt: %v", err)
			// Don't fail the request for read receipt errors
		}
	}

	// 11. Update access tracking and reset fail count on successful access
	_, err = srv.db.Exec(`
		UPDATE emails SET 
			last_access_at = CURRENT_TIMESTAMP,
			access_count = access_count + 1,
			fail_count = 0
		WHERE email_id = ?`,
		emailID,
	)
	if err != nil {
		log.Printf("Failed to update access tracking: %v", err)
		// Don't fail the request for tracking errors
	}

	// 12. Record successful access event
	clientIP := getClientIP(r)
	if err := srv.recordAccessEvent(r.Context(), emailID, userID, clientIP, r.UserAgent(), "", "", "", "", notification.AccessEventTypeSuccess); err != nil {
		log.Printf("Failed to record successful access event: %v", err)
	}

	// 12. Return the decrypted email content
	response := RecipientEmailResponse{
		EmailID:   emailID,
		SenderID:  senderID,
		Recipient: recipient,
		Subject:   subject,
		Body:      string(plaintext),
		CreatedAt: createdAt,
		Status:    "success",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	log.Printf("Recipient email access successful for email %s by user %s", emailID, userID)
}
