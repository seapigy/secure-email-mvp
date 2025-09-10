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
	Token            string `json:"token"`
	ExpiresAt        string `json:"expires_at"`
	AccountType      string `json:"account_type"`
	OrganizationID   string `json:"organization_id,omitempty"`
	OrganizationRole string `json:"organization_role,omitempty"`
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	// Lookup user with organization information
	var id string
	var storedHash string
	var mfaEnabled bool
	var emailVerified bool
	var accountType string
	var organizationID sql.NullString
	var organizationRole sql.NullString
	err := DB.QueryRow(`
		SELECT u.id, u.hashed_password, u.mfa_enabled, u.email_verified, u.account_type, om.organization_id, om.role
		FROM users u
		LEFT JOIN organization_members om ON u.id = om.user_id AND om.status = 'active'
		WHERE u.email = ? LIMIT 1
	`, req.Email).Scan(&id, &storedHash, &mfaEnabled, &emailVerified, &accountType, &organizationID, &organizationRole)
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

	// Add organization information if user is part of an organization
	if organizationID.Valid {
		resp.OrganizationID = organizationID.String
		if organizationRole.Valid {
			resp.OrganizationRole = organizationRole.String
		}
	}

	log.Printf("INFO login_success user_id=%s", id)

	// Log analytics event
	LogAnalyticsEvent(id, EventUserLogin, map[string]interface{}{
		"account_type": accountType,
		"success":      true,
	})
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
