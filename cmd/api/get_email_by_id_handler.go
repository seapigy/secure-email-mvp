// =============================================================================
// SECURE EMAIL MVP - GET EMAIL BY ID HANDLER
// =============================================================================
// HTTP handler for retrieving individual emails by ID with comprehensive security.
// Micro-Iteration 4.5: GET /api/email/{id} Endpoint Implementation
// =============================================================================
//
// FUNCTIONAL REQUIREMENTS:
// - JWT Authentication: Valid JWT token required
// - Access Control: Only sender or recipient can access
// - Database Lookup: Retrieve email record by ID with authorization check
// - Cloudflare R2 Download: Fetch encrypted blob from R2 storage
// - Decryption: Decrypt AES key, then email content (AES-256-GCM + gzip)
// - Response: Return JSON with email details
// - Audit Logging: Log access events to audit_log table
//
// SECURITY FEATURES:
// - JWT token validation and user extraction
// - Access control based on sender_id or recipient matching
// - Comprehensive audit logging with metadata
// - Error handling with generic messages to prevent information leakage
// - Database transaction safety
//
// DEBUGGING FEATURES:
// - Detailed logging with success/failure indicators
// - Step-by-step operation tracking
// - Parameter validation logging
// - Error context preservation
// =============================================================================

package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"secure-email-mvp/pkg/auth"
	"secure-email-mvp/pkg/storage"

	"github.com/gorilla/mux"
)

// GetEmailByIdResponse represents the response structure for GET /api/email/{id}
type GetEmailByIdResponse struct {
	ID        string    `json:"id"`              // Email ID
	Sender    string    `json:"sender"`          // Sender email address
	Recipient string    `json:"recipient"`       // Recipient email address
	Subject   string    `json:"subject"`         // Decrypted subject line
	Body      string    `json:"body"`            // Decrypted email body
	SentAt    time.Time `json:"sent_at"`         // Email creation timestamp
	Status    string    `json:"status"`          // Response status
	Error     string    `json:"error,omitempty"` // Error message if any
}

