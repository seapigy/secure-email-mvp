package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"secure-email-mvp/pkg/lockout"
)

// LockoutStatusResponse represents the response for lockout status queries
type LockoutStatusResponse struct {
	Email             string     `json:"email"`
	IsLockedOut       bool       `json:"is_locked_out"`
	FailedAttempts    int        `json:"failed_attempts"`
	MaxAttempts       int        `json:"max_attempts"`
	RemainingAttempts int        `json:"remaining_attempts"`
	LastFailedLogin   *time.Time `json:"last_failed_login,omitempty"`
	LockoutUntil      *time.Time `json:"lockout_until,omitempty"`
	LockoutRemaining  string     `json:"lockout_remaining,omitempty"`
	AttemptWindow     string     `json:"attempt_window"`
	IsWithinWindow    bool       `json:"is_within_window"`
}

// UnlockRequest represents the request to unlock a user account
type UnlockRequest struct {
	Email string `json:"email"`
}

// UnlockResponse represents the response for unlock operations
type UnlockResponse struct {
	Email     string `json:"email"`
	Unlocked  bool   `json:"unlocked"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

// lockoutStatusHandler handles GET /api/auth/lockout/status/{email}
// Returns the current lockout status for a user
func lockoutStatusHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract email from URL path
		email := r.URL.Query().Get("email")
		if email == "" {
			http.Error(w, `{"error":"Email parameter is required"}`, http.StatusBadRequest)
			return
		}

		// Get lockout status
		lockoutService := lockout.NewUserLockoutService(db)
		status, err := lockoutService.GetUserLockoutStatus(email)
		if err != nil {
			log.Printf("Failed to get lockout status for %s: %v", email, err)
			http.Error(w, `{"error":"Failed to retrieve lockout status"}`, http.StatusInternalServerError)
			return
		}

		// Build response
		response := LockoutStatusResponse{
			Email:             status.Email,
			IsLockedOut:       status.IsLockedOut(),
			FailedAttempts:    status.FailedAttempts,
			MaxAttempts:       status.MaxAttempts,
			RemainingAttempts: status.GetRemainingAttempts(),
			LastFailedLogin:   status.LastFailedLogin,
			LockoutUntil:      status.LockoutUntil,
			AttemptWindow:     status.AttemptWindow.String(),
			IsWithinWindow:    status.IsWithinAttemptWindow(),
		}

		// Add lockout remaining time if locked out
		if status.IsLockedOut() {
			remaining := status.GetLockoutRemainingTime()
			if remaining > 0 {
				response.LockoutRemaining = remaining.String()
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// unlockAccountHandler handles POST /api/auth/lockout/unlock
// Allows admin or user to unlock their account
func unlockAccountHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse request
		var req UnlockRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"Invalid request format"}`, http.StatusBadRequest)
			return
		}

		if req.Email == "" {
			http.Error(w, `{"error":"Email is required"}`, http.StatusBadRequest)
			return
		}

		// Check if user exists
		var userID string
		err := db.QueryRow("SELECT id FROM users WHERE email = ?", req.Email).Scan(&userID)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, `{"error":"User not found"}`, http.StatusNotFound)
				return
			}
			log.Printf("Failed to check user existence for %s: %v", req.Email, err)
			http.Error(w, `{"error":"Failed to process unlock request"}`, http.StatusInternalServerError)
			return
		}

		// Get current lockout status
		lockoutService := lockout.NewUserLockoutService(db)
		status, err := lockoutService.GetUserLockoutStatus(req.Email)
		if err != nil {
			log.Printf("Failed to get lockout status for %s: %v", req.Email, err)
			http.Error(w, `{"error":"Failed to retrieve lockout status"}`, http.StatusInternalServerError)
			return
		}

		// Check if account is actually locked
		if !status.IsLockedOut() {
			response := UnlockResponse{
				Email:     req.Email,
				Unlocked:  false,
				Message:   "Account is not currently locked",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}

		// Unlock the account by resetting failed attempts
		err = lockoutService.ResetUserFailedAttempts(req.Email)
		if err != nil {
			log.Printf("Failed to unlock account for %s: %v", req.Email, err)
			http.Error(w, `{"error":"Failed to unlock account"}`, http.StatusInternalServerError)
			return
		}

		log.Printf("Account unlocked for user: %s", req.Email)

		response := UnlockResponse{
			Email:     req.Email,
			Unlocked:  true,
			Message:   "Account successfully unlocked",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// lockoutConfigHandler handles GET /api/auth/lockout/config
// Returns the current lockout configuration
func lockoutConfigHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		config := lockout.DefaultConfig()

		response := map[string]interface{}{
			"max_attempts":             config.MaxAttempts,
			"lockout_duration":         config.LockoutDuration.String(),
			"attempt_window":           config.AttemptWindow.String(),
			"enabled":                  config.Enabled,
			"lockout_duration_minutes": int(config.LockoutDuration.Minutes()),
			"attempt_window_minutes":   int(config.AttemptWindow.Minutes()),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// lockoutStatsHandler handles GET /api/auth/lockout/stats
// Returns statistics about locked accounts (admin only)
func lockoutStatsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Count locked accounts
		var lockedCount int
		err := db.QueryRow(`
			SELECT COUNT(*) 
			FROM users 
			WHERE account_locked_until IS NOT NULL 
			AND account_locked_until > ?
		`, time.Now()).Scan(&lockedCount)
		if err != nil {
			log.Printf("Failed to count locked accounts: %v", err)
			http.Error(w, `{"error":"Failed to retrieve lockout statistics"}`, http.StatusInternalServerError)
			return
		}

		// Count accounts with failed attempts
		var failedAttemptsCount int
		err = db.QueryRow(`
			SELECT COUNT(*) 
			FROM users 
			WHERE failed_login_attempts > 0
		`).Scan(&failedAttemptsCount)
		if err != nil {
			log.Printf("Failed to count accounts with failed attempts: %v", err)
			http.Error(w, `{"error":"Failed to retrieve lockout statistics"}`, http.StatusInternalServerError)
			return
		}

		// Get recent lockout events (last 24 hours)
		var recentLockouts int
		err = db.QueryRow(`
			SELECT COUNT(*) 
			FROM users 
			WHERE last_failed_login > ?
		`, time.Now().Add(-24*time.Hour)).Scan(&recentLockouts)
		if err != nil {
			log.Printf("Failed to count recent lockouts: %v", err)
			http.Error(w, `{"error":"Failed to retrieve lockout statistics"}`, http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"currently_locked_accounts":     lockedCount,
			"accounts_with_failed_attempts": failedAttemptsCount,
			"recent_lockout_events_24h":     recentLockouts,
			"timestamp":                     time.Now().UTC().Format(time.RFC3339),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
