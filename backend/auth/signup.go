package auth

// DO NOT EDIT EXISTING CODE - new file added
// Signup handler (Go). Assumes a global DB variable is set by main application.
//
// Required env:
// - DATABASE_URL used by main to open DB and set auth.DB

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

// DB variable is declared in session.go

type signupRequest struct {
	Username      string `json:"username"`
	Email         string `json:"email"`
	Password      string `json:"password"`
	AccountType   string `json:"account_type,omitempty"`
	FallbackEmail string `json:"fallback_email"`
	SetupMFA      bool   `json:"setup_mfa,omitempty"`
}

type safeUserResp struct {
	ID          string        `json:"id"`
	Username    string        `json:"username"`
	Email       string        `json:"email"`
	AccountType string        `json:"account_type"`
	CreatedAt   string        `json:"created_at"`
	RecoveryKey string        `json:"recovery_key"`  // Only returned once at signup
	MFA         *mfaSetupData `json:"mfa,omitempty"` // MFA data if setup was requested
}

type mfaSetupData struct {
	Secret      string   `json:"secret"`
	QRCodeURL   string   `json:"qr_code_url"`
	BackupCodes []string `json:"backup_codes"`
}

// POST /api/auth/signup
func SignupHandler(w http.ResponseWriter, r *http.Request) {
	// Basic JSON decode
	var req signupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	req.Username = strings.TrimSpace(strings.ToLower(req.Username))
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.FallbackEmail = strings.TrimSpace(strings.ToLower(req.FallbackEmail))
	if req.Username == "" || req.Email == "" || req.Password == "" || req.FallbackEmail == "" {
		http.Error(w, "missing fields", http.StatusBadRequest)
		return
	}

	// Validate that fallback email is different from primary email
	if strings.EqualFold(req.FallbackEmail, req.Email) {
		http.Error(w, "fallback email must be different from primary email", http.StatusBadRequest)
		return
	}

	// Validate username characters (alphanumeric + . _ -)
	for _, ch := range req.Username {
		if !(ch == '.' || ch == '_' || ch == '-' || (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9')) {
			http.Error(w, "invalid username", http.StatusBadRequest)
			return
		}
	}

	// Validate password strength
	if err := ValidatePasswordStrength(req.Password); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Hash password
	hashed, err := HashPassword(req.Password)
	if err != nil {
		log.Printf("ERROR hashing password: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Default account type and validate
	if req.AccountType == "" {
		req.AccountType = "free"
	}

	// Validate account type
	if req.AccountType != "free" && req.AccountType != "premium" && req.AccountType != "enterprise" {
		http.Error(w, "invalid account type", http.StatusBadRequest)
		return
	}

	// For Premium/Enterprise signup, create placeholder subscription
	// This allows users to sign up without requiring live Stripe keys
	var subscriptionID string
	if req.AccountType == "premium" || req.AccountType == "enterprise" {
		subscriptionID = uuid.New().String()
	}

	// Insert user
	id := uuid.New().String()
	now := time.Now().UTC()

	tx, err := DB.Begin()
	if err != nil {
		log.Printf("ERROR begin tx: %v", err)
		http.Error(w, "database error", http.StatusServiceUnavailable)
		return
	}
	defer tx.Rollback()

	// Check duplicate email or username+domain uniqueness is handled in SQL indexes; here we proactively check email
	var exists int
	err = tx.QueryRow("SELECT 1 FROM users WHERE email = ? LIMIT 1", req.Email).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		// For Oracle/Postgres use $1 param; your DB layer must adapt to driver
		// TODO: use a proper query placeholder depending on driver
	}

	// Generate recovery private key (256-bit random)
	recoveryKey, err := GenerateRandomToken(32) // 32 bytes = 256 bits
	if err != nil {
		log.Printf("ERROR generating recovery key: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	hashedRecoveryKey := HashToken(recoveryKey)

	// Generate verification code for primary email
	verificationCode, err := GenerateRandomToken(6) // 6-character code
	if err != nil {
		log.Printf("ERROR generating verification code: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	hashedVerificationCode := HashToken(verificationCode)
	verificationExpiresAt := now.Add(30 * time.Minute) // 30-minute expiry

	// Note: Fallback email uses the same verification system as primary email

	// Try insert (use DB-specific placeholder as necessary)
	_, err = tx.Exec(
		`INSERT INTO users (id, username, email, hashed_password, account_type, account_status, verification_code, verification_code_expires_at, fallback_email, fallback_email_verified, recovery_private_key_hashed, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, 'pending_verification', ?, ?, ?, ?, ?, ?, ?)`,
		id, req.Username, req.Email, hashed, req.AccountType, hashedVerificationCode, verificationExpiresAt, req.FallbackEmail, false, hashedRecoveryKey, now, now,
	)
	if err != nil {
		// Log detailed error information
		log.Printf("ERROR inserting user - Full error: %v", err)
		log.Printf("ERROR inserting user - Username: %s, Email: %s, AccountType: %s", req.Username, req.Email, req.AccountType)
		log.Printf("ERROR inserting user - FallbackEmail: %s", req.FallbackEmail)

		// return nice error for duplicate
		if strings.Contains(err.Error(), "UNIQUE") || strings.Contains(err.Error(), "unique") {
			http.Error(w, "username or email already exists", http.StatusConflict)
			return
		}

		// Return the actual error for debugging
		http.Error(w, fmt.Sprintf("database error: %v", err), http.StatusServiceUnavailable)
		return
	}

	// Create placeholder subscription for Premium/Enterprise users
	if subscriptionID != "" {
		_, err = tx.Exec(`
			INSERT INTO subscriptions (id, user_id, status, plan, start_date, end_date, created_at, updated_at)
			VALUES (?, ?, 'active', ?, ?, ?, ?, ?)
		`, subscriptionID, id, req.AccountType, now, now.AddDate(0, 1, 0), now, now)
		if err != nil {
			log.Printf("ERROR creating placeholder subscription: %v", err)
			http.Error(w, "database error", http.StatusServiceUnavailable)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		log.Printf("ERROR commit: %v", err)
		http.Error(w, "database error", http.StatusServiceUnavailable)
		return
	}

	// Create default inbox and welcome message
	if err := CreateDefaultInbox(id); err != nil {
		log.Printf("ERROR creating default inbox: %v", err)
		// Don't fail signup if inbox creation fails, just log it
	}

	// Setup MFA if requested
	var mfaData *mfaSetupData
	if req.SetupMFA {
		mfaResponse, err := SetupMFAForUser(id, req.Username)
		if err != nil {
			log.Printf("ERROR setting up MFA during signup: %v", err)
			// Don't fail signup if MFA setup fails, just log it
		} else {
			mfaData = &mfaSetupData{
				Secret:      mfaResponse.Secret,
				QRCodeURL:   mfaResponse.QRCodeURL,
				BackupCodes: mfaResponse.BackupCodes,
			}
			log.Printf("INFO mfa_setup_completed user_id=%s", id)
		}
	}

	// Minimal safe response (recovery key will be sent after email verification)
	resp := safeUserResp{
		ID:          id,
		Username:    req.Username,
		Email:       req.Email,
		AccountType: req.AccountType,
		CreatedAt:   now.Format(time.RFC3339),
		MFA:         mfaData,
		// RecoveryKey: recoveryKey, // Will be sent after email verification
	}

	// Send verification code to external email (fallback email)
	if err := SendVerificationEmail(req.FallbackEmail, verificationCode, req.Username); err != nil {
		log.Printf("ERROR sending verification email: %v", err)
		// Don't fail signup if email sending fails, just log it
	} else {
		log.Printf("INFO verification_email_sent user_id=%s fallback_email=%s", id, req.FallbackEmail)
	}

	// Log verification code for development
	log.Printf("INFO verification_code_generated user_id=%s code=%s", id, verificationCode)

	// Log event (non-sensitive)
	log.Printf("INFO user_created id=%s account_type=%s", id, req.AccountType)

	// Log analytics event
	LogAnalyticsEvent(id, EventUserSignup, map[string]interface{}{
		"account_type": req.AccountType,
		"success":      true,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}