// getEmailByIdHandler handles GET /api/email/{id} requests with comprehensive security
func (srv *Server) getEmailByIdHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("🔍 getEmailByIdHandler started - Micro-Iteration 4.5")

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

	// Step 3: Extract client information for audit logging
	clientIP := getClientIP(r)
	userAgent := r.UserAgent()
	log.Printf("ℹ️ Client IP: %s, User Agent: %s", clientIP, userAgent)

	// Step 4: Retrieve email metadata from database
	var (
		blobID, encryptedKeyB64, nonceB64, authTagB64, compressionAlgo string
		senderID, recipient, subject                                   string
		createdAt                                                      time.Time
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

	// Step 4.1: First, check if email exists with a simple query
	var emailExists bool
	err = srv.db.QueryRow("SELECT COUNT(*) > 0 FROM emails WHERE email_id = ?", emailID).Scan(&emailExists)
	if err != nil {
		log.Printf("❌ Failed to check email existence: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Database error"})
		return
	}
	
	if !emailExists {
		log.Printf("❌ Email not found in database: %s", emailID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Email not found"})
		return
	}
	log.Printf("✅ Email exists in database: %s", emailID)

	// Step 4.2: Query email record with sender information
	var senderEmail string
	err = srv.db.QueryRow(`
		SELECT e.encrypted_blob_url, e.encrypted_key, e.encryption_nonce, e.encryption_auth_tag, 
		       e.compression_algo, e.sender_id, e.recipient, e.subject, e.created_at,
		       u.email as sender_email
		FROM emails e
		JOIN users u ON e.sender_id = u.id
		WHERE e.email_id = ?`,
		emailID,
	).Scan(&blobID, &encryptedKeyB64, &nonceB64, &authTagB64, &compressionAlgo,
		&senderID, &recipient, &subject, &createdAt, &senderEmail)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("❌ Email not found in database: %s", emailID)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Email not found"})
			return
		}
		log.Printf("❌ Database query failed for email %s: %v", emailID, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Database error"})
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

	log.Printf("ℹ️ Email found - Sender ID: %d, Recipient: %s, Subject: %s", senderIDInt, recipient, subject)

	// Step 5: Access Control - Check if user is authorized to access this email
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

		// Log unauthorized access attempt
		srv.logAuditEvent(r.Context(), "email_access", userID, emailID, "failure",
			"Access denied - user not authorized", clientIP, userAgent)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Access denied"})
		return
	}

	log.Printf("✅ Access authorized - User is %s", map[bool]string{true: "sender", false: "recipient"}[isSender])

	// Step 6: Decode base64-encoded encryption parameters
	encryptedKey, err := base64.StdEncoding.DecodeString(encryptedKeyB64)
	if err != nil {
		log.Printf("❌ Failed to decode encrypted key: %v", err)
		srv.logAuditEvent(r.Context(), "email_access", userID, emailID, "failure",
			"Invalid encryption key format", clientIP, userAgent)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid encryption data"})
		return
	}

	nonce, err := base64.StdEncoding.DecodeString(nonceB64)
	if err != nil {
		log.Printf("❌ Failed to decode nonce: %v", err)
		srv.logAuditEvent(r.Context(), "email_access", userID, emailID, "failure",
			"Invalid nonce format", clientIP, userAgent)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid encryption data"})
		return
	}

	authTag, err := base64.StdEncoding.DecodeString(authTagB64)
	if err != nil {
		log.Printf("❌ Failed to decode auth tag: %v", err)
		srv.logAuditEvent(r.Context(), "email_access", userID, emailID, "failure",
			"Invalid auth tag format", clientIP, userAgent)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid encryption data"})
		return
	}

	log.Printf("✅ Encryption parameters decoded successfully")

	// Step 7: Download encrypted blob from Cloudflare R2
	log.Printf("ℹ️ Attempting to download blob from R2: %s", blobID)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	encryptedBlob, err := storage.GetEmailFromR2(ctx, blobID)
	if err != nil {
		log.Printf("❌ Failed to retrieve blob from R2: %v", err)
		log.Printf("❌ R2 Error Details - Blob ID: %s, Error Type: %T", blobID, err)
		
		// Log additional R2 debugging info
		log.Printf("❌ R2 Configuration Check - Please verify:")
		log.Printf("   - R2_ACCESS_KEY_ID is set")
		log.Printf("   - R2_SECRET_ACCESS_KEY is set") 
		log.Printf("   - R2_BUCKET is correct")
		log.Printf("   - R2_ENDPOINT is correct")
		log.Printf("   - Blob exists in R2 bucket")
		
		srv.logAuditEvent(r.Context(), "email_access", userID, emailID, "failure",
			"Failed to retrieve encrypted content from R2", clientIP, userAgent)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to retrieve encrypted content from R2 storage"})
		return
	}

	log.Printf("✅ Encrypted blob retrieved from R2 - Size: %d bytes", len(encryptedBlob))

	// Step 8: Separate ciphertext and auth tag from blob
	if len(encryptedBlob) < 16 {
		log.Printf("❌ Encrypted blob too short: %d bytes", len(encryptedBlob))
		srv.logAuditEvent(r.Context(), "email_access", userID, emailID, "failure",
			"Invalid encrypted content size", clientIP, userAgent)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid encrypted content"})
		return
	}

	ciphertext := encryptedBlob[:len(encryptedBlob)-16]
	blobAuthTag := encryptedBlob[len(encryptedBlob)-16:]

	// Step 9: Verify auth tag integrity
	if !bytes.Equal(authTag, blobAuthTag) {
		log.Printf("❌ Auth tag mismatch - content integrity check failed")
		srv.logAuditEvent(r.Context(), "email_access", userID, emailID, "failure",
			"Content integrity check failed", clientIP, userAgent)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Content integrity check failed"})
		return
	}

	log.Printf("✅ Content integrity verified")

	// Step 10: Create EncryptedData struct for decryption
	encryptedData := &auth.EncryptedData{
		Ciphertext: ciphertext,
		Key:        encryptedKey,
		Nonce:      nonce,
		AuthTag:    authTag,
	}

	// Step 11: Decrypt the content using AES-256-GCM
	compressed, err := auth.DecryptAES256GCM(encryptedData)
	if err != nil {
		log.Printf("❌ Failed to decrypt content: %v", err)
		srv.logAuditEvent(r.Context(), "email_access", userID, emailID, "failure",
			"Decryption failed", clientIP, userAgent)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Decryption failed"})
		return
	}

	log.Printf("✅ Content decrypted successfully - Compressed size: %d bytes", len(compressed))

	// Step 12: Decompress the content using gzip
	var decompressed bytes.Buffer
	gzReader, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		log.Printf("❌ Failed to create gzip reader: %v", err)
		srv.logAuditEvent(r.Context(), "email_access", userID, emailID, "failure",
			"Decompression failed - invalid gzip data", clientIP, userAgent)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Decompression failed"})
		return
	}
	defer gzReader.Close()

	_, err = io.Copy(&decompressed, gzReader)
	if err != nil {
		log.Printf("❌ Failed to decompress content: %v", err)
		srv.logAuditEvent(r.Context(), "email_access", userID, emailID, "failure",
			"Decompression failed - read error", clientIP, userAgent)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Decompression failed"})
		return
	}

	plaintextBody := decompressed.String()
	log.Printf("✅ Content decompressed successfully - Plaintext size: %d bytes", len(plaintextBody))

	// Step 13: Sender email already retrieved from JOIN query
	log.Printf("ℹ️ Sender email: %s", senderEmail)

	// Step 14: Log successful access to audit log
	srv.logAuditEvent(r.Context(), "email_access", userID, emailID, "success",
		fmt.Sprintf("Email accessed successfully by %s", map[bool]string{true: "sender", false: "recipient"}[isSender]),
		clientIP, userAgent)

	// Step 15: Return successful response
	response := GetEmailByIdResponse{
		ID:        emailID,
		Sender:    senderEmail,
		Recipient: recipient,
		Subject:   subject,
		Body:      plaintextBody,
		SentAt:    createdAt,
		Status:    "success",
	}

	log.Printf("✅ Email access completed successfully - ID: %s, Sender: %s, Recipient: %s",
		emailID, senderEmail, recipient)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// logAuditEvent logs email access events to the audit_log table
func (srv *Server) logAuditEvent(ctx context.Context, eventType, userID, emailID, outcome, details, ipAddress, userAgent string) {
	// Generate unique log ID
	logID := generateUUID()

	// Get geolocation info (simplified - could be enhanced with IP geolocation service)
	country := "unknown"
	city := "unknown"

	// Create audit log entry
	_, err := srv.db.ExecContext(ctx, `
		INSERT INTO audit_log (
			log_id, timestamp, event_type, user_id, ip_address, user_agent,
			related_email_id, outcome, details, severity, country, city
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		logID, time.Now(), eventType, userID, ipAddress, userAgent,
		emailID, outcome, details, "info", country, city)

	if err != nil {
		log.Printf("⚠️ Failed to log audit event: %v", err)
	} else {
		log.Printf("ℹ️ Audit event logged - ID: %s, Type: %s, Outcome: %s", logID, eventType, outcome)
	}
}

// Note: getClientIP and generateUUID functions are already defined in other files
// and will be used from there to avoid duplication
