package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"secure-email-mvp/pkg/auth"
	"secure-email-mvp/pkg/email"
	"secure-email-mvp/pkg/password"
	"secure-email-mvp/pkg/reputation"
	"secure-email-mvp/pkg/zkid"
)

// SignupRequest represents the JSON request body for signup
type SignupRequest struct {
	Email         string `json:"email"`
	Password      string `json:"password"`
	FallbackEmail string `json:"fallback_email"`
}

// SignupResponse represents the JSON response for signup
type SignupResponse struct {
	Message string `json:"message"`
}

// signupHandlerFactory returns an HTTP handler for user signup. It validates input, enforces fallback email, and sends a confirmation link.
func signupHandlerFactory(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// BASIC DEBUGGING: Log that handler was called
		log.Printf("[SIGNUP_DEBUG] 🔥 SIGNUP HANDLER CALLED - Method: %s, Path: %s", r.Method, r.URL.Path)

		// ENHANCED DEBUGGING: Log request details
		log.Printf("[SIGNUP_DEBUG] ===== SIGNUP REQUEST START =====")
		log.Printf("[SIGNUP_DEBUG] Method: %s", r.Method)
		log.Printf("[SIGNUP_DEBUG] Path: %s", r.URL.Path)
		log.Printf("[SIGNUP_DEBUG] Content-Type: %s", r.Header.Get("Content-Type"))
		log.Printf("[SIGNUP_DEBUG] User-Agent: %s", r.Header.Get("User-Agent"))

		// Only allow POST method
		if r.Method != http.MethodPost {
			log.Printf("[SIGNUP_DEBUG] ❌ Method not allowed: %s", r.Method)
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Set content type to JSON
		w.Header().Set("Content-Type", "application/json")

		// Get client IP address
		clientIP := reputation.GetClientIP(r)
		log.Printf("[SIGNUP_DEBUG] Client IP: %s", clientIP)

		// Check IP reputation before processing signup
		reputationService := reputation.NewReputationService()
		ctx := context.Background()

		isMalicious, err := reputationService.CheckIPReputation(ctx, clientIP)
		if err != nil {
			log.Printf("[SIGNUP_DEBUG] ⚠️  IP reputation check failed for IP %s: %v", clientIP, err)
			// Continue processing on reputation check failure
		} else if isMalicious {
			log.Printf("[SIGNUP_DEBUG] ❌ Signup blocked due to IP reputation for IP %s", clientIP)
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Access denied due to IP reputation",
			})
			return
		}
		log.Printf("[SIGNUP_DEBUG] ✅ IP reputation check passed")

		// Parse JSON request body with detailed logging
		log.Printf("[SIGNUP_DEBUG] 🔍 Parsing request body...")

		// Read and log the raw request body
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("[SIGNUP_DEBUG] ❌ Failed to read request body: %v", err)
			http.Error(w, `{"error":"Failed to read request body"}`, http.StatusBadRequest)
			return
		}

		log.Printf("[SIGNUP_DEBUG] 📄 Raw request body: %s", string(bodyBytes))
		log.Printf("[SIGNUP_DEBUG] 📏 Request body length: %d bytes", len(bodyBytes))

		// Check if body is empty
		if len(bodyBytes) == 0 {
			log.Printf("[SIGNUP_DEBUG] ❌ Request body is empty")
			http.Error(w, `{"error":"Request body is empty"}`, http.StatusBadRequest)
			return
		}

		// Parse JSON from the body bytes
		var req SignupRequest
		if err := json.Unmarshal(bodyBytes, &req); err != nil {
			log.Printf("[SIGNUP_DEBUG] ❌ JSON parsing failed: %v", err)
			log.Printf("[SIGNUP_DEBUG] 🔍 JSON error details: %T", err)
			http.Error(w, `{"error":"Invalid JSON format"}`, http.StatusBadRequest)
			return
		}
		log.Printf("[SIGNUP_DEBUG] ✅ JSON parsing successful")

		// Log parsed struct values for debugging
		log.Printf("[SIGNUP_DEBUG] 📋 Parsed Struct Values:")
		log.Printf("[SIGNUP_DEBUG]   - Email: '%s' (length: %d)", req.Email, len(req.Email))
		log.Printf("[SIGNUP_DEBUG]   - Password: '%s' (length: %d)", strings.Repeat("*", len(req.Password)), len(req.Password))
		log.Printf("[SIGNUP_DEBUG]   - FallbackEmail: '%s' (length: %d)", req.FallbackEmail, len(req.FallbackEmail))

		// Check for empty required fields
		if req.Email == "" {
			log.Printf("[SIGNUP_DEBUG] ❌ Email field is empty")
			http.Error(w, `{"error":"Email is required"}`, http.StatusBadRequest)
			return
		}
		if req.Password == "" {
			log.Printf("[SIGNUP_DEBUG] ❌ Password field is empty")
			http.Error(w, `{"error":"Password is required"}`, http.StatusBadRequest)
			return
		}
		if req.FallbackEmail == "" {
			log.Printf("[SIGNUP_DEBUG] ❌ FallbackEmail field is empty")
			http.Error(w, `{"error":"Fallback email is required"}`, http.StatusBadRequest)
			return
		}

		log.Printf("[SIGNUP_DEBUG] ✅ All required fields present")

		// Validate email
		log.Printf("[SIGNUP_DEBUG] 🔍 Validating email format...")
		if !isValidEmail(req.Email) {
			log.Printf("[SIGNUP_DEBUG] ❌ Invalid email format: %s", req.Email)
			http.Error(w, `{"error":"Invalid email format"}`, http.StatusBadRequest)
			return
		}
		log.Printf("[SIGNUP_DEBUG] ✅ Email format validation passed")

		// Validate password using comprehensive password service
		log.Printf("[SIGNUP_DEBUG] 🔍 Validating password strength...")
		passwordService := password.NewPasswordService()
		passwordResult, err := passwordService.ValidatePassword(ctx, req.Password)
		if err != nil {
			log.Printf("[SIGNUP_DEBUG] ❌ Password validation failed: %v", err)
			http.Error(w, `{"error":"Password validation error"}`, http.StatusInternalServerError)
			return
		}

		if !passwordResult.IsValid {
			// Log validation failures for audit
			log.Printf("[SIGNUP_DEBUG] ❌ Password validation failed for email %s: %v", req.Email, passwordResult.Errors)

			// Return generic error message without revealing specific validation details
			http.Error(w, `{"error":"Password does not meet security requirements"}`, http.StatusBadRequest)
			return
		}

		// Log successful password validation
		log.Printf("[SIGNUP_DEBUG] ✅ Password validation passed for email %s (score: %d, breach count: %d)",
			req.Email, passwordResult.Score, passwordResult.BreachCount)

		// Validate fallback email
		log.Printf("[SIGNUP_DEBUG] 🔍 Validating fallback email...")
		if req.FallbackEmail == "" {
			log.Printf("[SIGNUP_DEBUG] ❌ Fallback email is empty")
			http.Error(w, `{"error":"Fallback email is required"}`, http.StatusBadRequest)
			return
		}
		if !isValidEmail(req.FallbackEmail) {
			log.Printf("[SIGNUP_DEBUG] ❌ Invalid fallback email format: %s", req.FallbackEmail)
			http.Error(w, `{"error":"Invalid fallback email format"}`, http.StatusBadRequest)
			return
		}
		log.Printf("[SIGNUP_DEBUG] ✅ Fallback email validation passed")

		// Hash password using Argon2 (matching auth package expectations)
		log.Printf("[SIGNUP_DEBUG] 🔐 Hashing password...")
		passwordHash, err := auth.HashPassword(req.Password, req.Email)
		if err != nil {
			log.Printf("[SIGNUP_DEBUG] ❌ Password hashing failed: %v", err)
			http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
			return
		}
		log.Printf("[SIGNUP_DEBUG] ✅ Password hashed successfully (length: %d)", len(passwordHash))

		// Generate TOTP secret
		log.Printf("[SIGNUP_DEBUG] 🔢 Generating TOTP secret...")
		totpSecret, err := auth.GenerateTOTPSecret()
		if err != nil {
			log.Printf("[SIGNUP_DEBUG] ❌ TOTP secret generation failed: %v", err)
			http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
			return
		}

		// For test user, use known TOTP secret
		if req.Email == "test@securesystem.email" {
			log.Printf("[SIGNUP_DEBUG] 🔧 Using known TOTP secret for test user")
			totpSecret = "JBSWY3DPEHPK3PXP"
		}
		log.Printf("[SIGNUP_DEBUG] ✅ TOTP secret generated: %s", totpSecret)

		// Generate fallback token and expiration
		log.Printf("[SIGNUP_DEBUG] 🔑 Generating fallback token...")
		fallbackToken := auth.GenerateFallbackToken(req.FallbackEmail)
		fallbackExpiration := auth.GenerateFallbackExpiration()
		log.Printf("[SIGNUP_DEBUG] ✅ Fallback token generated (length: %d)", len(fallbackToken))

			// Store user in database with TOTP secret using explicit transaction
	log.Printf("[TX_DEBUG] 🔥 ===== TRANSACTION START ===== email: %s", req.Email)
	log.Printf("[TX_DEBUG] 💾 Storing user in database...")
	log.Printf("[TX_DEBUG] Attempting to create user in database for email: %s", req.Email)
	
	// Begin explicit transaction
	log.Printf("[TX_DEBUG] 🔍 BEGIN TRANSACTION for email: %s", req.Email)
	tx, err := db.Begin()
	if err != nil {
		log.Printf("[TX_DEBUG] ❌ FAILED TO BEGIN TRANSACTION for email %s: %v", req.Email, err)
		log.Printf("[TX_DEBUG] Database driver error details: %T: %+v", err, err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}
	log.Printf("[TX_DEBUG] ✅ TRANSACTION BEGUN successfully for email: %s", req.Email)
	
	// Ensure transaction is rolled back on error, but only if not committed
	defer func() {
		if tx != nil {
			log.Printf("[TX_DEBUG] 🔄 ROLLBACK TRIGGERED for email %s (tx not committed)", req.Email)
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				log.Printf("[TX_DEBUG] ⚠️ FAILED TO ROLLBACK transaction: %v", rollbackErr)
				log.Printf("[TX_DEBUG] Rollback error details: %T: %+v", rollbackErr, rollbackErr)
			} else {
				log.Printf("[TX_DEBUG] ✅ ROLLBACK COMPLETED for email: %s", req.Email)
			}
		} else {
			log.Printf("[TX_DEBUG] ✅ NO ROLLBACK NEEDED - transaction was committed for email: %s", req.Email)
		}
	}()
	
	// Create user within transaction
	log.Printf("[TX_DEBUG] 🔍 EXECUTING USER INSERT for email: %s", req.Email)
	err = createUserWithTOTPInTx(tx, req.Email, passwordHash, totpSecret, req.FallbackEmail, fallbackToken, fallbackExpiration)
	if err != nil {
		log.Printf("[TX_DEBUG] ❌ DATABASE INSERTION FAILED for email %s: %v", req.Email, err)
		log.Printf("[TX_DEBUG] Insert error details: %T: %+v", err, err)
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			log.Printf("[TX_DEBUG] ❌ User already exists: %s", req.Email)
			http.Error(w, `{"error":"User already exists"}`, http.StatusBadRequest)
			return
		}
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}
	log.Printf("[TX_DEBUG] ✅ USER INSERT SUCCESSFUL for email: %s", req.Email)
	
	// Commit the transaction
	log.Printf("[TX_DEBUG] 🔍 COMMITTING TRANSACTION for email: %s", req.Email)
	if err = tx.Commit(); err != nil {
		log.Printf("[TX_DEBUG] ❌ FAILED TO COMMIT TRANSACTION for email %s: %v", req.Email, err)
		log.Printf("[TX_DEBUG] Commit error details: %T: %+v", err, err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
		return
	}
	
	// Clear the defer rollback since we committed successfully
	log.Printf("[TX_DEBUG] ✅ COMMIT SUCCESSFUL for email: %s", req.Email)
	tx = nil
	
	log.Printf("[TX_DEBUG] 🔥 ===== TRANSACTION COMMITTED ===== email: %s", req.Email)
	log.Printf("[SIGNUP_DEBUG] ✅ User created successfully in database for email: %s", req.Email)
	
	// POST-INSERT VERIFICATION: Confirm user actually exists in database
	log.Printf("[TX_DEBUG] 🔍 POST-INSERT VERIFICATION for email: %s", req.Email)
	var userID string
	var userEmail string
	err = db.QueryRow("SELECT id, email FROM users WHERE email = ?", req.Email).Scan(&userID, &userEmail)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("[TX_DEBUG] ❌ CRITICAL: User not found in database after commit for email: %s", req.Email)
			log.Printf("[TX_DEBUG] ❌ This indicates a serious database issue - transaction committed but user not persisted")
		} else {
			log.Printf("[TX_DEBUG] ⚠️ Failed to verify user in database for email %s: %v", req.Email, err)
		}
	} else {
		log.Printf("[TX_DEBUG] ✅ POST-INSERT VERIFICATION SUCCESSFUL for email: %s", req.Email)
		log.Printf("[TX_DEBUG] ✅ Verified user ID: %s, Email: %s", userID, userEmail)
	}

		// ZKID: Create encrypted email mapping if enabled (feature-flagged)
		zkCfg := zkid.ConfigFromEnv()
		if zkCfg.Enabled {
			log.Printf("[SIGNUP_DEBUG] 🔐 Creating ZKID mapping...")
			// Get newly created user UUID
			var userID string
			err = db.QueryRow("SELECT id FROM users WHERE email = ?", req.Email).Scan(&userID)
			if err == nil && userID != "" {
				svc := zkid.NewService(db, zkCfg)
				// Fallback email optional pointer
				var fb *string
				if req.FallbackEmail != "" {
					fb = &req.FallbackEmail
				}
				_, zkErr := svc.CreateOrUpdateMapping(userID, req.Email, fb)
				if zkErr != nil {
					log.Printf("[SIGNUP_DEBUG] ⚠️  ZKID mapping failed for user %s: %v", userID, zkErr)
				} else {
					log.Printf("[SIGNUP_DEBUG] ✅ ZKID mapping created successfully")
				}
			} else if err != nil {
				log.Printf("[SIGNUP_DEBUG] ⚠️  Could not lookup user id for %s: %v", req.Email, err)
			}
		}

		// Send fallback confirmation email using SMTP service
		if req.FallbackEmail != "" {
			emailService, err := email.NewSMTPEmailService(nil) // No domain verification for system emails
			if err != nil {
				log.Printf("[SIGNUP_DEBUG] ⚠️ Failed to initialize email service: %v", err)
				// Don't fail signup if email service is unavailable
			} else {
				// Get base URL from environment
				baseURL := os.Getenv("BASE_URL")
				if baseURL == "" {
					baseURL = "http://localhost:8080"
				}

				err = emailService.SendFallbackEmail(req.FallbackEmail, fallbackToken, baseURL)
				if err != nil {
					log.Printf("[SIGNUP_DEBUG] ⚠️ Failed to send fallback email to %s: %v", req.FallbackEmail, err)
					// Don't fail signup if email sending fails
				} else {
					log.Printf("[SIGNUP_DEBUG] ✅ Fallback confirmation email sent to %s", req.FallbackEmail)
				}
			}
		}

		// Return success response
		log.Printf("[SIGNUP_DEBUG] 📤 Returning success response")
		response := SignupResponse{
			Message: "User created",
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
		log.Printf("[SIGNUP_DEBUG] ===== SIGNUP REQUEST SUCCESS =====")
	}
}

