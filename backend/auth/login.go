package auth

// DO NOT EDIT EXISTING CODE - new file added
// Login handler: verify Argon2id password, create opaque session token, store hashed token.

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token       string `json:"token"`
	ExpiresAt   string `json:"expires_at"`
	AccountType string `json:"account_type"`
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	// Lookup user
	var id string
	var storedHash string
	var mfaEnabled bool
	var emailVerified bool
	var accountType string
	err := DB.QueryRow("SELECT id, hashed_password, mfa_enabled, email_verified, account_type_new FROM users WHERE email = ? LIMIT 1", req.Email).Scan(&id, &storedHash, &mfaEnabled, &emailVerified, &accountType)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		log.Printf("ERROR DB lookup: %v", err)
		http.Error(w, "database error", http.StatusServiceUnavailable)
		return
	}
	ok, err := VerifyPassword(req.Password, storedHash)
	if err != nil || !ok {
		// log failure non-sensitive
		log.Printf("WARN login_failed email=%s", req.Email)
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	// Check if email is verified
	if !emailVerified {
		http.Error(w, "email_not_verified", http.StatusForbidden)
		return
	}

	// If MFA is enabled, require the second step (not handled here; orchestrate frontend to call /api/auth/validate-mfa)
	if mfaEnabled {
		// indicate 2FA required
		http.Error(w, "mfa_required", http.StatusForbidden)
		return
	}

	// Generate opaque token and hash for storage
	rawToken, err := GenerateRandomToken(32)
	if err != nil {
		log.Printf("ERROR token generate: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	hashedToken := HashToken(rawToken)
	expiresAt := time.Now().Add(24 * time.Hour).UTC()

	// Store session
	_, err = DB.Exec(`INSERT INTO sessions (id, user_id, token_hash, expires_at, created_at) VALUES (?, ?, ?, ?, ?)`,
		uuid.New().String(), id, hashedToken, expiresAt, time.Now().UTC(),
	)
	if err != nil {
		log.Printf("ERROR store session: %v", err)
		http.Error(w, "database error", http.StatusServiceUnavailable)
		return
	}

	// Return raw token to client (store it client-side securely)
	resp := loginResponse{
		Token:       rawToken,
		ExpiresAt:   expiresAt.Format(time.RFC3339),
		AccountType: accountType,
	}

	log.Printf("INFO login_success user_id=%s", id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
