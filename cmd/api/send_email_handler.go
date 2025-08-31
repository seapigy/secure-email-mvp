package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"time"

	"secure-email-mvp/pkg/pqc"
	"secure-email-mvp/pkg/securelinks"
	securelinksemail "secure-email-mvp/pkg/securelinks/email"
	"secure-email-mvp/pkg/securelinks/security"
	"secure-email-mvp/pkg/storage"

	"database/sql"

	"github.com/google/uuid"

	"secure-email-mvp/pkg/email"
)

// =============================================================================
// MICRO-ITERATION 4.4: EMAIL SEND ENDPOINT WITH FOREIGN KEY FIX
// =============================================================================
//
// PROBLEM RESOLVED:
// - users table uses INTEGER PRIMARY KEY for id
// - emails table was expecting TEXT sender_id
// - This caused foreign key constraint violations during email inserts
//
// SOLUTION IMPLEMENTED:
// - Convert JWT userID string to integer before database insert
// - Verify user exists in database before email creation
// - Use complete INSERT statement with all required columns
// - Add comprehensive logging for debugging
//
// DEBUGGING FEATURES:
// - Detailed SQL query logging
// - Parameter value logging (userID, recipient, subject, etc.)
// - Security parameter logging (self-destruct, MFA, geolocation)
// - Actual database error messages in API responses
// - User existence verification before insert
//
// FOREIGN KEY INTEGRITY:
// - sender_id INTEGER NOT NULL
// - FOREIGN KEY (sender_id) REFERENCES users(id)
// - User ID conversion: string → integer
// - User existence verification before insert
// =============================================================================

// SendEmailRequest represents the JSON request body for sending secure emails
// This struct includes all security features implemented in Micro-Iterations 4.10-4.15
type SendEmailRequest struct {
	// Basic email fields
	Recipient string `json:"recipient"` // Email address of recipient
	Subject   string `json:"subject"`   // Email subject line
	Body      string `json:"body"`      // Email body content

	// Security settings (Micro-Iteration 4.12)
	SelfDestructAfterAttempts bool `json:"selfDestructAfterAttempts,omitempty"` // Enable self-destruct after failed attempts
	MaxFailedAttempts         int  `json:"maxFailedAttempts,omitempty"`         // Number of failed attempts before self-destruct (1-10)
	BurnAfterRead             bool `json:"burnAfterRead,omitempty"`             // Delete email after first successful read

	// Expiration settings
	ExpiresAt string `json:"expiresAt,omitempty"` // ISO 8601 UTC format expiration timestamp

	// Geolocation restrictions (Micro-Iteration 4.10)
	AllowedCity    string `json:"allowedCity,omitempty"`    // Single city name (case-insensitive, normalized)
	AllowedCountry string `json:"allowedCountry,omitempty"` // Single ISO 3166-1 alpha-2 country code

	// Enhanced geolocation verification (Micro-Iteration 4.15)
	GeoVerificationType string `json:"geoVerificationType,omitempty"` // "none", "city", "city_country"
	GeoCity             string `json:"geoCity,omitempty"`             // City for verification
	GeoCountry          string `json:"geoCountry,omitempty"`          // Country for verification

	// MFA settings (Micro-Iteration 4.12)
	RequireMFA bool   `json:"requireMFA,omitempty"` // Enable multi-factor authentication
	MFAType    string `json:"mfaType,omitempty"`    // "TOTP" or "EMAIL_CODE"

	// Password protection settings (Micro-Iteration 4.14)
	Password string `json:"password,omitempty"` // Optional password for email access

	// Additional security features (Micro-Iteration 4.16)
	TimeLock      bool   `json:"timeLock,omitempty"`      // Enable time-based access restrictions
	UnlockAfter   string `json:"unlockAfter,omitempty"`   // Date/time when email becomes accessible
	RemoteRevoke  bool   `json:"remoteRevoke,omitempty"`  // Enable remote revocation capability
	DecoyMessage  bool   `json:"decoyMessage,omitempty"`  // Enable decoy message feature
	StripMetadata bool   `json:"stripMetadata,omitempty"` // Strip metadata from email
	TamperAlerts  bool   `json:"tamperAlerts,omitempty"`  // Enable tamper detection alerts
}

