package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"secure-email-mvp/pkg/auth"

	"golang.org/x/crypto/bcrypt"
)

// LoginRequest represents the JSON request body for login
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse represents the JSON response for login
type LoginResponse struct {
	Token   string `json:"token"`
	Message string `json:"message"`
}

// loginHandlerFactory returns an HTTP handler for user login. It checks credentials, fallback confirmation, and issues a JWT on success.
func loginHandlerFactory(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only allow POST method
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Set content type to JSON
		w.Header().Set("Content-Type", "application/json")

		// Parse JSON request body
		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"Invalid JSON format"}`, http.StatusBadRequest)
			return
		}

		// Validate input
		if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" {
			http.Error(w, `{"error":"Email and password are required"}`, http.StatusBadRequest)
			return
		}

		// Look up user by email
		user, err := getUserByEmail(db, req.Email)
		if err != nil {
			if err == sql.ErrNoRows {
				// User not found - return same error as invalid password for security
				log.Printf("Login failed: user not found for email %s", req.Email)
				http.Error(w, `{"error":"Invalid email or password"}`, http.StatusUnauthorized)
				return
			}
			log.Printf("Database error during login for email %s: %v", req.Email, err)
			http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
			return
		}

		// Compare password with stored hash
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
		if err != nil {
			log.Printf("Login failed: invalid password for email %s", req.Email)
			http.Error(w, `{"error":"Invalid email or password"}`, http.StatusUnauthorized)
			return
		}

		// Check if fallback email is confirmed
		if !user.FallbackConfirmed {
			log.Printf("Login failed: fallback email not confirmed for email %s", req.Email)
			http.Error(w, `{"error":"Fallback email not confirmed"}`, http.StatusForbidden)
			return
		}

		// Generate JWT token
		token, err := auth.GenerateJWT(req.Email)
		if err != nil {
			log.Printf("Failed to generate JWT for email %s: %v", req.Email, err)
			http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
			return
		}

		// Login successful
		response := LoginResponse{
			Token:   token,
			Message: "Login successful",
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

// User represents a user from the database
type User struct {
	ID                int    `json:"id"`
	Email             string `json:"email"`
	Password          string `json:"password"`
	FallbackEmail     string `json:"fallback_email"`
	FallbackToken     string `json:"fallback_token"`
	FallbackConfirmed bool   `json:"fallback_confirmed"`
}

// getUserByEmail retrieves a user by email from the database, including fallback email status for security checks.
func getUserByEmail(db *sql.DB, email string) (*User, error) {
	query := `SELECT id, email, password, fallback_email, fallback_token, fallback_confirmed FROM users WHERE email = ?`

	var user User
	err := db.QueryRow(query, email).Scan(&user.ID, &user.Email, &user.Password, &user.FallbackEmail, &user.FallbackToken, &user.FallbackConfirmed)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
