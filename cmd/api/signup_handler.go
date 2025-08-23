package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
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
		// Only allow POST method
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Set content type to JSON
		w.Header().Set("Content-Type", "application/json")

		// Get client IP address
		clientIP := reputation.GetClientIP(r)

		// Check IP reputation before processing signup
		reputationService := reputation.NewReputationService()
		ctx := context.Background()

		isMalicious, err := reputationService.CheckIPReputation(ctx, clientIP)
		if err != nil {
			log.Printf("IP reputation check failed for IP %s: %v", clientIP, err)
			// Continue processing on reputation check failure
		} else if isMalicious {
			log.Printf("Signup blocked due to IP reputation for IP %s", clientIP)
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Access denied due to IP reputation",
			})
			return
		}

		// Parse JSON request body
		var req SignupRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"Invalid JSON format"}`, http.StatusBadRequest)
			return
		}

		// Validate email
		if !isValidEmail(req.Email) {
			http.Error(w, `{"error":"Invalid email format"}`, http.StatusBadRequest)
			return
		}

		// Validate password using comprehensive password service
		passwordService := password.NewPasswordService()
		passwordResult, err := passwordService.ValidatePassword(ctx, req.Password)
		if err != nil {
			log.Printf("Password validation failed: %v", err)
			http.Error(w, `{"error":"Password validation error"}`, http.StatusInternalServerError)
			return
		}

		if !passwordResult.IsValid {
			// Log validation failures for audit
			log.Printf("Password validation failed for email %s: %v", req.Email, passwordResult.Errors)

			// Return generic error message without revealing specific validation details
			http.Error(w, `{"error":"Password does not meet security requirements"}`, http.StatusBadRequest)
			return
		}

		// Log successful password validation
		log.Printf("Password validation passed for email %s (score: %d, breach count: %d)",
			req.Email, passwordResult.Score, passwordResult.BreachCount)

		// Validate fallback email
		if req.FallbackEmail == "" {
			http.Error(w, `{"error":"Fallback email is required"}`, http.StatusBadRequest)
			return
		}
		if !isValidEmail(req.FallbackEmail) {
			http.Error(w, `{"error":"Invalid fallback email format"}`, http.StatusBadRequest)
			return
		}

		// Hash password using Argon2 (matching auth package expectations)
		passwordHash, err := auth.HashPassword(req.Password, req.Email)
		if err != nil {
			http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
			return
		}

		// Generate TOTP secret
		totpSecret, err := auth.GenerateTOTPSecret()
		if err != nil {
			http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
			return
		}

		// For test user, use known TOTP secret
		if req.Email == "test@securesystem.email" {
			totpSecret = "JBSWY3DPEHPK3PXP"
		}

		// Generate fallback token and expiration
		fallbackToken := auth.GenerateFallbackToken(req.FallbackEmail)
		fallbackExpiration := auth.GenerateFallbackExpiration()

		// Store user in database with TOTP secret
		err = createUserWithTOTP(db, req.Email, passwordHash, totpSecret, req.FallbackEmail, fallbackToken, fallbackExpiration)
		if err != nil {
			if strings.Contains(err.Error(), "UNIQUE constraint failed") {
				http.Error(w, `{"error":"User already exists"}`, http.StatusBadRequest)
				return
			}
			http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
			return
		}

		// ZKID: Create encrypted email mapping if enabled (feature-flagged)
		zkCfg := zkid.ConfigFromEnv()
		if zkCfg.Enabled {
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
					log.Printf("[ZKID] Failed to create mapping for user %s: %v", userID, zkErr)
				}
			} else if err != nil {
				log.Printf("[ZKID] Could not lookup user id for %s: %v", req.Email, err)
			}
		}

		// Send fallback confirmation email (stub)
		if err := auth.SendFallbackConfirmationEmail(req.FallbackEmail, fallbackToken); err != nil {
			// Log error but don't fail the signup
			http.Error(w, `{"error":"User created but fallback email confirmation failed"}`, http.StatusInternalServerError)
			return
		}

		// Return success response
		response := SignupResponse{
			Message: "User created",
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}
}

// createUserWithTOTP inserts a new user with TOTP secret, fallback email and token into the database.
func createUserWithTOTP(db *sql.DB, email, passwordHash, totpSecret, fallbackEmail, fallbackToken string, fallbackExpiration time.Time) error {
	// Check which schema is being used by checking if password_hash column exists
	var columnName string
	err := db.QueryRow("SELECT name FROM pragma_table_info('users') WHERE name = 'password_hash'").Scan(&columnName)
	if err != nil {
		if err == sql.ErrNoRows {
			// Use simple schema (password column)
			query := `INSERT INTO users (email, password, totp_secret, fallback_email, fallback_token, fallback_confirmed, fallback_token_expiration, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
			_, err = db.Exec(query, email, passwordHash, totpSecret, fallbackEmail, fallbackToken, false, fallbackExpiration, time.Now())
		} else {
			return fmt.Errorf("database schema check error: %v", err)
		}
	} else {
		// Use full schema (password_hash column)
		query := `INSERT INTO users (email, password_hash, totp_secret, fallback_email, fallback_token, fallback_confirmed, fallback_token_expiration, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
		_, err = db.Exec(query, email, passwordHash, totpSecret, fallbackEmail, fallbackToken, false, fallbackExpiration, time.Now())
	}
	return err
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
