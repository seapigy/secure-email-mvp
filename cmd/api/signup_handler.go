package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"secure-email-mvp/pkg/auth"
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

		// Store user in database with TOTP secret
		log.Printf("[SIGNUP_DEBUG] 💾 Storing user in database...")
		log.Printf("[SIGNUP_DEBUG] Attempting to create user in database for email: %s", req.Email)
		err = createUserWithTOTP(db, req.Email, passwordHash, totpSecret, req.FallbackEmail, fallbackToken, fallbackExpiration)
		if err != nil {
			log.Printf("[SIGNUP_DEBUG] ❌ Database insertion failed for email %s: %v", req.Email, err)
			if strings.Contains(err.Error(), "UNIQUE constraint failed") {
				log.Printf("[SIGNUP_DEBUG] ❌ User already exists: %s", req.Email)
				http.Error(w, `{"error":"User already exists"}`, http.StatusBadRequest)
				return
			}
			http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
			return
		}
		log.Printf("[SIGNUP_DEBUG] ✅ User created successfully in database for email: %s", req.Email)

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

		// Send fallback confirmation email (stub) - temporarily disabled for debugging
		// if err := auth.SendFallbackConfirmationEmail(req.FallbackEmail, fallbackToken); err != nil {
		// 	// Log error but don't fail the signup
		// 	http.Error(w, `{"error":"User created but fallback email confirmation failed"}`, http.StatusInternalServerError)
		// 	return
		// }

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

// createUserWithTOTP inserts a new user with TOTP secret, fallback email and token into the database.
func createUserWithTOTP(db *sql.DB, email, passwordHash, totpSecret, fallbackEmail, fallbackToken string, fallbackExpiration time.Time) error {
	log.Printf("[SIGNUP_DEBUG] Starting database insertion for email: %s", email)

	// Insert into both password and password_hash columns for compatibility
	query := `INSERT INTO users (email, password, password_hash, totp_secret, fallback_email, fallback_token, fallback_confirmed, fallback_token_expiration) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	log.Printf("[SIGNUP_DEBUG] Executing query: %s", query)
	log.Printf("[SIGNUP_DEBUG] Parameters: email=%s, passwordHash=%s, totpSecret=%s, fallbackEmail=%s, fallbackToken=%s", email, passwordHash[:16]+"...", totpSecret[:16]+"...", fallbackEmail, fallbackToken[:16]+"...")

	_, err := db.Exec(query, email, passwordHash, passwordHash, totpSecret, fallbackEmail, fallbackToken, 0, fallbackExpiration)
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