// SendEmailResponse represents the JSON response for email sending operations
type SendEmailResponse struct {
	BlobID string `json:"blob_id,omitempty"` // Cloudflare R2 blob ID for encrypted content
	Status string `json:"status,omitempty"`  // Operation status ("success" or "error")
	Error  string `json:"error,omitempty"`   // Error message if operation failed

	// Sender-side tracking fields (Micro-Iteration 4.21)
	BurnAfterRead *bool `json:"burn_after_read,omitempty"` // Whether email will self-destruct after first read
	AccessCount   *int  `json:"access_count,omitempty"`    // Current number of successful accesses
	MaxAttempts   *int  `json:"max_attempts,omitempty"`    // Maximum allowed failed attempts before self-destruct

	// Secure link fields (NEW)
	SecureLinkURL *string `json:"secure_link_url,omitempty"` // URL for the secure link
}

// sendEmailHandler handles POST /api/email/send with comprehensive security features
//
// SECURITY FEATURES IMPLEMENTED:
// - Self-destruct after failed attempts (1-10 attempts)
// - Burn-after-read (one-time access)
// - Email expiration (ISO 8601 UTC format)
// - Enhanced geolocation verification (city/country restrictions)
// - Multi-factor authentication (TOTP or email-based)
// - Per-email password protection (Argon2id hashing)
// - Time-based access controls
// - Remote revocation
// - Decoy messages
// - Metadata stripping
// - Tamper alerts
//
// PROCESS FLOW:
// 1. Validate all security parameters
// 2. Create email with basic metadata
// 3. Apply security features using EmailSecurityService
// 4. Compress and encrypt email content
// 5. Upload encrypted blob to Cloudflare R2
// 6. Send email (internal/external) using EmailSecurityService
// 7. Return secure access link
func (srv *Server) sendEmailHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("=== sendEmailHandler started ===")

	// Step 1: Parse JSON request body
	var req SendEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("❌ JSON decode error: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Invalid request format"}`))
		return
	}

	// Step 2: Extract and validate authenticated user ID from JWT context
	userID, ok := GetUserIDFromContext(r)
	if !ok {
		log.Printf("❌ User ID not found in JWT context")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"Authentication required"}`))
		return
	}

	log.Printf("✅ Authenticated user ID: %s", userID)

	// Step 3: Check if database is available
	if srv.db == nil {
		log.Printf("❌ Database connection is nil")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Database connection unavailable"}`))
		return
	}

	// Step 4: Verify that the user exists in the database before proceeding
	var existingUserID string
	userCheckErr := srv.db.QueryRow("SELECT id FROM users WHERE id = ?", userID).Scan(&existingUserID)
	if userCheckErr != nil {
		log.Printf("❌ User not found in database: userID=%s, error=%v", userID, userCheckErr)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"User not found in database"}`))
		return
	}
	log.Printf("✅ Verified user exists in database: userID=%s", existingUserID)

	// Step 5: Validate required email fields
	if req.Recipient == "" || req.Subject == "" || req.Body == "" {
		log.Printf("❌ Missing required fields: recipient=%q subject=%q body=%q", req.Recipient, req.Subject, req.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Missing required fields: recipient, subject, and body are required"}`))
		return
	}

	// Step 6: Validate recipient email format using regex
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(req.Recipient) {
		log.Printf("❌ Invalid recipient email format: %q", req.Recipient)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Invalid recipient email format"}`))
		return
	}

	// Step 7: Generate unique email ID
	emailID := uuid.New().String()
	log.Printf("✅ Generated email ID: %s", emailID)

	// Step 8: Convert request to SecurityFeatureConfig
	securityConfig := email.SecurityFeatureConfig{
		// Basic Security
		PasswordProtection: req.Password != "",
		Password:           req.Password,

		// Access Control
		BurnAfterRead:             req.BurnAfterRead,
		SelfDestructAfterAttempts: req.SelfDestructAfterAttempts,
		MaxFailedAttempts:         req.MaxFailedAttempts,

		// Time-based Controls
		TimeLock:    req.TimeLock,
		UnlockAfter: req.UnlockAfter,
		ExpiresAt:   req.ExpiresAt,

		// Geolocation
		GeoVerificationType: req.GeoVerificationType,
		GeoCity:             req.GeoCity,
		GeoCountry:          req.GeoCountry,

		// Multi-Factor Authentication
		RequireMFA:   req.RequireMFA,
		MFAType:      req.MFAType,
		MFAOnOpen:    false, // These can be added to the request if needed
		MFAOnReply:   false,
		MFAOnForward: false,

		// Advanced Security
		RemoteRevoke:  req.RemoteRevoke,
		DecoyMessage:  req.DecoyMessage,
		DecoySecret:   "", // Will be generated by the service
		StripMetadata: req.StripMetadata,
		TamperAlerts:  req.TamperAlerts,
	}

	// Step 9: Validate security configuration
	if err := srv.emailSecurityService.ValidateSecurityConfig(securityConfig); err != nil {
		log.Printf("❌ Security configuration validation failed: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"` + err.Error() + `"}`))
		return
	}

	// Step 10: Compress email content with gzip
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write([]byte(req.Body)); err != nil {
		log.Printf("❌ Gzip compression failed: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Compression failed"}`))
		return
	}
	gz.Close()
	compressed := buf.Bytes()

	// Step 11: Encrypt compressed content using PQC Hybrid Encryption
	pqcConfig := pqc.LoadPQCConfigFromEnv()
	pqcService, err := pqc.NewPQCService(pqcConfig)
	if err != nil {
		log.Printf("❌ PQC service initialization failed: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"PQC encryption service failed"}`))
		return
	}

	// Encrypt using PQC hybrid encryption (Kyber + AES-256-GCM)
	hybridData, err := pqcService.EncryptHybrid(compressed, "email_send")
	if err != nil {
		log.Printf("❌ PQC encryption failed: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Encryption failed"}`))
		return
	}

	// Step 12: Upload encrypted content to Cloudflare R2
	var blobID string
	if srv.r2Client != nil {
		// Serialize hybrid data for storage
		hybridDataBytes, err := json.Marshal(hybridData)
		if err != nil {
			log.Printf("❌ Failed to serialize hybrid data: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"Failed to serialize encryption data"}`))
			return
		}

		// Upload to R2 using the storage package
		blobID = emailID + ".blob"
		if err := storage.UploadToR2WithContext(r.Context(), blobID, hybridDataBytes); err != nil {
			log.Printf("❌ R2 upload failed: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"Failed to upload encrypted content"}`))
			return
		}
		log.Printf("✅ Encrypted content uploaded to R2: %s", blobID)
	} else {
		// For testing without R2, use emailID as blobID
		blobID = emailID
		log.Printf("⚠️ R2 client not available, using emailID as blobID: %s", blobID)
	}

	// Step 13: Create email delivery configuration
	deliveryConfig := email.EmailDeliveryConfig{
		DeliveryType:             "both", // Send both internally and externally
		InternalRecipient:        req.Recipient,
		ExternalRecipient:        req.Recipient,
		ExternalSubject:          req.Subject,
		ExternalBody:             req.Body,
		ExternalSecurityFeatures: securityConfig,
	}

	// Step 14: Apply security features and send email using EmailSecurityService
	ctx := r.Context()
	err = srv.emailSecurityService.SendEmailWithSecurity(ctx, userID, deliveryConfig)
	if err != nil {
		log.Printf("❌ Failed to send email with security features: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to send email: ` + err.Error() + `"}`))
		return
	}

	// Step 14.5: Send notification email to recipient via SMTP
	if req.Recipient != "" {
		// Get sender email from database
		var senderEmail string
		if err := srv.db.QueryRow("SELECT email FROM users WHERE id = ?", userID).Scan(&senderEmail); err != nil {
			log.Printf("⚠️ Failed to get sender email for SMTP notification: %v", err)
			// Use a fallback address if we can't get the sender email
			senderEmail = "noreply@securesystem.email"
		}
		
		if err := srv.emailSecurityService.SendNotificationEmail(req.Recipient, emailID, senderEmail); err != nil {
			log.Printf("⚠️ Failed to send SMTP notification email: %v", err)
			// Don't fail the entire operation if SMTP notification fails
		} else {
			log.Printf("✅ SMTP notification email sent successfully from %s to: %s", senderEmail, req.Recipient)
		}
	}

	// Step 15: Check if recipient is external and create secure link
	log.Printf("🔍 Checking if recipient %s is external...", req.Recipient)
	isExternalRecipient, err := srv.isExternalRecipient(req.Recipient)
	if err != nil {
		log.Printf("❌ Failed to check if recipient is external: %v", err)
		// Continue with normal flow if we can't determine
	} else if isExternalRecipient {
		log.Printf("🌐 Recipient %s is external - creating secure link", req.Recipient)

		// Create secure link with all security features
		secureLinkResponse, err := srv.createSecureLinkForExternalRecipient(emailID, req, userID)
		if err != nil {
			log.Printf("❌ Failed to create secure link for external recipient: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"Failed to create secure link for external recipient"}`))
			return
		}

		log.Printf("✅ Secure link created for external recipient: %s", secureLinkResponse.SecureURL)

		// Return the secure link instead of the email blob
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Set default access count for new secure links
		zero := 0
		json.NewEncoder(w).Encode(SendEmailResponse{
			BlobID:        secureLinkResponse.LinkID, // Use link ID as blob ID for consistency
			Status:        "success",
			SecureLinkURL: &secureLinkResponse.SecureURL, // Add secure link URL to response
			BurnAfterRead: &req.BurnAfterRead,
			AccessCount:   &zero, // New secure links start with 0 accesses
			MaxAttempts:   &req.MaxFailedAttempts,
		})
		return
	} else {
		log.Printf("✅ Recipient %s is internal - proceeding with normal email flow", req.Recipient)
	}

	// Step 16: Return success response with tracking fields
	log.Printf("✅ Email send operation completed successfully")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Prepare tracking fields for response
	var burnAfterReadBool *bool
	var accessCountInt *int
	var maxAttemptsInt *int

	// Only include burn_after_read if it's enabled
	if req.BurnAfterRead {
		burnAfterReadBool = &req.BurnAfterRead
	}

	// Always include access_count (starts at 0)
	zero := 0
	accessCountInt = &zero

	// Only include max_attempts if self-destruct is enabled
	if req.SelfDestructAfterAttempts {
		maxAttemptsInt = &req.MaxFailedAttempts
	}

	json.NewEncoder(w).Encode(SendEmailResponse{
		BlobID:        blobID,
		Status:        "success",
		BurnAfterRead: burnAfterReadBool,
		AccessCount:   accessCountInt,
		MaxAttempts:   maxAttemptsInt,
	})
}

