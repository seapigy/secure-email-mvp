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
	"os"
	"strconv"
	"time"

	"secure-email-mvp/pkg/auth"
	"secure-email-mvp/pkg/storage"

	"github.com/gorilla/mux"
)

type ViewEmailResponse struct {
	EmailID                   string     `json:"email_id"`
	Recipient                 string     `json:"recipient"`
	Subject                   string     `json:"subject"`
	Body                      string     `json:"body"`
	CreatedAt                 time.Time  `json:"created_at"`
	Status                    string     `json:"status"`
	Error                     string     `json:"error,omitempty"`
	SelfDestructAfterAttempts bool       `json:"selfDestructAfterAttempts"`
	MaxFailedAttempts         int        `json:"maxFailedAttempts"`
	BurnAfterRead             bool       `json:"burnAfterRead"`
	IsConsumed                bool       `json:"isConsumed"`
	ExpiresAt                 *time.Time `json:"expiresAt,omitempty"`
	IsExpired                 bool       `json:"isExpired"`
}

// viewEmailHandler handles GET /api/email/view/{id}. It retrieves, decrypts, and returns a specific email.
// For burn-after-read emails, it deletes the email after first successful access.
// For self-destruct emails, it deletes the email after N failed access attempts.
func (srv *Server) viewEmailHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("viewEmailHandler started")

	// Get default max failed attempts from environment
	defaultMaxFailedAttempts := 3
	if envMax := os.Getenv("DEFAULT_MAX_FAILED_ATTEMPTS"); envMax != "" {
		if max, err := strconv.Atoi(envMax); err == nil && max > 0 {
			defaultMaxFailedAttempts = max
		}
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

	// 1. Retrieve email metadata from database including burn_after_read and access tracking
	var (
		blobID, encryptedKeyB64, nonceB64, authTagB64, compressionAlgo string
		senderID, recipient, subject                                   string
		createdAt                                                      time.Time
		selfDestructAfterAttempts                                      int
		maxFailedAttempts                                              int
		burnAfterRead                                                  int
		accessCount                                                    int
		expiresAt                                                      *time.Time
		selfDestructed                                                 int
		failedAccessAttempts                                           int
	)

	err := srv.db.QueryRow(`
		SELECT encrypted_blob_url, encrypted_key, encryption_nonce, encryption_auth_tag, 
		       compression_algo, sender_id, recipient, subject, created_at,
		       self_destruct_after_attempts, max_attempts, burn_after_read, access_count, expires_at,
		       self_destructed, failed_access_attempts
		FROM emails WHERE email_id = ?`,
		emailID,
	).Scan(&blobID, &encryptedKeyB64, &nonceB64, &authTagB64, &compressionAlgo,
		&senderID, &recipient, &subject, &createdAt, &selfDestructAfterAttempts,
		&maxFailedAttempts, &burnAfterRead, &accessCount, &expiresAt, &selfDestructed, &failedAccessAttempts)

	if err != nil {
		log.Printf("Database query failed: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"Email not found"}`))
		return
	}

	// 2. Check if the authenticated user is the sender of this email
	if senderID != userID {
		log.Printf("Unauthorized access attempt: user %s trying to access email from sender %s", userID, senderID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error":"Access denied"}`))
		return
	}

	// 3. Check if email has been self-destructed
	if selfDestructed == 1 {
		log.Printf("Email %s has been self-destructed - returning 410 Gone", emailID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusGone) // 410 Gone - resource no longer available
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":    "Email has been self-destructed and is no longer accessible",
			"status":   "self_destructed",
			"email_id": emailID,
		})
		return
	}

	// 6. Check burn-after-read logic
	isBurnAfterRead := burnAfterRead == 1
	isAlreadyConsumed := accessCount > 0

	if isBurnAfterRead && isAlreadyConsumed {
		log.Printf("Burn-after-read email %s already consumed by user %s", emailID, userID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusGone) // 410 Gone - resource no longer available
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":    "Email has already been read and consumed",
			"status":   "consumed",
			"email_id": emailID,
		})
		return
	}

	// 4. Check expiration logic
	if expiresAt != nil && time.Now().After(*expiresAt) {
		log.Printf("Email %s has expired (expired at %v) - deleting content", emailID, *expiresAt)

		// Delete from R2 storage
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := storage.DeleteBlob(ctx, blobID); err != nil {
			log.Printf("Failed to delete expired email from R2: %v", err)
			// Continue with response even if R2 deletion fails
		} else {
			log.Printf("Successfully deleted expired email blob %s from R2", blobID)
		}

		// Mark email as expired in database (soft delete approach)
		_, err = srv.db.Exec(`
			UPDATE emails SET 
				encrypted_blob_url = NULL,
				encrypted_key = NULL,
				encryption_nonce = NULL,
				encryption_auth_tag = NULL,
				sha256_hash = NULL
			WHERE email_id = ?`,
			emailID,
		)
		if err != nil {
			log.Printf("Failed to mark expired email as deleted in database: %v", err)
		} else {
			log.Printf("Successfully marked expired email %s as deleted in database", emailID)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusGone) // 410 Gone - resource no longer available
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":      "Email has expired and is no longer accessible",
			"status":     "expired",
			"email_id":   emailID,
			"expired_at": expiresAt.Format(time.RFC3339),
		})
		return
	}

	// 5. Check self-destruct logic for failed attempts
	if selfDestructAfterAttempts == 1 {
		// Use per-email max_attempts or fallback to default
		effectiveMaxAttempts := maxFailedAttempts
		if effectiveMaxAttempts <= 0 {
			effectiveMaxAttempts = defaultMaxFailedAttempts
		}

		// Check if we've exceeded the failed attempts threshold
		if failedAccessAttempts >= effectiveMaxAttempts {
			log.Printf("Self-destruct triggered for email %s: %d/%d failed attempts", emailID, failedAccessAttempts, effectiveMaxAttempts)

			// Delete from R2 storage
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			if err := storage.DeleteBlob(ctx, blobID); err != nil {
				log.Printf("Failed to delete self-destructed email from R2: %v", err)
				// Continue with response even if R2 deletion fails
			} else {
				log.Printf("Successfully deleted self-destructed email blob %s from R2", blobID)
			}

			// Mark email as self-destructed in database
			_, err = srv.db.Exec(`
				UPDATE emails SET 
					self_destructed = 1,
					deleted_at = CURRENT_TIMESTAMP,
					encrypted_blob_url = NULL,
					encrypted_key = NULL,
					encryption_nonce = NULL,
					encryption_auth_tag = NULL,
					sha256_hash = NULL
				WHERE email_id = ?`,
				emailID,
			)
			if err != nil {
				log.Printf("Failed to mark self-destructed email in database: %v", err)
			} else {
				log.Printf("Successfully marked email %s as self-destructed in database", emailID)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusGone) // 410 Gone - resource no longer available
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":    "Email has been self-destructed due to failed access attempts",
				"status":   "self_destructed",
				"email_id": emailID,
			})
			return
		}
	}

	// 7. Decode base64 fields
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

	// 8. Retrieve encrypted blob from R2
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

	// 9. Separate ciphertext and auth tag
	if len(encryptedBlob) < 16 {
		log.Printf("Encrypted blob too short: %d bytes", len(encryptedBlob))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Invalid encrypted content"}`))
		return
	}

	ciphertext := encryptedBlob[:len(encryptedBlob)-16]
	blobAuthTag := encryptedBlob[len(encryptedBlob)-16:]

	// 10. Verify auth tag matches
	if !bytes.Equal(authTag, blobAuthTag) {
		log.Printf("Auth tag mismatch")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Content integrity check failed"}`))
		return
	}

	// 11. Create EncryptedData struct for decryption
	encryptedData := &auth.EncryptedData{
		Ciphertext: ciphertext,
		Key:        encryptedKey,
		Nonce:      nonce,
		AuthTag:    authTag,
	}

	// 12. Decrypt the content
	compressed, err := auth.DecryptAES256GCM(encryptedData)
	if err != nil {
		log.Printf("Decryption failed: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Decryption failed"}`))
		return
	}

	// 13. Decompress the content
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

	// 14. Update access tracking and reset failed attempts on successful access
	_, err = srv.db.Exec(`
		UPDATE emails SET 
			last_access_at = CURRENT_TIMESTAMP,
			access_count = access_count + 1,
			failed_access_attempts = 0
		WHERE email_id = ?`,
		emailID,
	)
	if err != nil {
		log.Printf("Failed to update access tracking: %v", err)
		// Don't fail the request for tracking errors
	}

	// 15. Handle burn-after-read deletion
	if isBurnAfterRead {
		log.Printf("Burn-after-read email %s accessed by user %s - deleting content", emailID, userID)

		// Delete from R2 storage
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := storage.DeleteBlob(ctx, blobID); err != nil {
			log.Printf("Failed to delete burn-after-read email from R2: %v", err)
			// Continue with response even if R2 deletion fails
		} else {
			log.Printf("Successfully deleted burn-after-read email blob %s from R2", blobID)
		}

		// Mark email as consumed in database (soft delete approach)
		_, err = srv.db.Exec(`
			UPDATE emails SET 
				encrypted_blob_url = NULL,
				encrypted_key = NULL,
				encryption_nonce = NULL,
				encryption_auth_tag = NULL,
				sha256_hash = NULL
			WHERE email_id = ?`,
			emailID,
		)
		if err != nil {
			log.Printf("Failed to mark burn-after-read email as consumed in database: %v", err)
		} else {
			log.Printf("Successfully marked burn-after-read email %s as consumed in database", emailID)
		}
	}

	// 16. Return decrypted email with burn-after-read status
	response := ViewEmailResponse{
		EmailID:                   emailID,
		Recipient:                 recipient,
		Subject:                   subject,
		Body:                      string(plaintext),
		CreatedAt:                 createdAt,
		Status:                    "success",
		SelfDestructAfterAttempts: selfDestructAfterAttempts == 1,
		MaxFailedAttempts:         maxFailedAttempts,
		BurnAfterRead:             isBurnAfterRead,
		IsConsumed:                isBurnAfterRead && isAlreadyConsumed,
		ExpiresAt:                 expiresAt,
		IsExpired:                 expiresAt != nil && time.Now().After(*expiresAt),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handleFailedAccessAttempt increments the failed access attempts counter and triggers self-destruct if needed
func (srv *Server) handleFailedAccessAttempt(emailID string, blobID string, currentFailedAttempts int, maxAttempts int) error {
	// Increment failed attempts atomically
	newFailedAttempts := currentFailedAttempts + 1

	// Use a transaction to ensure atomicity
	tx, err := srv.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update failed attempts count
	_, err = tx.Exec(`
		UPDATE emails SET 
			failed_access_attempts = ?
		WHERE email_id = ?`,
		newFailedAttempts, emailID,
	)
	if err != nil {
		return fmt.Errorf("failed to update failed attempts: %w", err)
	}

	// Check if we've exceeded the threshold
	if newFailedAttempts >= maxAttempts {
		log.Printf("Self-destruct triggered for email %s: %d/%d failed attempts", emailID, newFailedAttempts, maxAttempts)

		// Delete from R2 storage (only if blobID is provided)
		if blobID != "" {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			if err := storage.DeleteBlob(ctx, blobID); err != nil {
				log.Printf("Failed to delete self-destructed email from R2: %v", err)
				// Continue with database update even if R2 deletion fails
			} else {
				log.Printf("Successfully deleted self-destructed email blob %s from R2", blobID)
			}
		} else {
			log.Printf("No blobID provided for R2 deletion, skipping R2 cleanup")
		}

		// Mark email as self-destructed in database
		_, err = tx.Exec(`
			UPDATE emails SET 
				self_destructed = 1,
				deleted_at = CURRENT_TIMESTAMP,
				encrypted_blob_url = NULL,
				encrypted_key = NULL,
				encryption_nonce = NULL,
				encryption_auth_tag = NULL,
				sha256_hash = NULL
			WHERE email_id = ?`,
			emailID,
		)
		if err != nil {
			return fmt.Errorf("failed to mark email as self-destructed: %w", err)
		}

		log.Printf("Successfully marked email %s as self-destructed in database", emailID)
	} else {
		log.Printf("Failed attempt for email %s: %d/%d", emailID, newFailedAttempts, maxAttempts)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// testSelfDestructHandler is a TEST ONLY endpoint to simulate failed access attempts
// This endpoint is only available when SIMULATE_SELF_DESTRUCT=1
func (srv *Server) testSelfDestructHandler(w http.ResponseWriter, r *http.Request) {
	// TEST ONLY: Check if simulation is enabled
	if os.Getenv("SIMULATE_SELF_DESTRUCT") != "1" {
		http.Error(w, "Self-destruct simulation not enabled", http.StatusForbidden)
		return
	}

	var req struct {
		EmailID string `json:"email_id"`
		Action  string `json:"action"` // "increment_failed" or "reset"
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.EmailID == "" {
		http.Error(w, "Email ID required", http.StatusBadRequest)
		return
	}

	// Get current failed attempts
	var currentFailedAttempts, maxAttempts int
	err := srv.db.QueryRow(`
		SELECT failed_access_attempts, max_attempts 
		FROM emails WHERE email_id = ?`,
		req.EmailID,
	).Scan(&currentFailedAttempts, &maxAttempts)

	if err != nil {
		http.Error(w, "Email not found", http.StatusNotFound)
		return
	}

	switch req.Action {
	case "increment_failed":
		// Simulate a failed access attempt
		if err := srv.handleFailedAccessAttempt(req.EmailID, "", currentFailedAttempts, maxAttempts); err != nil {
			log.Printf("Failed to simulate failed attempt: %v", err)
			http.Error(w, "Failed to simulate failed attempt", http.StatusInternalServerError)
			return
		}

		// Get updated failed attempts count
		var newFailedAttempts int
		err = srv.db.QueryRow(`
			SELECT failed_access_attempts FROM emails WHERE email_id = ?`,
			req.EmailID,
		).Scan(&newFailedAttempts)

		if err != nil {
			http.Error(w, "Failed to get updated count", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"email_id":        req.EmailID,
			"failed_attempts": newFailedAttempts,
			"max_attempts":    maxAttempts,
			"self_destructed": newFailedAttempts >= maxAttempts,
		})

	case "reset":
		// Reset failed attempts
		_, err = srv.db.Exec(`
			UPDATE emails SET failed_access_attempts = 0 WHERE email_id = ?`,
			req.EmailID,
		)
		if err != nil {
			http.Error(w, "Failed to reset attempts", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"email_id":        req.EmailID,
			"failed_attempts": 0,
			"max_attempts":    maxAttempts,
			"reset":           true,
		})

	default:
		http.Error(w, "Invalid action", http.StatusBadRequest)
	}
}
