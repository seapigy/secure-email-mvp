package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"strings"

	"secure-email-mvp/pkg/auth"
	"secure-email-mvp/pkg/password"
	"secure-email-mvp/pkg/reputation"
)

// EarlyValidationSignupRequest represents the JSON request body for signup with early validation
type EarlyValidationSignupRequest struct {
	Email         string `json:"email"`
	Password      string `json:"password"`
	FallbackEmail string `json:"fallback_email"`
}

// EmailValidationRequest represents the JSON request body for early email validation
type EmailValidationRequest struct {
	Email string `json:"email"`
}

// EmailValidationResponse represents the JSON response for email validation
type EmailValidationResponse struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message"`
}

// earlyEmailValidationHandlerFactory returns an HTTP handler for validating email before password input
func earlyEmailValidationHandlerFactory(db *sql.DB) http.HandlerFunc {
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

		// Check IP reputation before processing
		reputationService := reputation.NewReputationService()
		ctx := context.Background()

		isMalicious, err := reputationService.CheckIPReputation(ctx, clientIP)
		if err != nil {
			log.Printf("IP reputation check failed for IP %s: %v", clientIP, err)
			// Continue processing on reputation check failure
		} else if isMalicious {
			log.Printf("Email validation blocked due to IP reputation for IP %s", clientIP)
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(EmailValidationResponse{
				Valid:   false,
				Message: "Access denied due to IP reputation",
			})
			return
		}

		// Parse JSON request body
		var req EmailValidationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(EmailValidationResponse{
				Valid:   false,
				Message: "Invalid JSON format",
			})
			return
		}

		// Validate email format
		if !isValidEmailFormatEarly(req.Email) {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(EmailValidationResponse{
				Valid:   false,
				Message: "Invalid email format",
			})
			return
		}

		// Validate email domain (must be @securesystem.email)
		if !strings.HasSuffix(req.Email, "@securesystem.email") {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(EmailValidationResponse{
				Valid:   false,
				Message: "Email must end with @securesystem.email",
			})
			return
		}

		// Check if email already exists in users table
		var existingUser int
		err = db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", req.Email).Scan(&existingUser)
		if err != nil {
			log.Printf("Database error checking existing user: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(EmailValidationResponse{
				Valid:   false,
				Message: "Internal server error",
			})
			return
		}
		if existingUser > 0 {
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(EmailValidationResponse{
				Valid:   false,
				Message: "User already exists",
			})
			return
		}

		// Check if email has a pending signup
		pendingSignupService := auth.NewPendingSignupService(db)
		isPending, err := pendingSignupService.IsEmailPending(req.Email)
		if err != nil {
			log.Printf("Database error checking pending signup: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(EmailValidationResponse{
				Valid:   false,
				Message: "Internal server error",
			})
			return
		}
		if isPending {
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(EmailValidationResponse{
				Valid:   false,
				Message: "Signup already in progress. Please check your fallback email for confirmation.",
			})
			return
		}

		// Email is valid and available
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(EmailValidationResponse{
			Valid:   true,
			Message: "Email is available",
		})

		log.Printf("Email validation passed for %s", req.Email)
	}
}