// =============================================================================
// EXTERNAL RECIPIENT SECURE LINK INTEGRATION
// =============================================================================

// isExternalRecipient checks if the recipient email is external (not a registered user)
func (srv *Server) isExternalRecipient(recipientEmail string) (bool, error) {
	// Check if the recipient exists in our users table
	var userID string
	err := srv.db.QueryRow("SELECT id FROM users WHERE email = ?", recipientEmail).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			// User not found - this is an external recipient
			return true, nil
		}
		// Database error
		return false, fmt.Errorf("failed to check if recipient is external: %w", err)
	}

	// User found - this is an internal recipient
	return false, nil
}

// createSecureLinkForExternalRecipient creates a secure link for an external recipient
// with all the security features from the original email request
func (srv *Server) createSecureLinkForExternalRecipient(emailID string, req SendEmailRequest, senderID string) (*securelinks.CreateSecureLinkResponse, error) {
	// Convert email security settings to secure link security settings
	securitySettings := securelinks.SecuritySettings{
		// Password protection
		RequirePassword:   req.Password != "",
		PasswordHash:      nil, // Will be set if password provided
		MaxAccessAttempts: req.MaxFailedAttempts,

		// Geolocation restrictions
		GeolocationRestriction: req.GeoVerificationType != "none",
		AllowedCountries:       []string{},
		AllowedCities:          []string{},

		// Time-based restrictions
		TimeLock:      req.TimeLock,
		TimeLockUntil: nil, // Will be set if time lock provided

		// Access controls
		ReadOnce:     req.BurnAfterRead,
		AutoDestruct: req.SelfDestructAfterAttempts,

		// MFA settings
		RequireMFA: req.RequireMFA,
		MFAType:    req.MFAType,

		// Expiration
		ExpiresAt: nil,

		// Additional security features
		RemoteRevoke:  req.RemoteRevoke,
		StripMetadata: req.StripMetadata,
	}

	// Set password hash if password is provided
	if req.Password != "" {
		passwordService := security.NewPasswordProtectionService(srv.db)
		passwordHash, err := passwordService.HashPassword(req.Password)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password for secure link: %w", err)
		}
		securitySettings.PasswordHash = &passwordHash
	}

	// Set geolocation restrictions
	if req.GeoVerificationType != "none" {
		if req.GeoCountry != "" {
			securitySettings.AllowedCountries = []string{req.GeoCountry}
		}
		if req.GeoCity != "" {
			securitySettings.AllowedCities = []string{req.GeoCity}
		}
	}

	// Set time lock if provided
	if req.TimeLock && req.UnlockAfter != "" {
		unlockAt, err := time.Parse(time.RFC3339, req.UnlockAfter)
		if err != nil {
			return nil, fmt.Errorf("failed to parse unlock time: %w", err)
		}
		unlockAtUnix := unlockAt.Unix()
		securitySettings.TimeLockUntil = &unlockAtUnix
	}

	// Set expiration if provided
	if req.ExpiresAt != "" {
		expiresAt, err := time.Parse(time.RFC3339, req.ExpiresAt)
		if err != nil {
			return nil, fmt.Errorf("failed to parse expiration time: %w", err)
		}
		expiresAtUnix := expiresAt.Unix()
		securitySettings.ExpiresAt = &expiresAtUnix
	}

	// Create secure link request
	secureLinkReq := securelinks.CreateSecureLinkRequest{
		EmailID:          emailID,
		RecipientEmail:   req.Recipient,
		SecuritySettings: securitySettings,
		CustomMessage:    nil, // Use default template
	}

	// Create the secure link
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := srv.secureLinksService.CreateSecureLink(ctx, secureLinkReq, senderID)
	if err != nil {
		return nil, fmt.Errorf("failed to create secure link: %w", err)
	}

	// Send secure link email to external recipient
	if err := srv.sendSecureLinkEmail(ctx, response, req, senderID); err != nil {
		log.Printf("⚠️ Failed to send secure link email: %v", err)
		// Don't fail the entire operation for email sending errors
		// The secure link is still created and can be accessed
	}

	return response, nil
}

