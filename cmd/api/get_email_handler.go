package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"secure-email-mvp/pkg/auth"
	"secure-email-mvp/pkg/email"
	"secure-email-mvp/pkg/storage"
)

// FAIL_ATTEMPT_LIMIT is the maximum number of failed attempts before email deletion
const FAIL_ATTEMPT_LIMIT = 3

type GetEmailRequest struct {
	EmailID string `json:"email_id"`
}

type GetEmailResponse struct {
	EmailID   string    `json:"email_id"`
	SenderID  string    `json:"sender_id"`
	Recipient string    `json:"recipient"`
	Subject   string    `json:"subject"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	Status    string    `json:"status"`
	Error     string    `json:"error,omitempty"`
}

// getEmailHandler handles GET /api/email/get. It retrieves, decrypts, and decompresses email content.
func (srv *Server) getEmailHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("getEmailHandler started")

	// Parse request
	var req GetEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("JSON decode error: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Invalid request"}`))
		return
	}

	// Get authenticated user_id from context
	userID, ok := GetUserIDFromContext(r)
	if !ok {
		log.Printf("User ID not found in context")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"Authentication required"}`))
		return
	}

	if req.EmailID == "" {
		log.Printf("Missing email_id")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Missing email_id"}`))
		return
	}

	// 1. Retrieve email metadata from database
	var (
		blobID, encryptedKeyB64, nonceB64, authTagB64, compressionAlgo string
		senderID, recipient, subject                                   string
		createdAt                                                      time.Time
	)

	err := srv.db.QueryRow(`
		SELECT encrypted_blob_url, encrypted_key, encryption_nonce, encryption_auth_tag, 
		       compression_algo, sender_id, recipient, subject, created_at
		FROM emails WHERE email_id = ?`,
		req.EmailID,
	).Scan(&blobID, &encryptedKeyB64, &nonceB64, &authTagB64, &compressionAlgo,
		&senderID, &recipient, &subject, &createdAt)

	if err != nil {
		log.Printf("Database query failed: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"Email not found"}`))
		return
	}

	// Check if the authenticated user is the sender of this email
	if senderID != userID {
		log.Printf("Unauthorized access attempt: user %s trying to access email from sender %s", userID, senderID)

		// Handle failed access with new self-destruct logic (Micro-Iteration 4.8)
		emailSecurityDB := srv.createEmailSecurityDB()
		if err := emailSecurityDB.IncrementFailedAttempts(req.EmailID); err != nil {
			if _, ok := err.(email.SelfDestructError); ok {
				// Email should be destroyed due to too many failed attempts
				log.Printf("Self-destruct threshold reached for email %s, destroying email", req.EmailID)

				// Securely delete the email
				if deleteErr := emailSecurityDB.DeleteEmailSecure(req.EmailID); deleteErr != nil {
					log.Printf("Failed to securely delete email %s: %v", req.EmailID, deleteErr)
				}

				// Return specific error for self-destruct
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusGone)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "Email deleted due to too many failed attempts",
				})
				return
			}
			log.Printf("Failed to increment failed attempts: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error":"Access denied"}`))
		return
	}

	// Check email security toggles (Micro-Iteration 4.7)
	emailSecurityDB := srv.createEmailSecurityDB()
	accessError, err := emailSecurityDB.CheckEmailAccess(req.EmailID)
	if err != nil {
		log.Printf("Failed to check email access: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to verify email access"}`))
		return
	}

	if accessError != "" {
		log.Printf("Email access denied due to security toggles: %s", accessError)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": accessError,
		})
		return
	}

	// Check read-once consumption status (Micro-Iteration 4.10)
	isConsumed, consumedAt, err := emailSecurityDB.IsReadOnceConsumed(req.EmailID)
	if err != nil {
		log.Printf("Failed to check read-once consumption status: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to verify email access"}`))
		return
	}

	if isConsumed {
		log.Printf("Email access denied: read-once email already consumed at %s", consumedAt.Format(time.RFC3339))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Email has been revoked by sender",
		})
		return
	}

	// 2. Decode base64 fields
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

	// 3. Retrieve encrypted blob from R2
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Use injected R2 client if available, otherwise fall back to environment-based client
	var encryptedBlob []byte
	var r2Err error
	if srv.r2Client != nil {
		encryptedBlob, r2Err = srv.r2Client.GetEmail(ctx, blobID)
	} else {
		encryptedBlob, r2Err = storage.GetEmailFromR2(ctx, blobID)
	}
	if r2Err != nil {
		log.Printf("Failed to retrieve blob from R2: %v", r2Err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to retrieve encrypted content"}`))
		return
	}

	// 4. Separate ciphertext and auth tag
	if len(encryptedBlob) < 16 {
		log.Printf("Encrypted blob too short: %d bytes", len(encryptedBlob))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Invalid encrypted content"}`))
		return
	}

	ciphertext := encryptedBlob[:len(encryptedBlob)-16]
	blobAuthTag := encryptedBlob[len(encryptedBlob)-16:]

	// 5. Verify auth tag matches
	if !bytes.Equal(authTag, blobAuthTag) {
		log.Printf("Auth tag mismatch")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Content integrity check failed"}`))
		return
	}

	// 6. Create EncryptedData struct for decryption
	encryptedData := &auth.EncryptedData{
		Ciphertext: ciphertext,
		Key:        encryptedKey,
		Nonce:      nonce,
		AuthTag:    authTag,
	}

	// 7. Decrypt the content
	compressed, err := auth.DecryptAES256GCM(encryptedData)
	if err != nil {
		log.Printf("Decryption failed: %v", err)

		// Handle failed access with new self-destruct logic (Micro-Iteration 4.8)
		emailSecurityDB := srv.createEmailSecurityDB()
		if err := emailSecurityDB.IncrementFailedAttempts(req.EmailID); err != nil {
			if _, ok := err.(email.SelfDestructError); ok {
				// Email should be destroyed due to too many failed attempts
				log.Printf("Self-destruct threshold reached for email %s, destroying email", req.EmailID)

				// Securely delete the email
				if deleteErr := emailSecurityDB.DeleteEmailSecure(req.EmailID); deleteErr != nil {
					log.Printf("Failed to securely delete email %s: %v", req.EmailID, deleteErr)
				}

				// Return generic error to avoid information leakage
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "Email has been revoked by sender",
				})
				return
			}
			log.Printf("Failed to increment failed attempts: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Decryption failed"}`))
		return
	}

	// 8. Decompress the content
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

	// 9. Handle read-once consumption (Micro-Iteration 4.10)
	// Mark email as consumed before returning content to prevent race conditions
	consumerDevice := "unknown" // TODO: Extract from request headers or user agent
	consumedAt, err = emailSecurityDB.MarkReadOnceConsumed(req.EmailID, consumerDevice)
	if err != nil {
		if _, ok := err.(email.ReadOnceConsumedError); ok {
			// Another request already consumed this email
			log.Printf("Email access denied: read-once email already consumed by another request")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "Email has been revoked by sender",
			})
			return
		}
		log.Printf("Failed to mark read-once consumed: %v", err)
		// Don't fail the request for non-read-once emails
	} else {
		log.Printf("Successfully marked email %s as consumed at %s", req.EmailID, consumedAt.Format(time.RFC3339))

		// Check if email should be deleted after consumption
		if toggles, err := emailSecurityDB.GetEmailSecurityTogglesForAccess(req.EmailID); err == nil && toggles != nil && toggles.ShouldSelfDestructOnReadOnce() {
			log.Printf("Email %s configured for deletion after read, performing secure deletion", req.EmailID)
			if deleteErr := emailSecurityDB.DeleteEmailSecure(req.EmailID); deleteErr != nil {
				log.Printf("Failed to delete email after read-once consumption: %v", deleteErr)
				// Continue with returning content even if deletion fails
			} else {
				log.Printf("Successfully deleted email %s after read-once consumption", req.EmailID)
			}
		}
	}

	// 10. Update access tracking and reset failed attempts on successful access (Micro-Iteration 4.8)
	if err := emailSecurityDB.ResetFailedAttempts(req.EmailID); err != nil {
		log.Printf("Failed to reset failed attempts: %v", err)
		// Don't fail the request for tracking errors
	}

	// Update other access tracking
	_, err = srv.db.Exec(`
		UPDATE emails SET 
			last_access_at = CURRENT_TIMESTAMP,
			access_count = access_count + 1
		WHERE email_id = ?`,
		req.EmailID,
	)
	if err != nil {
		log.Printf("Failed to update access tracking: %v", err)
		// Don't fail the request for tracking errors
	}

	// 10. Return decrypted email
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(GetEmailResponse{
		EmailID:   req.EmailID,
		SenderID:  senderID,
		Recipient: recipient,
		Subject:   subject,
		Body:      string(plaintext),
		CreatedAt: createdAt,
		Status:    "success",
	})
}

// EmailDeletedError is returned when an email is deleted due to too many failed attempts
type EmailDeletedError struct {
	EmailID string
}

func (e EmailDeletedError) Error() string {
	return fmt.Sprintf("email %s deleted due to too many failed attempts", e.EmailID)
}

// handleFailedAccess increments the fail count and deletes the email if limit is reached
func (srv *Server) handleFailedAccess(emailID, blobID string, currentFailCount int, reason string) error {
	// Get client IP for logging
	clientIP := "unknown"
	if r := srv.getCurrentRequest(); r != nil {
		clientIP = r.RemoteAddr
	}

	// Log the failed attempt
	log.Printf("Failed access attempt for email %s: %s (IP: %s)", emailID, reason, clientIP)

	// Increment fail count
	newFailCount := currentFailCount + 1

	// Use a transaction to ensure atomicity
	tx, err := srv.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update fail count
	_, err = tx.Exec(`
		UPDATE emails SET failed_attempts = ? WHERE email_id = ?`,
		newFailCount, emailID,
	)
	if err != nil {
		return fmt.Errorf("failed to update fail count: %w", err)
	}

	// Check if limit reached
	if newFailCount >= FAIL_ATTEMPT_LIMIT {
		log.Printf("Fail limit reached for email %s: %d/%d attempts", emailID, newFailCount, FAIL_ATTEMPT_LIMIT)

		// Delete from R2 storage
		if blobID != "" {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			if err := storage.DeleteBlob(ctx, blobID); err != nil {
				log.Printf("Failed to delete email from R2: %v", err)
				// Continue with database deletion even if R2 deletion fails
			} else {
				log.Printf("Successfully deleted email blob %s from R2", blobID)
			}
		}

		// Delete email metadata from database
		_, err = tx.Exec(`DELETE FROM emails WHERE email_id = ?`, emailID)
		if err != nil {
			return fmt.Errorf("failed to delete email metadata: %w", err)
		}

		log.Printf("Deleted email %s due to too many failed attempts", emailID)

		// Commit the transaction
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

		// Return special error to indicate email was deleted
		return EmailDeletedError{EmailID: emailID}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// getCurrentRequest returns the current request from context (for logging purposes)
func (srv *Server) getCurrentRequest() *http.Request {
	// This is a simplified approach - in a real implementation you might want to pass the request
	// through context or use a different approach to get the client IP
	return nil
}