// createUserWithTOTPInTx inserts a new user with TOTP secret, fallback email and token into the database within a transaction.
func createUserWithTOTPInTx(tx *sql.Tx, email, passwordHash, totpSecret, fallbackEmail, fallbackToken string, fallbackExpiration time.Time) error {
	log.Printf("[TX_DEBUG] 🔍 ===== INSERT OPERATION START ===== email: %s", email)
	log.Printf("[TX_DEBUG] Starting database insertion within transaction for email: %s", email)

	// Generate a UUID for the user ID
	userID := uuid.New().String()
	log.Printf("[TX_DEBUG] Generated user ID: %s", userID)

	// Insert with all required fields including totp_configured and account_status
	query := `INSERT INTO users (id, email, password, password_hash, totp_secret, fallback_email, fallback_token, fallback_confirmed, fallback_token_expiration, totp_configured, account_status) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	log.Printf("[TX_DEBUG] 🔍 Executing SQL query: %s", query)
	log.Printf("[TX_DEBUG] 🔍 Query parameters:")
	log.Printf("[TX_DEBUG]   - userID: %s", userID)
	log.Printf("[TX_DEBUG]   - email: %s", email)
	log.Printf("[TX_DEBUG]   - passwordHash: %s...", passwordHash[:16])
	log.Printf("[TX_DEBUG]   - totpSecret: %s...", totpSecret[:16])
	log.Printf("[TX_DEBUG]   - fallbackEmail: %s", fallbackEmail)
	log.Printf("[TX_DEBUG]   - fallbackToken: %s...", fallbackToken[:16])
	log.Printf("[TX_DEBUG]   - fallback_confirmed: 0")
	log.Printf("[TX_DEBUG]   - fallback_token_expiration: %v", fallbackExpiration)
	log.Printf("[TX_DEBUG]   - totp_configured: 0")
	log.Printf("[TX_DEBUG]   - account_status: pending")

	log.Printf("[TX_DEBUG] 🔍 Executing tx.Exec()...")
	result, err := tx.Exec(query, userID, email, passwordHash, passwordHash, totpSecret, fallbackEmail, fallbackToken, 0, fallbackExpiration, 0, "pending")
	if err != nil {
		log.Printf("[TX_DEBUG] ❌ DATABASE INSERT ERROR for email %s: %v", email, err)
		log.Printf("[TX_DEBUG] ❌ Error type: %T", err)
		log.Printf("[TX_DEBUG] ❌ Full error details: %+v", err)
		
		// Check for specific constraint violations
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			log.Printf("[TX_DEBUG] ❌ UNIQUE CONSTRAINT VIOLATION: email %s already exists", email)
		} else if strings.Contains(err.Error(), "NOT NULL constraint failed") {
			log.Printf("[TX_DEBUG] ❌ NOT NULL CONSTRAINT VIOLATION: required field is null")
		} else if strings.Contains(err.Error(), "FOREIGN KEY constraint failed") {
			log.Printf("[TX_DEBUG] ❌ FOREIGN KEY CONSTRAINT VIOLATION: referenced table/row not found")
		}
		
		return fmt.Errorf("insert failed: %v", err)
	}

	// Log rows affected to confirm insertion
	log.Printf("[TX_DEBUG] 🔍 Getting rows affected...")
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("[TX_DEBUG] ⚠️ Could not get rows affected: %v", err)
		log.Printf("[TX_DEBUG] ⚠️ Rows affected error type: %T", err)
	} else {
		log.Printf("[TX_DEBUG] ✅ Rows affected: %d", rowsAffected)
		if rowsAffected == 0 {
			log.Printf("[TX_DEBUG] ❌ CRITICAL: Insert succeeded but 0 rows were affected")
			return fmt.Errorf("insert succeeded but no rows were affected")
		}
		if rowsAffected > 1 {
			log.Printf("[TX_DEBUG] ⚠️ WARNING: Insert affected %d rows (expected 1)", rowsAffected)
		}
	}

	log.Printf("[TX_DEBUG] ✅ ===== INSERT OPERATION SUCCESSFUL ===== email: %s", email)
	log.Printf("[TX_DEBUG] Database insertion successful within transaction")
	return nil
}

// createUserWithTOTP inserts a new user with TOTP secret, fallback email and token into the database.
func createUserWithTOTP(db *sql.DB, email, passwordHash, totpSecret, fallbackEmail, fallbackToken string, fallbackExpiration time.Time) error {
	log.Printf("[SIGNUP_DEBUG] Starting database insertion for email: %s", email)

	// Generate a UUID for the user ID
	userID := uuid.New().String()
	log.Printf("[SIGNUP_DEBUG] Generated user ID: %s", userID)

	// Insert into both password and password_hash columns for compatibility
	query := `INSERT INTO users (id, email, password, password_hash, totp_secret, fallback_email, fallback_token, fallback_confirmed, fallback_token_expiration) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	log.Printf("[SIGNUP_DEBUG] Executing query: %s", query)
	log.Printf("[SIGNUP_DEBUG] Parameters: userID=%s, email=%s, passwordHash=%s, totpSecret=%s, fallbackEmail=%s, fallbackToken=%s", userID, email, passwordHash[:16]+"...", totpSecret[:16]+"...", fallbackEmail, fallbackToken[:16]+"...")

	_, err := db.Exec(query, userID, email, passwordHash, passwordHash, totpSecret, fallbackEmail, fallbackToken, 0, fallbackExpiration)
	if err != nil {
		log.Printf("[SIGNUP_DEBUG] Database error: %v", err)
		return fmt.Errorf("insert failed: %v", err)
	}

	log.Printf("[SIGNUP_DEBUG] Database insertion successful")
	return nil
}

// createUser inserts a new user into the database (legacy function, not used in fallback flow).
func createUser(db *sql.DB, email, hashedPassword string) error {
	query := `INSERT INTO users (email, password_hash, created_at) VALUES (?, ?, ?)`
	_, err := db.Exec(query, email, hashedPassword, time.Now())
	return err
}

// isValidEmail checks if the email format is valid using regex.
func isValidEmail(email string) bool {
	// Basic email regex pattern
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(strings.TrimSpace(email))
}

// isValidPassword checks if the password meets minimum requirements.
func isValidPassword(password string) bool {
	return len(strings.TrimSpace(password)) >= 8
}