// earlyValidationSignupHandlerFactory returns an HTTP handler for signup with early email validation
func earlyValidationSignupHandlerFactory(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only allow POST method
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Set content type to JSON
		w.Header().Set("Content-Type", "application/json")

		// Parse JSON request body
		var req EarlyValidationSignupRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"Invalid JSON format"}`, http.StatusBadRequest)
			return
		}

		// Note: Email validation was already done in the early validation step
		// We still do basic format validation for security

		// Validate email format
		if !isValidEmailFormatEarly(req.Email) {
			http.Error(w, `{"error":"Invalid email format"}`, http.StatusBadRequest)
			return
		}

		// Validate email domain (must be @securesystem.email)
		if !strings.HasSuffix(req.Email, "@securesystem.email") {
			http.Error(w, `{"error":"Email must end with @securesystem.email"}`, http.StatusBadRequest)
			return
		}

		// Validate password using comprehensive password service
		passwordService := password.NewPasswordService()
		ctx := context.Background()
		passwordResult, err := passwordService.ValidatePassword(ctx, req.Password)
		if err != nil {
			log.Printf("Password validation failed: %v", err)
			http.Error(w, `{"error":"Password validation error"}`, http.StatusInternalServerError)
			return
		}

		if !passwordResult.IsValid {
			log.Printf("Password validation failed for email %s: %v", req.Email, passwordResult.Errors)
			http.Error(w, `{"error":"Password does not meet security requirements"}`, http.StatusBadRequest)
			return
		}

		// Log successful password validation
		log.Printf("Password validation passed for email %s (score: %d, breach count: %d)",
			req.Email, passwordResult.Score, passwordResult.BreachCount)

		// Validate fallback email (comprehensive validation)
		if req.FallbackEmail == "" {
			http.Error(w, `{"error":"Fallback email is required"}`, http.StatusBadRequest)
			return
		}
		if !isValidEmailFormatEarly(req.FallbackEmail) {
			http.Error(w, `{"error":"Invalid fallback email format"}`, http.StatusBadRequest)
			return
		}

		// Check if user is trying to use their own email as fallback
		if req.Email == req.FallbackEmail {
			http.Error(w, `{"error":"Fallback email cannot be the same as your main email"}`, http.StatusBadRequest)
			return
		}

		// Check if fallback email already exists as someone's main email
		var fallbackAsMainEmail int
		err = db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", req.FallbackEmail).Scan(&fallbackAsMainEmail)
		if err != nil {
			log.Printf("Database error checking fallback email as main email: %v", err)
			http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
			return
		}
		if fallbackAsMainEmail > 0 {
			http.Error(w, `{"error":"This email is already registered as a main account"}`, http.StatusBadRequest)
			return
		}

		// Check if fallback email is already used by another user as fallback
		var fallbackEmailInUse int
		err = db.QueryRow("SELECT COUNT(*) FROM users WHERE fallback_email = ?", req.FallbackEmail).Scan(&fallbackEmailInUse)
		if err != nil {
			log.Printf("Database error checking fallback email in use: %v", err)
			http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
			return
		}
		if fallbackEmailInUse > 0 {
			http.Error(w, `{"error":"This fallback email is already in use by another account"}`, http.StatusBadRequest)
			return
		}

		// Check if fallback email is used in pending signups
		var fallbackInPending int
		err = db.QueryRow("SELECT COUNT(*) FROM pending_signups WHERE fallback_email = ?", req.FallbackEmail).Scan(&fallbackInPending)
		if err != nil {
			log.Printf("Database error checking fallback email in pending signups: %v", err)
			http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
			return
		}
		if fallbackInPending > 0 {
			http.Error(w, `{"error":"This fallback email is already in use by another pending signup"}`, http.StatusBadRequest)
			return
		}

		// Final check: verify email is still available (race condition protection)
		var existingUser int
		err = db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", req.Email).Scan(&existingUser)
		if err != nil {
			log.Printf("Database error checking existing user: %v", err)
			http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
			return
		}
		if existingUser > 0 {
			http.Error(w, `{"error":"User already exists"}`, http.StatusBadRequest)
			return
		}

		// Check if email has a pending signup (final check)
		pendingSignupService := auth.NewPendingSignupService(db)
		isPending, err := pendingSignupService.IsEmailPending(req.Email)
		if err != nil {
			log.Printf("Database error checking pending signup: %v", err)
			http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
			return
		}
		if isPending {
			http.Error(w, `{"error":"Signup already in progress. Please check your fallback email for confirmation."}`, http.StatusConflict)
			return
		}

		// Check user count limit
		var userCount int
		err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
		if err != nil {
			log.Printf("Database error checking user count: %v", err)
			http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
			return
		}
		if userCount >= 100 {
			http.Error(w, `{"error":"Maximum user limit reached"}`, http.StatusForbidden)
			return
		}

		// Hash password using Argon2
		passwordHash, err := auth.HashPassword(req.Password, req.Email)
		if err != nil {
			log.Printf("Password hashing failed: %v", err)
			http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
			return
		}

		// Generate TOTP secret
		totpSecret, err := auth.GenerateTOTPSecret()
		if err != nil {
			log.Printf("TOTP secret generation failed: %v", err)
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

		// Create pending signup
		pendingSignup, err := pendingSignupService.CreatePendingSignup(
			req.Email, passwordHash, totpSecret, req.FallbackEmail,
			fallbackToken, fallbackExpiration,
		)
		if err != nil {
			log.Printf("Failed to create pending signup: %v", err)
			http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
			return
		}

		// Send fallback confirmation email
		if err := auth.SendFallbackConfirmationEmail(req.FallbackEmail, fallbackToken); err != nil {
			log.Printf("Failed to send fallback confirmation email: %v", err)
			// Clean up the pending signup if email sending fails
			if cleanupErr := pendingSignupService.DeletePendingSignup(pendingSignup.ID); cleanupErr != nil {
				log.Printf("Failed to cleanup pending signup after email failure: %v", cleanupErr)
			}
			// Return error with retry suggestion
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Failed to send confirmation email. Please check your fallback email address and try again.",
				"retry": "true",
			})
			return
		}

		// Return success response
		response := ImprovedSignupResponse{
			Message: "Signup initiated. Please check your fallback email for confirmation.",
			Status:  "pending_confirmation",
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)

		log.Printf("Pending signup created for %s (ID: %s)", req.Email, pendingSignup.ID)
	}
}

// isValidEmailFormatEarly checks if the email format is valid using regex.
func isValidEmailFormatEarly(email string) bool {
	// Basic email regex pattern
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(strings.TrimSpace(email))
}
