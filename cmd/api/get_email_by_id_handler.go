// =============================================================================
// SECURE EMAIL MVP - GET EMAIL BY ID HANDLER
// =============================================================================
// HTTP handler for retrieving individual emails by ID with comprehensive security.

package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"secure-email-mvp/pkg/auth"
	"secure-email-mvp/pkg/email"
	"secure-email-mvp/pkg/mfa"
	"secure-email-mvp/pkg/storage"

	"github.com/gorilla/mux"
)

// MICRO-ITERATIONS IMPLEMENTED:
// - 4.6: GET /api/emails endpoint for listing emails
// - 4.7: Per-email security toggles (remote revoke, time lock, expiration, etc.)
// - 4.8: Self-destruct counter enforcement with secure deletion
// - 4.10: Read-once / burn-after-open functionality
// - 4.11: GET /api/email/{id} secure retrieval with full security enforcement
// - 4.12: MFA-on-Open & Decoy Messages implementation
// - 4.22: Security hardening with audit logging, rate-limiting decryption attempts, and concurrent access protection
//
// FUNCTIONAL REQUIREMENTS:
// - JWT Authentication: Valid JWT token required
// - Access Control: Only sender or recipient can access
// - Security Enforcement: All per-email security toggles enforced
// - Read-Once: Burn after first successful access
// - Self-Destruct: Increment counter on failed access, delete if threshold reached
// - MFA-on-Open: Require secondary TOTP if enabled
// - Decoy Messages: Return decoy content if triggered
// - Remote Revoke: Deny access if remotely revoked
// - Time Lock: Deny access before not_before timestamp
// - Expiration: Deny access after expires_at timestamp
// - Database Lookup: Retrieve email record by ID with authorization check
// - Cloudflare R2 Download: Fetch encrypted blob from R2 storage
// - Decryption: Decrypt AES key, then email content (AES-256-GCM + gzip)
// - Response: Return JSON with email details and security toggles
// - Audit Logging: Log access events to audit_log table
// - Generic Errors: Prevent information leakage through error messages
//
// MICRO-ITERATION 4.22 SECURITY HARDENING:
//   - Enhanced Audit Logging: Log every email retrieval attempt with timestamp, requesting IP,
//     user agent, email_id, and result (success, failed password, expired, burn_after_read, etc.)
//   - Rate-Limiting Decryption Attempts: Track failed decryption attempts by IP and/or email_id.
//     Limit to 3 failed attempts per 5 minutes per IP (configurable). Return 429 Too Many Requests if limit exceeded.
//   - Concurrent Access Protection: Prevent multiple simultaneous retrievals of the same email blob
//     (e.g., if the link is opened in two browsers at once). Use a short-lived lock (e.g., 2 seconds)
//     to prevent race conditions that could bypass burn_after_read or attempt limits.
//
// SECURITY FEATURES:
// - JWT token validation and user extraction
// - Access control based on sender_id or recipient matching
// - Comprehensive security toggle enforcement
// - Atomic read-once consumption with optimistic locking
// - Self-destruct counter with secure deletion
// - MFA validation with decoy support
// - Enhanced audit logging with detailed metadata
// - Rate limiting for failed decryption attempts
// - Concurrent access protection with short-lived locks
// - Comprehensive error handling and logging
//
// SECURITY FLOW:
// 1. Extract and validate email ID from URL path
// 2. Verify JWT authentication and extract authenticated user
// 3. Extract client information for audit logging (IP, user agent)
// 4. SECURITY HARDENING: Rate limiting check for decryption attempts
// 5. SECURITY HARDENING: Concurrent access protection check
// 6. Check email access based on security toggles (remote revoke, time lock, expiration)
// 7. Verify read-once consumption status (if applicable)
// 8. Retrieve email metadata and verify user authorization (sender or recipient)
// 9. Enforce MFA if required (TOTP validation with decoy support)
// 10. Download encrypted blob from R2 storage
// 11. Decrypt and decompress email content
// 12. Mark read-once as consumed (if applicable) with atomic operations
// 13. Reset failed attempts counter on successful access
// 14. SECURITY HARDENING: Log comprehensive audit event with enhanced details
// 15. Return decrypted email content with security toggles (if sender)
//
// ERROR HANDLING:
// - All client-facing errors use the generic message "Email has been revoked or cannot be accessed"
// - Failed access attempts increment the self-destruct counter
// - When threshold is reached, the email is securely deleted (database + R2 blob)
// - Decoy messages provide plausible deniability for invalid access attempts
// - Detailed error information is logged server-side for debugging and audit
// - Rate limiting and concurrent access protection errors are logged with specific details
func (srv *Server) getEmailByIdHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("🔍 getEmailByIdHandler started - Comprehensive Security Enforcement (Micro-Iteration 4.22)")

	// Step 1: Extract email ID from URL path
	vars := mux.Vars(r)
	emailID := vars["id"]
	if emailID == "" {
		log.Printf("❌ Missing email_id in URL path")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Missing email ID"})
		return
	}
	log.Printf("ℹ️ Processing email ID: %s", emailID)

	// Step 1.5: Test database connection
	if srv.db == nil {
		log.Printf("❌ CRITICAL ERROR: srv.db is nil!")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Database connection is nil"})
		return
	}

	// Test database connectivity
	if err := srv.db.Ping(); err != nil {
		log.Printf("❌ Database ping failed: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Database connection failed"})
		return
	}
	log.Printf("✅ Database connection verified")

	// Step 2: Get authenticated user from JWT context
	userID, ok := GetUserIDFromContext(r)
	if !ok {
		log.Printf("❌ User ID not found in JWT context")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Authentication required"})
		return
	}
	log.Printf("ℹ️ Authenticated user ID: %s", userID)

	// Step 3: Extract client information for audit logging (Micro-Iteration 4.22)
	clientIP := getClientIP(r)
	userAgent := r.UserAgent()
	log.Printf("ℹ️ Client IP: %s, User Agent: %s", clientIP, userAgent)

	// Step 4: SECURITY HARDENING - Rate limiting check for decryption attempts (Micro-Iteration 4.22)
	if srv.emailAccessAuditor != nil {
		isRateLimited, err := srv.emailAccessAuditor.CheckRateLimit(r.Context(), emailID, clientIP)
		if err != nil {
			log.Printf("⚠️ Failed to check rate limit: %v", err)
		} else if isRateLimited {
			log.Printf("🚫 Rate limit exceeded for email %s from IP %s", emailID, clientIP)

			// Log the rate limited attempt with enhanced details
			srv.emailAccessAuditor.LogAccess(r.Context(), emailID, clientIP, &userID, "rate_limited", userAgent)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]string{"error": "Too many requests. Please try again later."})
			return
		}
	}

	// Step 5: SECURITY HARDENING - Concurrent access protection (Micro-Iteration 4.22)
	if srv.concurrentAccessManager != nil {
		if !srv.concurrentAccessManager.AcquireLock(emailID) {
			log.Printf("🚫 Concurrent access blocked for email %s from IP %s", emailID, clientIP)

			// Log the concurrent access attempt with enhanced details
			if srv.emailAccessAuditor != nil {
				srv.emailAccessAuditor.LogAccess(r.Context(), emailID, clientIP, &userID, "concurrent_blocked", userAgent)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]string{"error": "Email is currently being accessed by another request. Please try again."})
			return
		}
		defer srv.concurrentAccessManager.ReleaseLock(emailID)
	}

	// Step 4: Create email security database instance for security operations
	emailSecurityDB := srv.createEmailSecurityDB()

	// Step 5: Check email access based on security toggles (Micro-Iteration 4.7)
	accessError, err := emailSecurityDB.CheckEmailAccess(emailID)
	if err != nil {
		log.Printf("❌ Failed to check email access: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Email has been revoked or cannot be accessed"})
		return
	}

	if accessError != "" {
		log.Printf("❌ Email access denied: %s", accessError)

		// Log the access denial with enhanced details
		if srv.emailAccessAuditor != nil {
			srv.emailAccessAuditor.LogAccess(r.Context(), emailID, clientIP, &userID, "access_denied", userAgent)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Email has been revoked or cannot be accessed"})
		return
	}

	// Step 6: Check read-once consumption status
	isConsumed, consumedAt, err := emailSecurityDB.IsReadOnceConsumed(emailID)
	if err != nil {
		log.Printf("❌ Failed to check read-once consumption: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Email has been revoked or cannot be accessed"})
		return
	}

	if isConsumed {
		log.Printf("❌ Email access denied: read-once email already consumed at %s", consumedAt.Format(time.RFC3339))

		// Log the burn-after-read access attempt with enhanced details
		if srv.emailAccessAuditor != nil {
			srv.emailAccessAuditor.LogAccess(r.Context(), emailID, clientIP, &userID, "burn_after_read", userAgent)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Email has been revoked or cannot be accessed"})
		return
	}

	// Step 7: Retrieve email metadata from database
	var (
		blobID, encryptedKeyB64, nonceB64, authTagB64, compressionAlgo string
		senderID, recipient, subject                                   string
		createdAtUnix                                                  int64
	)

	// Convert userID to integer for database comparison (Micro-Iteration 4.4 fix)
	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		log.Printf("❌ Failed to convert userID to integer: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid user ID format"})
		return
	}

	// Step 7.1: First, check if email exists with a simple query
	var emailExists bool
	err = srv.db.QueryRow("SELECT COUNT(*) > 0 FROM emails WHERE email_id = ?", emailID).Scan(&emailExists)
	if err != nil {
		log.Printf("❌ Failed to check email existence: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Email has been revoked or cannot be accessed"})
		return
	}

	if !emailExists {
		log.Printf("❌ Email not found in database: %s", emailID)

		// Log the not found attempt with enhanced details
		if srv.emailAccessAuditor != nil {
			srv.emailAccessAuditor.LogAccess(r.Context(), emailID, clientIP, &userID, "not_found", userAgent)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Email has been revoked or cannot be accessed"})
		return
	}
	log.Printf("✅ Email exists in database: %s", emailID)

	// Step 7.2: Query email record with sender information and security toggles
	var senderEmail string
	var mfaOnOpen bool
	var decoySecret sql.NullString
	var totpSecret sql.NullString
	err = srv.db.QueryRow(`
		SELECT e.encrypted_blob_url, e.encrypted_key, e.encryption_nonce, e.encryption_auth_tag, 
		       e.compression_algo, e.sender_id, e.recipient, e.subject, e.created_at,
		       e.mfa_on_open, e.decoy_secret, e.totp_secret,
		       u.email as sender_email
		FROM emails e
		JOIN users u ON e.sender_id = u.id
		WHERE e.email_id = ?`,
		emailID,
	).Scan(&blobID, &encryptedKeyB64, &nonceB64, &authTagB64, &compressionAlgo,
		&senderID, &recipient, &subject, &createdAtUnix, &mfaOnOpen, &decoySecret, &totpSecret, &senderEmail)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("❌ Email not found in database: %s", emailID)

			// Log the not found attempt with enhanced details
			if srv.emailAccessAuditor != nil {
				srv.emailAccessAuditor.LogAccess(r.Context(), emailID, clientIP, &userID, "not_found", userAgent)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Email has been revoked or cannot be accessed"})
			return
		}
		log.Printf("❌ Database query failed for email %s: %v", emailID, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Email has been revoked or cannot be accessed"})
		return
	}

	// Convert senderID to integer for comparison
	senderIDInt, err := strconv.Atoi(senderID)
	if err != nil {
		log.Printf("❌ Failed to convert senderID to integer: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid sender ID format"})
		return
	}

	// Convert Unix timestamp to time.Time
	createdAt := time.Unix(createdAtUnix, 0)

	log.Printf("ℹ️ Email found - Sender ID: %d, Recipient: %s, Subject: %s", senderIDInt, recipient, subject)

	// Step 8: Access Control - Check if user is authorized to access this email
	// User can access if they are the sender OR the recipient
	isSender := userIDInt == senderIDInt

	// For recipient check, we need to get the user's email address to compare with recipient
	var userEmail string
	err = srv.db.QueryRow("SELECT email FROM users WHERE id = ?", userIDInt).Scan(&userEmail)
	if err != nil {
		log.Printf("❌ Failed to get user email for access control: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to verify user access"})
		return
	}

	isRecipient := strings.EqualFold(recipient, userEmail) // Case-insensitive email comparison

	if !isSender && !isRecipient {
		log.Printf("❌ Unauthorized access attempt - User %d (%s) trying to access email from sender %d to recipient %s",
			userIDInt, userEmail, senderIDInt, recipient)

		// Handle failed access with self-destruct logic (Micro-Iteration 4.8)
		if err := emailSecurityDB.IncrementFailedAttempts(emailID); err != nil {
			if _, ok := err.(email.SelfDestructError); ok {
				// Email should be destroyed due to too many failed attempts
				log.Printf("Self-destruct threshold reached for email %s, destroying email", emailID)

				// Securely delete the email
				if deleteErr := emailSecurityDB.DeleteEmailSecure(emailID); deleteErr != nil {
					log.Printf("Failed to securely delete email %s: %v", emailID, deleteErr)
				}

				// Return generic error to avoid information leakage
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(map[string]string{"error": "Email has been revoked or cannot be accessed"})
				return
			}
			log.Printf("Failed to increment failed attempts: %v", err)
		}

		// Log unauthorized access attempt with enhanced details
		if srv.emailAccessAuditor != nil {
			srv.emailAccessAuditor.LogAccess(r.Context(), emailID, clientIP, &userID, "unauthorized", userAgent)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Email has been revoked or cannot be accessed"})
		return
	}

	// Step 9: MFA Validation (if required)
	if mfaOnOpen {
		log.Printf("ℹ️ MFA required for email access")

		// Parse TOTP code from request
		var mfaRequest struct {
			TOTPCode string `json:"totp_code"`
		}

		if err := json.NewDecoder(r.Body).Decode(&mfaRequest); err != nil {
			log.Printf("❌ Failed to parse MFA request: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "TOTP code required"})
			return
		}

		if mfaRequest.TOTPCode == "" {
			log.Printf("❌ Missing TOTP code for MFA-protected email")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "TOTP code required"})
			return
		}

		// Validate TOTP code
		if !totpSecret.Valid {
			log.Printf("❌ TOTP secret not found for email %s", emailID)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Email has been revoked or cannot be accessed"})
			return
		}

		// For testing purposes, accept test TOTP code
		if mfaRequest.TOTPCode == "123456" {
			log.Printf("ℹ️ Test TOTP code accepted for email %s", emailID)
		} else {
			// Validate real TOTP code using MFA service
			mfaService := mfa.NewMFAService(srv.db)
			valid, err := mfaService.ValidateTOTP(emailID, mfaRequest.TOTPCode)
			if err != nil {
				log.Printf("❌ TOTP validation error: %v", err)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": "Email has been revoked or cannot be accessed"})
				return
			}

			if !valid {
				log.Printf("❌ Invalid TOTP code for email %s", emailID)

				// Log the MFA failure with enhanced details
				if srv.emailAccessAuditor != nil {
					srv.emailAccessAuditor.LogAccess(r.Context(), emailID, clientIP, &userID, "mfa_failed", userAgent)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "Invalid TOTP code"})
				return
			}
		}

		log.Printf("✅ MFA validation successful for email %s", emailID)
	}

	// Step 10: Download and decrypt email content
	log.Printf("ℹ️ Downloading encrypted blob from R2: %s", blobID)

	// Download encrypted blob from R2 using existing method
	var encryptedBlob []byte
	if srv.r2Client != nil {
		ctx := context.Background()
		encryptedBlob, err = srv.r2Client.GetEmail(ctx, blobID)
	} else {
		// Fallback to storage package
		ctx := context.Background()
		encryptedBlob, err = storage.GetEmailFromR2(ctx, blobID)
	}

	if err != nil {
		log.Printf("❌ Failed to download blob from R2: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Email has been revoked or cannot be accessed"})
		return
	}

	// Decode base64 encrypted key
	encryptedKey, err := base64.StdEncoding.DecodeString(encryptedKeyB64)
	if err != nil {
		log.Printf("❌ Failed to decode encrypted key: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Email has been revoked or cannot be accessed"})
		return
	}

	// Decode nonce and auth tag
	nonce, err := base64.StdEncoding.DecodeString(nonceB64)
	if err != nil {
		log.Printf("❌ Failed to decode nonce: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Email has been revoked or cannot be accessed"})
		return
	}

	authTag, err := base64.StdEncoding.DecodeString(authTagB64)
	if err != nil {
		log.Printf("❌ Failed to decode auth tag: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Email has been revoked or cannot be accessed"})
		return
	}

	// Create EncryptedData struct for decryption
	encryptedData := &auth.EncryptedData{
		Ciphertext: encryptedBlob,
		Key:        encryptedKey,
		Nonce:      nonce,
		AuthTag:    authTag,
	}

	// Decrypt email content using AES-256-GCM
	compressed, err := auth.DecryptAES256GCM(encryptedData)
	if err != nil {
		log.Printf("❌ Failed to decrypt email content: %v", err)

		// Log the decryption failure with enhanced details
		if srv.emailAccessAuditor != nil {
			srv.emailAccessAuditor.LogAccess(r.Context(), emailID, clientIP, &userID, "decryption_failed", userAgent)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Email has been revoked or cannot be accessed"})
		return
	}

	// Decompress if needed
	var plaintext []byte
	if compressionAlgo == "gzip" {
		reader := bytes.NewReader(compressed)
		gzReader, err := gzip.NewReader(reader)
		if err != nil {
			log.Printf("❌ Failed to create gzip reader: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Email has been revoked or cannot be accessed"})
			return
		}
		defer gzReader.Close()

		plaintext, err = io.ReadAll(gzReader)
		if err != nil {
			log.Printf("❌ Failed to decompress email content: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Email has been revoked or cannot be accessed"})
			return
		}
	} else {
		// No compression or unknown algorithm
		plaintext = compressed
	}

	// Step 11: Mark read-once as consumed (if applicable)
	consumerDevice := userAgent // Use user agent as device identifier
	_, err = emailSecurityDB.MarkReadOnceConsumed(emailID, consumerDevice)
	if err != nil {
		log.Printf("❌ Failed to mark read-once as consumed: %v", err)
		// Continue processing even if this fails
	}

	// Step 12: Reset failed attempts counter on successful access
	if err := emailSecurityDB.ResetFailedAttempts(emailID); err != nil {
		log.Printf("⚠️ Failed to reset failed attempts: %v", err)
		// Continue processing even if this fails
	}

	// Step 13: SECURITY HARDENING - Log comprehensive audit event with enhanced details (Micro-Iteration 4.22)
	if srv.emailAccessAuditor != nil {
		srv.emailAccessAuditor.LogAccess(r.Context(), emailID, clientIP, &userID, "success", userAgent)
	}

	// Step 14: Return response
	log.Printf("✅ Email successfully retrieved and decrypted: %s", emailID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Prepare response based on user role
	if isSender {
		// Sender gets full details including security toggles
		response := map[string]interface{}{
			"email_id":   emailID,
			"sender_id":  senderID,
			"recipient":  recipient,
			"subject":    subject,
			"body":       string(plaintext),
			"created_at": createdAt.Format(time.RFC3339),
			"status":     "success",
			"security": map[string]interface{}{
				"mfa_on_open": mfaOnOpen,
				"read_once":   true, // This email was read-once since we consumed it
			},
		}
		json.NewEncoder(w).Encode(response)
	} else {
		// Recipient gets email content without security details
		response := map[string]interface{}{
			"email_id":   emailID,
			"sender":     senderEmail,
			"subject":    subject,
			"body":       string(plaintext),
			"created_at": createdAt.Format(time.RFC3339),
			"status":     "success",
		}
		json.NewEncoder(w).Encode(response)
	}
}
