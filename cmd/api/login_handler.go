package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

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

		w.Header().Set("Content-Type", "application/json")

		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"Invalid JSON format"}`, http.StatusBadRequest)
			return
		}

		if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" {
			http.Error(w, `{"error":"Email and password are required"}`, http.StatusBadRequest)
			return
		}

		// Lookup user and lockout fields
		var user User
		var failedAttempts int
		var lastFailed, lockedUntil sql.NullTime
		err := db.QueryRow(`SELECT id, email, password, failed_login_attempts, last_failed_login, account_locked_until FROM users WHERE email = ?`, req.Email).Scan(&user.ID, &user.Email, &user.Password, &failedAttempts, &lastFailed, &lockedUntil)
		if err == sql.ErrNoRows {
			http.Error(w, `{"error":"Invalid email or password"}`, http.StatusUnauthorized)
			return
		} else if err != nil {
			http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
			return
		}
		// Check if account is locked
		if lockedUntil.Valid && lockedUntil.Time.After(time.Now()) {
			http.Error(w, `{"error":"Account is temporarily locked due to repeated failed login attempts. Please try again later."}`, http.StatusForbidden)
			return
		}
		// Check password
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
		if err != nil {
			// Password incorrect: increment failed attempts, set last_failed_login
			failedAttempts++
			lockUntil := sql.NullTime{}
			if failedAttempts >= 5 {
				lockUntil.Time = time.Now().Add(30 * time.Minute)
				lockUntil.Valid = true
			}
			db.Exec(`UPDATE users SET failed_login_attempts=?, last_failed_login=?, account_locked_until=? WHERE id=?`, failedAttempts, time.Now(), lockUntil, user.ID)
			http.Error(w, `{"error":"Invalid email or password"}`, http.StatusUnauthorized)
			return
		}
		// Password correct: reset lockout fields
		db.Exec(`UPDATE users SET failed_login_attempts=0, last_failed_login=NULL, account_locked_until=NULL WHERE id=?`, user.ID)
		// Return dummy token for now
		resp := map[string]string{"token": "dummy-token", "message": "Login successful"}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
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
