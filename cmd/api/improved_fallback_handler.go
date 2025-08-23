package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"secure-email-mvp/pkg/auth"
	"secure-email-mvp/pkg/zkid"

	"github.com/google/uuid"
)

// ImprovedFallbackConfirmResponse represents the JSON response for fallback confirmation
type ImprovedFallbackConfirmResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

// improvedConfirmFallbackHandlerFactory returns an HTTP handler for confirming fallback email.
// This handler creates the actual user account after fallback email verification.
func improvedConfirmFallbackHandlerFactory(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only allow GET method
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Set content type to JSON
		w.Header().Set("Content-Type", "application/json")

		// Get token from query parameter
		token := r.URL.Query().Get("token")
		if token == "" {
			response := map[string]string{"error": "Invalid or expired confirmation token"}
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		// Validate token format
		if !auth.ValidateFallbackToken(token) {
			response := map[string]string{"error": "Invalid or expired confirmation token"}
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		// Get pending signup by token
		pendingSignupService := auth.NewPendingSignupService(db)
		pendingSignup, err := pendingSignupService.GetPendingSignupByToken(token)
		if err != nil {
			log.Printf("Failed to get pending signup by token: %v", err)
			response := map[string]string{"error": "Invalid or expired confirmation token"}
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		// Check if token has expired
		if auth.IsTokenExpired(pendingSignup.FallbackTokenExpiration) {
			log.Printf("Fallback token expired for email %s", pendingSignup.Email)
			// Clean up expired pending signup
			if cleanupErr := pendingSignupService.DeletePendingSignup(pendingSignup.ID); cleanupErr != nil {
				log.Printf("Failed to cleanup expired pending signup: %v", cleanupErr)
			}
			response := map[string]string{"error": "Invalid or expired confirmation token"}
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		// Check if user already exists (race condition protection)
		var existingUser int
		err = db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", pendingSignup.Email).Scan(&existingUser)
		if err != nil {
			log.Printf("Database error checking existing user: %v", err)
			http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
			return
		}
		if existingUser > 0 {
			log.Printf("User already exists for email %s, cleaning up pending signup", pendingSignup.Email)
			// Clean up pending signup since user already exists
			if cleanupErr := pendingSignupService.DeletePendingSignup(pendingSignup.ID); cleanupErr != nil {
				log.Printf("Failed to cleanup pending signup: %v", cleanupErr)
			}
			response := map[string]string{"error": "User already exists"}
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(response)
			return
		}

		// Create the actual user account
		userID, err := createUserFromPendingSignup(db, pendingSignup)
		if err != nil {
			log.Printf("Failed to create user from pending signup: %v", err)
			http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
			return
		}

		// ZKID: Create encrypted email mapping if enabled (feature-flagged)
		zkCfg := zkid.ConfigFromEnv()
		if zkCfg.Enabled {
			svc := zkid.NewService(db, zkCfg)
			// Fallback email optional pointer
			var fb *string
			if pendingSignup.FallbackEmail != "" {
				fb = &pendingSignup.FallbackEmail
			}
			_, zkErr := svc.CreateOrUpdateMapping(userID, pendingSignup.Email, fb)
			if zkErr != nil {
				log.Printf("[ZKID] Failed to create mapping for user %s: %v", userID, zkErr)
			}
		}

		// Clean up the pending signup
		if err := pendingSignupService.DeletePendingSignup(pendingSignup.ID); err != nil {
			log.Printf("Failed to cleanup pending signup after user creation: %v", err)
			// Don't fail the request since user was created successfully
		}

		// Return success response
		response := ImprovedFallbackConfirmResponse{
			Message: "Fallback email confirmed successfully. Your account has been created and you may now log in.",
			Status:  "account_created",
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)

		log.Printf("User account created successfully for %s (ID: %s)", pendingSignup.Email, userID)
	}
}

// createUserFromPendingSignup creates a new user account from a confirmed pending signup
func createUserFromPendingSignup(db *sql.DB, pendingSignup *auth.PendingSignup) (string, error) {
	// Generate new user ID using UUID
	userID := uuid.New().String()

	// Check which schema is being used by checking if password_hash column exists
	var columnName string
	err := db.QueryRow("SELECT name FROM pragma_table_info('users') WHERE name = 'password_hash'").Scan(&columnName)
	if err != nil {
		if err == sql.ErrNoRows {
			// Use simple schema (password column)
			query := `INSERT INTO users (id, email, password, totp_secret, fallback_email, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`
			_, err = db.Exec(query, userID, pendingSignup.Email, pendingSignup.PasswordHash, pendingSignup.TOTPSecret, pendingSignup.FallbackEmail, time.Now(), time.Now())
		} else {
			return "", err
		}
	} else {
		// Use full schema (password_hash column)
		query := `INSERT INTO users (id, email, password_hash, totp_secret, fallback_email, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`
		_, err = db.Exec(query, userID, pendingSignup.Email, pendingSignup.PasswordHash, pendingSignup.TOTPSecret, pendingSignup.FallbackEmail, time.Now(), time.Now())
	}

	if err != nil {
		return "", err
	}

	return userID, nil
}
