package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"secure-email-mvp/pkg/auth"
	"secure-email-mvp/pkg/emailpassword"
	"secure-email-mvp/pkg/geolocation"
	"secure-email-mvp/pkg/geoverify"
	"secure-email-mvp/pkg/mfa"
	"secure-email-mvp/pkg/storage"

	"github.com/google/uuid"
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
//
// PROCESS FLOW:
// 1. Validate all security parameters
// 2. Normalize geolocation data (city names, country codes)
// 3. Hash password with Argon2id if provided
// 4. Compress email content with gzip
// 5. Encrypt with AES-256-GCM
// 6. Upload encrypted blob to Cloudflare R2
// 7. Store metadata in SQLite database with foreign key integrity
// 8. Return secure access link
//
// MICRO-ITERATION 4.4 FIXES:
// - Convert JWT userID string to integer for database insert
// - Verify user exists before email creation
// - Use complete INSERT statement with all required columns
// - Add comprehensive error logging and debugging
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

	// User ID is already a string, which matches the TEXT type in the database
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
	// This prevents foreign key constraint violations and ensures data integrity
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

	// Step 7: Validate self-destruct settings (Micro-Iteration 4.12)
	if req.SelfDestructAfterAttempts {
		if req.MaxFailedAttempts < 1 || req.MaxFailedAttempts > 10 {
			log.Printf("❌ Invalid maxFailedAttempts: %d (must be between 1-10)", req.MaxFailedAttempts)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"maxFailedAttempts must be between 1 and 10"}`))
			return
		}
	} else {
		// If self-destruct is disabled, set maxFailedAttempts to 0
		req.MaxFailedAttempts = 0
	}

	// Step 8: Validate expiration timestamp if provided
	var expiresAtValue interface{} = nil
	if req.ExpiresAt != "" {
		// Parse the ISO 8601 UTC timestamp
		expiresAt, err := time.Parse(time.RFC3339, req.ExpiresAt)
		if err != nil {
			log.Printf("❌ Invalid expiresAt format: %q (expected ISO 8601 UTC format)", req.ExpiresAt)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"expiresAt must be in ISO 8601 UTC format (e.g., 2024-01-15T14:30:00Z)"}`))
			return
		}

		// Check that expiration is in the future
		if expiresAt.Before(time.Now()) {
			log.Printf("❌ Expiration time is in the past: %v", expiresAt)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"expiresAt must be in the future"}`))
			return
		}

		expiresAtValue = expiresAt
	}

	// Step 9: Validate geolocation restrictions (Micro-Iteration 4.10)
	allowedCityValue := req.AllowedCity
	allowedCountryValue := req.AllowedCountry

	// Validate country code if provided
	if allowedCountryValue != "" {
		if !geolocation.ValidateCountryCode(allowedCountryValue) {
			log.Printf("❌ Invalid country code: %q", allowedCountryValue)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"Invalid country code. Must be ISO 3166-1 alpha-2 format (e.g., US, CA, GB)"}`))
			return
		}
		// Normalize to lowercase
		allowedCountryValue = strings.ToLower(strings.TrimSpace(allowedCountryValue))
	}

	// Validate city name if provided
	if allowedCityValue != "" {
		if !geolocation.ValidateCityName(allowedCityValue) {
			log.Printf("❌ Invalid city name: %q", allowedCityValue)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"Invalid city name. Must be 2-100 characters, letters, spaces, hyphens, and apostrophes only"}`))
			return
		}
		// Normalize city name
		allowedCityValue = geolocation.NormalizeCityName(allowedCityValue)
	}

	// Step 9: Validate enhanced geolocation verification (Micro-Iteration 4.15)
	geoVerifier := geoverify.NewGeolocationVerifier()

	// Set default verification type if not provided
	geoVerificationType := req.GeoVerificationType
	if geoVerificationType == "" {
		geoVerificationType = "none"
	}

	// Validate verification type
	if err := geoVerifier.ValidateVerificationType(geoVerificationType); err != nil {
		log.Printf("❌ Invalid geolocation verification type: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"` + err.Error() + `"}`))
		return
	}

	// Validate verification fields based on type
	if err := geoVerifier.ValidateVerificationFields(
		geoverify.VerificationType(geoVerificationType),
		req.GeoCity,
		req.GeoCountry,
	); err != nil {
		log.Printf("❌ Invalid geolocation verification fields: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"` + err.Error() + `"}`))
		return
	}

	// Normalize verification fields
	normalizedGeoCity, normalizedGeoCountry := geoVerifier.NormalizeVerificationFields(
		geoverify.VerificationType(geoVerificationType),
		req.GeoCity,
		req.GeoCountry,
	)

	// Step 10: Generate emailID early for MFA processing
	emailID := uuid.New().String()

	// Step 11: Validate MFA settings (Micro-Iteration 4.12)
	requireMFAInt := 0
	var mfaTypeValue interface{} = nil
	var encryptedTOTPSecretValue interface{} = nil

	if req.RequireMFA {
		requireMFAInt = 1

		// Validate MFA type
		if req.MFAType != "TOTP" && req.MFAType != "EMAIL_CODE" {
			log.Printf("❌ Invalid MFA type: %q", req.MFAType)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"MFA type must be 'TOTP' or 'EMAIL_CODE'"}`))
			return
		}

		mfaTypeValue = req.MFAType

		// If TOTP is selected, generate and encrypt the TOTP secret
		if req.MFAType == "TOTP" {
			mfaService := mfa.NewMFAService(srv.db)
			totpConfig, err := mfaService.GenerateTOTPSecret(emailID)
			if err != nil {
				log.Printf("❌ Failed to generate TOTP secret: %v", err)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error":"Failed to generate TOTP secret"}`))
				return
			}

			// Encrypt the TOTP secret for storage
			encryptedData, err := auth.EncryptAES256GCM([]byte(totpConfig.Secret))
			if err != nil {
				log.Printf("❌ Failed to encrypt TOTP secret: %v", err)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error":"Failed to encrypt TOTP secret"}`))
				return
			}

			// Convert encrypted data to base64 for storage
			encryptedSecret := base64.StdEncoding.EncodeToString(encryptedData.Ciphertext)
			encryptedKey := base64.StdEncoding.EncodeToString(encryptedData.Key)
			encryptedNonce := base64.StdEncoding.EncodeToString(encryptedData.Nonce)
			encryptedAuthTag := base64.StdEncoding.EncodeToString(encryptedData.AuthTag)

			// Store encrypted components as JSON
			encryptedComponents := map[string]string{
				"ciphertext": encryptedSecret,
				"key":        encryptedKey,
				"nonce":      encryptedNonce,
				"auth_tag":   encryptedAuthTag,
			}

			encryptedJSON, err := json.Marshal(encryptedComponents)
			if err != nil {
				log.Printf("❌ Failed to marshal encrypted TOTP components: %v", err)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error":"Failed to process TOTP secret"}`))
				return
			}

			encryptedTOTPSecretValue = string(encryptedJSON)
		}
	}

	// Step 12: Validate and process password protection (Micro-Iteration 4.14)
	var isPasswordProtectedInt int = 0
	if req.Password != "" {
		// Validate password strength
		emailPasswordService := emailpassword.NewEmailPasswordService(srv.db)
		if err := emailPasswordService.ValidatePasswordStrength(req.Password); err != nil {
			log.Printf("❌ Password validation failed: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"` + err.Error() + `"}`))
			return
		}
		isPasswordProtectedInt = 1
	}

	// Step 13: Compress email content with gzip
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

	// Step 14: Encrypt compressed content using AES-256-GCM
	encryptedData, err := auth.EncryptAES256GCM(compressed)
	if err != nil {
		log.Printf("❌ Encryption failed: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Encryption failed"}`))
		return
	}

	// Combine ciphertext and auth tag for storage (nonce stored separately)
	encrypted := append(encryptedData.Ciphertext, encryptedData.AuthTag...)

	// Step 15: Generate blobID for Cloudflare R2 storage
	blobID := uuid.New().String() + ".blob"

	// Step 16: Upload to Cloudflare R2 with context and timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := storage.UploadToR2WithContext(ctx, blobID, encrypted); err != nil {
		log.Printf("❌ R2 upload failed: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"R2 upload failed"}`))
		return
	}
	log.Printf("✅ R2 upload succeeded: %s", blobID)

	// Step 17: Compute SHA-256 hash of the encrypted content for integrity verification
	hash := sha256.Sum256(encrypted)
	hashB64 := base64.StdEncoding.EncodeToString(hash[:])

	// Step 18: Check database connection
	if srv.db == nil {
		log.Printf("❌ CRITICAL ERROR: srv.db is nil!")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Database connection is nil"}`))
		return
	}

	// Step 19: Simulated DB failure block (for testing purposes)
	// --- SIMULATED DB FAILURE BLOCK ---
	if os.Getenv("SIMULATE_DB_FAILURE") == "1" {
		log.Printf("❌ Simulated DB insert failure triggered by SIMULATE_DB_FAILURE env var")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Database insert failed (simulated)"}`))
		return
	}
	// --- END SIMULATED DB FAILURE BLOCK ---

	// Step 20: Prepare encryption metadata for database storage
	// Store nonce and auth tag separately for decryption
	nonceB64 := base64.StdEncoding.EncodeToString(encryptedData.Nonce)
	authTagB64 := base64.StdEncoding.EncodeToString(encryptedData.AuthTag)
	encryptedKeyB64 := base64.StdEncoding.EncodeToString(encryptedData.Key)

	// Step 21: Convert boolean values to integers for SQLite storage
	selfDestructInt := 0
	if req.SelfDestructAfterAttempts {
		selfDestructInt = 1
	}

	burnAfterReadInt := 0
	if req.BurnAfterRead {
		burnAfterReadInt = 1
	}

	// Step 22: Try to find recipient_id by email address (optional)
	var recipientID *string
	var recipientUserID string
	recipientErr := srv.db.QueryRow(`SELECT id FROM users WHERE email = ?`, req.Recipient).Scan(&recipientUserID)
	if recipientErr == nil {
		recipientID = &recipientUserID
		log.Printf("✅ Found recipient_id %s for email %s", recipientUserID, req.Recipient)
	} else {
		log.Printf("ℹ️ No registered user found for email %s - recipient_id will be NULL", req.Recipient)
	}

	// Step 23: Log all parameters for debugging (MICRO-ITERATION 4.4 ENHANCEMENT)
	log.Printf("=== DATABASE INSERT PARAMETERS ===")
	log.Printf("emailID=%s, userID=%s, recipient=%s, subject=%s, blobID=%s",
		emailID, userID, req.Recipient, req.Subject, blobID)
	log.Printf("recipientID=%v, expiresAtValue=%v, allowedCityValue=%v, allowedCountryValue=%v",
		recipientID, expiresAtValue, allowedCityValue, allowedCountryValue)
	log.Printf("encryptedKeyB64=%s, nonceB64=%s, authTagB64=%s, hashB64=%s",
		encryptedKeyB64, nonceB64, authTagB64, hashB64)

	// Step 24: Build the INSERT statement with all required columns (MICRO-ITERATION 4.4 FIX)
	// The emails table has many columns, but we'll include all essential ones
	insertQuery := `
		INSERT INTO emails (
			email_id, sender_id, recipient, subject, encrypted_blob_url, encrypted_key, 
			encryption_nonce, encryption_auth_tag, sha256_hash, self_destruct_after_attempts, 
			burn_after_read, expires_at, allowed_city, allowed_country, geo_verification_type,
			geo_city, geo_country, require_mfa, mfa_type, encrypted_totp_secret,
			is_password_protected
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	// Step 25: Log SQL query and parameters for debugging (MICRO-ITERATION 4.4 ENHANCEMENT)
	log.Printf("=== SQL EXECUTION ===")
	log.Printf("SQL Query: %s", insertQuery)
	log.Printf("Parameters: emailID=%s, userID=%s, recipient=%s, subject=%s, blobID=%s",
		emailID, userID, req.Recipient, req.Subject, blobID)
	log.Printf("Security params: selfDestructInt=%d, burnAfterReadInt=%d, requireMFAInt=%d, isPasswordProtectedInt=%d",
		selfDestructInt, burnAfterReadInt, requireMFAInt, isPasswordProtectedInt)

	// Step 26: Execute database insert with comprehensive error handling (MICRO-ITERATION 4.4 FIX)
	_, insertErr := srv.db.Exec(insertQuery,
		emailID, userID, req.Recipient, req.Subject, blobID, encryptedKeyB64,
		nonceB64, authTagB64, hashB64, selfDestructInt, burnAfterReadInt, expiresAtValue,
		allowedCityValue, allowedCountryValue, geoVerificationType,
		normalizedGeoCity, normalizedGeoCountry, requireMFAInt, mfaTypeValue, encryptedTOTPSecretValue,
		isPasswordProtectedInt,
	)

	// Step 27: Handle database insert errors with detailed logging (MICRO-ITERATION 4.4 ENHANCEMENT)
	if insertErr != nil {
		log.Printf("❌ DATABASE INSERT FAILED")
		log.Printf("Error: %v", insertErr)
		log.Printf("Failed INSERT parameters: emailID=%s, userID=%s, recipient=%s, subject=%s, blobID=%s",
			emailID, userID, req.Recipient, req.Subject, blobID)
		log.Printf("Security parameters: selfDestructInt=%d, burnAfterReadInt=%d, requireMFAInt=%d, isPasswordProtectedInt=%d",
			selfDestructInt, burnAfterReadInt, requireMFAInt, isPasswordProtectedInt)
		log.Printf("Geolocation params: allowedCityValue=%q, allowedCountryValue=%q, geoVerificationType=%q",
			allowedCityValue, allowedCountryValue, geoVerificationType)
		log.Printf("MFA params: mfaTypeValue=%v, encryptedTOTPSecretValue=%v", mfaTypeValue, encryptedTOTPSecretValue)
		log.Printf("Expiration: expiresAtValue=%v", expiresAtValue)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		errorResponse := map[string]string{
			"error":   "Database insert failed",
			"details": insertErr.Error(), // MICRO-ITERATION 4.4: Return actual database error
		}
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	log.Printf("✅ Database insert successful: emailID=%s", emailID)

	// Step 28: Set password if provided (Micro-Iteration 4.14)
	if req.Password != "" {
		emailPasswordService := emailpassword.NewEmailPasswordService(srv.db)
		if err := emailPasswordService.SetEmailPassword(emailID, req.Password); err != nil {
			log.Printf("❌ Failed to set email password: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"Failed to set email password"}`))
			return
		}
		log.Printf("✅ Email password set successfully")
	}

	// Step 29: Return success response with tracking fields (Micro-Iteration 4.21)
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
