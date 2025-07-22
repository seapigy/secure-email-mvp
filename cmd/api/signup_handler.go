package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"time"

	"secure-email-mvp/pkg/auth"

	"golang.org/x/crypto/bcrypt"
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

		// Validate password
		if !isValidPassword(req.Password) {
			http.Error(w, `{"error":"Password must be at least 8 characters long"}`, http.StatusBadRequest)
			return
		}

		// Validate fallback email
		if req.FallbackEmail == "" {
			http.Error(w, `{"error":"Fallback email is required"}`, http.StatusBadRequest)
			return
		}
		if !isValidEmail(req.FallbackEmail) {
			http.Error(w, `{"error":"Invalid fallback email format"}`, http.StatusBadRequest)
			return
		}

		// Hash password using bcrypt
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
			return
		}

		// Generate fallback token and expiration
		fallbackToken := auth.GenerateFallbackToken(req.FallbackEmail)
		fallbackExpiration := auth.GenerateFallbackExpiration()

		// Store user in database
		err = createUserWithFallback(db, req.Email, string(hashedPassword), req.FallbackEmail, fallbackToken, fallbackExpiration)
		if err != nil {
			if strings.Contains(err.Error(), "UNIQUE constraint failed") {
				http.Error(w, `{"error":"User already exists"}`, http.StatusBadRequest)
				return
			}
			http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
			return
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

// createUserWithFallback inserts a new user with fallback email and token into the database. Used for account recovery security.
func createUserWithFallback(db *sql.DB, email, hashedPassword, fallbackEmail, fallbackToken string, fallbackExpiration time.Time) error {
	query := `INSERT INTO users (email, password, fallback_email, fallback_token, fallback_confirmed, fallback_token_expiration, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := db.Exec(query, email, hashedPassword, fallbackEmail, fallbackToken, false, fallbackExpiration, time.Now())
	return err
}

// createUser inserts a new user into the database (legacy function, not used in fallback flow).
func createUser(db *sql.DB, email, hashedPassword string) error {
	query := `INSERT INTO users (email, password, created_at) VALUES (?, ?, ?)`
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
