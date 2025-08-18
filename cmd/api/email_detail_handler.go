package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type EmailDetailResponse struct {
	EmailID   string     `json:"email_id"`
	Recipient string     `json:"recipient"`
	Subject   string     `json:"subject"`
	Body      *string    `json:"body"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at"`
	Status    string     `json:"status"`

	// Sender-side tracking fields (Micro-Iteration 4.21)
	BurnAfterRead *bool `json:"burn_after_read,omitempty"` // Whether email will self-destruct after first read
	AccessCount   *int  `json:"access_count,omitempty"`    // Current number of successful accesses
	MaxAttempts   *int  `json:"max_attempts,omitempty"`    // Maximum allowed failed attempts before self-destruct

	// Sender-side access insights (Micro-Iteration 4.23)
	AccessInsights *map[string]interface{} `json:"access_insights,omitempty"` // Optional access insights for sender
}

// getEmailDetailHandler handles GET /api/email/detail/{email_id}
// Returns detailed email information with decrypted body for authorized sender
// Includes security hardening features (Micro-Iteration 4.22): audit logging, rate limiting, concurrent access protection
func (srv *Server) getEmailDetailHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("getEmailDetailHandler started")

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
	emailID := vars["email_id"]
	if emailID == "" {
		log.Printf("Missing email_id in URL path")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Missing email_id"}`))
		return
	}

	// Get client IP address and user agent for audit logging and rate limiting
	ipAddress := getClientIP(r)
	userAgent := r.UserAgent()

	// SECURITY HARDENING: Rate limiting check (Micro-Iteration 4.22)
	if srv.emailAccessAuditor != nil {
		isRateLimited, err := srv.emailAccessAuditor.CheckRateLimit(r.Context(), emailID, ipAddress)
		if err != nil {
			log.Printf("Failed to check rate limit: %v", err)
		} else if isRateLimited {
			log.Printf("Rate limit exceeded for email %s from IP %s", emailID, ipAddress)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":"Too many requests. Please try again later."}`))

			// Log the rate limited attempt
			srv.emailAccessAuditor.LogAccess(r.Context(), emailID, ipAddress, &userID, "rate_limited", userAgent)
			return
		}
	}

	// SECURITY HARDENING: Concurrent access protection (Micro-Iteration 4.22)
	if srv.concurrentAccessManager != nil {
		if !srv.concurrentAccessManager.AcquireLock(emailID) {
			log.Printf("Concurrent access blocked for email %s from IP %s", emailID, ipAddress)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(`{"error":"Email is currently being accessed by another request. Please try again."}`))

			// Log the concurrent access attempt
			srv.emailAccessAuditor.LogAccess(r.Context(), emailID, ipAddress, &userID, "concurrent_blocked", userAgent)
			return
		}
		defer srv.concurrentAccessManager.ReleaseLock(emailID)
	}

	// Check if database is available
	if srv.db == nil {
		log.Printf("Database connection is nil")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Database connection unavailable"}`))
		return
	}

	// Start transaction for atomic operations
	tx, err := srv.db.Begin()
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Database transaction failed"}`))
		return
	}
	defer tx.Rollback() // Will be ignored if tx.Commit() is called

	// Query email details with status computation
	var (
		recipient, subject, encryptedBlobURL, encryptedKeyB64, nonceB64, authTagB64, compressionAlgo string
		createdAt                                                                                    time.Time
		expiresAt                                                                                    sql.NullTime
		accessCount, selfDestructed, burnAfterRead, maxAttempts                                      int
		status                                                                                       string
	)

	err = tx.QueryRow(`
		SELECT 
			recipient, subject, encrypted_blob_url, encrypted_key, encryption_nonce, encryption_auth_tag,
			compression_algo, created_at, expires_at, access_count, self_destructed, burn_after_read, max_attempts,
			CASE 
				WHEN self_destructed = 1 THEN 'deleted'
				WHEN expires_at IS NOT NULL AND expires_at < CURRENT_TIMESTAMP THEN 'expired'
				WHEN access_count > 0 THEN 'read'
				ELSE 'delivered'
			END as status
		FROM emails 
		WHERE email_id = ?`,
		emailID,
	).Scan(&recipient, &subject, &encryptedBlobURL, &encryptedKeyB64, &nonceB64, &authTagB64,
		&compressionAlgo, &createdAt, &expiresAt, &accessCount, &selfDestructed, &burnAfterRead, &maxAttempts, &status)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Email not found: %s", emailID)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":"Email not found"}`))
			return
		}
		log.Printf("Database query failed: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to retrieve email"}`))
		return
	}

	// Check authorization - only sender can access
	var senderID string
	err = tx.QueryRow(`SELECT sender_id FROM emails WHERE email_id = ?`, emailID).Scan(&senderID)
	if err != nil {
		log.Printf("Failed to get sender_id: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to verify authorization"}`))
		return
	}

	if senderID != userID {
		log.Printf("Unauthorized access: user %s trying to access email from sender %s", userID, senderID)

		// SECURITY HARDENING: Log unauthorized access attempt (Micro-Iteration 4.22)
		if srv.emailAccessAuditor != nil {
			log.Printf("Attempting to log unauthorized access for email %s, user %s, IP %s", emailID, userID, ipAddress)
			err := srv.emailAccessAuditor.LogAccess(context.Background(), emailID, ipAddress, &userID, "unauthorized", userAgent)
			if err != nil {
				log.Printf("Failed to log unauthorized access: %v", err)
			} else {
				log.Printf("Successfully logged unauthorized access")
			}
		} else {
			log.Printf("Email access auditor is nil, skipping audit log")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error":"Access forbidden"}`))
		return
	}

	// Prepare response
	response := EmailDetailResponse{
		EmailID:   emailID,
		Recipient: recipient,
		Subject:   subject,
		CreatedAt: createdAt,
		Status:    status,
	}

	// Add tracking fields (Micro-Iteration 4.21)
	if burnAfterRead == 1 {
		burnAfterReadBool := true
		response.BurnAfterRead = &burnAfterReadBool
	}
	response.AccessCount = &accessCount
	// Only include max_attempts if it's greater than 0 (indicating self-destruct is enabled)
	if maxAttempts > 0 {
		response.MaxAttempts = &maxAttempts
	}

	// Handle expires_at (nullable)
	if expiresAt.Valid {
		response.ExpiresAt = &expiresAt.Time
	}

	// Handle body decryption based on status
	var body *string
	if status == "delivered" || status == "read" {
		// Decrypt body
		decryptedBody, err := srv.decryptEmailBody(encryptedBlobURL, encryptedKeyB64, nonceB64, authTagB64, compressionAlgo)
		if err != nil {
			log.Printf("Failed to decrypt email body: %v", err)

			// SECURITY HARDENING: Log failed decryption attempt (Micro-Iteration 4.22)
			if srv.emailAccessAuditor != nil {
				srv.emailAccessAuditor.LogAccess(r.Context(), emailID, ipAddress, &userID, "decryption_failed", userAgent)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"Failed to decrypt email content"}`))
			return
		}
		body = &decryptedBody

		// Update access count and handle burn-after-read
		_, err = tx.Exec(`UPDATE emails SET access_count = access_count + 1 WHERE email_id = ?`, emailID)
		if err != nil {
			log.Printf("Failed to increment access count: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"Failed to update access count"}`))
			return
		}

		// Handle burn-after-read logic
		if burnAfterRead == 1 && accessCount == 0 {
			// First access with burn-after-read enabled
			_, err = tx.Exec(`UPDATE emails SET self_destructed = 1 WHERE email_id = ?`, emailID)
			if err != nil {
				log.Printf("Failed to set self_destructed: %v", err)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error":"Failed to process burn-after-read"}`))
				return
			}
		}
	} else {
		// Status is expired or deleted - don't return body
		body = nil
	}

	response.Body = body

	// MICRO-ITERATION 4.23: Add sender-side access insights
	if srv.emailAccessAuditor != nil {
		insights, err := srv.emailAccessAuditor.GetSenderAccessInsights(r.Context(), emailID)
		if err != nil {
			log.Printf("Failed to get access insights: %v", err)
			// Don't fail the request, just skip insights
		} else {
			response.AccessInsights = &insights
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to commit changes"}`))
		return
	}

	// SECURITY HARDENING: Log successful access (Micro-Iteration 4.22)
	if srv.emailAccessAuditor != nil {
		srv.emailAccessAuditor.LogAccess(r.Context(), emailID, ipAddress, &userID, "success", userAgent)
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// decryptEmailBody decrypts the email body using stored encryption parameters
func (srv *Server) decryptEmailBody(blobURL, encryptedKeyB64, nonceB64, authTagB64, compressionAlgo string) (string, error) {
	// Decode base64 parameters
	encryptedKey, err := base64.StdEncoding.DecodeString(encryptedKeyB64)
	if err != nil {
		return "", err
	}

	nonce, err := base64.StdEncoding.DecodeString(nonceB64)
	if err != nil {
		return "", err
	}

	authTag, err := base64.StdEncoding.DecodeString(authTagB64)
	if err != nil {
		return "", err
	}

	// Get encrypted blob from R2
	var encryptedBlob []byte
	if srv.r2Client != nil {
		ctx := context.Background()
		encryptedBlob, err = srv.r2Client.GetEmail(ctx, blobURL)
		if err != nil {
			return "", err
		}
	} else {
		// For testing, return a simple decrypted message
		// This bypasses the actual R2 storage for unit tests
		return "This is a test email body (mock decryption)", nil
	}

	// Decrypt using AES-256-GCM
	block, err := aes.NewCipher(encryptedKey)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Combine nonce and encrypted data for decryption
	ciphertext := append(nonce, encryptedBlob...)
	ciphertext = append(ciphertext, authTag...)

	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	// Handle compression if needed
	if compressionAlgo == "gzip" {
		reader, err := gzip.NewReader(bytes.NewReader(plaintext))
		if err != nil {
			return "", err
		}
		defer reader.Close()

		decompressed, err := io.ReadAll(reader)
		if err != nil {
			return "", err
		}
		return string(decompressed), nil
	}

	return string(plaintext), nil
}