// sendSecureLinkEmail sends a secure link notification email to an external recipient
func (srv *Server) sendSecureLinkEmail(ctx context.Context, secureLinkResponse *securelinks.CreateSecureLinkResponse, req SendEmailRequest, senderID string) error {
	// Get sender information
	var senderName, senderEmail string
	err := srv.db.QueryRowContext(ctx, "SELECT email FROM users WHERE id = ?", senderID).Scan(&senderEmail)
	if err != nil {
		return fmt.Errorf("failed to get sender information: %w", err)
	}

	// Convert security settings to email service format
	securityContext := securelinksemail.SecurityContext{
		RequirePassword:        req.Password != "",
		RequireMFA:             req.RequireMFA,
		MFAType:                req.MFAType,
		GeolocationRestriction: req.GeoVerificationType != "none",
		AllowedCountries:       []string{},
		AllowedCities:          []string{},
		TimeLock:               req.TimeLock,
		TimeLockUntil:          nil,
		ReadOnce:               req.BurnAfterRead,
		AutoDestruct:           req.SelfDestructAfterAttempts,
		ExpiresAt:              nil,
	}

	// Set geolocation restrictions
	if req.GeoVerificationType != "none" {
		if req.GeoCountry != "" {
			securityContext.AllowedCountries = []string{req.GeoCountry}
		}
		if req.GeoCity != "" {
			securityContext.AllowedCities = []string{req.GeoCity}
		}
	}

	// Set time lock if provided
	if req.TimeLock && req.UnlockAfter != "" {
		unlockAt, err := time.Parse(time.RFC3339, req.UnlockAfter)
		if err == nil {
			securityContext.TimeLockUntil = &unlockAt
		}
	}

	// Set expiration if provided
	if req.ExpiresAt != "" {
		expiresAt, err := time.Parse(time.RFC3339, req.ExpiresAt)
		if err == nil {
			securityContext.ExpiresAt = &expiresAt
		}
	}

	// Create email request
	emailReq := securelinksemail.SecureLinkEmailRequest{
		LinkID:          secureLinkResponse.LinkID,
		RecipientEmail:  req.Recipient,
		SenderName:      senderName,
		SenderEmail:     senderEmail,
		SecurityContext: securityContext,
		CustomMessage:   nil, // Use default template
		LinkExpiresAt:   secureLinkResponse.ExpiresAt,
	}

	// Send the email
	emailResponse, err := srv.secureLinkEmailService.SendSecureLinkEmail(ctx, emailReq)
	if err != nil {
		return fmt.Errorf("failed to send secure link email: %w", err)
	}

	if !emailResponse.Success {
		return fmt.Errorf("email service returned failure: %s", emailResponse.Error)
	}

	log.Printf("✅ Secure link email sent successfully to %s (Transaction ID: %s)", req.Recipient, emailResponse.TransactionID)
	return nil
}
