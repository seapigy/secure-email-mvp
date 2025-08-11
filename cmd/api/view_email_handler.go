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
	"strings"
	"time"

	"secure-email-mvp/pkg/auth"
	"secure-email-mvp/pkg/bruteforce"
	"secure-email-mvp/pkg/emailpassword"
	"secure-email-mvp/pkg/geolocation"
	"secure-email-mvp/pkg/geoverify"
	"secure-email-mvp/pkg/iptracking"
	"secure-email-mvp/pkg/mfa"
	"secure-email-mvp/pkg/notification"
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
		allowedCity, allowedCountry                                    string
		geoVerificationType, geoCity, geoCountry                       string // NEW for Micro-Iteration 4.15
		requireMFA                                                     int
		mfaType                                                        string
		encryptedTOTPSecret                                            string
		mfaFailedAttempts                                              int
		mfaLockedUntil                                                 *time.Time
		isPasswordProtected                                            int
	)

	err := srv.db.QueryRow(`
		SELECT encrypted_blob_url, encrypted_key, encryption_nonce, encryption_auth_tag, 
		       compression_algo, sender_id, recipient, subject, created_at,
		       self_destruct_after_attempts, max_attempts, burn_after_read, access_count, expires_at,
		       self_destructed, failed_access_attempts, allowed_city, allowed_country,
		       geo_verification_type, geo_city, geo_country,
		       require_mfa, mfa_type, encrypted_totp_secret, mfa_failed_attempts, mfa_locked_until,
		       is_password_protected
		FROM emails WHERE email_id = ?`,
		emailID,
	).Scan(&blobID, &encryptedKeyB64, &nonceB64, &authTagB64, &compressionAlgo,
		&senderID, &recipient, &subject, &createdAt, &selfDestructAfterAttempts,
		&maxFailedAttempts, &burnAfterRead, &accessCount, &expiresAt, &selfDestructed, &failedAccessAttempts,
		&allowedCity, &allowedCountry, &geoVerificationType, &geoCity, &geoCountry,
		&requireMFA, &mfaType, &encryptedTOTPSecret, &mfaFailedAttempts, &mfaLockedUntil,
		&isPasswordProtected)

	if err != nil {
		log.Printf("Database query failed: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"Email not found"}`))
		return
	}

	// 2. Initialize brute-force protection (Micro-Iteration 4.12)
	bfProtection := bruteforce.NewBruteForceProtection(srv.db)

	// 3. Initialize IP tracking (Micro-Iteration 4.13)
	ipTracking := iptracking.NewIPTrackingService(srv.db)

	// 4. Check if the authenticated user is the sender of this email
	if senderID != userID {
		log.Printf("Unauthorized access attempt: user %s trying to access email from sender %s", userID, senderID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error":"Access denied"}`))
		return
	}

	// 5. Check IP-based lockout (Micro-Iteration 4.13)
	clientIP := getClientIP(r)
	ipLocked, err := ipTracking.CheckIPLockout(clientIP)
	if err != nil {
		log.Printf("Failed to check IP lockout for IP %s: %v", clientIP, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Access denied"}`))
		return
	}
	if ipLocked {
		log.Printf("IP %s is locked out due to repeated failed attempts", clientIP)

		// Record blocked access event
		if err := srv.recordAccessEvent(r.Context(), emailID, senderID, notification.AccessEventTypeBlocked, r, "IP lockout"); err != nil {
			log.Printf("Failed to record blocked access event: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error":"Access denied"}`))
		return
	}

	// 6. Check geolocation restrictions (Micro-Iteration 4.10)
	if allowedCity != "" || allowedCountry != "" {
		// Get client IP address
		clientIP := getClientIP(r)

		// Get geolocation service
		geoService := geolocation.NewGeolocationService()

		// Get location for the client IP
		location, err := geoService.GetLocationByIP(clientIP)
		if err != nil {
			log.Printf("Failed to get geolocation for IP %s: %v", clientIP, err)
			// If geolocation fails, deny access for security
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"error":"Access denied"}`))
			return
		}

		// Check geolocation restrictions with exact matching
		accessAllowed := true

		// Check country restriction if set
		if allowedCountry != "" {
			if strings.ToLower(strings.TrimSpace(location.Country)) != strings.ToLower(strings.TrimSpace(allowedCountry)) {
				accessAllowed = false
				log.Printf("Country mismatch for IP %s: expected %s, got %s", clientIP, allowedCountry, location.Country)
			}
		}

		// Check city restriction if set
		if allowedCity != "" {
			normalizedClientCity := geolocation.NormalizeCityName(location.City)
			normalizedAllowedCity := geolocation.NormalizeCityName(allowedCity)
			if normalizedClientCity != normalizedAllowedCity {
				accessAllowed = false
				log.Printf("City mismatch for IP %s: expected %s, got %s", clientIP, allowedCity, location.City)
			}
		}

		if !accessAllowed {
			log.Printf("Geolocation access blocked for IP %s (%s, %s)", clientIP, location.City, strings.ToUpper(location.Country))

			// Increment brute-force protection failed attempts
			if err := bfProtection.IncrementFailedAttempt(emailID); err != nil {
				log.Printf("Failed to increment brute-force attempts for geolocation failure: %v", err)
			}

			// Increment IP tracking failed attempts
			if err := ipTracking.IncrementFailedAttempt(clientIP); err != nil {
				log.Printf("Failed to increment IP attempts for geolocation failure: %v", err)
			}

			// Record blocked access event
			if err := srv.recordAccessEvent(r.Context(), emailID, senderID, notification.AccessEventTypeBlocked, r, "Geolocation restriction"); err != nil {
				log.Printf("Failed to record blocked access event: %v", err)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"error":"Access denied"}`))
			return
		}

		log.Printf("Geolocation check passed for IP %s (%s, %s)", clientIP, location.City, strings.ToUpper(location.Country))
	}

	// 6. Check enhanced geolocation verification (Micro-Iteration 4.15)
	if geoVerificationType != "" && geoVerificationType != "none" {
		// Get geolocation service
		geoService := geolocation.NewGeolocationService()

		// Get location for the client IP
		location, err := geoService.GetLocationByIP(clientIP)
		if err != nil {
			log.Printf("Failed to get geolocation for enhanced verification for IP %s: %v", clientIP, err)
			// If geolocation fails, deny access for security
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"error":"Access denied"}`))
			return
		}

		// Use the geolocation verifier to check access
		geoVerifier := geoverify.NewGeolocationVerifier()
		verificationResult := geoVerifier.VerifyLocation(
			geoverify.VerificationType(geoVerificationType),
			location,
			geoCity,
			geoCountry,
		)

		if !verificationResult.Allowed {
			log.Printf("Enhanced geolocation verification failed for IP %s (%s, %s): %s",
				clientIP, location.City, strings.ToUpper(location.Country), verificationResult.Reason)

			// Increment brute-force protection failed attempts
			if err := bfProtection.IncrementFailedAttempt(emailID); err != nil {
				log.Printf("Failed to increment brute-force attempts for enhanced geolocation failure: %v", err)
			}

			// Increment IP tracking failed attempts
			if err := ipTracking.IncrementFailedAttempt(clientIP); err != nil {
				log.Printf("Failed to increment IP attempts for enhanced geolocation failure: %v", err)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"error":"Access denied"}`))
			return
		}

		log.Printf("Enhanced geolocation verification passed for IP %s (%s, %s)",
			clientIP, location.City, strings.ToUpper(location.Country))
	}

	// 7. Check brute-force lockout
	locked, err := bfProtection.CheckLockout(emailID)
	if err != nil {
		log.Printf("Failed to check brute-force lockout for email %s: %v", emailID, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Access denied"}`))
		return
	}
	if locked {
		log.Printf("Email %s is locked out due to brute-force protection", emailID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error":"Access denied"}`))
		return
	}

	// 8. Check password protection (Micro-Iteration 4.14)
	if isPasswordProtected == 1 {
		// Check if password is provided in the request body
		var requestBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			log.Printf("Password required but no request body provided for email %s", emailID)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "Password required",
				"code":  "password_required",
			})
			return
		}

		password, ok := requestBody["password"].(string)
		if !ok || password == "" {
			log.Printf("Password required but no password provided for email %s", emailID)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "Password required",
				"code":  "password_required",
			})
			return
		}

		// Validate password
		emailPasswordService := emailpassword.NewEmailPasswordService(srv.db)
		valid, err := emailPasswordService.CheckEmailPassword(emailID, password)
		if err != nil {
			log.Printf("Password validation error: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"Password validation failed"}`))
			return
		}

		if !valid {
			// Increment brute-force protection failed attempts
			if err := bfProtection.IncrementFailedAttempt(emailID); err != nil {
				log.Printf("Failed to increment brute-force attempts for password failure: %v", err)
			}

			// Increment IP tracking failed attempts
			if err := ipTracking.IncrementFailedAttempt(clientIP); err != nil {
				log.Printf("Failed to increment IP attempts for password failure: %v", err)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":"Access denied"}`))
			return
		}

		// Reset failed attempts on successful password validation
		if err := bfProtection.ResetFailedAttempts(emailID); err != nil {
			log.Printf("Failed to reset brute-force attempts: %v", err)
		}

		if err := ipTracking.ResetFailedAttempts(clientIP); err != nil {
			log.Printf("Failed to reset IP attempts: %v", err)
		}

		log.Printf("Password validation successful for email %s", emailID)
	}

	// 9. Check MFA requirements
	if requireMFA == 1 {
		// Check if MFA is locked due to too many failed attempts
		if mfaLockedUntil != nil && time.Now().Before(*mfaLockedUntil) {
			log.Printf("MFA is locked for email %s until %v", emailID, *mfaLockedUntil)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "MFA is temporarily locked due to too many failed attempts",
				"code":  "mfa_locked",
			})
			return
		}

		// Check if MFA code is provided in the request
		mfaCode := r.URL.Query().Get("mfa_code")
		if mfaCode == "" {
			log.Printf("MFA required but no code provided for email %s", emailID)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":    "MFA code required",
				"code":     "mfa_required",
				"mfa_type": mfaType,
			})
			return
		}

		// Validate MFA code
		mfaService := mfa.NewMFAService(srv.db)
		var valid bool
		var err error

		switch mfaType {
		case "TOTP":
			valid, err = mfaService.ValidateTOTP(emailID, mfaCode)
		case "EMAIL_CODE":
			valid, err = mfaService.ValidateEmailCode(emailID, mfaCode)
		default:
			log.Printf("Invalid MFA type: %s", mfaType)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"Invalid MFA configuration"}`))
			return
		}

		if err != nil {
			log.Printf("MFA validation error: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"MFA validation failed"}`))
			return
		}

		if !valid {
			// Increment MFA failed attempts
			if err := mfaService.IncrementFailedAttempts(emailID); err != nil {
				log.Printf("Failed to increment MFA attempts: %v", err)
			}

			// Increment brute-force protection failed attempts
			if err := bfProtection.IncrementFailedAttempt(emailID); err != nil {
				log.Printf("Failed to increment brute-force attempts: %v", err)
			}

			// Increment IP tracking failed attempts
			if err := ipTracking.IncrementFailedAttempt(clientIP); err != nil {
				log.Printf("Failed to increment IP attempts for MFA failure: %v", err)
			}

			// Check if we should lock the account
			locked, _, err := mfaService.CheckMFALockout(emailID)
			if err != nil {
				log.Printf("Failed to check MFA lockout: %v", err)
			}

			if locked {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "Too many failed MFA attempts. Access is now locked.",
					"code":  "mfa_locked",
				})
			} else {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "Invalid MFA code",
					"code":  "invalid_mfa_code",
				})
			}
			return
		}

		// Reset failed attempts on successful validation
		if err := mfaService.ResetFailedAttempts(emailID); err != nil {
			log.Printf("Failed to reset MFA attempts: %v", err)
		}

		// Reset brute-force protection failed attempts on successful validation
		if err := bfProtection.ResetFailedAttempts(emailID); err != nil {
			log.Printf("Failed to reset brute-force attempts: %v", err)
		}

		// Reset IP tracking failed attempts on successful validation
		if err := ipTracking.ResetFailedAttempts(clientIP); err != nil {
			log.Printf("Failed to reset IP attempts: %v", err)
		}

		log.Printf("MFA validation successful for email %s", emailID)
	}

	// 5. Check if email has been self-destructed
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

	// Reset brute-force protection failed attempts on successful access
	if err := bfProtection.ResetFailedAttempts(emailID); err != nil {
		log.Printf("Failed to reset brute-force attempts on successful access: %v", err)
		// Don't fail the request for tracking errors
	}

	// Reset IP tracking failed attempts on successful access
	if err := ipTracking.ResetFailedAttempts(clientIP); err != nil {
		log.Printf("Failed to reset IP attempts on successful access: %v", err)
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

	// Record successful access event
	if err := srv.recordAccessEvent(r.Context(), emailID, senderID, notification.AccessEventTypeSuccess, r, ""); err != nil {
		log.Printf("Failed to record successful access event: %v", err)
		// Continue with response even if notification fails
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

// getClientIP extracts the client's IP address from the request
// Handles various proxy headers and falls back to RemoteAddr
func getClientIP(r *http.Request) string {
	// Check for X-Forwarded-For header (set by proxies)
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		if commaIndex := strings.Index(forwardedFor, ","); commaIndex != -1 {
			return strings.TrimSpace(forwardedFor[:commaIndex])
		}
		return strings.TrimSpace(forwardedFor)
	}

	// Check for X-Real-IP header (set by nginx)
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return strings.TrimSpace(realIP)
	}

	// Check for CF-Connecting-IP header (set by Cloudflare)
	if cfIP := r.Header.Get("CF-Connecting-IP"); cfIP != "" {
		return strings.TrimSpace(cfIP)
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if colonIndex := strings.LastIndex(ip, ":"); colonIndex != -1 {
		ip = ip[:colonIndex]
	}

	return ip
}
